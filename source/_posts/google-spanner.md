---
title: Spanner：Google 的全球分布式数据库
mathjax: true
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

​        这些特性都基于一个事实即 Spanner 会给事务分配在全球都有意义的提交时间戳，即使该事务可能是分布式的。时间戳反映了串行顺序。进一步的，串行顺序满足了外部一致性（或等价性，线性化 [20]）：如果事务 $T_1$ 在另一个事务 $T_2$ 开始之前提交，那么 $T_1$ 的提交时间戳便小于 $T_2$。Spanner 是首个在全球尺度提供如此承诺的系统。

​        使能这种属性的关键是一个新的 TrueTime API 和其实现。该 API 直接将时钟不确定性暴露出来，而对 Spanner时间戳的承诺依赖于实现提供的边界。假如这种不确定性很大，Spanner 就会停下来等待直到不确定性消失。Google 的集群管理软件提供了 TrueTime API 的实现。这个实现采用多个现代时钟基准（GPS 和原子钟）来确保不确定性很小（通常小于 10ms）。

​        第二节描述了 Spanner 实现的结构、特性集，和设计包含的工程决策。第三节描述了我们的新型 TrueTime API 和其实现的概览。第四节描述 Spanner 是如何使用 TrueTime 来实现外部一致性分布式事务、无锁只读事务以及原子 schema 更新的。第五节提供了一些对 Spanner 性能和 TrueTime 行为的测试，并讨论了 F1 的经验。第六、七、八节描述了相关未来工作，以及对我们结论的总结。

# 2 实现

这一节描述了 Spanner 的结构以及蕴含在实现下的理论基础。之后描述了目录抽象 - 它用于管理复制和局部性，是数据移动的基本单元。最后，描述了我们的数据模型、为什么 Spanner 看起来更像是关系型数据库而不是 k-v 存储，以及应用程序如何控制数据的局部性。

Spanner 的一份部署被称为一个 universe。鉴于 Spanner 全球管理数据的特点，将只有少数个 universe 在运行。目前我们运行了一个 测试/演示 universe，一个开发/生产 universe 和一个仅用于生产的 universe。

Spanner 被组织为一组 zone 集合，每一个 zone 都是一份大致类似 BigTable 的服务器部署 [9]。Zone 是管理部署的单元。这一组 zone 集合也是一组数据可被复制的位置集合。当新的数据中心被引入服务而旧的被关闭时，zone 能够在运行的系统中被添加或移除。Zone 也是物理隔离的单元：如果不同的应用程序的数据必须在同一数据中心的不同的服务器集上进行分区，那么在一个数据中心内就有可能分布一个或多个 zone。

{% asset_img figure-1.png %}

Figure 1 展示了在 Spanner universe 中的服务器。一个 zone 包含一个 zonemaster 和从一百至数千个 spanserver。zonemaster 给 spanserver 分配数据；spanserver 将数据提供给客户端。每个 zone 包含的 location proxy 用于客户端来定位能够提供数据的 spanserver。universe master 和 placement driver 目前都是单实例。universe master 主要作为一个控制台来展示所有 zone 的状态信息以用于交互式调试。placement driver 以分钟为单位处理跨 zone 的数据自动移动。placement driver 周期性的与 spanserver 通信，来寻找需要被移动的数据，以满足复制约束（replication constrain）被更新的情况或是负载均衡。篇幅考虑，我们将只详细描述 spanserver。

## 2.1 Spanserver 软件栈

{% asset_img figure-2.png %}

这一节聚焦于 spanserver 的实现，以说明复制和分布式事务在基于 BigTable 的实现上是如何分层的。Figure 2 展示了软件栈。在最底层，每一个 spanserver 都负责 100 到 1000 个称为 tablet 的数据结构的实例。一个 tablet 类似于 BigTable 的 tablet 抽象，因为它实现了如下映射包（mapping bag）：

`(key:string, timestamp:int64) → string`

与 BigTable 不同，Spanner 会给数据分配时间戳，这使得 Spanner 更像是一个多版本数据库而不只是一个 k-v 存储。一个 tablet 的状态会被保存在类 B-Tree 结构的文件以及写前日志文件中（write-ahead log），所有这些文件都被存放在分布式文件系统 Colossus（Google File System 的继任者 [15]） 内。

为了支持复制，每个 spanserver 都在每一个 tablet 上实现了一个 Paxos 状态机。（在 Spanner 的早期版本中，每个 tablet 能支持多个 Paxos 状态机，这能让复制配置更加灵活。但这种设计的复杂性让我们放弃了它。）每个状态机都将它的元数据以及日志存储在相关的 tablet 上。我们的 Paxos 实现支持基于时间租约的久存活（long-lived）leader，该租约默认为 10 秒。当前的 Spanner 实现会在日志中记录 Paxos 写两次：一次是在 tablet 的日志中，一次是在 Paxos 的日志中。这种选择是一个权宜之计，我们最终很可能会补救它。我们的 Paxos 实现是流水线的，以便在存在 WAN 延迟的情况下提升 Spanner 的吞吐；但是由 Paxos 应用的写入是按顺序的（我们将在第四节中依赖的一个事实）。

Paxos 状态机用于实现对映射包的一致性复制。每个副本的 k-v 映射状态都存储在它对应的 tablet 中。对状态的写操作必须由 Paxos leader 发起；访问状态时则直接从任意足够新的副本中的 tablet 内读取。副本集共同组成一个 Paxos 组。

在每个作为 leader 的副本中，spanserver 都实现了一个 lock table 用于并发控制。该 lock table 包含了两阶段锁（two-phase locking）的状态：他将一定范围的 key 映射到锁状态上。（注意一个久存活的 Paxos leader 对提升 lock table 的效率至关重要。）在 BigTable 和 Spanner 中，我们都是为久存活事务而设计的（例如生成报告，可能会需要以分钟为单位的顺序），这会导致当存在冲突时，乐观并发控制的性能较差。对需要同步的操作，例如事务读，会在 lock table 请求锁，其他按操作则会绕过 lock table。

在每个作为 leader 的副本中，spanserver 也都实现了一个事务管理器来支撑分布式事务。事务管理器用于实现一个 leader 参与者；而组内的其他副本则被称为 slave 参与者。如果一个事务只引入了一个 Paxos 组（大多数事务都是这样），由于 lock table 和 Paxos 组共同提供了事务性，因此会自动绕过事务管理器。假如一个事务引入了多于一个 Paxos 组，这些组的 leader 就会协调进行两阶段提交。其中的某个参与组被选为协调者：该组的 leader 参与者会被成为 leader 协调者，而组内的其他 slave 就成了 slave 协调者。每个事务管理器的状态都保存在其下的 Paxos 组中（所以也会一并被复制）。

## 2.2 目录和放置（placement）

在 k-v 映射包之上，Spanner 实现支持一个称为目录的桶抽象，它是由一组相邻的 key 共享一个通用的前缀。（选择术语目录是一个历史性意外，一个更好的术语应该是桶（bucket）。）我们会在 2.3 节解释前缀的来源。支持目录允许应用程序控制通过仔细的选择 key 来控制数据的位置。

{% asset_img figure-3.png %}

目录是数据放置的单元。一个目录中的所有数据都拥有相同的复制配置。Figure 3 展示了当数据在 Paxos 组之间移动时，是以目录为单位来移动的。Spanner 可能会通过移动一个目录来降低某个 Paxos 组的负载；将经常访问的目录放在同一个组中；或移动一个目录到距离访问者更近的地方。当客户端操作在进行中时也能够移动目录。可以预期一个 50MB 的目录能够在几秒钟内被移动。

