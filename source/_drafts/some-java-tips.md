---
title: 一些 Java 语言 Tips
date: 2020-04-04 20:56:16
tags: 
- java
- tips
categories:
- Java
---
## 一些 Java 语言 Tips
本文包括一些 Java 语言在使用中常见的小错误以及不佳实践，他们收集自我的日常开发、Code Review、以及书中所见。持续更新...

### 双括号初始化
在 Java 9 之前，想要在 Field 中初始化一个集合供其他方法使用，他的样子是尴尬而丑陋的：
```java
public class DoubleBraceInitExample {
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
    
    ......
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
