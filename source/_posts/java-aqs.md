---
title: 虚拟工厂：Java AQS 同步器
date: 2021-01-14 23:07:50
tags:
- java
- aqs
categories:
- Java
---

但凡提到 JUC，就一定会提到 AQS，我们能找到各种各样的文章，来分析 AQS 的实现原理，使用方法等。其原因，不仅是因为通过 AQS，JDK 衍生出了各种各样的同步工具，也因为 AQS 的优秀设计，能够使用户以非常简单的代码就能实现安全高效的同步，同时还能兼顾扩展性。

本文通过分析 AQS 的实现，来展现其优秀的设计架构与代码模型。

<!-- more -->

开始之前，先放出一个小例子，来看看使用 AQS 实现同步工具是多么的简单（本例参考了《Java 并发编程实战》中的例子）：

```java
public class Latch {
    private final Sync sync = new Sync();

    public void await() { sync.acquireShared(0); }

    public void release() { sync.release(0); }

    class Sync extends AbstractQueuedSynchronizer {
        @Override
        protected int tryAcquireShared(int arg) { return getState() == 1 ? 1 : -1; }

        @Override
        protected boolean tryRelease(int arg) {
            setState(1);
            return true;
        }
    }
}

@Test
public void should_release_after_10_seconds() throws InterruptedException {
    Latch latch = new Latch();

    Runnable waiter = () -> {
        latch.await();
        System.out.println(Thread.currentThread().getName() + " done");
    };
    Thread thread1 = new Thread(waiter);
    Thread thread2 = new Thread(waiter);

    System.out.println("Start at: " + System.currentTimeMillis());

    thread1.start();
    thread2.start();
    Thread.sleep(10000);

    latch.release();
    thread1.join();
    thread2.join();

    System.out.println("End at: " + System.currentTimeMillis());
}
```

上述例子描述了一个最简单的同步工具：闭锁。多个线程可以`await()`在其上，一旦闭锁`release()`时，所有线程得以释放。

上述例子的测试结果如下：

```shell
Start at: 1590683053181
Thread-4 done
Thread-3 done
End at:   1590683054190
```

通过 AQS，只要不到 20 行代码，就能实现闭锁功能，可见其极大的简化了工作。

## 总体结构

> 下文中源码部分使用的是 openjdk-15 的版本，与 jdk-8 的实现略有不同，但原理一致

从使用角度讲，AQS 的原理可以总结为一句话：

- AQS 委托 client 对一个 ”同步状态 state” 进行控制，以此来决定当前访问的线程是否需要进入一个线程队列阻塞等待。

因此，我们能设想，AQS 的作用，对 client 来说是类似一个 “同步器 helper” 的定位，它隐含了一些实现细节，并提供控制端点来帮助 client 更简单的实现同步器功能。

就如同前文的例子，闭锁代码通过定义 `tryAcquireShared(int arg)`，来使所有访问的线程都阻塞（初始 state == 0），只有当 `tryRelease(int arg)` 被调用，state 被设置为 1 后，队列中的线程被一一唤醒，且再次尝试  `tryAcquireShared(int arg)`，并能成功返回大于 0 的结果，因此线程得以继续执行。

同样的，假如我们想要实现一个独占锁，那么只要确保只有一个线程能够成功的将 state 置位（通过 AQS 提供的 CAS 方法），而其他线程置位失败后就会进入等待，直到锁的持有现成通过`release()` 将 state 重新清零为止。

所以，从代码结构上，我们能够将 AQS 的实现分为三层：

{% asset_img aqs_arch.png %}

在 API 层中，`acquireXXX` 与 `releaseXXX` 主要由当前访问的线程来触发，带`Shared` 后缀的方法都是共享访问方法，不带的是独占访问方法。`tryAcquire`与`tryRelease`由同步器的子类定义，通过对 `state` 进行操作和对比，来达到判断是否能获取/释放的目的。`state` 本身只是一个共享的 int 变量，用于帮助 API 层 `tryXXX` 方法记录、判断资源是否可获取。

Core Logic 层中，CLH 队列变体存放所有排队等待的线程。Try Lock State Machine 根据当前排队状态来决定如何处置当前线程（是入队等待还是出队获取资源）。Condition 则是一种等待队列的条件谓词实现。