一个 Paxos 组可能会包含多个目录的事实隐含了 Spanner 的 tablet 与 BigTable 的 tablet 的不同：Spanner 的 tablet 并不一定是单个按字符顺序分区的行空间。相反，Spanner 的 tablet 是一个可能包含多个行空间分区的容器。我们做出这样的决定是为了能够将经常访问的多个目录放在一起。

`Movedir` 是一个用于在 Paxos 组间移动数据的后台任务 [14]。由于 Spanner 还不支持基于 Paxos 的配置更改，`Movedir`也用于给 Paxos 组添加或移除副本 [25]。`Movedir` 并不被实现为单个事务，因此能它能避免在大量数据移动时阻塞正在进行中的读写。当它移动了所有的数据后，它会使用事务来原子对 “名义上的数据量” 进行移动，并更新被移动双方的 Paxos 元数据。

目录也是应用程序能够指定其地理复制属性（简称放置）的最小单元。我们的放置规范语言（placement-specification language）的设计分离了管理复制配置的责任。管理员可以控制两个维度：副本的数量和类型，以及对这些副本放置的地理位置。他们在这两个维度创建了命名选项的菜单（例如，North America, replicated 5 ways with 1 witness）。一个应用通过给每个数据库和/或独立的目录打标签来控制数据如何被复制，标签的内容就是上述选项的组合。例如，一个应用也许想将每个端用户的数据存储在他自己的目录下，那么用户 $A$ 的数据可以有三个副本在欧洲，而用户 $B$ 的数据可以有五份副本在北美。

为了解释清晰，我们简化了整个流程。实际上，当一个目录变得过大时，Spanner 会将其分片至多个片段（fragment）中。片段可能会从不同的 Paxos 组而来（也即不同的服务器中）。`Movedir` 实际上在组间是移动片段而不是整个目录。

## 2.3 数据模型

Spanner 将如下数据集特性暴露给应用程序：一个基于模式化半关系型表的数据结构，一种查询语言，以及通用事务。为了支持这些特性所做的工作受到了许多因素的推动。对于支持模式化半关系型表和同步复制的需求，Spanner 得到了流行的 Megstore 的推动 [5]。Google 内部至少有 300 个应用程序在使用 Megastore（即使它的性能相对较低），原因是它的数据模型比 BigTable 更容易管理，且支持跨数据中心的同步复制。（BigTable 只支持跨数据中心最终一致性。）著名的 Google 应用使用 Megasotore 的例子有 Gmail、Picasa、Calendar、Android Market 和 AppEngine。鉴于Dremel [28] 作为一种交互式数据分析工具的流行，在 Spanner 中支持类 SQL 查询语言的需求也很明确。最后，由于 BigTable 中跨行事务的缺失导致了频繁的抱怨；Percolator [32] 的构建部分是为了解决这个问题。一些作者声称，由于会引入性能或可用性的问题  [9, 10, 19]，对通用的两阶段提交的支持过于昂贵。我们认为，由于滥用事务而导致瓶颈出现时，最好让应用程序员来处理性能问题，而不是总围绕着缺失事务来编码。在 Paxos 上运行两阶段提交可以缓解可用性问题。

应用数据模型层建立在实现所支持的目录-桶 k-v 映射上。一个应用能在 universe 中创建一个或多个数据库。每一个数据库都能包含无限数量的模式化表。表看起来就像是关系型数据库的表一样，有行、列和有版本的值。我们不会深入讲解 Spanner 的查询语言。它与 SQL 很像，且包含了一些扩展来支持协议缓冲区值字段。

Spanner 的数据模型不是纯关系型的，因为行必须要有名称。更准确的讲，每个表都要求要有一个由一个或多个主键列组成的有序集合。这种要求使得 Spanner 仍然起来像 k-v 存储：主键构成了行的名称，每个表都定义了一个从主键列到非主键列的映射。只有当行的键（row's key）定义为某些值（即使它是 NULL）时，该行才存在。施加这种结构非常有用，因为这使应用程序通过选择 key 值来控制数据的位置。

{% asset_img figure-4.png %}

Figure 4 包含了一个 Spanner 为每一个用户及每一个相册存储图片元数据模式的例子。模式语言与 Megastore 类似，除此之外，每个 Spanner 数据库都要求必须由客户端划分为一个或多个表层次结构。客户端应用程序通过 `INTERLEAVE IN` 在数据库声明层级结构。层级之上的表是一个目录表。目录表中键为 `K` 的每一行，以及派生表中按字典顺序以`K`开头的所有行，构成一个目录。`ON DELETE CASCADE`是说当目录表中的一行被删除时，任何关联的子行也会被删除。该图也展示了示例数据库的交错布局：`Albums(2,1) ` 表示`user id 2`,`album id 1` 的 `Albums` 表中的行。这种将表交错以形成目录的方法非常重要，因为这允许客户端来描述多个表之间的位置关系，这对于在分片的、分布式数据库中获得良好性能是必要的的。如果没有这种方法，Spanner 就无法知道最重要的位置关系了。

# 3 TrueTime 

{% asset_img table-1.png %}

这一节描述了 TrueTime API 和其实现的概要。我们将更多的细节留给另一篇文章：我们的目的是展示拥有这种 API 的能力。Table 1 列出了该 API 的方法。TrueTime 显式将时间表示为一个 $TTinterval$ ，即一个包含有界时间不确定性的时间间隔（与标准时间接口不同，标准时间接口没有给客户带来不确定性这一概念）。$TTinterval$ 的端点来源于 $TTstamp$ 类型。$TT.now()$ 方法返回一个 $TTinterval$，该结果能保证包含在 $TT.now()$ 被调用时的绝对时间。该时间纪元类似于带有闰秒补偿（leap-second smearing）的 UNIX 时间。将瞬时误差边界定义为 $ε$，即区间宽度的一半，平均误差边界为 $\bar{ε}$。$TT.after()$ 和 $TT.before()$ 是对 $TT.now()$ 包装而成的简便方法。

用函数 $t_{abs}(e)$  表示事件 $e$ 的绝对时间。更正式的表述是，TrueTime 确保对于一次  $tt = TT.now()$ 的调用，$tt.earliest ≤ t_{abs}(e_{now}) ≤ tt.latest$ ，$e_{now}$ 是调用事件。

TrueTime 底层的时间参考是 GPS 和原子钟。TrueTime 使用两种时间参考的原因是它们有不同的故障模式。GPS 参考源漏洞包括天线和接收器错误，本地无线电干扰，相关性故障（比如，类似错误的闰秒处理及欺骗的设计错误），以及 GPS 系统中断故障。原子钟可能会以与 GPS 不相关的方式失效，在很长的一段时间内，由于频率误差，原子钟会产生显著的漂移。

TrueTime 由每个数据中心一组的 time master 和每台机器一个的 time slave 守护程序组成。大多数 master 都带有专用天线的 GPS 接收器；这些 master 被物理分离，以减小天线失效、无线电干扰与欺骗的影响。其余的 master（我们称之为世界末日（Armageddon）master）配备了原子钟。一台原子钟并没有那么贵：一个世界末日 master 的花费与一个 GPS master 处于同一数量级。所有 master 的时间参考会定期互相对比。每个 master 也会交叉检查它的时间参考与它本地时间前进的速率，如果存在较大的分歧，则会将自己移除。在同步间，世界末日 master 会显示出源于保守的最坏情况的时钟漂移而缓慢增加的时间不确定。GPS master 则显示不确定性通常接近于零。

