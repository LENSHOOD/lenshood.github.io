---
title: 译文: 一种易于理解的共识算法的研究（In Search of an Understandable Consensus Algorithm）
date: 2021-09-13 17:57:50
tags: 
- raft
categories:
- Software Engineering
---

> 本文是对 raft paper 的翻译，原文请见 https://raft.github.io/raft.pdf

## 摘要

Raft 是一种用于管理 log 复制的共识算法。它采用了与 Paxos 不同的算法结构，但能产出与 (multi-)Paxos 等价的结果，并且与 Paxos 一样高效；这令 Raft 成为了比 Paxos 更易于理解且能为构建实际系统提供更好的基础的一种算法。为了更加易懂，Raft 将共识算法中的几个关键元素（比如 leader 选举，log 复制，安全等）互相分离，并通过使用更强的一致性要求来减少必须要考虑的状态（state）的数量。用户调研的结果显示 Raft 相比 Paxos 而言更容易学习。同时，Raft 还包含了一种新的通过覆盖大多数来确保安全的机制来修改集群成员。

<!-- more -->



## 1 介绍

共识算法能够让一组机器以一个一致组的方试运行，并且能容忍其中一些成员的失效。正因为如此，共识算法在构建可靠的大规模软件系统中扮演了关键的角色。Paxos[15, 16] 在过去十年里主导了共识算法的讨论：多数共识算法的实现都基于 Paxos 或受到其影响，因此 Paxos 也就成为了共识算法教学的主要工具。

不幸的是，尽管存在多种让 Paxos 更加平易近人的尝试，其算法本身仍旧是难以理解。此外，Paxos 的架构需要进行复杂的更改才能支持实际的系统。结果，不论是学生还是系统工程师都受够了 Paxos。

在与 Paxos 斗争了一番以后，我们开始动身去寻找一种新的共识算法，期望能为教学或构建系统提供一种更好的基础算法。由于我们的首要目标是易懂，因此这让我们的尝试变得不同寻常：我们是否能够找到一种一种为了构建实际系统而定义的共识算法，它能采用一种明显比 Paxos 更简单的方式来描述？此外，我们期望这种算法能促进对直觉的开发，而这正是对构建系统的人所最为重要的。因此相比于仅仅是能让算法正常工作，更重要的是我们想让算法的原理变得显而易见。

我们最终研究出了一种共识算法，名为 Raft。在 Raft 的设计中，我们采用了特殊的技术来提升其易懂性，其包括解耦（Raft 将 leader 选举、log 复制和安全分开）与状态空间简化（相比 Paxos，Raft 减少了不确定性的程度以及服务器之间不一致的方式）。在一个 两个大学中共 43 个学生参与的用户调研显示，在易懂性上，Raft 显著优于 Paxos：在学习了 Raft 和 Paxos 之后，有 33 个学生在回答有关 Raft 的问题时比回答有关 Paxos 的问题时表现的好。

Raft 在许多地方都与现存的共识算法有相似之处（尤其是 Oki and Liskov 的 Viewstamped Replication [29, 22] 算法），但它也引入了一些新颖的特性：

- **强 leader**：Raft 采用了一种比其他共识算法更强的 leadership 模型。比如，log entries 只能从 leader 流向其他 follower。这种约定简化了 log 复制的管理并让 Raft 更易于理解。
- **leader 选举**：Raft 在 leader 选举上使用随机定时器。这只要在任何共识算法都必须的心跳机制中增加一点小的改动就能实现，随机定时能简单且迅速的解决冲突。
- **membership 更改** ：Raft 对更改集群中的 server 采用的机制是一种新的联合共识（joint consensus）手段，其大多数（majority）的转换过程中，两种不同的配置可以相互重叠。这允许集群在配置更改时仍然能连续工作。

我们认为 Raft 相比于 Paxos 以及其他的共识算法，在作为教育目的或作为实现的基础两方面都更优秀。它比其他算法更简单，也更易懂；他在满足实际系统的需求上也足够完备；它拥有许多开源的实现并被应用在许多公司中；其安全特性已经被正式的定义并被证明；其效率与其他算法相比也不相伯仲。

本文的余下部分介绍了复制状态机问题（第二节），讨论了 Paxos 的优缺点（第三节），描述了我们对易懂性采取的一般方法（第四节），展示了 Raft 共识算法（第五~八节），评估了 Raft（第九节），最后讨论了与之有关的工作（第十节）。



## 2 复制状态机

