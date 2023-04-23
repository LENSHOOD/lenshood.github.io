---
title: （WIP）理解 K8s 多集群（下）：解决方案的选择与演进趋势
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

Karmada 的应用跨集群调度实现的很完善，通过多个组件相互配合，不仅实现了传统的由用户指定的亲和性、权重、分组等调度偏好，还支持 Taint/Tolerantion、优先级、基于资源的调度，故障转移，动态重调度等更加自动化的调度方式。

在 Karmada 中与调度相关的组件关系如下图：

{% asset_img karmada-sched.jpg %}

在 Karmada 中，应用期望部署在哪些集群中，副本数有多少，实际状态是什么，都存储在 ResourceBinding 对象中。因此整个调度逻辑也都围绕着 ResourceBinding 来组织。如上图所示，与调度相关的组件包括了调度器 scheduler、重调度器 de-scheduler 以及污点管理器 taint-manager，这些组件产生的调度决策都会作用于 ResourceBinding 上。

##### 灵活的调度模式

在前面示例的 PropagationPolicy 中，已经展示了集群亲和性的配置，除了直接通过 Cluster 名称，还可以以采用如标签等方式选择合适的集群。此外，还可以通过 PropagationPolicy 中的 SpreadConstraint 特性来对集群进行分组。调度器 scheduler 通过读取 ResourceBinding 中与 PropagationPolicy 相关的信息来产生调度决策，并将决策结果写回 ResourceBinding。

假如 Member Cluster 出现了故障，集群控制器 cluster-controller 会在对应的 Karmada Cluster 对象上打污点，污点管理器 taint-manager 得知 Member Cluster 上存在污点后，也会修改原调度决策，并将修改作用在 ResourceBinding 上。

##### 动态重调度

重调度器 de-scheduler 通过 estimator 来获取应用在集群中的实际状态，进而决定是否修改原调度决策，这一修改也会落在 ResourceBinding 上。

每个 estimator 都对应一个 Member Cluster，estimator 通过检查应用在当前 Member Cluster 中的目标副本数，和实际副本数是否一致来发现调度失败的情况，de-scheduler 汇总所有 estimator 的信息，就能够知晓某个应用的现状，并决定是否要发起重调度。

#### 网络连通与服务发现

