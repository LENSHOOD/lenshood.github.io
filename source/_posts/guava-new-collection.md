---
title: （Guava 译文系列）新集合类型
date: 2019-11-15 11:19:30
tags:
- guava
- translation
categories:
- Guava
---
# 新集合类型

Guava 引入了许多 JDK 未包含但我们却广泛使用的的新集合类型。他们全都设计为可以和 JDK 的集合框架友好的共存， 而不是硬塞入 JDK 的集合类抽象中。

一般来说，Guava 的集合会精确地按照 JDK interface 所定义的契约来实现。

## Multiset

按照传统的 Java 习惯，计算单词在文档中出现的次数通常会按照如下方式：

```java
Map<String, Integer> counts = new HashMap<String, Integer>();
for (String word : words) {
  Integer count = counts.get(word);
  if (count == null) {
    counts.put(word, 1);
  } else {
    counts.put(word, count + 1);
  }
}
```

上述方案不仅尴尬，容易出错，而且不支持收集更多的有用的统计信息，比如单词总数等。我们可以做得更好。

Guava 提供了新的集合类型，[`Multiset`]，支持多次添加同一个元素。维基百科定义 Multiset 为在数学上 “集合概念的一种推广，其中成员可以出现多次...在 multisets 中，与 set 类似，与 tuple 相反，其与成员的顺序无关：{a, a, b} 和 {a, b, a} 是等同的 multiset”。

有两种主流的方式来理解：

- multiset 就像一个没有顺序限制的`ArrayList<E>`：顺序无关紧要。
- 这就像是一个`Map<E, Integer>`，包括元素和计数。

API结合了上述对`Multiset`的两种理解方式，如下：

- 当他被视为一个普通的`Collection`, `Multiset`表现的行为更像是一个未排序的`ArrayList`：
    *   调用`add(E)`将给定的元素添加单个引用。
    *   Multiset 的`iterator()`遍历每一个元素的每一个引用
    *   Multiset 的`size()`为每一个元素的每一个引用的总和
*   附加的查询操作及其性能特性，与你对`Map<E, Integer>`的期望一致。
    *   `count(Object)`返回与该元素相关联的计数。对`HashMultiset`，计数的复杂度是O(1)，对`TreeMultiset`，计数的复杂度O(log n)。
    *   `entrySet()` 返回一个`Set<Multiset.Entry<E>>`，与`Map`的 entrySet 类似。
    *   `elementSet()`返回一个对 multiset 元素去重的`Set<E>`，就像`Map`的`keySet()`。
    *   `Multiset`实现的内存消耗与不同元素的数量线性相关。

值得注意的是，`Multiset` 的契约与 `Collection` interface 完全一致，除非在极少数的情况下存在 JDK 自身的先例。特别的，`TreeMultiset` 像 `TreeSet` 一样, 使用比较来判断相等而不是`Object.equals`.另外，`Multiset.addAll(Collection)`方法，会在`Collection`中的元素每次出现时，对其引用加一，这样会比先前采用循环添加`Map`的方式方便很多。

Method               | Description
:------------------- | :----------
[`count(E)`]         | 计算被加入到 multiset 中的元素出现的次数
[`elementSet()`]     | 以`Set<E>`的形式查看去重了的 `Multiset<E>`元素
[`entrySet()`]       | 类似 `Map.entrySet()`, 返回一个 `Set<Multiset.Entry<E>>`, 包含支持 `getElement()` and `getCount()` 的 entry.
[`add(E, int)`]      | 给指定的元素增加指定的引用次数
[`remove(E, int)`]   | 给指定的元素减少指定的引用次数
[`setCount(E, int)`] | 给指定的元素设置指定的非负引用次数
`size()`             | 返回`Multiset`中的所有元素的所有引用次数的总和

### Multiset 不是 Map

请注意 `Multiset<E>` *不是* `Map<E, Integer>`, 即便是它可能属于 `Multiset` 实现的一部分. `Multiset` 是一个真正的 `Collection` 类型, 满足其所有相关的契约义务. 其他值得注意的不同点包括：

