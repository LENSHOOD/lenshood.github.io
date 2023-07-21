---
title: 性能工程实践
date: 2023-07-13 21:28:12
tags:
- performance engineering
- performance optimization
- performance analysis
categories:
- Software Engineering
---

{% asset_img header.jpg 500 %}

本文书接上篇《什么是性能工程》，通过介绍相关实践、方法与尝试，试图回答 “企业如何达成性能工程目标” 的问题。



## 前言

经过上一篇内容《什么是性能工程》的讨论，我们引出了 *“性能工程，是指通过设计、构建工具链和工作流，从而对系统性能进行持续改善和守护的一类实践方法”* 这一观点，并基于此定义了落地性能工程的目标：

**DevPerfOps：构建性能工程反馈闭环**

**固化专家经验形成知识库，沉淀性能优化标准实践**

**自助化性能分析，降低工具学习和使用成本**

本文将围绕着上述三个建设目标，通过介绍相关的尝试和方法，以讨论企业建设性能工程体系的最佳实践。

后文将介绍如下几个方法实践：

1. 以技术支撑研发流程再造的过程，实现性能左移。
2. 守护性能基线，及时发现性能劣化，并通过定界技术寻找性能瓶颈。
3. 构建性能工程平台，沉淀工具、知识和模式，打造一站式性能优化能力。



## 性能左移：技术支撑的流程再造

上篇文章提到过，现有的成熟研发流程中性能往往被忽视，导致与性能相关的工作不断被后置。

性能左移恰恰是尝试将性能相关的工作尽量前置，向研发流程的左侧移动。这包括设计阶段的性能建模，开发阶段的性能实践以及测试阶段的性能测试/仿真。

早期落地 DevOps 的经验告诉我们，由于人的因素，仅通过改变工作流程和内容，而不提供相应的技术支撑，是难以实现流程的再造和变革的。性能左移也不例外，为了实现性能左移的目标，我们扩展了自动化流水线，并提供了相应技术实践，扩展后的流水线如下所示：

{% asset_img 1.png %}

**性能建模与架构评议**

性能建模是整个流程的第一个环节。它一方面包含了在设计阶段需要明确的系统性能要求，如功能响应时间、关键路径吞吐量、服务并发量等，以及对资源利用率相关的要求和设计，如弹性扩缩指标，系统容量规划和预估等。

另一方面则包含了架构的性能关注点地图。关键的业务模块、性能阻塞概率大的组件以及各种交互接口都属于性能关注点的范畴。性能关注点地图能够在性能分析和优化时为我们提供逻辑、交互、算法和数据结构上的设计信息，并引导我们聚焦在关键路径和更易发生性能劣化的地方。

性能建模将与其他设计产出物一同进入架构评议环节进行评审。为了确保在后续流程中设计约束不被破坏，相关设计约束会以各类门限的形式挂到流水线上，例如对比指标是否达成，关键点测试覆盖率是否达标等。

**代码扫描与微基准测试**

在开发阶段，通过代码扫描工具，能够低成本的找出性能不友好的代码段，并提示开发者及时修改。目前市面上有众多静态扫描工具可供选择，它们不仅能识别性能问题，也能发现安全和质量问题。

被识别为性能关注点的功能，需要开发者为关键代码编写微基准测试。可以通过插桩、埋点等形式将性能关注点地图与分析系统关联起来，每一次进行流水线构建都会执行开发者编写的微基准测试，其测试结果也会通过插桩点上报给分析系统，以供统计和分析。

**测试框架与插件市场**

传统流程中，性能测试占比本身较低，因此通过简单的脚本式用例结合常见的性能探测工具就可以满足测试需求。但随着性能左移的落地，测试环节对性能测试的要求从孤立功能的通用指标测试转化为对系统整体的性能仿真以及多维度的指标验证，此时传统性能测试方法可能因效率低下而难以为继。

通过构建符合企业自身场景的性能测试体系，能够为性能测试提供基础能力支撑。成熟的测试框架，一方面集成了标准化的用例模板或脚手架，简化通用场景用例的编写，另一方面可为用例编写者提供负载生成、指标收集、执行任务编排、分析可视化等各类原子工具，方便按需组合。

最后，建立插件市场，将上述用例模板和原子化工具在企业内部开放并自下而上形成生态闭环。



