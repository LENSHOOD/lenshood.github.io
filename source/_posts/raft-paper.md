---
title: （译文）一种易于理解的共识算法的研究（In Search of an Understandable Consensus Algorithm）
mathjax: true
date: 2021-09-13 17:57:50
tags: 
- raft
categories:
- Software Engineering
---

{% asset_img header.jpg 500 %}

> 本文是对 raft paper 的翻译，原文请见 https://raft.github.io/raft.pdf

## 摘要

Raft 是一种用于管理 log 复制的共识算法。它采用了与 Paxos 不同的算法结构，但能产出与 (multi-)Paxos 等价的结果，并且与 Paxos 一样高效；这令 Raft 成为了比 Paxos 更易于理解且能为构建实际系统提供更好的基础的一种算法。为了更加易懂，Raft 将共识算法中的几个关键元素（比如 leader 选举，log 复制，安全等）互相分离，并通过使用更强的一致性要求来减少必须要考虑的状态（state）的数量。用户调研的结果显示 Raft 相比 Paxos 而言更容易学习。同时，Raft 还包含了一种新的通过覆盖大多数来确保安全的机制来修改集群成员。

<!-- more -->



## 1 介绍

共识算法能够让一组机器以一个一致组的方试运行，并且能容忍其中一些成员的失效。正因为如此，共识算法在构建可靠的大规模软件系统中扮演了关键的角色。Paxos[15, 16] 在过去十年里主导了共识算法的讨论：多数共识算法的实现都基于 Paxos 或受到其影响，因此 Paxos 也就成为了共识算法教学的主要工具。

不幸的是，尽管存在多种让 Paxos 更加平易近人的尝试，其算法本身仍旧是难以理解。此外，Paxos 的架构需要进行复杂的更改才能支持实际的系统。结果，不论是学生还是系统工程师都受够了 Paxos。

在与 Paxos 斗争了一番以后，我们开始动身去寻找一种新的共识算法，期望能为教学或构建系统提供一种更好的基础算法。由于我们的首要目标是易懂，因此这让我们的尝试变得不同寻常：我们是否能够找到一种为了构建实际系统而定义的共识算法，它能采用一种明显比 Paxos 更简单的方式来描述？此外，我们期望这种算法能促进对直觉的开发，而这正是对构建系统的人所最为重要的。因此相比于仅仅是能让算法正常工作，更重要的是我们想让算法的原理变得显而易见。

我们最终研究出了一种共识算法，名为 Raft。在 Raft 的设计中，我们采用了特殊的技术来提升其易懂性，其包括解耦（Raft 将 leader 选举、log 复制和安全分开）与状态空间简化（相比 Paxos，Raft 减少了不确定性的程度以及服务器之间不一致的方式）。在一个两所大学中共 43 个学生参与的用户调研显示，在易懂性上，Raft 显著优于 Paxos：在学习了 Raft 和 Paxos 之后，有 33 个学生在回答有关 Raft 的问题时比回答有关 Paxos 的问题时表现的好。

Raft 在许多地方都与现存的共识算法有相似之处（尤其是 Oki and Liskov 的 Viewstamped Replication [29, 22] 算法），但它也引入了一些新颖的特性：

- **强 leader**：Raft 采用了一种比其他共识算法更强的 leadership 模型。比如，log entries 只能从 leader 流向其他 follower。这种约定简化了 log 复制的管理并让 Raft 更易于理解。
- **leader 选举**：Raft 在 leader 选举上使用随机定时器。这只要在任何共识算法都必须的心跳机制中增加一点小的改动就能实现，随机定时能简单且迅速的解决冲突。
- **membership 更改** ：Raft 对更改集群中的 server 采用的机制是一种新的联合共识（joint consensus）手段，其大多数（majority）的转换过程中，两种不同的配置可以相互重叠。这允许集群在配置更改时仍然能连续工作。

我们认为 Raft 相比于 Paxos 以及其他的共识算法，在作为教育目的或作为实现的基础两方面都更优秀。它比其他算法更简单，也更易懂；他在满足实际系统的需求上也足够完备；它拥有许多开源的实现并被应用在许多公司中；其安全特性已经被正式的定义并被证明；其效率与其他算法相比也不相伯仲。

本文的余下部分介绍了复制状态机问题（第二节），讨论了 Paxos 的优缺点（第三节），描述了我们对易懂性采取的一般方法（第四节），展示了 Raft 共识算法（第五~八节），评估了 Raft（第九节），最后讨论了与之有关的工作（第十节）。



## 2 复制状态机

共识算法通常出现在复制状态机（replicated state machines）的上下文中[37]。在这种方法中，服务器集合上的状态机计算相同状态的相同副本，即使某些服务器关闭，也能继续运行。复制状态机用于解决在分布式系统中存在的各种容错（fault tolerance）问题。比如，在拥有单个集群 leader 的大规模系统中，如 GFS [8]，HDFS [38]，和 RAMCloud [33]，通常会使用一个分离的复制状态机来用于管理 leader 选举，并存储关键的配置信息，该复制状态机在 leader 崩溃后仍然能够继续存活。这类复制状态机的例子有 Chubby [2] 和 ZooKeeper [11] 等。

复制状态机通常采用 log 复制来实现，见图1。每个服务器都保存了一个包含一系列命令的 log，而状态机会按顺序执行该 log。每一个 log 都包含了相同顺序的相同命令，因此每一个复制状态机都能按照相同的顺序来执行命令。由于状态机是确定的，因此每个状态机都计算相同的状态并输出相同的序列。

{% asset_img 1.png %}

**图 1**：复制状态机架构。共识算法管理的一个 log 复制包含从客户端获得的状态机命令。状态机以相同的顺序执行 log 中的命令，因此他们会产生相同的输出。

共识算法的工作就是确保 log 复制的一致性。服务器上的共识模块接收从客户端发来的命令，并将之添加在 log 中。之后它与其他服务器上的共识模块通信，以确保每个服务器上的 log 最终都以相同的顺序保持相同的请求，即使某些服务器发生故障。一旦命令被成功复制，每个服务器的状态机就以 log 的顺序来执行这些命令，并将输出返回给客户端。最终，这些服务器看起来就像形成了一个单一的高可靠的状态机。

实际系统中的共识算法通常有如下属性：

- 它们在所有非拜占庭条件下确保*安全*（绝不返回错误的结果），这包括网络延迟、网络分区、丢包、包重复、包重排序等。

- 只要服务器集群中的任何大多数是可用的并且它能与其他服务器和客户端正常通信，那么整个集群就是可用的。因此，一个包含五个服务器的典型集群能容忍两个服务器宕机的错误。服务器都假定由于宕机而导致失效；他们可能从稳定存储的状态中恢复，并重新加入集群。
- 他们不依赖时间来确保 log 的一致性：错误的时钟和极端的消息延迟在最糟的情况下可能会导致可用性问题。
- 通常，只要集群中的大多数响应了一轮远程过程调用，就可认为一条命令完成了；少数响应较慢的服务器不会影响系统的总体性能。



## 3 Paxos 有什么问题？

在过去十年里，Leslie Lamport 的 Paxos 协议[15] 已经几乎成为了共识算法的同义词：它是课堂上最常被讲授的协议，也是大多数共识算法实现的起点。Paxos 首先定义了一个能够就单个决策达成一致的协议，例如复制单个 log entries。我们将此子集称为单一判定 Paxos（single-decree Paxos）。之后 Paxos 将该协议的多个实例组合起来，来达成一系列的决策，例如 log（multi-Paxos）。Paxos 确保了安全性和活性，且能支持集群成员变更。它的正确性已被证明，同时在常见的场景下它很有效。

不幸的是，Paxos 存在两个显著的缺点。第一个缺点是 Paxos 异常难懂。其完整的解释[15] 是臭名昭著的晦涩；只有少数人能顺利的读懂它，而且是在花费大量精力的前提下。因此，存在许多以更简单的术语来解释 Paxos 的尝试 [16, 20, 21]。这些解释聚焦于单一判定 Paxos 子集，并且它们也仍然具有挑战性。在一项对 NSDI 2012 与会者的非正式调查中，我们发现即使是在经验丰富的研究者之中，也只有少数人对 Paxos 感到舒适。我们自己也在与 Paxos 作斗争；我们难以理解完整的协议，直到读了一些简化的解释，以及开始尝试设计我们自己的替代性协议之后才搞懂它，而这一过程花了近一年时间。

我们假设 Paxos 的晦涩是源于其选择单一判定 Paxos 作为基础。单一判定 Paxos 本身就晦涩而不易懂：它被分成两个阶段，没有简单的直观解释，无法独立理解。因此，难以在单一判定协议的工作原理上建立直觉。而 multi-Paxos 的合并规则进一步增加了复杂度和不透明。我们相信，就多个决策（log 代替单个 entry）上达成共识的总体问题是可以通过其他更直接和清晰的方法来分解的。

Paxos的第二个问题是它没有为构建实际的实现提供良好的基础。其中一个原因是对 multi-Paxos 还没有一个被广泛认可的算法。Lamport 的描述也主要关注于单一判定的 Paxos；他只是大致描绘了达成 multi-Paxos 的可能性，但许多细节都是缺失的。有很多对 Paxos 进行充实与优化的尝试，例如[26]， [39]，和 [13]，但这些尝试互相之间并不相同，也不同于 Lamport 的描绘。类似的系统如Chubby [4] 实现了类似 Paxos 的算法，但大多数情况下其细节并没有被公布。

另外，Paxos 架构在构建实际系统上很匮乏；这是单一判定分解的另一个后果。例如，单独选取一组 log entries，然后将之合并到一个连续的 log 中，这种做法并没有什么好处；反倒增加了复杂度。围绕着 log 来设计的系统，新的 entries 会序列化的以一个限定的顺序追加，这样更简单且高效。另一个问题是 Paxos 在其核心使用了对称点到点的方法（尽管它最终建议采用一个更弱的 leadership 形式来优化性能）。这在只有下达单一决策的简化版世界中是合理的，但只有少数实际系统使用了这种方法。假如一系列的决策需要被下达，那么更简单且更快的方式是先选举一个 leader，之后由这个 leader 来协调决策。

