---
title: TiDB 学习课程 Lesson-2
date: 2020-08-19 22:51:46
tags:
- tidb
categories:
- TiDB
---

这一节，我们需要对一个完整的 TiDB 集群进行性能测试，为了实现这一目标，本文会通过以下几个步骤来介绍如何从零开始实现 TiDB 的性能测试：

1. **借助 kind 在单机模拟集群**

2. **通过 TiDB Operator 部署 TiDB 集群到 K8S**

3. **根据机器硬件调整 TiDB 线程池配置**

4. **sysbench 测试**

5. **go-ycsb 测试**

6. **go-tpc 测试**

7. **性能瓶颈分析**



### 借助 kind 在单机模拟集群

[kind](https://kind.sigs.k8s.io/) 能通过容器技术来模拟 K8S 集群：在 docker 容器中运行 K8S Node，再在 Node 中创建容器环境，并管理 Pod，名副其实的 docker in docker。

##### 安装 kind

首先，我们需要在本地安装 kind （这里忽略了 docker 的安装步骤，想要安装 docker 可以参见其[安装文档](https://www.docker.com/products/docker-desktop)）：

```shell
brew install kind
```

brew 安装的 kind 版本是 v0.8.1，如果想体验最新特性，可以从 [kind 的 github 仓库](https://github.com/kubernetes-sigs/kind/)获取最新代码并通过 go 编译后安装。

kind 本身并不依赖 kubectl，但为了方便，我们也一并将之安装：

```shell
brew install kubectl
```

kind 安装好之后，我们就可以着手创建我们的 K8S 集群了。

##### 构建 K8S

我们知道，任意一个 K8S Node，都至少需要一个 `kubelet` 和一个 `docker env` ，对于需要用作 control panel 的 node，还需要安装 `apiserver`、`etcd`、`scheduler` 、`controller-manager`等等组件，通常我们会使用`kubeadm`工具来安装。

不过 kind 作为将 docker 容器用作 node 的技术，其本身已经提供了完整的 node 镜像：`node-image`，并且我们只需要配置其 cluster 描述文件，即可快速创建出我们想要的集群。

以下是官方提供的基准 kind 配置文件，我在本地也使用了完全同样的配置（文件名 one-control-three-worker-kind-cluster.yaml）：

```yaml
# this config file contains all config fields with comments
# NOTE: this is not a particularly useful config file
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
# patch the generated kubeadm config with some extra settings
kubeadmConfigPatches:
- |
  apiVersion: kubelet.config.k8s.io/v1beta1
  kind: KubeletConfiguration
  evictionHard:
    nodefs.available: "0%"
# patch it further using a JSON 6902 patch
kubeadmConfigPatchesJSON6902:
- group: kubeadm.k8s.io
  version: v1beta2
  kind: ClusterConfiguration
  patch: |
    - op: add
      path: /apiServer/certSANs/-
      value: my-hostname
# 1 control plane node and 3 workers
nodes:
# the control plane node config
- role: control-plane
# the three workers
- role: worker
- role: worker
- role: worker
```

执行以下命令创建我们的 K8S 集群：

```shell
kind create cluster --config one-control-three-worker-kind-cluster.yaml
```

完成后，就可以通过 `kubectl cluster-info` 来查看集群状态了，执行结果：

```shell
Kubernetes master is running at https://127.0.0.1:57825
KubeDNS is running at https://127.0.0.1:57825/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
```

再来执行 `kubectl get nodes` 看一看 node 状态：

```shell
kind-control-plane   Ready    master   3m45s   v1.18.2
kind-worker          Ready    <none>   3m5s    v1.18.2
kind-worker2         Ready    <none>   3m6s    v1.18.2
kind-worker3         Ready    <none>   3m5s    v1.18.2
```

到此为止，我们的 K8S 集群就创建完毕了。