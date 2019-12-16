---
title: （Guava 译文系列）集合工具类
date: 2019-12-07 23:16:30
tags:
- guava
- translation
categories:
- Guava
---

# 集合工具类

任何体验过 JDK 集合框架的人都了解且喜爱[`java.util.Collections`]中的工具。
Guava 在这方面提供了更多工具：应用于所有集合的静态方法。
而且这些是 Guava 最流行也最成熟的部分。

与特定 interface 相关联的方法以相对直观的规则来分组：

Interface    | JDK or Guava? | Corresponding Guava utility class
:----------- | :------------ | :--------------------------------
`Collection` | JDK           | [`Collections2`]
`List`       | JDK           | [`Lists`]
`Set`        | JDK           | [`Sets`]
`SortedSet`  | JDK           | [`Sets`]
`Map`        | JDK           | [`Maps`]
`SortedMap`  | JDK           | [`Maps`]
`Queue`      | JDK           | [`Queues`]
[`Multiset`] | Guava         | [`Multisets`]
[`Multimap`] | Guava         | [`Multimaps`]
[`BiMap`]    | Guava         | [`Maps`]
[`Table`]    | Guava         | [`Tables`]

***在寻找变换、过滤等工具吗? 它们在我们的功能性语法下的
[functional] 文章里.***

## 静态构造器

在 JDK7 以前，构造新的泛型集合需要难看的重复代码：

```java
List<TypeThatsTooLongForItsOwnGood> list = new ArrayList<TypeThatsTooLongForItsOwnGood>();
```

我们应该都同意这挺难看。Guava 提供了在右边推断使用的泛型的静态方法：

```java
List<TypeThatsTooLongForItsOwnGood> list = Lists.newArrayList();
Map<KeyType, LongishValueType> map = Maps.newLinkedHashMap();
```

诚然，JDK7 的钻石操作符减少了一些麻烦：

```java
List<TypeThatsTooLongForItsOwnGood> list = new ArrayList<>();
```

但是 Guava 走的更远。通过工厂模式方法，我们能方便的在初始化集合时设置初始元素。

```java
Set<Type> copySet = Sets.newHashSet(elements);
List<String> theseElements = Lists.newArrayList("alpha", "beta", "gamma");
```

此外，通过对工厂方法命名的能力（Effective Java 第一条），我们能提升初始化指定大小集合的代码可读性：

```java
List<Type> exactly100 = Lists.newArrayListWithCapacity(100);
List<Type> approx100 = Lists.newArrayListWithExpectedSize(100);
Set<Type> approx100Set = Sets.newHashSetWithExpectedSize(100);
```

下面列出了提供的准确的静态工厂方法及其对应的工具类。

*注意：* Guava 提供的新集合类型并未暴露原始构造器，或在工具类内提供初始化方法。相反，他们直接暴露了静态工厂方法，例如：

```java
Multiset<String> multiset = HashMultiset.create();
```

## Iterables

任何可能的情况下，Guava 都更倾向于提供接受 `Iterable` 而不是 `Collection` 的工具。在 Google， 遇到某个并不存储在内存中的 “集合” 的情况并不少见，它的数据可能是从数据库，其他数据中心等处聚集而来的，因此实际上在并没有获取到所有元素的情况下，它并不能支持类似`size()` 等的操作。

因此，很多你期望支持所有 collections 的操作，都能在[`Iterables`]里面找到。此外，大多`Iterables` 方法都在 [`Iterators`] 中有对应的版本来接受一个 iterator。

`Iterables` 类中绝大多数的操作都是“懒操作”：除非绝对必要时才会提前准备迭代。返回`Iterables`的方法全都都返回懒计算的视图，而不会显式的在内存中构建 collection。

截止至 Guava 12，`Iterables` 通过一个包装了 `Iterable`且提供流式运算符的类 [`FluentIterable`] 来辅助。

以下是被选出的最常用的工具，尽管在`Iterables`中许多更函数式的方法会在[Guava 函数式语法][functional]中详述。

### 通用

