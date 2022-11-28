---
title: 可观察性成熟度模型
date: 2022-11-27 22:06:31
tags:
- observability
- SRE
- ops
categories:
- Software Engineering
---

> 本文是对 StackState 发布的 [*The Observability Maturity Model*](https://www.stackstate.com/white-paper/observability-maturity-model/) 的中文翻译



## 序言

At StackState, we have spent eight years in the monitoring and observability space. During this time, we have spoken with countless DevOps engineers, architects, SREs, heads of IT operations and CTOs, and we have heard the same struggles over and over. 

在StackState，我们在监控和可观察性领域已经花了八年时间。在这段时间里，我们与无数的DevOps工程师、架构师、SRE、IT运营主管和CTO交谈过，我们反复听到同样的挣扎。

Today’s consumers are used to great technology that works all the time. They have little tolerance for outages or performance issues. These expectations push businesses to stay competitive through frequent releases, ever-faster response and greater reliability. At the same time, the move towards cloudbased applications – with all of their ever-changing functions, microservices and containers – makes IT environments more complex and harder than ever to operate and monitor.

今天的消费者已经习惯了一直在工作的伟大技术。他们对故障或性能问题的容忍度很低。这些期望促使企业通过频繁的发布、更快的响应和更高的可靠性来保持竞争力。同时，向基于云的应用的转变--所有不断变化的功能、微服务和容器--使IT环境变得更加复杂，比以往更难操作和监控。

As a result, we have seen great commonalities in the monitoring challenges that are unfolding globally, such as this colorful issue described by a customer:

因此，我们看到在全球范围内展开的监测挑战有很大的共性，例如一位客户描述的这个丰富多彩的问题。

> “When something big broke in the infrastructure, storage, networking equipment or something like that... every time we saw the same movie. The monitoring gets red, red, red, thousands of alarms, nobody knows what’s the root cause. Everybody is panicked – real total chaos.”
>
> "当基础设施、存储、网络设备或类似的东西出现大故障时......每次我们都看到同样的电影。监控变得红色，红色，红色，成千上万的警报，没有人知道什么是根本原因。每个人都很惊慌--真正的完全混乱"。
>
> \- Georg Höllebauer, Enterprise Metrics Architect at APA-Tech
>
> \- Georg Höllebauer, APA-Tech的企业指标架构师



I witnessed this problem first-hand eight years ago when I was part of a team of two consultants working at a major Dutch bank, helping them improve the reliability of their mission-critical applications. They were a mature enterprise with multiple monitoring tools in place for their complex environment, but they could not find the root cause of problems quickly. As a result of many siloed tools and lack of unified view of their IT environment, customer experience was directly suffering. When something broke, it took too long to find and fix the core problem. We knew we had to find a better way, and the technology we built to meet this bank’s needs became the foundation for StackState.

八年前，我亲眼目睹了这个问题，当时我是一个由两名顾问组成的团队的一员，在一家大型荷兰银行工作，帮助他们提高其关键任务应用程序的可靠性。他们是一个成熟的企业，为其复杂的环境配备了多种监控工具，但他们无法迅速找到问题的根源。由于许多孤立的工具和缺乏对其IT环境的统一看法，客户体验直接受到影响。当有东西损坏时，要花很长时间才能找到并解决核心问题。我们知道我们必须找到一个更好的方法，而我们为满足这家银行的需求而建立的技术成为StackState的基础。

Since we released the original Monitoring Maturity Model in 2017, it has become clear that the original monitoring tools – which simply notified IT teams when something broke – were no longer sufficient for many other organizations as well. Today’s engineers need to immediately understand the priorities and context surrounding a problem: what’s the impact on customer experience and business results? Then, if the impact is high: why did it break and how do we fix it?

自从我们在2017年发布最初的监控成熟度模型以来，很明显，最初的监控工具--只是在出现故障时通知IT团队--对于许多其他组织也不再足够。今天的工程师需要立即了解一个问题的优先级和背景：对客户体验和业务成果的影响是什么？然后，如果影响很大：为什么会发生故障，我们该如何修复它？

The concept of observability has evolved from monitoring to answer those questions. Observability is vital in maintaining the level of service reliability needed for business success. Unfortunately, navigating the monitoring and observability space is hard, especially as AIOps enters the picture. Many vendors are making a lot of noise in the market and new open source projects are popping up left and right. It’s hard to know who really does what, and even harder to know which capabilities really matter.

可观察性的概念是从监测中发展出来的，以回答这些问题。可观察性对于保持业务成功所需的服务可靠性水平至关重要。不幸的是，在监控和可观察性的空间里航行是很困难的，特别是当AIOps进入画面时。许多供应商在市场上大肆宣传，新的开源项目也层出不穷。很难知道谁真正做了什么，更难知道哪些功能真正重要。

The Observability Maturity Model is based on extensive experience with real problems in live environments, discussions with customers and prospects, research into the latest technologies and conversations with leading analyst firms such as Gartner. We hope it will help you shine some light in the darkness. Our goal is not to present you with the perfect model of what your observability journey should look like. We know it doesn’t work like that. To quote [a famous British statistician](https://www.lacan.upc.edu/admoreWeb/2018/05/all-models-are-wrong-but-some-are-useful-george-e-p-box/), “All models are wrong, some are useful.” Rather, we wrote this Observability Maturity Model to help you identify where you are on the observability path, understand the road ahead and provide a map to help you find your way.

可观察性成熟度模型是基于对实际环境中真实问题的广泛经验、与客户和潜在客户的讨论、对最新技术的研究以及与Gartner等领先分析公司的对话。我们希望它能帮助你在黑暗中照亮一些光。我们的目标不是向你展示你的可观察性旅程应该是什么样子的完美模型。我们知道它并不像那样工作。引用[一位著名的英国统计学家](https://www.lacan.upc.edu/admoreWeb/2018/05/all-models-are-wrong-but-some-are-useful-george-e-p-box/)的话，"所有的模型都是错的，有些是有用的"。相反，我们编写这个可观察性成熟度模型是为了帮助你确定你在可观察性道路上的位置，了解前面的道路，并提供一张地图来帮助你找到你的路。

May this model be useful to you on your journey!

愿这个模型在你的旅程中对你有用!

Lodewijk Bogaards 

Co-founder and Chief Technology Officer 

StackState



## 引言：为什么要采用可观察性成熟度模型？

Monitoring has been around for decades as a way for IT operations teams to gain insight into the availability and performance of their systems. To meet market demands, innovate faster and better support business objectives, IT organizations require a deeper and more precise understanding of what is happening across their technology environments. Getting this insight is not easy, as today’s infrastructure and applications span multiple technologies, use multiple architectures and are more dynamic, distributed and modular in nature. 

作为IT运营团队深入了解其系统的可用性和性能的一种方式，监控已经存在了几十年。为了满足市场需求，更快地创新和更好地支持业务目标，IT组织需要更深入和更精确地了解他们的技术环境中正在发生什么。获得这种洞察力并不容易，因为今天的基础设施和应用程序跨越多种技术，使用多种架构，并且在本质上更加动态、分布和模块化。

Change is also a way of life in IT and research shows 76% of problems are caused by changes.*[1]* In order to maintain reliability in the face of all these challenges, a company’s monitoring strategy must evolve to observability.

变化也是IT行业的一种生活方式，研究表明76%的问题是由变化引起的。*[1]*为了在所有这些挑战面前保持可靠性，公司的监控策略必须向可观察性发展。

Most enterprises find it difficult to find the right monitoring strategy to manage their environments reliably. Over 65% of enterprise organizations have more than 10 monitoring tools, often running as siloed solutions.*[2]* This segregated structure limits the ability of SRE and IT operations teams to detect, diagnose and address performance issues quickly. When issues occur, teams try to find the root cause by combining teams, processes and tools, or by manually piecing together siloed data fragments. This traditional approach to monitoring is time consuming and does not provide the insights needed to improve business outcomes. Troubleshooting is just too slow and your most crucial customer-facing systems may be down for hours, resulting in millions in lost revenue.

大多数企业发现很难找到正确的监控策略来可靠地管理其环境。超过65%的企业组织有超过10个监控工具，通常作为孤立的解决方案运行。*[2]*这种分离的结构限制了SRE和IT运营团队快速检测、诊断和解决性能问题的能力。当问题发生时，团队试图通过结合团队、流程和工具，或通过手动拼凑筒状的数据碎片来找到根本原因。这种传统的监控方法很耗时，而且不能提供改善业务成果所需的洞察力。排除故障的速度太慢，你最关键的面向客户的系统可能会瘫痪几个小时，导致数百万的收入损失。

> 66% of MTTR is spent on identifying change that is causing a problem.*[3]*
>
> 66%的MTTR用于识别造成问题的变化*[3]*。



The move to dynamic cloud, containers, microservices and serverless architectures, combined with the need to maintain hybrid environments and legacy systems of record, further exacerbates the need for more advanced capabilities.

向动态云、容器、微服务和无服务器架构的转移，加上维护混合环境和传统记录系统的需要，进一步加剧了对更先进能力的需求。

Observability practices have evolved to meet these needs, combining advances in monitoring with a more holistic approach that provides deeper insights and a more precise understanding of what is happening across technology environments. The Observability Maturity Model defines four distinct levels in the evolution of observability, as described in Table 1 on the following page.

为了满足这些需求，可观察性实践已经发展起来，将监测方面的进展与更全面的方法结合起来，对整个技术环境中发生的事情提供更深入的洞察力和更精确的理解。可观察性成熟度模型在可观察性的发展过程中定义了四个不同的级别，如下页表 1 所述。

> Cloud and container migrations are driving the need for greater observability maturity.
>
> 云和容器的迁移正在推动对更大的可观察性成熟度的需求。



| Level                                  | Goal                                                         | Functionality                                                |
| -------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| \1. Monitoring                         | Ensure that individual components are working as expected.确保个别部件按预期工作。 | • Tracks basic health of individual components in IT systems • Looks at events; triggers alerts and notifications • Tells you that something went wrong… but not what- 追踪IT系统中各个组件的基本健康状况 - 查看事件；触发警报和通知 - 告诉你出了问题......但不是什么问题 |
| \2. Observability                      | Determine why the system is not working.确定系统不工作的原因。 | • Gives insights into system behavior by observing its outputs • Focuses on results inferred from metrics, logs and traces, combined with existing monitoring data • Delivers baseline data to help investigate what went wrong and why- 通过观察系统的输出，深入了解系统的行为 - 注重从指标、日志和跟踪中推断出的结果，结合现有的监测数据 - 提供基线数据，帮助调查出错的原因。 |
| \3. Casual Observability               | Find the cause of the incident and determine its impact across the system.找到事件的原因并确定其对整个系统的影响。 | • Provides more comprehensive insights to help determine what caused a problem • Adds ability to track topology changes in the IT stack over time, building on Level 1 and Level 2 foundations • Generates extensive, correlated information that helps reduce time needed to identify what went wrong, why the issue occurr提供更全面的见解，以帮助确定问题的原因 - 在一级和二级基础上，增加追踪IT堆栈中随时间变化的拓扑结构的能力 - 生成广泛的相关信息，帮助减少确定出错的原因、问题发生的原因所需的时间。 |
| \4. Proactive Observability With AIOps | Analyze large volumes of data, automate responses and prevent anomalies from becoming problems.分析大量的数据，自动化反应，防止异常情况变成问题。 | • Uses AI and ML to find patterns in large volumes of data • Combines AI/ML with data from Levels 1-3 to provide the most comprehensive analysis across the stack • Detects anomalies early and gives sufficient warnings to prevent failures- 使用人工智能和ML在大量数据中寻找模式 - 将人工智能/ML与1-3级的数据结合起来，在整个堆栈中提供最全面的分析 - 尽早发现异常，并发出足够的警告以防止失败 |

Table 1: Defining the levels of observability maturity

表1: 界定可观察性成熟度的级别

Each level of observability builds on the foundation established in previous levels to add capabilities in capturing, tracking and analyzing data. The new functionality enables deeper observability at each stage, resulting in improved IT reliability and customer satisfaction, as shown in Figure 1 below. Although you can marginally improve results within a level by enhancing processes, most teams need to collect new types of data to advance to the next maturity level and realize greater benefits.

每个级别的可观察性都建立在前几个级别建立的基础上，以增加捕捉、跟踪和分析数据的能力。新的功能使每个阶段的可观察性更加深入，从而提高了IT可靠性和客户满意度，如下图1所示。尽管你可以通过加强流程来略微改善一个级别内的结果，但大多数团队需要收集新类型的数据来提升到下一个成熟度级别并实现更大的收益。

{% asset_img figure1.png %}

Figure 1: Observability maturity and how it affects IT reliability

图1：可观察性的成熟度以及它如何影响IT可靠性

The Observability Maturity Model is based on research and conversations with enterprises across industries and has been validated with other practitioners, analysts and thought leaders. It is designed to help you: 

可观察性成熟度模型是基于对各行业企业的研究和对话，并经过其他从业者、分析家和思想领袖的验证。它旨在帮助你。

- Understand different types of data and how monitoring and observability practices can help your organization collect actionable information. 
- 了解不同类型的数据以及监测和可观察性实践如何帮助你的组织收集可操作的信息。
- Understand the differences between monitoring, observability and AIOps. 
- 理解监控、可观察性和AIOps之间的区别。
- Evaluate your organization’s current level of maturity. 
- 评估你的组织目前的成熟度。
- Guide your team to a higher level of maturity. 
- 引导你的团队达到一个更高的成熟度。

Use this model to learn clear steps you can take to improve observability in your organization so you can ultimately deliver more reliable and resilient applications to your customers. 

使用这个模型，你可以学习明确的步骤来提高你的组织中的可观察性，这样你就可以最终向你的客户提供更可靠和有弹性的应用。



## Level 1: Monitoring

*Goal: Ensure that individual components are working as expected*

The first level, Monitoring, is not new to IT. A monitor tracks a specific parameter of an individual system component to make sure it stays within an acceptable range. If the value moves out of the range, the monitor triggers an action, such as an alert, state change, notification or warning.

第一个层次，监控，对IT行业来说并不陌生。监控器跟踪单个系统组件的特定参数，以确保它保持在一个可接受的范围内。如果数值超出了范围，监控器就会触发一个动作，如警报、状态改变、通知或警告。

With traditional monitoring, which often encompasses application performance monitoring (APM), infrastructure monitoring, API monitoring, network monitoring and various other domain-centric tooling, the use case is, “Notify me when something is not operating satisfactorily.” You can think of monitoring in terms of traffic light colors:

- The component is available and healthy (green) 
- The component is at risk (orange or yellow) 
- The component is broken (red)

传统的监控通常包括应用性能监控（APM）、基础设施监控、API监控、网络监控和其他各种以领域为中心的工具，其用例是："当某些东西运行得不理想时通知我。" 你可以用交通灯的颜色来考虑监控。

- 该组件是可用的和健康的（绿色） 
- 该组件处于危险之中（橙色或黄色） 
- 该组件已损坏（红色）

Monitoring looks at pre-defined sets of values with pre-defined sets of failure modes. It focuses on basic component-level parameters, such as availability, performance and capacity and generates events that report on the state of the monitored value. 

监测着眼于预先定义的数值集和预先定义的故障模式集。它专注于基本的组件级参数，如可用性、性能和容量，并产生报告监测值状态的事件。

**Events** are noteworthy changes in the IT environment. Though events may be purely informative, they often describe critical incidents that require action. Events may trigger **alerts or notifications** that arrive via various channels, such as email, chat, a mobile app or an incident management system. 

**事件**是IT环境中值得注意的变化。虽然事件可能是纯粹的信息，但它们往往描述了需要采取行动的关键事件。事件可能触发**警报或通知，通过各种渠道到达，如电子邮件、聊天、移动应用程序或事件管理系统。

As a first step towards observability, implement monitoring to get basic insights into the health and status of individual components and be notified when something breaks. Below, Table 2 gives an overview of the key capabilities for Level 1.

作为实现可观察性的第一步，实施监控以获得对各个组件的健康和状态的基本了解，并在出现故障时得到通知。下面，表2给出了第1级的关键能力概述。

| Level 1: Monitoring                                          |
| ------------------------------------------------------------ |
| Use basic traffic-light monitoring to understand the availability of the individual components that make up your IT services.使用基本的交通灯监控来了解构成你的IT服务的各个组件的可用性。 |
| **System Input** Events and component-level metrics (e.g., “API response time is higher than our SLO of five seconds”) 事件和组件级指标（例如，"API响应时间高于我们的SLO五秒"）。**System Output** Alerts or notifications (e.g., “order fulfillment service is down”)警报或通知（例如，"订单执行服务中断了"）。 |
| **What You Get** • Basic information such as the health status of a component — is it working? • Alerts and notifications when issues occur • Easiest way to get started; many open-source and SaaS solutions are available - 基本信息，如一个组件的健康状态--它在工作吗？- 当问题发生时发出警报和通知 - 最简单的入门方式；有许多开源和SaaS解决方案可用 |

Table 2: Level 1 summary



### Next Step: Observability 

Monitoring gives you limited insights into the state of the overall environment. It shows you individual component health but generally no information about the big picture. It tells you something is broken but not why, who to call, nor when and where the original problem started. 

监测让你对整个环境的状态有有限的了解。它向你显示单个组件的健康状况，但通常没有关于大局的信息。它告诉你有些东西坏了，但没有告诉你原因，也没有告诉你该找谁，更没有告诉你原始问题是什么时候和什么地方开始的。

Setting up and maintaining monitoring checks and notification channels requires a lot of manual work. At Level 1, you also need to do root cause analysis and impact analysis manually and you have a limited set of data. Investigating the sources of problems takes time. In addition, a single issue may cause storms of alerts from multiple components, causing further confusion and delays in pinpointing the root cause. 

设置和维护监控检查和通知渠道需要大量的手工工作。在第1级，你还需要手动进行根本原因分析和影响分析，而且你的数据集有限。调查问题的来源需要时间。此外，一个问题可能会引起来自多个组件的警报风暴，造成进一步的混乱和延迟，无法准确地找出根本原因。

While monitoring can detect a limited number of known types of failures, or “known unknowns,” Level 2, Observability, can help you discover unknown and unexpected failure modes, or “unknown unknowns.” As you move from Level 1 to Level 2, you will gain more in-depth information that provides a better understanding of the availability, performance and behavior of your services. 

虽然监控可以检测到有限的已知故障类型，或 "已知的未知数"，但第2级，可观察性，可以帮助你发现未知和意外的故障模式，或 "未知的未知数"。随着你从第1级到第2级，你将获得更深入的信息，对你的服务的可用性、性能和行为有更好的了解。

## Level 2: Observability 

*Goal: Determine why the system is not working*



To keep today’s complex and dynamic IT systems running reliably, you need to not only know what’s working (monitoring) but also understand why it’s not working (observability). 

为了保持当今复杂多变的IT系统的可靠运行，你不仅需要知道什么在工作（监控），还需要了解它为什么不工作（可观察性）。

Traditional monitoring tracks the basic health of a component or system. Observability evolved naturally to provide deeper insights into the behavior of a system over time. When something goes wrong and your team receives an alert, you need to quickly figure out, “What happened? Where, when, why and who do we call?” Observability data helps you answer these questions. At its full maturity (Level 4), observability provides all the data you need, in the proper context, to automatically detect and remediate issues and even to proactively identify and prevent them. 

传统的监测是跟踪一个部件或系统的基本健康状况。可观察性自然而然地发展起来，以提供对一个系统随时间变化的行为的更深入的洞察力。当出了问题，你的团队收到警报时，你需要迅速搞清楚："发生了什么？在哪里，什么时候，为什么，我们应该找谁？" 可观察性数据帮助你回答这些问题。在其完全成熟时（第4级），可观察性在适当的背景下提供你所需要的所有数据，以自动检测和补救问题，甚至主动识别和预防问题。

When an alert pops up, you look to understand the state of your system to find the problem’s source. At Level 2, observability typically delivers system insights by focusing on three critical types of telemetry data: **metrics**, **logs** and **traces**. *[4]* These three pillars of observability are collected from IT components such as microservices, applications and databases to provide an overall perspective into a system’s behavior. Each pillar gives a different type of information, as outlined in Table 3 below.

当警报弹出时，你要了解系统的状态以找到问题的根源。在第二级，可观察性通常通过关注三种关键类型的遥测数据来提供系统洞察力。**指标**，**日志**和**跟踪**。*[4]* 可观察性的这三个支柱是从IT组件（如微服务、应用程序和数据库）中收集的，以提供对系统行为的整体看法。每个支柱都提供不同类型的信息，如下表3所示。

| Pillar      | Definition                                                   |
| ----------- | ------------------------------------------------------------ |
| **Metrics** | Numerical measurements that help you understand the performance and status of services — for example, the famous four golden signals: latency, traffic, error rate and saturation.*[5]*帮助你了解服务性能和状态的数字测量--例如，著名的四大黄金信号：延迟、流量、错误率和饱和度*[5]*。 |
| **Logs**    | Time-stamped records of relevant events that happen in a system (e.g., transactions, warnings, errors), which help you understand a system’s behavior at a given point in time.对系统中发生的相关事件（如事务、警告、错误）的时间戳记录，这有助于你了解系统在某一特定时间点的行为。 |
| **Traces**  | Detailed snapshots showing how data flows through an application from end to end (e.g., a user request), which help troubleshoot performance and sometimes give code-level visibility into how your app performs.详细的快照显示数据如何从头到尾流经一个应用程序（例如，用户请求），这有助于排除性能故障，有时还能让人看到你的应用程序的代码级性能。 |

Table 3: Three pillars of observability 

表3：可观察性的三大支柱 

These three pillars, along with events and alerts, are typically plotted on dashboards so teams can easily keep track of important activities. Some observability tools provide out-of-the box dashboards that bring together these different types of data on one screen and allow you to deep-dive into them for further investigation. 

这三个支柱，连同事件和警报，通常被绘制在仪表盘上，这样团队就可以轻松地跟踪重要的活动。一些可观察性工具提供了开箱即用的仪表盘，将这些不同类型的数据集中在一个屏幕上，并允许你深入研究这些数据以进一步调查。

Level 2 data has much greater breadth and depth than Level 1, and it often involves some data consolidation across your environment into a single view. You may need to build additional dashboards if you want more insights, especially if your environment has multiple domains and you are using multiple monitoring tools.

二级数据比一级数据有更大的广度和深度，它通常涉及到将整个环境的一些数据整合到一个单一的视图中。如果你想获得更多的洞察力，你可能需要建立额外的仪表盘，特别是当你的环境有多个域，并且你使用多个监控工具时。


| Level 2: Observability                                       |
| ------------------------------------------------------------ |
| Observe the behavior of IT environments by capturing metrics, logs and traces in addition to events and health state.除了事件和健康状态外，还通过捕捉指标、日志和跟踪来观察IT环境的行为。 |
| **System Input** Level 1 inputs + comprehensive metrics, logs and traces第1级输入+综合指标、日志和追踪 **System Output** Level 1 outputs + comprehensive dashboards with graphs, gauges, flame charts, logs, etc.1级输出+综合仪表盘，包括图表、仪表、火焰图、日志等。 |
| **What You Get** • Deeper, broader and more holistic view of overall system health by collecting additional data from more sources, which better supports problem diagnosis • Ability to discover unknown failure modes in addition to known types of failures • Beneficial insights from individual types of data — e.g., traces help identify performance bottlenecks, metrics make excellent KPIs and logs can be used to find software defects通过从更多的来源收集额外的数据，更深入、更广泛、更全面地了解整个系统的健康状况，从而更好地支持问题诊断 - 除了已知的故障类型外，还能发现未知的故障模式 - 从个别类型的数据中获得有益的见解 - 例如，跟踪有助于识别性能瓶颈，指标是优秀的KPI，日志可用于发现软件缺陷 |

Table 4: Level 2 summary



The challenge then becomes how to resolve information from too many dashboards. At Level 2, you can infer suspected reasons for incidents by manually correlating data, but this approach often involves complex manual queries across systems. 

那么挑战就变成了如何解决来自太多仪表盘的信息。在第2级，你可以通过手动关联数据来推断可疑的事故原因，但这种方法往往涉及跨系统的复杂手动查询。

At Level 2, teams have not yet developed an automated way to unify and correlate the siloed data from various tools and domains, so it is still labor intensive and time consuming to pinpoint the root cause of an issue. Consequently, MTTD and MTTR are higher than they should be, customers are more adversely affected and more revenue is lost than at higher maturity levels. 

在第2级，团队还没有开发出一种自动化的方法来统一和关联来自不同工具和领域的孤立数据，所以要找出问题的根本原因仍然需要耗费大量的人力和时间。因此，与更高的成熟度相比，MTTD和MTTR要高一些，客户受到的不利影响更大，损失的收入也更多。

### Next Step: Causal Observability

Observability generates a huge amount of data and sorting out the meaningful information can be difficult. 

可观察性产生了大量的数据，整理出有意义的信息可能很困难。

At Level 2, your team is likely challenged by both data silos and volume, which cause inefficiencies in cross-domain and cross-team troubleshooting. 

在第二级，你的团队可能受到数据孤岛和数据量的挑战，这导致跨领域和跨团队的故障排除效率低下。

When something goes wrong, too many people get involved because nobody knows where the problem is, resulting in incident ping-pong and blame games. You may need to build ad hoc solutions to query multiple observability silos to troubleshoot a single issue. Creating these queries requires practitioners with development skills, knowledge of data structures and understanding of system architecture. 

当出现问题时，由于没有人知道问题出在哪里，所以有太多人参与进来，从而导致了事件的乒乓化和指责游戏。你可能需要建立专门的解决方案来查询多个可观察性筒仓，以解决单一问题。创建这些查询需要从业人员具备开发技能、数据结构知识和对系统架构的理解。

In addition, the telemetry-centric and siloed views typical in Level 2 often require substantial manual work to extract actionable insights. Setting up efficient dashboards can take considerable time and they require ongoing maintenance. Root cause analysis, impact analysis and alert noise reduction are important in maintaining a reliable and resilient stack, but these activities are challenging at this level. 

此外，第2级中典型的以遥测为中心的筒仓式视图往往需要大量的手工工作来提取可操作的洞察力。设置高效的仪表盘可能需要相当长的时间，而且需要持续的维护。根源分析、影响分析和减少警报噪音对于维护一个可靠和有弹性的堆栈非常重要，但这些活动在这个级别是具有挑战性的。

Note: Teams are increasingly adopting the OpenTelemetry standard to facilitate the capture of metrics, logs and traces. OpenTelemetry is extremely helpful to efficiently collect these types of data, but it was not designed to bridge silos, create better context for data or to analyze the data. 

注意：团队越来越多地采用OpenTelemetry标准，以促进指标、日志和跟踪的采集。OpenTelemetry对于有效地收集这些类型的数据是非常有帮助的，但它并不是为了弥合孤岛，为数据创造更好的背景或分析数据而设计的。

In order to move to Level 3 and understand how your observability data is related, you need to provide context for events, logs, metrics and traces across the data silos in your IT environment. At Level 3, Causal Observability, you get a precise map of the topology of your business processes, applications and infrastructure and you can track how it all changes over time. When something goes wrong, you can use this contextual data combined with automation to quickly determine the cause of an issue without having to manually wade through silos of uncorrelated data.

为了进入第三级，了解你的可观察性数据是如何关联的，你需要为IT环境中的事件、日志、指标和跨数据仓的追踪提供背景。在第三级，即因果观察能力，你可以得到你的业务流程、应用程序和基础设施的拓扑结构的精确地图，你可以跟踪它如何随时间变化。当出现问题时，你可以利用这些上下文数据与自动化相结合，快速确定问题的原因，而不必手动处理不相关的数据孤岛。

## Level 3: Causal Observability

*Goal: Find the cause of the incident and determine its impact across the system.* 



It’s not surprising that most failures are caused by a change somewhere in a system, such as a new code deployment, configuration change, auto-scaling activity or auto-healing event. As you investigate the root cause of an incident, the best place to start is to find what changed. 

毫不奇怪，大多数故障是由系统中某个地方的变化引起的，比如新的代码部署、配置变化、自动扩展活动或自动修复事件。当你调查事件的根本原因时，最好的开始是找到什么变化。

To understand what change caused a problem and what effects propagated across your stack, you need to be able to see how the relationships between stack components have changed over time: 

- What did the stack look like when a problem began? 
- What components are affected? 
- How are all the alerts related? We call this level of insight, which lets you track cause and effect across your stack, causal observability — it builds on the foundation laid in Levels 1 and 2.

为了了解什么变化导致了问题，以及什么影响在你的堆栈中传播，你需要能够看到堆栈组件之间的关系是如何随时间变化的。

- 当一个问题开始时，堆栈是什么样子的？
- 哪些组件受到了影响？
- 所有的警报是如何关联的？我们把这种让你跟踪整个堆栈的因果关系的洞察力称为因果可观察性--它建立在第一和第二层次的基础上。

We call this level of insight, which lets you track cause and effect across your stack, causal observability — it builds on the foundation laid in Levels 1 and 2.

我们把这一层次的洞察力称为因果可观察性，它可以让你追踪整个堆栈的因果关系，它建立在第一和第二层次的基础上。

> “Deriving patterns from data within a topology will establish relevancy and illustrate hidden dependencies. Using topology as part of causality determination can greatly increase its accuracy and effectiveness.” 
>
> "从拓扑结构内的数据中推导出模式将建立相关性，并说明隐藏的依赖关系。将拓扑结构作为因果关系确定的一部分，可以大大增加其准确性和有效性。" 
>
> – Gartner® Market Guide for AIOps Platforms, May 2022, Pankaj Prasad, Padraig Byrne, Gregg Siegfried
>
> – Gartner® AIOps平台市场指南，2022年5月，Pankaj Prasad, Padraig Byrne, Gregg Siegfried



Topology is the first necessary dimension for causal observability. Topology is a map of all the components in your IT environment that spans all layers, from network to application to storage, showing how everything is related. Topology incorporates logical dependencies, physical proximity and other relationships between components to provide human-readable visualization and operationalized relationship data. 

拓扑结构是因果观察能力的第一个必要维度。拓扑是IT环境中所有组件的地图，它跨越了所有的层次，从网络到应用到存储，显示了所有东西的关系。拓扑结构包含了组件之间的逻辑依赖性、物理接近性和其他关系，以提供人类可读的可视化和操作化的关系数据。

> Topology describes the set of relationships and dependencies between the discrete components in an environment, for example, business services, microservices, load balancers, containers and databases. 
>
> 拓扑结构描述了环境中离散组件之间的关系和依赖性集合，例如，业务服务、微服务、负载均衡器、容器和数据库。
>
> In today’s modern environments, topologies evolve quickly as new code gets pushed into production continuously and the underlying infrastructure changes rapidly. Managing these dynamic environments requires the ability to track changes in topology over time (time-series topology), giving historical and real-time context to the activities happening in your stack.
>
> 在今天的现代环境中，随着新的代码不断被推入生产，以及底层基础设施的快速变化，拓扑结构迅速发展。管理这些动态环境需要有能力跟踪拓扑结构随时间的变化（时间序列拓扑），为你的堆栈中发生的活动提供历史和实时背景。



Modern environments consist of so many dynamic layers, microservices, serverless applications and network technology that adding an up-to-date topology to your observability mix is essential to separate cause from effect. Topology provides anchor points for thousands of unconnected data streams to give them structure, making previously invisible connections visible. Topology visualization lets you view telemetry from network, infrastructure, application and other areas in the context of full-stack activity; it also gives you crucial context to know how your business is affected when something breaks.

现代环境由许多动态层、微服务、无服务器应用和网络技术组成，因此在你的可观察性组合中加入最新的拓扑结构，对于区分因果关系至关重要。拓扑结构为数以千计的未连接的数据流提供了锚点，使它们具有结构性，使以前看不见的连接变得可见。拓扑可视化让你在全栈活动的背景下查看来自网络、基础设施、应用程序和其他领域的遥测数据；它还为你提供了重要的背景，让你知道当某些事情发生时，你的业务是如何受到影响的。



{% asset_img figure2.png %}

Figure 2: Causal observability requires the consolidation of topology information from all the sources in your environment.

图2：因果观察能力需要整合环境中所有来源的拓扑信息。

However, for most companies, adding topology is not enough to provide causal observability on its own. Especially in today’s dynamic modern environments with microservices, frequent deployments, everchanging cloud resources and containers spinning up and down, topology changes fast. What your stack looks like now is probably not what it looked like when a problem first began. So a second dimension is necessary to create the foundation for causal observability: time.

然而，对于大多数公司来说，增加拓扑结构本身并不足以提供因果观察能力。特别是在当今动态的现代环境中，微服务、频繁的部署、不断变化的云资源和容器的上下旋转，拓扑结构变化很快。你的堆栈现在是什么样子，可能不是问题刚开始时的样子。因此，第二个维度对于创建因果观察能力的基础是必要的：时间。



{% asset_img figure3.png %}

Figure 3: Capture time-series topology to track stack changes and quickly troubleshoot root cause.

图3：捕获时间序列拓扑结构，以跟踪堆栈的变化并迅速排除根本原因。

And finally, to understand the dynamic behaviors of modern IT environments and get the context required to achieve causal observability, you need to correlate your environment’s topology with its associated metric, log, event and trace data over time.

最后，为了理解现代IT环境的动态行为，并获得实现因果观察能力所需的背景，你需要将环境的拓扑结构与其相关的指标、日志、事件和跟踪数据随时间推移而关联起来。

{% asset_img figure4.png %}

Figure 4: Capture topology over time and correlate it with metrics, logs, events and traces to track changes in your stack. Later, when issues occur, you can go back to the exact moment in time the issue started and see what change caused it.

图4：随着时间的推移捕获拓扑结构，并将其与指标、日志、事件和跟踪联系起来，以跟踪堆栈中的变化。以后，当问题发生时，你可以回到问题开始的确切时刻，看看是什么变化造成的。

At Level 3, the additional dimensions of topology and time, correlated with telemetry data show you the cause and impact of any change or failure across the different layers, data silos, teams and technologies — significantly improving resolution times and business outcomes. You also have the foundation to begin automating root cause analysis, business impact analysis and alert correlation. This deeper level of data is also required for more advanced AIOps, as you’ll read about in Level 4.

在第三级，拓扑结构和时间的额外维度，与遥测数据相关联，向您展示任何变化或故障的原因和影响，跨越不同的层、数据仓、团队和技术--大大改善解决时间和业务成果。你也有了开始自动进行根本原因分析、业务影响分析和警报关联的基础。这种更深层次的数据也是更高级的AIOps所需要的，你会在第四级中读到这一点。

> **4 Key Steps to Build Causal Observability and a Foundation for AIOps** 
>
> 1. Consolidate: First, you need to ensure you have consolidated data from across your stack into one place so you have a complete view. 
> 2. Collect topology data: Next, you need to build a topology map of your environment, which is a map of the components in your stack showing how they all relate to each other. Visualizing topology quickly answers the questions, “What component depends on other components? If one service fails, what else will be affected?” 
> 3. Correlate: You need to correlate all this unified data so your entire IT environment can be analyzed as a whole, even across silos. Every component in the topology needs to be correlated with its associated metric, log, event and trace data. 
> 4. Track everything over time: Finally, if you want to see how a change in one component propagates across your stack, you need to correlate your topology data with metric, log and trace data over time.
>
> **建立因果观察能力和AIOps基础的四个关键步骤**。
>
> 1. 整合。首先，你需要确保你已经将整个堆栈的数据整合到一个地方，以便你有一个完整的视图。
> 2. 收集拓扑结构数据。接下来，你需要建立一个环境的拓扑图，这是一个堆栈中的组件的地图，显示它们之间的关系。拓扑结构的可视化可以快速回答问题："什么组件依赖于其他组件？如果一个服务出现故障，还有什么会受到影响？" 
> 3. 关联。你需要将所有这些统一的数据关联起来，这样你的整个IT环境就可以作为一个整体进行分析，甚至是跨筒仓。拓扑结构中的每个组件都需要与其相关的指标、日志、事件和跟踪数据相关联。
> 4. 随着时间的推移跟踪一切。最后，如果你想看看一个组件的变化是如何在你的堆栈中传播的，你需要将你的拓扑数据与指标、日志和跟踪数据随着时间的推移进行关联。




| Level 3: Causal Observability                                |
| ------------------------------------------------------------ |
| Contextualize telemetry data (metrics, traces, events, logs) through a single topology. Correlate all data over time to track changes as they propagate across your stack.通过单一的拓扑结构将遥测数据（指标、跟踪、事件、日志）上下文化。随着时间的推移，对所有数据进行关联，以跟踪在你的堆栈中传播的变化。 |
| **System Input** Levels 1 and 2 + time-series topology1级和2级+时间序列拓扑结构 **System Output** Levels 1 and 2 + correlated topology, telemetry and time data displayed in contextual visualizations, showing the effects of changes across your stack第1级和第2级+相关的拓扑结构、遥测和时间数据显示在上下文的可视化中，显示整个堆栈的变化效果 |
| **What You Get** • Consolidated, clear, correlated, contextual view of the environment’s state, through unification of siloed data in a time-series topology • Significant acceleration in root cause identification and resolution times through topology visualization and analysis to understand cause and effect • Foundation for basic automated investigations such as root cause analysis, business impact analysis and alert correlation • Context needed to automatically cluster alerts related to the same root cause, reducing noise and distractions • Ability to visualize the impact of network, infrastructure and application events on business services and customers通过统一时间序列拓扑中的孤岛数据，对环境状态进行综合的、清晰的、相关的上下文视图 - 通过拓扑可视化和分析了解因果关系，大大加快根源识别和解决时间 - 为基本的自动调查奠定基础，如根源分析、业务影响分析和警报关联 - 自动聚类与同一根源相关的警报所需的上下文，减少噪音和干扰 - 能够可视化网络、基础设施和应用程序事件对业务服务和客户的影响 |

Table 5: Level 3 summary



### Next Step: Proactive Observability With AIOps 

As noted above, Gartner points out that topology can greatly increase the accuracy and effectiveness of causal determination. Level 3 is a big step forward, but unifying data from different silos poses challenges in terms of data normalization, correlation and quality that may require new capabilities or even organizational changes to resolve. In addition, it is difficult to collect and operationalize high-quality topology data at scale, especially in less modern environments. 

如上所述，Gartner指出，拓扑结构可以大大增加因果判断的准确性和有效性。第3级是一个很大的进步，但统一来自不同筒仓的数据在数据规范化、关联性和质量方面带来了挑战，可能需要新的能力甚至组织变化来解决。此外，很难大规模地收集和操作高质量的拓扑数据，特别是在不太现代化的环境中。

Each topology source needs to continuously flow through into the master topology, so you need to ensure you have a system with the capability to store topology over time. Storing topology that is correlated with telemetry data over time presents an even bigger challenge. 

每个拓扑源都需要不断地流入主拓扑，所以你需要确保你有一个能够长期存储拓扑的系统。随着时间的推移，存储与遥测数据相关的拓扑结构是一个更大的挑战。

Consider these issues as you develop your implementation plan. Also keep in mind that the velocity, volume and variety of data at Level 3 is usually so large that to achieve your overall reliability goals, AI is likely necessary to help separate the signal from the noise. When you take the step to Level 4, you add artificial intelligence for IT operations (AIOps) on top of Levels 1-3 to gain more accurate insights.

在你制定实施计划时要考虑这些问题。还要记住，第3级的数据速度、数量和种类通常非常大，为了实现你的整体可靠性目标，可能需要人工智能来帮助从噪音中分离出信号。当你迈向第4级时，你在第1-3级的基础上增加了IT运营的人工智能（AIOps），以获得更准确的洞察力。

> “With data volumes reaching or exceeding gigabytes per minute across a dozen or more different domains, it is no longer possible, much less practical, for a human to analyze the data manually in service of operational expectations.” 
>
> "随着数据量达到或超过每分钟数千兆字节，跨越十几个不同的领域，由人类手动分析数据以服务于运营期望，已经不可能，更不现实。" 
>
> – Gartner® Market Guide for AIOps Platforms, May 2022, Pankaj Prasad, Padraig Byrne, Gregg Siegfried
>
> – Gartner® AIOps平台市场指南，2022年5月，Pankaj Prasad, Padraig Byrne, Gregg Siegfried





## Level 4: Proactive Observability With AIOps 

*Goal: Analyze large volumes of data, automate responses to incidents and prevent anomalies from becoming problems.*



Level 4, Proactive Observability With AIOps, is the most advanced level of observability. At this stage, artificial intelligence for IT operations (AIOps) is added to the mix. AIOps, in the context of monitoring and observability, is about applying AI and machine learning (ML) to sort through mountains of data looking for patterns that 

- drive better responses 
- at the soonest opportunity 
- by both humans and automated systems. 

第四级，主动观察与AIOps，是最先进的可观察水平。在这个阶段，IT运营的人工智能（AIOps）被加入到这个组合中。AIOps，在监控和可观察性的背景下，是关于应用人工智能和机器学习（ML）来整理堆积如山的数据，寻找模式，从而 

- 推动更好的反应 
- 在最短的时间内 
- 人和自动化系统都能做出更好的反应。



In Gartner’s “Market Guide for AIOps Platforms,” May 2022, by Pankaj Prasad, Padraig Byrne and Gregg Siegried, Gartner defines the characteristics of AIOps platforms in the following way:

在Pankaj Prasad、Padraig Byrne和Gregg Siegried撰写的Gartner《AIOps平台市场指南》（2022年5月）中，Gartner对AIOps平台的特点作了如下定义。

> “AIOps platforms analyze telemetry and events, and identify meaningful patterns that provide insights to support proactive responses. AIOps platforms have five characteristics: 
>
> 1. Cross-domain data ingestion and analytics 
> 2. Topology assembly from implicit and explicit sources of asset relationship and dependency 
> 3. Correlation between related or redundant events associated with an incident 
> 4. Pattern recognition to detect incidents, their leading indicators or probable root cause 
> 5. Association of probable remediation” 
>
> "AIOps平台分析遥测和事件，并识别有意义的模式，提供洞察力以支持主动反应。AIOps平台有五个特点。
>
> 1. 跨领域的数据摄取和分析 
> 2. 从资产关系和依赖性的隐性和显性来源进行拓扑结构组装 
> 3. 与事件相关的或多余的事件之间的关联性 
> 4. 模式识别，以检测事件、其领先指标或可能的根本原因 
> 5. 可能的补救措施的关联" 



We have the same view on AIOps as Gartner. AIOps builds on core capabilities from previous levels in this maturity model — such as gathering and operationalizing data, topology assembly and correlation of data — and adds in pattern recognition, anomaly detection and more accurate suggestions for remediating issues. Causal observability is a necessary foundation: time-series topology provides an essential framework. 

我们对AIOps的看法与Gartner相同。AIOps建立在这个成熟度模型中前几个级别的核心能力之上--比如收集和操作数据、拓扑组装和数据的相关性--并加入了模式识别、异常检测和更精确的问题补救建议。因果观察能力是一个必要的基础：时间序列拓扑结构提供了一个基本框架。

AIOps can help teams find problems faster and even prevent problems altogether. AI/ML algorithms look for changes in patterns that precede warnings, alerts and failures, helping teams know when a service or component starts to deviate from normal behavior and address the issue before something fails. 

AIOps可以帮助团队更快地发现问题，甚至完全预防问题。人工智能/ML算法寻找警告、警报和故障之前的模式变化，帮助团队了解服务或组件何时开始偏离正常行为，并在发生故障之前解决这个问题。

> “Spotting an anomaly is easy because they occur all the time. When you collect one billion events a day, a one-in-a-million incident happens every two minutes. The key for observability tools is to spot the anomaly that is relevant to the problem at hand, and then to link other bits of information from log files / metrics that are likely to be related. By surfacing correlated information in context, the operator can more quickly isolate the potential root cause of problems.” 
>
> "发现异常现象很容易，因为它们一直在发生。当你每天收集10亿个事件时，每两分钟就会发生一个百万分之一的事件。可观察性工具的关键是发现与手头问题相关的异常情况，然后将日志文件/指标中可能相关的其他信息点联系起来。通过浮现上下文中的相关信息，操作员可以更迅速地分离出问题的潜在根源"。
>
> – Gartner® “Innovation Insight for Observability,” March 2022, Padraig Byrne and Josh Chessman 
>
> – Gartner® "Innovation Insight for Observability"，2022年3月，Padraig Byrne和Josh Chessman 



However, anomalies occur frequently. They do not necessarily mean a problem will occur, nor that remediation should be a high priority. AIOps helps determine which anomalies require attention and which can be ignored. 

然而，反常现象经常发生。它们不一定意味着问题会发生，也不意味着补救措施应该是高度优先的。AIOps帮助确定哪些异常情况需要关注，哪些可以忽略。

Another goal of AIOps for observability is to drive automated remediation through IT service management (ITSM) and selfhealing systems. If these systems receive incorrect root cause input, for example, they can self-correct the wrong issue and cause bigger problems. AIOps delivers more accurate input that enhances their effectiveness. 

AIOps可观察性的另一个目标是通过IT服务管理（ITSM）和自我修复系统推动自动修复。例如，如果这些系统收到不正确的根本原因输入，它们可以自我纠正错误的问题并导致更大的问题。AIOps提供了更准确的输入，增强了它们的有效性。

> An ounce of prevention is worth a pound of cure. What better way to improve reliability than to stop incidents from ever happening at all?
>
> 一盎司的预防胜过一磅的治疗。还有什么比阻止事故发生更好的方法来提高可靠性呢？



At Level 4, you should notice more efficient and incident-free IT operations that deliver a better customer experience. To achieve these goals, set up AIOps to transcend silos and ingest data gathered from across the environment. The AI/ML models should analyze all the observability data types we discussed in previous levels: events, metrics, logs, traces, changes and topology, all correlated over time.

在第四级，你应该注意到更高效和无事故的IT运营，提供更好的客户体验。为了实现这些目标，设置AIOps以超越孤岛，并摄取从整个环境中收集的数据。AI/ML模型应该分析我们在前几个级别中讨论的所有可观察的数据类型：事件、指标、日志、跟踪、变化和拓扑结构，所有这些都是随时间推移而关联的。

> **A Word of Caution: Don’t Skip Level 3 **
>
> Proactive observability with AIOps is the best way to ensure reliable operation of your IT systems, but it’s a mistake to move directly to Level 4 and skip over the causal observability steps in Level 3 (data consolidation, topology, correlation of all data streams over time). 
>
> Each level in this Observability Maturity Model builds on capabilities established in previous levels, but having a complete foundation matters most for success in Level 4. If you apply AI/ ML without a comprehensive foundation of data, you can actually cause damage. For example, let’s say you use AI/ML on the front end of an automated self-healing system. If the algorithm determines an incorrect root cause, the self-healing system tries to remediate the wrong thing and can further break the system. If you apply AI/ML on top of insufficient data or poor-quality data, you may drive automation in the wrong direction as the algorithm learns the wrong thing. 
>
> Without topology data correlated with metric, log and trace data over time, AIOps tools will likely not understand the correlation between these various sorts of data as they come together. AIOps needs the additional context provided by topology and time in order to accurately assess root cause, determine business impact, detect anomalies and proactively determine when to alert SRE and DevOps teams.
>
> **提醒您。不要跳过第三级 **
>
> 利用AIOps进行主动观察是确保IT系统可靠运行的最佳方式，但直接进入第4级并跳过第3级的因果观察步骤（数据整合、拓扑结构、所有数据流随时间变化的关联性）是一个错误。
>
> 这个可观察性成熟度模型中的每个级别都是建立在前几个级别建立的能力之上的，但拥有一个完整的基础对第四级的成功最为重要。如果你在没有全面的数据基础的情况下应用AI/ML，你实际上会造成损害。例如，假设你在一个自动化自我修复系统的前端使用AI/ML。如果算法确定了一个不正确的根本原因，自我修复系统就会试图补救错误的东西，并可能进一步破坏系统。如果你在数据不足或数据质量差的基础上应用AI/ML，你可能会在算法学到错误的东西时推动自动化走向错误的方向。
>
> 如果没有拓扑数据与指标、日志和跟踪数据长期相关，AIOps工具很可能无法理解这些不同种类的数据之间的相关性，因为它们走到了一起。AIOps需要拓扑结构和时间提供的额外背景，以便准确评估根本原因，确定业务影响，检测异常情况，并主动确定何时提醒SRE和DevOps团队。



| Level 4: Proactive Observability With AIOps                  |
| ------------------------------------------------------------ |
| Use AIOps to sort through mountains of data and identify the most significant patterns and impactful events, so teams can focus their time on what matters.使用AIOps对堆积如山的数据进行分类，确定最重要的模式和有影响的事件，这样团队就可以把时间集中在重要的事情上。 |
| **System Input** Levels 1-3 + AI/ML models1-3级+AI/ML模型 **System Output** Levels 1-3 + proactive insights that enable fast MTTR and prevent failures1-3级+主动洞察，实现快速MTTR并防止故障发生 |
| **What You Get** • New insights into IT environment operations using AI/ML to gather and correlate actionable information from large volumes of data • Predictions and anomaly detection that highlight issues before they impact the business • Greater efficiency and reduced toil as teams focus effort on the most impactful events • Improved accuracy of automatic root cause analysis, business impact analysis and alert correlation • Incident data that is accurate enough to use effectively with automated ITSM and self-healing systems使用AI/ML从大量数据中收集和关联可操作的信息，对IT环境运行有新的见解 -预测和异常检测，在问题影响业务之前突出问题 -提高效率，减少劳累，因为团队将精力集中在最有影响的事件上 -提高自动根本原因分析、业务影响分析和警报关联的准确性 -事件数据足够准确，可以有效使用自动化ITSM和自愈系统 |

Table 6: Level 4 summary



### Next Steps 

Most AIOps solutions today require significant configuration and training time but often yield inaccurate results, especially if topology changes over time are not considered. Teams often implement them with unrealistic expectations and unclear goals, then find themselves disappointed. 

今天，大多数AIOps解决方案需要大量的配置和培训时间，但往往产生不准确的结果，特别是如果不考虑拓扑结构随时间的变化。团队经常在不切实际的期望和不明确的目标下实施它们，然后发现自己很失望。

Level 4 is the final observability maturity level for now, but as IT continues to evolve, we fully expect a Level 5 to emerge. 

4级是目前最后的可观察性成熟度级别，但随着IT的不断发展，我们完全期待5级的出现。

## Summary

For decades, IT operations teams have relied on monitoring for insight into the availability and performance of their systems. But the shift to more advanced IT technologies and practices is driving the need for more than monitoring – and so observability evolved. With infrastructures and applications that span multiple dynamic, distributed and modular IT environments, organizations need a deeper, more precise understanding of everything that happens within these systems. Observability provides that comprehensive insight, delivering clear capabilities at each level of maturity.

几十年来，IT运营团队一直依靠监控来洞察其系统的可用性和性能。但是，向更先进的IT技术和实践的转变推动了对监控以外的需求--因此，可观察性也随之发展。随着基础设施和应用跨越多个动态、分布式和模块化的IT环境，企业需要更深入、更精确地了解这些系统内发生的一切。可观察性提供了这种全面的洞察力，在每个成熟度级别上提供了明确的能力。

**Drivers to Improve Maturity**

| Level                                       | Drivers                                                      |
| ------------------------------------------- | ------------------------------------------------------------ |
| Level 1: Monitoring                         | Level 1 is sufficient for classic static infrastructure.1级对经典的静态基础设施来说是足够的。 |
| Level 2: Observability                      | Level 2 capabilities become more critical as you shift to cloud, container and microservices architectures and implement CI/CD.当你转向云、容器和微服务架构并实施CI/CD时，第二级能力变得更加关键。 |
| Level 3: Causal Observability               | Level 3 capabilities become essential for maintaining hybrid environments, expanding to multi-cloud platforms, implementing containers, microservices and more advanced CI/CD at scale.三级能力对于维护混合环境、扩展到多云平台、实施容器、微服务和更先进的CI/CD的规模变得至关重要。 |
| Level 4: Proactive Observability with AIOps | As companies attempt to automate systems for event correlation, automatic ticket creation, ticket consolidation, automatic remediation and self-healing, Level 4 capabilities for AIOps are required. The intelligence provided by AIOps delivers the data accuracy necessary for these systems.随着公司试图将事件关联、自动票据创建、票据整合、自动修复和自我修复的系统自动化，需要AIOps的4级能力。AIOps所提供的智能为这些系统提供了必要的数据准确性。 |

Figure 5: Typical technology environments that drive companies to advance their observability maturity.

图5：推动企业推进可观察性成熟度的典型技术环境。

Each level of observability is characterized by distinct goals, inputs, outputs and capabilities. You’ll also find commonalities in typical tooling at each level.

每个级别的可观察性都有不同的目标、输入、输出和能力的特点。你还会发现每个层次的典型工具的共同点。

|                                                              | Level 1: Monitoring | Level 2: Observability | Level 3: Causal Observability | Level 4: Proactive Observability With AIOps |
| ------------------------------------------------------------ | ------------------- | ---------------------- | ----------------------------- | ------------------------------------------- |
| **Observability Goals**                                      |                     |                        |                               |                                             |
| Ensure that individual components are working as expected.确保个别部件按预期工作。 | ✅                   | ✅                      | ✅                             | ✅                                           |
| Determine why the system is not working.确定系统不工作的原因。 |                     | ✅                      | ✅                             | ✅                                           |
| Find the cause of the incident and determine its impact across the system.找到事件的原因并确定其对整个系统的影响。 |                     |                        | ✅                             | ✅                                           |
| Analyze large volumes of data, automate responses and prevent anomalies from becoming problems.分析大量的数据，自动化反应，防止异常情况变成问题。 |                     |                        |                               | ✅                                           |
| **System Input**                                             |                     |                        |                               |                                             |
| Events and component-level metrics事件和组件级指标           | ✅                   | ✅                      | ✅                             | ✅                                           |
| Metrics, logs, traces (comprehensive)指标、日志、追踪（全面）。 |                     | ✅                      | ✅                             | ✅                                           |
| Time-series topology时间序列拓扑结构                         |                     |                        | ✅                             | ✅                                           |
| AI and ML modelsAI和ML模型                                   |                     |                        |                               | ✅                                           |
| **System Output**                                            |                     |                        |                               |                                             |
| Alerts报警                                                   | ✅                   | ✅                      | ✅                             | ✅                                           |
| Comprehensive dashboards全面的仪表板                         |                     | ✅                      | ✅                             | ✅                                           |
| Understand cause and effect of change理解变化的原因和效果    |                     |                        | ✅                             | ✅                                           |
| Automated root cause analysis自动化的根本原因分析            |                     |                        | ✅                             | ✅                                           |
| Automated business impact analysis自动化的业务影响分析       |                     |                        | ✅                             | ✅                                           |
| Correlated alerts / noise reduction相关警报/降噪             |                     |                        | ✅                             | ✅                                           |
| Predictive and preventative insights预测性和预防性的洞察力   |                     |                        |                               | ✅                                           |
| **Typical Tooling**                                          |                     |                        |                               |                                             |
| Classic domain-centric monitoring tools (e.g., infrastructure monitoring, application monitoring, API monitoring, synthetic monitoring, network monitoring, business monitoring), eventbased alerting.经典的以领域为中心的监控工具（例如，基础设施监控、应用监控、API监控、合成监控、网络监控、业务监控），基于事件的警报。 | ✅                   | ✅                      | ✅                             | ✅                                           |
| APM/observability tooling – APM tools, modern observability tools based on OpenTelemetry, observability data lakes (previously known as log aggregators), domain-agnostic combinations of open source metrics, trace and log tooling, sometimes unified in dashboard tooling.APM/可观察性工具--APM工具、基于OpenTelemetry的现代可观察性工具、可观察性数据湖（以前称为日志聚合器）、开源指标的领域诊断组合、跟踪和日志工具，有时统一于仪表盘工具。 |                     | ✅                      | ✅                             | ✅                                           |
| More advanced APM and observability tooling with causal reasoning and event correlation capabilities, powered by time-series topology. Level 3 is an emerging market area.更高级的APM和可观察性工具，具有因果推理和事件关联能力，由时间序列拓扑结构提供支持。3级是一个新兴的市场领域。 |                     |                        | ✅                             | ✅                                           |
| Data-agnostic AIOps solutions that can find patterns in large amounts of data to provide smart capabilities, such as anomaly detection and leading indicator detection/proactive alerting. Level 4 is an emerging market area.数据诊断型AIOps解决方案，可以在大量数据中找到模式，提供智能能力，如异常检测和领先指标检测/主动警报。第四级是一个新兴的市场领域。 |                     |                        |                               | ✅                                           |

Figure 6: Characteristics at each level of observability maturity. Where does your organization best fit?

图6：可观察性成熟度的各个层次的特征。你的组织最适合哪里？

The higher your maturity level, the more resilient and reliable your IT systems will be. You’ll be able to troubleshoot the root cause of problems more quickly, understand business impact of changes and failures and ultimately deliver a better experience for customers

你的成熟度越高，你的IT系统就越有弹性和可靠性。你将能更快地排除问题的根源，了解变化和故障的业务影响，最终为客户提供更好的体验。

## References

1. [“18 Key Areas Shaping IT Performance Markets in 2020,”](https://www.dej.cognanta.com/2020/01/24/18-key-areas-shaping-it-performance-markets-in-2020/) Digital Enterprise Journal (DEJ) 
2. Enterprise Management Associates (EMA) APM Tools Survey 
3. [“2022 State of Managing IT Performance Study – Key Takeaways,”](https://www.dej.cognanta.com/2022/07/14/2022-state-of-managing-it-performance-study-key-takeaways/) Digital Enterprise Journal (DEJ) 
4. [Distributed Systems Observability: A Guide to Building Robust Systems: A Guide to Building Robust Systems](https://www.oreilly.com/library/view/distributed-systems-observability/9781492033431/ch04.html), by Cindy Sridarhan, O’Reilly Media, 2018. 
5. [Site Reliability Engineering: How Google Runs Production Systems](https://sre.google/sre-book/monitoring-distributed-systems/), edited by Betsy Beyer, Chris Jones, Jennifer Petoff and Niall Richard Murphy, O’Reilly Media, 2016.