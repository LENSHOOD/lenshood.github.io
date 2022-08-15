---
title: 震惊！引入泛型竟能让 lfring 性能提升一倍！
date: 2022-08-01 23:31:32
tags: 
- lock free ring buffer
- go generics
- performance optimization
categories:
- Go
---

{% asset_img ring.png 500 %}

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
  - 环型队列中的元素个数，从 2 到 1024 变化
  - 生产者、消费者的比例，从 11:1 到 1:11 变化（测试机器一共 12 个硬件线程）
  - 总线程数量，从 2 到 48 变化（采用设置 `GOMAXPROCS` 的办法控制底层线程数，P 和 goroutine 仍旧一一对应。让线程竞争更多的表现在OS 调度器上而不是 go 调度器上）

此外，为了防止 go 版本更新而产生的其他优化对数据造成影响，使用 `interface{}` 和泛型的两套代码都在 `go version go1.19 darwin/amd64` 环境下测试。

平台是2019 款 MBP，2.6 GHz 6-Core Intel Core i7 CPU，超线程后共 12 个逻辑核。



#### 1.2 测试结果

1. 环形队列元素变化
{% iframe line-smooth.html 100% 500 %}

### 2. 关于泛型的讨论



### 3. 回到 lfring