## 持续改进：性能看护+定界分析

DevPerfOps 要求能实现性能工程的反馈闭环。性能左移已经让性能相关工作尽可能前置，但为了实现完整的反馈闭环，性能测试之后的阶段也非常关键。

系统上线后，真实的运行数据将不断验证性能指标是否达标，对发现的性能问题，也会开展性能优化工作。但系统并非一成不变，随着版本不断迭代，各类性能问题可能也会反复发生。为了能持续的监控系统性能，更快地识别性能变化并定位导致变化的代码块，需要支持性能看护和定界分析的能力。

**动态看护，统计性能变化**

看护的本质是建立基线，度量变化。上线前性能测试和上线后性能监控都可以产生大量性能指标数据，对指标数据进行清洗和筛选后，形成不同环境下的性能基线。在版本迭代过程中，通过对比基线，就能方便的发现性能变化点和劣化点。

{% asset_img 2.png %}

传统的性能基线大都来自于施加外部负载而得到的整体监控数据，这种将系统看做黑盒的基线缺少细节信息，导致即使通过基线对比发现了指标劣化，也难以快速追踪到问题所在。我们基于性能关注点地图，可以设置更细致的监控追踪点，对系统形成更深入的洞察。因此能够实现将黑盒观测数据与进程级、组件级观测数据相关联，使看护报告能提供进一步信息检视诊断的能力，帮助定位性能劣化根因。

理想情况下，性能指标数据的变化服从高斯分布，因此很容易统计和对比。然而实际场景下，性能指标的波动趋势远比仿真环境下复杂得多，通过概率统计方法对指标数据进行处理后，才能形成判断性能劣化的依据。常见的，指标分布中出现少量离群点，可能预示着出现了环境问题或是功能故障，而指标分布偏左或偏右表示可能存在未分离的变量影响，需要增加观测指标。

**定界分析，缩小责任边界**

看护报告可以指出哪几个关联指标出现劣化，顺着这些指标我们能大致将分析范围缩小到到服务或模块级别。但这种级别的指标数据粒度还是太粗，为了实施性能优化，需要分析更细粒度的性能数据，才能找到具体的劣化代码块。

为了深入的了解服务或模块的性能变化，一般采用性能剖析的手段，在函数级别进行数据采样。采样剖析可以生成函数调用栈、函数总耗时 / On CPU 耗时、函数内存消耗、堆内存分析等报告，可以通过它们寻找劣化代码块。在对劣化代码进行优化后，再次做剖析，前后差分对比可以分析优化效果。

成功的优化不仅能改善系统性能，还可以与相关的多个性能指标进行关联。收集优化前后各类指标数据的变化，积累起来形成指标特征，可训练生成性能劣化模型。在未来的性能看护场景下，发现指标劣化后，系统能自动尝试进行根因识别，给出可能的劣化点位置以及匹配到的优化案例，从而进一步降低性能优化的成本。



## 一站式优化：构建性能工程平台能力

通过前文提到的各种实践，研发团队基本实现了 DevPerfOps 的全流程反馈闭环。接下来，让我们回到性能工程的本质：“为了对系统性能进行持续改善和守护”。

性能优化并非易事。成功的性能优化，对实施者的技术水平、业务和架构的理解程度，以及对系统观测的深度都有较高的要求。得益于前文提到的各种实践打下的坚实基础，提升性能优化效率的路径变得逐渐清晰：建设平台化能力，固化知识、建设基础设施并提供自助式服务。

**性能知识图谱与专家协同**

积累了大量性能优化经验的专家永远是各个产品线争抢的竞争性资源，成为性能专家的路线需要多年的学习与总结、广阔的视野，以及深入系统实现细节和算法原理的研究性能力。性能专家很关键也很稀缺，对专家经验进行固化，是一种能有效降低成本并提升一线开发团队能力的手段。

性能优化包括分析和优化两个过程，分析要求逻辑性思维和关联演绎能力，优化更注重知识深度和经验方法。性能专家 Brandon Gregg 著名的 “性能分析黄金60秒” 就是一种固化的通用分析方法。通过 DevPerfOps 持续的性能反馈闭环，企业逐步积累了富有参考价值的性能分析优化的方法路径，我们可将这些过程总结成知识图谱：包括性能分析路径、性能优化模式以及案例等。知识图谱不仅能对新手产生足够的指引，形成更加顺畅的性能优化体验，也可以成为了企业不可多得的无形资产。

