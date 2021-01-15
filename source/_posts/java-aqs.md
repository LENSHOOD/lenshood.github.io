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

- AQS 委托 client 对一个 ”同步状态 state” 进行控制，以这种方式来决定当前访问的线程是否需要进入一个线程队列阻塞等待。

因此，我们能设想，AQS 的作用，对 client 来说是类似一个 “同步器 helper” 的定位，它隐含了一些实现细节，并提供控制端点来帮助 client 更简单的实现同步器功能。

就如同前文的例子，闭锁代码通过定义 `tryAcquireShared(int arg)`，来使所有访问的线程都阻塞（初始 state == 0），只有当 `tryRelease(int arg)` 被调用，state 被设置为 1 后，队列中的线程被一一唤醒，且再次尝试  `tryAcquireShared(int arg)`，并能成功返回大于 0 的结果，因此线程得以继续执行。

同样的，假如我们想要实现一个独占锁，那么只要确保只有一个线程能够成功的将 state 置位（通过 AQS 提供的 CAS 方法），而其他线程置位失败后就会进入等待，直到锁的持有现成通过`release()` 将 state 重新清零为止。

所以，从代码结构上，我们能够将 AQS 的实现分为三层：

{% asset_img aqs_arch.png %}