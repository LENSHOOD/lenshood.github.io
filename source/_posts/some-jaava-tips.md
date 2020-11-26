---
title: 一些 Java 语言 Tips
date: 2020-04-04 20:56:16
tags: 
- java
- tips
categories:
- Java
---

# 一些 Java 语言 Tips
本文包括一些 Java 语言在使用中常见的小错误以及不佳实践，他们收集自我的日常开发、Code Review、以及书中所见。持续更新...

<!-- more -->

## 双括号初始化（DBI）

在 Java 9 之前，想要在 Field 中初始化一个集合供其他方法使用，他的样子是尴尬而丑陋的：
```java
private static List<String> initListBeforeJava9 = new ArrayList<>();
static {
    initListBeforeJava9.add("Dopey");
    initListBeforeJava9.add("Doc");
    initListBeforeJava9.add("Bashful");
    initListBeforeJava9.add("Happy");
    initListBeforeJava9.add("Grumpy");
    initListBeforeJava9.add("Sleepy");
    initListBeforeJava9.add("Sneezy");
}
```
因此，或多或少的，我们会看到那个时候的代码有一种“更优雅”的实现法：
```java
private static List<String> doubleBraceCollectionInit =
            new ArrayList<String>(){{
                add("Dopey");
                add("Doc");
                add("Bashful");
                add("Happy");
                add("Grumpy");
                add("Sleepy");
                add("Sneezy");
            }};
```
这种方式有一个专门的名字：DBI（Double Brace Initialization）。然而，他只是看起来美一些，其实并不好。

首先，这种初始化方式的本质是：
- 第一层大括号，创建了一个 `ArrayList<String>`的匿名子类
- 第二层大括号，在匿名子类中构建一个代码块（也叫构造块），在代码块中调用父类的 `add` 方法来进行初始化

我们知道，Java 匿名类是可以直接访问外层类成员的：
```java
class Outer {
    private String outerField = "outer_flied";
    private Inner inner = new Inner() {
        void printOuter() {
            System.out.println(outerField);
        }
    };
}
```
之所以能够访问外层成员，是由于 Java 内部匿名类保存了对外层类对象的引用。那么，就存在一个问题：

假如采用 DBI 对某集合成员进行构造，在之后的某些逻辑中，将该集合成员**发布**了出去。由于存在引用关系，那么外层类对象，就再也无法被回收，直到这个被发布出去的集合对象被回收为止。

所以说，**DBI 会存在内存泄漏的风险**。

不过，话说回来，第一种初始化方式，确实有点不够美观，幸好我们有了 Java 9：
```java
private static List<String> java9CollectionInit =
            List.of("Dopey", "Doc", "Bashful", "Happy", "Grumpy", "Sleepy", "Sneezy");
```

## 伪唤醒 （Spurious Wakeup）
首先看一段代码：
```java
public class SpuriousWakeup {
    private static final Random RANDOM = new Random();
    private boolean condition = RANDOM.nextBoolean();
    
    void notifier() {
        synchronized (this) {
            ... ...
            notify();
        }
    }

    void wrongWaiter() throws InterruptedException {
        synchronized (this) {
            if (condition) {
                wait();
            }
            ... ...
        }
    }
}
```
上述代码中，`wrongWaiter()`根据 `condition` 来判断是否进行 `wait()`，而对应的，在 `notifier()`中进行 `notify()`。

类似这种 `wait - notify` 结构是代码中很常见的一种多线程协作结构，但上述代码中存在问题：

`wait()`有被伪唤醒的可能，什么是或怎么样伪唤醒先放一边，我们可以暂且认为被伪唤醒的线程，实际上没有达到唤醒条件。

那么会怎么样呢？ 假如 `wrongWaiter()` 被伪唤醒，则 `wait()` 语句阻塞解除，之后的逻辑会在不安全的情况下被执行，那么执行结果也就是不可信的了。

在某些情况下，这非常危险。

因此，对于防止伪唤醒，在使用 `wait()` 时，应该确保条件判断被放在 `while` 循环中，当被唤醒时再次判断条件，直到不满足为止，见如下代码：

```java
void correctWaiter() throws InterruptedException {
        synchronized (this) {
            while (condition) {
                wait();
            }
        }
    }
```

在《Java 并发编程实战》章节 14 中提到：当使用条件等待时：
- 在调用 `wait()` 前测试条件谓词
- 在循环中调用 `wait()`
- 确保先获取与条件队列相关的锁

#### 伪唤醒
> `wait()` 方法过早的返回

由于各种原因，`wait()` 可能会提前返回，可能的情况有：
- 由于调用 `notifyAll()` 使得所有正在等待的线程全部被唤醒
- 线程被中断
- `wait()` 超时
由于上述可能性，一个正在等待的线程，有可能在未满足条件谓词的情况下就被唤醒，此时如果不进行 re-check，将会导致错误。
#### 其他错误
上述代码中还有一点：唤醒使用的是`notify()` 而不是 `notifyAll()`。
通常情况下，`notifyAll()`的使用场景更加广泛，而未经仔细考虑的`notify()`容易出现问题。
`notify()`只唤醒一个线程，假设被`notify()`唤醒的线程不是程序期望执行逻辑的线程，那么真正期望被唤醒的线程就没有机会被唤醒，程序可能陷入死锁。
而使用`notifyAll()`会唤醒所有正在等待的线程，他们被唤醒后依次获取锁，并 re-check 条件谓词，进而选择是继续休眠还是开始工作。

《Java 并发编程实战》 中提到，只有两种情况才可以直接使用`notify()` :
> 1. **所有等待线程的类型都相同**：只有一个条件谓词与条件队列相关，并且每个线程在从 wait 返回后将执行相同的操作。
> 2. **单进单出**: 在条件变量上的每次通知，最多只能唤醒一个线程来执行。

## 慎用 Stream.peek()
工作中我经常会遇到这种情况：在使用 Stream 对某个对象集合进行 mapping 时，想要顺便修改其中的数据，例如，想要对一个 list 中对象的某个 field 统一赋值，并取第一个对其进行 mapping：
```java
class SomeClass {
    private int a;
    
    public void setA(int a) {
         this.a = a;
    } 
}

List<SomeClass> someList = fetchXxxList();

// 方法1：
somList.stream().forEach(e -> e.setA(someA));
Optional<OtherClass> first = somelist.stream().map(OtherClass::mapping).findFirst(); 

// 方法2：
Optional<OtherClass> first = somelist.stream().peek((e -> e.setA(someA))).map(OtherClass::mapping).findFirst(); 
```

很多时候我们都会觉得 peek 能用更优雅的方式实现我们的诉求，然而，会有静态检查器告诉我们：
`"Stream.peek" should be used with caution`

在Java Stream中，peek实际上是为了让我们做调试用的（就好像 peek 的释义一样），如果直接用它来实现功能，可能会遇到以下的情况：
1. peek 和 forEach 其实完全不同，peek 是一种中间操作（intermediate operation），而 forEach 是结束操作（terminal operation），由于 Stream 的 Lazy 策略， `Stream.of(“1”, “2”, “3”).peek(I -> println(i));` 什么也不会打印。
2. 即使使用是正确的，也保不齐在peek之后的终止阶段会因为什么特殊的原因只处理 Stream 中的几个元素。就如同上面的例子，代码的本意可能是对所有元素都setA，并输出第一个，但实际上，只有第一个元素的a值被set了。

基于上述的原因，我们还是按照代码中的 “方法1” 来写逻辑，更加健壮（即使不够简洁）。
