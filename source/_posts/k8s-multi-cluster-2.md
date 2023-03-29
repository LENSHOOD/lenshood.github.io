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

#### 应用模型扩展

Member Cluster 加入后，怎么定义跨集群的应用资源呢？在 KubeFed 的概念中，任何可以下发的应用资源类型，都需要被定义为 “联邦资源类型”（FederatedType，如上图），之后才能被 KubeFed 识别并调度。

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

[Karmada](https://karmada.io/) 是华为开源的多集群管理系统，目前是 CNCF 的沙箱项目。Karmada 在多集群管理上功能非常丰富，在集群管理、灵活调度、应用模型等方面都提供了较为完善的解决方案。

从部署架构上看（如下图），Karmard 多集群控制面是逻辑上的概念，其控制面进程支持部署在任意 K8s 集群或是 VM 上。控制面组件中，Karmada API Server 用于接受请求，创建的资源存储在 etcd，Karmada Scheduler 用于产生调度决策，而 Karmada 自有的资源，则分别被其对应的 Karmada Controllers 监听并执行实际的动作。可以看到 Karmada 控制面组件的设计与 K8s 非常相似。

![](https://karmada.io/zh/assets/images/architecture-37447d3b4fceeae700e488373138d808.png)

下图是 Karmada 在进行多集群管理过程中引入的一些概念和自有资源，其中：

- Resource Template：指在 Karmada 上下文中的 K8s 资源，如 Deployment，DaemonSet 等等
- PropagationPolicy：资源传播策略，用于定义某个资源模板需要以某些规则下发，可以认为是对调度器的提示
- ResourceBinding：产生调度决策后，会将资源模板与实际下发的集群绑定起来存储在 ResourceBinding 中
- OverridePolicy：对某些特定集群中下发的资源属性进行修改，覆盖模板值
- Work：实际操作资源下发的组件

![](https://karmada.io/zh/assets/images/karmada-resource-relation-0e46a98d960615afc1860e2b4e1f4ca1.png)

#### 应用模型扩展

Karmada 多集群管理的工作流程是通过用户在 Karmada 上下文（即通过 Karmada API Server）中创建 K8s 资源开始的，用户创建的 K8s 资源在 Karmada 中称为 Resource Template。

如下以一个最简单的 Nginx Deployment 为例：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 6
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
```

在单集群语境下，上述 Deployment 在创建后 K8s 集群会通过 Kubelet 在某个节点上实际启动 Nginx 容器，而在 Karmada 多集群语境下，该资源会被 Karmada 控制面保存，而不会立即开始尝试启动容器，Karmada 会将该 Deployment 视为一种创建 Nginx 应用的 ”模板“。

之后用户需要再创建一个 PropagationPolicy 资源来描述这个 Nginx 应用实际需要运行的集群：

```yaml
apiVersion: policy.karmada.io/v1alpha1
kind: PropagationPolicy
metadata:
  name: nginx-propagation
spec:
  resourceSelectors:
    - apiVersion: apps/v1
      kind: Deployment
      name: nginx
  placement:
    clusterAffinity:
      clusterNames:
        - member1
        - member2
    replicaScheduling:
      replicaDivisionPreference: Weighted
      replicaSchedulingType: Divided
      weightPreference:
        staticWeightList:
          - targetCluster:
              clusterNames:
                - member1
            weight: 1
          - targetCluster:
              clusterNames:
                - member2
            weight: 2
```

可见，PropagationPolicy 能通过 `resourceSelectors` 选中 Nginx Deployment（即选中了资源模板），之后在 `placement` 段中定义 Nginx 应用的调度策略，`clusterAffinity` 定义了该应用与 Member 集群的亲和性，在这里指 Nginx 需要在 `member1`和 `member2`中创建。`replicaScheduling` 段则描述了资源模板中的 `replicas: 6` 实际上在 Member 集群中的副本数，`replicaDivisionPreference` 和 `replicaSchedulingType` 指明六个副本需要按权重平均分配在 Member 集群中，由于 `weightPreference`  定义了 `member1`和 `member2 `的权重比例是 1:2，因此实际上在  `member1` 会部署 2 个副本，而 `member2` 会部署 4 个副本。

Karmada 的应用资源模型设计是前向兼容的，避免了 KubeFed 中需要修改原始应用资源的问题，从单集群演进而来的应用，最少只需要增加 PropagationPolicy 就能过渡到多集群。

#### 跨集群动态调度





### 1.3 OCM

### 1.4 Rancher

## 2. 演进趋势

1. 真的需要扁平网络吗？（跨集群 pod 同一个网络）
2. 跨集群方案怎么解决数据同步问题
3. 如何统一管理传统云资源，如 VM，块存储，VPC 等

## 3. 总结
