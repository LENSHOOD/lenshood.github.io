---
title: 引入泛型改善 lfring 性能的探索
date: 2022-08-01 23:31:32
tags: 
- lock free ring buffer
- go generics
- performance optimization
categories:
- Go
---

{% asset_img ring.png 300 %}

本文介绍了近期对先前的库 [go-lock-free-ring-buffer](https://github.com/LENSHOOD/go-lock-free-ring-buffer) （简称 lfring）改造泛型而产生的性能优化。
介绍该 lfring 的文章可见[这里](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)。

<!-- more -->

### 1. 引入泛型

随着 go1.18 的发布，等待了 10 年的泛型终于发布了。想想去年写的 [lfring](https://github.com/LENSHOOD/go-lock-free-ring-buffer) 的库，正是因为 go 没有泛型的支持，对数据的存取全部都用了 `interface{}`，导致整个流程中反复的进行类型转换，用户也需要大量类型推断。

虽然在实际当中，可以采用单态化的方式简化编程，改善性能，但作为一个通用的无锁队列，用`interface{}`来传递参数实属没有办法的办法。

泛型发布之后，各种评论褒贬不一，但是从 lfring 的角度看，我觉得引入泛型一定是利大于弊的。那么唯一的问题就是，引入了泛型之后，会不会真的如一些文章的测试结果所展示的，会产生性能劣化。lfring 本身是想通过 lock-free 的方式来改善性能从而在某些特定场景下替代 `channel` 的，假如引入了泛型导致性能出现劣化，就可能需要多多斟酌了。

#### 1.1 性能测试

性能测试沿用了[之前文章](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)里提到的测试办法：

- 数据源：构造 64 个元素的 int 类型数组，`Offer` 时按顺序从中取出
- 对照组：将 `channel` 按 lfring 的 api 进行封装，作为对比

- 测试指标：
  - 不同的 goroutine 分别进行`Offer` 和 `Poll` 的动作，由于生产消费之间的速度差异，以及不同 goroutine 之间竞争的原因，`Poll` 可能会拿不到值（返回 `success = false`，代表这一次取操作失败，需要重试），需要重试
  - 设置一个 counter，每一次成功的 `Poll` 都能让 counter+1，counter+1 代表了某个数成功的完成了一次`Offer` 到  `Poll` 的 handover
  - 显然，在相同的时间限制内，handover 次数越多，代表性能越高（考虑到一定的随机性，实际测试中采取的是多次采样，去除 3σ 区间外的离群值后取均值的办法）
- 测试类型：通过三个不同的维度变量来测试不同场景下的性能表现
  - 环型队列中的元素容量，从 2 到 1024 变化
  - 生产者、消费者的比例，从 11:1 到 1:11 变化（测试机器一共 12 个逻辑核）
  - 总线程数量，从 2 到 48 变化（采用设置 `GOMAXPROCS` 的办法控制底层线程数，P 和 goroutine 仍旧一一对应。让线程竞争更多的表现在OS 调度器上而不是 go 调度器上）

此外，为了防止 go 版本更新而产生的其他优化对数据造成影响，使用 `interface{}` 和泛型的两套代码都在 `go version go1.19 darwin/amd64` 环境下测试。

平台是2019 款 MBP，2.6 GHz 6-Core Intel Core i7 CPU，超线程后共 12 个逻辑核。



#### 1.2 测试结果

注：下图中的 `NodeBased` 和 `Hybrid` 是 lfring 库中队列的两种不同实现，我在[先前的博客](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)中对其进行了介绍，同时在文中也分析了 `NodeBased`的性能比 `Hybrid` 好的原因 。

1. 环形队列元素容量变化
  {% iframe with-capacity.html 100% 500 %}

  从图中可见，不论是哪一种类型，其容量在 2 和 4 时都没有太大的变化，而后才产生了差异，可以暂且假设在容量为 2 和 4 时，因为其他未明因素限制了发挥。

  除去容量 2 和 4 后，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 5% 到最多13.8%，平均提升约 8.4%，而对`Hybrid`类型的 lfring，泛型比 `interface{}` 性能提升并不明显，从最低 -4% 到最高 7.5%，平均提升约 3.1%。

  但令人吃惊的是，随着容量的不断上升，作为对照组的 `Channel` 的性能近乎与线性增长，并且在 64 之后超出了无锁队列的实现，这一现象将在后文中进行分析。单看 `Channel` 本身，在引入泛型后的性能提升巨大，从最小的 10.7% 一直到 77.1%，从趋势上看甚至如果容量更大时性能提升会更多，平均提升约为 50.4%。 
  
2. 生产者与消费者比例变化
  {% iframe with-producer.html 100% 500 %}

  在容量固定为 32 时，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 5.2% 到最多 27.2%，平均提升约 17.4%，整体上看在生产者比例较低时，性能提升更明显；而对`Hybrid`类型的 lfring，泛型比 `interface{}` 性能提升并不明显，从最低 -0.1% 到最高 15.3%，平均提升约 5%；对照组 `Channel` ，对比情况也类似，在生产者比例较低时，性能提升巨大，而在消费者比例较低时区别不大，从最低 -9% 到最高 134%，平均提升约 54.1%

3. 总线程数变化
  {% iframe with-thread.html 100% 500 %}

  在容量固定为 32 时，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 3% 到最多 35.9%，平均提升约 16.9%；而对`Hybrid`类型的 lfring，除了 2 线程（一个生产者，一个消费者）性能提升较为明显，其他几乎一致，可认为没有明显改善；对照组 `Channel` ，在 2 线程时性能有所降低，但其余情况下性能都有大幅提升，从最低 -12.2% 到最高 109%，平均提升约 46%



### 2. 关于泛型的讨论

#### 2.1 泛型概览

Golang 在 2009 年公开发布的时候，就没有包含泛型特性，从 2009 年开始，一直到 2020 年之间，不断地有人提出泛型的 feature request，有关泛型的各种讨论也层出不穷（参考有关泛型的 [issue](https://github.com/golang/go/issues?q=label%3Agenerics) ）。

为什么 Golang 一直没有引入泛型呢？其作者之一 rsc 在 2009 年的讨论 [The Generic Dilemma](https://research.swtch.com/generic) 提到了一个”泛型困境“：

对于泛型，市面上可见的三种处理方式：

1. 不处理（C ），不会对语言添加任何复杂度，但会让编码更慢（slows programers）
2. 在编译期进行单态化或宏扩展（C++），生成大量代码。不产生运行时开销，但由于单态化的代码会导致 icache miss 增多从而影响性能。这种方式会让编译更慢（slows compilation）
3. 隐式装箱/拆箱（Java），可以共用一套泛型函数，实例对象在运行时做强制转换。虽然共享代码可以提高 icache 效率，但运行时的类型转换、运行时通过查表进行函数调用都会限制优化、降低运行速度，因此这种方式会让执行更慢（slows execution）
4. （rsc 文中忽略的部分，C# 的实现）：



#### 2.2 不同的 GCShape



### 3. 回到 lfring

#### 3.1 简单看看汇编

#### 3.2 改变数据类型



### 4. Channel 迷思

