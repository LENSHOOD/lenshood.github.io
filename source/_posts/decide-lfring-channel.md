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

{% asset_img head-pic.jpg 300 %}

本文分析了 lfring 和 channel 在不同场景下的性能表现，并给出了在哪些场景中引入 lfring 才更佳的建议。

<!-- more -->

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

