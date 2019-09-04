---
title: 通用Object方法
date: 2019-09-04 22:42:30
tags:
- guava
- translation
categories:
- Guava
---
## 通用Object方法
### equals
假如你的对象成员变量可以为 `null`，那么实现 `Objec.equals`将会很痛苦，因为你得单独检查`null`的情况。使用[`Objects.equal`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Objects.html#equal-java.lang.Object-java.lang.Object-)可以让你采用 null 敏感的检查方式来实施`equals`，且不存在抛出`NullPointerException`的风险。

```java
Objects.equal("a", "a"); // returns true
Objects.equal(null, "a"); // returns false
Objects.equal("a", null); // returns false
Objects.equal(null, null); // returns true
```
注意：JDK7 中新引入的类`Objects`提供了相同的实现[`Objects.equals`](http://docs.oracle.com/javase/7/docs/api/java/util/Objects.html#equals(java.lang.Object,%20java.lang.Object))。

### hashCode
对一个`Object`中所有成员变量的哈希应当更为简单。Guava 的[`Objects.hashCode(Object...)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Objects.html#hashCode-java.lang.Object...-)为指定序列的成员变量创建了明智的、顺序敏感的哈希。使用`Objects.hashCode(field1, field2, ..., fieldn)`来替代手工构建哈希。

注意：JDK7 中新引入的类`Objects`提供了相同的实现[`Objects.hash(Object...)`](http://docs.oracle.com/javase/7/docs/api/java/util/Objects.html#hash(java.lang.Object...))。

### toString
一个优秀的`toString`方法会对 debug 产生难以估量的价值，然而他写起来很痛苦。使用[`MoreObjects.toStringHelper()`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/MoreObjects.html#toStringHelper-java.lang.Object-)来简单地创建一个实用的`toString`。以下包含一些简单的示例：
```java
   // Returns "ClassName{x=1}"
   MoreObjects.toStringHelper(this)
       .add("x", 1)
       .toString();

   // Returns "MyObject{x=1}"
   MoreObjects.toStringHelper("MyObject")
       .add("x", 1)
       .toString();
```

### compare/compareTo
实现一个`Comparator`，或者直接实现`Comparable`接口，会很痛苦。参考：
```java
class Person implements Comparable<Person> {
  private String lastName;
  private String firstName;
  private int zipCode;

  public int compareTo(Person other) {
    int cmp = lastName.compareTo(other.lastName);
    if (cmp != 0) {
      return cmp;
    }
    cmp = firstName.compareTo(other.firstName);
    if (cmp != 0) {
      return cmp;
    }
    return Integer.compare(zipCode, other.zipCode);
  }
}
```
上面的代码很容易搞混，难以查找 bug 且存在令人难受的冗余。我们应当做的更好才对。

据以上目标，Guava 提供了[`ComparisonChain`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/ComparisonChain.html)。

`ComparisonChain` 实现了“懒比较”：他只执行比较，直到发现了非零的结果，然后忽略进一步的输入。

```java
   public int compareTo(Foo that) {
     return ComparisonChain.start()
         .compare(this.aString, that.aString)
         .compare(this.anInt, that.anInt)
         .compare(this.anEnum, that.anEnum, Ordering.natural().nullsLast())
         .result();
   }
```
这种流式语法更易读，不容易出现意外的打字错误，并且足够聪明到除非必须否则不做任何额外的工作。更多的与比较相关的实用工具可以在 Guava 的“流式比较器”类[`Ordering`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/collect/Ordering.html)中找到，可在[这里](https://github.com/google/guava/wiki/OrderingExplained)找到详细解释。