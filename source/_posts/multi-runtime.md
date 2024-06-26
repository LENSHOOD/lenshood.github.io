---
title: 多运行时架构解决了什么问题？
date: 2022-12-06 13:44:50
tags: 
- multi-runtime
- mesh
- dapr
- layotto
categories:
- Software Engineering
---

{% asset_img header.png 500 %}

本文介绍了多运行时架构的概念以及一些实现方案，并讨论了目前的实现中存在的限制与挑战，最后讨论了未来可能的架构形态。

<!-- more -->

## 服务化演进中的问题

自从数年前微服务的概念被提出，到现在基本成了技术架构的标配。微服务的场景下衍生出了对分布式能力的大量需求：各服务之间需要相互协作和通信，以及共享状态等等，因此就有了各种中间件来为业务服务提供这种分布式能力。

{% asset_img 1.png 500 %}

我们熟知的 "Spring Cloud 全家桶" 正是凭借着对各种中间件优秀的集成与抽象能力，成为了当时炙手可热的项目。

然而随着业务的快速发展，组织规模的不断扩大，微服务越来越多，系统规模越来越大则是服务化体系架构演进的必然。这就带来了两方面复杂度的上升：

1. 服务治理与能力接入的复杂度
   - 服务治理代表了系统中服务资源的地图及其获取途径，例如通过注册发现服务提供图谱能力，路由、网关、LB 服务提供获取途径。
   - 能力接入则代表了如何使用系统中的服务能力，例如通过中间件提供的 API 协议或是封装的 SDK 来接入该中间件。
   - 各种业务服务越多、中间件越复杂，整个系统服务治理与接入复杂度就会急剧上升。
2. 团队协作的复杂度
   - 该复杂度主要体现在团队的认知负载上，复杂的依赖、沟通、协作将明显拖慢交付进度
   - 正如康威定律所述的，由于服务复杂度的上升，团队之间的交互成本也随之上升

> 如下是复杂度上升问题的一个显而易见的例子。
>
> {% asset_img 2.png 500 %}
>
> 当系统中的中间件都通过 SDK 作为其外化能力的控制方式，来封装协议、数据结构与操作方法。随着中间件数量和种类不断增多，大量孤立的 SDK 被绑定在业务服务上，导致两方面问题：
>
> 1. 版本升级困难：SDK 与业务服务的强依赖性导致想要升级 SDK 版本变得异常复杂与缓慢
> 2. 业务服务难以异构：SDK 所支持的语言反向限制了业务服务所能选择的语言，例如 Spring Cloud 几乎没有官方的多语言支持

如何治理这种不断上升的复杂度呢？复杂问题归一化是一种不错的手段。



## 什么是多运行时架构

多运行时微服务架构（Multi-Runtime Microservice Architecture）也被简称为多运行时架构，是由 Red Hat 的首席架构师 Bilgin Ibryam 在 2020 年初所提出的一种微服务架构形态，它相对完整的从理论和方法的角度阐述了多运行时架构的模型（实际上，在 2019 年末，微软的 dapr v0.1.0 就已经发布）。

暂时先抛开到底什么是 “多运行时” 不谈（因为多运行时这个名字个人觉得起的可能不太妥当），先看看多运行时架构都包括了哪些内容。

### 分布式应用四大类需求

上一节提到，为了治理不断上升的复杂度问题，归一化是手段之一。归一化的第一步就是对问题就行归类。

Bilgin Ibryam 梳理了分布式应用的各类需求后，将其划分到了四个领域内：

{% asset_img 3.webp %}

分别是：

- 生命周期：即应用从开发态到运行态之间进行打包、部署、扩缩容等需求
- 网络：分布式系统中各应用之间的服务发现、容错、灵活的发布模式、跟踪和遥测等需求
- 状态：我们期望服务是无状态的，但业务本身一定需要有状态，因此包含对缓存、编排调度、幂等、事务等需求
- 绑定：与外部服务之间进行集成可能面临的连接器、交互适配、协议转换等需求

Bilgin Ibryam 认为，应用之间对分布式能力的需求，无外乎这四大类。且在 Kubernetes 成为云原生场景下运行时的事实标准后，对生命周期这部分的需求已经基本被覆盖到了。

