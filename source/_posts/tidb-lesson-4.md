---
title: TiDB 学习课程 Lesson-4
date: 2020-09-06 23:43:18
tags:
- tidb
categories:
- TiDB
---

本次课程主要涉及 TiDB 中的 Executor 组件。

### TiDB SQL 层

在介绍 Executor 层之前，我们先从总体上来看一看 TiDB 的 SQL 架构，这里借用一张[官方博客的图](![SQL 层架构](https://download.pingcap.com/images/blog-cn/tidb-source-code-reading-2/2.png))：

![](https://download.pingcap.com/images/blog-cn/tidb-source-code-reading-2/2.png)

我们可以看到，从客户端发来的一个 SQL 请求，会先经过 Protocol Layer 进行预处理，每一个请求都转换成一个 Session Context，之后进入 SQL 层进行操作。粗略的看，我们能发现在 Executor 处，整个操作有可能会继续下探到 TiKV 层，也有可能直接由 Local Executor 返回。

仔细看 SQL 层的架构，实际上还是像一条生产线，通过对指令进行理解，来执行相应的操作。

1. 绿色部分：从协议层解析出来的 SQL 语句，是以文本形式存在的，那么如何让程序去理解这条语句的意图、嵌套结构，以及校验语句的正确性呢？通过 Parser（实际上采用了 yacc Parser Generator 来生成复杂的转换器）将 SQL 语句转换为一颗 AST（抽象语法树），对这颗树进行分析、校验后生成一个 stmt 结构。
2. 黄色部分：分析上述生成的 AST，并生成实际的执行计划，对该执行计划进行逻辑与物理优化，使之尽可能达到性能最优。这里生成最终的执行计划，即我们执行 `Explain xxx` 语句返回的执行计划了。
3. 深蓝色部分：执行计划会被转化为具体的执行器（Executor），执行器采用了 [Volcano 模型](https://paperhub.s3.amazonaws.com/dace52a42c07f7f8348b08dc2b186061.pdf)来实现，Volcano 模型简单来讲就是一颗操作树，树的每一层都先调用下层获取数据，之后对获取到的数据进行加工后返回给上层。

所以，我们可以从 SQL 语句的复杂性与多样性来得出判断：不同的操作（DML、DQL、join、sort、aggregate、index...）会生成不同的执行计划，不同的执行计划会转化为不同的执行器，因此需要通过多样的执行器来满足多样的 SQL 操作。

### Executor 的执行路径

上一节我们了解了在 SQL 层的执行路径，这一节我们一起梳理一下 Executor 的执行路径。

