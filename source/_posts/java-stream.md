---
title: 虚拟工厂：Java stream
date: 2020-05-03 23:07:27
tags:
- java
- stream
categories:
- Java
---

在我的文章 [虚拟工厂：Java 线程池]() 中简单介绍了采用 ”工厂“ 这样一个现实生活中的概念来抽象具体的线程操作，并定义了 `Worker` `TaskQueue` 等概念。用这种方式拉近了计算机域和真实世界域之间的距离，让代码表现现实的意图。

不仅仅是线程池，从 Java 8 开始引入的`Stream`也一样，它用代码构建了一套流水线体系，通过流水线环节的叠加来实现对流水线上元素的各种处理。

### 流水线（Pipeline）

在中文翻译中，我们时常会把 Assembly Line（装配线）以及 Pipeline（管线）都译为流水线，而在计算机领域，我们说的流水线通常都是 Pipeline，例如 CPU 的指令流水线（Instruction Pipeline）。

对于 Pipeline，韦氏词典的第一个解释即：

> a line of pipe with pumps, valves, and control devices for conveying liquids, gases, or finely divided solids

显然对于上述释义，Pipeline 翻译为 “管线” 似乎更为合理，不过在软件领域，对 Pipeline 进行了引申，一个 Pipeline 是一组计算过程（computing processes）的组合，并且以并行（Parallel）的方式执行。

那么实际上在计算机领域，Pipeline 与 Assembly Line 在概念上就没有太明确的区别了：我们可以将之类比为一种操作，它可以由多个工序构成，每个工序都将对上一个工序产出的工件进行进一步加工，工件从流水线入口进入，并在流水线上移动，最后被输出为某种产品。

所以根据上述描述我们可以抽象出与流水线相关的几个概念：

- 工件（workpiece）：即被流水线加工的元素
- 工序（process）：即执行单元，它以工件为输入，也以工件为输出
- 流水线（pipeline）：即流水线主体，他包括了入口，出口，在其上可以放置各种工序

那么我们就能得到如下的一个流水线结构：
{% asset_img pipeline.png %}

可见流水线上的工序理应能够灵活的组合与替换，采用各种简单而固定的工序，就能组合出来满足多种多样的需求。

### Stream

Java 8 中定义的流式操作，能够对流（Stream）叠加多种操作并进行处理。在 [Processing Data with Java SE 8 Streams, Part 1](https://www.oracle.com/technical-resources/articles/java/ma14-java-se-8-streams.html) 中提到对 Stream 的定义：

> a sequence of elements from a source that supports aggregate operations.

首先，任何单一对象、集合、数组都可以输入为一个 Stream，这就是定义里提到的 “source”。转换为 Stream 后，他们都将变身为流中的一个元素，即 “sequence of element”。在对所有元素进行一番处理之后，Stream 可以再次转换回到对象、集合、数组等，这里的处理即 “aggregate operations”。

那么，对于这 ”一番处理“，从处理的方式上划分，有以下几种[处理形式](https://docs.oracle.com/javase/8/docs/api/java/util/stream/package-summary.html#StreamOps)：

1. 中间操作（Intermediate Operations）：会返回一个新的 Stream，并且是 Lazy 操作

   - 无状态操作（StatelessOps）：对单个元素进行操作，操作之间没有联系，也不保存任何元素的状态。典型操作有：`filter` `maap` `peak` 等

   - 有状态操作（StatefulOps）：对流中元素的处理会依赖之前处理的结果，元素与元素之间有关系，下一个元素的处理依赖上一个元素的状态，或需要获取到所有的元素后才能进行操作。典型操作有：`distinct` `sorted` `limit`

2. 终止操作（Termination Operations）：穿越（traverse）整个流，得到结果或是副作用（side-effect），一旦执行了终止操作，整个流就认为已经被消费，并且无法再次执行任何操作。

#### Stream Shape

通常，Stream 中的元素都以对象引用的形式存在，但同时 Java 也考虑到了对基本类型的 Stream 支持，因此共定义了四种 “Stream Shape”：

- **Reference**：元素类型为对象引用，其行为由 `Stream` 定义
- **Int Value**：元素类型为 int，其行为由`IntStream`定义
- **Long Value**：元素类型为 long，其行为由`Longtream`定义
- **Double Value**：元素类型为 double，其行为由`DoubleStream`定义

下文中主要以 Reference 类型的 Stream 来举例说明流水线的工作原理。

### ReferencePipeline

对于元素为对象引用的 Stream（也是大多数 Stream 的形态）而言，`ReferencePipeline` 是其实现的核心。

`ReferencePipeline `的类继承关系如下图所示：

{% asset_img reference-pipeline.png %}

此外，由于`ReferencePipeline`本身实现了 `Streaam`接口，因此他实现了`Stream`中定义的所有行为，包括：`map`，`filter`，`reduce`，`collect`，`limit` 等等。所以我们可以说`ReferencePipeline`本身就是一个流水线的基础工序，基于这种基础工序，我们能构造出多种多样的工序来。

`ReferencePipeline` 作为基类，在其内部提供了如下三个实现类，这三个实现类进一步对不同种类的工序进行了定义：

- Head：
  - 与其他两个实现类相比，`Head`比较独特，也比较好理解：他会被作为整个 `Stream` 的头部，也即流水线的入口。又因为`Head`本身是一个`ReferencePipeline`，因此我们可以把`Head`理解为一道特殊的工序，他不对工件做任何处理，但作为头部（第一道工序），我们可以在其之后追加更多的工序。
- StatelessOp：无状态工序
  - 对应前文 ”中间操作“ 中的无状态操作
  - 无状态工序不依赖前后两个工件的状态，即不论前后工件如何，他只对当前工件进行加工，典型的无状态工序是`map`
- StatefulOp：有状态工序
  - 对应前文 ”中间操作“ 中的有状态操作
  - 有状态工序依赖与工件之间的状态，他能结合多个工件来进行处理，典型的有状态工序是`collect`

不出意外，我们也许能进一步想到：

- `map `操作其实就是一个行为是 ”对工件进行映射变换“ 的`StatelessOp` 工序
- `filter`操作其实就是一个行为是 ”对工件进行筛选“ 的`StatelessOp`工序
- `limit`操作其实就是一个行为是 ”只保留有限个工件“ 的`StatefulOp`工序
- `collect`操作其实就是一个行为是 ”收集所有工件进行整合“ 的`StatefulOp`工序

因此我们日常使用的形如：

```java
Stream.of("a", "b", "c", "1", "2", "3").filter(NumUtil::isNum).map(NumUtil::minusOne).limit(1);
```

其实就是先创造了一条这样的流水线：

`Head` -> `StatelessOp:filter`  -> `StatelessOp:map`   -> `StatelessOp:filter`  -> `StatefulOp:limit`

之后输入工件：`"a", "b", "c", "1", "2", "3"`，最后启动流水线产出结果。

#### 流水线的组装