Method                                | Description                                                                                            | See Also
:------------------------------------ | :----------------------------------------------------------------------------------------------------- | :-------
[`concat(Iterable<Iterable>)`]        | 返回懒视图的多个 iterable 的连接。                                       | [`concat(Iterable...)`]
[`frequency(Iterable, Object)`]       | 返回一个对象在 iterable 中出现的次数。                                                       | Compare `Collections.frequency(Collection, Object)`; see [`Multiset`]
[`partition(Iterable, int)`]          | 返回将 iterable 划分为指定块的不可变视图。           | [`Lists.partition(List, int)`], [`paddedPartition(Iterable, int)`]
[`getFirst(Iterable, T default)`]     | 返回 iterable 中的第一个元素，或当为空时返回默认值。                              | Compare `Iterable.iterator().next()`, [`FluentIterable.first()`]
[`getLast(Iterable)`]                 | 返回 iterable 中的最后一个元素，或当为空时快速失败并抛出 `NoSuchElementException` | [`getLast(Iterable, T default)`], [`FluentIterable.last()`]
[`elementsEqual(Iterable, Iterable)`] | 当两个 iterable 中的元素和顺序都一致时，返回 true。                                | Compare `List.equals(Object)`
[`unmodifiableIterable(Iterable)`]    | 返回给定 iterable 的一个不可变视图。                                                          | Compare `Collections.unmodifiableCollection(Collection)`
[`limit(Iterable, int)`]              | 返回一个 `Iterable` 并包含指定大小的元素。                              | [`FluentIterable.limit(int)`]
[`getOnlyElement(Iterable)`]          | 返回 `Iterable` 中的唯一元素. 当为空或存在多个元素时快速失败。  | [`getOnlyElement(Iterable, T default)`]

```java
Iterable<Integer> concatenated = Iterables.concat(
  Ints.asList(1, 2, 3),
  Ints.asList(4, 5, 6));
// concatenated has elements 1, 2, 3, 4, 5, 6

String lastAdded = Iterables.getLast(myLinkedHashSet);

String theElement = Iterables.getOnlyElement(thisSetIsDefinitelyASingleton);
  // if this set isn't a singleton, something is wrong!
```

### 类 collection

通常，collection 在其他 collection 上支持此类操作，但 iterable 不支持。

*当输入是一个 collection 时，每一个操作都交给对应的 `Collection` 接口方法委托。* 例如，如果 `Iterables.size` 传递给了 `Collection`，那么它实际上会调用 `Collection.size` 而不是遍历整个 iterator。

Method                                                  | Analogous `Collection` method      | `FluentIterable` equivalent
:------------------------------------------------------ | :--------------------------------- | :--------------------------
[`addAll(Collection addTo, Iterable toAdd)`]            | `Collection.addAll(Collection)`    |
[`contains(Iterable, Object)`]                          | `Collection.contains(Object)`      | [`FluentIterable.contains(Object)`]
[`removeAll(Iterable removeFrom, Collection toRemove)`] | `Collection.removeAll(Collection)` |
[`retainAll(Iterable removeFrom, Collection toRetain)`] | `Collection.retainAll(Collection)` |
[`size(Iterable)`]                                      | `Collection.size()`                | [`FluentIterable.size()`]
[`toArray(Iterable, Class)`]                            | `Collection.toArray(T[])`          | [`FluentIterable.toArray(Class)`]
[`isEmpty(Iterable)`]                                   | `Collection.isEmpty()`             | [`FluentIterable.isEmpty()`]
[`get(Iterable, int)`]                                  | `List.get(int)`                    | [`FluentIterable.get(int)`]
[`toString(Iterable)`]                                  | `Collection.toString()`            | [`FluentIterable.toString()`]

### FluentIterable

除了上述方法和[函数式语法][functional]，`FluentIterable` 包含一些方便的方法来复制为一个不可变集合：

Result Type          | Method
:------------------- | :-----------------------------------
`ImmutableList`      | [`toImmutableList()`]
`ImmutableSet`       | [`toImmutableSet()`]
`ImmutableSortedSet` | [`toImmutableSortedSet(Comparator)`]

### Lists

除了静态构造器方法和函数式编程方法，[`Lists`]提供了大量针对`List`对象的值工具方法。

Method                   | Description
:----------------------- | :----------
[`partition(List, int)`] | 返回底层列表的视图，该视图被分成指定大小的块。
[`reverse(List)`]        | 返回指定列表的反转. *注意*: 如果给定列表是不可变的, 考虑使用 [`ImmutableList.reverse()`] 来代替.

```java
List<Integer> countUp = Ints.asList(1, 2, 3, 4, 5);
List<Integer> countDown = Lists.reverse(theList); // {5, 4, 3, 2, 1}

List<List<Integer>> parts = Lists.partition(countUp, 2); // {{1, 2}, {3, 4}, {5}}
```

