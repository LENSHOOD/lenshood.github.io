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
