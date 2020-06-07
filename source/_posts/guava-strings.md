---
title: Guava 译文系列）Strings
date: 2020-06-07 23:11:22
tags:
- guava
- translation
categories:
- Guava
---

## Joiner

以某种分隔符连接连接字符串序列通常不应该那么麻烦，当序列中包含 null 值的时候尤甚。流式风格的 [`Joiner`] 简化了这种操作。

```java
Joiner joiner = Joiner.on("; ").skipNulls();
return joiner.join("Harry", null, "Ron", "Hermione");
```

上述代码将返回字符串：”Harry; Ron; Hermione“。除了`skipNulls`以外，还可以用`useForNull(String)`来指定将 null 替换为给定的字符串。

你也可以直接对对象使用 `Joiner` ，这会自动调用对象的`toString()`方法后进行拼接。

```java
Joiner.on(",").join(Arrays.asList(1, 5, 7)); // returns "1,5,7"
```

**注意**：joiner 实例是不可变的。因此 joiner 的配置方法总会返回一个新的 `Joiner`，并通过这种方式来获取所需语义的 joiner。这种特性让`Joiner`能够线程安全，并可以将之定义为一个`static final`的常量。

## Splitter

Java 内置的字符串分割器存在一些古怪的行为。例如：`String.split` 会默默地丢弃最尾部的分隔符，而 `StringTokenizer`只用五个空格来工作。（译者注：原文 “StringTokenizer respects exactly five whitespace characters and nothing else”。这里应该是指`StringTokenizer`默认采用`\n` `\r`, 空格，Tab 符，换页符这五种符号来分隔字符串，而这五种符号在字符串中都以类似“空格”的形式显示。）

小测验:  `",a,,b,".split(",")` 会返回什么结果？

1.  `"", "a", "", "b", ""`
1.  `null, "a", null, "b", null`
1.  `"a", null, "b"`
1.  `"a", "b"`
1.  以上都不对

正确答案是以上都不对，返回的实际是`"", "a", "", "b"`。只有尾部的分隔符被忽略掉了。我甚至都搞不懂为什么要这么做。

[`Splitter`] 采用令人安心且直截了当的流式风格来允许用户完全的控制它并避免出现所有那些令人困惑的行为。

```java
Splitter.on(',')
    .trimResults()
    .omitEmptyStrings()
    .split("foo,bar,,   qux");
```

上述代码会返回一个包含了"foo", "bar", "qux" 的`Iterable<String>`。`Splitter`可以被配置为使用`Pattern`, `char`, `String`或 `CharMatcher`来进行分割。

#### 基本工厂方法

| Method                                                     | Description                                                  | Example                                                      |
| :--------------------------------------------------------- | :----------------------------------------------------------- | :----------------------------------------------------------- |
| [`Splitter.on(char)`]                                      | 在出现特定的单个字符时分割。                                 | `Splitter.on(';')`                                           |
| [`Splitter.on(CharMatcher)`]                               | 在某个类别中出现任意字符时分割。                             | `Splitter.on(CharMatcher.BREAKING_WHITESPACE)`<br>`Splitter.on(CharMatcher.anyOf(";,."))` |
| [`Splitter.on(String)`]                                    | 以`String`字面量分割。                                       | `Splitter.on(", ")`                                          |
| [`Splitter.on(Pattern)`]<br>[`Splitter.onPattern(String)`] | 以正则表达式分割。                                           | `Splitter.onPattern("\r?\n")`                                |
| [`Splitter.fixedLength(int)`]                              | 将字符串以固定长度字符数来分割。最后一个部分的长度有可能小于`length`，但一定不会为空。 | `Splitter.fixedLength(3)`                                    |

#### 修饰方法