### 静态工厂

`Lists` 提供了下述静态工厂方法:

Implementation | Factories
:------------- | :--------
`ArrayList`    | [basic][newArrayList], [with elements][newArrayList(E...)], [from `Iterable`][newArrayList(Iterable)], [with exact capacity][newArrayListWithCapacity], [with expected size][newArrayListWithExpectedSize], [from `Iterator`][newArrayList(Iterator)]
`LinkedList`   | [basic][newLinkedList], [from `Iterable`][newLinkedList(Iterable)]

## Sets

[`Sets`] 类包含了许多厉害的方法.

### 集合论操作

我们提供了大量集合论操作, 作为参数集合上的视图实现. 这些返回[`SetView`]的方法可以用于:

*   直接用作 `Set`，因为他实现了 `Set` 接口
*   通过[`copyInto(Set)`]将他复制为另一个可变集合
*   通过 [`immutableCopy()`] 创建一个不可变副本

Method                            |
:-------------------------------- |
[`union(Set, Set)`]               |
[`intersection(Set, Set)`]        |
[`difference(Set, Set)`]          |
[`symmetricDifference(Set, Set)`] |

例如：

```java
Set<String> wordsWithPrimeLength = ImmutableSet.of("one", "two", "three", "six", "seven", "eight");
Set<String> primes = ImmutableSet.of("two", "three", "five", "seven");

SetView<String> intersection = Sets.intersection(primes, wordsWithPrimeLength); // contains "two", "three", "seven"
// I can use intersection as a Set directly, but copying it can be more efficient if I use it a lot.
return intersection.immutableCopy();
```

### 其他集合工具

Method                          | Description                                                                             | See Also
:------------------------------ | :-------------------------------------------------------------------------------------- | :-------
[`cartesianProduct(List<Set>)`] | 返回从每个集合中取一个元素可以得到的每个可能的列表. | [`cartesianProduct(Set...)`]
[`powerSet(Set)`]               | 返回指定集合的子集集合.                                        |

```java
Set<String> animals = ImmutableSet.of("gerbil", "hamster");
Set<String> fruits = ImmutableSet.of("apple", "orange", "banana");

Set<List<String>> product = Sets.cartesianProduct(animals, fruits);
// {{"gerbil", "apple"}, {"gerbil", "orange"}, {"gerbil", "banana"},
//  {"hamster", "apple"}, {"hamster", "orange"}, {"hamster", "banana"}}

Set<Set<String>> animalSets = Sets.powerSet(animals);
// {{}, {"gerbil"}, {"hamster"}, {"gerbil", "hamster"}}
```

### 静态工厂

`Sets` 提供了以下静态工厂方法:

Implementation  | Factories
:-------------- | :--------
`HashSet`       | [basic][newHashSet], [with elements][newHashSet(E...)], [from `Iterable`][newHashSet(Iterable)], [with expected size][newHashSetWithExpectedSize], [from `Iterator`][newHashSet(Iterator)]
`LinkedHashSet` | [basic][newLinkedHashSet], [from `Iterable`][newLinkedHashSet(Iterable)], [with expected size][newLinkedHashSetWithExpectedSize]
`TreeSet`       | [basic][newTreeSet], [with `Comparator`][newTreeSet(Comparator)], [from `Iterable`][newTreeSet(Iterable)]

## Maps

[`Maps`] 包含了许多有用的工具，值得我们单独讨论.

### `uniqueIndex`

[`Maps.uniqueIndex(Iterable, Function)`] 处理了一种常见的情况：存在许多对象，每一个都有一些自己特定的属性, 期望基于这些属性来查找这些对象.

比如说我们有很多字符串，我们知道他们有独一无二的长度，我们期望查找特定长度的字符串。

```java
ImmutableMap<Integer, String> stringsByIndex = Maps.uniqueIndex(strings, new Function<String, Integer> () {
    public Integer apply(String string) {
      return string.length();
    }
  });
```

假如索引*不*唯一, 可以参见 `Multimaps.index`.

### `difference`

[`Maps.difference(Map, Map)`] 允许你比较两个 Map 之间的不同. 他返回一个 `MapDifference` 对象, 他把文氏图分解为:

