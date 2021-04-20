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
- 内存结构紧凑（避免 GC），读写效率高（常数时间的入队、出队、随机访问）
- 难以扩容

{% asset_img 1.png %}

### Ideology

原理上 Ring Buffer 简单优雅：

```go
type ring struct {
	head int,
  tail int,
	element []interface{},
  capacity int
}

func (r *ring) Offer(value interface{}) {
  // full
  if tail - head == capacity-1 || tail - head == -1 {
    return
  }
  
  tail++
  
  // turn around
  if tail == capacity-1 {
  	tail = 0;
	}
  
  element[tail] = value
}

func (r *ring) Poll() interface{} {
  // enpty
  if (tail == head) {
    return nil
  }
  
  v := element[head]
  head++
  
  // turn around
  if head == capacity-1 {
  	head = 0;
  }
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

### 前期准备

我们知道，在传统 SMP 架构下，确保多线程程序 memory ordering 的方式是采用 memory barrier +  cache coherence protocol 来实现共享变量的访问在 cpu 不同核之间保持一致性与顺序性。

因此，我们需要根据共享变量（在本文的上下文中指的就是 `tail`/`head`/`element[]`）的访问方式来确定其具体使用编程语言中的哪一种 memory order 抽象。

C++ 和 Java 都提供了较为细粒度的 memory order 抽象（[c++ memory order](https://www.cplusplus.com/reference/atomic/memory_order/)、[java varhandle](http://gee.cs.oswego.edu/dl/html/j9mm.html)），但 go 提供的 `atomic` 包对 memory order 的控制比较粗略，因此我们能用得上的库函数只有如下几种：

```go
// 以下的 “XXX” 代表各种不同的类型如：uint32、int64、uintptr、unsafe.Pointer 等等
func LoadXXX(addr *XXX) (val XXX)
func StoreInt64(addr *XXX, val XXX)
func CompareAndSwapXXX(addr *XXX, old, new XXX) (swapped bool)
```

CAS 需要的 barrier 限制毋庸置疑，但对于 `Load`/`Store`，Go 在[官方文档](https://golang.org/pkg/sync/atomic/)中只是讲这些提到这些函数能够 “atomically” 执行操作，但并没有说细节。不过从[这里](https://groups.google.com/g/golang-dev/c/vVkH_9fl1D8/m/azJa10lkAwAJ)我们看到 rsc 在回复中提到说 Go 的 atomic 实现 ”能够保证 sequential consistency，就像 c++ 的 seqconst“。那么我们就可以推断对共享变量进行 `Load`/`Store`操作就类似于在 java 中对 `volatile` 变量的操作行为。这也意味着最终我们的 lock-free 实现可能会相对较慢。

### 改造一下

#### 原子化的 turn around

在改造为 lock-free 版本之前，我们发现了前述代码中的一个问题：

每当 head / tail 的值达到 capacity 后，我们需要将其重置为 0（即 turn around 操作）。`head/tail++` 与判断并重置为 0 是两个操作，无锁状态下它们无法同步。当然我们也可以先计算出新值，再用 CAS 来更新，但做得越多出错的概率越大，我们期望以更简单的方式来实现 ”turn around“。

我们知道计算机补码的溢出性质：当一个无符号数向上溢出时，它变为 0。我们可以利用这一性质来限定 `tail`/`head`，但直接使用的问题是 capacity 只能是 2^8/16/32/64。

使用 Mask 就可以让 capacity 的选取变得更灵活：执行 `head/tail & (2^n - 1)` 就可以支持任意的二次幂 capacity。这样，我们对 capacity 的限制就只是要求二次幂了。

#### 二次幂转换

我们不期望用户在创建 ring buffer 时需要了解到 ”capacity 必须是二次幂“ 这种细节问题，最简单的处理就是把用户输入的任意大于 0 的 capacity 值向上取值为其最近的二次幂。

如下代码能够即高效又简单的完成这一工作：

```go
func findPowerOfTwo(givenMum uint64) uint64 {
	givenMum--
	givenMum |= givenMum >> 1
	givenMum |= givenMum >> 2
	givenMum |= givenMum >> 4
	givenMum |= givenMum >> 8
	givenMum |= givenMum >> 16
	givenMum |= givenMum >> 32
	givenMum++

	return givenMum
}
```

#### 替换 atomic 函数

经过前面的讨论，我们就可以初步的给出改造结果了：

```go
func (r *ring) Offer(v interface{}) bool {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isFull(oldTail, oldHead) {
		return false
	}

	newTail := (oldTail+1) & r.mask
	if !atomic.CompareAndSwapUint64(&r.tail, oldTail, newTail) {
		return false
	}

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail])), unsafe.Pointer(&v))
	return true
}

func (r *ring) Poll() (v interface{}, success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isEmpty(oldTail, oldHead) {
		return nil, false
	}

	newHead := (oldHead+1) & r.mask
	if !atomic.CompareAndSwapUint64(&r.head, oldHead, newHead) {
		return nil, false
	}

  headNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead])))
	return *(*interface{})(headNode), true
}

func (r *ring) isEmpty(tail uint64, head uint64) bool {
	return (tail - head) & r.mask == 0
}

func (r *ring) isFull(tail uint64, head uint64) bool {
	return (tail - head)  & r.mask == r.capacity - 1
}
```



## 非同步下的各种问题

### 不可控的调度

### ABA



## 性能测试

### 基准

### 对比

### 对比另一种实现