| Method                       | Description                                                  | Example                                                      |
| :--------------------------- | :----------------------------------------------------------- | :----------------------------------------------------------- |
| [`omitEmptyStrings()`]       | 自动忽略结果中包含的空字符串。                               | `Splitter.on(',').omitEmptyStrings().split("a,,c,d")` returns `"a", "c", "d"` |
| [`trimResults()`]            | 删除结果中的空格符，与`trimResults(CharMatcher.WHITESPACE)`的语义相同。 | `Splitter.on(',').trimResults().split("a, b, c, d")` returns `"a", "b", "c", "d"` |
| [`trimResults(CharMatcher)`] | 删除结果中匹配`CharMatcher`的字符。                          | `Splitter.on(',').trimResults(CharMatcher.is('_')).split("_a ,_b_ ,c__")` returns `"a ", "b_ ", "c"`. |
| [`limit(int)`]               | 在指定数量的字串被返回后停止分割。                           | `Splitter.on(',').limit(3).split("a,b,c,d")` returns `"a", "b", "c,d"` |

如果你想要返回`List`，使用[`splitToList()`][`Splitter.splitToList(CharSequence)`]来代替`split()`。

**注意**：splitter 实例是不可变的。因此 joiner 的配置方法总会返回一个新的 `Splitter`，并通过这种方式来获取所需语义的 joiner。这种特性让`Splitter`能够线程安全，并可以将之定义为一个`static final`的常量。

#### Map 分割器

You can also use a splitter to deserialize a map by specifying a second
delimiter using [`withKeyValueSeparator()`][`Splitter.withKeyValueSeparator()`].

The resulting [`MapSplitter`] will split the input into entries using the
splitter's delimiter, and then split those entries into keys and values using
the given key-value separator, returning a `Map<String, String>`.

你也可以用splitter来将字符串反序列化为一个 map，通过[`withKeyValueSeparator()`][`Splitter.withKeyValueSeparator()`]来指定第二个分隔符。

[`MapSplitter`] 的结果会通过分隔符来将输入分割为entries，之后通过给定的 key-value 分隔器将这些entrie分割成 key 和 value，最后返回`Map<String, String>`。

<!-- Hidden Section (why?)

## Escaper

Escaping strings correctly -- converting them into a format safe for inclusion
in e.g. an XML document or a Java source file -- can be a tricky business, and
critical for security reasons. Guava provides a flexible API for escaping text,
and a number of built-in escapers, in the com.google.common.escape package.

