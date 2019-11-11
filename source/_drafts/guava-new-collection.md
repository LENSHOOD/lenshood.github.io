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

### Views

`Multimap` also supports a number of powerful views.

*   [`asMap`] views any `Multimap<K, V>` as a `Map<K, Collection<V>>`. The
    returned map supports `remove`, and changes to the returned collections
    write through, but the map does not support `put` or `putAll`. Critically,
    you can use `asMap().get(key)` when you want `null` on absent keys rather
    than a fresh, writable empty collection. (You can and should cast
    `asMap.get(key)` to the appropriate collection type -- a `Set` for a
    `SetMultimap`, a `List` for a `ListMultimap` -- but the type system does not
    allow `ListMultimap` to return `Map<K, List<V>>` here.)
*   [`entries`] views the `Collection<Map.Entry<K, V>>` of all entries in the
    `Multimap`. (For a `SetMultimap`, this is a `Set`.)
*   [`keySet`] views the distinct keys in the `Multimap` as a `Set`.
*   [`keys`] views the keys of the `Multimap` as a `Multiset`, with multiplicity
    equal to the number of values associated to that key. Elements can be
    removed from the `Multiset`, but not added; changes will write through.
*   [`values()`] views all the values in the `Multimap` as a "flattened"
    `Collection<V>`, all as one collection. This is similar to
    `Iterables.concat(multimap.asMap().values())`, but returns a full
    `Collection` instead.

### Multimap Is Not A Map

A `Multimap<K, V>` is *not* a `Map<K, Collection<V>>`, though such a map might
be used in a `Multimap` implementation. Notable differences include:

*   `Multimap.get(key)` always returns a non-null, possibly empty collection.
    This doesn't imply that the multimap spends any memory associated with the
    key, but instead, the returned collection is a view that allows you to add
    associations with the key if you like.
*   If you prefer the more `Map`-like behavior of returning `null` for keys that
    aren't in the multimap, use the `asMap()` view to get a `Map<K,
    Collection<V>>`. (Or, to get a `Map<K,`**`List`**`<V>>` from a
    `ListMultimap`, use the static [`Multimaps.asMap()`] method. Similar methods
    exist for `SetMultimap` and `SortedSetMultimap`.)
*   `Multimap.containsKey(key)` is true if and only if there are any elements
    associated with the specified key. In particular, if a key `k` was
    previously associated with one or more values which have since been removed
    from the multimap, `Multimap.containsKey(k)` will return false.
*   `Multimap.entries()` returns all entries for all keys in the `Multimap`. If
    you want all key-collection entries, use `asMap().entrySet()`.
*   `Multimap.size()` returns the number of entries in the entire multimap, not
    the number of distinct keys. Use `Multimap.keySet().size()` instead to get
    the number of distinct keys.

### Implementations

`Multimap` provides a wide variety of implementations. You can use it in most
places you would have used a `Map<K, Collection<V>>`.

Implementation             | Keys behave like... | Values behave like..
:------------------------- | :------------------ | :-------------------
[`ArrayListMultimap`]      | `HashMap`           | `ArrayList`
[`HashMultimap`]           | `HashMap`           | `HashSet`
[`LinkedListMultimap`] `*` | `LinkedHashMap``*`  | `LinkedList``*`
[`LinkedHashMultimap`]`**` | `LinkedHashMap`     | `LinkedHashSet`
[`TreeMultimap`]           | `TreeMap`           | `TreeSet`
[`ImmutableListMultimap`]  | `ImmutableMap`      | `ImmutableList`
[`ImmutableSetMultimap`]   | `ImmutableMap`      | `ImmutableSet`

Each of these implementations, except the immutable ones, support null keys and
values.

`*` `LinkedListMultimap.entries()` preserves iteration order across non-distinct
key values. See the link for details.

`**` `LinkedHashMultimap` preserves insertion order of entries, as well as the
insertion order of keys, and the set of values associated with any one key.

Be aware that not all implementations are actually implemented as a `Map<K,
Collection<V>>` with the listed implementations! (In particular, several
`Multimap` implementations use custom hash tables to minimize overhead.)

If you need more customization, use [`Multimaps.newMultimap(Map,
Supplier<Collection>)`] or the [list][newListMultimap] and [set][newSetMultimap]
versions to use a custom collection, list, or set implementation to back your
multimap.

## BiMap

The traditional way to map values back to keys is to maintain two separate maps
and keep them both in sync, but this is bug-prone and can get extremely
confusing when a value is already present in the map. For example:

```java
Map<String, Integer> nameToId = Maps.newHashMap();
Map<Integer, String> idToName = Maps.newHashMap();