每个守护程序都会从多个 master 轮询数据 [29] 来降低任意 master 错误的可能性。一部分是从附近数据中心选取的 GPS master；其余的是较远的数据中心的 GPS master，还有一些是世界末日 master。守护程序应用 Marzullo 算法的一个变体来探测并拒绝假信息，并将本地机器时钟与正确的 master 同步。为了防止损坏本地时钟，那些显示频率偏移大于组件规格和操作环境的最坏情况边界的机器将被移除。

在同步间隔中，一个守护程序显示为缓慢的增加时间不确定性。$ε$ 源于保守应用的最坏情况下的本地时钟漂移。$ε$ 也依赖于 time master 的不确定性和与 time master 通信的延迟。在我们的生产环境，$ε$ 通常是时间的锯齿函数，在每个轮询间隔内，从 1 到 7ms 不等。 因此 $\bar{ε}$ 在大多数情况下是 4ms。守护程序的轮询间隔目前是 30 秒，且目前应用的漂移率被设定为 200us/s， 这两个值共同构成了 0 到 6ms 的锯齿边界。剩下的 1ms 是从与 time master 的通信延迟而来。当出现故障时，这种锯齿可能会有偏差。例如，time master 的偶尔不可用会导致数据中心范围的 $ε$ 的增加。同样的，过载的机器和网络连接也会导致 $ε$ 的局部峰值。

# 4 并发控制
这一节描述 TrueTime 是如何用于保证并发控制属性的正确性，以及这些属性是如何用于实现类似外部一致性事务、无锁只读事务以及旧数据非阻塞读等特性的。这些特性能够使能例如保证在时间戳 $t$ 下的全数据库审计读能看到截止 $t$ 时提交的所有事务。

进一步的，区分 Paxos 能看到的写（除非上下文清晰，否则我们称为 Paxos 写）与 Spanner 客户端产生的写是很重要的。例如，两阶段提交会在准备阶段生成一个 Paxos 写，而这个写操作与 Spanner 客户端写并没有关系。

## 4.1 时间戳管理
{% asset_img table-2.png %}

表 2 列出了Spanner 支持的操作类型。Spanner 支持了读写事务、只读事务（预声明快照隔离事务），以及快照读。单机写被实现为读写事务；非快照单机读被实现为只读事务。二者都会内部重试（客户端不需要再编写他们自己的重试循环）。
只读事务是一种具有快照隔离 [6] 性能优势的事务。只读事务必须提前声明不包含任何写；并不能简单认为它是一种不包含写操作的读写事务。只读事务中的读会在系统选择的时间戳下无锁的执行，因此不会阻塞到来的写操作。在只读事务中执行的读操作可以在任意满足最新数据的副本上执行（4.1.3 节）。
快照读是在过去的数据上执行无锁的读。客户端可以为快照读指定一个时间戳，也可以提供一个期望的过期时间戳的上界来让 Spanner 选择时间戳。这两种情况下快照读都可以在任意满足最新数据的副本上执行。
无论是只读事务还是快照读，一旦选定了时间戳，提交就不可避免，除非是在该时间戳上的数据已经被垃圾回收了。因此，客户端可以避免在一个重试循环中暂存结果。当一个服务器失效时，客户端会在内部通过重复时间戳和当前读位置来继续将该操作执行在一个不同的服务器上。

### 4.1.1 Paxos Leader 租约
Spanner 的 Paxos 实现使用定时租约来实现 leader 久存活（默认10秒）。一个潜在的leader发送定时租约投票请求；在收到额定数量的租约投票时，leader就能知道它拥有了租约。在一次成功的写之后，副本会隐含的自动延长该租约投票，且当租约投票将要到期时，它会请求延长租约投票。当某个副本探测到它拥有了额定的租约投票数量时，就可定义一个 leader 租约期的开始，而当它不再拥有额定的租约投票时（因为有一些投票过期了）可定义 leader 租约期结束。Spanner 依赖如下离散不变性：对每个 Paxos 组，每个 Paxos leader 的租约期与其他 leader 的租约期不相交。附录 A 描述了这种不变性是如何强制保证的。
Spanner 的实现允许一个 Paxos leader 通过释放 slave 的的租约投票来放弃 leader 角色。为了保持离散不变性，Spanner 会在放弃 leader 被允许时做出约束。定义 $S_{max}$ 为 leader 使用的最大时间戳。后续章节会描述何时  $S_{max}$ 会增加。在放弃 leader 之前，leader 必须等待 $TT.after( S_{max})$ 为 True。

### 4.1.2 为读写事务分配时间戳
事务读写使用两阶段锁。因此，可以在所有锁都被获取之后，以及任意锁被释放之前的任何时间点分配时间戳。对于给定的事务，Spanner 会给它分配 Paxos 分配给代表事务提交的 Paxos write 操作的时间戳。

Spanner 依赖下述单调不变性：在每个 Paxos 组中，Spanner 都使用单调递增的顺序给 Paxos write 分配时间戳，哪怕是跨 leader 的情况。一个单 leader 的副本可以简单的按单调递增分配时间戳。这种不变性通过使用如下离散不变性来强制跨 leader：一个 leader 必须仅在它的租约期内分配时间戳。注意无论何时一个时间戳 $s$ 被分配，$S_{max}$ 都会被提升至 $s$ 以此来保持离散性。

Spanner 也强制如下外部一致不变性：如果一个事务 $T_2$ 在事务 $T_1$ 提交后开始，那么 $T_2$ 的提交时间戳必须要大雨 $T_1$ 的提交时间戳。定义事务 $T_i$ 的开始和提交事件为 $e^{start}_i$ 和 $e^{commit}_i$，$T_i$ 的提交时间戳为 $s_i$。该不变性变为 $t_{abs}(e^{commit}_1) < t_{abs}(e^{start}_2) ⇒ s_1 < s_2$。执行事务和分配时间戳的协议服从两个规则，这两个规则会共同确保该不变性，详情下述。定义协调者 leader 的写事务 $T_i$ 的提交请求到达事件为 $e^{server}_i$。

**Start** 协调者 leader 给写事务 $T_i$ 分配一个不比 $TT.now().latest$ 的值小的时间戳 $s_i$，该时间戳在 $e^{server}_i$ 之后计算。注意参与者 leader 在这里并不相关；4.2.1 节将会描述它们是如何在第二条规则内引入的。

**Commit Wait** 协调者 leader 确保客户端在 $TT.after(s_i)$ 为 true 之前不会看到任何由 $T_i$ 提交的数据。提交会等待确保 $s_i$ 小于 $T_i$ 的绝对提交时间，或 $s_i < t_{abs}(e^{commit}_i)$。Commit Wait 的实现描述在4.2.1节。证明：

$s_1 < t_{abs}(e^{commit}_1)$  (commit wait)

$t_{abs}(e^{commit}_1) < t_{abs}(e^{start}_2)$  (assumption)

$t_{abs}(e^{start}_2) ≤ t_{abs}(e^{server}_2)$  (causality)

$t_{abs}(e^{server}_2) ≤ s_2$  (start)

$s_1 < s_2$  (transitivity)