因此实际上我们更关注的是如何归一化其他三种需求。

> **与单机应用的类比**
>
> 单机应用一般大都是以用户态进程的形式运行在操作系统上。显然，与微服务类似，单机应用的核心关注点也是业务逻辑，与业务关系不大的支撑能力，都要依赖操作系统来完成。
>
> 因此上述由 Bilgin 归纳的分布式应用四大类需求，其实我们很容易就可以和单机应用进行合理的类比：
>
> | 支撑能力 | 单机应用                     | 分布式应用                        |
> | -------- | ---------------------------- | --------------------------------- |
> | 生命周期 | 用户态进程                   | Kubernetes                        |
> | 网络     | 网络协议、域名解析、路由服务 | 服务发现/注册、负载均衡、流量管理 |
> | 状态     | 文件系统                     | 数据库、对象存储、块存储          |
> | 绑定     | 标准库、系统调用             | 事件分发、分布式事务、消息路由    |
>
> 从上述类比来看我们发现，单单是 Kubernetes 可能还不足以称为是 “云原生操作系统”，除非有一种解决方案，能在分布式环境下，把其他几项支撑能力也进行归一化整合，才能理直气壮的冠此大名。



### Service Mesh 的成功

Service Mesh 在近几年的高速发展，让我们认识到网络相关的需求是如何被归一化并与业务本身解耦的：

通过流量控制能力实现多变的发布模式以及对服务韧性的灵活配置，通过安全能力实现的开箱即用的 mTLS 双向认证来构建零信任网络，通过可观察性能力实现的网络层Metrics，Logging 和 Tracing 的无侵入式采集。

而上述服务治理能力，全部被代理到 Sidecar进程中完成。这就实现了 codebase level 的解耦，网络相关的分布式能力完全抛弃 SDK。

{% asset_img 4.png 400 %}

伴随着 Service Mesh 的成功，我们不禁会想到，是否可以将另外的两种需求 --- 状态和绑定 --- 也进行 Mesh 化改造呢？

### 分布式能力 Mesh 化

基于对 Service Mesh 的拓展，我们大可以将其他的能力也进行 Mesh 化，每一类能力都以 Sidecar 的形式部署和运作：

{% asset_img 5.png 500 %}

在业界也有不少从某些能力角度切入的方案：

{% asset_img 6.png 500 %}

我们可以发现，各类方案都有自己的一套对某些能力需求的 Mesh 化方案，合理的选择它们，的确满足了分布式能力 Mesh 化的要求，但却引入了新的问题：

- 复杂度从业务服务下沉到了 Mesh 层：多种 Mesh 化方案之间缺乏一致性，导致选型和运维的成本很高
- 多个 Sidecar 进程会带来不小的资源开销，很多解决方案还需要搭配控制面进程，资源消耗难以忽视

对业务复杂度上升的归一化，现在变成了对 Mesh 复杂度上升的归一化。

### Multi-Runtime = Micrologic + Mecha

Bilgin Ibryam 在多运行时微服务架构中，对前述讨论的各种问题点进行了整合，提出了 Micrologic + Mecha 的架构形态：

{% asset_img 7.png 500 %}

在 Micrologic 中只包含业务逻辑，尽可能的把分布式系统层面的需求剥离出去，放到 Mecha 中。从 Mecha 的命名就可以明白它的功能：

{% asset_img 8.png 300 %}

由提供各种分布式能力的 "机甲" 组成的 Sidecar 进程，与 "裸奔的" 业务逻辑一起部署。因为是 Micrologic 进程和 Mecha 进程共同部署的这种多个 "运行时" 的架构，所以称之为 "多运行时架构"。

Mecha 不仅成功的将分布式能力从耦合的业务进程中抽取出来，还整合了方案，避免了多种方案混合的额外成本。可以说 Mecha 在本质上提供了一个**分布式能力抽象层**。

因此与其叫 "多运行时架构"，不如叫 "面向能力的架构"。



## 微软的尝试：dapr