共识算法通常出现在复制状态机（replicated state machines）的上下文中[37]。在这种方法中，服务器集合上的状态机计算相同状态的相同副本，即使某些服务器关闭，也能继续运行。复制状态机用于解决在分布式系统中存在的各种容错（fault tolerance）问题。比如，在拥有单个集群 leader 的大规模系统中，如 GFS [8]，HDFS [38]，和 RAMCloud [33]，通常会使用一个分离的复制状态机来用于管理 leader 选举，并存储关键的配置信息，该复制状态机在 leader 崩溃后仍然能够继续存活。这类复制状态机的例子有 Chubby [2] 和 ZooKeeper [11] 等。

复制状态机通常采用 log 复制来实现，见图1。每个服务器都保存了一个包含一系列命令的 log，而状态机会按顺序执行该 log。每一个 log 都包含了相同顺序的相同命令，因此每一个复制状态机都能按照相同的顺序来执行命令。由于状态机是确定的，因此每个状态机都计算相同的状态并输出相同的序列。

{% asset_img 1.png Figure 1: Replicated state machine architecture. The consensus algorithm manages a replicated log containing state machine commands from clients. The state machines process identical sequences of commands from the logs, so they produce the same outputs. %}

**图 1**：复制状态机架构。共识算法管理的一个 log 复制包含从客户端获得的状态机命令。状态机以相同的顺序执行 log 中的命令，因此他们会产生相同的输出。

共识算法的工作就是确保 log 复制的一致性。服务器上的共识模块接收从客户端发来的命令，并将之添加在 log 中。之后它与其他服务器上的共识模块通信，以确保每个服务器上的 log 最终都以相同的顺序保持相同的请求，即使某些服务器发生故障。一旦命令被成功复制，每个服务器的状态机就以 log 的顺序来执行这些命令，并将输出返回给客户端。最终，这些服务器看起来就像形成了一个单一的高可靠的状态机。

实际系统中的共识算法通常有如下属性：

- 它们在所有非拜占庭条件下确保*安全*（绝不返回错误的结果），这包括网络延迟、网络分区、丢包、包重复、包重排序等。

- 只要服务器集群中的任何大多数是可用的并且它能与其他服务器和客户端正常通信，那么整个集群就是可用的。因此，一个包含五个服务器的典型集群能容忍两个服务器宕机的错误。服务器都假定由于宕机而导致失效；他们可能从稳定存储的状态中恢复，并重新加入集群。
- 他们不依赖时间来确保 log 的一致性：错误的时钟和极端的消息延迟在最糟的情况下可能会导致可用性问题。
- 通常，只要集群中的大多数响应了一轮远程过程调用，就可认为一条命令完成了；少数响应较慢的服务器不需要影像系统的总体性能。



## 3 Paxos 有什么问题？

在过去十年里，Leslie Lamport 的 Paxos 协议[15] 已经几乎成为了共识算法的同义词：它是课堂上最常被讲授的协议，也是大多数共识算法实现的起点。Paxos 首先定义了一个能够就单个决策达成一致的协议，例如复制单个 log entries。我们将此子集称为单一判定 Paxos（single-decree Paxos）。之后 Paxos 将该协议的多个实例组合起来，来达成一系列的决策，例如 log（multi-Paxos）。Paxos 确保了安全性和活性，且能支持集群成员变更。它的正确性已被证明，同时在常见的场景下它很有效。

不幸的是，Paxos 存在两个显著的缺点。第一个缺点是 Paxos 异常难懂。其完整的解释[15] 是臭名昭著的晦涩；只有少数人能顺利的读懂它，而且是在花费大量精力的前提下。因此，存在许多尝试以更简单的术语来解释 Paxos 的尝试 [16, 20, 21]。这些解释聚焦于单一判定 Paxos 子集，并且它们也仍然具有挑战性。在一项对 NSDI 2012 与会者的非正式调查中，我们发现即使是在经验丰富的研究者之中，也只有少数人对 Paxos 感到舒适。我们自己也在与 Paxos 作斗争；我们难以理解完整的协议，直到我们读了一些简化的解释，以及开始尝试设计我们自己的替代性协议之后才搞懂它，而这一过程花了近一年时间。

我们假设 Paxos 的晦涩是源于其选择单一判定 Paxos 作为基础。单一判定 Paxos 本身就晦涩而不易懂：它被分成两个阶段，没有简单的直观解释，无法独立理解。因此，它难以对单一判定协议的工作原理上建立直觉。而 multi-Paxos 的合并规则进一步增加了复杂度和不透明。我们相信，就多个决策（log 代替单个 entry）上达成共识的总体问题是可以通过其他更直接和清晰的方法来分解的。