nameToId.put("Bob", 42);
idToName.put(42, "Bob");
// what happens if "Bob" or 42 are already present?
// weird bugs can arise if we forget to keep these in sync...
```

A [`BiMap<K, V>`] is a `Map<K, V>` that

*   allows you to view the "inverse" `BiMap<V, K>` with [`inverse()`]
*   ensures that values are unique, making [`values()`][BiMap.values] a `Set`

`BiMap.put(key, value)` will throw an `IllegalArgumentException` if you attempt
to map a key to an already-present value. If you wish to delete any preexisting
entry with the specified value, use [`BiMap.forcePut(key, value)`] instead.

```java
BiMap<String, Integer> userId = HashBiMap.create();
...

String userForId = userId.inverse().get(id);
```

### Implementations

Key-Value Map Impl | Value-Key Map Impl | Corresponding `BiMap`
:----------------- | :----------------- | :--------------------
`HashMap`          | `HashMap`          | [`HashBiMap`]
`ImmutableMap`     | `ImmutableMap`     | [`ImmutableBiMap`]
`EnumMap`          | `EnumMap`          | [`EnumBiMap`]
`EnumMap`          | `HashMap`          | [`EnumHashBiMap`]

*Note:* `BiMap` utilities like `synchronizedBiMap` live in [`Maps`].

## Table

```java
Table<Vertex, Vertex, Double> weightedGraph = HashBasedTable.create();
weightedGraph.put(v1, v2, 4);
weightedGraph.put(v1, v3, 20);
weightedGraph.put(v2, v3, 5);

weightedGraph.row(v1); // returns a Map mapping v2 to 4, v3 to 20
weightedGraph.column(v3); // returns a Map mapping v1 to 20, v2 to 5
```

Typically, when you are trying to index on more than one key at a time, you will
wind up with something like `Map<FirstName, Map<LastName, Person>>`, which is
ugly and awkward to use. Guava provides a new collection type, [`Table`], which
supports this use case for any "row" type and "column" type. `Table` supports a
number of views to let you use the data from any angle, including

*   [`rowMap()`], which views a `Table<R, C, V>` as a `Map<R, Map<C, V>>`.
    Similarly, [`rowKeySet()`] returns a `Set<R>`.
*   [`row(r)`] returns a non-null `Map<C, V>`. Writes to the `Map` will write
    through to the underlying `Table`.
*   Analogous column methods are provided: [`columnMap()`], [`columnKeySet()`],
    and [`column(c)`]. (Column-based access is somewhat less efficient than
    row-based access.)
*   [`cellSet()`] returns a view of the `Table` as a set of [`Table.Cell<R, C,
    V>`]. `Cell` is much like `Map.Entry`, but distinguishes the row and column
    keys.

Several `Table` implementations are provided, including:

*   [`HashBasedTable`], which is essentially backed by a `HashMap<R, HashMap<C,
    V>>`.
*   [`TreeBasedTable`], which is essentially backed by a `TreeMap<R, TreeMap<C,
    V>>`.
*   [`ImmutableTable`]
*   [`ArrayTable`], which requires that the complete universe of rows and
    columns be specified at construction time, but is backed by a
    two-dimensional array to improve speed and memory efficiency when the table
    is dense. `ArrayTable` works somewhat differently from other
    implementations; consult the Javadoc for details.

## ClassToInstanceMap

Sometimes, your map keys aren't all of the same type: they *are* types, and you
want to map them to values of that type. Guava provides [`ClassToInstanceMap`]
for this purpose.

In addition to extending the `Map` interface, `ClassToInstanceMap` provides the
methods [`T getInstance(Class<T>)`] and [`T putInstance(Class<T>, T)`], which
eliminate the need for unpleasant casting while enforcing type safety.

`ClassToInstanceMap` has a single type parameter, typically named `B`,
representing the upper bound on the types managed by the map. For example:

```java
ClassToInstanceMap<Number> numberDefaults = MutableClassToInstanceMap.create();
numberDefaults.putInstance(Integer.class, Integer.valueOf(0));
```

Technically, `ClassToInstanceMap<B>` implements `Map<Class<? extends B>, B>` --
or in other words, a map from subclasses of B to instances of B. This can make
the generic types involved in `ClassToInstanceMap` mildly confusing, but just
remember that `B` is always the upper bound on the types in the map -- usually,
`B` is just `Object`.

Guava provides implementations helpfully named [`MutableClassToInstanceMap`] and
[`ImmutableClassToInstanceMap`].

**Important**: Like any other `Map<Class, Object>`, a `ClassToInstanceMap` may
contain entries for primitive types, and a primitive type and its corresponding
wrapper type may map to different values.

## RangeSet

A `RangeSet` describes a set of *disconnected, nonempty* ranges. When adding a
range to a mutable `RangeSet`, any connected ranges are merged together, and
empty ranges are ignored. For example:

```java
   RangeSet<Integer> rangeSet = TreeRangeSet.create();
   rangeSet.add(Range.closed(1, 10)); // {[1, 10]}
   rangeSet.add(Range.closedOpen(11, 15)); // disconnected range: {[1, 10], [11, 15)}
   rangeSet.add(Range.closedOpen(15, 20)); // connected range; {[1, 10], [11, 20)}
   rangeSet.add(Range.openClosed(0, 0)); // empty range; {[1, 10], [11, 20)}
   rangeSet.remove(Range.open(5, 10)); // splits [1, 10]; {[1, 5], [10, 10], [11, 20)}
