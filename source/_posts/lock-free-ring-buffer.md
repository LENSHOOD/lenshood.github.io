---
title: 一个简单的 Lock Free Ring Buffer，有多简单？
date: 2021-04-19 22:46:04
tags: 
- lock free
- ring buffer
categories:
- Go
---

> 本文涉及到的代码见：https://github.com/LENSHOOD/go-lock-free-ring-buffer

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



## 非同步下不可控的调度产生的问题

由于我们的代码中完全没有用到任何锁或同步块，因此单线程（或正确同步）下的一些代码假设就不再成立：

- 实时性：某两行代码，执行完第一行，下一行会立即执行
- Happens-before：第一行一定在第二行之前执行（只有添加了 memory barrier 的操作才会限制编译器、CPU 的重排序）

所以上一节的代码看似没问题，实际上是无法通过并发测试的，当我们拿多个线程一边写一边读时，很可能会出现如下两种情况：

1. 结果错误：写入的数据和读出的数据对不上，要么读出的少了，要么同一个值读出好几次
2. 死锁：致命的问题，用户在使用 ring buffer 时，通常的做法是当 `Offer`/`Poll` 失败后立即重试（或者让出 CPU 等待下一次调度），由于 lock-free，线程没有 park，死锁会导致 CPU 飙升

