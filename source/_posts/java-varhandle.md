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

{% asset_img cache-spped.jpeg %}

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

#### Memory reordering

从前文中我们能够知道，CPU 通过操作缓存来优化性能，而为了保持缓存间的一致性，需要通过类似 MESI 的一致性协议来确保缓存中不存在错误的数据。

但是由于 Invalidation Queue 与 Store Buffer 的影响，有些时候：

- 如果在 Invalidation Queue 中暂存的失效事件，被 Cache 响应之前，受影响的数据又被 CPU 读取，导致读到了过期的数据
- 由于 Store Buffer 的存在，Store 的真实执行时间也许会比看起来要晚

由于上述可能发生的场景，导致真实的执行顺序看起来与指令顺序不一致。

#### Compiler Reordering

通常，在程序编译阶段，编译器也会做出大量的优化，其中就包括了对语句的简化，其简化的主要手段就是对相互不依赖的语句进行优化。

比如，对于语句 `a = 1; b = 2; x = a + b;`而言，第三条语句依赖前两条语句的结果，因此他一定要在前两条语句之后执行，才能保证程序的正确性。但对于互相之间没有依赖的语句，就可以随意的进行重排、合并、省略等等操作了，因为他们不会影响到最终程序执行的结果，例如：

```go
a := p.x
...
b := p.x
```

可能会被优化为：

```go
a := p.x
...
b := a
```

但是编译器的优化都只针对单线程程序，多线程程序之间的依赖性，编译器难以获悉。因此假如上述程序的实际执行是这样的：

```go
thread-1:     |    thread-2:
a := p.x			|    ...
...						|		 p.x = p.x + 1
b := p.x			|    ...
```

那么如果还是按照前面的办法进行优化 `b := a`，则真实程序的执行顺序看起来就会像是 `p.x = p.x + 1` 在 `b := p.x` 之后才执行。

#### CPU Reordering

我们都知道现代 CPU 在执行指令时，是以指令流水线的模式来运行的。由于执行每一条指令的过程，都可以大致分解为如下阶段：

取指（Fetch，IF）、译码（Decode，ID）、执行（Execute，EX）、存取（Memory，MEM）、写回（Write-Back， WB）

在 CPU 中不同的执行阶段会涉及到不同的功能部件，那么只要我们将指令阶段拆分开，尽量让每一个部件都不空闲的持续运行，就能使性能达到最优：

{% asset_img five-stage-pipeline.png %}

但现实并不总是我们期望的，假如当出现相互依赖的指令时，例如上述的例子： `a = 1; b = 2; x = a + b;`，由于 `x = a + b` 依赖于 `a` 与 `b` 的加载，那么当该指令在流水线中执行时，发现它依赖的数据还没有准备好，就会将流水线中断，中断会导致流水线中所有的环节开始等待，很影响性能。

因此，既然要等待，我们不如将等待的时间用来执行其他的指令，这样就能最大化的减少中断。指令重排序就能实现这一目的。

### Memory Barries

上一节提到的各个层面的 Reordering，的确能够对性能提升起到极大的帮助，但同时也带来了很多问题。

在多线程场景下，线程间如果存在对共享变量的操作，那么无论是在编译器层面、CPU 层面还是 Cache 层面，如果不做特殊处理，它们都无法保证程序的正确性。

基于此，Memory Barrier 的概念被引入。

Memory Barrier（也称 Memory Fence）是一类 CPU 指令，它能够影响编译器的编译和 CPU 的执行，使得所有在 Memory Barrier 之后的指令，一定不会在 Barrier 之前执行。因此可以认为，Memory Barrier 的引入，能够抑制所有编译器、CPU 以及其他设备的各种优化手段，包括重排序、内存操作的延迟和组合、预测执行、分支预测等等技术。Memory Barrier 能让 CPU 老老实实的按指令顺序执行，当然也就导致了性能的下降，因此通常只有在不得已的情况下才会使用它。

如下几种 Memory Barriers:

