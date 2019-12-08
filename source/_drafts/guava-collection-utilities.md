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

As a result, many of the operations you might expect to see supported for all
collections can be found in [`Iterables`]. Additionally, most `Iterables`
methods have a corresponding version in [`Iterators`] that accepts the raw
iterator.

The overwhelming majority of operations in the `Iterables` class are *lazy*:
they only advance the backing iteration when absolutely necessary. Methods that
themselves return `Iterables` return lazily computed views, rather than
explicitly constructing a collection in memory.

As of Guava 12, `Iterables` is supplemented by the [`FluentIterable`] class,
which wraps an `Iterable` and provides a "fluent" syntax for many of these
operations.

The following is a selection of the most commonly used utilities, although many
of the more "functional" methods in `Iterables` are discussed in [Guava
functional idioms][functional].

### General

Method                                | Description                                                                                            | See Also
:------------------------------------ | :----------------------------------------------------------------------------------------------------- | :-------
[`concat(Iterable<Iterable>)`]        | Returns a lazy view of the concatenation of several iterables.                                         | [`concat(Iterable...)`]
[`frequency(Iterable, Object)`]       | Returns the number of occurrences of the object.                                                       | Compare `Collections.frequency(Collection, Object)`; see [`Multiset`]
[`partition(Iterable, int)`]          | Returns an unmodifiable view of the iterable partitioned into chunks of the specified size.            | [`Lists.partition(List, int)`], [`paddedPartition(Iterable, int)`]
[`getFirst(Iterable, T default)`]     | Returns the first element of the iterable, or the default value if empty.                              | Compare `Iterable.iterator().next()`, [`FluentIterable.first()`]
[`getLast(Iterable)`]                 | Returns the last element of the iterable, or fails fast with a `NoSuchElementException` if it's empty. | [`getLast(Iterable, T default)`], [`FluentIterable.last()`]
[`elementsEqual(Iterable, Iterable)`] | Returns true if the iterables have the same elements in the same order.                                | Compare `List.equals(Object)`
[`unmodifiableIterable(Iterable)`]    | Returns an unmodifiable view of the iterable.                                                          | Compare `Collections.unmodifiableCollection(Collection)`
[`limit(Iterable, int)`]              | Returns an `Iterable` returning at most the specified number of elements.                              | [`FluentIterable.limit(int)`]
[`getOnlyElement(Iterable)`]          | Returns the only element in `Iterable`. Fails fast if the iterable is empty or has multiple elements.  | [`getOnlyElement(Iterable, T default)`]

```java
Iterable<Integer> concatenated = Iterables.concat(
  Ints.asList(1, 2, 3),
  Ints.asList(4, 5, 6));
// concatenated has elements 1, 2, 3, 4, 5, 6

String lastAdded = Iterables.getLast(myLinkedHashSet);

String theElement = Iterables.getOnlyElement(thisSetIsDefinitelyASingleton);
  // if this set isn't a singleton, something is wrong!
```

### Collection-Like

Typically, collections support these operations naturally on other collections,
but not on iterables.

*Each of these operations delegates to the corresponding `Collection` interface
method when the input is actually a `Collection`.* For example, if
`Iterables.size` is passed a `Collection`, it will call the `Collection.size`
method instead of walking through the iterator.

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

Besides the methods covered above and in the functional idioms [article]
[functional], `FluentIterable` has a few convenient methods for copying
into an immutable collection:

Result Type          | Method
:------------------- | :-----------------------------------
`ImmutableList`      | [`toImmutableList()`]
`ImmutableSet`       | [`toImmutableSet()`]
`ImmutableSortedSet` | [`toImmutableSortedSet(Comparator)`]

### Lists

In addition to static constructor methods and functional programming methods,
[`Lists`] provides a number of valuable utility methods on `List` objects.

Method                   | Description
:----------------------- | :----------
[`partition(List, int)`] | Returns a view of the underlying list, partitioned into chunks of the specified size.
[`reverse(List)`]        | Returns a reversed view of the specified list. *Note*: if the list is immutable, consider [`ImmutableList.reverse()`] instead.

