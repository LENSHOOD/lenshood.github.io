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
先来看一段代码：
```
public class SynchronizedVisibility {
    private static boolean ready;
    private static int num;

    public static class ReaderThread extends Thread {
        @Override
        public void run() {
            while (!ready) {
                Thread.yield();
            }
            System.out.println(num);
        }
    }

    public static void main(String[] args) {
        new ReaderThread().start();
        num = 42;
        ready = true;
    }
}
```
上述代码摘录自`《Java 并发编程实战》程序清单 3-1`，显然，对于两个共享变量`ready`， `num`的可见性，在上述代码中存在未知。

按照书中所述，上述代码存在三种情况：
1. 程序正常结束，打印 42
2. 程序正常结束，打印 0 (num 与 ready 的赋值顺序被重排，导致循环在 num 赋值前结束)
3. 程序陷入循环无法结束(在循环内 ready未被修改，对 ready 的判断可能被提至循环体外部)
上述情况，除了第一种符合我们的预期，其他的两种情况，在单线程模型下是无法想象的。然而在多线程情况下，确有一定几率会出现奇怪的程序行为。

> 需要说明的是，我在本地环境中运行上述代码 5w + 次，并未出现一例错误，经 Google 后得知，对指令的重排与 CPU 架构，JIT 等等都有关，虽然无法复现，但从 JVM 的设计角度讲上述情况是可能发生的。

#### Reordering
上述因为可见性导致的问题，都可归于 Java 的重排序问题。

重排序是由[ Java 内存模型](https://docs.oracle.com/javase/specs/jls/se8/html/jls-17.html)的设计而产生的一种自动对程序代码执行顺序的重排优化。无论是 JIT、Javac 还是处理器硬件，都可能会因优化考虑，而对代码指令进行重排序。

此外，Java 内存模型（JMM）中提到了`intra-thread semantics`的概念，即在单线程程序内，重排序在不影响最终执行结果的前提下进行，换句话说，假如后一条语句依赖前一条语句所修改的变量值结果，则为了保证最终结果一致性，这种语句关系不会被重排。然而这并不适用于多线程情况下。

因此前述的情况 2 和情况 3 就显而易见。在情况 2 中，对 main thread 而言，`num = 42；` 与 `ready = true` 没有任何依赖关系，因此对这两条语句的重排是合法的，至于要不要重排则两说。情况 3 也一样，对 `!ready == true` 的判断，`ready`变量的值直到循环结束都没有被改变，那么将判断提前，类似于:
``` Java
if (!ready) {
    while(true) {
        .....
    }
}
```
是完全合法的。因此在多线程情况下就可能出现由于重排序而导致的错误。

请注意，上述示例程序并没有进行任何的同步处理。在 JMM 中讲到：

```
The semantics of the Java programming language allow compilers and microprocessors to perform optimizations that can interact with incorrectly synchronized code in ways that can produce behaviors that seem paradoxical. Here are some examples of how incorrectly synchronized programs may exhibit surprising behaviors.
```

#### synchronized 关键字
由于优化的原因，非正确同步的代码会产生令人惊讶的行为。那么我们只要保证程序被正确的同步，则就不会出现上述异常的情况。

因此最简单的，将上述程序修改：
``` java
public class SynchronizedVisibility {
    private static boolean ready;
    private static int num;

    public static class ReaderThread extends Thread {
        @Override
        public void run() {
            while (!ready) {
                Thread.yield();
            }
            System.out.println(num);
        }
    }

    public static void main(String[] args) {
        new ReaderThread().start();
        synchronized (SynchronizedVisibility.class) {
            num = 42;
            ready = true;
        }
    }
}
```

由`synchronized` 关键字保证的同步性，使得无论在同步块内的语句被如何重排，只要主线程当前执行至同步块内，Reader 线程则无法在类锁释放前访问其静态成员，因此保证了 ready 和 num 对 Reader 线程的可见性。

#### volatile
除了基本的同步以外，Java 语言提供了`volatile`关键字，来提供一种“弱同步”，即对线程共享变量提供可见性(仅提供可见性而不包括原子性)。

Java SE8 官方文档中提到：
```
The Java programming language allows threads to access shared variables. As a rule, to ensure that shared variables are consistently and reliably updated, a thread should ensure that it has exclusive use of such variables by obtaining a lock that, conventionally, enforces mutual exclusion for those shared variables.

The Java programming language provides a second mechanism, volatile fields, that is more convenient than locking for some purposes.

A field may be declared volatile, in which case the Java Memory Model ensures that all threads see a consistent value for the variable.
```
为了实现在保证并发的同时提供可见，JVM 简单的将操作带有`volatile`关键字的共享变量的语句禁止重排，因此也就杜绝了由于重排导致的可见性问题。

`volatile`与`synchronized`不同，在访问`volatile`的变量时并不会加锁，因此就提升了并发性，提升了性能。然而由于`volatile`的弱同步性，在涉及复杂逻辑处使用之应慎重，以防止出现更脆弱也更难以理解的代码。

引用`《Java 并发编程实战》 P31`的一段话：
```
仅当 volatile 变量能简化代码的实现以及对同步策略的验证时，才应该使用它们。如果在验证正确性时需要对可见性进行复杂的判断，那么就不要使用 volatile 变量。volatile 变量的正确使用方法包括：确保他们自身状态的可见性，确保他们所引用对象的状态的可见性，以及标识一些重要的程序生命周期事件的发生(例如，初始化或关闭)。
```

### 参考
1. [Java Memory Model](https://docs.oracle.com/javase/specs/jls/se8/html/jls-17.html#jls-17.4)
2. [volatile Fields](https://docs.oracle.com/javase/specs/jls/se8/html/jls-8.html#jls-8.3.1.4)