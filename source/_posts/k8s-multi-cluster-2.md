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

[KubeFed](https://github.com/kubernetes-sigs/kubefed) 是 Kubernetes 多集群特别兴趣小组构建的一套多集群管理方案，是相对较早的试图解决多集群管理问题的开源方案。KubeFed 主要聚焦于通过定义集群联邦来解决跨集群应用模型的定义、应用调度以及服务发现问题。由于多方面的原因，目前 KubeFed 已经归档，不再活跃更新。

KubeFed 的整体架构如下图所示：

![](https://github.com/kubernetes-sigs/kubefed/blob/master/docs/images/concepts.png?raw=true)

首先，在部署模型上，KubeFed 作为控制面独占一个 K8s 集群，称为 Host Cluster，而实际部署应用的集群，称为 Member Cluster。任何 K8s 集群，想要成为集群联邦内的一个成员，都需要通过 KubeFed 来进行 “Join”，“Join” 实际上是在 Host Cluster 中创建了一种类型为 `KubeFedCluster` 的自定义资源（即上图中的 “Cluster Configuration”），在其中描述了待加入集群的 API 端点，证书、Token 等用于从外部连接到集群的信息，KubeFed 会通过这类信息来向 Member Cluster 下达指令。

#### 资源模型扩展

Member Cluster 加入后，怎么定义跨集群的资源呢？在 KubeFed 的概念中，任何可以下发的资源类型，都需要被定义为 “联邦资源类型”（FederatedType，如上图），之后才能被 KubeFed 识别并调度。

举例说明，当期望在集群联邦中创建标准的 Deployment 资源时，需要：

**1. 创建 CRD，类型起名叫 “FederatedDeployment”，代表新增一种 FederatedType：**

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

**2. 创建了 FederatedType 之后，还需要将其注册至 KubeFed**

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

通过上述内容我们会发现，KubeFed 定义的资源模型，是无法前向兼容的，这也就导致如果从单集群迁移到 KubeFed 多集群，需要花费大量成本来适配。

#### 跨集群调度

显然，按上述方式定义的联邦资源，其调度方式属于全静态的预定义资源分发，在 placement 中设定的是什么，KubeFed 就会按照预定值来调度应用。然而这种方式非常不灵活，假如某个 Member Cluster 资源不足，则就会出现调度 Pending。

KubeFed 提供了名为 `ReplicaSchedulingPreference` 的调度策略来解决动态调度问题：

```yaml
apiVersion: scheduling.kubefed.io/v1alpha1
kind: ReplicaSchedulingPreference
metadata:
  name: test-deployment
  namespace: test-namespace
spec:
  targetKind: FederatedDeployment
  totalReplicas: 10
  rebalance: true
  clusters:
   cluster1:
     weight: 2
   cluster2:
     weight: 3
```

上述 Spec 中，totalReplicas 代表目标资源的总副本数（10 个），clusters 中通过 weight 来分配不同集群中的副本数（cluster1 分 4 个，cluster2 分 6 个），rebalance 则代表假如某集群资源不足则自动重平衡。基于此，KubeFed 就可以根据集群实际的状态来分配资源，当某个集群中该资源的期望副本数与实际副本数不符时，就可以自动进行动态调度。

当然，这种动态调度只是基于比较简单的规则，并没有对调度策略做过多细化。

#### 服务注册与发现

KubeFed 在最初的设计中通过引入 ”ServiceDNSRecord“ 类型来通过与外部的 DNS 服务交互来实现跨集群的服务发现。

任何需要跨集群发布的服务都应创建 ServiceDNSRecord 类型来告知 KubeFed 可以将对应的服务地址注册到外部的 DNS 服务中，以实现服务的注册和发现。

但是在 KubeFed v2 的 [KEP](https://github.com/kubernetes-sigs/kubefed/blob/master/docs/keps/20200619-kubefed-pull-reconciliation.md#other-main-reconciliation-loops) 中提到，KubeFed 将不再使用 ServiceDNSRecord 相关的功能，而是寻求其他方案例如 Service Mesh 等来实现服务注册于发现的功能，因此现在我们会看到，KubeFed 的源码中已经不再包含相关的控制逻辑了。

### 1.2 Karmada

### 1.3 OCM

### 1.4 Rancher

## 2. 演进趋势

1. 真的需要扁平网络吗？（跨集群 pod 同一个网络）
2. 跨集群方案怎么解决数据同步问题
3. 如何统一管理传统云资源，如 VM，块存储，VPC 等

## 3. 总结