```java
List<Integer> countUp = Ints.asList(1, 2, 3, 4, 5);
List<Integer> countDown = Lists.reverse(theList); // {5, 4, 3, 2, 1}

List<List<Integer>> parts = Lists.partition(countUp, 2); // {{1, 2}, {3, 4}, {5}}
```

### Static Factories

`Lists` provides the following static factory methods:

Implementation | Factories
:------------- | :--------
`ArrayList`    | [basic][newArrayList], [with elements][newArrayList(E...)], [from `Iterable`][newArrayList(Iterable)], [with exact capacity][newArrayListWithCapacity], [with expected size][newArrayListWithExpectedSize], [from `Iterator`][newArrayList(Iterator)]
`LinkedList`   | [basic][newLinkedList], [from `Iterable`][newLinkedList(Iterable)]

## Sets

The [`Sets`] utility class includes a number of spicy methods.

### Set-Theoretic Operations

We provide a number of standard set-theoretic operations, implemented as views
over the argument sets. These return a [`SetView`], which can be used:

*   as a `Set` directly, since it implements the `Set` interface
*   by copying it into another mutable collection with [`copyInto(Set)`]
*   by making an immutable copy with [`immutableCopy()`]

Method                            |
:-------------------------------- |
[`union(Set, Set)`]               |
[`intersection(Set, Set)`]        |
[`difference(Set, Set)`]          |
[`symmetricDifference(Set, Set)`] |

For example:

```java
Set<String> wordsWithPrimeLength = ImmutableSet.of("one", "two", "three", "six", "seven", "eight");
Set<String> primes = ImmutableSet.of("two", "three", "five", "seven");

SetView<String> intersection = Sets.intersection(primes, wordsWithPrimeLength); // contains "two", "three", "seven"
// I can use intersection as a Set directly, but copying it can be more efficient if I use it a lot.
return intersection.immutableCopy();
```

### Other Set Utilities

Method                          | Description                                                                             | See Also
:------------------------------ | :-------------------------------------------------------------------------------------- | :-------
[`cartesianProduct(List<Set>)`] | Returns every possible list that can be obtained by choosing one element from each set. | [`cartesianProduct(Set...)`]
[`powerSet(Set)`]               | Returns the set of subsets of the specified set.                                        |

```java
Set<String> animals = ImmutableSet.of("gerbil", "hamster");
Set<String> fruits = ImmutableSet.of("apple", "orange", "banana");

Set<List<String>> product = Sets.cartesianProduct(animals, fruits);
// {{"gerbil", "apple"}, {"gerbil", "orange"}, {"gerbil", "banana"},
//  {"hamster", "apple"}, {"hamster", "orange"}, {"hamster", "banana"}}

Set<Set<String>> animalSets = Sets.powerSet(animals);
// {{}, {"gerbil"}, {"hamster"}, {"gerbil", "hamster"}}
```

### Static Factories

`Sets` provides the following static factory methods:

Implementation  | Factories
:-------------- | :--------
`HashSet`       | [basic][newHashSet], [with elements][newHashSet(E...)], [from `Iterable`][newHashSet(Iterable)], [with expected size][newHashSetWithExpectedSize], [from `Iterator`][newHashSet(Iterator)]
`LinkedHashSet` | [basic][newLinkedHashSet], [from `Iterable`][newLinkedHashSet(Iterable)], [with expected size][newLinkedHashSetWithExpectedSize]
`TreeSet`       | [basic][newTreeSet], [with `Comparator`][newTreeSet(Comparator)], [from `Iterable`][newTreeSet(Iterable)]

## Maps

[`Maps`] has a number of cool utilities that deserve individual explanation.

### `uniqueIndex`

[`Maps.uniqueIndex(Iterable, Function)`] addresses the common case of having a
bunch of objects that each have some unique attribute, and wanting to be able to
look up those objects based on that attribute.

Let's say we have a bunch of strings that we know have unique lengths, and we
want to be able to look up the string with some particular length.

```java
ImmutableMap<Integer, String> stringsByIndex = Maps.uniqueIndex(strings, new Function<String, Integer> () {
    public Integer apply(String string) {
      return string.length();
    }
  });
```

