---
title: TiDB 学习课程 Lesson-5
date: 2020-09-19 22:14:32
tags:
- tidb
categories:
- TiDB
---

这一节，我们涉及的知识是 TiDB 如何实现 DDL。

## TiDB DDL 工作原理

我们知道，TiDB 的元信息，与数据一样是以 k-v 的形式存储在 TiKV 的，DDL 操作，包括了对元数据本身的修改，以及必要时对相关数据的修改。

那么作为一个多实例分布式部署形态的数据集群，TiDB 是如何完成 DDL 操作的呢？他又是如何避免多实例之间不一致的问题呢？TiDB 在 DDL 期间，会像 MySQL 一样产生锁表的动作吗？

我们来一一解答上述问题。

### 任务分发模型

由于 TiDB 本身是无状态的服务，用户的 DDL 请求，可能会发送到任意一个 TiDB 实例上，进而要求对基于元数据考虑分布式一致性问题。为了简化实现，TiDB 采用了 owner 的概念，即每个集群内只有一个实例作为 DDL owner 来真正的执行 DDL 任务，而不论哪个实例接收到 DDL 请求，最终该 DDL 任务的实际执行都会被分发给 owner。

为了实现这一目标，TiDB 的 DDL 操作采用了基于 worker 的任务分发模型实现上述要求。

简单来说，一个 DDL 被处理的流程如下：

1. TiDB 接收到 DDL 请求，将该请求解析并封装为一个 DDL job，扔进 ddl job queue（该队列放在 TiKV）
2. 每一个 TiDB 实例都会启动一个 worker goroutine，用来从 ddl job queue 中领取任务来执行。但注意前文的要求是只有 owner 才能执行任务，因此 worker 会判断自己所在的实例是否是 owner，如果不是则不会领取任务。当 owner worker 执行完任务后，将执行结果扔进 history ddl job queue（该队列放在 TiKV）
3. 发起 DDL 任务的 TiDB 实例会持续等待 history ddl job queue 中返回对应的 job，获取成功后整个 ddl 操作结束。

根据上述描述，一切 DDL 操作都会遵循该流程来执行，这种方式实现了在线无锁 DDL。

下图是 TiDB 官方博客中提供的关于 DDL 执行过程的示意图：