*   一个`Multiset<E>`只能包含记录正数次数的元素。没有元素可以拥有负数计数，且计数为`0`的值被认为不包含在 multiset 中。他们不会在`elementSet()` 或 `entrySet()`中出现。
*   `multiset.size()`返回集合的尺寸，等于所有元素的计数之和。如果想要了解去重后的元素的数量，使用`elementSet().size()`。（所以，例如`add(E)`会对`multiset.size()`增加一。）
*   `multiset.iterator()`对每个元素的每个引用进行遍历，所以该迭代器的长度等于`multiset.size()`。
*   `Multiset<E>`支持添加元素，删除元素，或直接设定元素的计数值。`setCount(elem, 0)`相当于删除该元素的所有引用。
*   `multiset.count(elem)` 对不包含在 multiset 中的元素总会返回`0`。

### 实现类

Guava 提供了许多`Multiset`的实现，*大致*与 JDK 的实现相对应。

Map                 | Corresponding Multiset     | Supports `null` elements
:------------------ | :------------------------- | :-----------------------
`HashMap`           | [`HashMultiset`]           | Yes
`TreeMap`           | [`TreeMultiset`]           | Yes
`LinkedHashMap`     | [`LinkedHashMultiset`]     | Yes
`ConcurrentHashMap` | [`ConcurrentHashMultiset`] | No
`ImmutableMap`      | [`ImmutableMultiset`]      | No

### SortedMultiset

[`SortedMultiset`]是`Multiset`接口的一种变体，它能够支持按照特定范围来高效的取出 sub-multisets。例如你可以使用`latencies.subMultiset(0, BoundType.CLOSED, 100, BoundType.OPEN).size()`来测算你的网站中有多少点击延迟小于 100ms 并且与`latencies.size()`相比较来测算与点击总量的占比。

`TreeMultiset` 实现了 `SortedMultiset` 接口。在本文撰写时，`ImmutableSortedMultiset`对 GWT 的兼容性测试仍在进行中。

## Multimap

每一个有经验的 Java 程序员，在某些时刻，都尝试实现并处理过一种令人尴尬的结构： `Map<K, List<V>>` 或 `Map<K, Set<V>>`。例如 `Map<K, Set<V>>` 是一种典型的表示未标记有向图的方式。Guava 的 [`Multimap`]框架让处理一个 key 与多个 value 之间的映射变得简单。`Multimap` 是一种通用的关联单个 key 与任意多个 value 的方法。

理解 Multimap 的概念有两种办法：看作是单个 key 对单个 value 的映射的集合：

```
a -> 1
a -> 2
a -> 4
b -> 3
c -> 5
```

或者作为唯一 key 到 value 集合的映射：

```
a -> [1, 2, 4]
b -> [3]
c -> [5]
```

通常，`Multimap` 接口根据第一种视角来理解是最好的，但他也允许你按照另一种视角来看待，即以`asMap()`的视角，他会返回 `Map<K, Collection<V>>`。重要的是一个 key 映射一个空集合的形式并不存在：一个 key 映射到至少一个 value，否则他就不会在`Multimap`中存在。

很少会直接使用 `Multimap` 接口，反之，更多情况你会使用 `ListMultimap` 或 `SetMultimap`，对应分别将 key 映射为一个`List` 或一个 `Set`。

### 构建

创建一个 `Multimap` 最直接的方式就是使用 [`MultimapBuilder`]，它允许你配置 key 和 value 以怎样的形式展现。比如：

```java
// creates a ListMultimap with tree keys and array list values
ListMultimap<String, Integer> treeListMultimap =
    MultimapBuilder.treeKeys().arrayListValues().build();

// creates a SetMultimap with hash keys and enum set values
SetMultimap<Integer, MyEnum> hashEnumMultimap =
    MultimapBuilder.hashKeys().enumSetValues(MyEnum.class).build();
```

你也可以选择直接在实现类上使用 `create()` 方法，只不过这样相比于 `MultimapBuilder` 有一些不妥。