1. Write (or store) memory barriers:

   - 只作用于 store 指令。确保所有在 barrier 之前的 store 操作在所有 barrier 之后的 store 操作**之前发生（happen before）**。

   - 从 CPU 的角度看，在所有 barrier 之前的 store 将数据存入内存之后，barrier 后面的 store 才会开始执行。

   - Write memory barrier 通常会与后面将会提到的 Read、Data Dependency barrier 一同使用。

2. Data dependency barriers:
   - 两个有依赖关系的 load 操作，barrier 之前的 load 会在之后的 load 之前发生。
   - 在单 CPU 内相互依赖的 load 似乎没有必要加 barrier，但由于 Invalidation Queue 和 Store Buffer 导致缓存一致性协议只能保证最终一致性，因此其他 CPU 的 Cache 可能会先感知到后面的 load，再感知到前面的 load，导致程序错误。
3. Read (or load) memory barriers:
   - 只作用于 load 指令。确保所有在 barrier 之前的 load 操作在所有 barrier 之后的 load 操作**之前发生（happen before）**。
   - Read barrier 隐含了 data dependency barrier。
4. General memory barriers:
   - 确保所有在 barrier 之前的 load/store 操作在所有 barrier 之后的 load/store 操作**之前发生（happen before）**。
   - 是最强的 barrier，它隐含了所有前三种 barriers。

如下是两种隐含类型，

5. ACQUIRE operations（等同于 LOCK）:
   - 确保所有在 ACQUIRE 之后的内存操作看起来一定在 ACQUIRE 之后发生。
   - 但 ACQUIRE 之前的内存操作也可能在 ACQUIRE 之后发生
6. RELEASE operations（等同于 UNLOCK）:
   - 确保所有在 RELEASE 之前的内存操作看起来一定在 RELEASE 之前发生。
   - 但 RELEASE 之后的内存操作也可能在 RELEASE 之前发生
   - 与 ACQUIRE 操作结合，就不再需要内存屏障

## JMM 的抽象

内存模型（Memory Model）是用于在多处理器环境下定义当前处理器对内存的写入是否对其他处理器可见，以及其他处理器对内存的写入是否对当前处理器可见的充分或必要条件。

不同的 CPU 可能会定义不同的内存模型，通常可以分两类：

1. 强内存模型

   支持强内存模型的 CPU （如 x86 系列）通常会持续追踪正在进行中而未完成的内存操作。即强内存模型的 CPU 能够了解他们所定义的内存模型，并可以当某些操作破坏了其内存模型时回滚数据。因此强内存模型的 CPU 设计起来更复杂，但对编码而言更简单。

2. 弱内存模型

   支持弱内存模型的 CPU（如 ARM， POWER 等）为其重排 load/store 操作提供了更大的空间。并且在多核环境下更容易改变程序原本的意图。因此这种内存模型的 CPU 设计起来更简单，但对编码要求更高（通常需要更多的插入 Memory Barrier 来纠正 CPU 的行为）。

### 抽象的内存划分

Java 语言是运行在 JVM 上的语言，因此 JVM 在设计中需要考虑不同平台下 Java 程序执行的一致性。

因此 Java 在语言层面定义了 Java 的内存模型（Java Memory Model），映射为不同线程之间，哪些操作是合法的，以及线程之间如何通过内存来进行交互。从编译器设计的角度看，JMM 定义了一套规则来约束在什么条件下，不允许对 field 或 monitor 的某些操作指令进行重排序。

通过定义一套语言层面的内存模型，无论 JVM 运行在什么平台下，Java 多线程程序的运行结果都是可以预测的。



### Volatile

### Memory Order

## Lock-Free 编程

## Reference

1. [Cache coherency primer](https://fgiesen.wordpress.com/2014/07/07/cache-coherency/)
2. [The Java Memory Model](http://www.cs.umd.edu/~pugh/java/memoryModel/)
3. [The JSR-133 Cookbook for Compiler Writers](http://gee.cs.oswego.edu/dl/jmm/cookbook.html)
4. [既然CPU有缓存一致性协议（MESI），为什么JMM还需要volatile关键字？](https://www.zhihu.com/question/296949412)
5. [LINUX KERNEL MEMORY BARRIERS](https://www.kernel.org/doc/Documentation/memory-barriers.txt)