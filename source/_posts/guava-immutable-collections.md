---
title: （Guava 译文系列）不可变集合
date: 2019-09-25 23:17:43
tags:
- guava
- translation
categories:
- Guava
---

## 不可变集合

### 示例
```java
public static final ImmutableSet<String> COLOR_NAMES = ImmutableSet.of(
  "red",
  "orange",
  "yellow",
  "green",
  "blue",
  "purple");

class Foo {
  final ImmutableSet<Bar> bars;
  Foo(Set<Bar> bars) {
    this.bars = ImmutableSet.copyOf(bars); // defensive copy!
  }
}
```

### 为何要这么做？
不可变对象有非常多的优点，包括：
- 安全的被不受信任的库使用。
- 线程安全：能够被许多线程使用且不存在竞态条件的风险。
- 不需要支持变更，基于此假设，能够节省时间和空间。所有的不可变集合的实现都比他们可变集合兄弟的内存效率要高。（[分析](https://github.com/DimitrisAndreou/memory-measurer/blob/master/ElementCostInDataStructures.txt)）
- 可以当做常量使用，能够期望他会保持固定不变

为对象创建不可变副本是一项优秀的防御式编程技术。Guava 为标准`Collection`类型提供了简单、易于使用的不可变版本，包括 Guava 自己的`Collection` 变体。

JDK 提供了`Collections.unmodifiableXXX`方法，但在我们看来，它：
- 笨重且冗长；在任何你想要做防御副本的地方使用都令人不快
- 不安全：只有当没有人持有原集合的引用时，才能会返回真正的不可变集合
- 低效：数据结构仍然包含了可变集合的所有开销，包括并发修改检查，哈希表所需的额外空间等等

**当你不期望修改一个集合，或期望一个集合保持不变时，一个不错的实践是将之防御性的拷贝入一个不可变的集合。**

**重要**：所有 Guava 的不可变集合实现都*拒绝 null*  值。基于我们对 Google 代码库的彻底研究表明，只有 5% 的情况下，`null`被集合所接受，其余 95% 的情况都受益于对`null`快速失败。假如你需要使用`null`，那就考虑使用`Collections.unmodifiableList`或类似的允许`null`值的实现吧。

### 如何使用？
一个`ImmutableXXX`集合可以通过以下几种方式创建：
- 使用`copyOf`方法，如`ImmutableSet.copyOf(set)`
- 使用`of`方法，如`ImmutableSet.of("a", "b", "c")`或`ImmutableMap.of("a", 1, "b", 2)`
- 使用`Builder`，如
```java
public static final ImmutableSet<Color> GOOGLE_COLORS =
   ImmutableSet.<Color>builder()
       .addAll(WEBSAFE_COLORS)
       .add(new Color(0, 191, 255))
       .build();
```
除了已经排序的集合外，**集合顺序会在构造时被确定**。如
`ImmutableSet.of("a", "b", "c", "a", "d", "b")`
的元素会以"a","b","c","d"的顺序被遍历。

#### `copyOf`比你想象的更聪明
很有用的一点是，请记住`ImmutableXXX.copyOf`会在安全的前提下尽量避免复制数据 -- 具体细节未明，但实现通常很“智能”。例如：
```java
ImmutableSet<String> foobar = ImmutableSet.of("foo", "bar", "baz");
thingamajig(foobar);

void thingamajig(Collection<String> collection) {
   ImmutableList<String> defensiveCopy = ImmutableList.copyOf(collection);
   ...
}
```
在上述代码中，`ImmutableList.copyOf(foobar)`会足够聪明的直接返回`foobar.asList()`，即`ImmutableSet`的常量耗时视图。

作为一般性的启发，`ImmutableXXX.copyOf(ImmutableCollection)`尝试在下述时刻避免进行线性耗时拷贝：
- 有可能在常量时间内使用底层数据结构时。如`ImmutableSet.copyOf(ImmutableList)`则无法在常量时间内完成。
- 不会造成内存泄漏时。例如，假定你有一个`ImmutableList<String> hugeList`，然后你执行了`ImmutableList.copyOf(hugeList.subList(0, 10))`，则显示复制会被执行，这能够避免意外的持有`hugeList`的引用，这显然并无必要。
- 不会改变语义时。因此`ImmutableSet.copyOf(myImmutableSortedSet)`会执行显示拷贝，因为`ImmutableSet`所持有的`hashCode()`和`equals()`与基于比较行为的`ImmutableSortedSet`的语义并不相同。

这能帮助最小化防御式编程风格所造成的性能开销。

#### `asList`
所有不可变集合都能通过`asList()`来提供`ImmutableList`的视图，所以，例如你有一个`ImmutableSortedSet`，则你可以通过`sortedSet.asList().get(k)`来获取第`k`小的元素。

返回的`ImmutableList`通常 -- 不总是但通常 -- 是一个常量开销的视图，而不是显示拷贝的。这就是说，它通常比普通的`List`更聪明，例如他会用原有集合的更高效的`contains`方法。

### 详细
#### 在哪儿？

| Interface                                                    | JDK or Guava? | Immutable Version                                            |
| ------------------------------------------------------------ | ------------- | ------------------------------------------------------------ |
| `Collection`                                                 | JDK           | [`ImmutableCollection`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableCollection.html) |
| `List`                                                       | JDK           | [`ImmutableList`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableList.html) |
| `Set`                                                        | JDK           | [`ImmutableSet`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableSet.html) |
| `SortedSet/NavigableSet`                                     | JDK           | [`ImmutableSortedSet`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableSortedSet.html) |
| `Map`                                                        | JDK           | [`ImmutableMap`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableMap.html) |
| `SortedMap`                                                  | JDK           | [`ImmutableSortedMap`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableSortedMap.html) |
| [`Multiset`](https://github.com/google/guava/wiki/NewCollectionTypesExplained#Multiset) | Guava         | [`ImmutableMultiset`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableMultiset.html) |
| `SortedMultiset`                                             | Guava         | [`ImmutableSortedMultiset`](http://google.github.io/guava/releases/12.0/api/docs/com/google/common/collect/ImmutableSortedMultiset.html) |
| [`Multimap`](https://github.com/google/guava/wiki/NewCollectionTypesExplained#Multimap) | Guava         | [`ImmutableMultimap`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableMultimap.html) |
| `ListMultimap`                                               | Guava         | [`ImmutableListMultimap`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableListMultimap.html) |
| `SetMultimap`                                                | Guava         | [`ImmutableSetMultimap`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableSetMultimap.html) |
| [`BiMap`](https://github.com/google/guava/wiki/NewCollectionTypesExplained#BiMap) | Guava         | [`ImmutableBiMap`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableBiMap.html) |
| [`ClassToInstanceMap`](https://github.com/google/guava/wiki/NewCollectionTypesExplained#ClassToInstanceMap) | Guava         | [`ImmutableClassToInstanceMap`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableClassToInstanceMap.html) |
| [`Table`](https://github.com/google/guava/wiki/NewCollectionTypesExplained#Table) | Guava         | [`ImmutableTable`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ImmutableTable.html) |