### 4.1.3  以时间戳读
在4.1.2节描述的单调递增不变性允许Spanner正确确定副本的状态是否足够新以满足读取要求。每个副本都会跟踪一个称为安全时间点 $T_{safe}$ 的值，该值是副本最新的最大时间戳。一个副本可以满足一个在时间戳 $t < T_{safe}$ 的读取请求。定义 $t_{safe} = min(t^{Paxos}_{safe} , t^{TM}_{safe})$， 每一个Paxos 状态机都包含一个安全时间点 $t^{Paxos}_{safe}$，每个事务管理器都包含一个安全时间点 $t^{TM}_{safe}$。$t^{Paxos}_{safe}$更简单：它是应用做多的Paxos写的时间戳。由于时间戳单调递增切写按顺序执行，因此写不会在 $t^{Paxos}_{safe}$ 或其之后发生。

当存在零个准备（但未提交）事务 -- 即两阶段提交中两个阶段之间的事务 -- 时，$t^{TM}_{safe} = ∞$。（对于参与者 slave，$t^{TM}_{safe}$ 实际上指的是副本leader的事务管理器，slave 可以通过传递的元数据来推断其状态）。如果有任何这种事务，那么受到这些事务影响的状态是不确定的：一个参与者副本并不知道这种事务是否要提交。就像我们在4.2.1节讨论的，提交协议确保每一个参与者都知道准备好的事务的时间戳下界。每个参与者leader（对组 $g$）会给事务 $T_i$ 为其准备阶段记录分配一个准备阶段时间戳 $s^{prepare}_{i,g}$。协调者leader确保该事务的提交时间戳在所有参与组 $g$ 之间都满足 $s_i >= s^{prepare}_{i,g}$ 。因此对于每一个在组 $g$ 中的副本，在 $g$ 准备的所有事务 $T_i$ 上，$t^{TM}_{safe} = mini(s^{prepare}_{i,g}) − 1$ 在 $g$ 准备的所有事务上。

### 4.1.4 为RO事务分配时间戳
一个只读事务分两个阶段执行：分配一个时间戳 $s_{read}$ [8]，之后以快照读在 $s_{read}$ 来执行事务读。快照读可以在任何满足最新时间戳要求的副本进行。

在事务执行的任何时候，$s_{read} = TT.now().latest$ 的简单赋值通过一个类似4.1.2节中的写操作的参数来确保外部一致性。然而，假如 $t_{safe}$ 不足够提前，这种时间戳也许需要在 $s_{read}$ 读数据的执行时阻塞。（此外，注意选择一个 $s_{read}$ 的值也有可能比 $s_{max}$ 提前来保持不连续性。）为了降低被阻塞的概率，Spanner 应该分配最旧的时间戳来保持外部一致性。4.2.2节解释了这种时间戳需要如何选择。

## 细节
这一节解释了前文省略的读写事务与只读事务的实用细节，以及用于实现原子schema变更的特殊事务类型的实现。之后描述了一些基本方案的改进。

### 4.2.1 读写事务
像 BigTable一样，发生在事务内的写会缓存在客户端直到提交时。因此，事务内的读不会看到事务内写的效果。这种设计在Spanner工作的很好，原因是读会返回被读取数据的时间戳，而未提交的写也并没有被分配时间戳。

读写事务内的读使用 wound-wait [33] 来防止死锁。客户端向合适组的 leader 副本发出读操作，这会获取读锁且读最新的数据。当客户端事务保持开启时，他会发送keepalive信息来防止事务在参与者 leader 上被超时。当一个客户端已经完成了所有的读，并缓存了所有的写后，它开始进行两阶段提交。该客户端选择一个协调者组并发送包含协调者以及任何被缓存写信息的提交信息给每一个参与者leader。让客户端来驱动两阶段提交能避免在大范围的连接上发送两次数据。

一个非协调者、参与者leader会先获取写锁。之后选择一个准备时间戳，该时间戳必须比任何它给先前事务分配过的时间戳要大（为了保证单调递增），然后将准备记录通过 Paxos 进行log。之后每个参与者会通知该协调者它们的准备时间戳。

协调者leader也会先获取写锁，但是跳过准备阶段。它在收取到所有其他参与者leader的时间戳之后，选择一个作为整个事务的时间戳。该提交时间戳 $s$ 必须大于或等于所有的准备时间戳（为了满足4.1.3节所述的约束），在协调者收到它们的提交消息时大于 $TT.now().latest$ ，且大于任何该leader曾经分配给先前事务的时间戳（还是为了保证单调递增）。协调者leader之后将提交记录通过 Paxos 进行log（或者在等待其他参与者时由于超时而被中断）。

在允许任何协调者副本应用提交记录之前，协调者 leader 等待直到 $TT.after(s)$，以便遵循在4.1.2节描述的 commit-wait 规则。由于协调者 leader 基于 $TT.now().latest$ 来选择 $s$，现在要等到该时间戳确保成为过去，所以期望的等待时间至少是 $2 ∗ \bar{ε}$。这个等待通常会与和 Paxos 通信的时间相重合。在提交等待之后，协调者发送提交时间戳给客户端和其他所有的参与者 leader。每个参与者 leader 都将事务的结果通过Paxos进行log。所有参与者都在相同的时间戳应用事务之后释放锁。

### 4.2.2 只读事务
分配一个时间戳需要一个在所有涉及到读的 Paxos 组之间的协商阶段。所以，Spanner 的每个只读事务都需要一个作用域表达式，来汇总所有将要在整个事务中被读取到的key。Spanner 会自动推断某个独立查询的作用域。

假如作用域内的值只由单个 Paxos 组来提供，那么客户端会向该组的 leader 发起只读事务。（目前的 Spanner 的实现仅为 Paxos leader 上的只读事务选择一个时间戳。）该 leader 分配 $s_{read}$ 并执行该读取。对单个站点的读而言，Spanner通常会比 $TT.now().latest$ 做得更好。定义 $LastTS()$ 为 Paxos 组中最后提交的时间戳。如果没有准备好的事务，赋值 $s_{read} = LastTS()$ 一般能满足外部一致性：事务能够看到最近的写的结果，因此会在其顺序之后。

假如作用域内的值由多个 Paxos 组提供，那么会有几种选择。最复杂的选择是在所有 Paxos 组的 leader 之间轮流通信来基于 $LastTS()$ 协商 $s_{read}$。目前 Spanner 的实现选择了一个更简单的选项。客户端避免轮流协商，仅将读操作在 $sread = TT.now().latest$ 时执行（也许会等待安全时间过后）。在事务内的所有读都可以被发给副本来满足更新。

### 4.2.3 Schema 变更事务
TrueTime 使 Spanner 能支持原子的 schema 变更。由于参与者的数量（数据库内组的数量）可能会有上百万，因此不可能使用标准事务。Bigtable 支持在单个数据中心内原子的进行 schema 变更，但它会阻塞所有操作。

Spanner 的 schema 变更事务是一个通用的标准事务非阻塞变量。首先，在准备阶段，它显式的分配一个将来的时间戳。因此，跨数千个服务器的 schema 变更能在对其他并发的动作造成最小中断的情况下完成。之后，隐含依赖该 schema 的读写操作，与任意已注册的 schema 变更时间戳 $t$ 相同步：当它们的时间戳先于 $t$ 时它们可能会执行，但是假如他们的时间戳晚于 $t$ 时，它们必须阻塞等待 schema 变更事务。如果没有 TrueTime，定义 schema 变更在 $t$ 时发生将会没有意义。