### 修改

[`Multimap.get(key)`] 返回与指定 key 关联的所有值的*视图*，即使当前并没有值。对 `ListMultimap` 他会返回一个 `List`，对 `SetMultimap` 他会返回一个 `Set`。

修改操作通过底层的 `Multimap` 来进行写。例如,

```java
Set<Person> aliceChildren = childrenMultimap.get(alice);
aliceChildren.clear();
aliceChildren.add(bob);
aliceChildren.add(carol);
```

通过底层 multimap 来写.

Other ways of modifying the multimap (more directly) include:
另外的（更直接的）修改 multimap 的方法包括：

Signature                         | Description                                                                                                                                                                                                           | Equivalent
:-------------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------
[`put(K, V)`]                     | 增加一个 key 和 value 的关联 | `multimap.get(key).add(value)`
[`putAll(K, Iterable<V>)`]        | 按顺序增加 key 与 value 的关联。 | `Iterables.addAll(multimap.get(key), values)`
[`remove(K, V)`]                  | 删除一个 `key` 与 `value` 的关联并在 multimap 被修改后返回 `true`. | `multimap.get(key).remove(value)`
[`removeAll(K)`] | 删除所有与 key 相关联的 value 并将这些 value 返回. 返回的集合可能可以也可能不可以被修改, 但不论怎样对其进行修改都不会影响原 multimap. (会返回何时的集合类型。) | `multimap.get(key).clear()`
[`replaceValues(K, Iterable<V>)`] | 删除所有与 `key` 关联的 `value` 并且将 `key` 与新的 `values` 关联. 返回先前与`key` 关联的所有值。| `multimap.get(key).clear(); Iterables.addAll(multimap.get(key), values)`

### 视图

`Multimap` 也支持多种强大的视图.

*   [`asMap`] 将任何 `Multimap<K, V>` 以 `Map<K, Collection<V>>` 的形式展示。返回的 map 支持 `remove`，修改会写回原对象，但他并不支持`put` 或 `putAll`。重要的是，当你期望对不存在的 key 返回 `null` 而不是空的可写集合时，你可以使用 `asMap().get(key)`。（你可以也应该将 `asMap.get(key)` 的返回值转换为适当的集合类型 -- 对 `SetMultimap` 返回 `Set`，对 `ListMultimap` 返回 `List` -- 然而类型系统不允许 `ListMultimap` 返回 `Map<K, List<V>>`。）
*   [`entries`] 以 `Collection<Map.Entry<K, V>>` 形式展示 `Multimap` 的所有 entries。（对 `SetMultimap` 会返回一个 `Set`。）
*   [`keySet`] 以的 `Set` 形式展示 `Multimap` 中所有不重复的 key。
*   [`keys`] 以 `Multiset` 的形式展示 `Multimap` 的所有 key，相同 key 的数量与对应的值的数量一致。允许从 `Multiset` 中删除元素，但不允许添加，相应的更改会被写入原对象。
*   [`values()`] 以“拍平”的`Collection<V>`形式展示 `Multimap` 所有的值，值都处于同一个集合中。这与 `Iterables.concat(multimap.asMap().values())` 很类似，只是返回了一个满的`Collection`。

### Multimap 不是 Map

`Multimap<K, V>` *不是* `Map<K, Collection<V>>`，尽管上述 map 可能被用于 `Multimap` 的实现。其显著的差异包括：