Support 层基本由对 Unsafe 包提供的方法进行封装（或直接使用）来实现 CAS 和线程调度等支撑性功能。

## Core Logic 层实现

### CLH Queue Variant

在 AQS 中，实现了一个 CLH 的变体用作等待队列。CLH 队列最早是由 Craig，Landin 和 Hagersten，分别在两篇独立的论文中提出的一个相似的观点，即通过排队自旋的方式来公平的取用资源，从而避免竞争所产生的的资源消耗。

AQS 中的等待队列，是类似 CLH 锁队列的一个变体，相比单纯的自旋，AQS 中更多的采用了对线程进行阻塞的方式来等待资源。

CLH 等待队列的节点实现如下所示：

```java
// Node status bits, also used as argument and return values
static final int WAITING   = 1;          // must be 1
static final int CANCELLED = 0x80000000; // must be negative
static final int COND      = 2;          // in a condition wait

abstract static class Node {
    volatile Node prev;       // initially attached via casTail
    volatile Node next;       // visibly nonnull when signallable
    Thread waiter;            // visibly nonnull when enqueued
    volatile int status;      // written by owner, atomic bit ops by others

    // methods for atomic operations
    final boolean casPrev(Node c, Node v) {  // for cleanQueue
      	return U.weakCompareAndSetReference(this, PREV, c, v);
    }
    final boolean casNext(Node c, Node v) {  // for cleanQueue
      	return U.weakCompareAndSetReference(this, NEXT, c, v);
    }
    final int getAndUnsetStatus(int v) {     // for signalling
      	return U.getAndBitwiseAndInt(this, STATUS, ~v);
    }
    final void setPrevRelaxed(Node p) {      // for off-queue assignment
      	U.putReference(this, PREV, p);
    }
    final void setStatusRelaxed(int s) {     // for off-queue assignment
      	U.putInt(this, STATUS, s);
    }
    final void clearStatus() {               // for reducing unneeded signals
      	U.putIntOpaque(this, STATUS, 0);
    }

    private static final long STATUS
      	= U.objectFieldOffset(Node.class, "status");
    private static final long NEXT
      	= U.objectFieldOffset(Node.class, "next");
    private static final long PREV
      	= U.objectFieldOffset(Node.class, "prev");
}
```

显然，从数据结构的角度讲，等待队列实际上是一个双向链表。

定义了前驱、后继节点 `prev` 和 `next`（由于前驱后继节点通常都是由不同的线程来创建和访问，因此采用 `volatile` 语法确保不同线程访问的可见性），当前节点的实际内容有两个：a. 实际等待线程的引用。b. 当前节点的状态，状态定义为 `WATING` ，`CANCLELLED`，`COND`。

`Node` 中提供了一些方法来对 field 进行操作，他们全部使用 `Unsafe` 提供的方法来实现（Jdk9 版本当中大都采用 `VarHandle` 实现，目前我还不清楚为什么在后续版本中又回到了 `Unsafe`）。其中有采用 CAS 的方法，也有单纯的 get/set 方法。其中的 `setXXXRelaxed` 方法实际上就是传统的 setter 方法（这里也要用 `Unsafe` 也许是为了与其他几个方法保持一致），Relaxed 后缀，是 JDK9 通过 `VarHandle`引入的 Memory Order 中的概念([Doug Lea 的解释](http://gee.cs.oswego.edu/dl/html/j9mm.html))，实际上应该多少借鉴了 [C++ 11 的 Memory Order 模型](https://www.zhihu.com/question/24301047)。

对于 CLH 节点，在 AQS 中还定义了 `head`，`tail` 等概念，来维护一个完整的链表队列，其入队、出队的操作也都在 `acquire` 与 `release` 方法中实现。

### Try Lock State Machine

当一个 client 确认某个访问线程需要排队等待获取资源时，AQS 会将访问线程封装为一个 CLH Node，并进入一个类似 State Machine 的循环，来根据当前等待队列的情况，采取不同的逻辑，状态转换图如下所示：



### Condition

## Unsafe 支撑

### CAS

### Thread 调度

## 多样的同步器示例

### ReentrantLock

### ReadWriteLock

### CountDownLatch

### CyclicBarrier

### Semaphore