因此，实际系统与Paxos几乎没有相似之处。每个实现都是始于 Paxos，在实施中发现了其困难点，最终开发出一个显著不同的架构来。这种方式耗时且易出错，而理解 Paxos 的困难加剧了问题。Paxos 的表述方法也许更利于理论证明其正确性，但实际的实现与 Paxos 是如此的不同，以至于其证明没什么价值。Chubby 的开发者有如下的经典评价：

> Paxos 算法的描述与真实世界的需求之间存在巨大的鸿沟... 最终的系统将基于未被证明的协议[4]。

由于这些问题，我们得出结论，Paxos 并不能为构建系统或教学而提供一个好的基础。鉴于共识算法在大规模软件系统中的重要性，我们决定尝试看看是否能自己设计一个更好的替代算法。Raft 就是这一尝试的结果。



## 4 为了易懂性而设计

我们在设计 Raft 中有几点目标：它必须能为构建系统提供一个完整的和实际的基础，这样就能极大的减少对开发者的设计工作要求；他必须能在所有条件下安全，且能在典型的操作条件下可用；它必须在通用操作下高效。不过，我们最重要的目的（也是最困难的挑战）是易懂性。它必须能让大多数人能舒适的理解其算法本质。此外，它还必须能够培养对算法的直觉，使系统构建者可以实施在真实世界当中不可避免的扩展。

在设计 Raft 的过程中，我们存在大量的要点需要在各种替代方案中进行选择。在这种场景下，我们评估替代方案就是基于易懂性：对每一种替代方案，它到底有多难解释（例如，其状态空间有多复杂，是否包含有不清晰的含义？），以及对于读者，完全的理解这种方法及其含义有多容易？

我们意识到在这种分析当中存在很高的主观性；尽管如此，我们使用了两种普适的技术。第一种技术是众所周知的问题拆分法：只要可能，我们就将问题拆成独立的可解决、可解释、可相对独立进行理解的片段。例如，在 Raft 中我们将 leader 选举、log 复制、安全性以及成员变更进行了拆分。

我们的第二个方法是通过减少需要考虑的状态数量来简化状态空间，让系统更加连贯且尽可能的消除不确定性。特别的，log 中不允许存在空洞，Raft 也限制了让 log 之间可能变得不连续的途径。尽管我们尽可能的在多数场景下消除了不确定性，但在某些场景下，不确定性实际上提升了易懂性。在实际当中，随机性方法引入了不确定性，但这却趋向于通过以同一种方法处理所有可能的选择，从而减少了状态空间（“选择任何一种，这无关紧要”）。我们采用随机化来简化 Raft 的 leader 选举算法。



## 5 Raft 共识算法

Raft 是一种以第二节中描述的形式来管理 log 复制的算法。图 2 以简明的形式总结了该算法以供参考，图 3 列举了该算法的关键属性；这些图中的元素将在本节的其余部分中分段讨论。

{% asset_img 2.png %}

**图 2**：Raft 共识算法的一个简明总结（除了成员变更与 log 压缩）。左上框中的服务器行为描述为一组独立且重复触发的规则。章节编号，如§5.2，表明讨论特定特性的位置。一个正式的规范[31] 更精确地描述了该算法。

{% asset_img 3.png %}

**图 3**：Raft 保证这些属性中的每一个在所有时间里都成立。小节编号表明每一个属性讨论的具体位置。

Raft 通过首先选举出一个重要的 leader，并让该 leader 全权负责管理 log 复制，来实现共识。leader 从客户端处接受 log entries，将其复制到其他服务器上，并且通知它们何时将 log entries 应用到状态机中是安全的。拥有一个 leader 简化了对 log 复制的管理。比如，leader 可以在不咨询其他服务器的情况下决定在日志中放置新条目的位置，数据以简单的方式从leader流向其他服务器。leader 可能会失效或与其他服务器断开连接，这时，新的 leader 将被选举出来。

鉴于这种 leader 的方法，Raft 将共识问题分解成了三个相对独立的小问题，将在一下小节中进行讨论：

- **leader 选举**：当现存的 leader 失效时，一个新的 leader 必须要被选出。（5.2 小节）
- **log 复制**：leader 必须从客户端接受 log entries 并且将其在整个集群之间复制，强制其他的 log 与自己的保持一致（5.3 小节）
- **安全**：Raft 的关键安全属性是图 3 中的 “State Machine Safety Property”：假如任何服务器已经将某个特定的 log entry 应用于其状态机上，那么任何其他的服务器都不可能再在同一个 log index 上应用一个不同的命令了。5.4 小节描述了 Raft 是怎么确保这一属性的；解决方案涉及了一个在 5.2 节描述的选举机制上的额外限制。

在展示了共识算法之后，这一节将讨论可用性问题以及时间在系统中的作用。

### 5.1 Raft 基础

一个 Raft 集群包含了许多个服务器；五是一个典型数量，它允许整个系统可以容忍两个服务器失效。在任意时间，每一个服务器都会处于如下三种状态中的一种：leader，follower，candidate。在正常运行时只会存在一个 leader，剩余服务器都是 follower。follower 是被动的：它们自己不主动发起请求，而是简单地对从 leader 和 follower 到来的请求做出响应。leader 处理所有的客户端请求（假如客户端连接了一个 follower，那么该 follower 就会将此请求重定向到 leader）。第三种状态 candidate，是用于选举新的 leader 的，这将在 5.2 节描述。图 4 展示了状态及其流转；如何流转将在下文讨论。

{% asset_img 4.png %}

**图 4**：服务器状态。follower 仅仅对从其他服务器到来的请求做出响应。假如一个 follower 无法正常通信，它便转换为一个 candidate 并发起一次选举。当一个 candidate 收到了集群中大多数的选票时，它就变成了 leader。leader 通常会一直运行直到它失效为止。

Raft 将时间分割成任意长度的*term（任期）*，如图 5 所示。term 使用连续的整数编号。每个 term 都从一场选举开始，其中一名或多名 candidate 将会尝试变成 5.2 节将描述的 leader。假如某个 candidate 在这次选举中胜出，那么在这个 term 之后的时间里它都以 leader 的身份服务。在某些场景下，选举可能最终会得到分裂的投票结果。那么这一个 term 将以无 leader 而结束；一个新的 term 马上会开始（包含一次新的选举）。Raft 确保在一个给定的 term 中至多存在一个 leader。

{% asset_img 5.png %}

**图 5**：时间被分割成为 term，每一个 term 都以一次选举为开始。在选举成功之后，一个单独的 leader 一直管理整个集群直到 term 结束。假如某些情况下选举失败了，那么这个 term 就将以没有 leader 而结束。term 之间的转换，在不同的服务器上可能会被观察到发生在不同的时间点。

不同的服务器可能会在不同的时间点上观察到 term 的转换，且在某些情况下一个服务器可能无法观察到一次选举或甚至是整个 term。term 扮演了逻辑时钟[14] 的角色，它们允许服务器观察到已过时的信息，例如一个已经过时的 leader。每个服务器都保存着 *current term（当前的 term）* 编号，该编号是单调递增的。current term 会在一切的通信之间被传递；假如某一个服务器的 current term 小于其他的服务器，那么他将会将自己的 current term 更新为更大的 term 值。如果某个 candidate 或者 leader 发现它自己的 term 过期了，那么会立即回退为 follower 状态。假如一个服务器收到了一个包含了过时的 term 号的请求，它便会拒绝该请求。

Raft 服务器通过远程过程调用（RPC）来互相通信，基本的共识算法只需要两类 RPC。RequestVote RPC 是由 candidate 在选举时发起（见 5.2 节），而 AppendEntries RPC 则是由 leader 发起用于 log 复制以及心跳（见 5.3 节）。第七节增加了第三种 RPC 用于在服务器之间传递快照（snapshot）。假如服务器没有及时地收到响应，那么它就会尝试重发，服务器都采用并行的方式发送 RPC 以达到最佳性能。

### 5.2 leader 选举

Raft 使用心跳机制来触发选举。当服务器启动时，它们都是 follower。只要服务器收到了其他 leader 或 candidate 的有效 RPC，那么它就保持为 follower。leader 会定期的发送心跳（不包含任何 log entries 的空 AppendEntries RPC）给所有 follower 来维持自己的 leader 权威。假如一个 follower 在一段时间内都没有收到任何的通信信息，那么就会发生选举超时（election timeout），之后它就会假定目前没有一个有效的 leader，并开启一次选举来选出一个新的 leader。

为了开启一次选举，follower 会将自己的 current term 增加，之后转换为 candidate 状态。然后，它先投票给自己，之后并行的给其他集群内的服务器发起 RequestVote RPC。一个 candidate 会保持自己的状态直到如下三件事发生：（a）它赢得了选举，（b）另一个服务器将自己确立为 leader，（c）一段时间过后，仍然没有赢家。下文各段将分别讨论这些结果。

当一个 candidate 收到了同一集群在同一 term 内的大多数的选票，即认为选举获胜。每个服务器在一个 term 期间都会基于先到先服务的原则（注意，在 5.4 节对投票增加了一种额外的限制），至多给一个 candidate 投票。大多数原则确保了在一个特定的 term 内（见图 3 的 Election Safety Property）至多只能有一个 candidate 赢得选举。一旦某个 candidate 赢得了选举，它便是 leader。之后它会给其他所有的服务器发送心跳消息来建立自己的权威，并且阻止新的选举的发生。

在等待投票时，candidate 可能会收到从其他已经声称自己为 leader 的服务器发来的 AppendEntries RPC 消息。如果这个 leader 的 term（包含在 RPC 消息内）至少与 candidate 当前的 current term 一样大，那么这个 candidate 就会意识到这个 leader 是合法的，因此将自己回退到 follower 状态。但如果这个 RPC 中包含的 term 小于该 candidate 的 current term，那么它就会拒绝该 RPC 并且继续保持自己的 candidate 状态。

