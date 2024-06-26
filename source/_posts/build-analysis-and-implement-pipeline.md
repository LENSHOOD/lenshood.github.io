---
title: 构建分析设计与落地交付的迭代式工作流
date: 2022-07-16 19:44:11
tags:
- business analysis
- project management
categories:
- Software Engineering
---

{% asset_img header.jpg 500 %}

本文介绍了一种将 inception 分析部分拆细，已迭代的形式掺入落地阶段的一种办法。



<!-- more -->

### 交付模式的问题

在 ThoughtWorks，一般我们在项目正式进入交付之前，都会预设一个数天到数周的 Inception，用来明确需求，控制风险，以及进行前期设计。

理想情况下，通过一个中短期的 Inception 阶段，我们不仅与客户一起明确了想要达成的业务目标和要实现的业务价值，也梳理清楚了交付范围并制定了相对详细的落地计划。

在手握迭代计划、故事列表以及技术详设之后，交付同学就可以遵循各种敏捷实践，干的又快又好，顺利交付项目。

然而实际情况总是复杂且令人头疼的，由于各种原因，在项目交付的过程中，总是会遇到依赖阻塞、需求膨胀、错误返工等等预期外的状况，不仅会造成项目风险上升，同时也会降低士气，影响质量，甚至导致项目失败。

在交付过程中遇到的预期外的状况，我们在复盘的时候，经常会将之归咎于业务分析阶段做的不扎实：或是业务没有完全分析清楚，或是某些关键节点的风险没能正确识别等等。进而就很容易得出直接的原因：分析阶段时间太紧张、客户不配合、业务过于杂乱等等，但正因为分析阶段已经过去了，所以解决的办法只能是拖时间再分析，进而延长合同时间，面临扯皮、扣款甚至终止合作。最后在项目结束时大家一起唏嘘：下一个项目，一定要把分析做足！

但所谓 ”将业务分析阶段做扎实“，是否可能是一个伪命题？



#### 分析与落地难以泾渭分明

正如前文所述，在落地阶段持续的进行分析工作，是交付过程中的常见现象。

现实中，即使是在 Inception 阶段已经产生了完整的故事列表和技术方案，交付过程中仍旧会存在很多由于异常事件和不确定性而产生的未知风险。

在敏捷交付流程中，我们已经有了许多实践来应对这些风险。

为了尽早暴露风险，我们采用按迭代 showcase 并在下个迭代开始前 IPM 的办法，尽早得收集客户意见并明确变化点。同时，我们会通过 Spike 来快速明确问题，以应对交付中产生的突发不确定。此外，Story Kick Off 的流程也会帮助 dev 和 ba 澄清业务价值，如果遇到难以澄清的业务，ba 可以快速与客户进一步沟通。

上述的各种实践，佐证了分析设计与交付落地天然就不是泾渭分明的。



#### Inception 难以面面俱到

Inception 的产出是相对粗粒度的，不太能面面俱到，这主要由时间和复杂度两方面因素导致。

从客户的角度讲，着急的客户会期望需求抛出后，我们能尽早尽快的开始分析、设计、开发，尽早落地，尽早上线。而即使是需求不那么紧急的客户，只要是以交付为目的，也都不太会接受长时间的 Inception。

从业务复杂度的角度讲，越复杂的业务，梳理起来越困难，分析的时候就更加倾向于厘清主干；如果上下游系统很多，技术方案也就会更多考虑系统如何集成、服务进程间如何交互。

时间短 + 复杂度高会导致 Inception 阶段的工作强度大，压力也比较大。产出的业务分析就可能抓大放小，不够细致，技术方案也会偏向于整体层面。

因此，不论分析团队与交付团队是不是一拨人，站的角度不同，产生的体验就会产生差异。分析团队会认为在压力较大的客户现场我们已经做的比较详尽了，而交付团队仍旧会在出现交付风险时归咎于分析的不充分。



### 构建分析+落地的迭代式工作流

前文我们讨论了交付项目中，Inception 阶段所存在的分析设计不够扎实细致的问题，以及产生这一问题的原因。同时也提到，在实际的敏捷交付中，已经存在了一些应对分析不足问题的方法。

这一节我们将介绍从流程上明确分析设计迭代的实践，并讨论这一实践所需要包含的内容及产出。



#### Inception 都包含了那些内容？

##### 分析调研

调研阶段主要工作是结合当下现状，与客户做需求的梳理和澄清，产出的内容可能会包含商业画布、用户画像、服务蓝图等。

##### 方案设计

在明确了需求之后就可以进行方案的设计，其中包括了业务方案规划、原型设计、技术方案规划、领域建模等部分。当然在实际进行方案设计的时候，通常对业务的理解也是一个逐步清晰的过程，所以与客户在一起工作很重要。