All escapers in Guava extend the
[http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/escape/Escaper.html
Escaper] abstract class, and support the method String escape(String). Built-in
Escaper instances can be found in several classes, depending on your needs:
[http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/html/HtmlEscapers.html
HtmlEscapers],
[http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/xml/XmlEscapers.html
XmlEscapers],
[http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/escape/SourceCodeEscapers.html
SourceCodeEscapers],
[http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/net/UriEscapers.html
UriEscapers], or you can build your own with
[http://google.github.io/guava/releases/snapshot/api/docs/ an Escapers.Builder].
To inspect an Escaper, you can use Escapers.computeReplacement to find the
replacement string for a given character.
-->

## CharMatcher

In olden times, our `StringUtil` class grew unchecked, and had many methods like
these:

*   `allAscii`
*   `collapse`
*   `collapseControlChars`
*   `collapseWhitespace`
*   `lastIndexNotOf`
*   `numSharedChars`
*   `removeChars`
*   `removeCrLf`
*   `retainAllChars`
*   `strip`
*   `stripAndCollapse`
*   `stripNonDigits`

They represent a partial cross product of two notions:

1.  what constitutes a "matching" character?
1.  what to do with those "matching" characters?

To simplify this morass, we developed `CharMatcher`.

Intuitively, you can think of a `CharMatcher` as representing a particular class
of characters, like digits or whitespace. Practically speaking, a `CharMatcher`
is just a boolean predicate on characters -- indeed, `CharMatcher` implements
[`Predicate<Character>`] -- but because it
is so common to refer to "all whitespace characters" or "all lowercase letters,"
Guava provides this specialized syntax and API for characters.

But the utility of a `CharMatcher` is in the _operations_ it lets you perform on
occurrences of the specified class of characters: trimming, collapsing,
removing, retaining, and much more. An object of type `CharMatcher` represents
notion 1: what constitutes a matching character? It then provides many
operations answering notion 2: what to do with those matching characters? The
result is that API complexity increases linearly for quadratically increasing
flexibility and power. Yay!

```java
String noControl = CharMatcher.javaIsoControl().removeFrom(string); // remove control characters
String theDigits = CharMatcher.digit().retainFrom(string); // only the digits
String spaced = CharMatcher.whitespace().trimAndCollapseFrom(string, ' ');
  // trim whitespace at ends, and replace/collapse whitespace into single spaces
String noDigits = CharMatcher.javaDigit().replaceFrom(string, "*"); // star out all digits
String lowerAndDigit = CharMatcher.javaDigit().or(CharMatcher.javaLowerCase()).retainFrom(string);
  // eliminate all characters that aren't digits or lowercase
```

**Note:** `CharMatcher` deals only with `char` values; it does not understand
supplementary Unicode code points in the range 0x10000 to 0x10FFFF. Such logical
characters are encoded into a `String` using surrogate pairs, and a
`CharMatcher` treats these just as two separate characters.

### Obtaining CharMatchers

Many needs can be satisfied by the provided `CharMatcher` factory methods:

*   [`any()`]
*   [`none()`]
*   [`whitespace()`]
*   [`breakingWhitespace()`]
*   [`invisible()`]
*   [`digit()`]
*   [`javaLetter()`]
*   [`javaDigit()`]
*   [`javaLetterOrDigit()`]
*   [`javaIsoControl()`]
*   [`javaLowerCase()`]
*   [`javaUpperCase()`]
*   [`ascii()`]
*   [`singleWidth()`]

Other common ways to obtain a `CharMatcher` include:

| Method                  | Description                                                  |
| :---------------------- | :----------------------------------------------------------- |
| [`anyOf(CharSequence)`] | Specify all the characters you wish matched. For example, `CharMatcher.anyOf("aeiou")` matches lowercase English vowels. |
| [`is(char)`]            | Specify exactly one character to match.                      |
| [`inRange(char, char)`] | Specify a range of characters to match, e.g. `CharMatcher.inRange('a', 'z')`. |

Additionally, `CharMatcher` has [`negate()`], [`and(CharMatcher)`], and
[`or(CharMatcher)`]. These provide simple boolean operations on `CharMatcher`.

### Using CharMatchers

`CharMatcher` provides a [wide variety] of methods to operate on occurrences of
the specified characters in any `CharSequence`. There are more methods provided
than we can list here, but some of the most commonly used are:

| Method                                      | Description                                                  |
| :------------------------------------------ | :----------------------------------------------------------- |
| [`collapseFrom(CharSequence, char)`]        | Replace each group of consecutive matched characters with the specified character. For example, `WHITESPACE.collapseFrom(string, ' ')` collapses whitespaces down to a single space. |
| [`matchesAllOf(CharSequence)`]              | Test if this matcher matches all characters in the sequence. For example, `ASCII.matchesAllOf(string)` tests if all characters in the string are ASCII. |
| [`removeFrom(CharSequence)`]                | Removes matching characters from the sequence.               |
| [`retainFrom(CharSequence)`]                | Removes all non-matching characters from the sequence.       |
| [`trimFrom(CharSequence)`]                  | Removes leading and trailing matching characters.            |
| [`replaceFrom(CharSequence, CharSequence)`] | Replace matching characters with a given sequence.           |

(Note: all of these methods return a `String`, except for `matchesAllOf`, which
returns a `boolean`.)

## Charsets

Don't do this:

```java
try {
  bytes = string.getBytes("UTF-8");
} catch (UnsupportedEncodingException e) {
  // how can this possibly happen?
  throw new AssertionError(e);
}
```

Do this instead:

```java
bytes = string.getBytes(Charsets.UTF_8);
```

[`Charsets`] provides constant references to the six standard `Charset`
implementations guaranteed to be supported by all Java platform implementations.
Use them instead of referring to charsets by their names.

TODO: an explanation of charsets and when to use them

(Note: If you're using JDK7, you should use the constants in
[`StandardCharsets`]

## CaseFormat

`CaseFormat` is a handy little class for converting between ASCII case
conventions &mdash; like, for example, naming conventions for programming
languages. Supported formats include:

| Format               | Example            |
| :------------------- | :----------------- |
| [`LOWER_CAMEL`]      | `lowerCamel`       |
| [`LOWER_HYPHEN`]     | `lower-hyphen`     |
| [`LOWER_UNDERSCORE`] | `lower_underscore` |
| [`UPPER_CAMEL`]      | `UpperCamel`       |
| [`UPPER_UNDERSCORE`] | `UPPER_UNDERSCORE` |

Using it is relatively straightforward:

```java
CaseFormat.UPPER_UNDERSCORE.to(CaseFormat.LOWER_CAMEL, "CONSTANT_NAME")); // returns "constantName"
```

We find this especially useful, for example, when writing programs that generate
other programs.

## Strings

A limited number of general-purpose `String` utilities reside in the [`Strings`]
class.

[`Joiner`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Joiner.html
[`Splitter`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html
[`Splitter.on(char)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#on-char-
[`Splitter.on(CharMatcher)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#on-com.google.common.base.CharMatcher-
[`Splitter.on(String)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#on-java.lang.String-
[`Splitter.on(Pattern)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#on-java.util.regex.Pattern-
[`Splitter.onPattern(String)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#onPattern-java.lang.String-
[`Splitter.fixedLength(int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#fixedLength-int-
[`Splitter.splitToList(CharSequence)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#splitToList-java.lang.CharSequence-
[`Splitter.withKeyValueSeparator()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#withKeyValueSeparator-java.lang.String-
[`MapSplitter`]: https://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.MapSplitter.html
[`omitEmptyStrings()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#omitEmptyStrings--
[`trimResults()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#trimResults--
[`trimResults(CharMatcher)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#trimResults-com.google.common.base.CharMatcher-
[`limit(int)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Splitter.html#limit-int-
[`any()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#any--
[`none()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#none--
[`whitespace()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#whitespace--
[`breakingWhitespace()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#breakingWhitespace--
[`invisible()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#invisible--
[`digit()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#digit--
[`javaLetter()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#javaLetter--
[`javaDigit()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#javaDigit--
[`javaLetterOrDigit()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#javaLetterOrDigit--
[`javaIsoControl()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#javaIsoControl--
[`javaLowerCase()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#javaLowerCase--
[`javaUpperCase()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#javaUpperCase--
[`ascii()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#ascii--
[`singleWidth()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#singleWidth--
[`anyOf(CharSequence)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#anyOf-java.lang.CharSequence-
[`is(char)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#is-char-
[`inRange(char, char)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#inRange-char-char-
[`negate()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#negate--
[`and(CharMatcher)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#and-com.google.common.base.CharMatcher-
[`or(CharMatcher)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#or-com.google.common.base.CharMatcher-
[wide variety]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#method_summary
[`collapseFrom(CharSequence, char)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#collapseFrom-java.lang.CharSequence-char-
[`matchesAllOf(CharSequence)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#matchesAllOf-java.lang.CharSequence-
[`removeFrom(CharSequence)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#removeFrom-java.lang.CharSequence-
[`retainFrom(CharSequence)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#retainFrom-java.lang.CharSequence-
[`trimFrom(CharSequence)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#trimFrom-java.lang.CharSequence-
[`replaceFrom(CharSequence, CharSequence)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CharMatcher.html#replaceFrom-java.lang.CharSequence-java.lang.CharSequence-
[`Charsets`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Charsets.html
[`StandardCharsets`]: http://docs.oracle.com/javase/7/docs/api/java/nio/charset/StandardCharsets.html
[`LOWER_CAMEL`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CaseFormat.html#LOWER_CAMEL
[`LOWER_HYPHEN`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CaseFormat.html#LOWER_HYPHEN
[`LOWER_UNDERSCORE`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CaseFormat.html#LOWER_UNDERSCORE
[`UPPER_CAMEL`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CaseFormat.html#UPPER_CAMEL
[`UPPER_UNDERSCORE`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/CaseFormat.html#UPPER_UNDERSCORE
[`Strings`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Strings.html
[`Predicate\<Character>`]: FunctionalExplained#predicate