第三种可能发生的情形是某个 candidate 既没有赢得也没有输掉当前的选举：假如有许多 follower 都同时变成了 candidate，那么选票就可能分裂导致没有一个 candidate 能获得大多数选票。当这种情况发生时，每个candidate 都将超时，然后通过增加自己的 term 并且给其他服务器发送一轮 RequestVote RPC 来开启一次新的选举。然而，如果不采取额外的措施，这种投票分裂的情况可能会无限的持续下去。

Raft 采用了选举超时随机的方式来确保投票分裂罕有发生并且在发生时能快速的解决掉。为了在一开始就防止投票分裂（译者注：在大家都是 follower 的时候），选举的超时时间在一个固定的范围内随机选取（比如 150~300ms）。这将会把服务器分散开，以便在大多数情况下只有一台服务器会超时；它将会赢得选举，并且在其他服务器超时之前发送心跳。相同的机制被用于处理投票分裂。每个 candidate 都会在开启一轮选举之前随机的重置自己的选举超时计时，它会等待该超时时间流逝完毕，然后再启动下一轮选举；这降低了在新的选举中再次出现投票分裂的可能性。9.3 节展示了这种方法能快速的选出 leader。

选举正是一种展现我们以易懂性来指导在各种设计替代之间做出选择的例子。最开始我们计划使用一种评级系统：每个 candidate 都被指定为一个特定的等级，用于在 candidate 之间评比选择。假如某个 candidate 发现有其他 candidate 的评级比它高，那么它就会回退为 follower 状态因此更高评级的 candidate 会更容易在下一次选举中胜出。我们发现这种办法会围绕可用性产生一些细微的问题（一个低评级的服务器也许需要在某个高评级服务器失效时超时并尝试再次转变为 candidate，但假如这个时间过短，它就可能会重置某个选举过程）。我们对这个算法做了多次调整，但每一次调整都会出现新的极端情况。最终我们总结得出，随机重试的办法才是最明确也最易懂的。

### 5.3 log 复制

一旦一个 leader 被选出，他就开始服务客户端请求。每个客户端请求都包含一个能被复制状态机执行的命令（command）。leader 将该命令追加到它自己的 log 最后面，作为一个新的 log entry，之后并行的给所有其他服务器发送 AppendEntries RPC 来复制该 entry。当这个 log entry 被安全的复制之后（这将在下面详述），leader 将这个 log entry 应用于他自己的状态机中，并将执行结果返回给客户端。假如 follower 崩溃或者运行迟缓，以及发生网络丢包时，leader 会无限的重发 AppendEntries RPC（即使它已经给客户端返回了结果）直到所有的 follower 最终都存储了所有 log entries。

log 以图 6 的格式组织。当每一条 log entry 被 leader 接受到时，其内部 entry 都保存了一个状态机命令和它当前所处的 term 号。term 号用于探测不同 log 之间的不一致性来确保图 3 中的某些属性。每个 log entry 还包含一个整数的索引（index）来定位它在 log 当中的位置。

{% asset_img 6.png %}

**图 6**：log 以 entry 组成，并以顺序编号。每个 entry 都包含创建它时的 term 号（每个框中的数字）以及状态机的命令。当一个 entry 被安全的应用在状态机之后，就认为它已经 *committed（被提交）*。

leader 能够决定何时将 log entry 应用到状态机上是安全的；这样的 entry 成为 committed 的。Raft 确保 committed entries 是持久的并且最终能够被所有可用的状态机执行。一旦当 leader 将一个由当前 leader 创建的 log entry 复制到了大多数服务器，那么该 entry 就是 committed 的（例如图 6 中的 entry 7）。这样会将 leader log 中之前的所有 entries 提交，包括由之前的 leader 所创建的 entries。5.4 节讨论了在 leader 变更之后，应用这一规则过程中的一些细节，同时也表明对于提交的定义是安全的。leader 会持续跟踪它所知道的将要被 committed 的 entry 的最高 index 值，然后它将这个 index 包含在未来的 AppendEntries RPC 中（包括心跳）以便其他服务器最终发现。一旦某个 follower 发觉到某个 log entry 已经 committed，它就会将之应用到它自己本地的状态机中（以 log 的顺序）。

我们设计了 Raft 日志机制，以保持不同服务器上日志之间的高度一致性。这不仅简化了系统行为使其更易于预测，这还成为了确保安全的一个重要组件。Raft 维护着如下的属性，它们共同组成了图 3 中的 “Log Matching Property”（日志匹配属性）：

- 假如不同 log 中的两个 entry 拥有相同的 term 和 index，那么它们一定保存了相同的命令。

- 假如不同 log 中的两个 entry 拥有相同的 term 和 index，那么所有该 entry 前面的 log 也都是相同的。

第一个属性源于这样一种事实，leader 在给定的 term 和给定的 log index 上，最多只能创建一条 entry，且 log entries 绝不会改变它们在 log 当中的位置。而第二个属性由 AppendEntries 执行的简单的一致性检查来保证。当发送一条 AppendEntries RPC 时，leader 会将紧跟着当前新 entry 的前导 entry 的 term 以及 index 信息包含在请求体中。假如 follower 在它自己的 log 中相同 term，相同 index 的位置上，找不到这一前导 entry，那么它将拒绝这个新增的 entry。这种一致性检查就如同一个归纳步骤：日志的初始空状态满足 Log Matching Property，并且一致性检查在日志扩展时能继续延续 Log Matching Property。最终，无论 AppendEntries 是否返回成功，leader 总是能通过新的 entry 来确认其 follower 的 log 与自己的相同。

在正常运行中，leader 和 follower 的 log 保持一致，因此 AppendEntries 的一致性检查绝不会失败。然而，leader 崩溃则会导致 log 发生不一致（旧 leader 可能还没有完全将其 log 复制到其他 follower）。这种不一致会在一系列的 leader 和 follower 崩溃中持续加剧。图 7 展示了 follower 的 log 与新 leader 的 log 不同的几种情况。一个 follower 有可能会缺失当前 leader 中存在的 entries，它也可能拥有当前 leader 的 log 中不存在的额外 entries，或二者皆存在。日志中丢失的和无关的条目可能跨越多个 term。

{% asset_img 7.png %}

**图 7**：当最顶部的 leader 生效时，在如下 follower 场景（a-f）中的任意一种都有可能发生。每个框代表一条 log entry；框中的数字代表 term。一个 follower 有可能丢失 entry（a-b），有可能拥有额外的未提交 entries（c-d），也可能都有（e-f）。举例说明，在场景 f 中，该 server 可能在 term2 时是 leader，添加了许多 entries 到自己的 log 中，接着就在提交这些 entries 之前崩溃了；它很快重启，变成了 term3 的 leader，并且添加了少量新的 entries 到自己的 log 中；在任何不论是 term2 还是 term3 中的 entries 被提交之前，它又崩溃了，并且在后续的几个 term 中持续崩溃。

在 Raft 中，leader 处理不一致的方式是强迫 follower 复制它自己的 leader log。这意味着在 follower 中存在冲突的 entries 会被来自 leader 的 log 所覆盖。第 5.4 节将说明，当再加上一个限制时，就安全了。

为了使 follower 的 log 与其自己的 log 保持一致，leader 必须找到最近的两个 log 一致的 entry 位置点，删除该点之后 follower 日志中的所有 entries，并在该点之后将 leader 的所有 entries 发送给 follower。所有这些都发生在响应 AppendEntries RPC 的一致性检查时。leader 会维护每一个 follower 的 *nextIndex（下一个索引）*，即下一个 leader 将会发给 follower 的 log entry 的 index。当 leader 刚开始生效时，它将所有 nextIndex 初始化为它自己的最后一个 log entry 的下一个值（图 7 中的值是 11）。假如某个 follower 的 log 与 leader 不一致，那么 AppendEntries 的一致性检查在下一次 AppendEntries RPC 时就会失败。经过一次 rejection（请求拒绝），leader 会将该 follower 的 nextIndex 减少，然后重发 AppendEntries RPC。最终 nextIndex 将会到达一个 leader 和 follower 的 log 匹配的点。此时，AppendEntries 就会成功，并且会删除该 follower log 中任何发生冲突的 entries 然后将 leader log 中的新 entries （如果有的话）追加在自己的 log 后面。一旦 AppendEntries 成功后，follower 的 log 就与 leader 一致了，并且会在接下来的 term 中持续的一致下去。

> 如果需要，该协议可以被优化来减少被 reject 的 AppendEntries RPC 的数量。比如，当 rejecting 一个 AppendEntries 请求时，follower 可以将其发生冲突的 entry 的 term 以及它的 log 中该 term 的第一条 index，包含在响应中。当获得了这一信息后，leader 就能在减少 nextIndex 是直接跳过所有当前 term 中发生冲突的 entries，而不是通过一个个的 RPC 来减少。实际当中，我们怀疑这一优化的必要性，因为错误的发生是低频的且一般也不太会存在大量不一致的 entries。

以这种机制，当 leader 生效时，它就不需要采用任何特殊的手段来恢复 log 一致性。它只要开始正常运行，log 就会在响应 AppendEntries 一致性检查失败后自动收敛。leader 绝不覆盖或删除它自己的 log （图 3 中的 “Leader Append-Only Property” leader 仅追加属性）。

这种 log 复制机制具有第2节中描述的理想的一致性属性：只要大多数服务器是有效的，Raft 就能够接收、复制并应用新的 log entries；在通常情况下一个新的 entry 可以通过一轮 RPC 复制到集群的大多数中；并且单个缓慢的 follower 不影响整体性能。

### 5.4 安全性

前面的章节描述了 Raft 如何进行 leader 选举与 log 复制。然而，目前所描述的机制并不足以确保每个状态机都能恰好以相同的顺序执行相同的命令。比如，当 leader 提交一些 log entries 的时候，某 follower 也许不可用，之后该 follower 可能会被选举为 leader，并且用新的 entries 来覆盖原先的 entries；结果，不同的状态机就有可能执行不同的命令序列。

