---
title: 一个简单的 Lock Free Ring Buffer，有多简单？
date: 2021-04-19 22:46:04
tags: 
- lock free
- ring buffer
categories:
- Go
---

## Ring Buffer

Ring Buffer 是一种极其简单的数据结构，它具有如下常见的特性：

- 容量固定的有界队列，进出队列操作不需要整体移动队内数据
- 内存结构紧凑（避免 GC），读写效率高（常数时间的入队、出队、索引）
- 难以扩容

{% asset_img 1.png %}

### Ideology

原理上 Ring Buffer 是极其简单优雅的：

```go
type ring struct {
	head int,
  tail int,
	element []interface{},
  capacity int
}

func (r *ring) Offer(value interface{}) {
  if (tail - head) > capacity-1 {
    return
  }
  
  element[tail & (capacity-1)] = value
  tail++
}

func (r *ring) Poll() interface{} {
  if (tail == head) {
    return nil
  }
  
  v := element[head & (capacity-1)]
  head++
  return v
}
```

### Reality

既然作为一种缓冲器，我们可以预见到 Ring Buffer 的一个主要场景就是在 Producer - Consumer 模式下均衡数据交换速率，削峰填谷。

因此天然的，我们期望一个 Ring Buffer 是 thread safe 的。

我们有很多种方式来避免入队和出队操作相互竞争，其中最简单的就是把共享变量：`head`、`tail`、`element[]` 用互斥锁保护起来。

但我们会发现，在前面代码中几乎每一条语句都涉及到共享变量的操作，所以我们只能用锁将整个函数体都包裹起来，导致临界区([Critical Section](https://en.wikipedia.org/wiki/Critical_section)) 很大。

**考虑到 Ring Buffer 超级简单，我们可以用 Lock Free 的方式来改造它吗？**

1. 虽然 `element[]` 整体上是一个共享变量，但由于只有`head` 和 `tail` 的持有者才能访问数据，所以不同的持有者访问`element[]` 不同的 "slot"，并不会发生竞争
2. `tail` 只被 Producer 更新，`head` 只被 Consumer 更新，`head` / `tail`之间不存在原子更新关系。
3. `tail` 和 `head` 都是单个整数类型变量，对其读写适用于CPU支持的原子操作（read-modify-write）

这么看来直接使用一些 atomic 操作，就能实现 lock-free 的 ring buffer 了！



## Lock-Free Ring Buffer