*   `Multimap.get(key)` 总是返回一个非 null，但可能为空的集合。这并不意味着 multimap 会耗费内存来与 key 做关联，相反，返回的集合是一个视图，能允许你根据需要添加与该 key 的关联。
*   假如你更喜欢当 multimap 中不存在 key 时返回 `null` 这种类似 `Map` 的行为，使用 `asMap()` 视图来获得一个`Map<K,Collection<V>>`，（或者使用静态方法 [`Multimaps.asMap()`] 来从`ListMultimap`中获取`Map<K,`**`List`**`<V>>`，类似的方法在`SetMultimap` 和 `SortedSetMultimap` 亦存在）
*   当且仅当对指定的 key 存在相关联的任意个元素时，`Multimap.containsKey(key)`返回 true。特别的，如果一个 key `k` 曾经与一个或多个元素关联，但这些元素已经被移除，则 `Multimap.containsKey(key)` 返回 false。
*   `Multimap.entries()` 返回 `Multimap` 中所有 key 的所有 entries。假如你想要一个 key-collection 的 entries，则使用 `asMap().entrySet()`。
*   `Multimap.size()` 返回整个 multimap 中所有 entries 的数量，而不是无重复的 key 的数量。想要获得无重复 key 的数量，使用 `Multimap.keySet().size()`。

### 实现类

`Multimap` 提供了大量的实现类。你可以在大多数情况下使用他来替代 `Map<K, Collection<V>>`。

Implementation             | Keys behave like... | Values behave like..
:------------------------- | :------------------ | :-------------------
[`ArrayListMultimap`]      | `HashMap`           | `ArrayList`
[`HashMultimap`]           | `HashMap`           | `HashSet`
[`LinkedListMultimap`] `*` | `LinkedHashMap``*`  | `LinkedList``*`
[`LinkedHashMultimap`]`**` | `LinkedHashMap`     | `LinkedHashSet`
[`TreeMultimap`]           | `TreeMap`           | `TreeSet`
[`ImmutableListMultimap`]  | `ImmutableMap`      | `ImmutableList`
[`ImmutableSetMultimap`]   | `ImmutableMap`      | `ImmutableSet`

除了不可变的类以外，每一个实现类都支持 null key 和 value。

`*` `LinkedListMultimap.entries()` 在重复的 key 和 value 之间保留迭代顺序。点击链接查看详情。

`**` `LinkedHashMultimap` 保留 entries 的插入顺序，也即是 key 的插入顺序，以及与任何一个 key 关联的 value 集。

注意并不是所有列出的实现的其本质都是 `Map<K, Collection<V>>`！（尤其是一些 `Multimap`的实现采用定制化的哈希表来使得开销最小。）

假如你想要更多的定制，使用[`Multimaps.newMultimap(Map, Supplier<Collection>)`]或 [list][newListMultimap] 以及 [set][newSetMultimap] 的变体或使用自定义集合，列表或集来实现支持 multimap。

## BiMap

将 value 映射回 key 的传统方法是维护两个独立的 map 并且使他们保持同步，然而，这种方式容易出错，且当一个 value 已经存在于 map 中时，会变得非常令人困惑。例如：

```java
Map<String, Integer> nameToId = Maps.newHashMap();
Map<Integer, String> idToName = Maps.newHashMap();

nameToId.put("Bob", 42);
idToName.put(42, "Bob");
// 假如 "Bob" 或 42 已经存在时会怎么样?
// 如果我们忘记使他们保持一致，那么可能会出现奇奇怪怪的 bug...
```

一个 [`BiMap<K, V>`] 是一个 `Map<K, V>`，它能够：

*   允许你通过[`inverse()`]来获得反向视图 `BiMap<V, K>`
*   确保 values 是唯一的，返回的[`values()`][BiMap.values] 是一个 `Set`

假如你尝试将 key 映射至一个已经存在的 value 时，`BiMap.put(key, value)`会抛出`IllegalArgumentException`。假如你期望删除具有指定 value 的任何已经存在的 entry，可以使用 [`BiMap.forcePut(key, value)`]

```java
BiMap<String, Integer> userId = HashBiMap.create();
...

String userForId = userId.inverse().get(id);
```

### 实现类

Key-Value Map Impl | Value-Key Map Impl | Corresponding `BiMap`
:----------------- | :----------------- | :--------------------
`HashMap`          | `HashMap`          | [`HashBiMap`]
`ImmutableMap`     | `ImmutableMap`     | [`ImmutableBiMap`]
`EnumMap`          | `EnumMap`          | [`EnumBiMap`]
`EnumMap`          | `HashMap`          | [`EnumHashBiMap`]