Method                   | Description
:----------------------- | :----------
[`entriesInCommon()`]    | 在两个 map 中都存在的 entries, key 和 value 都匹配.
[`entriesDiffering()`]   | 相同 key 但不同 value 的 entries. 在该 map 中的值属于 [`MapDifference.ValueDifference`] 类型, 能让你查找左值和右值.
[`entriesOnlyOnLeft()`]  | 返回 key 在左边 map 出现，但未出现在右边 map 的 entries。
[`entriesOnlyOnRight()`] | 返回 key 在右边 map 出现，但未出现在左边 map 的 entries。

```java
Map<String, Integer> left = ImmutableMap.of("a", 1, "b", 2, "c", 3);
Map<String, Integer> right = ImmutableMap.of("b", 2, "c", 4, "d", 5);
MapDifference<String, Integer> diff = Maps.difference(left, right);

diff.entriesInCommon(); // {"b" => 2}
diff.entriesDiffering(); // {"c" => (3, 4)}
diff.entriesOnlyOnLeft(); // {"a" => 1}
diff.entriesOnlyOnRight(); // {"d" => 5}
```

### `BiMap` utilities

Guava `BiMap` 的工具基于 `Maps`, 因此 `BiMap` 也同样是一个 `Map`.

`BiMap` utility              | Corresponding `Map` utility
:--------------------------- | :---------------------------------
[`synchronizedBiMap(BiMap)`] | `Collections.synchronizedMap(Map)`
[`unmodifiableBiMap(BiMap)`] | `Collections.unmodifiableMap(Map)`

#### 静态工厂

`Maps` 提供了如下静态工厂方法.

Implementation    | Factories
:---------------- | :--------
`HashMap`         | [basic][newHashMap], [from `Map`][newHashMap(Map)], [with expected size][newHashMapWithExpectedSize]
`LinkedHashMap`   | [basic][newLinkedHashMap], [from `Map`][newLinkedHashMap(Map)]
`TreeMap`         | [basic][newTreeMap], [from `Comparator`][newTreeMap(Comparator)], [from `SortedMap`][newTreeMap(SortedMap)]
`EnumMap`         | [from `Class`][newEnumMap(Class)], [from `Map`][newEnumMap(Map)]
`ConcurrentMap`   | [basic][newConcurrentMap]
`IdentityHashMap` | [basic][newIdentityHashMap]

## Multisets

标准`Collection`操作，例如`containsAll`，忽略了 multiset 中的元素计数，只关心元素是否存在于 multiset。[`Multisets`] 提供了大量考虑到 multiset 中元素多样性的操作。

Method                                                        | Explanation                                                                                               | Difference from `Collection` method
:------------------------------------------------------------ | :-------------------------------------------------------------------------------------------------------- | :----------------------------------
[`containsOccurrences(Multiset sup, Multiset sub)`]           | 返回 `true` 假如对于所有 `o`， `sub.count(o) <= super.count(o)`.                                           | `Collection.containsAll` 忽略了数量, 只测试元素是否被包含.
[`removeOccurrences(Multiset removeFrom, Multiset toRemove)`] | 将 `removeFrom` 中出现的元素，在 `toRemove` 中移除.                   | `Collection.removeAll` 移除 `toRemove` 中的所有元素，哪怕只出现了一次.
[`retainOccurrences(Multiset removeFrom, Multiset toRetain)`] | 确保对所有的`o`， `removeFrom.count(o) <= toRetain.count(o)`.                                   | `Collection.retainAll` 保留 `toRetain` 中出现的所有元素，哪怕只出现了一次.
[`intersection(Multiset, Multiset)`]                          | 返回两个 multisets 中交叉的部分; `retainOccurrences` 的一个无副作用替代品. | 没有类似对比.

```java
Multiset<String> multiset1 = HashMultiset.create();
multiset1.add("a", 2);

Multiset<String> multiset2 = HashMultiset.create();
multiset2.add("a", 5);

multiset1.containsAll(multiset2); // returns true: all unique elements are contained,
  // even though multiset1.count("a") == 2 < multiset2.count("a") == 5
Multisets.containsOccurrences(multiset1, multiset2); // returns false

multiset2.removeOccurrences(multiset1); // multiset2 now contains 3 occurrences of "a"

multiset2.removeAll(multiset1); // removes all occurrences of "a" from multiset2, even though multiset1.count("a") == 2
multiset2.isEmpty(); // returns true
```

其他的 `Multisets` 工具包括:

Method                                         | Description
:--------------------------------------------- | :----------
[`copyHighestCountFirst(Multiset)`]            | 返回 multiset 的一个不可变副本，按照元素出现频率降序迭代。
[`unmodifiableMultiset(Multiset)`]             | 返回 multiset 的一个不可变视图。
[`unmodifiableSortedMultiset(SortedMultiset)`] | 返回有序 multiset 的一个不可变视图。

```java
Multiset<String> multiset = HashMultiset.create();
multiset.add("a", 3);
multiset.add("b", 5);
multiset.add("c", 1);

ImmutableMultiset<String> highestCountFirst = Multisets.copyHighestCountFirst(multiset);

// highestCountFirst, like its entrySet and elementSet, iterates over the elements in order {"b", "a", "c"}
```

## Multimaps

[`Multimaps`] 提供了大量通用工具操作，值得我们一一讨论.

### `index`

`Maps.uniqueIndex`的表亲, [`Multimaps.index(Iterable, Function)`] 回答了你想要从所有对象中查找某些特定的共有而不是排他的属性的问题。

比方说我们想要通过字符串长度来对其进行分组。

```java
ImmutableSet<String> digits = ImmutableSet.of(
    "zero", "one", "two", "three", "four",
    "five", "six", "seven", "eight", "nine");
Function<String, Integer> lengthFunction = new Function<String, Integer>() {
  public Integer apply(String string) {
    return string.length();
  }
};
ImmutableListMultimap<Integer, String> digitsByLength = Multimaps.index(digits, lengthFunction);
/*
 * digitsByLength maps:
 *  3 => {"one", "two", "six"}
 *  4 => {"zero", "four", "five", "nine"}
 *  5 => {"three", "seven", "eight"}
 */
```

### `invertFrom`

由于 `Multimap` 可以将多个 key 映射至单个 value, 也可以将一个 key 映射至 多个 value, 那么转置一个 `Multimap` 就变得很有用处。 Guava 提供了 [`invertFrom(Multimap toInvert, Multimap dest)`] 让你来做这件事，而不是帮你选择一个实现。

*注意:* 如果你是用的是 `ImmutableMultimap`, 考虑替换为 [`ImmutableMultimap.inverse()`]。

```java
ArrayListMultimap<String, Integer> multimap = ArrayListMultimap.create();
multimap.putAll("b", Ints.asList(2, 4, 6));
multimap.putAll("a", Ints.asList(4, 2, 1));
multimap.putAll("c", Ints.asList(2, 5, 3));

TreeMultimap<Integer, String> inverse = Multimaps.invertFrom(multimap, TreeMultimap.<String, Integer> create());
// note that we choose the implementation, so if we use a TreeMultimap, we get results in order
/*
 * inverse maps:
 *  1 => {"a"}
 *  2 => {"a", "b", "c"}
 *  3 => {"c"}
 *  4 => {"a", "b"}
 *  5 => {"c"}
 *  6 => {"b"}
 */
```

### `forMap`

想在`Map`上使用`Multimap`的方法？[`forMap(Map)`] 将 `Map` 展示为 `SetMultimap` 视图. 这非常有用, 比如, 与 `Multimaps.invertFrom`结合在一起。

```java
Map<String, Integer> map = ImmutableMap.of("a", 1, "b", 1, "c", 2);
SetMultimap<String, Integer> multimap = Multimaps.forMap(map);
// multimap maps ["a" => {1}, "b" => {1}, "c" => {2}]
Multimap<Integer, String> inverse = Multimaps.invertFrom(multimap, HashMultimap.<Integer, String> create());
// inverse maps [1 => {"a", "b"}, 2 => {"c"}]
```

### Wrappers
`Multimaps` 提供了传统的包装方法, 以及 `Map` 和`Collection`实现的定制化 `Multimap` 工具。

Multimap type       | Unmodifiable                      | Synchronized                      | Custom
:------------------ | :-------------------------------- | :-------------------------------- | :-----
`Multimap`          | [`unmodifiableMultimap`]          | [`synchronizedMultimap`]          | [`newMultimap`]
`ListMultimap`      | [`unmodifiableListMultimap`]      | [`synchronizedListMultimap`]      | [`newListMultimap`]
`SetMultimap`       | [`unmodifiableSetMultimap`]       | [`synchronizedSetMultimap`]       | [`newSetMultimap`]
`SortedSetMultimap` | [`unmodifiableSortedSetMultimap`] | [`synchronizedSortedSetMultimap`] | [`newSortedSetMultimap`]

