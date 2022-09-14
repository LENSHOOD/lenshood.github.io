---
title: lfring 和 channel，到底要怎么选？
date: 2022-09-04 23:41:10
tags: 
- lock free ring buffer
- channel
- comparison
categories:
- Go
---

{% asset_img head-pic.jpg 500 %}

本文分析了 lfring 和 channel 在不同场景下的性能表现，并给出了在哪些场景中引入 lfring 才更佳的建议。

<!-- more -->

---

Lock Free Ring Buffer 系列文章：

1. [一个简单的 Lock Free Ring Buffer，有多简单？](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)
2. [探索引入泛型对 lfring 产生的性能影响](https://lenshood.github.io/2022/08/01/optimize-lfring-performance/)
3. [lfring 和 channel，到底要怎么选？](https://lenshood.github.io/2022/09/04/decide-lfring-channel/)

---

### Generics 与对照组 Channel

在 lfring 的引入 generics 一文中，我们仅对比了在`interface{}` 和 generics 两种场景下，lfring 的性能变化，而没有涉及到对照组 channel 在引入 generics 后的变化。

从前一篇[介绍 generics 对 lfring 影响的文章](https://lenshood.github.io/2022/08/01/optimize-lfring-performance/)中我们已经明确的分析得到，造成 `interface{}` 性能低下的主要原因是由于编译器在转换 `interface{}` 途中产生的一次额外的堆内存分配，而对比了将这一次堆内存分配前移到测试之外（直接传分配好了的对象指针）的办法后，我们会发现传入 `interface{}` 的测试结果已经与 generics 基本一致了。

那么将这一结论扩展到对照组 channel 以后，我们得到了如下的结果：

{% iframe capacity.html 100% 500 %}

{% iframe producer.html 100% 500 %}

{% iframe thread.html 100% 500 %}

对于上述三类测试的详细介绍可参考[这里](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)。

显而易见，对于非指针类型且传入 `interface{}` 的测试 case，在三类测试当中表现都最差，generics 作为对照组，不论是指针还是非指针类型，性能上没有差异。而指针类型且传入`interface{}` 的 case 甚至在某些特殊点上性能更好。

上述测试得出，对于 channel，引入 generics，在存储非指针对象的情况下，性能相比`interface{}` 也有明确的提升。



### Channel 比 lfring 更强？

在前面的第一个 capacity 作为变量的测试里，我们隐约的发现，当 capacity 逐步增加后，channel 的性能稳步提升。这一点在定性分析上是合理的：容量增大意味着池子更大了，也更难以被装满，因此线程之间的竞争就更小了。

再看实测值，似乎比 lfring 性能还要更好。为了进一步的定量分析，我们再次将 lfring 和 channel 放在一起进行测试：

{% iframe capacity-all.html 100% 600 %}

{% iframe thread-all.html 100% 600 %}

{% iframe producer-all.html 100% 600 %}

结合上面三个图我们会发现，在生产者消费者相等，容量变化的场景下，channel 展现了一条近似线性增长的曲线：容量越大，性能越好。同时，元素数超过 64，其性能就已经超越了 lfring。

再看从第二张线程变化的图上，channel 和其他两条 lfring 的曲线区分度不高，只有在单线程和超出 CPU 逻辑核数后，lfring 和 channel 产生了偏离。因此线程数产生的影响相对一致。

而最后第三张展现生产者消费者比例变化的图上，我们发现 lfring 在生产者消费者不均衡（可认为生产端和消费端不等速）时所表现出的性能稳定性更好，而 channel 更适合等速的场景。



### 哪些场景下应该用 lfring？

基于上述测试结果，我们可以假设：

用 channel 来实现队列，更适合大容量、生产消费速度差小的场景，而 lfring 更适合容量偏小，生产消费不等速的场景（实际当中不等速才是日常，不然干嘛要叫 buffer 呢？）

基于上述假设，我们控制线程数不变，在两个轴上分别对队列容量、生产者消费者比例进行改变，得到了如下的三维图：

{% iframe 3d-capacity-producer.html 100% 500 %}

结果是明确的：channel 的性能表现像一个山丘，而 lfring 更像是平原。从中，我们可以得出如下结论：

1. lfring 更适合生产消费不等速的场景，这也相对符合对 buffer 的定义
2. lfring 更适合容量不大的场景
3. 与最初的测试一样，NodeBased 实现性能更好
4. channel 在大容量和生产消费等速的场景下性能非常亮眼，很适合在多个 goroutine 之间高速交换数据，此外 channel 所提供的的有序、排队等特点，也很适合多数的情况

综上我们发现，lfring 所提供的是某些特定情况下更高性能的、可替代 channel 的解决方案。
