---
title: 并发原子性与可见性
date: 2019-08-08 22:37:15
tags:
- concurrency
- atomicity
- visibility
categories:
- Java
---
并发编程是 Java 编程的基础，同时也是提升效率，改善性能表现的利器。说到并发，就一定会说到同步，`synchronized`关键字是 Java 中关于同步最基础的设计，它能够简单明确的提供对变量、代码块、方法、类的同步支持。我们大都知道，同步能够为代码提供原子性，但有时我们会忽略，同步还有一个重要的作用，就是提供了代码的可见性。

下文将从两个方面，以 Java 为例简述同步带来的原子性和可见性，并在可见性部分引出了经常令 Java 程序员困惑的 `volatile`关键字。

### 同步 - 原子性
给出以下类：
``` java
public class SynchronizedAtomicity {

    private int i = 0;

    public void NonThreadSafeCounter() {
        counter();
    }

    public synchronized void ThreadSafeCounter() {
        counter();
    }

    private void counter() {
        i++;
        System.out.println("Thread" + Thread.currentThread().getId() + " say: i is: " + i);
    }
}
```
非线程安全：
``` java
@Test
public void validate_non_thread_safe_counter() {
	SynchronizedAtomicity synchronizedAtomicity = new SynchronizedAtomicity();

	Runnable task = () -> {
        int stopNum = 100;
        while (stopNum-- > 0) {
        	synchronizedAtomicity.NonThreadSafeCounter();
            try {
            	Thread.sleep(0, 10);
            } catch (InterruptedException e) {
            	e.printStackTrace();
            }
        }
    };

	Thread t1 = new Thread(task);
	Thread t2 = new Thread(task);

    t1.start();
    t2.start();

    t1.join();
    t2.join();
}
```
得到结果：
```
······
Thread10 say: i is: 184
Thread11 say: i is: 185
Thread10 say: i is: 186
Thread11 say: i is: 187
Thread10 say: i is: 187
Thread11 say: i is: 189
Thread10 say: i is: 189
```
线程安全：
``` java
@Test
public void validate_thread_safe_counter() throws InterruptedException {
    SynchronizedAtomicity synchronizedAtomicity = new SynchronizedAtomicity();

    Runnable task = () -> {
        int stopNum = 100;
        while (stopNum-- > 0) {
            synchronizedAtomicity.ThreadSafeCounter();
            try {
                Thread.sleep(0, 10);
            } catch (InterruptedException e) {
                e.printStackTrace();
            }
        }
    };

    Thread t1 = new Thread(task);
    Thread t2 = new Thread(task);

    t1.start();
    t2.start();

    t1.join();
    t2.join();
}
```
得到结果：
```
······
Thread11 say: i is: 194
Thread11 say: i is: 195
Thread10 say: i is: 196
Thread11 say: i is: 197
Thread10 say: i is: 198
Thread11 say: i is: 199
Thread10 say: i is: 200
```

显然没有`synchronize`保护的非线程安全方法在多线程执行时不具备原子性，导致在两个循环 100 次累加的线程中执行后，总值小于 200，而具有原子性的方法总值等于 200。

以上即以最简单的例子说明了同步的原子性。

### 同步 - 可见性