If indices are *not* unique, see `Multimaps.index` below.

### `difference`

[`Maps.difference(Map, Map)`] allows you to compare all the differences between
two maps. It returns a `MapDifference` object, which breaks down the Venn
diagram into:

Method                   | Description
:----------------------- | :----------
[`entriesInCommon()`]    | The entries which are in both maps, with both matching keys and values.
[`entriesDiffering()`]   | The entries with the same keys, but differing values. The values in this map are of type [`MapDifference.ValueDifference`], which lets you look at the left and right values.
[`entriesOnlyOnLeft()`]  | Returns the entries whose keys are in the left but not in the right map.
[`entriesOnlyOnRight()`] | Returns the entries whose keys are in the right but not in the left map.

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

The Guava utilities on `BiMap` live in the `Maps` class, since a `BiMap` is also
a `Map`.

`BiMap` utility              | Corresponding `Map` utility
:--------------------------- | :---------------------------------
[`synchronizedBiMap(BiMap)`] | `Collections.synchronizedMap(Map)`
[`unmodifiableBiMap(BiMap)`] | `Collections.unmodifiableMap(Map)`

#### Static Factories

`Maps` provides the following static factory methods.

Implementation    | Factories
:---------------- | :--------
`HashMap`         | [basic][newHashMap], [from `Map`][newHashMap(Map)], [with expected size][newHashMapWithExpectedSize]
`LinkedHashMap`   | [basic][newLinkedHashMap], [from `Map`][newLinkedHashMap(Map)]
`TreeMap`         | [basic][newTreeMap], [from `Comparator`][newTreeMap(Comparator)], [from `SortedMap`][newTreeMap(SortedMap)]
`EnumMap`         | [from `Class`][newEnumMap(Class)], [from `Map`][newEnumMap(Map)]
`ConcurrentMap`   | [basic][newConcurrentMap]
`IdentityHashMap` | [basic][newIdentityHashMap]

## Multisets

Standard `Collection` operations, such as `containsAll`, ignore the count of
elements in the multiset, and only care about whether elements are in the
multiset at all, or not. [`Multisets`] provides a number of operations that take
into account element multiplicities in multisets.

Method                                                        | Explanation                                                                                               | Difference from `Collection` method
:------------------------------------------------------------ | :-------------------------------------------------------------------------------------------------------- | :----------------------------------
[`containsOccurrences(Multiset sup, Multiset sub)`]           | Returns `true` if `sub.count(o) <= super.count(o)` for all `o`.                                           | `Collection.containsAll` ignores counts, and only tests whether elements are contained at all.
[`removeOccurrences(Multiset removeFrom, Multiset toRemove)`] | Removes one occurrence in `removeFrom` for each occurrence of an element in `toRemove`.                   | `Collection.removeAll` removes all occurences of any element that occurs even once in `toRemove`.
[`retainOccurrences(Multiset removeFrom, Multiset toRetain)`] | Guarantees that `removeFrom.count(o) <= toRetain.count(o)` for all `o`.                                   | `Collection.retainAll` keeps all occurrences of elements that occur even once in `toRetain`.
[`intersection(Multiset, Multiset)`]                          | Returns a view of the intersection of two multisets; a nondestructive alternative to `retainOccurrences`. | Has no analogue.

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

Other utilities in `Multisets` include:

Method                                         | Description
:--------------------------------------------- | :----------
[`copyHighestCountFirst(Multiset)`]            | Returns an immutable copy of the multiset that iterates over elements in descending frequency order.
[`unmodifiableMultiset(Multiset)`]             | Returns an unmodifiable view of the multiset.
[`unmodifiableSortedMultiset(SortedMultiset)`] | Returns an unmodifiable view of the sorted multiset.

```java
Multiset<String> multiset = HashMultiset.create();
multiset.add("a", 3);
multiset.add("b", 5);
multiset.add("c", 1);

ImmutableMultiset<String> highestCountFirst = Multisets.copyHighestCountFirst(multiset);

// highestCountFirst, like its entrySet and elementSet, iterates over the elements in order {"b", "a", "c"}
```