定制的 `Multimap` 实现让你能选定某个特定实现作用于返回的 `Multimap` 上. 请注意:

*   multimap 假设完全拥有通过工厂方法创建的 map 和 list。 这些对象不应该被手动更新, 他们在被提供时应该为空, 且他们并不应该被作为软引用、弱引用、幻象引用。
*   **不做保证** 在你修改`Multimap`后，`Map` 也应该一同改变。
*   multimap 在并发修改时并不是线程安全的, 即使构建 map 和其实例的工厂方法确保他们线程安全，即使并发读操作能正常工作。如需必要，使 `synchronized` 包装器来绕过它。
*   假如map, 工厂, 工厂创建的 list, 以及 multimap 的内容都可序列化时，该 multimap 可被序列化。
*   `Multimap.get(key)`返回的集合与`Supplier`返回的集合类型*不同*，但如果 supplier 返回`RandomAccess` list，`Multimap.get(key)`返回的list 也将是random access。

注意定制的`Multimap` 方法接受一个 `Supplier` 参数来创建全新的 collection。 下面是一个例子展示一个基于`TreeMap`的`ListMultimap` 映射为`LinkedList`。

```java
ListMultimap<String, Integer> myMultimap = Multimaps.newListMultimap(
  Maps.<String, Collection<Integer>>newTreeMap(),
  new Supplier<LinkedList<Integer>>() {
    public LinkedList<Integer> get() {
      return Lists.newLinkedList();
    }
  });
```

## Tables

[`Tables`] 类提供了一些趁手的工具.

### `customTable`

与`Multimaps.newXXXMultimap(Map, Supplier)` 工具相比,
[`Tables.newCustomTable(Map, Supplier<Map>)`] 允许你指定一个 `Table`
实现， 使用任意的行列映射。

```java
// use LinkedHashMaps instead of HashMaps
Table<String, Character, Integer> table = Tables.newCustomTable(
  Maps.<String, Map<Character, Integer>>newLinkedHashMap(),
  new Supplier<Map<Character, Integer>> () {
    public Map<Character, Integer> get() {
      return Maps.newLinkedHashMap();
    }
  });
```

### `transpose`

[`transpose(Table<R, C, V>)`] 方法允许你将`Table<C, R, V>` 展示为 `Table<R, C, V>`视图。

### Wrappers

我们提供了你熟悉且喜爱的不可变包装器. 但你可以考虑在大多数情况下使用 [`ImmutableTable`]。

*   [`unmodifiableTable`]
*   [`unmodifiableRowSortedTable`]

