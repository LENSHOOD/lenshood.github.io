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
    - [使用和避免使用 null](https://lenshood.github.io/2019/08/20/guava-using-and-avoiding-null-explained/)：null 可能会引起模棱两可，可能会造成令人迷惑的错误，有时只是不够优雅。 许多 Guava 工具不会盲目的接受 null，而是会拒绝 null 且会对 null 值快速失败。
    - [前置条件](https://github.com/google/guava/wiki/PreconditionsExplained)：更简单的在方法中测试前置条件。
    - [通用Object方法](https://github.com/google/guava/wiki/CommonObjectUtilitiesExplained)：对 Object 方法的简化实现，例如 hashCode() 和 toString()。
    - [排序](https://github.com/google/guava/wiki/OrderingExplained)：Guava 强大的“流式 Comparator”类。
    - [异常](https://github.com/google/guava/wiki/ThrowablesExplained)：简化的异常和错误的传播和检查方案。

- 集合：对 JDK 集合生态的 Guava 扩展。这里包含部分Guava 中非常成熟和流行的部分。
    - [不可变集合](https://github.com/google/guava/wiki/ImmutableCollectionsExplained)，用于防御性编程，常量集合，提升效率。
    - [新集合类型](https://github.com/google/guava/wiki/NewCollectionTypesExplained)，包含 JDK 集合类未实现的场景：multiset，multimap，tables，双向 map等等。
    - [强大的集合工具](https://github.com/google/guava/wiki/CollectionUtilitiesExplained)，JDK 集合类未提供的通用操作。
    - [扩展工具](https://github.com/google/guava/wiki/CollectionHelpersExplained)：写个集合装饰器？实现一个迭代器？我们将这些变得简单。

- [图](https://github.com/google/guava/wiki/GraphsExplained)：对[图形化](https://en.wikipedia.org/wiki/Graph_(discrete_mathematics))数据建模的库，即实体及其之间的关系。关键特性包括：
	- [图](https://github.com/google/guava/wiki/GraphsExplained#graph)：边为不包含任何身份或其个体信息的匿名实体的一种图。
	- [有值图](https://github.com/google/guava/wiki/GraphsExplained#valuegraph)：边与非唯一值相关联的一种图。
	- [网络](https://github.com/google/guava/wiki/GraphsExplained#network)：边缘是唯一对象的一种图。
	- 支持可变与不可变、有向和无向以及其他一些属性的图。

- [缓存](https://github.com/google/guava/wiki/CachesExplained)：本地缓存，支持多种过期行为。
- [函数式语法](https://github.com/google/guava/wiki/FunctionalExplained)：谨慎使用，Guava的函数式语法能够显著的简化代码。
- 并发：很强大，其简单的抽象使得编写正确的并发代码更为简单。
	- [ListenableFuture](https://github.com/google/guava/wiki/ListenableFutureExplained)：Future，当其结束时进行回调。
	- [Service](https://github.com/google/guava/wiki/ServiceExplained)：帮你处理启动、停止中复杂的状态逻辑。

- [Strings](https://github.com/google/guava/wiki/StringsExplained)：一些超有用的 String 工具类：splitting, joining, padding 等等。

- [扩展基本类型](https://github.com/google/guava/wiki/PrimitivesExplained)：对 JDK 未包含的基本类型，例如`int`，`char`等等的操作，包括一些类型的无符号变量。

- [Ranges](https://github.com/google/guava/wiki/RangesExplained)：强大的 Guava API，可处理`Comparable`类型的范围数据，包括连续和离散的数据。

- [I/O](https://github.com/google/guava/wiki/IOExplained)：对所有 Java 5 和 6 的流、文件的简化 I/O 操作。

- [Hashing](https://github.com/google/guava/wiki/HashingExplained)：比`Object.hashCode()`更成熟的 Hashing 工具，包括布隆过滤器。

- [事件总线](https://github.com/google/guava/wiki/EventBusExplained)：采用发布订阅模型进行通信，元素之间无需相互注册。

- [数学运算](https://github.com/google/guava/wiki/MathExplained)：对 JDK 未提供的数学运算工具的优化，经过了仔细的验证。

- [反射](https://github.com/google/guava/wiki/ReflectionExplained)：Java 反射能力的 Guava 工具类。

- 提示：用 Guava 来使你的应用程序按照你的想法来执行：
	- [Guava 的思想](https://github.com/google/guava/wiki/PhilosophyExplained)：Guava 是什么，不是什么，以及我们的目标。
	- [在构建中使用 Guava](https://github.com/google/guava/wiki/UseGuavaInYourBuild)：在构建工具中使用 Guava，例如 Maven、Gradle 等等。
	- [使用 ProGuard](https://github.com/google/guava/wiki/UsingProGuardWithGuava)：避免在你的 JAR 中打包未使用到的 Guava 组件。
	- [Apache Commons 对照](https://github.com/google/guava/wiki/ApacheCommonCollectionsEquivalents)：帮助你从 Apache Commons 集合类转换到 Guava。
	- [兼容性](https://github.com/google/guava/wiki/Compatibility)：Guava 的版本详细。
	- [创意墓地](https://github.com/google/guava/wiki/IdeaGraveyard)：最终被拒绝的功能请求。
	- [友情项目](https://github.com/google/guava/wiki/FriendsOfGuava)：一些我们喜欢且仰慕的开源项目。
	- [如何贡献](https://github.com/google/guava/wiki/HowToContribute)：如何对 Guava 进行贡献。

**留意**：如需对本 wiki 的内容进行讨论，请仅使用 guava-discuss 邮件列表。