*注意：* `BiMap` 工具类例如 `synchronizedBiMap` 在 [`Maps`] 中存在.

## Table

```java
Table<Vertex, Vertex, Double> weightedGraph = HashBasedTable.create();
weightedGraph.put(v1, v2, 4);
weightedGraph.put(v1, v3, 20);
weightedGraph.put(v2, v3, 5);

weightedGraph.row(v1); // returns a Map mapping v2 to 4, v3 to 20
weightedGraph.column(v3); // returns a Map mapping v1 to 20, v2 to 5
```

通常，当你尝试同时对超过一个 key 进行索引时，你最终会选择类似 `Map<FirstName, Map<LastName, Person>>` 的难看而尴尬的解决方案。Guava 提供了一个新的集合类， [`Table`]，用于支持 "row" 类型和 "column" 类型的情况。`Table` 支持多种视图，来提供多角度的数据展示，包括：

*   [`rowMap()`], 将 `Table<R, C, V>` 展示为 `Map<R, Map<C, V>>`. 同样的, [`rowKeySet()`] 返回一个 `Set<R>`.
*   [`row(r)`] 返回一个非 null `Map<C, V>`. 写入这个 `Map` 写入底层的 `Table`.
*   类似的也提供了类方法： [`columnMap()`], [`columnKeySet()`], 和 [`column(c)`]. (基于列的查询会比基于行的效率更低一些)
*   [`cellSet()`] 的返回将 `Table` 展示为一组  [`Table.Cell<R, C, V>`]. `Cell` 就像 `Map.Entry`, 但区分了行和列的 key。

多种 `Table` 的实现类如下：

*   [`HashBasedTable`], 实际上底层是 `HashMap<R, HashMap<C,V>>`。
*   [`TreeBasedTable`], 实际上底层是 `TreeMap<R, TreeMap<C,V>>`。
*   [`ImmutableTable`]
*   [`ArrayTable`]，它要求在创建时即指定完整的行和列，当表密集时通过二维数组来提升速度与内存效率。`ArrayTable`与其他的实现上有些许不同，查看 javadoc 来获取详情。

## ClassToInstanceMap

有些时候，map 中的 key 不一定全都是同样的类型：他们*是各种*类型，且你想要让某种类型映射对应其类型的值。Guava 提供了 [`ClassToInstanceMap`] 来实现它。

除了扩展 `Map` 接口外，`ClassToInstanceMap` 提供了
[`T getInstance(Class<T>)`] 和 [`T putInstance(Class<T>, T)`] 方法，以此来避免在强制类型安全时不愉快的类型转换。

`ClassToInstanceMap` 有一个单一类型的参数，通常命名为 `B`，代表由 map 管理的类型上界。例如：

```java
ClassToInstanceMap<Number> numberDefaults = MutableClassToInstanceMap.create();
numberDefaults.putInstance(Integer.class, Integer.valueOf(0));
```

技术上讲，`ClassToInstanceMap<B>` 实现了 `Map<Class<? extends B>, B>` -- 或者换句话说，一个 B 的子类映射到 B 的一个实例。这让 `ClassToInstanceMap` 引入泛型变得有点混乱，但只要记住 `B` 总是 map 的上边界 -- 通常 `B` 只是 `Object`。

Guava 提供了名为 [`MutableClassToInstanceMap`] 和
[`ImmutableClassToInstanceMap`] 的实现类。

**重要**：就像其他的`Map<Class, Object>`一样，`ClassToInstanceMap` 也许会包含基本类型的 entries，因此基本类型和其对应的包装类也许会映射至不同的值。

## RangeSet

A `RangeSet` describes a set of *disconnected, nonempty* ranges. When adding a
range to a mutable `RangeSet`, any connected ranges are merged together, and
empty ranges are ignored. For example:
`RangeSet` 描述了一种存储 *非连续，非空* 范围的 set。当给不可变的 `RangeSet` 增加了一个范围时，任何连续的范围被合并，空范围被忽略。例如：

