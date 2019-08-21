---
title: 使用和避免使用 null
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

由于上述原因，大多数 Guava 工具都设计为只要能找到 null 的替代方案，就对null 采取快速失败处理，不允许使用null。此外，Guava 提供了许多工具，让 你在必须使用 `null` 的时候能更简单，也能帮助你避免使用 `null`。

## 具体案例
如果你尝试在 `set` 中用 `null` 做值或在 `map` 中用 `null` 做 key -- 别这么干；在查找操作中明确的将 `null` 用作特殊值会使逻辑更清晰（也会少一些惊喜）。

假如你想在 Map 中用 `null` 作为值 -- 把这个 entry 取出来把；确保只存在独立的非 null `Set`（或 null `Set`）。`Map` 中存在一个entry 的值为 `null` 和 entry 不存在这两种情况非常容易混淆，因此更好的办法是让这两种情况完全独立起来，并且仔细考虑在应用中遇到 `null` 时应该代表什么样的意义。

假如在 `List` 中使用 null -- 如果这个 list 中的元素不多时，也许使用 `Map<Integer, E>`是个更好的选择？这种方案实际上更高效，且可能实际上更准确的符合你的应用程序的真实需求。

有时我们可以考虑使用一个原生的 null 对象。例如，对枚举类增加一个常量来代表你期望的 null 值，在`java.math.RoundingMode`中提供了一个 `UNNECESSARY`值来代表“不做舍入，如果需要做舍入则抛出异常”。

如果你真的需要 null 值，同时你在非 null 友好的集合实现中遇到了问题，那么可以尝试使用其他类型的实现。例如，用`Collections.unmodifiableList(Lists.newArrayList())`替代`ImmutableList`。

## Optional
许多时候，程序员使用`null`都是为了表示某些类型的缺失：也许某处本应有值，但是没有，或某些东西找不到。就好比`Map.get`当找不到 key 对应的 value 时返回 `null`。

`Optional<T>`是一种替换可为 null 的`T`引用非 null 值的方式。一个`Optional`可能会包含一个非空的`T`引用(我们称这种情况的引用为 ”present“)，也可能什么都不包含(我们称这种情况为”absent”)。并不存在“包含null”的情况。

```java
Optional<Integer> possible = Optional.of(5);
possible.isPresent(); // returns true
possible.get(); // returns 5
```
