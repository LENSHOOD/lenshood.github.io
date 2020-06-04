---
title: 虚拟工厂：Java AQS 同步器
date: 2020-05-24 22:50:50
tags:
- java
- aqs
categories:
- Java
---

但凡提到 JUC，就一定会提到 AQS，我们能找到各种各样的文章，来分析 AQS 的实现原理，使用方法等。其原因，不仅是因为通过 AQS，JDK 衍生出了各种各样的同步工具，也因为 AQS 的优秀设计，能够使用户以非常简单的代码就能实现安全高效的同步，同时还能兼顾扩展性。

本文通过分析 AQS 的实现，来展现其优秀的设计架构与代码模型。

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