```java
   RangeSet<Integer> rangeSet = TreeRangeSet.create();
   rangeSet.add(Range.closed(1, 10)); // {[1, 10]}
   rangeSet.add(Range.closedOpen(11, 15)); // disconnected range: {[1, 10], [11, 15)}
   rangeSet.add(Range.closedOpen(15, 20)); // connected range; {[1, 10], [11, 20)}
   rangeSet.add(Range.openClosed(0, 0)); // empty range; {[1, 10], [11, 20)}
   rangeSet.remove(Range.open(5, 10)); // splits [1, 10]; {[1, 5], [10, 10], [11, 20)}
```

注意当合并类似`Range.closed(1, 10)` 和 `Range.closedOpen(11, 15)` 时，你需要通过 [`Range.canonical(DiscreteDomain)`] 来对范围进行预处理，例如 `DiscreteDomain.integers()`。

**NOTE**: `RangeSet` is not supported under GWT, nor in the JDK 1.5 backport;
`RangeSet` requires full use of the `NavigableMap` features in JDK 1.6.
**注意**：`RangeSet` 在 GWT 和 JDK 1.5 之前都不受支持，`RangeSet`依赖 JDK 1.6 的 `NavigableMap` 功能。

### 视图

`RangeSet` 的实现类支持非常广泛的范围视图，例如：

*   `complement()`: `RangeSet` 的补集视图。`complement` 也是一个 `RangeSet`，即包含非连续、非空的范围。
*   `subRangeSet(Range<C>)`: 返回 `RangeSet` 中指定 `Range` 的交叉视图。这可以概括为传统有序集合的`headSet`, `subSet`,  和 `tailSet` 视图。
*   `asRanges()`: 以可迭代的 `Set<Range<C>>` 形式展示 `RangeSet`。
*   `asSet(DiscreteDomain<C>)` (`ImmutableRangeSet` only): 以 `ImmutableSortedSet<C>` 的形式展示 `RangeSet<C>`，展示其中的元素，而不是范围本身。（这个操作不支持在 `DiscreteDomain` 和 `RangeSet` 都无上界或下界的情况）

### 查询

除了对视图的操作以外，`RangeSet` 还支持多种直接查询操作，其中最突出的有：

*   `contains(C)`: `RangeSet` 最基础的操作，查询 `RangeSet` 的任意范围中是否包含给定的元素。
*   `rangeContaining(C)`: 返回包围了指定元素的范围，若不存在则返回 `null`。
*   `encloses(Range<C>)`: 直截了当，测试 `RangeSet` 中是否存在任何 `Range` 包含了给定的范围。
*   `span()`: 返回一个最小的 `Range`，他能 `encloses` 所有 `RangeSet` 中的范围。

## RangeMap

`RangeMap` 是一个集合类型，描述一个不相交、非空的范围与值的映射。不像`RangeSet`，`RangeMap` 从不“合并”相邻的映射， 即使相邻的映射指向相同的值。例如：

```java
RangeMap<Integer, String> rangeMap = TreeRangeMap.create();
rangeMap.put(Range.closed(1, 10), "foo"); // {[1, 10] => "foo"}
rangeMap.put(Range.open(3, 6), "bar"); // {[1, 3] => "foo", (3, 6) => "bar", [6, 10] => "foo"}
rangeMap.put(Range.open(10, 20), "foo"); // {[1, 3] => "foo", (3, 6) => "bar", [6, 10] => "foo", (10, 20) => "foo"}
rangeMap.remove(Range.closed(5, 11)); // {[1, 3] => "foo", (3, 5) => "bar", (11, 20) => "foo"}
```

### 视图

`RangeMap` 提供两种视图:

*   `asMapOfRanges()`: 以 `Map<Range<K>, V>` 的形式展示 `RangeMap`。例如这可以用于遍历 `RangeMap`。
*   `subRangeMap(Range<K>)` 将 `RangeMap` 与指定的 `Range` 相交为一个 `RangeMap`。这可以概括为传统 `headMap`, `subMap`, 和 `tailMap`操作。

