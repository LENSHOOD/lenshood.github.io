---
title: Java 基于 ZooKeeper 实现分布式锁需要注意什么
date: 2020-03-25 23:26:16
tags: 
- distributed lock
- zookeeper
categories:
- Java
---

在前一篇有关 Redis 分布式锁的[文章](https://lenshood.github.io/2020/02/04/redis-distributed-lock/)中，我们讨论了几点有关分布式锁的要求：
1. 操作原子性
2. 可重入性
3. 效率

为了满足上述条件，采用 `本地锁 + Redis 锁` 的方式解决了问题。不过在文章末尾提到，Redis 
不保证强一致性，因此对一致性要求很高的场景会存在安全隐患。

本文将讨论使用满足 CP 要求的 ZooKeeper 来实现强一致性的分布式锁。

### Zookeeper 分布式锁原理
结合 Redis 的分布式锁实现，我们能够想到最直接的 zk lock 实现方式，可能会是以 `ZNode` 来类比 redis 的 kv pair：创建一个 `ZNode`，通过判断其是否存在、以及其值是否与当前 client id 一致来尝试获取一个锁。

然而，结合 zk 的诸多优秀特性，实际上我们能更优雅的实现这一过程：
1. 创建一个路径为 `locknode/{guid}-lock-` 的 znode，同时将之设置为 `EPHEMERAL_SEQUENTIAL`, 其中的 `guid` 是为了解决一种边缘 case*。因此，我们会创建形如 `locknode/{guid}-lock-0000000012` 的一个节点。
2. 尝试获取 `locknode` 下的所有节点，对其进行排序，若刚刚创建的节点处在第一位，则获取锁成功，退出当前流程。
3. 若不为第一位，则对整个序列中排在自己持有的路径前一位的路径添加一个 watcher，并检查该前一位节点是否存在
4. 若前一位节点不存在，跳转至第二步，否则休眠等待。当被 watch 的路径发生变化时（通常是被删除），等待被唤醒并跳转至第二步。


可以看到，上述实现分布式锁的流程，用到了 zk 的两个特性：
1. sequence node
    - 通过 zk 内部保证的序列来确保获取锁公平（回顾 Redis 的方案，每隔 100ms 重试，是一种抢占式的非公平策略）
    - 每一次获取锁的尝试都会被如实的记录下来，易于观察整个获取锁的过程，也易于 debug
2. watcher
    - watcher 避免了轮询，每个等待中的路径都只观察其前一位路径，确保锁释放时只会有一个等待者（而不是所有）被唤醒，避免了羊群效应 （herd effect）。

> 注* guid 的特殊 case：对于`EPHEMERAL_SEQUENTIAL`节点的创建，假设节点创建成功，但 zk server 在返回创建结果之前 crash，那么在 client 重新连接至 zk 后，其 session 仍然有效，因此节点亦存在。
>
> 这时将出现诡异的一幕：某种情况下，该 client 以为自己没有获取到锁（实际上已经拿到了），这时他会再次创建一个 path，并休眠，而另一个 client 一直在等待第一位 path 被释放，但却永远也等不到（本来持有锁的 client 却休眠了）。
>
> 通过给 path 增加 guid 前缀的办法，当 client 检测到 create 非正常返回时，会启动 retry 流程：获取所有 children，若其中包含有 guid 的节点，则认为节点已经创建成功。

### 代码实现