[Dapr](https://dapr.io/) 是微软主导开发并开源的一种 Mecha runtime，high level 上看它处在整个架构的中间层：

{% asset_img 9.jpeg %}

自上而下分别是业务层、Dapr Runtime层、基础设施层。Dapr 通过 Http 或 gRPC API 向业务层提供分布式能力抽象，通过称为 "Component" 的接口定义，实现对具体基础设施的插件式管理。

### Building Blocks

作为一个合格的 Mecha，最关键的就是如何定义分布式能力抽象层。如何把各类中间件提供的分布式能力定义清楚是一项挑战。Dapr 中定义的分布式能力抽象层，称为 Building Blocks。顾名思义，就是一系列的 "构建块"，每一个块定义了一种分布式能力。

{% asset_img 10.png %}

其中有一些 Blocks 的能力由 dapr 自己就能实现，有一些则需要由实际的基础设施或中间件来实现。选取几个典型举例说明：

- Service-to-service Invocation：提供服务间调用的能力，其中也隐含了服务的注册与发现。该 Block 的能力由 dapr 直接实现。
- State management：提供状态管理能力，最简单的就是存取状态。该 Block 需要其他基础设施通过 Component 的形式实现，例如定义一个 Redis Component。
- Publish and subscribe：提供消息发布和订阅的能力，这是非常典型的一种分布式能力。也需要通过基础设施来实现，如定义一个 Kafka Component。

### 使用示例

下图所示的是一个 demo 场景：Checkout 服务与 Order Processor 服务，分别处于订单结算业务的上下游。当用户通过 Checkout 服务进行订单付款后，Checkout 服务会发出一条 "xx 订单已付款" 的事件，该事件将被 Order Processor 服务订阅，用于进行订单的下一步操作，如出库、发货等动作。

{% asset_img 11.png %}

显然在引入 dapr 之前，Checkout 和 Order Processor 服务势必都需要依赖某种消息服务的 SDK，而引入 dapr 后，dapr runtime 作为 Sidecar 与业务服务共同部署。

此时，对于 Checkout 服务，想要发送事件，只需要 POST 调用本地 dapr 进程所提供的端点：`/publish/orderpubsub/orders`。其中 `publish` 代表需要使用的 Building Block 类型，即 Publish and subscribe，而 `orderpubsub` 代表使用具体哪一个 Block，最后 `orders` 则是消息 topic。

对于 Order Processor 服务，想要订阅事件，需要提供一个专为 dapr 使用的端点 `/dapr/subscribe`，dapr runtime 启动后，会自动与该端点通信，获取到订阅消息所需的关键信息，如：订阅哪一个 Block（`orderpubsub`）中的哪一个 topic（`orders`）？以及当 dapr 收到消息后的 callback 地址是什么（`POST /orders`）？

对于 Publish and subscribe，dapr 自身无法实现该能力，需要定义外部消息服务的 Component，在上图中我们可见该 Component 的 Block 名称就是 `orderpubsub`，而实际定义的类型是一个 Redis，该 Redis 的访问方式则在`spec.metadata` 中详述。

最后，引入 dapr 之后，Checkout 将事件发送给 dapr，左右两个 dapr 分别连接 Redis 进行消息的发送与订阅，Order Processor 的 `/orders` 端点会在事件到来时被 dapr 调用。

### Kubernetes 下部署 dapr

在 K8S 下，dapr 除了以 Sidecar 形式存在的 runtime 以外，还包含了一些控制面的实例：

{% asset_img 12.png %}

- Actor partition placement(Placement)：管理 dapr actor 与创建它的 Pod 之间的映射关系（Actor 是一种 building block，用于以 actor 模型执行用户自定义的任务）
- Dapr runtime injector(Sidecar injector)：与 istio 的自动注入能力类似，监听 Pod 创建并判断是否自动注入 dapr sidecar
- Cert authority and Identity(Sentry)：用于管理和分发证书，以支持 dapr sidecar 之间的 mTLS 通信
- Update component changes(Operator)：用于监听 Component CR 的变更，并将变更应用到每一个 dapr sidecar 上



## Dapr 的限制与挑战

Dapr 期望通过定义一个能容纳所有需求的分布式能力抽象层，来彻底解放业务逻辑。从归一化的角度看，不得不说这是一种大胆而富有野心的尝试，理想条件下的确能非常优雅地解决问题。但现实总是充斥着各种跳脱出理想的情况，dapr 在推广的过程中遇到了很多限制与挑战。

### 与 Service Mesh 整合

作为面向开发侧提供的能力抽象层，dapr 在网络能力上包含了 mTLS、Observability 与 Resiliency（即超时重试熔断等），但并没有包含诸如负载均衡、动态切换、金丝雀发布等运维侧的流量管理能力。

{% asset_img 14.png %}

因此对于不断走向成熟的业务系统，可能既要 Service Mesh 在运维侧的流量管理能力，又要 dapr 在开发侧的分布式抽象能力，不管谁先谁后，都将面临一个问题：怎样搭配使用它们才是正确的？某些场景下可以做适配，如：

- 对于 distributed tracing 的能力，如果采用 Service Mesh 来实现，则需要考虑将原本 dapr 直连的中间件也加入 mesh 网络，否则会 trace 不到。但从 distributed tracing 本身功能角度讲，更应该使用 dapr。
- mTLS 应该只在 dapr 或者 Service Mesh 中开启，而不应该都开启。

但 dapr 与 Service Mesh 配合使用中难以避免的是开销的问题，包括资源开销和性能开销。

每个应用 Pod 携带两种 sidecar，再加上 dapr 和 Service Mesh 自己的控制面应用（高可用方案主备或多副本），这些资源开销是无法忽略，甚至是非常可观的。而由于 Service Mesh 网络代理的流量劫持，网络调用需要先经过 dapr sidecar，再经过网络代理 sidecar，被代理两次，也会造成一定的性能开销。

下表是汇总的 [Dapr 官方标注的 daprd 资源与性能消耗数据](https://doczs.dapr.io/operations/performance-and-scalability/perf-service-invocation/#data-plane-performance)，以及 [Istio v1.16（最新版未找到）官方标注的 envoy 资源与性能消耗数据](https://istio.io/v1.16/docs/ops/deployment/performance-and-scalability/#performance-summary-for-istio-hahahugoshortcode-s0-hbhb)：

|           | **CPU**   | **Mem** | **Latency P90** **@1k TPS** |
| --------- | --------- | ------- | --------------------------- |
| **daprd** | 0.48 vCPU | 23 MiB  | 1.4ms                       |
| **envoy** | 0.35 vCPU | 40 MiB  | 2.65ms                      |

简单计算一下就会发现，当拥有 1000 个业务实例时，dapr + istio 的 Sidecar 进程可能会消耗 800+ vCPU 和 60+ GiB 内存。

随着分布式能力抽象层的不断扩展，到底哪些属于开发侧，哪些属于运维侧，也许不会像现在这样泾渭分明了。因此已经有对 Multi-Runtime 与 Service Mesh 能力边界越来越模糊的讨论。

### Sidecarless？

从上一节的表格我们发现，资源消耗以及性能的问题其实不只是 dapr 下的场景，实际上它是 sidecar 模式自有的限制，因此在 Service Mesh 领域的讨论中，已经有提出 Sidecarless 的概念了，即通过 DaemonSet 而不是 Sidecar 的形式来部署网络代理。

{% asset_img 18.png %}

对于网络代理的 Sidecarless 化，支持方认为它能带来高性能、低资源消耗的优点，而反对方则认为它会导致安全性与隔离性差、故障的爆炸半径过大等缺点。

那么，Mecha 是否也可能会走向 Sidecarless 呢？

与网络代理的 Sidecarless 类似，如果将 Mecha 做成 DaemonSet，其优劣势也差不多。而 DaemonSet 形式的 Mecha，由于只启动一次，可能会在 Serverless 的场景下大幅缩短 Serverless 函数的执行时间。对此 dapr 项目也有相关的[讨论](https://github.com/dapr/dapr/issues/5385)。

就像今年 Cilium 发布支持 Service Mesh 能力的办法，通过 eBPF 在内核态实现 L3 L4 层能力，而对应的 L7 层能力则交给用户态的 Envoy 处理这种将问题一分为二的思想，也许多运行时架构的未来方案也可能是折中或是多种方式结合的。例如采用在 Node 上按 Service Account 或 Namespace 运行多实例，或是轻量级 Sidecar 做协议转换＋DaemonSet 做流量管理和网络调用。

当然 DaemonSet 也有其固有的缺陷，资源被共享从而降低消耗的同时，故障也被共享了，而且故障产生的伤害面也变大了，此外还会导致 DaemonSet 被应用使用的争抢问题，以及应用之间的数据暴露风险。到底后续将会如何演进，我们拭目以待。

### 定义抽象能力的（API）的困境

分布式能力抽象层，是对分布式场景下需求的抽象性定义，抽象作为一种共识，其要义就在于保留共性而排除个性。但实际当中会发现，同类型中间件的差异化恰恰体现在了一些高级的、细分的专有特性上，很多业务对中间件选型的原因也在于这些专有特性上。

这就引出了一个困境：抽象能力所覆盖的需求，其丰富程度与可移植性成反比。

{% asset_img 13.png 300 %}

就如上图所示，如果抽象能力范围只覆盖到红色的部分，则组件 ABC 的专有特性都无法被引入，而如果抽象能力范围覆盖到绿色，那么就无法迁移到组件 C。

Dapr 的 Building Blocks 中，State management 就存在这样的一个例子：

State management 定义了基于事务操作的能力 `/v1.0/state/<storename>/transaction`，支持 State management 能力的 Component 有很多，对于支持事务的中间件如 Redis 就一切正常，但有一些并不支持事务的如 DynamoDB，则这种能力就无法使用。

定义抽象能力的困境，本质上是一种对能力收敛的权衡，这种权衡可能是与具体的业务需要高度相关的。

关于如何降低专有特性对能力集合可移植性的冲击，敖小剑在他的文章[《死生之地不可不察：论API标准化对Dapr的重要性》](https://skyao.io/talk/202111-important-of-api-standardization-for-dapr/) 中提到了四种解决思路：

1. 在 Mecha 层弥补能力缺失

   如果缺失的能力支持用基础能力来间接实现，就可以在 Mecha 内做处理。例如对于不支持批量写入的基础设施，在 dapr 中通过 forloop 连续调用单次写入也能间接的弥补这一能力（虽然无法做到性能一致）。

   然而这样也可能导致 dapr 越来越臃肿，怎么权衡见仁见智。

2. 在 Component 层弥补能力缺失

   Component 作为某种具体基础设施与 dapr 的适配器，可以将 1 中的方案下沉到 Component 里面，避免 dapr 本身的臃肿，然而这种办法的缺陷在于每种基础设施只要想弥补缺失的能力，就都要分别在自己的 Component 中实现一遍。

3. 直接忽略某些缺失的能力

   例如在 State management 中对多副本强一致性的配置属性 consistency，假如实际的存储中间件是单副本架构，那么就可以直接忽略掉该属性。

4. 其余的情况，只能在业务侧处理

   就像前文提到的事务能力，对于不支持的基础设施必须要明确报错，否则可能导致业务不正确。这种情况就只能在业务侧做限制，本质上是侵入了业务层。

这四种解决思路从权衡与折中的角度，覆盖了绝大多数能力缺失的场景，本质上这些思路属于 **“坚守 API 能力交集”** 的办法。假如跳出“抽象共识”这一限制，我们是否可以试图构建出一套包含了所有分布式能力的“大全集”呢？显然只是理论可行，但不现实。

然而，在企业实际的场景下，这个“全集”的规模可能并不一定像我们想象的那么庞大，因此就有可能提供额外的一种思路，即对分布是抽象层进行扩展，将有限规模的“个性”全部包含进去，形成 **“并集”** 从而规避上述问题。

{% asset_img 19.png %}

蚂蚁 Layotto 的设计中体现了这种方案，详见下文。



## 蚂蚁金服的方案：layotto

蚂蚁金服作为 dapr 的早起使用者，在落地的过程中结合遇到的问题及业务思考，在 2021 年年中推出了自研的 Mecha 方案：layotto。

### 单一 Sidecar

{% asset_img 15.png %}

非常有趣的一点是，layotto 是以 MOSN 为基座的。MOSN 是蚂蚁金服自研的网络代理，可用于 Service Mesh 数据面。因此 layotto 类似于是 MOSN 的一个特殊的插件，向业务侧提供分布式能力抽象层，并且仍然以 Component 的形式封装各种中间件的访问与操作，而在这之下的所有网络层交互全部代理给 MOSN。

由于 layotto 在运行态上是与 MOSN 绑定在一个 Sidecar 内的，因此就减少了一部分前文提到的两个 Sidecar 之间通信的开销。当然 layotto 可以这样做也有一部分原因在于 MOSN 本身已经在蚂蚁内部大规模落地，同时蚂蚁也有足够的研发强度来支撑 layotto 的开发。

### “私有协议”与“可信协议”

Layotto 的开发者，在讨论多运行时架构以及 layotto 落地实践的[文章](https://www.infoq.cn/article/5n0ahsjzpdl3mtdahejx)中，尝试对可移植性的概念进行了扩展，将支撑分布式能力的协议划分为 “可信协议” 与 “私有协议”。

其中，可信协议指代的是一类影响力很大的协议如 Redis 协议、S3 协议、SQL 标准等。这一类协议由于用户众多，且被各类云厂商所支持，因此可以认为它们本身就具有可移植性。

私有协议则指代一些企业内部自研的、闭源或影响力小的开源软件提供的协议。显然这一类协议才更需要考虑抽象与可移植性。

因此实际上的所谓分布式能力抽象层可能会是如下图所示的样子：

{% asset_img 16.webp 500 %}

各类可信协议不再二次抽象，而是直接支持（通过 MOSN 底座），对其余的私有协议再进行抽象。这种直接支持开源协议的思路，部分缓解了定义抽象能力的困境问题。

### 灵活的扩展模型

前文提到的 API 扩展形成 “并集”，Layotto 通过提供 In-Tree 形式的私有 API 注册点，实现了不修改 Layotto 代码就能扩展 API 能力：

![](https://user-images.githubusercontent.com/26001097/131614952-ccfc7d11-d376-41b0-b16c-2f17bfd2c9fc.png)

从代码角度看，Layotto 是通过暴露 API 注册钩子，暴露启动入口，来允许用户自行扩展代码，之后再调用启动函数启动进程。这样扩展 API 代码与 Layotto package 级隔离，但编译后可形成同一个二进制文件。

另外，通过 MOSN 的 WASM 插件能力，Layotto 也支持通过 WASM 镜像来扩展 API Filter。

## 未来展望

虽然多运行时架构这种理念从提出到现在只有两年，但已经很少有人会否认它所带来的价值，不论是 dapr 还是 layotto 的快速发展，都明确了头部企业对这一领域的投资逻辑。

当然目前从理论到实践可能都不够成熟，大家在落地实践的过程中也都会或多或少遇到前文提到的一些局限。但这些局限所处的层次大都是工程化、技术选择等具体的问题，相信随着各方技术的不断整合，实践的不断完善，问题都能解决。

对多运行时架构实践的未来，结合当下的限制、挑战以及趋势，我们也许能勾勒出某种未来可能的架构形态：

{% asset_img 17.png 500 %}

在这一架构形态下：

- 分布式能力抽象层提供标准能力抽象，以及灵活扩展的私有协议的能力
- 既成标准协议（对前文 "可信协议" 的另一种提法）作为 "既成的" 抽象能力，在Mecha 层只做协议转换或直接透传
- Mecha 与网络代理层进程级耦合，各类特性不再明确区分开发侧与运维侧
- 进程在 Node 上按租户/namespace 以及高可用要求划分多实例
- 接入现代化的可观测性体系，提升对故障的洞察分析能力，降低由于架构分层带来的问题诊断困难

总之，不管是架构形态怎么变、能力怎么抽象，让业务逻辑不断内聚，越来越面向接口、面向能力编程的趋势不会改变，服务化体系的未来值得期待。



## Reference

1. [Multi-Runtime Microservices Architecture](https://www.infoq.com/articles/multi-runtime-microservice-architecture/)
2. [Dapr](https://dapr.io/)
3. [死生之地不可不察：论API标准化对Dapr的重要性](https://skyao.io/talk/202111-important-of-api-standardization-for-dapr/) 
4. [Layotto](https://mosn.io/layotto/#/zh/README)
5. [如何看待 Dapr、Layotto 这种多运行时架构？](https://www.infoq.cn/article/5n0ahsjzpdl3mtdahejx)