##### 形成计划

在方案设计过程中，会不断的与客户拉通对齐，设计、梳理的差不多之后，就比较有信心能把大的方案拆细，形成故事列表，根据工作量的预估情况，就能产出交付计划了。



#### 拆分 Inception，按迭代进行分析设计

明确了 Inception 所要做的工作，我们就可以尝试对其中的任务进行拆解了。

在此之前，我们需要明确拆解分析设计迭代的原则：

1. 哪些内容需要提前确定？

   显然，在交付之前进行 Inception 是有其明确价值和缘由的，不是所有的部分都可以拆解、并推迟进行。一定要在前期阶段就明确的内容，除了毋庸置疑的分析调研阶段外，还包括了：

   业务模块划分：划分清楚粗粒度的业务模块，能加深对业务全景的理解，也有助于安排交付计划。

   技术架构与规范：根据当前的技术现状，进行技术可行性分析。结合业务模块和服务集成关系，分析并明确系统上下文，制定技术规范，如数据一致性，错误处理，隐私安全等等。

2. 可以拆分并推迟的部分

   业务流程：只需要完成最近一个交付迭代所包含的业务流程分析。拆分业务流程的前提是迭代划分必须合理，否则可能导致后面的迭代推翻了前面的工作。可以按照业务上下文结合逐步增添特性的办法来划分。

   领域建模：对应业务流程，只需要包含最近迭代涉及到的业务的模型，其他上下游的部分可以简化，只留相对稳定的接口。

   技术选型：随着业务的不断分析，涉及到的技术方案和选型会逐步显现，对前期迭代的一些设计可能也存在改进。对跨功能的技术选型，可根据实际情况分析论证。

   部署架构：部署架构通常也会随着迭代的进行不断的演进，例如从单体到微服务的过程。



##### 分析迭代产出物和粒度

从前面的讨论中，我们能明确，分析迭代的工作重点是具体业务的梳理以及相关技术建模、规则和数据的梳理等，这一类内容能明确的指导落地实施。

- 业务推演：主干业务是什么样子的

  - 梳理业务流程图，明确每一个业务节点，与相关责任方（如领域专家）达成一致
  - 业务流程中还应该包含上下游的交互关系，最好能简单的表述上下游的业务，以实现更完整的逻辑串联
  - 和客户或领域专家共同完成业务推演是很重要的，近距离工作所能传递的知识细节是远程协作无法比拟的

- 实现依赖：纸上的业务要实现成代码，还需要哪些依赖

  - 业务规则：上游数据经过怎样的规则和演算，才能得到满足业务目标的结果。规则代表了程序的行为
  - 支撑数据：为了实现业务规则，需要的支撑性数据，如产品主数据、业务数据、财务数据等
  - 集成关系：上下游之间的集成关系，本系统与其他外部系统之间的交互流程

- 技术建模：模型可以直接指导落地开发，因此应采用更面向程序员的语言来描述

  - 静态模型：
    - 关键节点的业务凭证
    - 业务规则映射
    - 上下游之间的接口
  - 动态流程
    - 采用时序图、流程图等描述静态模型中欠缺的分支和数据流向
    - 描述业务输入经过怎样的映射、组装，得到业务输出
    - 结合动态流程，就可以补全静态模型中业务凭证的生命周期

- 跨功能需求

  - 本迭代内可能涉及到的交互模式、中间件选型等
  - 部署架构的演进、更改等

  

##### 迭代工作流

将分析迭代和交付迭代组合在一起，我们就能够得到如下的完整工作流：

{% asset_img 1.jpg %}

以上述方式，分析迭代正好比交付迭代早一个周期，整个工作流就可以按照类似流水线的方式并行运转。

不过分析迭代和交付迭代之间并不是割裂的，虽然明确了产出物，但仅仅靠产出物的交付，是难以完整的传递知识的。因此在实际当中，我们所说的 IPM 和 IKM 的流程，就会作为分析团队和交付团队的重合工作区，再此期间大家共同协作来传递业务知识，产出故事卡，并对故事卡进行 kick off，正如下图所示：

{% asset_img 2.jpg 500 %}



### 总结

本文就项目交付中由于 Inception 不充分而产生交付风险的问题，讨论了 Inception 的局限性，并给出了一种可能的改进方法，即除了 Inception 之外，将分析设计的工作分散到每一个交付迭代之前，与交付有机结合的方法。

此外，基于分析设计迭代这一阶段的引入，对 Inception 本身与分析设计迭代之间的流程、产出内容等方面做了讨论。

最后，给出了分析设计与交付结合起来之后的完整迭代工作流图，解释了通过在 IPM 和 IKM 环节进行重叠，可以更平滑的引入分析设计迭代。