```

Note that to merge ranges like `Range.closed(1, 10)` and `Range.closedOpen(11,
15)`, you must first preprocess ranges with [`Range.canonical(DiscreteDomain)`],
e.g. with `DiscreteDomain.integers()`.

**NOTE**: `RangeSet` is not supported under GWT, nor in the JDK 1.5 backport;
`RangeSet` requires full use of the `NavigableMap` features in JDK 1.6.

### Views

`RangeSet` implementations support an extremely wide range of views, including:

*   `complement()`: views the complement of the `RangeSet`. `complement` is also
    a `RangeSet`, as it contains disconnected, nonempty ranges.
*   `subRangeSet(Range<C>)`: returns a view of the intersection of the
    `RangeSet` with the specified `Range`. This generalizes the `headSet`,
    `subSet`, and `tailSet` views of traditional sorted collections.
*   `asRanges()`: views the `RangeSet` as a `Set<Range<C>>` which can be
    iterated over.
*   `asSet(DiscreteDomain<C>)` (`ImmutableRangeSet` only): Views the
    `RangeSet<C>` as an `ImmutableSortedSet<C>`, viewing the elements in the
    ranges instead of the ranges themselves. (This operation is unsupported if
    the `DiscreteDomain` and the `RangeSet` are both unbounded above or both
    unbounded below.)

### Queries

In addition to operations on its views, `RangeSet` supports several query
operations directly, the most prominent of which are:

*   `contains(C)`: the most fundamental operation on a `RangeSet`, querying if
    any range in the `RangeSet` contains the specified element.
*   `rangeContaining(C)`: returns the `Range` which encloses the specified
    element, or `null` if there is none.
*   `encloses(Range<C>)`: straightforwardly enough, tests if any `Range` in the
    `RangeSet` encloses the specified range.
*   `span()`: returns the minimal `Range` that `encloses` every range in this
    `RangeSet`.

## RangeMap

`RangeMap` is a collection type describing a mapping from disjoint, nonempty
ranges to values. Unlike `RangeSet`, `RangeMap` never "coalesces" adjacent
mappings, even if adjacent ranges are mapped to the same values. For example:

```java
RangeMap<Integer, String> rangeMap = TreeRangeMap.create();
rangeMap.put(Range.closed(1, 10), "foo"); // {[1, 10] => "foo"}
rangeMap.put(Range.open(3, 6), "bar"); // {[1, 3] => "foo", (3, 6) => "bar", [6, 10] => "foo"}
rangeMap.put(Range.open(10, 20), "foo"); // {[1, 3] => "foo", (3, 6) => "bar", [6, 10] => "foo", (10, 20) => "foo"}
rangeMap.remove(Range.closed(5, 11)); // {[1, 3] => "foo", (3, 5) => "bar", (11, 20) => "foo"}
```

### Views

`RangeMap` provides two views:

*   `asMapOfRanges()`: views the `RangeMap` as a `Map<Range<K>, V>`. This can be
    used, for example, to iterate over the `RangeMap`.
*   `subRangeMap(Range<K>)` views the intersection of the `RangeMap` with the
    specified `Range` as a `RangeMap`. This generalizes the traditional
    `headMap`, `subMap`, and `tailMap` operations.

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
