---
title: （Guava 译文系列）Throwables
date: 2019-09-02 23:18:24
tags:
- guava
- translation
categories:
- Guava
---

## Throwables
Guava 的[`Throwables`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Throwables.html)工具能经常简化与异常相关的操作。

### 传播
有些时候，当你捕获一个异常时，你想要将之再次抛出给下一个 try...catch 块。这种情况经常出现在遇到`RuntimeException`和`Error`的时候，这些异常不需要被捕获，但却仍会被 try...catch 块捕获，然而你并不想要这样做。

Guava 提供了许多工具来简化异常传播。例如：
```java
try {
  someMethodThatCouldThrowAnything();
} catch (IKnowWhatToDoWithThisException e) {
  handle(e);
} catch (Throwable t) {
  Throwables.propagateIfInstanceOf(t, IOException.class);
  Throwables.propagateIfInstanceOf(t, SQLException.class);
  throw Throwables.propagate(t);
}
```
这些方法每一个都会抛出异常，但抛出结果 - 例如 `throw Throwables.propagate(t)` - 对向编译器证明将抛出异常很有用。

以下为 Guava 提供的异常传播方法的小结：

Signature | Explanation
---|---
[`RuntimeException propagate(Throwable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Throwables.html#propagate-java.lang.Throwable-) | 传播`RuntimeException` 和 `Error`，或将异常包装为`RuntimeException`并以其它方式抛出
[`void propagateIfInstanceOf(Throwable, Class<X extends Exception>) throws X`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Throwables.html#propagateIfInstanceOf-java.lang.Throwable-java.lang.Class-) | 传播是`X`的实例的异常
[`void propagateIfPossible(Throwable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Throwables.html#propagateIfPossible-java.lang.Throwable-) | 传播是`RuntimeException`或`Error`的实例的异常
[`void propagateIfPossible(Throwable, Class<X extends Throwable>) throws X`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Throwables.html#propagateIfPossible-java.lang.Throwable-java.lang.Class-) | 传播是`RuntimeException`或`Error`或`X`的实例的异常

#### `Throwables.propagate`的使用
**模拟 Java 7 的 multi-catch 和 rethrow**
通常如果你想要让异常传播至调用栈上一层，`catch`块是完全不需要的。由于你不会从异常中恢复，所以可能并不应该对异常进行 log 或做其他的操作。你也许想要做一些清理工作，但是通常无论运行是否成功，清理工作都需要进行，所以需要在最后使用`finally`块。然而，一个可以再次抛出异常的`catch`块有时也有用：也许你想要在向上传播异常之前记录错误数，或者也许你只想在某些情况下才传播异常。

对于单个异常的情况，获取并再次抛出异常的过程直接简单。然而当存在多异常的情况时，问题会变得很糟：
```java
@Override public void run() {
  try {
    delegate.run();
  } catch (RuntimeException e) {
    failures.increment();
    throw e;
  } catch (Error e) {
    failures.increment();
    throw e;
  }
}
```
Java 7 采用[multicatch](http://docs.oracle.com/javase/7/docs/technotes/guides/language/catch-multiple.html)解决了此问题：
```java
} catch (RuntimeException | Error e) {
  failures.increment();
  throw e;
}
```
然而非 Java 7 的用户被卡住了。他们想要采用以下方法来解决问题，然而编译器并不允许他们抛出一个`Throwable`类型的变量。
```java
} catch (Throwable t) {
  failures.increment();
  throw t;
}
```
解决的办法就是用`throw Throwables.propagate(t)`来替换`throw t`。在有限的场景下， `Throwables.propagate`表现的与原先的代码行为一致。但是，采用`Throwables.propagate`的方式，还能简单的包含其他隐藏的行为。特别要注意的是，以上模式只可用于`RuntimeException`和`Error`的情况。假如`catch`块也可能会捕获检查类异常，则需要通过数个`propagateIfInstanceOf`来确保行为正常，因为`Throwables.propagate`无法直接传播检查类异常。

总而言之，使用`propagate`的方式还 ok。当然在 Java 7 之后他就完全没必要了。在其他版本下，它能够减少一点点的重复代码，但是一个简单的方法抽取（Extract Method）重构也能实现同样的效果。

另外，`propagate`的用法[很容易意外的包装检查类异常](https://github.com/google/guava/commit/287bc67cac97052b13cbbc0358aed8054b14bd4a)。