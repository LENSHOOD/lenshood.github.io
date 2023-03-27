---
title: Kubernetes 多集群（下）：方案与演进
date: 2023-03-26 13:19:34
tags:
- multi-cluster
- multi-cloud
- k8s
categories:
- Kubernetes
---

{% asset_img header.jpg 500 %}

本文（分上下两部分）介绍了 K8s 多集群的由来以及实现多集群所面临的核心问题，之后分析并探讨了现有的 K8s 多集群方案，最后根据目前实现方案的痛点与挑战，设想了未来的演进趋势。

本篇是下半部分，主要讨论目前实现 K8s 多集群的开源方案、对现状问题的讨论以及可能的演进方向。

<!-- more -->

## 1. 几种方案对比

上一篇我们已经讨论了实现多集群管理所涉及到的 4 个核心问题：

- 部署模型：包括了多集群管理控制面所处的位置、集群间网络连通性以及跨集群的服务注册于发现
- 跨集群应用调度：涉及了通用调度模型以及需要通过不同的调度策略对应用和集群的属性进行匹配
- 应用模型扩展：应用模型需要在规格和状态上进行扩展，同时也应考虑前向兼容性以及支持自定义资源
- 集群即资源：为了更灵活的自动扩缩，将集群视为可以进行生命周期管理的资源，并考虑合理的状态模型

接下来我们通过探究常见的一些开源多集群管理方案，来对比它们之间的特性。

### 1.1 KubeFed

[KubeFed](https://github.com/kubernetes-sigs/kubefed) 是 Kubernetes 多集群特别兴趣小组构建的一套多集群管理方案，是相对较早的试图解决多集群管理问题的开源方案。KubeFed 主要聚焦于通过定义集群联邦来解决跨集群应用模型的定义、应用调度以及服务发现问题。

KubeFed 的整体架构如下图所示：

![](https://github.com/kubernetes-sigs/kubefed/blob/master/docs/images/concepts.png?raw=true)

首先，在部署模型上，KubeFed 作为控制面独占一个 K8s 集群，称为 Host Cluster，而实际部署应用的集群，称为 Member Cluster。任何 K8s 集群，想要成为集群联邦内的一个成员，都需要通过 KubeFed 来进行 “Join”，“Join” 实际上是在 Host Cluster 中创建了一种类型为 `KubeFedCluster` 的自定义资源（即上图中的 “Cluster Configuration”），在其中描述了待加入集群的 API 端点，证书、Token 等用于从外部连接到集群的信息，KubeFed 会通过这类信息来向 Member Cluster 下达指令。

Member Cluster 加入后，怎么定义跨集群的资源呢？在 KubeFed 的概念中，任何可以下发的资源类型，都需要被定义为 “联邦资源类型”（FederatedType，如上图），之后才能被 KubeFed 识别并调度。

举例说明，当我们期望在集群联邦中创建标准的 Deployment 资源时，需要：

**创建 CRD，类型起名叫 “FederatedDeployment”，代表新增一种 FederatedType：**

一个 FederatedType，在 Spec 定义中必须包含三种元素即：Template，Placement 和 Overrides（如上图），其中 Template 表示 KubeFed 将实际创建的真实的资源，在这里便是原生的 Deployment；Placement 代表该 Deployment 将被部署在哪些 Member Cluster 中；而 Overrides 则是用于当某些 Member Cluster 中部署的 Deployment 与其他集群不太一样，差异化的部分会定义在 Overrides 中。

[示例](https://github.com/kubernetes-sigs/kubefed/blob/master/example/sample1/federateddeployment.yaml)如下：

```yaml
apiVersion: types.kubefed.io/v1beta1
kind: FederatedDeployment
metadata:
  name: test-deployment
  namespace: test-namespace
spec:
  template:
    ... ...
    spec:
      replicas: 3
      selector:
        matchLabels:
          app: nginx
      ... ...
  placement:
    clusters:
    - name: cluster2
    - name: cluster1
  overrides:
  - clusterName: cluster2
    clusterOverrides:
    - path: "/spec/replicas"
      value: 5
    ... ...
```

**创建了 FederatedType 之后，还需要将其注册至 KubeFed**

上述 FederatedDeployment，是由用户定义的 CRD，为了使 KubeFed 能真正的监听这个 CRD 从而实现应用分发，还需要注册一下。

通过[定义一个 “FederatedTypeConfig” 对象](https://github.com/kubernetes-sigs/kubefed/blob/d6f10d29c3785dc55abc954dec0afeddce4893ef/charts/kubefed/templates/federatedtypeconfig.yaml#L39)实现注册：

```yaml
apiVersion: core.kubefed.io/v1beta1
kind: FederatedTypeConfig
metadata:
  name: deployments.apps
spec:
  federatedType:
    group: types.kubefed.io
    kind: FederatedDeployment
    pluralName: federateddeployments
    scope: Namespaced
    version: v1beta1
  propagation: Enabled
  targetType:
    group: apps
    kind: Deployment
    pluralName: deployments
    scope: Namespaced
    version: v1
```

如此就实现了对 FederatedDeployment 和 Deployment 的关联，KubeFed 在实际创建资源时，会创建 Spec 取自 FederatedDeployment 中 template 的 Deployment 资源。

> 实际上，标准的 K8s 资源 API 都已经被创建好，在安装 KubeFed 时会一并安装，而如果有自定义资源需要被 Federated，KubeFed 也提供了 kubefedctl CLI 工具来简化操作。



### 1.2 Karmada

### 1.3 OCM

### 1.4 Rancher

## 2. 演进趋势

1. 真的需要扁平网络吗？（跨集群 pod 同一个网络）
2. 跨集群方案怎么解决数据同步问题
3. 如何统一管理传统云资源，如 VM，块存储，VPC 等

## 3. 总结