Karmada 支持借助 [Submariner](https://submariner.io/) 或 [Istio](https://istio.io/) 来实现 Member Cluster 之间的网络连通性。通过实现 [Multi-Cluster Services API](https://github.com/kubernetes/enhancements/blob/master/keps/sig-multicluster/1645-multi-cluster-services-api/README.md) 来支持跨集群的服务注册与发现能力。

##### 网络连通

Submariner 为 K8s 多集群提供了基于 Overlay 网络的连通方案，能够实现跨 L3 层的 Overlay，且不要求多集群使用相同的 CNI 插件。

![](https://submariner.io/images/submariner/architecture.jpg)

上图是 Submariner 的总体架构，可以看到它本质上是通过 VxLan 技术建立的 Overlay 网络。每个集群内安装 Gateway 来接受跨集群的流量，所有跨集群流量都会通过 Route Agent 路由至 Gateway。Gateway 之间通过公有网络建立基于 IPSec 的加密通道，在其上传输跨集群流量。Broker 可以视为 Submariner 的控制平面，用于控制和协调。

#####服务发现

Karmada 基于 Multi-Cluster Services API 的 ServiceExport 和 ServiceImport 实现了相关控制逻辑来构建跨集群的服务发现。

{% asset_img karmada-sd.jpg %}

如上图所示，当 Member Cluster 0 中的某个服务 Service 需要被导出时，先由用户创建 ServiceExport。此时 Karmada 的 ServiceExport 控制器会监听到 Member Cluster 0 中创建的 ServiceExport 对象，并配合 EndpointSlice 控制器将 Member Cluster 0 中需要导出的 Service 和对应的 Endpoints 复制一份到 Karmada 控制集群，以备后续导出。

接下来，用户在需要导入服务的集群中（图中是 Member Cluster 1）创建 ServiceImport。一旦 ServiceImport 对象被创建，ServiceImport 控制器就会基于待发布的 Service（Original Service） 创建出对应的 “派生” Service（Derived Service），并和 Endpoints 一并通过 Propagation 机制下发到 Member Cluster 1，实现对 Service 的导入。

### 1.3 OCM

[OCM](https://open-cluster-management.io/) 即 Open Cluster Management，是阿里与红帽共同推出的一种 K8s 多集群实现方案。其最大的特色在于采用借鉴了 K8s 控制面组件+Kubelet 架构模式的所谓 [“Hub-Agent” 设计架构](https://open-cluster-management.io/concepts/architecture/#hub-spoke-architecture)，通过一个小型轻量级的控制集群，就能够管理多至数千个集群。

![](https://github.com/open-cluster-management-io/OCM/raw/main/assets/ocm-arch.png)

上图所示的是 OCM 的总体架构，可以发现它与 KubeFed 或 Karmada 最大的区别就在于，其每一个工作集群（OCM 中称为 Managed Cluster）中都安装有一个 “Klusterlet” 组件（恰好类比于 Kubelet）。在多集群管理流程中，控制集群（OCM 中称为 Hub Cluster）只负责生成各个 ManagedCluster 中应当被下发的应用资源模板（OCM 中称为 “处方”），实际的资源管理与状态上报工作，是由 Klusterlet 主动向 Hub Cluster 拉取处方，基于处方的内容管理应用的生命周期，并定期推送应用资源的状态。

正如 OCM 的架构概念所述：“试想，如果Kubernetes中没有kubelet，而是由控制平面直接操作容器守护进程，那么对于一个中心化的控制器，管理一个超过5000节点的集群，将会极其困难。 同理，这也是OCM试图突破可扩展性瓶颈的方式，即将“执行”拆分卸入各个单独的代理中，从而让hub cluster可以接受和管理数千个集群。” 比对 Karmada 是通过在控制面创建每个工作集群对应一个的 Work 组件来实施集群管理，OCM 的 Klusterlet 就类似于把 Karmada 的 Work 放在了工作集群上运行。

#### 应用模型扩展

OCM 是通过名为 `ManifestWork` 的对象来描述应用：

```yaml
apiVersion: work.open-cluster-management.io/v1
kind: ManifestWork
metadata:
  namespace: <target managed cluster>
  name: hello-work-demo
spec:
  workload:
    manifests:
      - apiVersion: apps/v1
        kind: Deployment
        metadata:
          name: hello
          namespace: default
        spec:
          ... ...
status:
  conditions:
    ... ...
  resourceStatus:
    manifests:
      - conditions:
          ... ...
        resourceMeta:
          group: apps
          kind: Deployment
          name: hello
          ... ...
```

一个 `ManifestWork` 能够描述多个应用资源，此外每一个 Managed Cluster 在 Hub Cluster 中都拥有一个命名空间，`ManifestWork` 创建在哪个命名空间中，对应 Managed Cluster 的 Klusterlet 就会将其拉取下来，并如实的创建应用资源。而应用资源实际的状态信息，也会由 Klusterlet 更新回 `ManifestWork` 中。

不过，显然 OCM 的应用模型扩展并没有考虑前向兼容的问题。

#### 动态调度

与 Karmada 很类似，OCM 也是通过名为 `Placement` 的对象来实现动态调度：

```yaml
apiVersion: cluster.open-cluster-management.io/v1beta1
kind: Placement
metadata:
  name: placement1
spec:
  numberOfClusters: 3
  clusterSets:
    - prod
  predicates:
    - requiredClusterSelector:
        labelSelector:
          matchLabels:
            purpose: test
        claimSelector:
          matchExpressions:
            - key: platform.open-cluster-management.io
              operator: In
              values:
                - aws
```

当`Placement` 创建过后，调度逻辑会按照其描述来生成名为 `PlacementDecision` 的调度决策：

```yaml
apiVersion: cluster.open-cluster-management.io/v1beta1
kind: PlacementDecision
metadata:
  labels:
    cluster.open-cluster-management.io/placement: placement1
  name: placement1-decision-1
status:
  decisions:
    - clusterName: cluster1
    - clusterName: cluster2
    - clusterName: cluster3
```

相关控制器监听到调度决策后就会按要求在 Managed Cluster 的命名空间中创建 `ManifestWork`，完成调度流程。

另外，OCM 的 Add-on 插件体系也提供了灵活的框架来允许用户自定义并扩展内建的调度逻辑。

### 1.4 Gardener

[Gardener](https://gardener.cloud/) 是 SAP 开源的 K8s 多集群解决方案，与前面几种方案不同，Gardener 专注于 K8s 集群即服务（Kubernetes-as-a-Service）。

在设计概念上，Gardener 期望作为一种管理大量 K8s 集群的组件，借助 Gardner，用户能够实现在接入各种不同类型底层基础设施的同时，方便的在其上构建标准的 K8s 集群。

与 Cluster-Api 不同，Gardener 更进一步，除了能在各种差异化基础设施上管理 K8s 集群的生命周期，还能够确保在这些基础设施上运行的 K8s 集群具有完全相同的版本、配置和行为，这能简化应用的多云迁移。Gardener 还提供了[一个页面](https://k8s-testgrid.appspot.com/conformance-all)专门介绍其不同版本的标准化K8s 集群与不同云提供商的兼容情况。

![](https://gardener.cloud/__resources/gardener-architecture-detailed_945c90.png)

上图所示的是 Gardener 的整体架构图。从垂直分层的角度看，Gardener 自身及其管理的 K8s 集群可分为三层：

- Garden Cluster：Gardener 的控制集群，主要用于定义并管理实际的 K8s 工作集群。

- Seed Cluster：Gardener 并不是直接在基础设施上创建工作集群的，相反，Gardener 定义了 Seed 集群的概念。在 Seed 集群中以标准 K8s Workload 的形式运行着多个工作集群的控制面。这种 ”K8s in K8s“ 的形式简化了工作集群控制面的高可用设计，也非常易于扩展。

- Shoot Cluster：工作集群的数据面节点，可以视为实际的工作集群。

在 Garden Cluster 中用户可以通过构建 `Seed` 对象（Gardener 定义的一种 CR，下同）来描述 Seed 集群，而通过 `Shoot` 对象来描述 Shoot 集群，最后通过构建 `CloudProfile` 对象来描述下层基础设施的配置。通常在每一个 IaaS Region 中都会运行一个 Seed Cluster，由它来持有当前 IaaS Region 下工作集群的控制面。每个 Seed Cluster 中都运行着 Gardenlet 用来从 Garden Cluster 中获取工作集群的创建需求。实际的集群创建动作也是由 Gardenlet 来完成的。

因此 Gardener 的设计类似于将 K8s 的概念扩展到了多集群领域：

- Kubernetes 控制面 = Garden 集群
- Kubelet = Gardenlet
- Node = Seed 集群
- Pod = Shoot 集群

## 2. 演进趋势

通过上述内容，我们了解到了目前的一些实现 K8s 多集群管理的开源解决方案，它们或多或少的实现了一些我们认为的实现 K8s 多集群的核心要素。

事实上，K8s 多集群的诞生本身就是由于业务发展的需要，因此即使是目前的各种开源方案已经实现了多集群管理的许多能力，但仍然有一些领域支持的并不完善。我们将在这一节讨论 K8s 多集群未来可能需要支持的能力和趋势。

### 2.1 多租户

提供多租户服务厂商的一项共识就是：“不要信任任何租户”，因为恶意租户是难以避免的。因此讨论多租户时，我们经常会讨论租户间的隔离性和安全性问题。

K8s 本身对多租户的支持一直不太完善，从目前来看大致存在三种多租户实现方案：

#### 基于 Namespace 的隔离

通过 K8s 的 namespace 机制（之后简称 ns），可以把不同的工作负载进行分组和隔离。不同的 ns 之间可以存在同名的工作负载，RBAC 设置权限的粒度也是由 ns 定义的。

因此基于 ns 的逻辑隔离是相对简单的一种多租户形式。只要结合访问控制策略，将租户用户对集群的访问权限限定在某个 ns 下，就能实现最基础的多租。通过设置 ResourceQuota 对象，也可以限制 ns 的资源配额。

-----------------------图-----------------------

逻辑隔离最大的问题在于这是一种 “存在于控制面” 的隔离，即仅在与控制面交互时会受到隔离的限制。某一个 ns 下的租户，可能无法通过 API Server 查看或修改其他 ns 下的工作负载，但实际上由于 [K8s 对数据面网络连通性的要求](https://kubernetes.io/zh-cn/docs/concepts/services-networking/#the-kubernetes-network-model)，Pod 之间默认是连通的，因此如果不加以限制，租户之间的应用实际上完全可以相互访问。

通过设置合理的 [NetworkPolicy](https://kubernetes.io/zh-cn/docs/concepts/services-networking/network-policies/) 来控制 Pod 之间的网络流量策略，能够在一定程度上解决上述问题，不过 NetworkPolicy 需要通过 CNI 插件来实现，因此对集群选择的 CNI 也提出了要求。

另外，数据面隔离还涉及到容器运行时的问题。传统的容器运行时采用的都是共享系统内核的策略，那么就有理由相信恶意容器可能会利用内核漏洞突破容器的限制，访问在同一节点上其他租户的数据。有诸如 Kata Containers、gVisor、Firecracker 等安全容器方案能缓解该问题。

#### 基于多集群的隔离

上述逻辑隔离的策略毕竟安全性和隔离性都比较低，在需要更高隔离等级的多租户场景下并不适合。

因此本文重点讨论的多集群方案就能很容易的被应用在租户隔离场景下。每个租户拥有自己独立的集群，虽然基础设施可能都处于同一家云提供商，但通过 AZ、VPC 等手段能够方便快捷的实现租户间的物理隔离。

-----------------------图-----------------------

通过 K8s 多集群来实现租户隔离，的确是一种隔离性和安全性都更佳的方案，但 K8s 集群本身的复杂性也导致了小规模集群场景下控制面组件对资源的过多消耗。假如租户应用实际只需要 2 个数据面节点就足够，但为了集群的正常运转，仍旧需要最少 3 个节点来部署控制面组件，以实现最低的组件选举要求。

K8s 集群控制面资源在总资源中的占比，与数据面节点数量成反比，假如有大量小规模集群租户的存在，那么要么会导致多租户服务提供商的资源成本过高，要么会导致租户使用服务的底价过高，这两点都不利于业务发展。

#### 基于虚拟集群的隔离



### 2.2 有状态应用的调度



### 2.3 统一管理云资源



1. 真的需要扁平网络吗？（跨集群 pod 同一个网络）
2. 跨集群方案怎么解决数据同步问题
3. 如何统一管理传统云资源，如 VM，块存储，VPC 等
3. 多租 [k8s multiple tenancy wg](https://github.com/kubernetes/community/blob/master/wg-multitenancy/annual-report-2020.md)

## 3. 总结