本节通过添加对哪些服务器可以当选为 leader 进行限制来完善 Raft 算法。这一限制确保了对任意给定的 term，其 leader log 中需要包含所有在之前的 term 中已提交的 entries（即图 3 中的 “Leader Completeness Property” leader 完整性属性）。基于选举限定，我们将提交规则变得更加精确。最后，我们给出了一个对 Leader Completeness Property 的简要证明并且展示了它是如何引导复制状态机的正确行为的。

#### 5.4.1 选举的限定

在任何基于 leader 的共识算法中，leader 最终都必须保存所有已提交的 log entries。在一些共识算法中，例如 Viewstamped Replication [22]，即使最开始不包含所有的已提交 entries，某个服务器也可以被选举为 leader。这些算法中包含了一些额外的机制来定位缺失的 entries 然后将其传递到新的 leader 中，这一过程可能发生在选举期间，也可能在选举结束后的段时间内完成。不幸的是，这需要引入相当复杂的额外机制。Raft 使用了一种更简单的方法来确保所有先前 term 中已提交的 entries 都能够在选举时就存在于每一个新的 leader 当中，而不需要将 entries 传送给 leader。这意味着 log entries 将只会向单一的方向流动，即从 leader 到 follower，而 leader 绝不会覆盖其 log 中已存在的 entries。

Raft 通过一个投票过程来防止 log 中没有包含所有已提交 entries 的 candidate 赢得选举。为了获得选票，candidate 必须要与集群中的大多数取得联系，这就意味着每一个已提交的 entry 都必须至少在一个服务器当中存在。假如该 candidate 的 log 至少与大多数中的任何其他 log 一样是最新的（之后会精确的定义这种 “最新”），那么它就持有了所有已提交的 entries。RequestVote RPC 中实现了这一限制：该请求中包含了当前 candidate 的 log，假如投票人的 log 比收到 RequestVote RPC 中的 log 更新，那么将拒绝给它投票。

Raft 通过比较 log 中最后的 entries 的 term 和 index 来判断两组 log 谁是最新。如果这两组 log 最后 entries 的 term 不同，那么谁的 term 更大则更新。而如果是相同的 term，就需要比较哪组 log 更长，则更新。

#### 5.4.2 提交先前 term 中的 entries

在 5.3 节中已经讲过，当一个 entry 被存储在大多数服务器上之后，leader 就能知道当前 term 中的这个 entry 是已提交的了。假如一个 leader 在提交一条 entry 之前就崩溃了，那么新的 leader 就会尝试去完成这条 entry 的复制。然而，leader 并不能立即得出这样的结论，即前一个 term 中的 entry 一旦存储在大多数服务器上，就会被提交。图 8 展示了一种场景，一个旧的 log entry 已经在多数服务器上存储，但仍旧被新的 leader 覆盖掉了。

{% asset_img 8.png %}

**图 8**：一个时间序列展示了为什么 leader 不能通过旧的 term 中的 entry 来决定是否提交。在（a）中 S1 是 leader，并且部分复制了index 2 的 entry。 在（b）中 S1 崩溃了；S5 通过 S3 S4 和它自己的选票，被选为 term3 的 leader，然后接收了一个在 index 2 处不同的 entry。到了（c），S5 崩溃了；S1 又重启，且被选定为 leader，继续复制。这时 term2 的 log entry 已经被复制到大多数服务器上，但还没有提交。假如这时 S1 又在（d）崩溃了（译者注：leader 先追加了新的 Cmd - Term4，之后收到了 S3 已完成 Term2 复制的消息，但还没有来得及执行 committed 的代码就崩溃了，导致 S1 的 committed 仍然是 Term1），S5 就有可能被选为 leader （通过 S2，S3，S4 的选票）并且会用他自己的 term3 的 entry 覆盖到所有其他服务器上。然而，假如 S1 在崩溃之前，已经成功的将它 current term 中的 entry 复制到大多数，就如同（e）所展示的，那么这个 entry 就会被提交（这时 S5 就不能赢得选举了）。这时，所有先前的 entries 也就都被提交了。

为了消除图 8 中存在的问题，Raft 不会通过计算副本数量来提交先前 term 中的 log entries。只有 leader 当前 term 的 log entries，会以副本计数的方式来判定提交；一旦当前 term 中的某个 entry 以这种方式成功提交，则所有先前的 entries 也就由于  Log Matching Property 而被间接提交。在某些情况下，leader 确实可以安全的认为旧的 log entries 都已经提交了（例如，一个 entry 在所有的服务器上都存在），但 Raft 为了简单，而采用了更保守的方法。

Raft 在提交规则中引入了额外的复杂度，其原因是当 leader 复制先前 term 中的 entries 时，这些 log entries 中包含的是原先的 term 号。在其他的共识算法中，假如一个新的 leader 从先前的 term 中重新复制了 entries，那么这些 entries 就必须使用新的 term 号。Raft 的方法中，由于 log entries 的 term 号不会随着时间的变化以及 log 之间的交互而改变，因此更容易进行推理。此外，Raft 中的新 leader 相比其他算法会发送更少的先前 term 中的 log entries（其他算法必须要发送冗余的 log entries 并在它们可被提交之前重新编号）。

#### 5.4.3 安全论证

基于完整的 Raft 算法，我们现在可以更加精确的论证 Leader Completeness Property 是成立的（这种论证是基于将在 9.2 节讨论的安全性证明）。如果我们假设 Leader Completeness Property 不成立，然后我们证明其矛盾性即可。假定 term T 的 leader （ leaderT）从它自己的 term 中提交了一个 log entry，但这个 log entry 在未来的 term 中并没有被存储。考虑大于 T 的最小的 term U 的 leader （leaderU ）没有存储这一 entry。

1. 已提交的 entry 在被选中时必须不在 leaderU 的 log 中（因为 leaderU 从不删除或覆盖 entry）。
2. leaderT 在大多数集群中复制了该 entry，leaderU 则获得了大多数集群的投票。因此，至少有一个服务器（“投票者”）都接受了leaderT 的条目并投票给 leaderU，如图 9 所示。选民是达成矛盾的关键。
3. 投票人必须在投票给 leaderU 之前接受 leaderT 的 entry committed；否则，它将拒绝 leaderT 的 AppendEntries 请求（由于其 current term 高于 T）。
4. 投票人在投票给 leaderU 时仍然存储该 entry，因为每个介入的 leader 都包含 entry（根据假设），leader 从不删除条目，而 follower 只有在与 leader 发生冲突时才会删除 entries。
5. 选民将投票给了 leaderU，因此 leaderU 的 log 必须与选民的 log 一样是最新的。这导致了两个矛盾中的一个。
6. 首先，如果投票者和 leader 共享相同的最后一个 log entry，那么 leader 的 log 必须至少与投票者的 log 一样长，因此其 log 包含投票者 log 中的每个 entries。这是就是一个矛盾，因为投票者包含了提交的 entries，而 leader 则被认为没有。
7. 否则，leaderU 的最后一个 term 肯定比投票者的任期长。此外，它也比 T 大，因为投票者的最后一个 log entry 至少是 T（它包含来自 T 的已提交 entry）。创建 leaderU 最后一个 log entry 的较早的 leader 必须在其 log 中包含已提交的 entry（根据假设）。然后，根据 Log Matching Property，leaderU 的 log 还必须包含已提交的 entry，这是一个矛盾。
8. 这就完成了矛盾。因此，所有 term 都大于 T 的 leader 必须包含在 term T 中已提交的来自 term T 的所有 entry。
9. Log Matching Property 保证未来的 leader 也将包含间接提交的 entries，如图8（d）中的 index 2。

{% asset_img 9.png %}

**图 9**：如果 S1 （term T 的 leader）从自己的 term 中提交了一个新的 log entry，然后 S5 在之后的 term U 中胜选成为 leader，那么就一定存在至少一个服务器（S3），即接受了该 log entry，又投票给了 S5。

>  通过 Leader Completeness Property，我们能证明图 3 中的 State Machine Safety Property，即假如一个服务器已经在状态机应用了一个给定 index 的 log entry，那么其他服务器就不能再在其状态机上应用一个相同 index 的不同的 log entry 了。在服务器应用 log entry 到其状态机时，其 log 必须与 leader 的 log 在直到这一条 entry 时是相同的，且该 entry 必须已被提交。现在考虑任何服务器应用给定log index 的最小 term；Log Completeness Property 保证所有更高 term 的 leader 都将存储相同的 log entry，因此在之后 term 中应用该 index 的服务器将会应用相同的值。故 State Machine Safety Property 也成立。
>
> 
>
> 最后，Raft 要求服务器按照 log index 的顺序来应用 entries。与  State Machine Safety Property 结合起来，这就意味着所有的服务器都会在其状态机中已完全相同的顺序，应用完全相同的一组 log entries。

#### 5.5 Follower 和 candidate 崩溃

到现在为止，我们已经讨论过了 leader 失效的问题。而 follower 和 candidate 崩溃的处理则相对而言简单得多，而且它们也都采用相同的方式来处理。假如一个 follower 或 candidate 崩溃了，那么之后发往它的 RequestVote 和 AppendEntries RPC 都将失败。Raft 对这种问题的处理办法就是无限重试；假如崩溃的服务器重启了，那么 RPC 就能正常响应。如果一个服务器在完成 RPC 但在返回响应之前崩溃，那么它将会在重启之后再次收到相同的 RPC。Raft 的 RPC 是幂等的，因此这不会产生什么危害。比如，假如一个 follower 收到了一个 AppendEntries 请求，该请求包含的 log entries 已经存在于其自己的 log 中了，那么它就会忽略这些 entries。

#### 5.6 时间与可用性

