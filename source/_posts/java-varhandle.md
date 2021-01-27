---
title: 使用 VarHandle 进行更细致 lock-free 编程
date: 2021-01-27 22:45:22
tags: 
- java
- VarHandle
categories:
- Java
---

从 Java 9 开始，jdk 中提供了一个全新的类用于部分替代原本 `Unsafe` 所承担的偏底层的工作，即 `java.lang.invoke.VarHandle`。

新颖的是，在 `VarHandle` 中定义了对变量以不同的 Memory Order 来进行访问的几种模式：

`Plain`，`Opaque`，`Acquire/Release`，`Volatile`

通过这几种不同的模式，在无锁编程（lock-free）中能够更灵活的控制对共享变量的操作，也能基于此更细致的进行设计从而提升程序的性能。

那么，该如何理解 `VarHandle` 、Memory Order、lock-free 这些概念之间的关系呢？接下来我们会从底层说起，一步步将他们串起来。

## CPU 缓存的麻烦

通常 CPU 是无法直接访问内存的，物理结构上，CPU 和内存之间，还隔着一层缓存（Cache）。缓存的目的与我们在应用服务与数据库之间隔一层缓存的目的一样：CPU 和内存的速度差距太大了。而且为了能平滑连接高速设备与低速设备，缓存层通常会分出 L1 L2 L3 层，容量逐层递增，访问速度逐层递减，最后才是内存、硬盘，整体结构像一座金字塔。

{% asset_img cache-spped.png %}

上面这幅图（[来源](https://www.quora.com/How-fast-can-the-L1-cache-of-a-CPU-reach)）展示了缓存与内存的速度对比，可以看到，L1 Cache 的 latency 只有 0.8ns，而 Memory 有 45.5ns，速度相差 57 倍之多，试想如果 CPU 直接与 Memory 交互，整个世界估计都会变慢。

### 古早的缓存结构

早先的 CPU，只有一个核心，也只有一个缓存，呈现如下的结构：

{% asset_img 1-1cache.png %}

对于这种结构而言，Cache 的使用非常简单：读写都优先通过缓存。

读的时候：

- 假如缓存中恰好有需要的数据（Cache Hit）则直接返回
- 假如没有需要的数据（Cache Miss）则继续向下，从 Memory 中读取

写的时候：

- 先写入缓存，再继续向下写入 Memory。这种同步更新的策略叫 **Write-Through**。这种策略简单，并且能保证在任何时间里，Cache 与 Memory 的数据都是最新的，但缺点是同步操作很慢。
- 先写入缓存，并标记 dirty，之后异步的进行 "flush"。这种异步更新的策略叫 **Write-Back**。这种策略能极大地提升性能，但 Cache 与 Memory 之间存在一段时间的不一致，有点类似最终一致性。

但无论如何，在这种简单的架构下，不同的线程只是分享了 CPU 不同的时间片，线程与线程间的共享资源同一时刻只可能有一个线程进行访问，能确保资源对所有线程完全可见。

### 真并行 - 多核

事情不会像上一节那么简单的，因为实际上如今的 CPU 都是多核。多核意味着同一时刻可以有多个任务真正并行的执行，对上一节的架构进行扩展，我们可以得到：

{% asset_img n-1cache.png %}

对于这种架构，我们必须要限制 CPU 对 Cache 的访问：同一时刻只能有一个 core 访问 Cache，否则就会产生并发修改，出现不一致，导致程序的结果不可预知。

然而这样做，似乎失去了多核存在的意义，由于缓存的限制，多个核心在读写指令上被强制串行化了，这就好像给读写指令加了一把独占锁，导致性能很差。

### 分布式缓存

如果不能同时访问同一个 Cache，那么我们索性将 Cache 拆分开，给每个 core 都分配一块缓存：

{% asset_img n-ncache.png %}

### 还是不够快

## CPU 实在太快了

### 指令重排

### Memory Barries

## JMM 的抽象

### 抽象的内存划分

### Volatile

### Memory Order

## Lock-Free 编程