### 4.2.4 改进
上述定义的 $t^{TM}_{safe}$ 存在弱点，即单个准备好的事务会阻碍 $t_{safe}$ 前进。因此，在之后的时间戳中不会有读操作发生，即使读操作与该事务并不冲突。这种错误的冲突可以通过从 key range 到准备好的事务时间戳的细粒度映射以扩大 $t^{TM}_{safe}$ 来消除。这类信息可以保存在 lock 表中，lock 表已经将 key range 与锁元信息进行了映射。当读到来时，它只需要对 key range 的细粒度安全时间进行检查来确保没有读冲突。

上述定义的 $LastTS()$ 也存在类似的弱点：假如一个事务刚刚提交，一个非冲突的只读事务也必须被分配 $S_{read}$ 以便跟踪该事务。因此，读操作的执行可能会被延迟。这个弱点也可以通过从 key range 到准备好的事务时间戳的细粒度映射以扩大 $LastTS()$ 来解决。（目前我们还没有实现该优化。）当一个只读事务到来时，可以通过为事务冲突的 key range 取 $LastTS()$ 的最大值来分配时间戳，除非与准备好的事务有冲突。（可以通过细粒度的安全时间来确定。）

上述定义的 $t^{Paxos}_{safe}$ 存在弱点，它无法在缺失 Paxos 写的情况下前进。即一个在 $t$ 的快照读无法在 Paxos 组的最后一次写发生在 $t$ 之前的情况下执行。Spanner 通过 leader 租约之间的不连续性解决了这个问题。每个 Paxos leader 通过保持一个将来将要发生的写时间戳来保证 $t^{Paxos}_{safe}$ 的前进：他维护了一个 Paxos 序列号 $n$ 到可能会被分配给 Paxos 序列号 $n+1$ 的最小时间戳的映射 $MinNextTS(n)$。当一个副本应用过 $n$ 后，可以将 $t$ 推进至 $MinNextTS(n) − 1$。

单个 leader 可以很容易的执行 $MinNextTS()$ 承诺。因为 $MinNextTS()$ 承诺的时间戳在一个 leader 租约期限内，不连续的不变量在 leader 之前强制 $MinNextTS()$ 承诺。假如一个 leader 想要在租约结束之后推进 $MinNextTS()$ ，它必须先扩展它的租约。注意 $S_{max}$ 总是比 $MinNextTS()$ 的最新值更新以保持离散性。

leader 默认每 8 秒推进一次 $MinNextTS()$。因此在准备好的事务缺失时，最坏情况下，空闲 Paxos 组中健康的 slave 可以在大于 8 秒的时间戳上执行读取。

# 5 评估
我们首先测量了关于 Spanner 的复制、事务、可用性的性能。之后我们提供了一些 TrueTime 的行为数据，以及一个场景来研究我们的首个客户，F1。

## 5.1 微基准测试
{% asset_img table-3.png %}

Table 3 展示了一些对 Spanner 的微基准测试。这些测量值都是在共享时间的机器上完成的：每个 spanserver 都运行在 4GB RAM + 4 core（AMD Barcelona 2200MHz）的调度单元上。客户端在分立的机器上运行。每个 zone 包含一个 spanserver。客户端和 zone 被放置在一组网络距离小于 1ms 的数据中心中。（这种布局应该很常见：大多数应用斌不需要将他们的数据分布在全世界。）测试数据库被创建为50 个 Paxos 组，2500 个目录。操作是单独的 4KB 读写。压缩之后，所有的读都会耗尽内存，因此我们只测量 Spanner 的调用栈的开销。此外，最开始会执行一轮不测量的读来对各个地方的缓存进行预热。

对于延迟实验，客户端发起足够少的操作来避免服务端排队。对于单副本实验，提交等待大约 5ms，Paxos 延迟大约 9ms。随着副本数量的增加，延迟大致保持不变，标准差较小。这是因为 Paxos 会在副本之间并行的执行。随着副本数量的增加，在一个 slave 副本上，完成一次 quorum 的延迟对减慢程度不敏感。

对于吞吐实验，客户端发起足够多的操作来让服务器的 CPU 饱和。快照读可以在任意足够新的副本上执行，因此它的吞吐量随着副本数的增加接近线性增长。单个读的只读事务由于 leader 必须要分配时间戳因此只会在 leader 执行。只读事务的吞吐量随着副本数的增加而增加，原因是有效 spanserver 的数量也在不断增加：在实验环境下，spanserver 的数量与副本数量一致，且 leader 会随机的分布在 zone 中。写吞吐受益于相同的实验工件（解释了为何副本数从3到5的吞吐增长），但随着副本数量的增加，每次写所执行的工作量呈线性增加，这远远超出了前述的益处。

{% asset_img table-4.png %}

Table 4 展示了两阶段提交能扩大到一个合理的参与者数量：它总结了一组跨 3 个 zone 运行的的实验，每个区域有 25 个 spanserver。从平均值和第99个百分位来看，将参与者增加到50个是合理的，并且在100个参与者时，延迟开始显著增加。

## 5.2 可用性
{% asset_img figure-5.png %}

Figure 5 展示了可用性能受益于在多个数据中心执行 Spanner。 它显示了在同一时间标度下的三种数据中心故障实验对吞吐量的影响。测试 universe 由 5 个 zone $Z_i$ 组成，每一个都包含 25 个 spanserver。测试数据库分片为 1250 个 Paxos 组，100 个客户端以 50K reads/second 的聚合速率不断地发起非快照读。所有的 leader 都被显式的指定在 $Z_1$。在每次测试开始 5 秒后，一个 zone 的所有服务器被 kill 掉：non-leader 实验 kill $Z_2$；leader-hard 实验 kill $Z_1$；leader-soft 实验 kill $Z_1$，但它先会给所有能够交接领导权的服务器发出提醒。

kill $Z_2$ 对读吞吐没有影响。当给 leader 时间来交接领导权到其他 zone 后 kill $Z_1$，仅造成了很小的影响：吞吐量的下降在图中难以察觉，但实际是大约 3-4%。另一方面，在没有警告的前提下直接 kill $Z_1$ 会造成严重的影响：完成率降低到接近 0，随着 leader 被重选，系统的吞吐量以大约 100K reads/second 的速率上升，这源于实验中的两个要素：系统有额外的容量，且操作会在 leader 不可用时排队。因此，系统的吞吐量会先上升，然后再以稳态速率稳定下来。

我们同时能看到 Paxos leader 的租约期设置为 10 秒所造成的影响。当我们 kill zone 的时候，组的 leader 租约过期时间应该均匀的在接下来的 10 秒内分布。不久后每个宕机 leader 的租约会过期，新的 leader 被选出。在 kill 过后大约 10 秒，所有的组都会再次拥有 leader，吞吐量恢复正常。更短的租约期能够在可用性上降低服务器宕机造成的影响，但是更频繁的租约刷新会导致增加大量的网络开销。我们正在设计开发一种机制，能够让 leader 失效时 slave 自动释放 Paxos leader 租约。

