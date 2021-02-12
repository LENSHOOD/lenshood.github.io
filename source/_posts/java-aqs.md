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

Core Logic 层中，CLH 队列变体存放所有排队等待的线程。Try Lock Loop 根据当前排队状态来决定如何处置当前线程（是入队等待还是出队获取资源）。Condition 则是一种等待队列的条件谓词实现。

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

`Node` 中提供了一些方法来对 field 进行操作，他们全部使用 `Unsafe` 提供的方法来实现（Jdk9 版本当中大都采用 `VarHandle` 实现，注释中提到，后续版本又回到 `Unsafe`的原因是 *"avoid potential VM bootstrap issues"* ）。其中有采用 CAS 的方法，也有单纯的 get/set 方法。其中的 `setXXXRelaxed` 方法实际上就是传统的 setter 方法（这里也要用 `Unsafe` 也许是为了与其他几个方法保持一致），Relaxed 后缀，是 JDK9 通过 `VarHandle`引入的 Memory Order 中的概念([Doug Lea 的解释](http://gee.cs.oswego.edu/dl/html/j9mm.html))，实际上应该多少借鉴了 [C++ 11 的 Memory Order 模型](https://www.zhihu.com/question/24301047)。

对于 CLH 节点，在 AQS 中还定义了 `head`，`tail` 等概念，来维护一个完整的链表队列，其入队、出队的操作也都在 `acquire` 与 `release` 方法中实现。

### Try Lock Loop

当一个 client 确认某个访问线程需要排队等待获取资源时，AQS 会将访问线程封装为一个 CLH Node，并进入一个等锁的循环，来根据当前等待队列的情况，采取不同的逻辑（相关逻辑在 `final int acquire(Node node, int arg, boolean shared, boolean interruptible, boolean timed, long time)`方法中），流程转换图如下所示：

{% asset_img try_lock_loop.png %}

上图看似复杂，实际上只包括了如下几个过程：

1. 创建 Node，入队
2. 线程休眠等待
3. 被唤醒（head 线程已经做完所有工作），位列第一，此时是除 head 节点外第一顺位的 Node
4. 尝试获取资源，获取到后将自己设置为 head，并退出等锁循环，继续执行线程逻辑
5. 若获取资源失败（被其他未排队的线程抢占，即非公平抢占）
   - 自旋等待锁释放
   - 若自旋太久重新进入 park（但顺位仍是 first）

对比上述过程我们发现，实际的实现代码，是把 “创建 Node，入队，休眠等待” 这件事，拆成了多个阶段（创建，入队，设置 Waiting 状态），而在逐步进行这些阶段之间，在节点入队前会尽可能尝试 tryAcquire，这一点在类注释中讲到：

`在队列中排名第一并不能保证获取到资源，这只代表获得了竞争的权利。我们平衡了吞吐、开销、公平性之后，允许线程在入队前“抢占”的尝试获取锁。`

另一方面，在入队并休眠前，拆分阶段也使得当前节点对前驱节点取消的响应更加及时。不止如此，`acquire` 方法在实现过程中，考虑了许多优化点来提升性能：

- 非公平性，假如持有锁的线程在释放后又立即 try lock，对于公平锁而言，它只能在队尾排队等待，而非公平锁允许它尝试抢占。这样就避免了入队后等待以及被唤醒的两次线程切换操作。（但非公平锁可能导致线程 "starving"，因此 ReentrantLock 就分别提供了公平、非公平的实现）

- 为了让 GC 更易于回收，在入队前，Node 的 field 都默认为 `null`，因为 “在 Node 在被使用前就已经被丢弃的现象并不少见”
- 对于 CLH 需要的一个 dummy head（哨兵节点），AQS 在创建的时候并不会将其一起创建出来，而是在出现第一次竞争时才创建，以减少无效的开销。（可能 AQS 被创建后很久，都没有遇到过竞争的情况）

最后贴上代码实现：

```java
final int acquire(Node node, int arg, boolean shared,
                      boolean interruptible, boolean timed, long time) {
  Thread current = Thread.currentThread();
  byte spins = 0, postSpins = 0;   // retries upon unpark of first thread
  boolean interrupted = false, first = false;
  Node pred = null;                // predecessor of node when enqueued

  for (;;) {
    if (!first && (pred = (node == null) ? null : node.prev) != null &&
        !(first = (head == pred))) {
      if (pred.status < 0) {
        cleanQueue();           // predecessor cancelled
        continue;
      } else if (pred.prev == null) {
        Thread.onSpinWait();    // ensure serialization
        continue;
      }
    }
    if (first || pred == null) {
      boolean acquired;
      try {
        if (shared)
          acquired = (tryAcquireShared(arg) >= 0);
        else
          acquired = tryAcquire(arg);
      } catch (Throwable ex) {
        cancelAcquire(node, interrupted, false);
        throw ex;
      }
      if (acquired) {
        if (first) {
          node.prev = null;
          head = node;
          pred.next = null;
          node.waiter = null;
          if (shared)
            signalNextIfShared(node);
          if (interrupted)
            current.interrupt();
        }
        return 1;
      }
    }
    if (node == null) {                 // allocate; retry before enqueue
      if (shared)
        node = new SharedNode();
      else
        node = new ExclusiveNode();
    } else if (pred == null) {          // try to enqueue
      node.waiter = current;
      Node t = tail;
      node.setPrevRelaxed(t);         // avoid unnecessary fence
      if (t == null)
        tryInitializeHead();
      else if (!casTail(t, node))
        node.setPrevRelaxed(null);  // back out
      else
        t.next = node;
    } else if (first && spins != 0) {
      --spins;                        // reduce unfairness on rewaits
      Thread.onSpinWait();
    } else if (node.status == 0) {
      node.status = WAITING;          // enable signal and recheck
    } else {
      long nanos;
      spins = postSpins = (byte)((postSpins << 1) | 1);
      if (!timed)
        LockSupport.park(this);
      else if ((nanos = time - System.nanoTime()) > 0L)
        LockSupport.parkNanos(this, nanos);
      else
        break;
      node.clearStatus();
      if ((interrupted |= Thread.interrupted()) && interruptible)
        break;
    }
  }
  return cancelAcquire(node, interrupted, interruptible);
}
```



### Condition

Condition 可以看做是对 `Object.wait()` 与 `Object.notify()` 的对象式封装。它的优点在于，我们可以根据不同的条件来创建不同的 Condition，而这些 Condition 能够共同作用与同一组资源竞争者，从而实现更为灵活的逻辑控制。

AQS 将 Condition 的等待/唤醒调度也融合在了 CLH 队列中。它将与 Condition 相关的线程封装为一个单独的 `ConditionNode` 节点，与之对应的，还有 `ExclusiveNode` 和 `SharedNode`。只不过 `ConditionNode` 还实现了 `ForkJoinPool.ManagedBlocker` 接口：

```java
static final class ConditionNode extends Node
  implements ForkJoinPool.ManagedBlocker {
  ConditionNode nextWaiter;            // link to next waiting node

  /**
   * Allows Conditions to be used in ForkJoinPools without
   * risking fixed pool exhaustion. This is usable only for
   * untimed Condition waits, not timed versions.
   */
  public final boolean isReleasable() {
    return status <= 1 || Thread.currentThread().isInterrupted();
  }

  public final boolean block() {
    while (!isReleasable()) LockSupport.park();
    return true;
  }
}
```

实现 `ForkJoinPool.ManagedBlocker`  的目的是为了在 `Condition.await()` 时交由 `ForkJoinPool` 来协助执行状态检查并控制当前线程进入等待。

AQS 又设计了 `ConditionObject` 类，作为真正的条件对象。`Condition` 的通常使用场景是，由于不满足某个条件，某个线程被挂起，并由另外的线程在条件满足时将其唤醒。由于涉及到多个线程之间对于同一条件（也是一种资源）的操作，这显然是一个需要用到锁的场景，因此 AQS 在其内部实现了 `ConditionObject` ，能直接与条件判断逻辑中的锁关联在一起。

所以，当应用程序期望使用 `Condition` 来调度线程时，需要的动作如下：

1. 创建锁对象： `new Lock()`
2. 创建一个或多个条件对象：`Lock.newCondition()`
3. 判断条件前先获取锁，`Lock.lock()`
4. 不满足条件，进入等待：`Condition.await()`，此时先前获取到的锁被自动释放
5. 另一线程的动作导致条件被满足，重新唤醒：`Condition.singal()`，实际当中更多的会用`Condition.signalAll()` 防止[伪唤醒](https://lenshood.github.io/2020/04/04/some-jaava-tips/#%E4%BC%AA%E5%94%A4%E9%86%92-spurious-wakeup)
6. 等待的线程被唤醒，在执行下一步动作之前，还需要再次获取锁，因为这部分逻辑是被锁包裹的
7. 获取锁成功，继续执行

基于上面的步骤，我们来看看 `ConditionObject` 真正的实现：

```java
public final void await() throws InterruptedException {
  ...
  ConditionNode node = new ConditionNode();
  int savedState = enableWait(node);
  ...
  while (!canReacquire(node)) {
    ...
        ForkJoinPool.managedBlock(node);
    ...
  }
  ...
  acquire(node, savedState, false, false, false, 0L);
  ...
}

private int enableWait(ConditionNode node) {
  if (isHeldExclusively()) {
    node.waiter = Thread.currentThread();
    node.setStatusRelaxed(COND | WAITING);
    ConditionNode last = lastWaiter;
    if (last == null)
      firstWaiter = node;
    else
      last.nextWaiter = node;
    lastWaiter = node;
    int savedState = getState();
    if (release(savedState))
      return savedState;
  }
  node.status = CANCELLED; // lock not held or inconsistent
  throw new IllegalMonitorStateException();
}

private boolean canReacquire(ConditionNode node) {
  // check links, not status to avoid enqueue race
  return node != null && node.prev != null && isEnqueued(node);
}
```

以上是 `await()` 相关的实现。我们可以看到，在创建了 `ConditionNode` 之后，会先通过 `enableWait()` 检查当前是否持有锁，并对 node 进行初始化。注意，这里我们发现，在 `ConditionObject` 里面，还维护了一个单独的 `ConditionNode` 队列，专门用于管理由于等待条件而挂起的线程。最后，在节点入队后，将当前的锁释放。

`ForkJoinPool.managedBlock(node);` 这句话就是用 `ForkJoinPool` 来帮助维护挂起了，其执行逻辑，类似：

```java
while (!blocker.isReleasable())
  if (blocker.block())
    break;
```

可以看到，当前线程被重新唤醒后，仍然要进入 `acquire(node, savedState, false, false, false, 0L);`的流程，这就是重新获取锁的过程（所以如果这时有其他线程占用着锁，当前被唤醒的线程又会重新被挂起，这在 `signalAll` 时会出现）。

```java
public final void signal() {
  ConditionNode first = firstWaiter;
  if (!isHeldExclusively())
    throw new IllegalMonitorStateException();
  if (first != null)
    doSignal(first, false);
}

private void doSignal(ConditionNode first, boolean all) {
  while (first != null) {
    ConditionNode next = first.nextWaiter;
    if ((firstWaiter = next) == null)
      lastWaiter = null;
    if ((first.getAndUnsetStatus(COND) & COND) != 0) {
      enqueue(first);
      if (!all)
        break;
    }
    first = next;
  }
}

final void enqueue(Node node) {
  if (node != null) {
    for (;;) {
      Node t = tail;
      node.setPrevRelaxed(t);        // avoid unnecessary fence
      if (t == null)                 // initialize
        tryInitializeHead();
      else if (casTail(t, node)) {
        t.next = node;
        if (t.status < 0)          // wake up to clean link
          LockSupport.unpark(node.waiter);
        break;
      }
    }
  }
}
```

以上是 `signal()` 相关的逻辑，在条件满足被 `signal()` 后，会选择先从 `firstWaiter` 开始唤醒，唤醒前将 `ConditionNode` 插入CLH等锁队列中。假如是 `signalAll()`则会在唤醒 `firstWatier` 之后继续唤醒下一个 `ConditionNode`。

## Unsafe 支撑

作为 AQS 中对 CLH 队列的操作（包括 lock-free 的入队以及对线程的控制等）的支撑，`jdk.internal.misc.Unsafe` 类承担了绝大多数的职责。

AQS 通过如下语句来获取 `Unsafe`：

```java
private static final Unsafe U = Unsafe.getUnsafe();
```

### CAS

CAS 即 compare and set 或 compare and swap，在 lock-free 编程中有着广泛的应用。

多数 CPU 都提供了具有 CAS 语义的指令，将 compare and set 这样的动作在一条指令中原子的执行，`Unsafe` 中包装了一些 CAS 方法：

- `compareAndSetXXX(Object o, long offset, Object expected, Object x)`：在对象 o 的 offset 处判断当前值是否为 expected，如果是则将其设置为 x，并返回 true，否则返回 false。其中 expected 与 x 根据具体不同的方法，也可以是 primitive 类型
- `compareAndExchangeXXX(Object o, long offset, Object expected, Object x)`：与 `compareAndSet` 类似的语义。
- `weakCompareAndSetXXX(Object o, long offset, Object expected, Object x)`：与 `compareAndSet` 类似的语义，但提供了更弱的内存语义，因此在即使实际值与 expected 一致时，也可能会由于内存竞争而失败。

因此，CLH 队列在入队时，由于可能同时有很多个线程尝试入队，因此采用了 CAS 的方法来设置队尾：

```java
} else if (pred == null) {          // try to enqueue
  node.waiter = current;
  Node t = tail;
  node.setPrevRelaxed(t);         // avoid unnecessary fence
  if (t == null)
    tryInitializeHead();
  else if (!casTail(t, node))
    node.setPrevRelaxed(null);  // back out
  else
    t.next = node;
}
```

而由于出队的时候，只会有一个线程参与操作，就不需要 CAS 了：

```java
if (acquired) {
  if (first) {
    node.prev = null;
    head = node;
    pred.next = null;
    node.waiter = null;
    if (shared)
      signalNextIfShared(node);
    if (interrupted)
      current.interrupt();
  }
  return 1;
}
```

### Thread 调度

`Unsafe` 也提供了对线程的调度操作：

```java
// block current thread
public native void park(boolean isAbsolute, long time);

// unblock the given thread
public native void unpark(Object thread);
```

可以看到，上面的方法可以实现对线程进行 block 或 unblock。这里要回顾一下线程的状态：

- NEW：Thread 还未启动
- RUNNABLE：从 JVM 的角度看，Thread 正在执行中。但在操作系统层面可能处于等待资源的状态
- BLOCKED：正在等待 monitor lock 的 Thread。可代表正在等待 `synchronized` 块的 Thread 状态。
- WAITING：等待其他线程执行动作。如下操作后，Thread 可以进入 WAITING 状态：
  - `Object.wait()`
  - `Thread.join()`
  - `LockSupport.park()`：LockSupport 在 `park()` 中调用了 `Unsafe.park()`
- TIMED_WAITING：与 WAITING 类似，只不过调用的方法都带有 `wait time`参数
- TERMINATED：Thread 已经终止。

因此，在`Unsafe.park` 之后，线程就进入了 WAITING 状态。所以在 AQS `acquire` 方法的最后，就是将线程 park。

AQS 中实际使用的 `LockSupport.park()` 与 `Unsfae.park()` 的主要区别在于，`LockSupport.park` 提供了包装逻辑来在等待线程中设置被等待的对象：`blocker` 。`blocker` 可以用于调试、监控等目的。	  

## 多样的同步器示例

### ReentrantLock

### ReadWriteLock

### CountDownLatch

### CyclicBarrier

### Semaphore