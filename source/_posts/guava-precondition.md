---
title: （Guava 译文系列）先决条件
date: 2019-09-02 23:18:24
tags:
- guava
- translation
categories:
- Guava
---

## 先决条件
Guava 提供了大量的判断先决条件的工具。我们强烈建议将之通过 static import 引入。

<!-- more -->

每一种方法，都有以下三种变体形式：
1. 无额外入参。任何抛出的异常都不包含错误描述。
2. 存在一个`Object`参数。任何抛出异常的错误描述都通过`Object.toString()`获取。
3. 存在一个`String`参数加数个额外的`Object`参数。上述参数表现为类似 printf 方法的使用形式（当然更高效且与 GWT 兼容），但只支持`%s`指示器。
    - 注意：`checkNotNull`，`checkArgument`和`checkState`有大量的重载采用基本类型和`Object`类型结合的方式实现，而不是采用 varargs 数组。这可以使调用在大多数情况下无需进行自动装箱与 varargs 数组分配的工作。

上述第三中变体的示例：
``` java
checkArgument(i >= 0, "Argument was %s but expected nonnegative", i);
checkArgument(i < j, "Expected i < j, but %s >= %s", i, j);
```

| 方法签名（不包括附加入参）                                   | 描述                                                         | 错误抛出的异常              |
| ------------------------------------------------------------ | ------------------------------------------------------------ | --------------------------- |
| [`checkArgument(boolean)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Preconditions.html#checkArgument-boolean-) | 检查输入的`boolean`为`true`。用于校验方法的输入参数          | `IllegalArgumentException`  |
| [`checkNotNull(T)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Preconditions.html#checkNotNull-T-) | 检查值不为 null，并直接返回该值，因此本方法可直接用与行内嵌  | `NullPointerException`      |
| [`checkState(boolean)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Preconditions.html#checkState-boolean-) | 检查对象的状态，不依赖于方法参数。例如，`Iterator`可以使用本方法来确保在调用`remove`之前，`next`先被调用。 | `IllegalStateException`     |
| [`checkElementIndex(int index, int size)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Preconditions.html#checkElementIndex-int-int-) | 检查`index`是一个在给定长度的 List、String 或 array 中有效的*元素*。一个元素的 index 可能会从 0（包含）至`size`（**不包含**）。你无需传入 List，String 或 array，只需传入他们的大小。本方法会将 `index` 返回。 | `IndexOutOfBoundsException` |
| [`checkPositionIndex(int index, int size)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Preconditions.html#checkPositionIndex-int-int-) | 检查`index`是一个在给定长度的 List、String 或 array 中有效的*位置*。一个位置的 index 可能会从 0（包含）至`size`（**包含**）。你无需传入 List，String 或 array，只需传入他们的大小。本方法会将 `index` 返回。 | `IndexOutOfBoundsException` |
| [`checkPositionIndexes(int start, int end, int size)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Preconditions.html#checkPositionIndexes-int-int-int-) | 检查给定的范围`[start, end)`是一个在给定长度的 List、String 或 array 中有效的子范围。本方法包含一个给定的错误信息。 | `IndexOutOfBoundsException` |

我们更喜欢使用我们自己的先决条件检查工具而不是例如 Apache Commons 提供的比较工具类，是因为：
- 通过 static import 后，Guava 的方法更清晰、准确。`checkNotNull`的名字清晰地告知了他会做什么，且会抛出什么样的异常。
- `checkNotNull`会在检查完毕后返回输入参数，这允许在构造函数中使用简单的单行实现：`this.field = checkNotNull(field);`。
- 简单、多样（支持 varargs）的异常信息。（正是这一项优势，因此我们更倾向于使用`checkNotNull`而不是[`Objects.requireNonNull`](http://docs.oracle.com/javase/7/docs/api/java/util/Objects.html#requireNonNull(java.lang.Object,java.lang.String))）

我们建议你将多个先决条件检查语句分开多行书写，这可以帮助你在  debug 时找到究竟是哪一个先决条件失败。此外，你也应提供有帮助性的错误信息，这能让处于独立行的先决条件检查更简单。