我们对 Raft 的一个要求就是，安全性一定不能依赖于时间：系统一定不能仅仅因为一些事件发生的比预期过快或过慢就产生错误的结果。然而，可用性（系统及时响应客户的能力）则不可避免的依赖于时间。例如，如果消息交换花费的时间比两次服务器崩溃的间隔时间还要长，那么 candidate 就没法保持足够长的时间来赢得选举；没有一个稳定的 leader，Raft 就无法工作。

leader 选举，是 Raft 中时间最为关键的一面。只要系统满足如下对时间的要求，那么 Raft 就能够实施选举，并维护一个稳定的 leader：

$broadcastTime ≪ electionTimeout ≪ MTBF$

在这个不等式中，broadcastTime 是一个服务器并行的给集群中所有的服务器发送 RPC，并收到它们的响应所花费的平均时间；electionTimeout 是 5.2 节所描述的选举超时时间；MTBF 则是单个服务器两次失效的平均时间间隔。broadcastTime 应该比 electionTimeout 小一个数量级，以便 leader 能够可靠地发送阻止 follower 开始选举所需的心跳信息；由于 electionTimeout 采用了随机数的方法，这个不等式也能使得投票分裂变得不太可能发生。electionTimeout 应该比 MTBF 小几个数量级，以便系统稳定地进行。当 leader 崩溃时，系统将在差不多 electionTimeout 的时间内不可用；我们希望这只代表总时间的一小部分。

broadcastTime 和 MTBF 属于底层系统的属性，而 electionTimeout 则必须由我们来选定。Raft 的 RPC 通常要求接收者将信息持久保存到稳定的存储中，因此 broadcast 可能在 0.5ms 到 20ms 之间，具体取决于存储技术。所以，electionTimeout 可能会从 10ms 到 500ms 之间选取。典型的服务器 MTBF 是几个月或更久，因此很容易满足时间要求。



## 6 集群成员变更

到目前为止，我们仍假设集群的配置（参与共识算法的服务器集合）是固定的。在实际中，偶尔需要修改该配置，比如替换掉失效的服务器，或是修改复制等级。即使这能够通过将整个集群下线，更新配置，再重启集群的方式完成，但这会导致在更改配置的过程中集群不可用。此外，任何手工的步骤都存在操作员出错的风险。为了避免这些问题，我们决定将配置更改自动化，并且将其合并到 Raft 共识算法内。

为了确保配置变更机制的安全性，在过渡期间不得出现两个 leader 在同一 term 内当选的情况。不幸的是，任何尝试将服务器直接从旧配置切换为新配置的尝试都是不安全的。我们不可能自动的一次性切换所有的服务器，因为这样集群可能会在转换时被分裂成两个独立的大多数（如图 10）。

{% asset_img 10.png %}

**图 10**：直接将一种配置切换为另一种配置是不安全的，因为不同的服务器会在不同的时间点发生切换。在这个例子中，集群从三主机切换为五主机。不幸的是，在某个时间点，有两个不同的 leader 能够在同一个 term 内被选举，一个是在旧配置（$C_{old}$）下的大多数，另一个则是新配置（$C_{new}$）下的大多数。

为了保证安全，配置变更必须使用两阶段的方式。有很多具体的实现方法可以实施两阶段。比如，一些系统（例如, [22]）在第一个阶段禁用旧配置，因此这时集群无法响应客户端请求；然后在第二个阶段使新配置生效。在 Raft 中，集群首先切换到一种过度配置，我们称之为联合共识（joint consensus）；一旦联合共识被成功提交，之后系统就可以转换为新的配置。联合共识阶段合并了旧配置与新配置：

- log entries 会复制到两种配置中所有的服务器上。

- 两种配置中的任一服务器可能会成为 leader。

- 为了达成一致（用于选举和 entry 提交），要求比如分别获得新旧两个的配置中的大多数。

联合共识允许单个服务器在不同的时间与不同的配置之间转换，而不会影响安全性。此外，联合共识允许整个集群在配置变更的过程中依然能服务客户端请求。

集群的配置通过 log 复制特殊的 entries 来进行存储与传递；图 11 展示了配置变更的过程。当 leader 收到一个将配置从 $C_{old}$ 变更到 $C_{new}$ 的请求时，它将其作为用于联合共识（图中的 $C_{old, new}$ ）的配置存储为 log entry 然后将这个 entry 采用之前描述过的机制进行复制。一旦某个服务器将这个新的配置 entry 添加到其 log 中后，它就将这个配置应用到所有未来的决策当中（服务器总是会使用其 log 当中最近的配置，无论该配置 entry 是否被提交）。这意味着 leader 将会使用 $C_{old, new}$ 中的规则来决定 $C_{old, new}$ 中的 log entry 在何时会被提交。假如该 leader 崩溃了，一个新的 leader 将会基于  $C_{old}$ 或  $C_{old, new}$  的配置来被选举，这主要依赖于获胜的服务器是否已经接收到了  $C_{old, new}$ 的配置。在任何情况下， $C_{new}$ 都不能在此期间做出单方面的决定。

{% asset_img 11.png %}

**图 11**：配置变更时间线。虚线显示已创建但未提交的配置 entry，实线显示最近提交的配置 entry。leader 首先在其自己的 log 中创建了 $C_{old, new}$  配置 entry，然后将其提交到 $C_{old, new}$ （$C_{old}$  的大多数和 $C_{new}$  的大多数）。之后它创建了 $C_{new}$  的 entry 并且将之提交到$C_{new}$  的大多数中。并不存在任何一个 $C_{old}$  和 $C_{new}$  都能独立的做出决定的时间点。

一旦 $C_{old, new}$  被提交后，不论是 $C_{old}$  还是 $C_{new}$ 都不能在拿到对方的许可前做出任何决定， 而 Leader Completeness Property 则确保了只有处于 $C_{old, new}$ 配置下的服务器，才能被选为 leader。现在当 leader 创建 $C_{new}$ 的 log entry 并尝试复制到集群时，就是安全的了。同样，一旦看到此配置，它将在每个服务器上生效。当新配置根据 $C_{new}$ 的规则被提交后，旧的配置就无关了，而未处于新配置中的服务器就可以被关闭。如图 11 所示，$C_{old}$ 和 $C_{new}$ 都不能单方面做出决定；这保证了安全。

对于重新配置，还有三个问题需要解决。第一个问题是，新加入的服务器可能还没有初始化存储任何 log entries。假如它就以这样的状态加入集群，那么需要很长一段时间才能赶上其他的服务器，在此期间可能无法提交新的 log entries。为了避免这种可用性缺口，Raft 在配置变更前引入了一个额外的阶段，来让新的服务器以非投票成员（non-voting members）的角色加入集群（leader 仍然会复制 log 给它们，但它们不会被认为是大多数中的一员）。一旦这个新服务器赶上了集群中其他的服务器，配置变更过程就能够按先前描述的方式进行。

第二个问题是，集群的 leader 也许并不是新配置中的一员。在这种情况下，一旦当 leader 的  $C_{new}$ log entry 被提交之后，它就会退出（回到 follower 状态）。这意味着会有一段时间里（在 leader 提交  $C_{new}$ 的这段时间），有一个不属于集群的 leader 在管理着整个集群；它一边在复制 log，而复制过程中自己却不被计算在大多数中。当  $C_{new}$ 被提交后，leader 过渡将开始，因为这是新配置可以独立运行的第一个时间点（总是能从 $C_{new}$ 中选举一个 leader）。在这个时间点以前，可能只有来自 $C_{old}$ 的服务器才能当选领导人。

第三个问题是，移除服务器（不在  $C_{new}$ 当中的）可能会扰乱整个集群。这些服务器将不再会收到心跳，因此他们会超时，并发起新的选举。之后他们会发送包含了新的 term 编号的 RequestVote RPC，这将导致当前的 leader 被迫回退到 follower 状态。一个新的 leader 最终将被选出，但这些被移除的服务器会再次超时，而这样的流程会不断的重复，最终导致可用性变差。

为了避免这一问题，当服务器认为当前 leader 存在时，会忽略掉 RequestVote RPC。特别的，如果某个服务器在接收到当前 leader 的消息的最小选举超时时间内收到了一个 RequestVote RPC，它将不会更新自己的 term，也不会为其投票。这并不影响正常的选举，每一个服务器在发起选举之前，都至少会等待足够的最小选举超时时间。然而，这种办法能帮助避免由已被移除的服务器带来的干扰：如果一个 leader 能够从它的集群中获得心跳，那么它就不会被更大的 term 编号推翻。



## 7 Log 压缩

Raft的 log 在正常操作期间会增长，以包含更多的客户端请求，但在实际系统中，它不能不受限制地增长。随着 log 的不断变长，它会占用更多的空间，并需要花费更多的时间来重放（replay）。如果没有某种机制来丢弃 log 中积累的过时信息，这最终会导致可用性问题。

快照是实现压缩的最简单的方法。当进行快照时，整个当前系统的状态会被写入一个稳定存储的 *snapshot （快照）*中，之后该点以前的所有 log 都会被丢弃。在 Chubby 和 ZooKeeper 中都用到了快照技术，本节接下来的内容就会描述 Raft 中的快照。

当然某些增量压缩的方法，例如 log cleaning[36] 和 log-structured merge trees[30, 5] 也是可行的。这类方法每次操作数据的一小部分，因此压缩的负荷会在时间上更均匀的分布。它们首先选取一个累积了许多删除、覆盖对象的数据区，之后重写存活的对象并对其进行压缩，然后释放该数据区。与快照相比，这需要更多的机制和复杂性，快照通过始终对整个数据集进行操作简化了问题。log cleaning 要求对 Raft 进行修改，不过状态机能够使用现有的快照接口来实现 LSM tree。

