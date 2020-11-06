---
title: Spanner: Google 的全球分布式数据库
date: 2020-11-06 17:40:44
tags:
- database
- spanner
categories:
- TiDB
---

> 本文是对 《Spanner: Google’s Globally-Distributed Database》的翻译，原文请见：https://static.googleusercontent.com/media/research.google.com/en//archive/spanner-osdi2012.pdf

# 摘要

​        Spanner 是 Google 的可伸缩、多版本、全球分布且同步复制的数据库。它是首个在全球范围分布数据且支持外部一致性分布式事务的系统。本文描述了 Spanner 的结构、特性集、其多项设计决策下蕴含的理论基础以及一个能够暴露时钟不确定性（clock uncertainly）的新颖的时间 API。该 API 及其实现是作为关键角色来支撑外部一致性和其他许多涉及整个 Spanner 的强大特性，如对过往数据的非阻塞读（nonblocking read in the past），无锁的读事务，原子的 schema 变更等。

# 1 介绍

​        Spanner 是由 Google 设计、构造并部署的可伸缩全球分布式数据库。其最顶层的抽象，可以看做是一个在遍布世界各地的许多 Paxos [21] 集合之间分片数据的数据库。复制用于全局可用性以及地理位置性；客户端自动在不同的副本间故障转移。当数据量或服务器数量发生变化时，Spanner 会自动在机器之间重分片（reshard）数据，并且能在机器之间自动迁移数据（哪怕是跨数据中心）来平衡负载和应对故障。Spanner 被设计为能够在数百个数据中心的数万亿个数据库行之间扩展到数百万台机器。

​        应用程序能够使用 Spanner 来实现高可用，即使面对大面积的自然灾害，它也能靠在洲际内甚至跨洲际来进行数据复制。我们的内部客户是 F1 [35]，一个重写的 Google 广告后端。F1 使用了遍布美国的五个副本。大多数其他的应用大概会将他们的数据复制到 3 到 5 个处于同一个地理区域内数据中心，但故障模式相对独立。即只要他们能够在 1 到 2 个数据中心失效的情况下存活下来，大多数应用都会在低延迟与高可用之间选择低延迟。

​        Spanner 主要聚焦于跨数据中心复制数据的管理，但我们也花了非常多精力在我们的分布式基础设施之上设计和实现重要的数据库特性。即使大多数项目都乐于使用 BigTable [9]，但我们也持续的收到了用户的抱怨：BigTable 难以使用在某些较为复杂、schema 经常演进，或是想要在大面积复制中保持强一致性的应用中。（其他作者也有类似的说法 [37]。）尽管存在相对较弱的写吞吐，Google 的许多应用程序仍然会选择使用 Megastore [5]，因为其半关系型数据模型以及支持同步复制的特性。因此，Spanner 从一个类 BigTable 的有版本 kv 存储（versioned key-value store）进化为一个基于时间版本的多版本数据库。数据存储在 schema 化的半关系型表中；数据有版本，且每一个版本都自动以其提交时间作为时间戳；旧版本的数据受可配置的垃圾收集策略所控制；应用也可以读取旧时间戳上的数据。Spanner 支持通用事务，且提供了一个基于 SQL 的查询语言。

​        作为一个全球分布的数据库。Spanner 提供了许多有趣的特性。首先，对于数据复制的配置可以被应用程序细粒度的动态控制。应用能够指定一些约束（constrains）来控制哪个数据中心包含哪些数据，数据距离用户有多远（来控制读延迟），副本之间有多远（来控制写延迟），以及维护了多少份副本（来控制持久性、可用性和读性能）。为了均衡数据中心之间的利用率，数据能够动态且透明的在不同数据中心之间移动。其次，Spanner 提供了两个在分布式数据库中较难实现的特性：提供外部一致性 [16] 读写，和基于时间戳的跨数据中心全球一致性读。这些特性使得 Spanner 能支持一致性备份，执行一致性 MapReduce [12]，以及原子的 schema 更新，所有这些都是全球尺度，甚至是在正在执行事务时。

​        这些特性都基于一个事实即 Spanner 会给事务分配在全球都有意义的提交时间戳，即使该事务可能是分布式的。时间戳反映了串行顺序。进一步的，串行顺序满足了外部一致性（或等价性，线性化 [20]）：如果事务 T1 在另一个事务 T2 开始之前提交，那么 T1 的提交时间戳便小于 T2。Spanner 是首个在全球尺度提供如此承诺的系统。

​        使能这种属性的关键是一个新的 TrueTime API 和其实现。该 API 直接将时钟不确定性暴露出来，而对 Spanner时间戳的承诺依赖于实现提供的边界。假如这种不确定性很大，Spanner 就会停下来等待直到不确定性消失。Google 的集群管理软件提供了 TrueTime API 的实现。这个实现采用多个现代时钟基准（GPS 和原子钟）来确保不确定性很小（通常小于 10ms）。

​        第二节描述了 Spanner 实现的结构、特性集，和设计包含的工程决策。第三节描述了我们的新型 TrueTime API 和其实现的概览。第四节描述 Spanner 是如何使用 TrueTime 来实现外部一致性分布式事务、无锁只读事务以及原子 schema 更新的。第五节提供了一些对 Spanner 性能和 TrueTime 行为的测试，并讨论了 F1 的经验。第六、七、八节描述了相关未来工作，以及对我们结论的总结。

# 2 实现

​        这一节描述了 Spanner 的结构以及蕴含在实现下的理论基础。之后描述了目录抽象 - 它用于管理复制和局部性，是数据移动的基本单元。最后，描述了我们的数据模型、为什么 Spanner 看起来更像是关系型数据库而不是 k-v 存储，以及应用程序如何控制数据的局部性。

​        A Spanner deployment is called a universe. Given that Spanner manages data globally, there will be only a handful of running universes. We currently run a test/playground universe, a development/production universe, and a production-only universe.

Spanner 的一份部署被称为一个 universe。鉴于 Spanner 全球管理数据的特点，将只有少数个 universe 在运行。目前我们运行了一个 测试/演示 universe，