![](https://download.pingcap.com/images/blog-cn/tidb-source-code-reading-17/1.png)

### 异步 Schema 变更原理

TiDB 的 Scheme 变更过程，借鉴了 [Google F1 的异步 Schema 变更算法](http://static.googleusercontent.com/media/research.google.com/zh-CN//pubs/archive/41376.pdf)。

实际上 TiDB 的设计基础本身就是基于 Google 的 Spanner / F1，在 F1 的 Schema 变更论文摘要中就已经提到，他的算法：

1. asynchronous—it allows different servers in the database system to transition to a new schema at different times
2. online—all servers can access and update all data during a schema change.

这种算法不仅不要求各节点 Schema 的状态保持一致，还实现了 Schema 变更完全的无锁，我们可以一边更新数据，一边更新数据表。

我们可以想一想，对于 TiDB 这种分布式数据库，如果期望实现简单的同步带锁的 Schema 变更，似乎也不那么容易：

1. TiDB 节点本身无状态，且没有一个地方可以保存所有节点列表的最新状态（这本身就需要同步），那么想要让所有节点的 Schema 在同一时刻变更到新的状态就会非常复杂，且不可靠。
2. 由于节点众多，时间不统一，数据规模庞大等等多种因素，Schema 变更的时间可能不会太快，如果在此期间与对应表相关的 DML/DQL 不可用，那么将极大地影响正常业务（MySQL 5.6 以后也引入了 Online DDL 来在绝大多数 DDL 场景下保证业务连续性）

那么既然客观条件要求 Schema 变更期间数据可用，同时无法实现状态同步变更，那么异步无锁的 Schema 变更策略就是唯一的选择了。

#### 异步变更存在的问题

从上文中我们已经知道，很难保证所有 TiDB 节点暂存的 Schema 信息都是最新状态的，那么异步变更时就会存在问题：

假如有 A/B/C 三个 TiDB 节点，目前对 table t 增加了一列 c，且当前 A/B 节点都已经更新的该状态，即表 t 包含了列 c，那么业务方此时就可能在 insert t 的时候，加上 c 相关的信息，例如：

`insert into t (c) values('c-value')`

假如该语句被定向到节点 C 时，就会产生错误，因为 C 并不知道表 t 中已经包含了 c 列，这就产生了数据的插入失败。

类似的，当我们进行添加索引操作时，假如部分节点已经获知索引的信息，而另一部分并未获知，则索引列就很可能产生有的值有索引，有的值没有索引的情况。

#### 状态拆分解决不一致问题

既然我们无法确保所有节点在同一时刻达到统一状态，那么有没有一个折中的办法，允许状态不一致的情况暂时存在，并且最终达到一致呢？

F1 引入了中间状态的概念，仍旧以添加列为例：

直接状态变更：

`original ---> column added`

中间状态变更：

`original ---> column added but not valid ---> column added`

对于中间状态变更，我们只要保证，同一时刻所有的节点只能处于两个相邻状态中的一个，例如：

1. A/B/C 三个节点，A/B 节点处于初始状态，C 节点已经变更，进入中间状态，column 已经增加，但不发布。
  - 此时，即使 C 节点已经先添加了列，但并没有发布，从外部看，表仍然是原来的样子，因此所有包含新列的 insert 语句都会报错
2. A/B 节点逐步到达中间状态，此时 C 节点进入最终态，增加 column 的状态被发布出来。
  - 此时，对于 insert 语句中包含新列的请求，如果进入 C 节点，一切按照正常流程，而如果进入 A/B 节点，则由于 A/B 也已经存在新列，因此 insert 语句也可以成功
3. 最后 A/B/C 都达到终态

我们看到，采用这种方式就可以逐步地让整个集群的状态趋于一致。那么如何确保，同一时刻只有两种状态可以共存呢？

在 F1 中，采用 **租约(lease)** 来实现，租约本身有时间限制，节点会强制在租约到期后重新加载 Schema。这种方式是采用宽松的时间段来规避了各节点时间不统一的问题。

假设节点的Schema 租约时间是 30s，A/B/C 节点的本地时间分别有 5s 的延迟误差，因此，A 节点首先刷新租约进入中间状态，5 秒后 B 节点进入，再 5 秒后 C 节点进入，这样就保证了同一时刻只有两个 Schema 状态可以同时存在。另外我们也能看到，假如节点间的时间误差大于租约期时，就可能破坏我们的预期（因此我们可以通过提升时间精度的办法来降低租约期，提升 DDL 速度）。

##### 真实示例

显然，很多 DDL 操作可能会比较复杂，因此实际中，仅有一个中间状态是不现实的，通常至少会有两个中间状态：

- **write only**：变更的 schema 元素仅对写操作有效，读操作不生效
- **delete only**：变更的 schema 元素仅对删除操作有效，读操作不生效

那么实际的添加 column 操作中，状态变更流程是这样的：

```
absent --> delete only --> write only --(reorg)--> public
```

其中，`reorg` 不指代一个状态而是在 write only 状态后，需要做的一系列操作，例如对所有现存数据增加索引，增加默认值等等。

而对于删除列，则正好相反：

```
public --> write only --> delete only --(reorg)--> absent
```

1. public：能够对该列数据进行读写
2. write only：列数据不可读，只可写，确保写操作正常进行（此时仍存在 public 节点，因此写操作仍然需要支持，删除操作会被忽略）
3. delete only：列数据不可读，只可删除，处于该状态的节点，已经不会再有任何被删除列上的新数据被插入了。
4. reorg：此时可以开始逐步删除列上数据，就算有节点处于 delete only 状态，且成功删除了列上的数据，reorg 时只需要忽略删除失败的行即可。
5. absent：最终状态

## TiDB DDL 实现

