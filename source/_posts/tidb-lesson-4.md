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

Executor 的所有逻辑都放置在 `executor` 包下，其中：`adapter.go` 对应了包模块的上层出入口。

`adapter.go`顾名思义是用于适配，它主要与执行计划交互，用于由外部调用来执行操作，并返回操作结果。`adapter.go` 提供了 `ExecStmt` 来实现构造执行器并执行，提供了 `RecordSet` 通过执行器获取结果集。

与下层 tikv 的交互逻辑，则散落在各种不同的 Executor 实现中。

Executor 都提供了 `Open()` 与 `Next()` 方法来实现对自身的初始化以及实际的执行。

#### Executor 类型

根据不同的操作，Executor 包含了很多种类，不过总体来看，所有的 Executor 都能分成两类：

- 需要返回结果的类型：例如各种单表、连表查询
- 不需要返回结果的类型：例如插入、更新等

对于不需要返回结果类型的 Executor，其 `Next()` 会立即执行，相关的逻辑在`adapter.go`的`handleNoDelay()` 方法中实现。

而对于需要返回结果的 Executor，不会立即执行，而是会在`Open()`被调用后，构造一个 `ResultSet` 结构，包含相关上下文，最终的读取过程在`conn.go` 的 `handleStmt()` 方法中 `err = cc.writeResultset(ctx, rs, false, status, 0)` 这句话里实际的执行，并获取结果。

接下来会指定两个具有代表性的 Executor 来分别介绍上述两种类型的操作过程。

#### InsertExec 执行过程介绍

连接 TiDB 后执行：

```sql
> create table test (
	id int primary key,
  name varchar(20)
);

> explain insert into test values(1, 'a');
+----------+---------+------+---------------+---------------+
| id       | estRows | task | access object | operator info |
+----------+---------+------+---------------+---------------+
| Insert_1 | N/A     | root |               | N/A           |
+----------+---------+------+---------------+---------------+
```

我们可以看到对于一个最简单的插入语句，TiDB 给出的执行计划是仅使用 Insert 执行器来执行。



#### TableReaderExec 执行过程介绍