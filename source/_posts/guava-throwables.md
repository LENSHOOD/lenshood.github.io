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

**无需再将`throws Throwable`转换为`throws Exception`**
一些 API，尤其是 Java 反射类以及 JUnit (JUnit 大量使用了反射)，声明方法会抛出`Throwable`。与这些 API 交互会很痛苦，因为即使是最通用的 API 通常也只声明`throws Exception`。`Throwables.propagate`被一些知道他一定不会抛出`Exception`和`Error`的调用方使用。这里有一个定义`Callable`来执行 JUnit 测试的例子：
```java
public Void call() throws Exception {
  try {
    FooTest.super.runTest();
  } catch (Throwable t) {
    Throwables.propagateIfPossible(t, Exception.class);
    Throwables.propagate(t);
  }

  return null;
}
```
这里并不需要`propagate()`，因为第二行与`throw new RuntimeException(t)`相同。(题外话：这个例子也提醒到我`propagateIfPossible`存在潜在的混淆，因为它不只是传播了给定的异常类型，还会传播`RuntimeException`和`Errors`。)

上述模式(或其变体例如`throw new RuntimeException(t)`)在 Google 的代码库中出现了不下 30 次。（搜索`'propagateIfPossible[^;]* Exception.class[)];'`。）其中只有一小部分采用了`throw new RuntimeException(t)`的实现。也许我们想要一个`throwWrappingWeirdThrowable`方法来进行`Throwable`和`Exception`的转换，但却采用了上述两行代码来代替，除非我们将`propagateIfPossible`设为过时，否则也许并没有什么必要使用新方法。

####对`Throwables.propagate`存在争议的用法
**争议：将检查类异常转换为非检查异常**
原则上将，非检查异常意味着 bug，检查异常意味着超出你控制范围的问题。事实上，连 JDK 有的时候[也](https://docs.oracle.com/javase/6/docs/api/java/lang/Object.html#clone%28%29)[搞](https://docs.oracle.com/javase/6/docs/api/java/lang/Integer.html#parseInt%28java.lang.String%29)[错了](https://docs.oracle.com/javase/6/docs/api/java/net/URI.html#URI%28java.lang.String%29)(或者至少对于某些方法，[没有对所有人都正确的答案](http://docs.oracle.com/javase/6/docs/api/java/net/URI.html#create%28java.lang.String%29))。

结果就是，调用者有时需要在这两种异常之间做转换。

```java
try {
  return Integer.parseInt(userInput);
} catch (NumberFormatException e) {
  throw new InvalidInputException(e);
}
```

```java
try {
  return publicInterfaceMethod.invoke();
} catch (IllegalAccessException e) {
  throw new AssertionError(e);
}
```

有的时候一些调用者会使用`Throwables.propagate`。那么使用他有什么不好的地方呢？

一个主要的问题是这会使代码变得不够清晰。`throw Throwables.propagate(ioException)`是干什么的？`throw new RuntimeException(ioException)`是干什么的？这二者做着同样的事情，但后者显然更直截了当。前者引出了问题："这到底是干什么的？他应该不只是包装了`RuntimeException`对吧？如果是的话，那干嘛还要包装一层呢？"

诚然，问题的一部分出在“propagate”本身就是一个模糊地命名。这是一种[抛出未声明异常的方法吗](http://www.eishay.com/2011/11/throw-undeclared-checked-exception-in.html)？也许叫做“wrapIfChecked“可能会更好。但是即使调用了该方法，在已知的检查异常上调用它也没有任何好处。他甚至还可能存在一个额外的问题：也许相比`RuntimeException`，抛出`IllegalArgumentException`会更好。

我们有时也会看到`propagate`使用在*也许*只会抛出检查异常的地方。结果就是这与通常的方式相比更小一点，也更不直接一点：
```java
} catch (RuntimeException e) {
  throw e;
} catch (Exception e) {
  throw new RuntimeException(e);
}
```

```java
} catch (Exception e) {
  throw Throwables.propagate(e);
}
```

然而，不容忽视的是将检查异常转换为非检查异常的通用实践，在有些时候是毋庸置疑的，然而更常见的是他多被用于避免处理合法的检查异常。这就引出了一个争论，即检查异常总体上是否是一个坏点子。我在这里并不想谈论这些。以上内容已经足以说明`Throwables.propagate`并不存在用以鼓励 Java 用户忽略`IOException`及类似异常的目的。

**争议：从其他线程再抛出异常**
```java
try {
  return future.get();
} catch (ExecutionException e) {
  throw Throwables.propagate(e.getCause());
}
```

这里需要考虑很多事情：
1. 上述 cause 可能是一个检查异常。可见上文”将检查异常转换为非检查异常”。但假如我们已知这个 task 不会抛出检查异常呢？(如果他是一个`Runable`的结果。)按上述讨论，你可以捕获他，并抛出一个`AssertionError`；`propagate`可以提供稍多一点功能。尤其是对`Future`，可以考虑[`Future.get`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/util/concurrent/Futures.html#getUnchecked-java.util.concurrent.Future-)。
2. 上述 cause 可能不会抛出任何`Exception`和`Errors`。(好吧，这可能不是真的，但是如果你直接将之再抛出，则编译器确实会强制你考虑这种可能性。)可见上文：将`throws Throwable`转换为 `throws Exception`。
3. 上述 cause 可能是一个非检查异常或是 `Error`。如果是，则他会被直接抛出。不幸的是，栈追踪信息会显示最初创建的线程的异常，而不是当前线程的传播该异常处。通常最好在异常链中包含两个线程的栈追踪信息，就像`get`抛出的`ExecutionException`一样。(这个问题实际上与`propagate`无关；他与任何尝试在不同的线程再抛出异常的代码有关。)

### 因果链
Guava 为了能让研究一个异常的因果链变得简单一点，提供了三个有用的方法签名自解释的方法：
- [`Throwable getRootCause(Throwable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Throwables.html#getRootCause-java.lang.Throwable-)
- [`List<Throwable> getCausalChain(Throwable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Throwables.html#getCausalChain-java.lang.Throwable-)
- [`String getStackTraceAsString(Throwable)`](http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Throwables.html#getStackTraceAsString-java.lang.Throwable-)
