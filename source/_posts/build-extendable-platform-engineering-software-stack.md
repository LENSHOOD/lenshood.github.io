---
title: 构建可扩展的平台工程软件栈架构
date: 2023-05-09 22:08:49
tags: 
- multi-runtime
- platform engineering
- multi-cluster
- multi-cloud
categories:
- Software Engineering
---

{% asset_img head-pic.jpg 500 %}

本文介绍了企业在构建平台工程能力时如何通过可扩展的软件栈架构来满足多样的场景与诉求。



<!-- more -->

## 平台工程是新概念吗？

[平台工程（Platform Engineering）](https://platformengineering.org/blog/what-is-platform-engineering)，是通过在企业内部建设一整套工具链、资源和知识库，来更好的为研发人员构建工作流，以使研发团队在 DevOps 实践上更易于达成 [DORA Metrics](https://cloud.google.com/blog/products/devops-sre/using-the-four-keys-to-measure-your-devops-performance) 所描绘的高效能团队要求。企业通过设立小规模的[平台工程团队](https://www.thoughtworks.com/zh-cn/radar/techniques/platform-engineering-product-teams)，建设 IDP 内部开发者平台（Internal Developer Platforms），使研发人员能够方便灵活的对业务应用进行开发和交付，而不必将大量时间花费在学习和使用交付工作流所依赖的各种底层设施上。

实际上，很多人在了解到平台工程这一概念后，都会心生疑惑：“平台工程是个新概念吗，听着耳熟啊？”，“这不就是我们公司的 xxx 平台吗？ （这里的 xxx 平台可能是 **研发管理平台** / **DevOps 平台** / **开发者平台** / **一站式运维平台** 等等”。

的确，至少在笔者所经历过的雇主、客户等诸多场景中，就接触过（甚至咨询过）不下 5 个类似的平台，虽然它们的名称、内部代号各不相同，但至少都会提供包括 CI/CD 流水线、配置中心、制品库、资源管理、APM等基础能力。通过这些能力，业务研发团队实现了一定程度的交付自动化。

那有了这些基础能力，是否意味着企业的平台工程就建成了呢？并不是。

平台工程是为了让企业更易于达到高成熟度的 DevOps 实践而诞生的，但许多企业没有精力和人才来构建整个 DevOps 体系（工具、流程、文化甚至组织变革），就只好先从便宜且收效快的地方入手了：从运维团队抽调几个人，搭一套开源方案就能凑合用。但这些零散的工具和脚本除了提供一定的自动化能力外，并不能让研发人员从 DevOps 的工作中解放出来（[《“扯淡的DevOps，我们开发者根本不想做运维！”》](https://mp.weixin.qq.com/s/ZLIdcZOAAKHRl2KvRsxkGA)）。

虽说 DevOps 精神提倡 “You build it, you run it”，然而普通开发人员玩不转或者根本不想考虑代码到底如何才能交付上线。而由于这些工作很可能最终会落到更资深的人员头上，导致他们的时间也被明显侵占了。更头疼的是随着云原生的不断发展，各类新的概念和实践被提出，研发人员愈发的感觉到知识追赶的压力，也就越来越抵触做 DevOps 相关的事情。

我们会观察到，真正具备了高 DevOps 成熟度的企业，一般能看到如下两种不同的实践：

1. 团队自组织的 DevOps 实践：完全开放的独立式团队，由成熟的研发人员组成，践行敏捷文化，根据实际情况自建或选择工具链（如 Github Actions），通过 IaC 确保流程和基础设施全部代码化管理，自主选择构建配置管理、Key 管理、可观测性等能力的方案。
2. 企业统一构建成熟的内部开发者平台：通过在企业内部搭建非常完善的平台、工具甚至基础设施，推行全公司一致的研发流程和团队管理实践，并通过构建复杂的指标体系来评估不同的研发团队并与绩效挂钩，自上而下的推动 DevOps 能力提升。

上述两种不同的实践中，第一类更适合分散作战的小型精英团队，能快速且成本可控的交付软件。而对于第二类实践，其建设投入非常可观，更适合大型企业。事实上很多大型企业自建的内部开发者平台，已经和平台工程的概念内容逐步趋同了。

因此平台工程是个新概念吗？是也不是，它的一部分也许早已融入了研发流程中，而缺少的那部分则能够真正解放研发人员的双手。



## 平台工程能力的演进

前面提到，许多企业都可能期望通过引入一些工具或构建平台来弥补其 DevOps 能力的欠缺，不同的投入水平会得到不同的效果。而大型企业在经过持续的投入和演进后，也已经构建了趋近于平台工程概念所描绘的 IDP。

在 CNCF App Delivery TAG 发布的 [*平台工程白皮书 Platforms White Paper*](https://tag-app-delivery.cncf.io/whitepapers/platforms/) 中，对企业平台工程能力的成熟度做了总结（成熟度由低到高）：

>1. 产品开发人员可以根据需要配置相关能力，并立即使用这些能力来运行系统，如计算、存储、数据库或身份认证。
>2. 产品开发人员可以根据需要配置服务空间，并在服务空间中运行流水线和任务，来保存制品和配置，以及收集遥测数据。
>3. 第三方软件的管理员可以按需配置依赖，如数据库等，并轻松安装和运行该依赖软件。
>4. 产品开发人员可以通过模板（templates）配置完整的环境，这些模板结合了专用场景（如web开发或MLOps）下所需的运行时（run-time）和开发时（development-time）服务。
>5. 产品开发人员和管理人员可以通过自动数据采集和标准看板来观测已部署服务的功能、性能和成本。

可以看出，按成熟度模型的划分，IDP 所能提供的能力从基础到高级可归纳为*「基础设施 -> CICD -> 自助 PaaS -> 模板自动化 -> 可观测」* 。当然，以上的一切都是云原生的。

纵然在这篇 [*何为平台工程*](https://platformengineering.org/blog/what-is-platform-engineering) 中提到，只要组织规模达到 20~30 人，就可以开始关注 IDP 了。但小规模企业对平台工程能力建设的主要考量点可能还是**基本可用**和**成本最低**。通过尽量使用便宜的甚至免费的服务来满足 DevOps 的要求就是最好的选择，因此小规模企业可能更容易接受 SaaS 化的解决方案，例如云提供商打包出售的开发者服务，或者类似 [Github Pro](https://github.com/pricing)、[Coding](https://coding.net/) 等方案。

而当企业达到一定规模，并且处于高速发展期时，其诉求可能变化为更关注软件交付的速度，以及 IDP 的快速扩展能力。毕竟高速发展的业务需要平台成支撑其快速试错、快速占领市场的目标，同时大量扩张的业务和用户量也要求 IDP 能更快、更简单的扩展新功能以支持业务发展。

大规模企业在平台工程能力的建设上，就更倾向于要求易用、合规、稳定、并且能降本增效。为了支撑大型企业中各式各样的业务需求，IDP 在用户交互上需要简明易用，降低学习门槛。而由于存在跨国跨地区的业务，以及内部用户众多，合规性和稳定性的要求也非常重要，IDP 可以没有存在感，但不能因为 IDP 而导致企业发生损失。最后，大规模企业流程多、效率差，降本增效是持续的主题，IDP 需要尽可能的自助化、傻瓜化从而降低人员使用成本，提升效率。

据此可见，企业对平台工程能力建设的诉求，在不同规模下不同，在不同时期下也不同。那么在设计 IDP 的整体架构时，如何才能满足不同的诉求呢？下文我们将介绍基于标准抽象构建的模块化软件栈架构来设计平台，以充分释放可扩展性。



## 可扩展的平台工程软件栈架构

软件栈（或称[解决方案栈](https://en.wikipedia.org/wiki/Solution_stack)）是通过一组软件子系统或组件来构建的一个完整平台，在该平台之上可以运行特定的应用程序。这种基于抽象分层的栈模式在软件系统中十分常见。

对于前面所提到的构建平台工程能力过程中，将企业需要自研的部分尽可能收敛和内聚，通过设计标准抽象层，并尽可能借助开源社区的力量，构建分层的、可扩展的、组件化的软件栈架构，就能够充分应对建设平台能力过程中在成本和功能上多变的场景化诉求。

基于上述思想构建的可扩展的平台工程架构，如下图所示：

{% asset_img arch.jpg %}

设计自上而下的抽象层，在层内组合多种开源组件，就可以搭建起灵活可扩展的平台工程架构。

### 可扩展的能力抽象层

作为平台架构的最上层，能力抽象层最核心的目标是通过提供合理的抽象与接口，让应用在开发态和运行态都能更关注于业务本身，而将运维能力侧和公共组件侧的需求尽可能代理出去。通过能力抽象层提供的各种抽象，应用研发团队只需要对应用所需公共能力和运维的需求进行定义，之后即可放心的将配置和实施的工作交给平台，从而节省大量原先需要做的 DevOps 工作和时间。

因此，能力抽象层的设计，在应用运维需求侧，考虑构建标准化可扩展的应用运维模型（见如下“统一应用模型”），而在公共能力的需求侧，考虑构建通用的公共能力抽象（见如下“分布式能力抽象”）。

#### 统一应用模型

通常意义上，应用在 Day2 阶段的持续时间要远大于 Day0 和 Day1，这也就导致了交付和运维的工作是繁杂和冗长的。软件部署早已不是把包丢到服务器上然后启动进程就完事，还包括配置管理、服务拓扑、副本扩缩、流量分发、监控、审计、成本优化等等各种要求。

为了满足这些要求，需要通过各种工具和手段来实现，而正如前文提到的，很多开发者可能并不想去了解这些玩意儿到底如何使用。毕竟作为应用开发者，他们可能更清楚自己软件的入口是`/index`，而不清楚从主域名经过几层 Nginx 才能路由到到`/index`。

通过定义一套标准的应用交付模型，就可以将应用开发团队和平台团队的关注点进行分离，由开发者定义特定应用在交付和运行中的需求，由平台团队来实现这些需求。

[OAM（Open Application Model](https://github.com/oam-dev/spec/blob/master/README.md)就是这样的一种模型标准。

<img src="https://github.com/oam-dev/spec/blob/master/assets/overview.png?raw=true" style="zoom: 50%;" />

平台团队提供 ComponentDefinition 来描述不同应用的部署模型，如`通过 K8s Deployment 部署的无状态后端服务`，`直接推送 CDN 的静态前端页面` 等。应用开发者只需挑选某个 Component 来描述应用，并设置一些属性参数，如镜像名、ENV、端口号等等即可。

同时，平台团队还提供了对 Traits 和 Scopes 的定义，这允许开发者为他们的应用添加运维特征如动态扩缩，灰度发布，负载均衡等，以及分组特征如安全组，AZ等。

因此，开发者只需要将应用模型以类似 `yaml` 的形式维护在代码仓内，随着 [GitOps](https://www.weave.works/blog/what-is-gitops-really) 流程，应用就会顺滑的交付上线。

通过 OAM 的抽象隔离，对应用团队提出的各种运维诉求，平台团队可以灵活的采用各种手段来实现，且实现方案能放心的替换和扩展。

> [KubeVela]([kubevela.io](https://kubevela.io/)) 实现并扩展了 OAM

#### 分布式能力抽象

在成规模的服务化体系中，应用可能依赖了越来越多的中间件以及三方服务，它们提供了应用实现业务目标所需要的各种分布式能力。然而，传统的 SDK 集成方式让这些分布式能力变成了一个个的孤岛，难以统一治理，导致维护困难、升级困难，降低了整体的研发效率。

平台可以将各种分布式能力进行归类和抽象，为应用提供统一的分布式能力抽象层，因而应用只需要通过抽象层调用标准化能力，由平台团队维护实际的能力实现组件。

[多运行时架构](https://www.lenshood.dev/2022/12/06/multi-runtime/)就是基于上述思想提出的解决方案：

在多运行时架构的理念中，将各种分布式能力归纳为 4 个部分：生命周期、网络、状态以及绑定。传统场景下，这四大能力是由各类基础设施和中间件来提供的。

<img src="https://www.lenshood.dev/2022/12/06/multi-runtime/7.png" style="zoom:50%;" />

通过`Mircologic + Mecha`，即微业务与所谓“机甲”相组合的方式，将业务对分布式能力的需求全部交给 Mecha 运行时来代理，而真实提供分布式能力的组件，通过 Mecha 与业务应用隔离。

平台通过为每一个业务应用提供 Mecha 运行时，隔离需求与实现，因此能够方便的对各种中间件进行维护和扩展。

> 多运行时架构的实现方案有 [Dapr](https://dapr.io/)、[Layotto](https://mosn.io/layotto/#/zh/README) 等



###可扩展的 DevOps 组件层

得益于能力抽象层对应用需求的抽象，在 DevOps 组件层，平台团队可以放心大胆的提供和尝试各类工具和实践。同时，在丰富的 DevOps 组件生态中，也形成了许多标准化协议和流程，让各类组件本身也能灵活扩展。

云原生不仅让应用获得了高度灵活的资源利用和弹性能力，也为基础设施侧组件提供了标准化的运行环境和管理方法，因此 DevOps 组件层提供的组件也将默认采用云原生的方案。

#### CI/CD

DevOps 组件层所提供的最基本的能力应该就是 CI/CD Pipeline 了。通过 Pipeline 来拉取 Repo 中的代码，并一步步的执行编译、检查、测试、部署、发布，这就是所有企业尝试 DevOps 工具的第一步。

CI/CD Pipeline 的运行，本质上是执行了一个 DAG，各个阶段具体做的事情只是挂载在 DAG 节点上的细节。因此大多数的 CI/CD 工具实际上都是一个 Workflow Engine。

可以基于 DAG 来定义标准的 Pipeline 抽象，在不同的抽象工作节点上执行如编译、检查、运行测试、打包等任务，具体的工作节点实现可自定义方案。

上一节提到的实现了 OAM 的 KubeVela，就基于 OAM 模型扩展了 Workflow 的定义，通过 Workflow 扩展定义，可以将 Pipeline 的部分直接集成在 OAM 模型当中，通过[workflow 插件](https://github.com/kubevela/workflow/tree/main)支持了部分预定义步骤的执行（如编译镜像等）。KubeVela 对 CI/CD Pipeline 的抽象称为 [Unified Declarative CI/CD](https://kubevela.io/docs/tutorials/s2i)。

<img src="https://camo.githubusercontent.com/3c9f24e500b84bc31c744c623573b11e00247f49364fe7bc673e03faa56ee631/68747470733a2f2f7374617469632e6b75626576656c612e6e65742f696d616765732f312e362f776f726b666c6f772d617263682e706e67" style="zoom: 50%;" />

#### 资源管理 + IaC

正如前文在分布式能力抽象中所描述的，业务应用通常会依赖大量由 CSP 提供的 PaaS 中间件服务。通过公共能力抽象，应用开发者能够不必关心具体中间件的实现和维护，但对于平台团队而言，大多数 PaaS 服务都是非 K8s 环境的，企业可能会维护大量 IaC 脚本来实现对资源的自动化操作，但想要将它们纳入 K8s 下统一管理仍需要可观的人力付出，尤其是在跨云场景下更加复杂。

[Crossplane](https://www.crossplane.io/) 正是为了解决这一问题而诞生。

<img src="https://docs.crossplane.io/media/composition-how-it-works.svg" style="zoom: 50%;" />

Crossplane 通过 Provider 实现对具体 PaaS 资源的操作，在其[官方市场](https://marketplace.upbound.io/)中已经有数十家 CSP 开发的 Providers。Crossplane 允许用户自己定义资源，并通过标准的控制器模式来完成对资源的管理。因此，通过 Crossplane 可以声明式的管理 PaaS 资源。

实际的资源申请场景中，Crossplane 借鉴了 K8s PV 与 PVC 的概念，资源提供方通过创建资源定义，来发布可用的资源，而资源使用方通过构建 Claim 来发出对资源的请求。最后，通过 Crossplane 控制器就能完成这一需求匹配过程。

通过类似 Crossplane 的技术，可以将各种孤立的 PaaS 中间件与企业平台结合起来，并实现代码化、自动化管理。

#### 可观测性

常见的可观测性能力通过 Metrics、Logs 以及 Tracing 来分别采集系统的指标、日志和链路追踪。

在 [OpenTelemetry](https://opentelemetry.io/) 出现以前，上述三种不同种类的观测数据各自存在特定的探针、数据格式以及标准，从而导致后端系统的选择绑定了几种方案而难以替换和扩展。

OpenTelemetry（简称 otel） 作为可观测性系统前后端之间的抽象层，整合了一套标准数据模型，使得数据探针和数据处理系统不在相互依赖。

<img src="https://opentelemetry.io/img/otel_diagram.png" style="zoom:67%;" />

除了各种开源方案对 otel 的支持外，包括 AWS、Azure 和 GCP 在内的许多 CSP 都在其产品内支持了 otel 标准。因此引入 otel 能够极大的增强在可观测性能力上的扩展性。



### 可扩展的资源编排层

根据 [CNCF 对云原生的定义](https://github.com/cncf/toc/blob/main/DEFINITION.md)，云原生应用应该能在云上自由的弹性扩展。从这一点看，理想的云原生基础设施，应该能为应用提供无限的资源和难以察觉的扩缩速度，虽然这并不现实，但平台也有义务为上层提供灵活的、开箱即用的、标准化的生命周期管理与资源编排能力，而包括平台本身在内的上层用户，可以自由的按需使用资源，而不需要考虑资源的管理、分配细节。

上层用户可能对资源提出各类场景化诉求，如高可用、租户隔离、法律合规、使用成本等。因此资源编排层的职责就是通过对差异化资源的灵活组合，叠加精准的编排调度能力，为上层用户提供匹配其诉求的抽象资源。

为了支撑上述需求，资源编排层需要实现三大能力：应用灵活调度、标准资源抽象、基础设施动态扩缩。

#### 跨集群动态调度

多集群本质上是为了让应用运行在更合适的位置。为了高可用、成本控制、合规性等目的，业务应用可能会被动态调度到不同的 K8s 集群上。实施调度的关键是调度策略。

通过调度策略，我们期望能解决 ”什么样的应用” 需要被调度到 “哪类集群” 的问题。显然，应用有其自身独特的属性集，集群也一样。从属性集的角度看，调度策略问题就可以转化为应用与集群属性集之间的最优匹配问题。

<img src="https://www.lenshood.dev/2023/03/09/k8s-multi-cluster-1/sched-chain.jpg" style="zoom:50%;" />

应用的属性集可能包括：命名空间、资源依赖、副本数、镜像名、租户归属、应用亲和性/反亲和性、最小资源需求等等，集群的属性集可能包括 AZ、地区（Region）、节点数、已分配 Pod 数、资源总量/余量、污点（Taint）等等。

进行调度决策时，待调度应用与待选集群的属性集依次通过所有过滤型决策器和打分型决策器，最终找到一个（或一组，考虑多副本高可用）分数最高的集群，调度完成。而下达调度决策的前提，是多集群控制面能准确的获悉集群中的各种状态，因此状态数据的收集也至关重要。

>  [Karmada](https://karmada.io/zh/) 是一种多集群解决方案，它实现了动态调度

#### 应用模型扩展

为了满足多集群管理的需求，传统的单集群应用模型需要进行扩展，以允许应用开发者能定义应用的部署偏好，从而更符合实际的业务目标。

对应用模型的扩展可以分为规格扩展和状态扩展。

对于规格扩展，又可分为两类：限制（constrains）和提示（hints）：

- 限制（constrains）：代表了应用对跨集群管理的强制性要求，如亲和性/反亲和性，最小副本数，污点容忍性（Taints Tolerations）等等
- 提示（hints）：代表了对多集群管理决策与动作的非强制性提示，如优先级，副本分配偏好，资源需求等等

而状态扩展主要扩展的是应用在多个集群上的状态。这包括应用实际在每个集群上的副本数，运行健康状况，曾经被调度的历史等等。状态扩展恰恰是为了将收集到的状态数据聚合在应用模型上，以便于调度器的工作。

根据前文描述的 OAM 模型，平台团队可以针对限制和提示定义多种 Traits，以供应用开发者选用。

#### 集群标准化

基于公版 K8s 扩展和改造的 K8s 发行版层出不穷，这包括各类 CSP 提供的云托管服务（如 EKS），也包括许多开源的方案（如 K3s），发行版 K8s 本质上是在不同场景下提供更合适的容器平台方案。

K8s 社区一方面鼓励定制版的研发以提升社区生态的健康发展，另一方面，CNCF 提出了 [K8s一致性认证](https://www.cncf.io/certification/software-conformance/) 用来确保所有发行版在 API 层面与开源公版兼容，从而在一定程度上达成约束：可以相对安全的在通过认证的发行版之间切换。

但即使是一致性认证的存在，企业在选择 K8s 时也仍然可能面临多集群之间版本一致性、版本升级的困难。

[Gardener](https://gardener.cloud/) 是 SAP 开源的 K8s 多集群解决方案，在设计概念上，Gardener 期望作为一种管理大量 K8s 集群的组件，借助 Gardner，用户能够实现在接入各种不同类型底层基础设施的同时，方便的在其上构建标准的 K8s 集群。

![](https://gardener.cloud/__resources/gardener-architecture-overview_2bd462.png)

Gardener 除了能在各种差异化基础设施上管理 K8s 集群的生命周期，还能够确保在这些基础设施上运行的 K8s 集群具有完全相同的版本、配置和行为，这能简化应用的多云迁移。Gardener 还提供了[一个页面](https://k8s-testgrid.appspot.com/conformance-all)专门介绍其不同版本的标准化K8s 集群与不同云提供商的兼容情况。

除此之外，许多 2B 企业在销售其产品时，也会遇到客户基础设施中的各种 “魔改” 版本（显然不会通过一致性测试）。针对这类场景下的产品兼容性需求，企业一方面会尽量收敛产品对 K8s API 的依赖，同时也会自研类似代理层模块，来屏蔽五花八门的 “魔改” 版本带来的冲击。



### 可扩展的基础设施层

在可扩展的资源编排层中，抽象资源通过 K8s 以标准的 CRI、CSI、CNI 形式提供给上层，这给了基础设施层很大的灵活度和扩展性。基础设施层可以自由的通过多云和混合云等技术来根据实际需要向上提供底层的计算、存储、网络资源。

#### 集群即资源

K8s 是对基础设施层的抽象，因此从整体上看，基础设施层的目标就是向资源编排层交付 K8s 集群。然而不论是基于云虚拟机自建 K8s，还是直接使用 CSP 定制化和代管的集群（如 EKS，GKE 等），在各种 CSP 上构建 K8s 集群都涉及到许多适配性工作。

[ClusterAPI（简称 CAPI）](https://github.com/kubernetes-sigs/cluster-api) 是 K8s “Cluster Lifecycle SIG（集群生命周期特别兴趣小组）” 发起的项目，CAPI 尝试通过定义标准基础设施 API 来统一集群生命周期管理，各类 CSP 自行提供实现了标准 API 的 “Provider” 来支持自动化操作集群资源，由于其官方背景，目前已有[数十种 Provider](https://cluster-api.sigs.k8s.io/reference/providers.html) 可供选择（不仅包含 aws 等公有云，还包含了 OpenStack，OCI 等其他方案）。

<img src="https://cluster-api.sigs.k8s.io/images/management-cluster.svg" alt="CAPI 架构" style="zoom:67%;" />

CAPI 的价值不仅在于对各种 CSP 的全面适配，更重要的是通过它能够实现集群的自动化创建和销毁，也就是实现了**集群即资源**，从极大的缩短了资源编排层的扩展速度。

#### 跨云网络

通常，CSP 会默认提供 VPC 来实现灵活的隔离，而跨云之间更是完全隔离，想要互相通信只能通过专线或公网路由。

在多云架构下或多或少会存在跨云网络连通需求，采用 Overlay 网络的方式，可以在不同的云或 VPC 之间建立虚拟扁平网络，实现直接网络连通。

![](https://www.lenshood.dev/2023/03/09/k8s-multi-cluster-1/network-overlay.jpg)

Overlay 网络的本质是隧道技术，通常是在三层网络上构建隧道传输二层网络包来实现虚拟网络。Overlay 网络的优势在于它构建的虚拟扁平网络让上层通信不再依赖复杂的路由策略，但类似 [VxLan](https://en.wikipedia.org/wiki/Virtual_Extensible_LAN) 的技术，只对网络数据包进行了再封装，并没有任何安全性可言，因此当用于公网间建立隧道的时候会采用加密协议传输，如 [IPSec](https://en.wikipedia.org/wiki/IPsec)，[WireGuard](https://www.wireguard.com/) 等。

> 相关开源组件有 [Kilo](https://kilo.squat.ai/)、[Submariner](https://submariner.io/) 等



## 总结

本文首先讨论了企业在建设平台工程能力过程中存在差异化的场景和诉求，基于这一原因，我们期望通过构建可扩展的软件栈架构，使企业能够尽可能的借助开源社区的力量，在成本可控的前提下建设自己的内部平台。

可扩展的前提是足够的抽象，通过划分四个抽象层，我们能够在应用能力诉求、DevOps 组件、应用编排调度和基础设施资源这四个维度上分别实现灵活可扩展。叠加云原生生态中的丰富开源方案，组合起来，就构成了完整的软件栈。