图 12 展示了 Raft 中快照的基本原理。每个服务器都会独立的构建快照，包含 log 中已经提交的 entries。大部分的工作是由状态机将其当前状态写入快照。Raft 也在快照中包含一些小规模的元数据：最近包含的 index 是 log 中最后一个被快照替换的 log entry 的 index（最近的一个被状态机应用的 entry），最近包含的 term 是该 entry 所处的 term。保留这些元数据的目的是为了支持在做完快照之后的第一次 AppendEntries 的一致性检查，因为一致性检查需要之前最近的一个 entry 的 index 与 term 信息。为了使集群配置变更能正常进行（第六节），快照同样也会包含 log 中最近一次配置的 index。一旦某个服务器完成了写快照，它就可能会删掉所有包含在快照中的 log entries，以及先前的快照。

{% asset_img 12.png %}

**图 12**：一个服务器将其 log 中已提交的 entries （index 1~5）替换为一个新的快照，新快照只存储当前状态（在这个例子中是变量 x 和 y）。这个快照最后包含的 index 和 term 用于将快照定位至 log entry 6 之前。

尽管服务器通常会独立的生成快照，leader 还必须偶尔将其快照发送给进度落后的 follower。这发生在当 leader 丢弃了需要发送给某个 follower 的下一个 log entry 之后。幸运的是，这种场景不是一个常见的操作：一个已经赶上 leader 的 follower 已经包含了这个 entry。然而，某个发生异常导致缓慢的 follower 或者是一个新加入集群的 follower 则不包含该 entry。一种让这些 follower 能追上 leader 进度的方法就是通过网络给它们发送快照。

leader 使用一种新的称为 InstallSnapshot 的 RPC 来给落后太多的 follower 发送快照；见图 13。当一个 follower 通过 InstallSnapshot RPC 接收到了一个快照时，它必须决定如何处置自己已经存在的 log entries。通常在快照中都会包含一些在接收者 log 中不存在的新信息。在这种情况下，follower 会丢弃它自己的所有 log；它全部被快照取代，并且可能包含一些与快照冲突的未提交 entries。相反，假如 follower 收到了一个描述了其前缀 log 的快照（由于重发或是错误），那么被快照覆盖的 log entries 会被删掉，而在快照之后的 entries 仍然有效且必须被保留。

{% asset_img 13.png %}

**图 13**：InstallSnapshot RPC 的总结。快照被分割为多个块用于传输；这种 RPC 中的每一个块都会让 follower 认为是一个集群存活的迹象，因此每一次 follower 都可以重置其选举超时定时器。

这种快照方法背离了 Raft 的 strong leader 原则，因为 follower 可以在 leader 不知情的情况下生成快照。然而，我们认为这种背离是合理的。虽然有一个 leader 有助于避免达成共识时出现的决策冲突，但当生成快照时，共识早已经达成，因此并没有决策冲突。数据仍然是从 leader 流向 follower，只是现在 follower 能够重新组织他们的数据。

我们考虑一种基于 leader 的替代方案，在这种方案中只有 leader 才能创建快照，之后它会将这些快照发送给每一个 follower。但这种方法存在两个缺点。首先，发送快照会浪费网络带宽并拖慢整个快照机制的流程。每个 follower 已经包含了产生它自己的快照的所需信息，而且通常服务器从自己的本地状态中生成快照要比在网络中发送并接收快照的成本更低。其次，这种方案中 leader 的实现要复杂的多。比如，leader 需要将快照发送给 follower 的同时，进行新增 log entries 的复制，以免阻塞客户端请求。

还有两个问题会影响到快照机制的性能。首先，服务器必须决定何时生成快照。如果生成的太频繁，那么会浪费磁盘空间和能源；而如果生成的太稀疏，则有耗尽存储容量的风险，并且也会增加在重启时重放 log 所需要的时间。一个简单的策略是当 log 达到了一个固定的字节大小时生成快照。如果此大小设置为明显大于快照的预期大小，则快照的磁盘带宽开销将很小。

第二个性能问题是写快照可能需要花费大量的时间，而我们并不期望这会影响到正常的运行。解决办法是采用写时复制的技术，以便在不影响正在写的快照的情况下接受新的更新。例如，使用函数式数据结构构建的状态机自然支持这一点。或者，操作系统的写时复制支持（例如，Linux 上的 fork）可用于创建整个状态机的内存快照（我们的实现就使用了这种方法）。	



## 8 客户端交互

这一节描述 Raft 如何与客户端交互，包括客户端如何找到集群 leader 以及 Raft 如何支持可线性化语义[10]。这些问题适用于所有的基于共识的系统中，而 Raft 的解决方案与其他系统类似。

Raft 的客户端会将所有的请求都发给 leader。当客户端启动时，它会随机连接一台主机。假如这台主机不是 leader，那么它将拒绝客户端请求并提供最近的 leader 信息（AppendEntries 请求中包含了 leader 的网络地址）返回给客户端。假如 leader 崩溃，客户端请求会超时；那么客户端会再次尝试随机连接一台其他的主机。

我们对 Raft 的目标是能实现可线性化语义（在调用和响应之间的某个点上，每个操作似乎都会立即执行一次）。然而，据已有的描述我们可知，Raft 会多次执行同一个命令：比如，如果 leader 在提交了 log entry 但在响应客户端之前崩溃了，那么客户端就会向新的 leader 重发请求，导致该请求被执行两次。客户端的解决办法是在每一条命令中都携带一个不重复的流水号。那么，状态机会追踪每个客户端最近执行的流水号以及相关的响应。假如它收到了一个命令发现其流水号表明已经执行过了，那么它就会立即返回，而不是再次执行一遍命令。

只读操作可以不用给 log 添加任何信息而直接处理 。然而，如果没有额外的措施，这将有返回过期数据的风险，因为响应请求的 leader 可能已在其不知情的情况下被新 leader 取代了。线性化读取不应该返回旧的数据，Raft 需要两个额外的预防措施，而不是用 log 来保证这一点。首先，leader 必须拥有最新的哪些 entries 已经被提交的信息。 Leader Completeness Property 能保证 leader 拥有所有已提交的 entries，但在其 term 刚开始时，它可能并不知道这些已提交的 entries 都是谁。为了搞清楚这一点，leader 需要提交一个它当前 term 的 entry。Raft 通过让每个 leader 在 term 刚开始时都向其自己的 log 提交一个空操作的 entry 来解决这个问题。之后，leader 在处理一个只读请求前，必须要检查自己是否被罢免了（假如一个更近的 leader 刚刚被选举，那么它自己的信息就可能是过期的）。Raft通过让 leader 在响应只读请求之前与集群的大多数成员交换心跳消息来处理此问题。或者，leader 可以依赖心跳机制来提供一种 lease（租约）[9]，但为了安全，这将依赖于时间（假定时钟偏移是有限的）。



## 9 实现和评估

我们已经实现了 Raft 用于为 RAMCloud [33] 保存配置信息的复制状态机的一部分，来协助 RAMCloud 协调器的故障转移。这一 Raft 实现包含了除测试，注释，空行以外的大约 2000 行 C++ 代码。源码可以免费获取[23]。现在也有大约 25 个独立的基于本文草稿的第三方 Raft 开源实现[34]，它们处于多个开发阶段中。同样的，很多公司也都部署了基于 Raft 的系统[34]。本节后面的内容会采用三个关键点来评估 Raft：易懂性，正确性和性能。

### 9.1 易懂性

为了衡量 Raft 对比 Paxos 的易懂性，我们进行了一场实验性研究，参与者为斯坦福大学高阶操作系统课程与加州大学伯克利分校的分布式计算课程的高年级研究生。我们分别录制了 Paxos 和 Raft 的教学视频，并创建了一些相关的测验。Raft 课程包含了本文中除了 log 压缩以外的所有内容；Paxos 的课程则包含了足够的材料来创建等效的复制状态机，包括单一判定 Paxos，多判定 Paxos，重配置，以及在实践中需要的一些优化（比如 leader 选举）。小测验则会考察对算法的基本理解，还要求学生对极端情况进行推理。每个学生先看一个视频，参与相关的小测验，然后再看另一个视频，在完成另一个测验。大约一半的参与者先完成 Paxos 的部分而另一半则先完成 Raft 的部分，以便于将参与者的个体差异以及可能从第一部分的学习中获得的经验所导致的影响考虑在内。我们对比参与者在每一场测验中的成绩来判定该参与者是否对 Raft 理解的更好。

我们尝试尽可能公平的比较 Paxos 和 Raft。该实验在两方面更有利于 Paxos：43 个参与者中有 15 人都报告说他们以往有一些 Paxos 的经验，而且 Paxos 的视频，相对 Raft 的视频长了 14%。通过表 1 的总结可以得知，我们已采取措施减少潜在的偏见来源。我们所有的资料都可以被检查[28, 31]。

{% asset_img t1.png %}

**表 1**：研究中可能对 Paxos 产生偏见的担忧，针对每种偏见采取的措施，以及可用的其他材料。

平均上，参与者在 Raft 测验中的得分会比在 Paxos 测验中的得分高 4.9 分（总分60分，Raft平均得分为25.7，Paxos平均得分为20.8）；图 14 展示了他们单独的成绩。配对 t-检验表明，在 95% 的置信度下，Raft 分数的真实分布的平均值至少比 Paxos 分数的真实分布大 2.5 分。我们也创建了一个线性回归模型基于如下三个因素来预测新学生的测验成绩：他们做的哪一套测验，他们对于 Paxos 的先期经验，以及他们学习这两个算法的顺序。该模型预测，测验类型的选择产生了 12.5 分的差异，这有利于Raft。这比观察到的 4.9 分的差异大得多，那是因为很多实际的学生都有一些 Paxos 的先期经验，这对 Paxos 有很大的帮助，而对 Raft 帮助甚小。有趣的是，该模型还预测出对于已经参与过 Paxos 测验的人，他们的在 Raft 的成绩会低 6.3 分；尽管我们不知道为什么，但这确实在统计上是显著的。

{% asset_img 14.png %}

**图 14**：一个比较 43 个参与者在 Raft 和 Paxos 测验中的表现的散点图。对角线上方的点（33）代表该参与者在 Raft 的分数更高。