## Multimaps

[`Multimaps`] provides a number of general utility operations that deserve
individual explanation.

### `index`

The cousin to `Maps.uniqueIndex`, [`Multimaps.index(Iterable, Function)`]
answers the case when you want to be able to look up all objects with some
particular attribute in common, which is not necessarily unique.

Let's say we want to group strings based on their length.

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

Since `Multimap` can map many keys to one value, and one key to many values, it
can be useful to invert a `Multimap`. Guava provides [`invertFrom(Multimap
toInvert, Multimap dest)`] to let you do this, without choosing an
implementation for you.

*NOTE:* If you are using an `ImmutableMultimap`, consider
[`ImmutableMultimap.inverse()`] instead.

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

Need to use a `Multimap` method on a `Map`? [`forMap(Map)`] views a `Map` as a
`SetMultimap`. This is particularly useful, for example, in combination with
`Multimaps.invertFrom`.

```java
Map<String, Integer> map = ImmutableMap.of("a", 1, "b", 1, "c", 2);
SetMultimap<String, Integer> multimap = Multimaps.forMap(map);
// multimap maps ["a" => {1}, "b" => {1}, "c" => {2}]
Multimap<Integer, String> inverse = Multimaps.invertFrom(multimap, HashMultimap.<Integer, String> create());
// inverse maps [1 => {"a", "b"}, 2 => {"c"}]
```

### Wrappers

`Multimaps` provides the traditional wrapper methods, as well as tools to get
custom `Multimap` implementations based on `Map` and `Collection`
implementations of your choice.

Multimap type       | Unmodifiable                      | Synchronized                      | Custom
:------------------ | :-------------------------------- | :-------------------------------- | :-----
`Multimap`          | [`unmodifiableMultimap`]          | [`synchronizedMultimap`]          | [`newMultimap`]
`ListMultimap`      | [`unmodifiableListMultimap`]      | [`synchronizedListMultimap`]      | [`newListMultimap`]
`SetMultimap`       | [`unmodifiableSetMultimap`]       | [`synchronizedSetMultimap`]       | [`newSetMultimap`]
`SortedSetMultimap` | [`unmodifiableSortedSetMultimap`] | [`synchronizedSortedSetMultimap`] | [`newSortedSetMultimap`]

The custom `Multimap` implementations let you specify a particular
implementation that should be used in the returned `Multimap`. Caveats include:

*   The multimap assumes complete ownership over of map and the lists returned
    by factory. Those objects should not be manually updated, they should be
    empty when provided, and they should not use soft, weak, or phantom
    references.
*   **No guarantees are made** on what the contents of the `Map` will look like
    after you modify the `Multimap`.
*   The multimap is not threadsafe when any concurrent operations update the
    multimap, even if map and the instances generated by factory are. Concurrent
    read operations will work correctly, though. Work around this with the
    `synchronized` wrappers if necessary.
*   The multimap is serializable if map, factory, the lists generated by
    factory, and the multimap contents are all serializable.
*   The collections returned by `Multimap.get(key)` are *not* of the same type
    as the collections returned by your `Supplier`, though if you supplier
    returns `RandomAccess` lists, the lists returned by `Multimap.get(key)` will
    also be random access.

Note that the custom `Multimap` methods expect a `Supplier` argument to generate
fresh new collections. Here is an example of writing a `ListMultimap` backed by
a `TreeMap` mapping to `LinkedList`.

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

The [`Tables`] class provides a few handy utilities.

### `customTable`

Comparable to the `Multimaps.newXXXMultimap(Map, Supplier)` utilities,
[`Tables.newCustomTable(Map, Supplier<Map>)`] allows you to specify a `Table`
implementation using whatever row or column map you like.

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

The [`transpose(Table<R, C, V>)`] method allows you to view a `Table<R, C, V>`
as a `Table<C, R, V>`.

### Wrappers

These are the familiar unmodifiability wrappers you know and love. Consider,
however, using [`ImmutableTable`] instead in most cases.

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

