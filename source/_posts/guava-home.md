---
title: Guava 用户手册
date: 2019-08-17 22:29:30
tags:
- guava
- translation
categories:
- Guava
---

> 本文系对 Guava 用户手册的中文译文，原文详见 https://github.com/google/guava/wiki

## 用户手册

Guava 项包含了被 Google 的 Java 项目依赖的多个核心库：集合、缓存、基本类型支持、并发库、通用注解、字符串处理、I/O 等等。Googler 们每天都在生产服务中使用这些工具。

不过通过搜索 Javadoc 并不是一个学习如何更好的使用这些库的高效方式。这里我们尝试提供一个对 Guava 的可阅读且令人愉悦的解释文稿，这其中包含一些最流行和最有用的功能。

*本 wiki 文档仍然在更新中，其中一些段落可能还没有完成。*

- 基本工具类：让 Java 语言用起来更愉快
    - [使用和避免使用 null](https://github.com/google/guava/wiki/UsingAndAvoidingNullExplained)：null 可能会引起模棱两可，可能会造成令人迷惑的错误，有时只是不够优雅。 许多 Guava 工具不会盲目的接受 null，而是会拒绝 null 且会对 null 值快速失败。
    - [前置条件](https://github.com/google/guava/wiki/PreconditionsExplained)：更简单的在方法中测试前置条件。
    - [通用Object方法](https://github.com/google/guava/wiki/CommonObjectUtilitiesExplained)：对 Object 方法的简化实现，例如 hashCode() 和 toString()。
    - [排序](https://github.com/google/guava/wiki/OrderingExplained)：Guava 强大的“流式 Comparator”类。
    - [异常](https://github.com/google/guava/wiki/ThrowablesExplained)：简化的异常和错误的传播和检查方案。

- 集合：对 JDK 集合生态的 Guava 扩展。这里包含部分Guava 中非常成熟和流行的部分。
    - [不可变集合](https://github.com/google/guava/wiki/ImmutableCollectionsExplained)，用于防御性编程，常量集合，提升效率。
    - [新集合类型](https://github.com/google/guava/wiki/NewCollectionTypesExplained)，包含 JDK 集合类未实现的场景：multiset，multimap，tables，双向 map等等。
    - [强大的集合工具](https://github.com/google/guava/wiki/CollectionUtilitiesExplained)，JDK 集合类未提供的通用操作。
    - [扩展工具](https://github.com/google/guava/wiki/CollectionHelpersExplained)：写个集合装饰器？实现一个迭代器？我们将这些变得简单。