在参与者完成了测验之后，我们还对他们做了调查，询问有关哪一个算法让他们觉得更容易实现或解释；这一结构可见图 15。绝大多数的参与者都报告称 Raft 要更容易实现和解释（41 个问了问题的人中的 33 个）。然而，这种自主报告与参与测验成绩相比，还是更不可靠一些，并且参与者也有可能因为我们认为 Raft 更易于理解这一假设而受到影响。

{% asset_img 15.png %}

**图 15**：采用5分制，要求参与者（左）认为哪种算法更容易在一个正常、正确和高效的系统中实现，以及（右）哪种算法更容易向 CS 研究生解释。

更详细的关于 Raft 用户研究的讨论可见[31]。

### 9.2 正确性

我们已经为第五节中描述的共识机制制定了正式规范和安全性证明。正式规范[31] 使用 TLA+ 规范语言[17] 使图 2 中总结的信息完全精确。作为证明主体，它大约有 400 行长。它本身对任何实现 Raft 的人都很有用。我们已经使用 TLA 证明系统[7] 机械地证明了 Log Completeness Property。然而，该证明依赖于未经机械检查的不变量（例如，我们尚未证明规范的类型安全性）。此外，我们还编写了State Machine Safety Property 的非正式证明[31]，该证明完整（仅依赖于规范）且相对精确（长度约 3500 字）。

### 9.3 性能

Raft 的性能与其他共识算法如 Paxos 类似。对性能而言，最重要的场景就是何时一个已经建立好的 leader 能开始复制新的 log entries。Raft 采用最小消息数量法（从 leader 到一半的集群的一次往返）实现了这一场景。在将来改善 Raft 的性能也是有可能的。比如，很容易就能够通过对请求进行批次化、流水线化来提高吞吐，降低延迟。在其他算法的文献中，已经提出了许多的优化手段；而它们当中许多都能应用在 Raft 上，我们将这些工作留给将来。

我们使用自己实现的 Raft 来测量 Raft leader 选举算法的性能，并且回答了两个问题。一，选举过程会迅速收敛吗？二，在 leader 崩溃之后，可以实现的最小停机时间是多少？

为了测量 leader 选举，我们在一个五主机的集群中反复的让 leader 崩溃然后记录集群经过多久会检测到问题并选举一个新的 leader（如图 16）。为了构造最糟情况的场景，每个试验中的服务器都有不同的 log 长度，因此某些 candidate 是不能成为 leader 的。此外，为了“鼓励” 投票分裂，我们的测试脚本会在 leader 停止运行前触发一次心跳 RPC 的同步广播（这类似于 leader 在宕机前进行了一次新 log entry 复制的行为）。leader 在心跳间隔内均匀的随机宕机，这大约是所有测试当中的最小选举超时时间的一半。因此，可能的最小停机时间约为最小选举超时时间的一半。

{% asset_img 16.png %}

**图 16**：检测并替换崩溃的 leader 的时间。顶部的图表改变选举超时的随机性，底部的图表缩放最小选举超时。每行代表 1000 次试验（“150–150毫秒” 的100次试验除外），并对应于特定的选举超时选择；例如，“150–155ms” 意味着选举超时是在 150ms 和 155ms 之间随机均匀选择的。测量是在一个由五台服务器组成的集群上进行的，广播时间大约为 15 毫秒。九台服务器组成的集群的结果类似。

图 16 中的顶图显示，少量随机化选举超时足以避免选举中的投票分裂。在没有随机性的情况下，在我们的测试中，由于存在许多投票分裂，leader 选举持续花费了 10 秒以上的时间。增加 5 毫秒的随机性有很大帮助，使得平均停机时间为 287 毫秒。使用更多的随机性可以改善最坏情况下的行为：使用 50 毫秒的随机性，最坏情况下的完成时间（超过1000次试验）为 513 毫秒。

图 16 中的底图显示，可以通过减少选举超时来减少停机时间。当选举超时时间处于 12-24 毫秒时，leader 选举平均只需 35 毫秒（最长一次时间为 152 毫秒）。然而，将超时时间降低到低于这一点会违反 Raft 的时间要求：在其他服务器开始新的选举之前，leader 很难广播心跳。这可能会导致不必要的 leader 变更，降低总体系统可用性。我们建议使用保守的选举超时，如 150–300 毫秒；此类超时不太可能导致不必要的 leader 变更，并且仍将提供良好的可用性。



## 10 相关的工作

有非常多于共识算法相关的出版物，它们大多数都能归到以下的类目中：

- Lamport 对 Paxos 最早的描述[15]，以及尝试将之解释的更清楚[16, 20, 21]。
- 对 Paxos 的详细说明，填补了缺失的细节，对算法进行了修改来为实现算法提供更好的基础[26, 39, 13]。
- 实现了共识算法的系统，例如 Chubby [2, 4]， ZooKeeper [11, 12]， 和 Spanner [6]。Chubby 和 Spanner 的算法细节并没有公开，但它们都声称自己基于 Paxos。ZooKeeper 的算法已经公开了许多细节，但和 Paxos 还是有许多不同。
- 能够应用在 Paxos 上的性能优化 [18, 19, 3, 25, 1, 27]。
- Oki 和 Liskov 的 Viewstamped Replication (VR)，与 Paxos 相同时期的共识算法替代。其最初的描述[29] 与分布式事务的协议交织在一起，不过在后续的一次更新中，其核心共识协议被拆分开了[22]。VR 使用基于 leader 的方式，且和 Raft 有许多相似的地方。

Raft 与 Paxos 最大的不同就是 Raft 的强领导性（strong leadership）：Raft 将 leader 选举作为共识协议中的基础部分，并且它尽可能的专注于 leader。这种方式得到了一个更简单、更易于理解的算法。例如，在 Paxos 中，leader 选举与基本的共识协议之间是正交的：它只是一种性能优化的手段而对达成共识而言并不是必须的。然而，这就导致了需要额外的机制：Paxos 在基本的共识协议中包含了两阶段协议，而又有一个额外独立的机制用于 leader 选举。相反，Raft 将 leader 选举直接融入共识算法中而且将其用于两阶段共识的第一个阶段当中。这就得到了比 Paxos 更少的机制。

和 Raft 一样，VR 和 ZooKeeper 也是基于 leader 的，因此与 Paxos 相比，Raft 有许多优势。然而，Raft 的机制少于 VR 或ZooKeeper，因为它将非 leader 的功能最小化。比如，Raft 中，log entries 的流向是单向的：只从 leader 的 AppendEntries RPC 流出。VR 的 log entries 是双向流动的（leader 能在选举流程里接收 log entries）；这就导致了额外的机制和复杂度。ZooKeeper 已经公开的描述中也提到会将 log entries 进/出 leader，但其实现看起来更像 Raft[35]。

据我们了解，Raft 的消息类型比任何其他基于共识的 log 复制算法都要少。比如，我们计算了 VR 和 ZooKeeper 用于基本共识以及成员变更的消息类型（除了 log 压缩和客户端交互）。VR 和 ZooKeeper 都定义了 10 种不同的消息类型，而 Raft 只有 4 种（两种 RPC 请求，以及它们的响应）。Raft 的消息内容相对其他算法更丰富一点，但它们都比较简单。此外，VR 和 ZooKeeper 的描述中，leader 变更时将会传输完整的 log；因此需要额外的消息类型来优化这些机制，使其实用。

Raft 的 strong leadership 机制简化了算法，但一些性能优化方法就会被排除在外。例如，在 leaderless 的条件下，Egalitarian Paxos (EPaxos) 能实现更高的性能。EPaxos 利用了命令中的可交换性。任何服务器都可以只通过一轮通信来提交一个命令，只要有其他正在并发提交的命令能够与让它承载。但是，如果并发提交的命令不能互相承载，EPaxos 需要额外的一轮通信。因为任何服务器都可能会提交命令，EPaxos 就能很好的负载均衡并且能够实现在 WAN 下比 Raft 更低的延迟。然而，这极大的增加了 Paxos 的复杂性。

对集群成员变更，有许多不同的方法被提出，或在一些其他的工作中被应用，包括 Lamport 最初的提案[15]，VR[22]，以及 SMART[24]。我们为 Raft 选择了联合共识方法是因为它利用了共识协议的其余部分，所以只需要增加少量的额外机制，就能实现成员变更。Raft 不能选择 Lamport 的 α-based 方法，因为它假定可以在没有 leader 的情况下达成共识。而相比于 VR 和 SMART，Raft 的重配置算法的优点在于成员变更可以不受限的在处理任何普通请求时发生；相比之下，VR 会在配置变更时停止所有正常的工作，而 SMART 对未完成请求的数量施加了类似 α 的限制。



## 11 总结

算法的主要设计目标通常要考虑正确性、效率和/或简洁性。尽管这些都是非常有价值的目标，但我们认为易懂性同样重要。只有当开发者将算法转化为实际的实现，上述的的目标才能实现，这将不可避免的偏离和扩展已发布的形式。除非开发者对算法有着深刻的认识，并能对其产生直觉，否则对开发者而言，很难在实现过程中保留其可取的特性。

在这篇文章中我们讨论了分布式共识问题，其中一个被广泛接受但难以理解的算法 Paxos 多年来一直在挑战学生和开发人员。我们开发了一种新的算法，Raft，我们已经展示了它比 Paxos 更加易懂。同时我们也相信 Raft 为系统构建提供了一个更好的基础。让易懂性成为首要目标，改变了我们设计 Raft 的方式；随着设计的进行，我们发现自己重复使用了一些技术，例如问题分解以及简化状态空间。这些技术不仅提升了 Raft 的易懂性，也让我们更容易相信它的正确性。



## 12 致谢