## 5.3 TrueTime
关于 TrueTime，有两个问题必须要回答：$ε$ 真的是时钟不确定性的边界吗，$ε$ 到底有多糟糕？对于前者，最严重的问题是，假如本地时钟的漂移超过了 200us/sec：这会破坏 TrueTime 的假设。我们的机器统计数据显示 CPU 故障的可能性是时钟的 6 倍。即相比于其他严重的硬件故障，时钟问题发生的频率极低。因此，我们相信 TrueTime 的实现能够像 Spanner 的其他依赖软件一样值得信赖。

{% asset_img figure-6.png %}

Figure 6 展示了在距离最多 2200km 的数据中心之间的数千个 spanserver 中获取的 TrueTime 数据。它绘制了 90，99 和 99.9 分位的 $ε$，在 timeslave 守护程序从 timemaster 拉取时间后立即从  timeslave 中采样。由于本地时钟的不确定性，采样忽略了 $ε$ 中的锯齿。因此是 timemaster 的不确定性（通常为 0）与通信延迟之和。

数据显示前述两种因子在确定基准值时通常不是个问题。然而，它们可能会产生显著的尾部延迟问题而导致 $ε$ 的值升高。从3月30日开始，由于网络改进暂时降低了网络连接拥塞使得从3月30日开始尾部延迟有所降低。4月13日大约一小时的升高，是因为例行维护而关闭了两台 timemaster 导致的。我们将继续调查并消除 TrueTime 峰值的原因。

## 5.4 F1
2011 年初，作为重写的 Google 广告后端 F1 [35] 的一部分，Spanner 开始在生产负载下进行实验性评估。该后端最初是基于 MySQL 的，并采用多种方式进行手动分片。未压缩的数据集有数十 TB 大，虽然对许多 NoSQL 实例来说这些数据并不多，但它已经足够大到使分片过的 MySQL 时造成困难。MySQL 会将 schema 分配给每个用户，且所有相关的数据都是分片固定的。这种布局允许在每个用户的基础上使用索引和复杂查询，但需要了解应用程序业务逻辑的分片。随着用户及其数据的增加，当在这种收入关键型数据库上进行重分片的成本非常高。最后一次重分片在两年时间内花费了巨大的努力，涉及了数十个团队的协调和测试以最小化风险。这种操作对于日常维护太过复杂：因此，团队需要通过将一些数据存储在 Bigtable 上来限制 MySQL 上的数据增长，但这损害了事务行为和跨数据查询的能力。

F1 团队选择使用 Spanner 有如下几个原因。其一，使用 Spanner 不在需要人工分片。其二，Spanner 提供了同步复制和自动故障转移功能。以 MySQL 的主从模式，故障转移很困难，会有数据丢失和一段时间不可用的风险。其三，F1 要求强一致性事务语义，这让其他 NoSQL 系统变得不切实际。应用程序语义需要一致性读和跨任意数据事务。F1 团队需要在数据上支持二级索引（由于 Spanner 还不能自动支持二级索引），且能够通过 Spanner 事务来实现它们自己的一致性全局索引。

所有应用程序的写入现在都默认从 F1 透传至 Spanner，这代替了基于 MySQL 的应用程序栈。F1 在美国西海岸拥有两个副本，东海岸有三个。副本站的选择一方面是为了对付潜在的重大自然灾害，另一方面也与它们的前端站点有关。有趣的是，Spanner 的自动故障转移对 F1 来说近乎是透明的。尽管在过去的几个月里发生了一些意外的集群故障，但是 F1 团队需要做的最多的就是是升级它们的数据库 schema 来告诉 Spanner 优先把 Paxos leader 放在哪里，以期能让这些 leader 在它们的前端移动时与之保持最近。

Spanner 的时间戳语义让 F1 能更高效的通过数据库状态来计算的内存数据结构。F1 维护了一个逻辑上的修改历史日志，该日志作为每个事物的一部分也被写入了 Spanner。F1 以时间戳获取全量快照，来初始化其数据结构，之后读取增量数据来更新该数据结构。

{% asset_img table-5.png %}

Table 5 展示了 F1 的每个目录分片数量的分布情况。每个目录通常都会与一个在 F1 应用程序栈之上的客户相关。绝大多数目录（以及其对应的客户）仅由一个分片组成，这意味着对这类客户数据的读写能确保只在单个服务器上进行。多于 100 个分片的目录用于存放F1二级索引的表：一次写入此类表的多个分片是非常罕见的。F1 团队只在事务内进行未经调整的大容量数据加载时遇到过这种行为。

{% asset_img table-6.png %}

Table 6 展示从 F1 服务器测量的 Spanner 的操作延迟。在选择 Paxos leader 的时候，东海岸数据中心的副本拥有更高的选择优先级。表中的数据是从这些数据中心的 F1 服务器中测量的。在写延迟中较大的标准偏差是由于锁冲突而导致的相当大的尾部造成的。导致在读延迟中更大的标准偏差的部分原因是 Paxos leader 分布在两个数据中心间，且只有一个数据中心的机器拥有 SSD。此外，测量包含系统中来自两个数据中心的每次读取：读取字节的平均值和标准差分别为 1.6KB 和 119KB。

# 6 相关工作

Megastore [5] 和 DynamoDB [3] 作为存储服务已经提供了跨数据中心一致性复制。DynamoDB 给出了一个 key-value 接口，仅在 region 内部复制。Spanner 跟随 Megastore 提供了半关系型数据模型，甚至是类似的 schema 语言。Megastore 未实现高性能。处于 Bigtable 层次之上，这回带来很高的通信成本。它也不支持久存活 leader：有多个副本都可能发起写。来自不同副本的所有写操作都不可避免的会在 Paxos 协议中发生冲突，即使他们在逻辑上不存在冲突：Paxos 组的吞吐量以每秒几次写的速度崩溃。Spanner 提供了高性能、通用事务，和外部一致性。

Pavlo 等人 [31] 对比了数据库与  MapReduce [12] 的性能。他们指出，为探索分布式键值存储 [1,4,7,41] 上的数据库功能而进行的其他几项努力证明了这两个世界正在融合。我们同意这一结论，但是证明集成多个层有它的优点：例如，集成并发控制和复制可以减少 Spanner 中等待提交的成本。

在复制的存储之上构建事务分层的概念可以追溯至Gifford 的论文中 [16]。Scatter [17] 是一个事务分层在一致性复制之上的基于 DHT 的 key-value 存储。Spanner 专注于提供一个比 Scatter 更高层的接口。Gray and Lamport [18] 描述了一个基于 Paxos 的非阻塞提交协议。他们的协议会比两阶段提交产生更多消息传低成本，这回加重在宽范围分布的组上提交的成本。Walter [36] 提供了一种快照隔离的变体，可在数据中心内而不是跨数据中心工作。相反的，我们的只读事务提供了一个更自然的语义，因为我们所有的操作都支持外部一致性。

最近有大量降低或消除锁开销的工作。Calvin [40] 消除了并发控制：它会预先分配时间戳之后按照时间戳顺序执行事务。HStore [39] 和 Granola [11] 都分别支持了它们自己的事务类型分类，有一些能够避免锁。但这些系统没有一个能提供外部一致性。Spanner 通过提供快照隔离的支持来解决争用问题。

VoltDB [42] 是一个分片内存数据库，支持在广域范围内进行主从复制来进行灾难恢复，但不支持更通用的复制配置。它是一个被称为 NewSQL 的例子，是由市场推动的支持伸缩的 SQL [38]。大把商业数据库实现了读过去的数据，例如 MarkLogic [26] 和 Oracle’s Total Recall [30]. Lomet and Li [24] 描述了对这种有时态数据库的实现策略。