Paxos的第二个问题是它没有为构建实际的实现提供良好的基础。其中一个原因是对 multi-Paxos 还没有一个被广泛认可的算法。Lamport 的描述也主要关注于单一判定的 Paxos；他只是大致描绘了达成 multi-Paxos 的可能性，但许多细节都是缺失的。有很多对 Paxos 进行充实与优化的尝试，例如[26]， [39]，和 [13]，但这些尝试互相之间并不相同，也不同于 Lamport 的描绘。类似的系统如Chubby [4] 实现了类似 Paxos 的算法，但大多数情况下其细节并没有被公布。

另外，Paxos 架构在构建实际系统上很匮乏；这是单一判定分解的另一个后果。例如，单独选取一组 log entries，然后将之合并之一个连续的 log 中，这种做法并没有什么好处；反倒增加了复杂度。围绕着 log 来设计的系统，新的 entries 会序列化的的以一个限定的顺序追加，这样更简单且高效。另一个问题是 Paxos 在其核心使用了对称点到点的方法（尽管它最终建议采用一个更弱的 leadership 形式来优化性能）。这在只有下达单一决策的简化版世界中是合理的，但只有少数实际系统使用了这种方法。假如一系列的决策需要被下达，那么更简单且更快的方式是先选举一个 leader，之后由这个 leader 来协调决策。

因此，实际系统与Paxos几乎没有相似之处。每个实现都是始于 Paxos，在实施中发现了其困难点，最终开发出一个显著不同的架构来。这种方式耗时且易出错，而理解 Paxos 的困难加剧了问题。Paxos 的表述方法也许更利于理论证明其正确性，但实际的实现与 Paxos 是如此的不同，以至于其证明没什么价值。Chubby 的开发者有如下的经典评价：

> Paxos 算法的描述与真实世界的需求之间存在巨大的鸿沟... 最终的系统将基于未被证明的协议[4]。

由于这些问题，我们得出结论，Paxos 并不能为构建系统或教学而提供一个好的基础。鉴于共识算法在大规模软件系统中的重要性，我们决定尝试看看是否能自己设计一个更好的替代算法。Raft 就是这一尝试的结果。

## 4 为了易懂性而设计

We had several goals in designing Raft: it must provide a complete and practical foundation for system building, so that it significantly reduces the amount of design work required of developers; it must be safe under all conditions and available under typical operating conditions; and it must be efficient for common operations. But our most important goal—and most difficult challenge—was understandability. It must be possible for a large audience to understand the algorithm comfortably. In addition, it must be possible to develop intuitions about the algorithm, so that system builders can make the extensions that are inevitable in real-world implementations



There were numerous points in the design of Raft where we had to choose among alternative approaches. In these situations we evaluated the alternatives based on understandability: how hard is it to explain each alternative (for example, how complex is its state space, and does it have subtle implications?), and how easy will it be for a reader to completely understand the approach and its implications?



We recognize that there is a high degree of subjectivity in such analysis; nonetheless, we used two techniques that are generally applicable. The first technique is the well-known approach of problem decomposition: wherever possible, we divided problems into separate pieces that could be solved, explained, and understood relatively independently. For example, in Raft we separated leader election, log replication, safety, and membership changes.



Our second approach was to simplify the state space by reducing the number of states to consider, making the system more coherent and eliminating nondeterminism where possible. Specifically, logs are not allowed to have holes, and Raft limits the ways in which logs can become inconsistent with each other. Although in most cases we tried to eliminate nondeterminism, there are some situations where nondeterminism actually improves understandability. In particular, randomized approaches introduce nondeterminism, but they tend to reduce the state space by handling all possible choices in a similar fashion (“choose any; it doesn’t matter”). We used rando



## 5 Raft 共识算法

Raft is an algorithm for managing a replicated log of the form described in Section 2. Figure 2 summarizes the algorithm in condensed form for reference, and Figure 3 lists key properties of the algorithm; the elements of these figures are discussed piecewise over the rest of this section.



Raft implements consensus by first electing a distinguished leader, then giving the leader complete responsibility for managing the replicated log. The leader accepts log entries from clients, replicates them on other servers, and tells servers when it is safe to apply log entries to their state machines. Having a leader simplifies the management of the replicated log. For example, the leader can decide where to place new entries in the log without consulting other servers, and data flows in a simple fashion from the leader to other servers. A leader can fail or become disconnected from the other servers, in which case a new leader is elected. 



Given the leader approach, Raft decomposes the consensus problem into three relatively independent subproblems, which are discussed in the subsections that follow:



