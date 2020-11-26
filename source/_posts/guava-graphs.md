---
title: （Guava 译文系列）图
date: 2020-06-20 21:42:32
tags:
- guava
- translation
categories:
- Guava
---

Guava 的`common.graph`库对图结构进行了建模，[图](https://en.wikipedia.org/wiki/Graph_\(discrete_mathematics\))，是一种包含实体及其之间关系的数据结构。这种结构的例子包括 web 页面与超链接、科学家和他们写的论文、机场以及机场之间的航路、以及个人与其家庭关系（谱系树）。图结构的目的在于能提供一种通用且可扩展的语言来描述上述这类数据。

<!-- more -->

## 定义

一个图包含了一组**节点（node）**（也叫顶点）和一组**边（edge）**（也叫连接或弧线）；每一条边连接了两个节点。与边关联的节点成为**端点（endpoints）**

（当我们在下文中介绍`Graph`接口时，我们使用小写的"graph"来指代图这种数据结构。当我们想要指代`Graph`接口时，我们会用大写。（译者注：译文中图数据结构与`Graph`接口的名字不存在混淆。））。

如果一条边拥有定义好的起点（他的**源 source**）和终点（他的**目标 target**，也叫目的地），那么这条边是**有向的（directed）**。否则便是**无向的（undirected）**。有向边适合表示不对称关系（派生自...， 连接到...， 被...撰写），而无向边适合表示对称关系（“与...为共同作者”，“两者之间距离...”， “与...是兄弟”）。

假如一个图的所有边都是有向的，则称为有向图，反正如果一个图的所有边都是无向的，则成为无向图。（`common.graph` 不支持同时包含有向边和无向边的图）。

举例如下：

```java
graph.addEdge(nodeU, nodeV, edgeUV);
```

*   `nodeU` 和 `nodeV` 互为相邻（ **adjacent**）
*   `edgeUV` 关联（ **incident**）  `nodeU` 到 `nodeV` (反之亦然)

如果 `graph` 是有向图，那么：

*   `nodeU` 是 `nodeV` 的一个前任（**predecessor**）
*   `nodeV` 是`nodeU` 的一个后继（**successor** ）
*   `edgeUV` 是`nodeU`的一条外向（ **outgoing**）边
*   `edgeUV` 是`nodeV`的一条内向（ **incoming**）边
*   `nodeU` 是 `edgeUV`的一个源（**source**）
*   `nodeV` 是`edgeUV`的一个目标（ **target** ）

如果 `graph` 是无向图，那么：

*   `nodeU` 是 `nodeV`的一个前任和后继
*   `nodeV` 是`nodeU`的一个前任和后继
*   `edgeUV` 是 `nodeU`的一条外向与内向边
*   `edgeUV` 是 `nodeV`的一条外向与内向边

所有这些关系，都与 `graph`有关。

**自循环（self-loop）**指一条边连接了一个节点到他自身，同样的，这也代表一条边的两个端点是同一个节点。如果一个自循环是有向的，那么他既是关联节点的外向边，也是内向边，且与之关联的节点既是源也是目标。

如果两条边以相同的顺序连接了相同的节点，则称这两条边**平行（parallel）**，而如果他们以相反的顺序连接了相同的节点，则称**反平行（antiparallel）**。（无向边不能成为反平行。）

举例如下：

```java
directedGraph.addEdge(nodeU, nodeV, edgeUV_a);
directedGraph.addEdge(nodeU, nodeV, edgeUV_b);
directedGraph.addEdge(nodeV, nodeU, edgeVU);

undirectedGraph.addEdge(nodeU, nodeV, edgeUV_a);
undirectedGraph.addEdge(nodeU, nodeV, edgeUV_b);
undirectedGraph.addEdge(nodeV, nodeU, edgeVU);

```

在 `directedGraph` 中，`edgeUV_a` 和 `edgeUV_b` 互为平行, 且都与 `edgeVU`互为反平行。

在 `undirectedGraph`中， `edgeUV_a`， `edgeUV_b`， 和 `edgeVU` 都互为平行。

## 能力

`common.graph` 聚焦于提供接口和类来支撑基于图的开发工作。它不提供类似于 I/O 的功能或是可视化的支持，它的实用工具选择也很有限。详情可见[FAQ](#faq)。

总体上， `common.graph` 支持了一下几种图：

*   有向图
*   无向图
*   能够关联值（权重、标签等等）的节点和边
*   允许/不允许自循环的图
*   允许/不允许平行边的图（有平行边的图有时会被称为multigraph）
*   图的节点/边支持以插入顺序、可排序或无序。

具体哪种`common.graph`支持哪些图，已在 javadoc 中说明。而对内置的各种图的实现类，其说明文档处于对应的`Builder`类的 javadoc 中。在库内对特殊类型的*实现类*（特别是第三方实现）并不要求支持所有种类，且可能在后续添加支持其他种类。

对于底层数据结构的选择，该库是不可知的：根据实现者对不同用例的优化，图关系的描述可以通过矩阵、邻接表、邻接映射等结构来表示。

`common.graph` 目前不包括对如下图种类的*明确*支持， 即使他们能够用现有的类型来建模：

*   树、森林
*   具有相同类型的元素（节点或边）具有不同的数据类型（例如：二部/k部图，多模图）
*   超图

`common.graph` 不允许图既包含有向边，又包含无向边。

 [`Graphs`] 类提供了一些基本实用工具（例如，对图的复制和比较）

## 图的种类

顶层的图接口有三个，以边的表示方式可以区分为：`Graph`， `ValueGraph`，和 `Network`。他们互为同级，不存在一种为另一种的子类型的关系。

每一个“顶层”接口都扩展自[`SuccessorsFunction`] 和 [`PredecessorsFunction`]。这些接口的意义在于当作只需要获取前任和后继节点的图算法（例如广度优先算法）的参数。当图的拥有者已经有了一种行之有效的图表示方法，且他们并不想专门把自己的表示法转化为`common.graph`而只想采用一种图算法的时候，这种顶层接口十分有用。

### Graph

[`Graph`]是最基础、最简单的一种图类型。他定义了一些底层操作符，来处理节点之间的关系，例如`successors(node)，` `adjacentNodes(node)，` 和 `inDegree(node)`。他的节点是第一等的唯一对象；你可以将之在`Graph`内部数据结构类比为`Map`的键。

`Graph`的边是完全匿名的；他们仅仅由其端点来定义。

示例用法：`Graph<Airport>`的边俩连接了能搭乘直达航班到达的机场。

### ValueGraph

[`ValueGraph`] 一样拥有 [`Graph`] 所拥有的所有与节点相关的方法，但增加了两个从特定边获取值的方法。

`ValueGraph`的每一条边都关联着一个用户指定的值。这些值不要求唯一（因为节点是唯一的）；一个`ValueGraph`和一个  `Graph ` 的关系可以类比为`Map` 和 `Set`；一个`Graph`的边是一对端点，而一个`ValueGraph`的边是一对端点与其值的映射。

[`ValueGraph`] 提供了一个 `asGraph()` 方法来返回一个`ValueGraph`的`Graph`视图。这允许以`Graph`为参数的方法同样能处理`ValueGraph`实例。

示例用法：`ValueGraph<Airport, Integer>`的边代表了被一条边连接的两个`Airport`之间的旅途时间。

### Network

[`Network`] 一样拥有 [`Graph`] 所拥有的所有与节点相关的方法，但增加了操作边与操作节点-边关系的相关方法，例如`outEdges(node)`，`incidentNodes(edge)`和 `edgesConnecting(nodeU, nodeV)`。

`Network`的边是第一等的唯一对象，就像所有图类型中的节点一样。边的唯一性约束使得 [`Network`]原生支持平行边，以及边之间关系的方法和节点-边之间关系的方法。

[`Network`] 提供了一个 `asGraph()` 方法来返回一个`Network`的 `Graph` 视图。这允许以`Graph`为参数的方法同样能处理`Network`实例。

示例用法：`Network<Airport, Flight>`的边代表了一个人能从一个机场到另一个机场可以搭乘的具体航班。

### 选择合适的 graph 类型

上述三种 graph，其本质的区别在于他们对边的不同表示方式。

[`Graph`] 中节点之间连接的边是匿名的，边本身并不拥有任何标识或属性。当每一对节点都最多被一条边连接，且不需要在边上关联任何信息的时候，你应该使用`Graph`。

[`ValueGraph`] 的边拥有自己的唯一或不唯一的值（例如边的权重或标签等）。当每一对节点都最多被一条边连接，且需要在边上关联信息，不同边上的信息并不要求唯一的时候（例如边的权重），你应该使用`ValueGraph`。

[`Network`] 的边是第一等唯一对象，就像节点一样。当边对象唯一，且期望实施对其引用的查询时，你应该使用`Network`。（注意这种唯一性允许`Network`支持平行边。）

## 构造 graph 实例

`common.graph`提供的实现类在设计上并不是 public 的。这减少了用户需要了解的 public 类型类的数量，也使浏览内置实现类提供的多种能力变得更容易，而不会让只想创建一个 graph 的用户感到不知所措。

为了创建 graph 类型中的某一种内建实现类的实例，可以使用对应的 `Builder` 类： [`GraphBuilder`]，[`ValueGraphBuilder`]，或 [`NetworkBuilder`]。例如：

```java
// Creating mutable graphs
MutableGraph<Integer> graph = GraphBuilder.undirected().build();

MutableValueGraph<City, Distance> roads = ValueGraphBuilder.directed()
    .incidentEdgeOrder(ElementOrder.stable())
    .build();

MutableNetwork<Webpage, Link> webSnapshot = NetworkBuilder.directed()
    .allowsParallelEdges(true)
    .nodeOrder(ElementOrder.natural())
    .expectedNodeCount(100000)
    .expectedEdgeCount(1000000)
    .build();

// Creating an immutable graph
ImmutableGraph<Country> countryAdjacencyGraph =
    GraphBuilder.undirected()
        .<Country>immutable()
        .putEdge(FRANCE, GERMANY)
        .putEdge(FRANCE, BELGIUM)
        .putEdge(GERMANY, BELGIUM)
        .addNode(ICELAND)
        .build();
```

*   你可以通过以下两种方式来获得一个 graph `Builder` 实例:
    *   调用静态方法 `directed()` 或 `undirected()`。每一个`Builder`提供的 graph 实例都会是有向或无向的。
    *   调用静态方法`from()`，他能返回一个基于已存在的 graph 实例的`Builder`。
*   在你创建好`Builder`实例之后，你可以选择指定其他的特性和能力。
*   构建可变的 graph 实例
    *   你可以通过对同一个`Builder`实例调用多次`build()`方法来构建相同配置的多个不同实例。
    *   你不需要指定`Builder `的元素类型，在 graph 类型本身上指定他们就足够了。
    *    `build()` 方法会返回一个对应 graph 类型的`Mutable` 子类型，他提供了修改方法，更多细节可见下文的["`Mutable` and `Immutable` graphs"](#mutable-and-immutable-graphs)章节。
*   构建不可变的 graph 实例
    *   在同一个`Builder`上多次调用`immmutable()`来获得多个相同配置的`ImmutableGraph.Builder`实例。
    *   你需要在调用`immutable`时指定元素类型。

### Builder 的约束 vs. 优化提示

`Builder`类型通常提供了两类可选项：约束和优化提示。

约束指定了一个由`Builder`创建的 graph 实例必须要满足的行为和属性，例如：

*   graph 是否有向
*   graph 是否允许自循环
*   graph 的边是否可排序

等等。

graph 的实现类可以选择性的使用优化提示来提高效率，例如，决定类型或是内部数据结构的初始大小。优化提示并不保证有任何效果。

每个 graph 类型都提供与其特定`Builder`约束相关的访问器，但并不提供优化提示的访问器。

## `Mutable` 和 `Immutable` 图

### `Mutable*` 类型

每个 graph 类型都有一个与之相关联的 `Mutable*` 子类型： [`MutableGraph`]，[`MutableValueGraph`]，和 [`MutableNetwork`]。这些子类型定义了对其进行修改的方法：

*   添加或删除节点功能：
    *   `addNode(node)` 和 `removeNode(node)`
*   添加或删除边功能：
    *   [`MutableGraph`]
        *   `putEdge(nodeU, nodeV)`
        *   `removeEdge(nodeU, nodeV)`
    *   [`MutableValueGraph`]
        *   `putEdgeValue(nodeU, nodeV, value)`
        *   `removeEdge(nodeU, nodeV)`
    *   [`MutableNetwork`]
        *   `addEdge(nodeU, nodeV, edge)`
        *   `removeEdge(edge)`

这种方式与传统的 Java 集合框架（也包括 Guava 的新集合类型）的工作方式不同；每种类型都包含（可选的）修改方法签名。我们选择将这些修改方法剥离开并放入子类型，有一部分鼓励防御型编程的考虑：通常，如果你的代码只是检查或遍历一个 graph 而并不改变他，那么代码的输入应该被指定为 [`Graph`]， [`ValueGraph`]，或
[`Network`] 而不是可变子类型。另一方面，如果你的代码的确需要修改一个对象，在一个带有“Mutable”标签的类型上工作有助于提醒你注意他会被修改这个事实。

由于 [`Graph`] 等都是接口，即使他们不包含可变方法，向调用者提供该接口实例也*不保证*不会被调用者修改，就像（实际上他是一个`Mutable*`子类型的实例一样），调用者可以把它强制转换为一个可变子类型。如果你想要提供一个契约性的保证，即作为方法参数或返回值的 graph 不可被改变，你应该使用`Immutable`实现类，详情见下文。

### `Immutable*` 实现

每一种 graph 类型还有一个相关联的 `Immutable`  实现。这些类与 Guava 的 `ImmutableSet` 、`ImmutableList` 、 `ImmutableMap` 类似：一旦创建，他们就再也不能被编辑了，同时，他们内部采用了高效的不可变数据结构。

与 Guava 的其他 `Immutable` 类型不同，这些实现并没有任何可变的方法签名，所以他们并不需要在被尝试改变时抛出  `UnsupportedOperationException` 异常。

你可以通过以下两种方式创建一个 `ImmutableGraph` 的实例。

使用 `GraphBuilder` ：

```java
ImmutableGraph<Country> immutableGraph1 =
    GraphBuilder.undirected()
        .<Country>immutable()
        .putEdge(FRANCE, GERMANY)
        .putEdge(FRANCE, BELGIUM)
        .putEdge(GERMANY, BELGIUM)
        .addNode(ICELAND)
        .build();
```

使用 `ImmutableGraph.copyOf()`:

```
ImmutableGraph<Integer> immutableGraph2 = ImmutableGraph.copyOf(otherGraph);
```

不可变图总能提供对关联边顺序稳定的保证。如果使用 `GraphBuilder` 来填充一个图，那么相关边的顺序将会在可能的情况下使用插入顺序（通过[`ElementOrder.stable()`]了解更多细节）。当使用 `copyOf` 时，相关边的顺序将会采用他们在被访问并复制时的顺序。

#### 保证

每一个 `Immutable*` 类型都能做出如下保证：

*   **浅不变性（shallow immutability）**: 元素不可被增加、删除或被替换
    (这些类并不实现 `Mutable*` 接口)
*   **确定性迭代（deterministic iteration）**: 迭代的顺序总与输入图的顺序一致
*   [**线程安全（thread safety）**](#synchronization): 多线程访问是安全的
*   **完整性（integrity）**: 该类型不能在包外被创建子类 (子类会让上述保证被破坏)

#### 把这些类当作是 "interfaces"， 而不是实现

每一个 `Immutable*` 类型都提供有意义的保证行为 -- 而不仅仅是具体的某个实现。你应当将他们视同接口。

若存储一个 `Immutable*` 的字段或方法返回值（类似`ImmutableGraph`）应该被声明为  `Immutable*` 类而不是其关联的接口类型（例如 `Graph`）。这向调用者传递了所有上述列举的语义保证，这是一种非常有用的信息。

另一方面，一个 `ImmutableGraph` 类型的参数通常会让调用者不快。因此，接受 `Graph` 更合适。

**警告**：就像[下文中提到的](#elements-and-mutable-state)，修改一个集合中包含的元素（在某种程度上影响了他的 `equals()` 行为），多数情况下是个坏主意。这会导致未定义的行为和一些 bug。所以最好的是使用不可变对象用作`Immutable*`实例的元素，因为用户可能希望你的“不可变”对象是完全不可变的。

## Graph 元素 (节点和边)

### 元素必须可用作 `Map` 的 key

用户提供 graph 元素应该被视作是 graph 内部实现维护的内部数据结构的 key。所以，作为代表 graph 元素的类，必须实现 `equals()` 和 `hashCode()`，或包含下面列举的属性。

#### 唯一性

如果 `A` 和 `B` 满足 `A.equals(B) == true` 那么这两个对象中至多有一个能作为 graph 的 key。

#### `hashCode()` 和 `equals()` 之间的一致性

`hashCode()` 必须与由[`Object.hashCode()`](https://docs.oracle.com/javase/8/docs/api/java/lang/Object.html#hashCode--) 定义的`equals()`保持一致。

####  `equals()` 的顺序一致性

假如节点是有序的（例如，通过 `GraphBuilder.orderNodes()` 创建的 graph），那么其顺序一定要与`equals()`保持一致，就如同 [`Comparator`] 和 [`Comparable`] 定义的一样。

#### 非递归性

`hashCode` 和 `equals()` *一定不能*递归引用其他元素，例如：

```java
// DON'T use a class like this as a graph element (or Map key/Set element)
public final class Node<T> {
  T value;
  Set<Node<T>> successors;

  public boolean equals(Object o) {
    Node<T> other = (Node<T>) o;
    return Objects.equals(value, other.value)
        && Objects.equals(successors, other.successors);
  }

  public int hashCode() {
    return Objects.hash(value, successors);
  }
}
```

当给 `common.graph`使用上述类作为其元素类型时 (例如， `Graph<Node<T>>`) 存在如下问题：

*   **冗余**： 由 `common.graph` 提供的 `Graph` 的内部实现已经提供了类似的关系。
*   **低效**： 添加/访问该元素时将会调用  `equals()` (可能还会调用 `hashCode()`)，这将需要 O(n) 的时间复杂度
*   **不可行**： 如果 graph 中包含环， `equals()` 和
    `hashCode()` 将无法终止

取而代之的， 仅使用 `T` 值自身来作为节点类型 (假设`T` 本身能用做 `Map` 的 key)。

### 元素和可变状态

如果 graph 的元素包含可变状态：

*   该可变状态一定不能反映在 `equals()/hashCode()` 方法中（本条详情在  `Map` 的文档中有讨论）
*   不要构建多个彼此相等的元素，并期望他们可以互换。尤其是，当将这种元素加入 graph 后，如果需要在创建过程中多次引用这些元素，则应该创建一次并存储引用（而不是将 `new MyMutableNode(id)` 传递给每个 `add*()` 调用）。

如果你需要存储每个元素的可变状态，一种选择是使用不可变元素并将可变状态存储在单独的数据结构中（例如，一个元素到状态的映射）。

### 元素必须非 null

向 graph 中添加元素的方法按照契约需要拒绝 null 元素。

## Graph 库的契约与行为

本节将会讨论`common.graph`类型内置实现的行为。

### 变更

你可以给一个还未被添加进 graph 的节点增加一个对应的边。若他们还没有准备好展示，则他们会静默的被添加进 graph：

```java
Graph<Integer> graph = GraphBuilder.directed().build();  // graph is empty
graph.putEdge(1, 2);  // this adds 1 and 2 as nodes of this graph, and puts
                      // an edge between them
if (graph.nodes().contains(1)) {  // evaluates to "true"
  ...
}
```

### Graph `equals()` 和 graph 相等

截止至 Guava 22，，每一个 `common.graph `的 graph 类型都以一种对特定类型合理的方式定义了  `equals()` ：

*   `Graph.equals()` 定义了两个 `Graph` 相等的条件是他们拥有相同的节点与边集合（即每一条边都有相同的终点和方向）。
*   `ValueGraph.equals()` 定义了两个  `ValueGraph` 相等的条件是他们拥有相同的节点与边集合，且相同的边拥有相同的值。
*   `Network.equals()` 定义了两个 `Network` 相等的条件是他们拥有相同的节点与边集合，且每一条边的对象都在相同的方向上连接了相同的节点。（如果有的话）

另外,对每一种 graph 类型，两个 graph 仅当他们的边拥有相同的方向性时才相等（要么二者都有向，要么都无向）。

当然，每一种 graph 类型的 `hashCode()` 都与 `equals()` 定义一致。

如果你只想基于连通性来比较两个 `Network` 或 `ValueGraph`，或是比较一个`Network` 或一个  `ValueGraph` 与一个 `Graph`，你可以使用 `Network` 和 `ValueGraph` 的`Graph` 视图。

```java
Graph<Integer> graph1, graph2;
ValueGraph<Integer, Double> valueGraph1, valueGraph2;
Network<Integer, MyEdge> network1, network2;

// compare based on nodes and node relationships only
if (graph1.equals(graph2)) { ... }
if (valueGraph1.asGraph().equals(valueGraph2.asGraph())) { ... }
if (network1.asGraph().equals(graph1.asGraph())) { ... }

// compare based on nodes, node relationships, and edge values
if (valueGraph1.equals(valueGraph2)) { ... }

// compare based on nodes, node relationships, and edge identities
if (network1.equals(network2)) { ... }
```

### 访问器方法

访问器将会返回集合：

*   也许是 graph 的视图；可能会影响视图的对 graph 的修改（例如，在用 `nodes()` 迭代时调用 `addNode(n)` 或 `removeNode(n)`）可能会抛出`ConcurrentModificationException`。
*   假如输入合法但并没有元素满足该请求时，将会返回空集合（例如：如果`node` 并没有相邻节点时，`adjacentNodes(node)`会返回空集合）。

假如传入的元素并在 graph 中，那么访问器会抛出`IllegalArgumentException`。

Java 的集合框架中的一些方法就像`contains()`会接受`Object`类型的参数而不是合适的泛型类，截止至 Guava 22，`common.graph` 的方法都会接受泛型类说明符来提升类型安全性。

### 同步

不同的 graph 实现会自主决定他们的同步策略。默认情况下，未定义的行为可能是由于调用正在被另一个线程所更改的 graph 中的任意方法引起的。

通常来说，内置的可变实现不提供任何同步保证，但`Immutable*`类是线程安全的（凭借他的不可变性）。

### 元素对象

你添加到 graph 中的节点、边、值等对象都与内置实现无关；他们只用作内部数据结构的 key。这表明节点/边也许可以在 graph 实例之间共享。

默认情况下，节点和边的对象遵从插入顺序（即，通过 `Iterator` 的  `nodes()` 和 `edges()` 访问的顺序就是他们被添加进 graph 的顺序，就像 `LinkedHashSet` 一样）。

## 实现者须知

### 存储模型

`common.graph` 支持多种机制来存储 graph 的拓扑，包括：

*   the graph implementation stores the topology (for example, by storing a
    `Map<N, Set<N>>` that maps nodes onto their adjacent nodes); this implies
    that the nodes are just keys, and can be shared among graphs
*   由 graph 的实现来存储拓扑（例如，通过存储一个 `Map<N, Set<N>>` 来将节点映射到他们相邻的节点）；这种实现中节点只作为 key，因此可以在 graph 之间共享。
*   由节点来存储拓扑（例如，通过存储一个相邻节点的  `List<E>`）；这种实现（通常）是单个 graph 专有的。
*   由一个独立的数据仓库（例如数据库）来存储拓扑

注意：`Multimap` 并不能满足用作Graph 实现需要支持节点隔离的要求（在节点没有关联边时），这是因为 `Multimap` 限制了一个 key 要么会映射到至少一个 value，要么就不会出现在`Multimap`中。

### 访问器行为

For accessors that return a collection, there are several options for the
semantics, including:

对于返回一个集合的访问器，在语义上有一些可选项，包括：

1.  当其集合是一个不可变副本时（例如 `ImmutableSet`）:任何尝试修改该集合的行为都会抛出一个异常，对 graph 的任何修改，都**不会**反映在该集合上。
2.  当其集合是一个不可变视图时（例如`Collections.unmodifiableSet()`）：任何尝试修改该集合的行为都会抛出一个异常，对 graph 的修改会反映在该集合上。
3.  当其集合是一个可变副本时：他可以被修改，但对 graph 的任何修改，都**不会**反映在该集合上。
4.  当其集合是一个可变视图时：他可以被修改，对 graph 的修改也会反映在该集合上。

（理论上，可以返回在一个方向上的写操作集合，但不能返回另一个方向（集合到 graph 或反之），但这基本上永远不会被用到，所以别这么干:)）

（1）和（2）通常更好；直到撰写本文时，内置实现通常都使用（2）。

（3）是一个可行的选项，但可能会在当用户期望对 graph 或集合的修改会影响另一方的时让用户感到混淆。

（4）是一种危险的设计选择，使用时应该特别小心，因为保持内部数据结构的一致性非常困难。

### `Abstract*` 类

每一个 graph 类型都对应了一个 `Abstract` 类：`AbstractGraph`，等等。

如果可能的话，对该 graph 接口的实现者应该继承合适的抽象类而不是直接去实现接口。抽象类提供了许多难以正确设计的关键方法实现，以及能够帮助给出一致性的实现，例如：

*   `*degree()`
*   `toString()`
*   `Graph.edges()`
*   `Network.asGraph()`

## 代码示例

### Graph 包含`node` 吗？

```java
graph.nodes().contains(node);
```

### 在节点 `u` 和 `v`之间存在边吗 （是 graph 中已知的吗）？

当 graph 是无向时， 下例中参数 `u` 和 `v` 的顺序无关。

```java
// This is the preferred syntax since 23.0 for all graph types.
graphs.hasEdgeConnecting(u, v);

// These are equivalent (to each other and to the above expression).
graph.successors(u).contains(v);
graph.predecessors(v).contains(u);

// This is equivalent to the expressions above if the graph is undirected.
graph.adjacentNodes(u).contains(v);

// This works only for Networks.
!network.edgesConnecting(u, v).isEmpty();

// This works only if "network" has at most a single edge connecting u to v.
network.edgeConnecting(u, v).isPresent();  // Java 8 only
network.edgeConnectingOrNull(u, v) != null;

// These work only for ValueGraphs.
valueGraph.edgeValue(u, v).isPresent();  // Java 8 only
valueGraph.edgeValueOrDefault(u, v, null) != null;
```

### 基础 `Graph` 示例

```java
ImmutableGraph<Integer> graph =
    GraphBuilder.directed()
        .<Integer>immutable()
        .addNode(1)
        .putEdge(2, 3) // also adds nodes 2 and 3 if not already present
        .putEdge(2, 3) // no effect; Graph does not support parallel edges
        .build();

Set<Integer> successorsOfTwo = graph.successors(2); // returns {3}
```

### 基础 [`ValueGraph`] 示例

```java
MutableValueGraph<Integer, Double> weightedGraph = ValueGraphBuilder.directed().build();
weightedGraph.addNode(1);
weightedGraph.putEdgeValue(2, 3, 1.5);  // also adds nodes 2 and 3 if not already present
weightedGraph.putEdgeValue(3, 5, 1.5);  // edge values (like Map values) need not be unique
...
weightedGraph.putEdgeValue(2, 3, 2.0);  // updates the value for (2,3) to 2.0
```

### 基础 [`Network`] 示例

```java
MutableNetwork<Integer, String> network = NetworkBuilder.directed().build();
network.addNode(1);
network.addEdge("2->3", 2, 3);  // also adds nodes 2 and 3 if not already present

Set<Integer> successorsOfTwo = network.successors(2);  // returns {3}
Set<String> outEdgesOfTwo = network.outEdges(2);   // returns {"2->3"}

network.addEdge("2->3 too", 2, 3);  // throws; Network disallows parallel edges
                                    // by default
network.addEdge("2->3", 2, 3);  // no effect; this edge is already present
                                // and connecting these nodes in this order

Set<String> inEdgesOfFour = network.inEdges(4); // throws; node not in graph
```

### 逐节点遍历无向图

```java
// Return all nodes reachable by traversing 2 edges starting from "node"
// (ignoring edge direction and edge weights, if any, and not including "node").
Set<N> getTwoHopNeighbors(Graph<N> graph, N node) {
  Set<N> twoHopNeighbors = new HashSet<>();
  for (N neighbor : graph.adjacentNodes(node)) {
    twoHopNeighbors.addAll(graph.adjacentNodes(neighbor));
  }
  twoHopNeighbors.remove(node);
  return twoHopNeighbors;
}
```

### 逐边遍历有向图

```java
// Update the shortest-path weighted distances of the successors to "node"
// in a directed Network (inner loop of Dijkstra's algorithm)
// given a known distance for {@code node} stored in a {@code Map<N, Double>},
// and a {@code Function<E, Double>} for retrieving a weight for an edge.
void updateDistancesFrom(Network<N, E> network, N node) {
  double nodeDistance = distances.get(node);
  for (E outEdge : network.outEdges(node)) {
    N target = network.target(outEdge);
    double targetDistance = nodeDistance + edgeWeights.apply(outEdge);
    if (targetDistance < distances.getOrDefault(target, Double.MAX_VALUE)) {
      distances.put(target, targetDistance);
    }
  }
}
```

## FAQ

### 为什么 Guava 要引入 `common.graph`？

正如 Guava 所做的其他事情一样，引入 graph 也是基于同样的理由：

*   代码重用、互用、范例统一：很多事情都与 graph 处理有关
*   效率：有多少代码都在使用低效的 graph 表示？太多了（例如矩阵的表示）
*   正确性：有多少代码都将 graph 分析做错了？
*   推广 graph 用作 ADT：当 graph 用起来很简单的时候，有多少人都会用？
*   简单性：假如显式的使用比喻，处理 graph 的代码会变得更容易理解。

### `common.graph` 支持哪些类型的 graph？

请见上文章节： ["能力"](#能力) 。

### `common.graph` 并不包含某个特性/算法，你们可以增加吗？

也许吧，你可以给我们的邮箱`guava-discuss@googlegroups.com`发邮件或 [在 Github 上提 Issue](https://github.com/google/guava/issues)。

我们的处世哲学是只有当某种东西（a）与 Guava 的核心使命相匹配且（b）有一个好的理由来期望他能够合理广泛的被使用时，才应该是 Guava 的一部分。

`common.graph`也许永远也不会提供可视化或者 I/O 的能力；这些都是他们自己项目中的内容，与 Guava 的使命并不相符。

类似遍历、过滤、变换等的能力才更符合，因此也更有可能被引入，虽然最终我们仍期望其他的 graph 库能提供大部分能力。

###  超大规模的 Graph 会被支持吗（例如 MapReduce 规模）？

现在还不行。Graph 在较小的百万级别节点下应该能工作，但你考虑应该将本库类比为 Java 的集合框架类型（`Map`， `List`，`Set`等等）。

### 我如何能定义 `successors(node)`的顺序？

在 graph builder 中设置`incidentEdgeOrder()`为 [`ElementOrder.stable()`]就能确保  `successors(node)` 会以边的插入顺序返回 `node`的后继。这在对其他与边和节点相关方法（例如`incidentEdges(node)`）时也有效。

### 为什么我要用 Guava Graph 库而不是其他库呢？

**太长不读**：你应该使用对你奏效的，但当本库不支持你的需求时，请让我们知道！

本库的主要竞争者（对 Java）是：[JUNG](https://github.com/jrtom/jung) 和 [JGraphT](http://jgrapht.org/)。

`JUNG` 是 Joshua O'Madadhain（ `common.graph` 的带头人）在 2003年与其他人共同创建的，
他现在仍在维护它。 JUNG 是一个相当成熟、功能齐全且被广泛使用的库，但在很多地方粗陋且低效。 现在 `common.graph` 已经对外发布了，他目前工作在一个新的 `JUNG`版本上，该版本试用了`common.graph` 来作为他的数据模型。

`JGraphT` 是另一个已经存在了一段时间的第三方 Java graph 库。我们对它并不熟悉，所以我们并不能评价他的细节，但是它至少在一些地方与 `JUNG` 是相同的。这个库也包含了很多 [适配器类](https://jgrapht.org/javadoc/org/jgrapht/graph/guava/package-summary.html)来将`common.graph`适配到`JGraphT`。

如果你有非常特别的需求的话，推出自己的解决方案有时是正确的方法。但是就像通常你不会在 Java 中实现自己的 hash table（而不是使用 `HashMap` 或 `ImmutableMap`） 一样，基于以上列出的所有原因，你应该考虑使用 `common.graph` （或者，如果有需要，使用其他现存的 graph 库）。

## 主要贡献者

`common.graph` 是一个团队合作的成果，我们受到了 Google 内外的各种人的帮助，但这些人的影响最大。

*   **Omar Darwish** 完成了很多早期实现， 并设置了测试覆盖标准。
*   [**James Sexton**](https://github.com/bezier89) 是对项目最多产的个人，他在方向和设计上拥有显著的影响力。他负责一些核心特性，以及我们提供实现的效率。
*   [**Joshua O'Madadhain**](https://github.com/jrtom) 在反思了他也参与创建的 [JUNG ](http://jung.sf.net)的优劣势之后，开启了 `common.graph` 项目。他作为项目带头人审阅或编写了设计和代码的几乎各个方面。
*   [**Jens Nyman**](https://github.com/nymanjens) 贡献了非常多近期的插件例如 [`Traverser`] 以及不可变 graph 的 builder。他对项目的未来发展方向也有重大影响。

[`Comparable`]: https://docs.oracle.com/javase/8/docs/api/java/lang/Comparable.html
[`Comparator`]: https://docs.oracle.com/javase/8/docs/api/java/util/Comparator.html
[`ElementOrder.stable()`]: https://guava.dev/releases/snapshot/api/docs/com/google/common/graph/ElementOrder.html#stable--
[`Graph`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/Graph.html
[`GraphBuilder`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/GraphBuilder.html
[`Graphs`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/Graphs.html
[`ImmutableGraph`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/ImmutableGraph.html
[`ImmutableNetwork`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/ImmutableNetwork.html
[`MutableGraph`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/MutableGraph.html
[`MutableNetwork`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/MutableNetwork.html
[`MutableValueGraph`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/MutableValueGraph.html
[`Network`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/Network.html
[`NetworkBuilder`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/NetworkBuilder.html
[`PredecessorsFunction`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/PredecessorsFunction.html
[`SuccessorsFunction`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/SuccessorsFunction.html
[`Traverser`]: https://guava.dev/releases/snapshot/api/docs/com/google/common/graph/Traverser.html
[`ValueGraph`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/ValueGraph.html
[`ValueGraphBuilder`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/graph/ValueGraphBuilder.html