Farsite 导出了与可信时钟基准相关的时钟不确定性（比 TrueTime 宽松得多）界限：Farsite 服务器维护租约使用了与 Spanner 维护 Paxos 租约类似的方式。在之前的工作中，松散同步的时钟被用于并发控制 [2, 23]。我们展示了 TrueTime 允许一个关于跨 Paxos 状态机集合的全局时间的原因。

# 7 未来的工作

去年我们花了绝大多数时间与 F1 团队一起工作来将 Google 的广告后端从 MySQL 迁移到 Spanner。我们积极地改善它的监控和支持工具，以及调优它的性能。此外，我们也致力于提升我们的备份/恢复系统的功能性与性能。我们目前正在实现 Spanner 的模式语言，二级索引的自动维护以及基于负载的自动重分片。更远期的，我们有一些特性计划去考察。并行乐观读也许是一个有追求价值的策略，但初步的试验证明正确实现它并非易事。此外，我们计划最终支持直接对 Paxos 进行配置修改 [22, 34]。

考虑到我们期望许多应用程序能够跨彼此相对更近的数据中心来复制它们的数据，因此 TrueTime 也许会明显的影响到性能。目前我们认为将延迟降低到 1ms 以下并非是不可逾越的障碍。Time-master-query 间隔时间可以降低，更好的石英钟也相对便宜。Time-master 的查询延迟可以通过改善网络技术来降低，甚至可以通过交替的时间分布技术来避免。

最后，有很多地方显然可以改进。即使 Spanner 可以在数个节点之间伸缩，然而由于节点被设计为简单的 key-value 查询，因此节点本地的数据结构在复杂 SQL 查询时的性能还是相对较弱。来自数据库文献中的算法和数据结构能够显著的提升单节点的性能。其次，自动在数据中心之间移动数据以响应客户端负载的变化一直是我们的目标，但为了让该目标变得高效，我们还需要在数据中心之间自动的、协调的迁移客户端应用程序进程的能力。

# 8 总结

总的来说，Spanner 结合并扩展了两个研究社区的观点：从数据库社区引入了一个熟悉的、易用的、半关系型接口、事务以及基于 SQL 的查询语言；从系统社区引入了可伸缩、自动分片、故障容忍、一致性复制、外部一致性以及广域分布式。从 Spanner 开始，我们花了超过五年时间来迭代直到今天的设计和实现。花费如此漫长迭代的一部分原因是我们慢慢意识到 Spanner 应该做的不仅仅是解决全局复制命名空间的问题，还应该关注 Bigtable 所缺少的数据库特性。

我们设计里有一个突出的特点：Spanner 的特性集中最关键的就是 TrueTime。我们已经展示了在时间 API 中具象化时间不确定性使得构建具有更强时间语义的分布式系统成为可能。此外，由于底层系统对时钟不确定性实施了更严格的限制，因此强语义的开销会减少。作为社区，我们不应该再依赖松散的同步时钟和孱弱的时间 API 来设计分布式算法。

# 致谢

许多人都帮助我们完善了这篇论文：我们的引领人 Jon Howell，它超越了自己的职责；匿名审阅人；以及许多 Googler：Atul Adya, Fay Chang, Frank Dabek, Sean Dorward, Bob Gruber, David Held, Nick Kline, Alex Thomson, and Joel Wein。我们的管理层一直非常支持我们的工作和论文的发表，他们是：Aristotle Balogh, Bill Coughran, Urs Holzle, Doron Meyer, Cos Nicolaou, Kathy Polizzi, Sridhar Ramaswany, 和 Shivakumar Venkataraman。

We have built upon the work of the Bigtable and Megastore teams. The F1 team, and Jeff Shute in particular, worked closely with us in developing our data model and helped immensely in tracking down performance and correctness bugs. The Platforms team, and Luiz Barroso and Bob Felderman in particular, helped to make TrueTime happen. Finally, a lot of Googlers used to be on our team: Ken Ashcraft, Paul Cychosz, Krzysztof Ostrowski, Amir Voskoboynik, Matthew Weaver, Theo Vassilakis, and Eric Veach; or have joined our team recently: Nathan Bales, Adam Beberg, Vadim Borisov, Ken Chen, Brian Cooper, Cian Cullinan, Robert-Jan Huijsman, Milind Joshi, Andrey Khorlin, Dawid Kuroczko, Laramie Leavitt, Eric Li, Mike Mammarella, Sunil Mushran, Simon Nielsen, Ovidiu Platon, Ananth Shrinivas, Vadim Suvorov, and Marcel van der Holst.

我们建立在 Bigtable 和 Megastore 团队的工作之上。F1 团队，特别是 Jeff Shute，与我们一同工作，来开发我们的数据模型并极大地帮助了我们追踪性能和正确性的 bug。平台团队，特别是 Luiz Barroso 和 Bob Felderman，帮助实现了 TrueTime。最后，很多 Googler 都曾是我们团队的一员，它们是：Ken Ashcraft, Paul Cychosz, Krzysztof Ostrowski, Amir Voskoboynik, Matthew Weaver, Theo Vassilakis, 和 Eric Veach；还有近期刚刚加入我们团队的人：Nathan Bales, Adam Beberg, Vadim Borisov, Ken Chen, Brian Cooper, Cian Cullinan, Robert-Jan Huijsman, Milind Joshi, Andrey Khorlin, Dawid Kuroczko, Laramie Leavitt, Eric Li, Mike Mammarella, Sunil Mushran, Simon Nielsen, Ovidiu Platon, Ananth Shrinivas, Vadim Suvorov, 以及 Marcel van der Holst。

# 参考文献

[1] Azza Abouzeid et al. “HadoopDB: an architectural hybrid of MapReduce and DBMS technologies for analytical workloads”. Proc. of VLDB. 2009, pp. 922–933. 

[2] A. Adya et al. “Efficient optimistic concurrency control using loosely synchronized clocks”. Proc. of SIGMOD. 1995, pp. 23– 34.

[3] Amazon. Amazon DynamoDB. 2012. 

[4] Michael Armbrust et al. “PIQL: Success-Tolerant Query Processing in the Cloud”. Proc. of VLDB. 2011, pp. 181–192. 

[5] Jason Baker et al. “Megastore: Providing Scalable, Highly Available Storage for Interactive Services”. Proc. of CIDR. 2011, pp. 223–234. 

[6] Hal Berenson et al. “A critique of ANSI SQL isolation levels”. Proc. of SIGMOD. 1995, pp. 1–10. 

[7] Matthias Brantner et al. “Building a database on S3”. Proc. of SIGMOD. 2008, pp. 251–264. 

[8] A. Chan and R. Gray. “Implementing Distributed Read-Only Transactions”. IEEE TOSE SE-11.2 (Feb. 1985), pp. 205–212. 

[9] Fay Chang et al. “Bigtable: A Distributed Storage System for Structured Data”. ACM TOCS 26.2 (June 2008), 4:1–4:26. 

[10] Brian F. Cooper et al. “PNUTS: Yahoo!’s hosted data serving platform”. Proc. of VLDB. 2008, pp. 1277–1288. 

[11] James Cowling and Barbara Liskov. “Granola: Low-Overhead Distributed Transaction Coordination”. Proc. of USENIX ATC. 2012, pp. 223–236.

