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

### 真并行 - 多核

### 分布式缓存

### 还是不够快

## CPU 实在太快了

### 指令重排

### Memory Barries

## JMM 的抽象

### 抽象的内存划分

### Volatile

### Memory Order

## Lock-Free 编程