没有如下同仁的支持，我们的用户调研是不可能完成的。他们是：Ali Ghodsi, David Mazieres, 以及 Berkeley 的 CS 294-91 和 Stanford 的 CS 240 的同学们。 Scott Klemmer 帮助我们设计了用户调研，Nelson Ray 为我们提供统计分析方面的建议。用于用户调研的 Paxos 演示文稿深度借鉴了 Lorenzo Alvisi 最初创建的演示文稿。特别感谢 David Mazieres 和 Ezra Hoch 细致的找到了 Raft 的一些 bug。许多人都提供了对本文以及用户调研材料很有帮助的反馈，他们包括 Ed Bugnion，Michael Chan，Hugues Evrard，Daniel Giffin，Arjun Gopalan，Jon Howell，Vimalkumar Jeyakumar， Ankita Kejriwal，Aleksandar Kracun，Amit Levy，Joel Martin，Satoshi Matsushita，Oleg Pesok，David Ramos，Robbert van Renesse，Mendel Rosenblum， Nicolas Schiper，Deian Stefan，Andrew Stone，Ryan Stutsman，David Terei，Stephen Yang，Matei Zaharia，24名匿名会议评论员（有重复），特别是我们的导师 Eddie Kohler。Werner Vogels 在推特上发了一条链接，链接到了一份早期的草稿，这给了 Raft 很大的曝光率。这项工作得到了 Gigascale 系统研究中心和 Multiscale 系统中心的支持，这两个研究中心是由半导体研究公司 Focus Center Research Program 资助的六个研究中心中的两个，Focus Center Research Program 由 STARnet 资助，Focus Center Research Program 由 MARCO 和 DARPA 赞助，由国家科学基金会授予第 0963859 号资助，并由Facebook、Google、Mellanox、NEC、NETAPP、SAP 和三星赠款。Diego Ongaro 得到了Junglee Corporation 斯坦福研究生奖学金的支持。



## 参考文献

[1] BOLOSKY, W. J., BRADSHAW, D., HAAGENS, R. B., KUSTERS, N. P., AND LI, P. Paxos replicated state machines as the basis of a high-performance data store. In Proc. NSDI’11, USENIX Conference on Networked Systems Design and Implementation (2011), USENIX, pp. 141–154.

[2] BURROWS, M. The Chubby lock service for looselycoupled distributed systems. In Proc. OSDI’06, Symposium on Operating Systems Design and Implementation (2006), USENIX, pp. 335–350. 

[3] CAMARGOS, L. J., SCHMIDT, R. M., AND PEDONE, F. Multicoordinated Paxos. In Proc. PODC’07, ACM Symposium on Principles of Distributed Computing (2007), ACM, pp. 316–317.

[4] CHANDRA, T. D., GRIESEMER, R., AND REDSTONE, J. Paxos made live: an engineering perspective. In Proc. PODC’07, ACM Symposium on Principles of Distributed Computing (2007), ACM, pp. 398–407.

 [5] CHANG, F., DEAN, J., GHEMAWAT, S., HSIEH, W. C., WALLACH, D. A., BURROWS, M., CHANDRA, T., FIKES, A., AND GRUBER, R. E. Bigtable: a distributed storage system for structured data. In Proc. OSDI’06, USENIX Symposium on Operating Systems Design and Implementation (2006), USENIX, pp. 205–218. 

[6] CORBETT, J. C., DEAN, J., EPSTEIN, M., FIKES, A., FROST, C., FURMAN, J. J., GHEMAWAT, S., GUBAREV, A., HEISER, C., HOCHSCHILD, P., HSIEH, W., KANTHAK, S., KOGAN, E., LI, H., LLOYD, A., MELNIK, S., MWAURA, D., NAGLE, D., QUINLAN, S., RAO, R., ROLIG, L., SAITO, Y., SZYMANIAK, M., TAYLOR, C., WANG, R., AND WOODFORD, D. Spanner: Google’s globally-distributed database. In Proc. OSDI’12, USENIX Conference on Operating Systems Design and Implementation (2012), USENIX, pp. 251–264.

[7] COUSINEAU, D., DOLIGEZ, D., LAMPORT, L., MERZ, S., RICKETTS, D., AND VANZETTO, H. TLA+ proofs. In Proc. FM’12, Symposium on Formal Methods (2012), D. Giannakopoulou and D. M´ery, Eds., vol. 7436 of Lecture Notes in Computer Science, Springer, pp. 147–154. 

[8] GHEMAWAT, S., GOBIOFF, H., AND LEUNG, S.-T. The Google file system. In Proc. SOSP’03, ACM Symposium on Operating Systems Principles (2003), ACM, pp. 29–43. 

[9] GRAY, C., AND CHERITON, D. Leases: An efficient faulttolerant mechanism for distributed file cache consistency. In Proceedings of the 12th ACM Ssymposium on Operating Systems Principles (1989), pp. 202–210. 

[10] HERLIHY, M. P., AND WING, J. M. Linearizability: a correctness condition for concurrent objects. ACM Transactions on Programming Languages and Systems 12 (July 1990), 463–492. 

[11] HUNT, P., KONAR, M., JUNQUEIRA, F. P., AND REED, B. ZooKeeper: wait-free coordination for internet-scale systems. In Proc ATC’10, USENIX Annual Technical Conference (2010), USENIX, pp. 145–158. 

[12] JUNQUEIRA, F. P., REED, B. C., AND SERAFINI, M. Zab: High-performance broadcast for primary-backup systems. In Proc. DSN’11, IEEE/IFIP Int’l Conf. on Dependable Systems & Networks (2011), IEEE Computer Society, pp. 245–256. 

[13] KIRSCH, J., AND AMIR, Y. Paxos for system builders. Tech. Rep. CNDS-2008-2, Johns Hopkins University, 2008. 

[14] LAMPORT, L. Time, clocks, and the ordering of events in a distributed system. Commununications of the ACM 21, 7 (July 1978), 558–565. 

[15] LAMPORT, L. The part-time parliament. ACM Transactions on Computer Systems 16, 2 (May 1998), 133–169. 

[16] LAMPORT, L. Paxos made simple. ACM SIGACT News 32, 4 (Dec. 2001), 18–25. 

[17] LAMPORT, L. Specifying Systems, The TLA+ Language and Tools for Hardware and Software Engineers. AddisonWesley, 2002. 

[18] LAMPORT, L. Generalized consensus and Paxos. Tech. Rep. MSR-TR-2005-33, Microsoft Research, 2005.

[19] LAMPORT, L. Fast paxos. Distributed Computing 19, 2 (2006), 79–103. 

[20] LAMPSON, B. W. How to build a highly available system using consensus. In Distributed Algorithms, O. Baboaglu and K. Marzullo, Eds. Springer-Verlag, 1996, pp. 1–17. 

[21] LAMPSON, B. W. The ABCD’s of Paxos. In Proc. PODC’01, ACM Symposium on Principles of Distributed Computing (2001), ACM, pp. 13–13. 

[22] LISKOV, B., AND COWLING, J. Viewstamped replication revisited. Tech. Rep. MIT-CSAIL-TR-2012-021, MIT, July 2012. 

[23] LogCabin source code. http://github.com/ logcabin/logcabin

[24] LORCH, J. R., ADYA, A., BOLOSKY, W. J., CHAIKEN, R., DOUCEUR, J. R., AND HOWELL, J. The SMART way to migrate replicated stateful services. In Proc. EuroSys’06, ACM SIGOPS/EuroSys European Conference on Computer Systems (2006), ACM, pp. 103–115. 

[25] MAO, Y., JUNQUEIRA, F. P., AND MARZULLO, K. Mencius: building efficient replicated state machines for WANs. In Proc. OSDI’08, USENIX Conference on Operating Systems Design and Implementation (2008), USENIX, pp. 369–384. 

[26] MAZIERES , D. Paxos made practical. http: //www.scs.stanford.edu/˜dm/home/ papers/paxos.pdf, Jan. 2007. 

[27] MORARU, I., ANDERSEN, D. G., AND KAMINSKY, M. There is more consensus in egalitarian parliaments. In Proc. SOSP’13, ACM Symposium on Operating System Principles (2013), ACM. 

[28] Raft user study. http://ramcloud.stanford. edu/˜ongaro/userstudy/. 

[29] OKI, B. M., AND LISKOV, B. H. Viewstamped replication: A new primary copy method to support highly-available distributed systems. In Proc. PODC’88, ACM Symposium on Principles of Distributed Computing (1988), ACM, pp. 8–17. 

[30] O’NEIL, P., CHENG, E., GAWLICK, D., AND ONEIL, E. The log-structured merge-tree (LSM-tree). Acta Informatica 33, 4 (1996), 351–385. 

[31] ONGARO, D. Consensus: Bridging Theory and Practice. PhD thesis, Stanford University, 2014 (work in progress).http://ramcloud.stanford.edu/˜ongaro/ thesis.pdf. 

[32] ONGARO, D., AND OUSTERHOUT, J. In search of an understandable consensus algorithm. In Proc ATC’14, USENIX Annual Technical Conference (2014), USENIX. 

[33] OUSTERHOUT, J., AGRAWAL, P., ERICKSON, D., KOZYRAKIS, C., LEVERICH, J., MAZIERES ` , D., MITRA, S., NARAYANAN, A., ONGARO, D., PARULKAR, G., ROSENBLUM, M., RUMBLE, S. M., STRATMANN, E., AND STUTSMAN, R. The case for RAMCloud. Communications of the ACM 54 (July 2011), 121–130. 

[34] Raft consensus algorithm website. http://raftconsensus.github.io. 

[35] REED, B. Personal communications, May 17, 2013. 

[36] ROSENBLUM, M., AND OUSTERHOUT, J. K. The design and implementation of a log-structured file system. ACM Trans. Comput. Syst. 10 (February 1992), 26–52. 

[37] SCHNEIDER, F. B. Implementing fault-tolerant services using the state machine approach: a tutorial. ACM Computing Surveys 22, 4 (Dec. 1990), 299–319. 

[38] SHVACHKO, K., KUANG, H., RADIA, S., AND CHANSLER, R. The Hadoop distributed file system. In Proc. MSST’10, Symposium on Mass Storage Systems and Technologies (2010), IEEE Computer Society, pp. 1–10. 

[39] VAN RENESSE, R. Paxos made moderately complex. Tech. rep., Cornell University, 2012.