本节提到的并发测试代码可以见[这里](https://github.com/LENSHOOD/go-lock-free-ring-buffer/blob/master/mpmc_concurrency_test.go)。

### 场景 1：CAS 与 Read/Write Value 不同步

最简单且可能发生的问题就是 `head`/`tail` 已经被更新，但值迟迟没有写入：

{% asset_img 2.png %}

如上图所示，开始状态 buffer 为空，`Consumer` 因为空而停止消费。这时 `Produer` 开始送入数据，显然，图中`Producer` 的 CAS 操作已经成功，`tail++`已经发布给 `Consumer` 。意味着 `Consumer`已经被授权可以继续消费数据了。

但问题在于 `Producer` 并没有完成值的写入，而是被调度暂停。在这期间 `Consumer` 已经完成了数据的读取，那么显然读到了错误的数据。

要解决这样的问题，就在于需要在 `element[]` 的 slot 中告诉 `Producer`/`Consumer` 当前的值是否可用（已被发布/读取）：

```go
func (r *ring) Offer(v interface{}) bool {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isFull(oldTail, oldHead) {
		return false
	}

	newTail := (oldTail+1) & r.mask
  // ----------- BEGIN --------------
  tailNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail])))
	if tailNode != nil {
		return false
	}
  // ----------- END --------------
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
  // ----------- BEGIN --------------
  headNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead])))
  if headNode == nil {
		return nil, false
	}
  // ----------- END --------------
	if !atomic.CompareAndSwapUint64(&r.head, oldHead, newHead) {
		return nil, false
	}

  // ----------- BEGIN --------------
  atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead])), nil)
  // ----------- END --------------
	return *(*interface{})(headNode), true
}
```

`Consumer` 每次读取数据，都要判断是否为 `nil`，如果是则说明新值还有没写入；对应的，`Producer` 在写入之前也要判断值是否不为`nil`，如果是就说明还没有被消费完毕。

`nil` 值就像一道闸门，隔开了生产和消费的过程。

### 场景 2：Load Tail/Head 与 CAS 不同步（即 ABA）

ABA 问题算是 lock-free 中的经典问题了，讲的就是在 CAS 的时候，需要比较旧值和新值，但当值变化了数次之后，恰巧又变为与旧值相同的值时，CAS 是没法判断到底中间有没有变化的。

{% asset_img 3.png %}

上面的图是说：`Producer 0` 顺利的通过了 `isFull` ，`value == nil` 的检查，准备 CAS，但这时被调度出让执行权，之后其他的 `Producer` 和 `Consumer` 欢快的执行了一整圈，然后 `Producer 0` 又拿到 CPU，开始执行 CAS。

按道理，过了这么久，CAS 必然会执行失败。但很悬的就是，`tail` 转了一圈，现在又恰巧和 `Produer 0` 先前读到的 `oldTail` 一样了，所以 CAS 顺利的执行完成，`Producer 0` 也愉快的写入了他本该在很早以前就写入的值。

但万万没想到的是，`Producer 0` 写入的数据，覆盖掉了原本存在，但没有被读取到的数据，因此 `Consumer 0` 在不知情的情况下读到了重复的数据。这就是典型的 ABA 问题。

怎么解决呢？

通常规避 ABA 问题就是引入版本号或者戳（stamp），这样在实现上让绕了一圈回到旧值的情况不可能或极难发生，就避免了 ABA：

```go
func (r *ring) Offer(value interface{}) (success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isFull(oldTail, oldHead) {
		return false
	}

	newTail := oldTail + 1
	tailNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail & r.mask])))
	// not published yet
	if tailNode != nil {
		return false
	}
	if !atomic.CompareAndSwapUint64(&r.tail, oldTail, newTail) {
		return false
	}

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail & r.mask])), unsafe.Pointer(&value))
	return true
}

func (r *ring) Poll() (value interface{}, success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isEmpty(oldTail, oldHead) {
		return nil, false
	}

	newHead := oldHead + 1
	headNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead & r.mask])))
	// not published yet
	if headNode == nil {
		return nil, false
	}
	if !atomic.CompareAndSwapUint64(&r.head, oldHead, newHead) {
		return nil, true
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead & r.mask])), nil)

	return *(*interface{})(headNode), true
}
```

和前文对比看，其实改动很小，原本计算 `tail/head` 是通过对其加一之后按位与 `mask` 来限制其值不会超限。现如今我们不限制`head/tail` 让它们可以无限制的加一，直到达到 `uint64` 的最大值后归零。

但在读写`element[]` 的时候对索引值进行按位与 `mask`，我们可以知道不论 `head/tail` 有多大，按位与 `mask` 之后的索引值一定不会超限导致 `out of range`。看起来似乎逻辑上和先前没差，但实际上由于 `head/tail` 的极限值被极大的增大了，现在再想要发生 ABA 问题，线程调度间隔之间要执行 `2^64` 次读/写，这几乎不可能。

### 场景 3：Load Tail 与 Load Head 不同步 

从原理上讲，`tail < head` 这种情况绝对不可能发生（除了边界点 `head=capacity-1`，`tail=0`），但不同线程的视角看到的结果很可能不同：



正如上图所示，`Consumer 0` 在读取了 `tail=5` 后，本应继续读取 `head`，但由于被调度器出让 cpu，在一段时间内其他几个线程已经执行了数次操作，等到 `Consumer 0` 再次获得 cpu 后，它读到 `head==7`，因此在 `Consumer 0`的视角看，此时的 ring buffer 出现了”不可能“ 情况：`tail==5, head==7`。

重要的是：目前实际上 `tail==head`，表示 buffer 已经全部被读取，不包含有效数据了。但显然在 `Consumer 0` 看来 `tail - head != 0`，所以它不认为 buffer 已经为空，这时它继续工作：

```go
newHead := (oldHead+1) & r.mask
if !atomic.CompareAndSwapUint64(&r.head, oldHead, newHead) {
  return nil, false
}

headNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead])))
```

一切都顺利的执行完成，之后会导致两个结果：

1. 由于 `Poll()` 在读取结束后并没有清空的动作，因此 `Consumer 0` 会读取到已经被读过一次的数据，产生重复消费。
2. 由于 `CAS(head++)` 正常执行完毕，现在 `head==0`，而 `tail==7`，于是很诡异，`tail - head == capacity-1`， 代表 buffer 已满，所有的`Prouder`都停止工作，然后某个 `Consumer` 再错误的读取一次数据（使`head==1`）然后`Producer` 才能继续工作。





## 性能测试

### 基准

### 对比

### 对比另一种实现