[`Multiset`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multiset.html
[`count(E)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multiset.html#count-java.lang.Object-
[`elementSet()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multiset.html#elementSet--
[`entrySet()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multiset.html#entrySet--
[`add(E, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multiset.html#add-java.lang.Object-int-
[`remove(E, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multiset.html#remove-java.lang.Object-int--
[`setCount(E, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multiset.html#setCount-E-int-
[`HashMultiset`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/HashMultiset.html
[`TreeMultiset`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/TreeMultiset.html
[`LinkedHashMultiset`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/LinkedHashMultiset.html
[`ConcurrentHashMultiset`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ConcurrentHashMultiset.html
[`ImmutableMultiset`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableMultiset.html
[`SortedMultiset`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/SortedMultiset.html
[`Multimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html
[`Multimap.get(key)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#get-K-
[`put(K, V)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#put-K-V-
[`putAll(K, Iterable<V>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#putAll-K-java.lang.Iterable-
[`remove(K, V)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#remove-java.lang.Object-java.lang.Object-
[`removeAll(K)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#removeAll-java.lang.Object-
[`replaceValues(K, Iterable<V>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#replaceValues-K-java.lang.Iterable-
[`asMap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#asMap--
[`entries`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#entries--
[`keySet`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#keySet--
[`keys`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#keys--
[`values()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimap.html#values--
[`MultimapBuilder`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/MultimapBuilder.html
[`Multimaps.asMap()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#asMap-com.google.common.collect.ListMultimap-
[`ArrayListMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ArrayListMultimap.html
[`HashMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/HashMultimap.html
[`LinkedListMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/LinkedListMultimap.html
[`LinkedHashMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/LinkedHashMultimap.html
[`TreeMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/TreeMultimap.html
[`ImmutableListMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableListMultimap.html
[`ImmutableSetMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableSetMultimap.html
[`Multimaps.newMultimap(Map, Supplier<Collection>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#newMultimap-java.util.Map-com.google.common.base.Supplier-
[newListMultimap]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#newListMultimap-java.util.Map-com.google.common.base.Supplier-
[newSetMultimap]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#newSetMultimap-java.util.Map-com.google.common.base.Supplier-
[`BiMap<K, V>`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/BiMap.html
[`inverse()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/BiMap.html#inverse--
[BiMap.values]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/BiMap.html#values--
[`BiMap.forcePut(key, value)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/BiMap.html#forcePut-java.lang.Object-java.lang.Object-
[`HashBiMap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/HashBiMap.html
[`ImmutableBiMap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableBiMap.html
[`EnumBiMap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/EnumBiMap.html
[`EnumHashBiMap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/EnumHashBiMap.html
[`Maps`]: CollectionUtilitiesExplained#Maps
[`Table`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.html
[`rowMap()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.html#rowMap--
[`rowKeySet()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.html#rowKeySet--
[`row(r)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.html#row-R-
[`columnMap()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.html#columnMap--
[`columnKeySet()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.html#columnKeySet--
[`column(c)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.html#column-C-
[`cellSet()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.html#cellSet--
[`Table.Cell<R, C, V>`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Table.Cell.html
[`HashBasedTable`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/HashBasedTable.html
[`TreeBasedTable`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/TreeBasedTable.html
[`ImmutableTable`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableTable.html
[`ArrayTable`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ArrayTable.html
[`ClassToInstanceMap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ClassToInstanceMap.html
[`T getInstance(Class<T>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ClassToInstanceMap.html#getInstance-java.lang.Class-
[`T putInstance(Class<T>, T)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ClassToInstanceMap.html#putInstance-java.lang.Class-java.lang.Object-
[`MutableClassToInstanceMap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/MutableClassToInstanceMap.html
[`ImmutableClassToInstanceMap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableClassToInstanceMap.html
[`Range.canonical(DiscreteDomain)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Range.html#canonical-com.google.common.collect.DiscreteDomain-
