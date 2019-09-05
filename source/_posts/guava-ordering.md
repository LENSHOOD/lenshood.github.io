---
title: 排序
date: 2019-09-05 21:34:32
tags:
- guava
- translation
categories:
- Guava
---
## 排序
### 示例
```java
assertTrue(byLengthOrdering.reverse().isOrdered(list));
```
### 简述
[`Ordering`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html)是 Guava 的“流式”`Comparator`类。它能够用于构建复杂的比较器，并且能应用于集合对象上。

一个`Ordering`实例的核心仅仅是一个特殊的`Comparator`实例。`Ordering`简单的取用一些依赖`Comparator`的静态方法（例如 `Collections.max`）将之改造为实例方法。此外，`Ordering`类还提供了链式方法来改进、增强现有的比较器。

### 创建
通用的排序实例可由静态方法提供：

方法 | 描述
---|---
[`natural`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#natural--) | 使用*正常序列*的 Comparable 类型
[`usingToString`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#usingToString--) | 以字典序对对象的字符串表示（由`toString()`返回）进行排序

将一个现存的`Comparator`来构造`Ordering`，最简单的方法是使用[`Ordering.from(Comparator)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#from-java.util.Comparator-)。

但是更通用的构造一个自定义`Ordering`的方法是完全跳过`Comparator`而直接扩展`Ordering`的抽象类：

```java
Ordering<String> byLengthOrdering = new Ordering<String>() {
  public int compare(String left, String right) {
    return Ints.compare(left.length(), right.length());
  }
};
```
### 链式调用
一个给定的`Ordering`可以被包装并获取派生的`Ordering`。如下是一些最常用的变体：

方法 | 描述
---|---
[`reverse()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#reverse--) | 返回反向排序
[`nullsFirst()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#nullsFirst--) | 返回一个新的`Ordering`，会将 null 对象至于 non- null 对象之前，如果没有 null 对象，则表现为与原始`Ordering`的行为一致。类似的可见[`nullsLast`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#nullsLast--)
[`compound(Comparator)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#compound-java.util.Comparator-) | 返回一个专用的`Ordering`，当遇到相等情况时可进行进一步比较
[`lexicographical()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#lexicographical--) | 返回可对 iterables 的元素按照字典序排序的`Ordering`
[`onResultOf(Function)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#onResultOf-com.google.common.base.Function-) | 返回一个`Ordering`，对被比较值先执行 Function，对返回值按初始`Ordering`进行排序

举例说明，假如你想要构造一个下述类的比较器：
```java
class Foo {
  @Nullable String sortedBy;
  int notSortedBy;
}
```
该比较器能处理可为 null 的`sortBy`成员。以下是一个建立在链式调用方法上的解决方案：
```java
Ordering<Foo> ordering = Ordering.natural().nullsFirst().onResultOf(new Function<Foo, String>() {
  public String apply(Foo foo) {
    return foo.sortedBy;
  }
});
```
当链式调用的`Ordering`进行读取时，按从右至左倒序工作。以上例子通过读取`sortedBy`成员变量值来对`Foo`进行排序，首先将所有为 null 的`sortedBy`移动至最前面，然后对剩下的进行字符串的正常排序。之所以出现倒序，是因为每一个链式调用，都将上一个`Ordering`进行封装成为一个新的`Ordering`。

（对“倒序”规则的一个例外：当链式调用涉及`compound`时，读取从左至右。为了避免混淆，不要将`compound`与其他类型的链式调用一起使用。）

超过几个调用的调用链将会难以理解。就像上述例子一样，我们推荐限制链式调用的调用长度不大于 3 个。即使这样，你也许会期望进一步简化，将 中间对象 - 如`Function`实例 - 分离出来，就像这样：
```java
Ordering<Foo> ordering = Ordering.natural().nullsFirst().onResultOf(sortKeyFunction);
```

### 应用
Guava 提供了许多方法来通过`Ordering`对值或集合进行操作或检查。我们在以下列出了最常用的：

方法 | 描述 | 类似的
---|---|---
[`greatestOf(Iterable iterable, int k)	`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#greatestOf-java.lang.Iterable-int-) | 返回指定 iterable 中`k`个最大的元素，按照该`Ordering`从大到小排序 | [`leastOf`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#leastOf-java.lang.Iterable-int-)
[`isOrdered(Iterable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#isOrdered-java.lang.Iterable-) | 测试指定的 `Iterable` 是否按照`Ordering`非递减排序。 | [`isStrictlyOrdered`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#isStrictlyOrdered-java.lang.Iterable-)
[`sortedCopy(Iterable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#sortedCopy-java.lang.Iterable-) | 返回一个对指定元素进行排序的副本`List` | [`immutableSortedCopy`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#immutableSortedCopy-java.lang.Iterable-)
[`min(E, E)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#min-E-E-) | 按照`Ordering`返回输入参数中较小的一个。假如相等，则返回第一个 | [`max(E, E)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#max-E-E-) 
[`min(E, E, E, E...)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#min-E-E-E-E...-) | 按照`Ordering`返回输入参数中较小的一个。假如存在多个最小值，则返回第一个 | [`max(E, E, E, E...)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#max-E-E-E-E...-)
[`min(Iterable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#min-java.lang.Iterable-) | 返回指定`Iterable`中最小的元素。假如`Iterable`为空则抛出`NoSuchElementException`异常 | [`max(Iterable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#max-java.lang.Iterable-), [`min(Iterator)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#min-java.util.Iterator-), [`max(Iterator)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html#max-java.util.Iterator-)