[12] Jeffrey Dean and Sanjay Ghemawat. “MapReduce: a flexible data processing tool”. CACM 53.1 (Jan. 2010), pp. 72–77. 

[13] John Douceur and Jon Howell. Scalable Byzantine-FaultQuantifying Clock Synchronization. Tech. rep. MSR-TR-2003- 67. MS Research, 2003. 

[14] John R. Douceur and Jon Howell. “Distributed directory service in the Farsite file system”. Proc. of OSDI. 2006, pp. 321–334. 

[15] Sanjay Ghemawat, Howard Gobioff, and Shun-Tak Leung. “The Google file system”. Proc. of SOSP. Dec. 2003, pp. 29–43. 

[16] David K. Gifford. Information Storage in a Decentralized Computer System. Tech. rep. CSL-81-8. PhD dissertation. Xerox PARC, July 1982. 

[17] Lisa Glendenning et al. “Scalable consistency in Scatter”. Proc. of SOSP. 2011. 

[18] Jim Gray and Leslie Lamport. “Consensus on transaction commit”. ACM TODS 31.1 (Mar. 2006), pp. 133–160. 

[19] Pat Helland. “Life beyond Distributed Transactions: an Apostate’s Opinion”. Proc. of CIDR. 2007, pp. 132–141. 

[20] Maurice P. Herlihy and Jeannette M. Wing. “Linearizability: a correctness condition for concurrent objects”. ACM TOPLAS 12.3 (July 1990), pp. 463–492. 

[21] Leslie Lamport. “The part-time parliament”. ACM TOCS 16.2 (May 1998), pp. 133–169. 

[22] Leslie Lamport, Dahlia Malkhi, and Lidong Zhou. “Reconfiguring a state machine”. SIGACT News 41.1 (Mar. 2010), pp. 63– 73. 

[23] Barbara Liskov. “Practical uses of synchronized clocks in distributed systems”. Distrib. Comput. 6.4 (July 1993), pp. 211– 219. 

[24] David B. Lomet and Feifei Li. “Improving Transaction-Time DBMS Performance and Functionality”. Proc. of ICDE (2009), pp. 581–591. 

[25] Jacob R. Lorch et al. “The SMART way to migrate replicated stateful services”. Proc. of EuroSys. 2006, pp. 103–115. 

[26] MarkLogic. MarkLogic 5 Product Documentation. 2012. 

[27] Keith Marzullo and Susan Owicki. “Maintaining the time in a distributed system”. Proc. of PODC. 1983, pp. 295–305. 

[28] Sergey Melnik et al. “Dremel: Interactive Analysis of WebScale Datasets”. Proc. of VLDB. 2010, pp. 330–339. 

[29] D.L. Mills. Time synchronization in DCNET hosts. Internet Project Report IEN–173. COMSAT Laboratories, Feb. 1981. 

[30] Oracle. Oracle Total Recall. 2012. 

[31] Andrew Pavlo et al. “A comparison of approaches to large-scale data analysis”. Proc. of SIGMOD. 2009, pp. 165–178. 

[32] Daniel Peng and Frank Dabek. “Large-scale incremental processing using distributed transactions and notifications”. Proc. of OSDI. 2010, pp. 1–15. 

[33] Daniel J. Rosenkrantz, Richard E. Stearns, and Philip M. Lewis II. “System level concurrency control for distributed database systems”. ACM TODS 3.2 (June 1978), pp. 178–198. 

[34] Alexander Shraer et al. “Dynamic Reconfiguration of Primary/Backup Clusters”. Proc. of USENIX ATC. 2012, pp. 425– 438. 

[35] Jeff Shute et al. “F1 — The Fault-Tolerant Distributed RDBMS Supporting Google’s Ad Business”. Proc. of SIGMOD. May 2012, pp. 777–778. 

[36] Yair Sovran et al. “Transactional storage for geo-replicated systems”. Proc. of SOSP. 2011, pp. 385–400

[37] Michael Stonebraker. Why Enterprises Are Uninterested in NoSQL. 2010. 

[38] Michael Stonebraker. Six SQL Urban Myths. 2010. 

[39] Michael Stonebraker et al. “The end of an architectural era: (it’s time for a complete rewrite)”. Proc. of VLDB. 2007, pp. 1150– 1160.

[40] Alexander Thomson et al. “Calvin: Fast Distributed Transactions for Partitioned Database Systems”. Proc. of SIGMOD. 2012, pp. 1–12. 

[41] Ashish Thusoo et al. “Hive — A Petabyte Scale Data Warehouse Using Hadoop”. Proc. of ICDE. 2010, pp. 996–1005. 

[42] VoltDB. VoltDB Resources. 2012

# 附录 A：Paxos Leader 租约管理	

确保 Paxos leader 租约的不连续性最简单的方法是，只要租约间隔被延长，leader 就发出一个同步 Paxos 写操作。后继 leader 会读取该间隔并等待直到间隔过后。

TrueTime 可用于在不需要额外的日志写入时保证不连续性。潜在的第 $i$ 个 leader 保持从副本 $r$ 租约投票开始时的下限 $v^{leader}_{i,r} = TT.now().earliest$，在 $e^{send}_{i,r}$ （定义为租约请求被 leader 发出的时刻）之前计算得到。每个副本 $r$ 都在租约 $e^{grant}_{i,r}$ 时获取租约，这会在 $e^{receive}_{i,r}$ 之后发生（当副本收到一个租约请求时）；该租约在 $t^{end}_{i,r} = TT.now().latest + 10$ 时终止，该时刻由 $e^{receive}_{i,r}$ 计算得到。一个副本 $r$ 服从**单一投票（single-vote）**规则：在 $TT.after(t^{end}_{i,r}) == true$ 之前它不会再授予其他租约。为了在不同的 $r$ 中保证这一点，在副本授予租约之前，Spanner 会在日志中记录租约的投票；这种日志写入可以在现有的 Paxos 协议日志写入的基础上进行。

当第 $i$ 个 leader 收到投票 quorum 后（$e^{quorum}_i$ 事件），它将自己的租约期计算为 $lease_i = [TT.now().latest, min_r(v^{leader}_{i,r}) + 10]$。当 $TT.before(min_r(v^{leader}_{i,r}) + 10)  == false$ 时，leader 上的租约被视为过期。为了证明不连续性，我们利用了第 $i$ 个和第 $(i + 1)$ 个 leader 在它们的 quorum 中必须有一个共同的副本这一事实。我们称该副本为 $r0$。证明：

$lease_i.end = min_r(v^{leader}_{i,r}) + 10$  (by definition)

$min_r(v^{leader}_{i,r}) + 10 ≤ v^{leader}_{i,r0} + 10$  (min) 

$v^{leader}_{i,r0} + 10 ≤ t_{abs}(e^{send}_{i,r0}) + 10$  (by definition)

$t_{abs}(e^{send}_{i,r0}) + 10 ≤ t_{abs}(e^{receive}_{i,r0}) + 10$ (causality)

$t_{abs}(e^{receive}_{i,r0}) + 10 ≤ t^{end}_{i,r0}$  (by definition)

$t^{end}_{i,r0} < t_{abs}(e^{grant}_{i+1,r0})$  (single-vote)

$t_{abs}(e^{grant}_{i+1,r0}) ≤ t_{abs}(e^{quorum}_{i+1})$  (causality)

$t_{abs}(e^{quorum}_{i+1}) ≤ lease_{i+1}.start$  (by definition)