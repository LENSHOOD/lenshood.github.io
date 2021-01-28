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

### CPU 核私有缓存

如果不能同时访问同一个 Cache，那么我们索性将 Cache 拆分开，给每个 core 都分配一块缓存：

{% asset_img n-ncache.png %}

这样每个 core 都能愉快的与自己私有的 Cache 通信，而 Cache 与 Memory 之间通过一条总线连接。有了私有缓存，CPU 不同 core 之间不再需要互相等待了，因此能够高效的并行执行程序。

但同时问题也来了：假如 core 1 在自己的缓存里保存了共享变量 `a = 1`，并且对其进行 `+1` 操作，那么在 core 1 的视角来看，现在`a = 2`。这时候，core 2 也想执行同样的 `+1` 操作，由于 core 2 对 `a` 的读取没有命中，因此从 Memory 中读取出 `a = 1`，之后也执行了 `a = a+1` 。这种情况下，假如 core 2 的缓存优先回写到了 Memory 中，那么 core 1 的更新操作就丢失了。

上述问题在我们日常编程当中经常遇到，解决的办法通常是加锁，通过牺牲一定的效率来确保正确性。那么在 CPU 中呢？

也一样，在这种缓存架构下，为了保证数据的正确性，提出了**缓存一致性协议（Cache Coherence）**来解决问题。

#### Cache Coherence

通常，缓存一致性协议分两类：

- Snoopy 嗅探协议：每个 Cache 上产生的任何操作，都会在总线中广播，而其他 Cache 可以嗅探总线上的消息，从而做出反应
- Directory 目录协议：Snoopy 协议由于有大量的广播消息占用总线，在 core 数量 > 8~16 时会产生严重拥堵。基于目录的协议是选择某一个 core 作为 host，来通过一张目录表保存当前缓存内容的全局状态。但基于目录的协议开销大，速度慢。

#### MESI

MESI 协议是一种经典的 Snoopy 一致性协议，他给所有缓存中保存的值（这里的值指的是以[缓存行Cache Line](https://medium.com/software-design/why-software-developers-should-care-about-cpu-caches-8da04355bb8a#:~:text=A%20cache%20line%20is%20the,region%20is%20read%20or%20written.)为单位）都指定了状态，共有四种：

- **I**nvalid：处于该状态下的数据要么在 Cache 中不存在，要么是存在但已经过期。Invalid 的数据会被 Cache 直接忽略。
- **S**hared：处于该状态下的数据是对 Memory 中数据的干净拷贝（clean copy）。Shared 的数据可以被多个 Cache 所拥有，但只可读，不可写。
- **E**xclusive：处于该状态的数据也是对 Memory 中数据的干净拷贝，但与 Shared 不同的是，Exclusive 下的数据，一定只会被某一个 core 独占。而假如其他 Cache 中可能存在相同的数据，那么当它被某个 core 独占后，其他相同的数据都会变成 Invalid。

- **M**odified：该状态标记的数据是脏数据（dirty）。表明数据在 Cache 内被修改，与 Exclusive 一样，任何与之相同的数据，都会变为 Invalid。此外，Modified 的数据必须在失效之前被写回 Memory。

与单核单缓存 write-back 策略的架构比较，我们发现，I、S、M 状态都能找到相应映射，而 MESI 的独特之处就在于 E 状态。Exclusive 解决了某个 Cache 在需要修改数据之前需要先通知其他 Cache 的问题：当数据处于 E 或 M 时，表明当前的这份数据是唯一有效的，那么我们就能大胆的对其进行修改了。

{% asset_img mesi.png %}

从上面的状态图（[图源](http://www.broadview.com.cn/article/347)）中我们可以看到：

1. 初始状态下，内存地址映射到 Cache 的数据（Cache Line）的状态是 Invalid，当任一 core 1 对该地址进行读取后，状态转至 Exclusive。如果这时有另一个 core 2 对相同地址发起读请求，则 core 1 嗅探到请求后，会复制数据给 core 2，此时两个 Cache 中的数据状态都变为 Shard。
2. 假如 core 1 想要对某初始状态的数据进行修改，则状态会从 Invalid 转至 Modified。假如这时 core 2 想读取该数据，则 core 1 会先将数据 Write-Back，然后复制一份交给 core 2，状态此时变为 Shard。
3. 但如果 2 中的 core 2 不是读，而是想要修改，那么 core 1 也会先拦截 core 2 的这一动作，将数据 Write-Back，之后将数据状态改为 Invalid。之后 core 2 就可以以正常的流程来修改数据。

从 MESI 的定义上看，只要 Cache 之间能遵循协议要求，及时的将操作进行广播，并及时的对总线上发生的事件进行响应，那么就能实现多核 CPU 对多个私有缓存操作的顺序一致性（Sequential Consistency）

### 还是不够快

从前一节我们知道，Cache 之间能够通过一致性协议如 MESI 来确保数据不会被错误的并发修改，那么我们更进一步思考，以 MESI 为例，为了确保顺序一致性：

1. 所有 Cache 都必须能立即对总线中发生的事件进行响应
2. CPU 完全按照程序指令，尽职的将内存操作事件进行广播，并且等待一条指令执行完毕后，再执行下一条

我们当然可以想办法允许 Cache 和 CPU 满足上述条件，但：

1. Cache 可能正在执行其他工作，无法及时响应事件，导致 Cache 之间为了需要相互等待
2. CPU 对 Cache 的操作实际上是一种 IO 操作，那么为了逐条执行指令，这种 IO 操作的指令也必须等待

各种各样的等待导致 CPU 的整体效率还是不够高，处理速度还是不够快。

如果不采取等待的策略，那么：

- 在 Cache 繁忙时接收到的 Invalidation 事件可以通过一个队列保存，在 Cache 空闲后一一处理。那么就需要在 Cache 中增加一个 **Invalidation Queue**
- CPU 无需等待 IO 操作完成，而是在这段时间里执行其他的操作以至于不会浪费周期进行空转。这样的话，从指令执行的角度看，CPU 的执行顺序，可能会与程序指令顺序不一致，即乱序执行。
- 对于 Cache 的写操作，即 store 操作，相比 load 操作要复杂一些，是一个两阶段的过程：首先想要 store 的 core 必须先拿到 Exclusive 状态的所有权，之后再进行写入动作。第一步涉及到一个与其他 core 协商的过程。协商需要等待，因此为了性能考虑，将等待协商的写操作也放入一个队列中，直到可以开始写为止。因此我们还需要在 Cache 中增加一个 **Store Buffer**

最后，我们的 CPU 缓存架构变成了这样：

{% asset_img n-n-queue.png %}

## CPU 实在太快了

### 指令重排

### Memory Barries

## JMM 的抽象

### 抽象的内存划分

### Volatile

### Memory Order

## Lock-Free 编程