[`java.util.Collections`]: http://docs.oracle.com/javase/7/docs/api/java/util/Collections.html
[`Collections2`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Collections2.html
[`Lists`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html
[`Sets`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html
[`Maps`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html
[`Queues`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Queues.html
[`Multiset`]: NewCollectionTypesExplained#Multiset
[`Multisets`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multisets.html
[`Multimap`]: NewCollectionTypesExplained#Multimap
[`Multimaps`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html
[`BiMap`]: NewCollectionTypesExplained#BiMap
[`Table`]: NewCollectionTypesExplained#Table
[`Tables`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Tables.html
[functional]: FunctionalExplained
[`Iterables`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html
[`Iterators`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterators.html
[`FluentIterable`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/FluentIterable.html
[`concat(Iterable<Iterable>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#concat-java.lang.Iterable-
[`concat(Iterable...)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#concat-java.lang.Iterable...-
[`frequency(Iterable, Object)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#frequency-java.lang.Iterable-java.lang.Object-
[`partition(Iterable, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#partition-java.lang.Iterable-int-
[`Lists.partition(List, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#partition-java.util.List-int-
[`paddedPartition(Iterable, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#paddedPartition-java.lang.Iterable-int-
[`getFirst(Iterable, T default)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#getFirst-java.lang.Iterable-T-
[`FluentIterable.first()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/FluentIterable.html#first--
[`getLast(Iterable)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#getLast-java.lang.Iterable-
[`getLast(Iterable, T default)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#getLast-java.lang.Iterable-T-
[`FluentIterable.last()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/FluentIterable.html#last--
[`elementsEqual(Iterable, Iterable)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#elementsEqual-java.lang.Iterable-java.lang.Iterable-
[`unmodifiableIterable(Iterable)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#unmodifiableIterable-java.lang.Iterable-
[`limit(Iterable, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#limit-java.lang.Iterable-int-
[`FluentIterable.limit(int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/FluentIterable.html#limit-int-
[`getOnlyElement(Iterable)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#getOnlyElement-java.lang.Iterable-
[`getOnlyElement(Iterable, T default)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#getOnlyElement-java.lang.Iterable-T-
[`addAll(Collection addTo, Iterable toAdd)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#addAll-java.util.Collection-java.lang.Iterable-
[`contains(Iterable, Object)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#contains-java.lang.Iterable-java.lang.Object-
[`FluentIterable.contains(Object)`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#contains-java.lang.Object-
[`removeAll(Iterable removeFrom, Collection toRemove)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#removeAll-java.lang.Iterable-java.util.Collection-
[`retainAll(Iterable removeFrom, Collection toRetain)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#retainAll-java.lang.Iterable-java.util.Collection-
[`size(Iterable)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#size-java.lang.Iterable-
[`FluentIterable.size()`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#size--
[`toArray(Iterable, Class)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#toArray-java.lang.Iterable-java.lang.Class-
[`FluentIterable.toArray(Class)`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#toArray-java.lang.Class-
[`isEmpty(Iterable)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#isEmpty-java.lang.Iterable-
[`FluentIterable.isEmpty()`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#isEmpty--
[`get(Iterable, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#get-java.lang.Iterable-int-
[`FluentIterable.get(int)`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#get-int-
[`toString(Iterable)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Iterables.html#toString-java.lang.Iterable-
[`FluentIterable.toString()`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#toString--
[`toImmutableList()`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#toImmutableList--
[`toImmutableSet()`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#toImmutableSet--
[`toImmutableSortedSet(Comparator)`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/FluentIterable.html#toImmutableSortedSet-java.util.Comparator-
[`partition(List, int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#partition-java.util.List-int-
[`reverse(List)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#reverse-java.util.List-
[`ImmutableList.reverse()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableList.html#reverse--
[newArrayList]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#newArrayList--
[newArrayList(E...)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#newArrayList-E...-
[newArrayList(Iterable)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#newArrayList-java.lang.Iterable-
[newArrayListWithCapacity]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#newArrayListWithCapacity-int-
[newArrayListWithExpectedSize]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#newArrayListWithExpectedSize-int-
[newArrayList(Iterator)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#newArrayList-java.util.Iterator-
[newLinkedList]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#newLinkedList--
[newLinkedList(Iterable)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Lists.html#newLinkedList-java.lang.Iterable-
[`SetView`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.SetView.html
[`copyInto(Set)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.SetView.html#copyInto-S-
[`immutableCopy()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.SetView.html#immutableCopy--
[`union(Set, Set)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#union-java.util.Set-java.util.Set-
[`intersection(Set, Set)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#intersection-java.util.Set-java.util.Set-
[`difference(Set, Set)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#difference-java.util.Set-java.util.Set-
[`symmetricDifference(Set, Set)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#symmetricDifference-java.util.Set-java.util.Set-
[`cartesianProduct(List<Set>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#cartesianProduct-java.util.List-
[`cartesianProduct(Set...)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#cartesianProduct-java.util.Set...-
[`powerSet(Set)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#powerSet-java.util.Set-
[newHashSet]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newHashSet--
[newHashSet(E...)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newHashSet-E...-
[newHashSet(Iterable)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newHashSet-java.lang.Iterable-
[newHashSetWithExpectedSize]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newHashSetWithExpectedSize-int-
[newHashSet(Iterator)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newHashSet-java.util.Iterator-
[newLinkedHashSet]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newLinkedHashSet--
[newLinkedHashSet(Iterable)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newLinkedHashSet-java.lang.Iterable-
[newLinkedHashSetWithExpectedSize]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newLinkedHashSetWithExpectedSize-int-
[newTreeSet]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newTreeSet--
[newTreeSet(Comparator)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newTreeSet-java.util.Comparator-
[newTreeSet(Iterable)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Sets.html#newTreeSet-java.lang.Iterable-
[`Maps.uniqueIndex(Iterable, Function)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#uniqueIndex-java.lang.Iterable-com.google.common.base.Function-
[`Maps.difference(Map, Map)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#difference-java.util.Map-java.util.Map-
[`entriesInCommon()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/MapDifference.html#entriesInCommon--
[`entriesDiffering()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/MapDifference.html#entriesDiffering--
[`MapDifference.ValueDifference`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/MapDifference.ValueDifference.html
[`entriesOnlyOnLeft()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/MapDifference.html#entriesOnlyOnLeft--
[`entriesOnlyOnRight()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/MapDifference.html#entriesOnlyOnRight--
[`synchronizedBiMap(BiMap)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#synchronizedBiMap-com.google.common.collect.BiMap-
[`unmodifiableBiMap(BiMap)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#unmodifiableBiMap-com.google.common.collect.BiMap-
[newHashMap]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newHashMap--
[newHashMap(Map)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newHashMap-java.util.Map-
[newHashMapWithExpectedSize]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newHashMapWithExpectedSize-int-
[newLinkedHashMap]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newLinkedHashMap--
[newLinkedHashMap(Map)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newLinkedHashMap-java.util.Map-
[newTreeMap]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newTreeMap--
[newTreeMap(Comparator)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newTreeMap-java.util.Comparator-
[newTreeMap(SortedMap)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newTreeMap-java.util.SortedMap-
[newEnumMap(Class)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newEnumMap-java.lang.Class-
[newEnumMap(Map)]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newEnumMap-java.util.Map-
[newConcurrentMap]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newConcurrentMap--
[newIdentityHashMap]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Maps.html#newIdentityHashMap--
[`containsOccurrences(Multiset sup, Multiset sub)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multisets.html#containsOccurrences-com.google.common.collect.Multiset-com.google.common.collect.Multiset-
[`removeOccurrences(Multiset removeFrom, Multiset toRemove)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multisets.html#removeOccurrences-com.google.common.collect.Multiset-com.google.common.collect.Multiset-
[`retainOccurrences(Multiset removeFrom, Multiset toRetain)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multisets.html#retainOccurrences-com.google.common.collect.Multiset-com.google.common.collect.Multiset-
[`intersection(Multiset, Multiset)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multisets.html#intersection-com.google.common.collect.Multiset-com.google.common.collect.Multiset-
[`copyHighestCountFirst(Multiset)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multisets.html#copyHighestCountFirst-com.google.common.collect.Multiset-
[`unmodifiableMultiset(Multiset)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multisets.html#unmodifiableMultiset-com.google.common.collect.Multiset-
[`unmodifiableSortedMultiset(SortedMultiset)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multisets.html#unmodifiableSortedMultiset-com.google.common.collect.SortedMultiset-
[`Multimaps.index(Iterable, Function)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#index-java.lang.Iterable-com.google.common.base.Function-
[`invertFrom(Multimap toInvert, Multimap dest)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#invertFrom-com.google.common.collect.Multimap-M-
[`ImmutableMultimap.inverse()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableMultimap.html#inverse--
[`forMap(Map)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#forMap-java.util.Map-
[`unmodifiableMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#unmodifiableMultimap-com.google.common.collect.Multimap-
[`unmodifiableListMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#unmodifiableListMultimap-com.google.common.collect.ListMultimap-
[`unmodifiableSetMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#unmodifiableSetMultimap-com.google.common.collect.SetMultimap-
[`unmodifiableSortedSetMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#unmodifiableSortedSetMultimap-com.google.common.collect.SortedSetMultimap-
[`synchronizedMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#synchronizedMultimap-com.google.common.collect.Multimap-
[`synchronizedListMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#synchronizedListMultimap-com.google.common.collect.ListMultimap-
[`synchronizedSetMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#synchronizedSetMultimap-com.google.common.collect.SetMultimap-
[`synchronizedSortedSetMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#synchronizedSortedSetMultimap-com.google.common.collect.SortedSetMultimap-
[`newMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#newMultimap-java.util.Map-com.google.common.base.Supplier-
[`newListMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#newListMultimap-java.util.Map-com.google.common.base.Supplier-
[`newSetMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#newSetMultimap-java.util.Map-com.google.common.base.Supplier-
[`newSortedSetMultimap`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Multimaps.html#newSortedSetMultimap-java.util.Map-com.google.common.base.Supplier-
[`Tables.newCustomTable(Map, Supplier<Map>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Tables.html#newCustomTable-java.util.Map-com.google.common.base.Supplier-
[`transpose(Table<R, C, V>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Tables.html#transpose-com.google.common.collect.Table-
[`ImmutableTable`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableTable.html
[`unmodifiableTable`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Tables.html#unmodifiableTable-com.google.common.collect.Table-
[`unmodifiableRowSortedTable`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Tables.html#unmodifiableRowSortedTable-com.google.common.collect.RowSortedTable-