知识图谱能显著提升通用场景性能优化的效率，但性能专家本身是不可或缺的，毕竟仍然存在大量孤立场景和疑难杂症。通过开发平台化能力，研发团队可以在平台上创建 “性能优化专项”，将服务架构图、业务场景描述、已有的性能分析结果等相关上下文都汇总在项目里，并通过预约系统申请相关领域的专家资源。这样性能专家一旦进入项目，就能以最快的速度熟悉上下文，并通过项目授权的临时凭证访问相关系统，以及随时通过视频会议进行远程协助。

**启发式自助分析流**

不常进行性能优化的研发人员，即使有知识图谱作指导，在面临性能优化工作时也总感到无从下手。类似 “性能分析黄金60秒” 的实践就是为了帮助实施者快速入手，并了解系统概貌的一种实践。

既然我们已经拥有了观测系统、工具插件市场、分析优化知识图谱、性能看护数据和劣化特征模型，基于此完全有条件构建性能工程平台，整合这些优秀实践，为研发人员提供启发式的自助分析工作流。

图

研发人员可在平台上创建项目，选择关注的服务或系统后，平台根据历史观测数据绘制服务调用/依赖图，并展示节点的基本概况、性能指标和关联的性能优化案例。平台还可基于以往的劣化特征模型，建议用户关注可能存在问题的服务或实例。之后，用户下钻到某个感兴趣的实例，并对该实例启动分析，分析会先执行通用采集器，绘制基于时间轴联动或组件联动的数据采集结果，根据初次采集结果，用户可进一步选择进程或线程，进行二次性能剖析，通过调用栈、火焰图甚至TopDown分析等手段，发现问题。为了识别具体的代码劣化点，用户可以调取同一服务上一版本的性能剖析历史，并进行差分对比，尝试发现变更点和问题来源，还可以通过函数符号关联代码库或是反汇编代码，直接在代码级别进行 Diff 来发现问题代码。



## 性能工程的成熟度路线

从性能工程的挑战、目标以及实践来看，企业建设性能工程体系的诉求逐渐加强，条件也在不断成熟：云原生基础设施、全链路可观测、AI 辅助等技术的不断发展，企业越来越有信心能建成性能工程系统。

结合前文的内容，我们认为性能工程体系的演进，从成熟度路线上可分为以下几个层次：

1. 流程化：
   - 将性能建模等新实践引入研发工作流，在团队中建立性能优化的意识
   - 企业研发自己的测试框架和工具集并能集成性能测试/仿真到流水线
   - 扩展可观测系统，建成对性能侧指标数据的收集和分析能力
2. 数字化：
   - 建立性能基线，并实现持续性能看护和性能定界
   - 积累性能指标数据和变化趋势，驱动持续优化和架构决策
3. 资产化
   - 性能专家形成资源池，专家经验被固化，建立性能知识图谱形成知识库
   - 总结性能分析路径和优化模式，最终沉淀为公共能力组件，在更底层解决性能问题
4. 平台化
   - 整合各类实践和资产，将性能工程能力扩充到企业内部平台中
   - 不断总结工程方法，演进平台化能力，性能活动成为研发流程的常态



## 总结

性能工程，是指通过设计、构建工具链和工作流，从而对系统性能进行持续改善和守护的一类实践方法。

伴随着硬件发展不断放缓，系统架构愈发复杂，维护和演进成本高企，性能问题在近年来逐步暴露。这给企业研发带来了许多问题和负担，但也为工程实践领域引入了新的机遇。

经过上一篇《什么是性能工程》和本篇《性能工程实践》，我们从硬件和软件的发展角度讨论了性能在软件研发流程中的重要性，也提到了在当下研发流程中嵌入性能相关工作的痛点和挑战。

通过性能工程实践方法，我们期望能通过性能左移、性能看护和定界分析构建 DevPerfOps 反馈闭环，并构建性能工程平台化能力来固化专家经验、沉淀方法，以至于最终形成自助式的性能分析和优化流程。

相信越来越多的开发者和组织都会开始关注性能工程领域，探索更优秀的方法和实践，以更好的驾驭复杂系统。