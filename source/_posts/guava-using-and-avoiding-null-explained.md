---
title: （Guava 译文系列）使用和避免使用 null
date: 2019-08-20 23:05:17
tags:
- guava
- translation
categories:
- Guava
---

# 使用和避免使用 null
> "null 烂透了“ - [Doug Lea](http://en.wikipedia.org/wiki/Doug_Lea)
> "这是我犯得值 10 亿刀的错误" - [Sir C. A. R. Hoare](http://en.wikipedia.org/wiki/C._A._R._Hoare) 提到他发明的 null 时如是说

对`null`的粗心使用会导致各种各样令人难以置信的错误。对 Google 的基础代码进行研究后，我们发现大约 95% 的集合中都不应有任何 null 值，如果对 `null` 快速失败而不是默默地接受，便会对开发者有所帮助。

此外，`null` 也会产生令人不悦的模糊情况。通常很少有能准确的揣摩到返回`null`值原本意义场景，例如，当 `Map.get(key)`返回`null` 时，一种可能是 key 对应的值本来就是`null`，而另一种情况则是该 key 并不存在。`Null` 能代表失败，能代表成功，能代表几乎任何事。如果能用其他的值来代替`null`，会使表意更加清晰。

即便如此，有些时候使用 `null` 仍旧是正确合理的。从内存占用和速度角度讲，`null` 非常划算，而且不可避免的会在对象数组中使用。然而，相较于库代码，在应用代码中， null 是造成逻辑困扰、难以理解的 bug 以及令人不悦的模糊的主要来源。例如，当 `Map.get` 返回 null 时，null 可能代表值不存在，或者值存在且等于 null。最要命的，null 不会对他所代表的的意义给出任何提示。

由于上述原因，大多数 Guava 工具都设计为只要能找到 null 的替代方案，就对 null 采取快速失败处理，不允许使用 null。此外，Guava 提供了许多工具，让你在必须使用 `null` 的时候能更简单，也能帮助你避免使用 `null`。

<!-- more -->

## 具体案例

如果你尝试在 `set` 中用 `null` 做值或在 `map` 中用 `null` 做 key -- 别这么干；在查找操作中明确的将 `null` 用作特殊值会使逻辑更清晰（也会少一些惊喜）。

假如你想在 Map 中用 `null` 作为值 -- 把这个 entry 取出来把；确保只存在独立的非 null `Set`（或 null `Set`）。`Map` 中存在一个 entry 的值为 `null` 和 entry 不存在这两种情况非常容易混淆，因此更好的办法是让这两种情况完全独立起来，并且仔细考虑在应用中遇到 `null` 时应该代表什么样的意义。

假如在 `List` 中使用 null -- 如果这个 list 中的元素不多时，也许使用 `Map<Integer, E>`是个更好的选择？这种方案实际上更高效，且可能实际上更准确的符合你的应用程序的真实需求。

有时我们可以考虑使用一个原生的 null 对象。例如，对枚举类增加一个常量来代表你期望的 null 值，在`java.math.RoundingMode`中提供了一个 `UNNECESSARY`值来代表“不做舍入，如果需要做舍入则抛出异常”的情况。

如果你真的需要 null 值，同时你在对 null 不友好的集合实现中遇到了问题，那么可以尝试使用其他类型的实现。例如，用`Collections.unmodifiableList(Lists.newArrayList())`替代`ImmutableList`。

## Optional
许多时候，程序员使用`null`都是为了表示某些类型的缺失：也许某处本应有值，但是没有，或某些东西找不到。就好比`Map.get`当找不到 key 对应的 value 时返回 `null`。

`Optional<T>`是一种替换可为 null 的`T`引用非 null 值的方式。一个`Optional`可能会包含一个非空的`T`引用(我们称这种情况的引用为 ”present“)，也可能什么都不包含(我们称这种情况为”absent”)。`Optional`并不存在“包含 null ”的情况。

```java
Optional<Integer> possible = Optional.of(5);
possible.isPresent(); // returns true
possible.get(); // returns 5
```

`Optional` 并不打算直接模拟其他编程环境下的"option"或者"maybe"结构，尽管他们可能会有一些相似之处。

以下我们列举了一些最常用的`Optional`操作。

### 构建一个 Optional
以下每一个都是`Optional`的静态方法。

Method | Description
---|---
[`Optional.of(T)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Optional.html#of-T-) | 构建一个包含非 null 值的 Optional，或当遇到 null 时快速失败。
[`Optional.absent()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Optional.html#absent--) | 返回一个缺失了某种类型的 Optional。
[`Optional.fromNullable(T)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Optional.html#fromNullable-T-) | 将一个可能为 null 的引用转换为 Optional，将非 null 值视为 present，将 null 视为 absent。

### 查询方法

Method | Description
---|---
[`boolean isPresent()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Optional.html#isPresent--) | 假如该`Optional` 包含一个非 null 实例，则返回 `true`。
[`T get()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Optional.html#get--) | 返回包含的`T`类型实例，当该实例不存在时抛出`IllegalStateException`异常。
[`T or(T)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Optional.html#or-T-) | 返回包含的`T`类型实例，当该实例不存在时返回给定的默认值。
[`T orNull()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Optional.html#orNull--) | 返回包含的`T`类型实例，当该实例不存在时返回`null`。本方法是`fromNullable`的反方法。
[`Set<T> asSet()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Optional.html#asSet--) | 返回一个包含该`Optional`实例的不可变单例 `Set`，如果为空则返回一个空不可变`Set`。

`Optional`提供了除上述以外更多方便的工具，详情请见 Javadoc。

### 意义何在？
除了给 `null`设置了一个名字能增加可读性外，Optional 最大的优点是防呆设计。他会强制你积极地思考可能的值缺失情况，并迫使你必须将 Optional 反解开来处理这种缺失情况，只有这样才能通过编译。让人感到不安的是，null 非常容易让人忘记考虑值缺失的情况。尽管用 FindBugs 能有所改善，但我们完全不认为他能解决 null 的问题。

当面对**返回值**可能存在，也可能不存在的情况时，上述问题尤为严重。相比你可能会忘记`other.method(a, b)`的参数**a**可能存在 null 的情况，你(和其他人)更可能会忘记`other.method(a, b)`也许会返回一个 null 值。而返回一个 `Optional`的值，会使上述情况变得不可能发生，因为调用者为了使编译通过，必须将返回的`Optional`反解开才能进一步处理。

### 简便方法
无论何时你想要用一个默认值来替换`null`时，请使用[`MoreObjects.firstNonNull(T, T)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/MoreObjects.html#firstNonNull-T-T-)。就好像方法名所描述的一样，如果两个输入值都为`null`，则会抛出`NullPointerException`。当然如果你使用的是`Optional`，会有些更好的替代方法，例如：`first.or(second)`。

在`Strings`中提供了一些处理可能为空的`String`的方法。
具体来说，我们提供了下述名称：
- [`emptyToNull(String)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Strings.html#emptyToNull-java.lang.String-)
- [`isNullOrEmpty(String)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Strings.html#isNullOrEmpty-java.lang.String-)
- [`nullToEmpty(String)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Strings.html#nullToEmpty-java.lang.String-)

我们想要强调，这些方法主要用于对接一些令人不悦的 API(这些 API 认为 null 字符串与 empty 字符串等同)。每当你将 null 字符串与 empty 字符串混在一起使用的时候，你都会听到 Guava 团队在你的耳边哭泣。(当然，如果 null 字符串与 empty 字符串代表不同的意义时可能会好一些，但是将他们视为等同的做法完全是一个令人不安且常见的代码坏味道。)