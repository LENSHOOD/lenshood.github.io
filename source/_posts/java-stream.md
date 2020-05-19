---
title: 虚拟工厂：Java stream
date: 2020-05-03 23:07:27
tags:
- java
- stream
categories:
- Java
---

在我的文章 [虚拟工厂：Java 线程池]() 中简单介绍了采用 ”工厂“ 这样一个现实生活中的概念来抽象具体的线程操作，并定义了 `Worker` `TaskQueue` 等概念。用这种方式拉近了计算机域和真实世界域之间的距离，让代码表现现实的意图。

不仅仅是线程池，从 Java 8 开始引入的`Stream`也一样，它用代码构建了一套流水线体系，通过流水线环节的叠加来实现对流水线上元素的各种处理。

### 流水线（Pipeline）

在中文翻译中，我们时常会把 Assembly Line（装配线）以及 Pipeline（管线）都译为流水线，而在计算机领域，我们说的流水线通常都是 Pipeline，例如 CPU 的指令流水线（Instruction Pipeline）。

对于 Pipeline，韦氏词典的第一个解释即：

> a line of pipe with pumps, valves, and control devices for conveying liquids, gases, or finely divided solids

显然对于上述释义，Pipeline 翻译为 “管线” 似乎更为合理，不过在软件领域，对 Pipeline 进行了引申，一个 Pipeline 是一组计算过程（computing processes）的组合，并且以并行（Parallel）的方式执行。

那么实际上在计算机领域，Pipeline 与 Assembly Line 在概念上就没有太明确的区别了：我们可以将之类比为一种操作，它可以由多个工序构成，每个工序都将对上一个工序产出的工件进行进一步加工，工件从流水线入口进入，并在流水线上移动，最后被输出为某种产品。

所以根据上述描述我们可以抽象出与流水线相关的几个概念：

- 工件（workpiece）：即被流水线加工的元素
- 工序（process）：即执行单元，它以工件为输入，也以工件为输出
- 流水线（pipeline）：即流水线主体，他包括了入口，出口，在其上可以放置各种工序

那么我们就能得到如下的一个流水线结构：
{% asset_img pipeline.png %}

可见流水线上的工序理应能够灵活的组合与替换，采用各种简单而固定的工序，就能组合出来满足多种多样的需求。

### Stream

Java 8 中定义的流式操作，能够对流（Stream）叠加多种操作并进行处理。在 [Processing Data with Java SE 8 Streams, Part 1](https://www.oracle.com/technical-resources/articles/java/ma14-java-se-8-streams.html) 中提到对 Stream 的定义：

> a sequence of elements from a source that supports aggregate operations.

首先，任何单一对象、集合、数组都可以输入为一个 Stream，这就是定义里提到的 “source”。转换为 Stream 后，他们都将变身为流中的一个元素，即 “sequence of element”。在对所有元素进行一番处理之后，Stream 可以再次转换回到对象、集合、数组等，这里的处理即 “aggregate operations”。

那么，对于这 ”一番处理“，从处理的方式上划分，有以下几种[处理形式](https://docs.oracle.com/javase/8/docs/api/java/util/stream/package-summary.html#StreamOps)：

1. 中间操作（Intermediate Operations）：会返回一个新的 Stream，并且是 Lazy 操作

   - 无状态操作（StatelessOps）：对单个元素进行操作，操作之间没有联系，也不保存任何元素的状态。典型操作有：`filter` `maap` `peak` 等

   - 有状态操作（StatefulOps）：对流中元素的处理会依赖之前处理的结果，元素与元素之间有关系，下一个元素的处理依赖上一个元素的状态，或需要获取到所有的元素后才能进行操作。典型操作有：`distinct` `sorted` `limit`

2. 终止操作（Termination Operations）：穿越（traverse）整个流，得到结果或是副作用（side-effect），一旦执行了终止操作，整个流就认为已经被消费，并且无法再次执行任何操作。

#### Stream Shape

通常，Stream 中的元素都以对象引用的形式存在，但同时 Java 也考虑到了对基本类型的 Stream 支持，因此共定义了四种 “Stream Shape”：

- **Reference**：元素类型为对象引用，其行为由 `Stream` 定义
- **Int Value**：元素类型为 int，其行为由`IntStream`定义
- **Long Value**：元素类型为 long，其行为由`Longtream`定义
- **Double Value**：元素类型为 double，其行为由`DoubleStream`定义

下文中主要以 Reference 类型的 Stream 来举例说明流水线的工作原理。

### ReferencePipeline

对于元素为对象引用的 Stream（也是大多数 Stream 的形态）而言，`ReferencePipeline` 是其实现的核心。

`ReferencePipeline `的类继承关系如下图所示：

{% asset_img reference-pipeline.png %}

此外，由于`ReferencePipeline`本身实现了 `Streaam`接口，因此他实现了`Stream`中定义的所有行为，包括：`map`，`filter`，`reduce`，`collect`，`limit` 等等。所以我们可以说`ReferencePipeline`本身就是一个流水线的基础工序，基于这种基础工序，我们能构造出多种多样的工序来。

`ReferencePipeline` 作为基类，在其内部提供了如下三个实现类，这三个实现类进一步对不同种类的工序进行了定义：

- Head：入口工序
  - 与其他两个实现类相比，`Head`比较独特，也比较好理解：他会被作为整个 `Stream` 的头部，也即流水线的入口。又因为`Head`本身是一个`ReferencePipeline`，因此我们可以把`Head`理解为一道特殊的工序，他接收原料作为工件，且不对工件做任何进一步处理，但作为头部（第一道工序），我们可以在其之后追加更多的工序。
- StatelessOp：无状态工序
  - 对应前文 ”中间操作“ 中的无状态操作
  - 无状态工序不依赖前后两个工件的状态，即不论前后工件如何，他只对当前工件进行加工，典型的无状态工序是`map`
- StatefulOp：有状态工序
  - 对应前文 ”中间操作“ 中的有状态操作
  - 有状态工序依赖与工件之间的状态，他能结合多个工件来进行处理，典型的有状态工序是`sorted`

通过上述三种基于 `ReferencePipeline` 实现的操作，定义了构造一条流水线的入口（Head）及中间操作（intermediate operation），那么现在就只剩下流水线的出口（terminal operation）还没有定义了，所以，JDK 中还对流水线的出口即终止操作进行了定义：

- TerminalOp：出口工序
  - 对应前文”终止操作“
  - 将流水线各工序最后输出的工件进行整合并转换为产品，典型的终止操作是`reduce`与`collect`

综合上述几种操作类型，不出意外，我们也许能进一步想到：

- `map `操作其实就是一个行为是 ”对工件进行映射变换“ 的`StatelessOp` 工序
- `filter`操作其实就是一个行为是 ”对工件进行筛选“ 的`StatelessOp`工序
- `limit`操作其实就是一个行为是 ”只保留有限个工件“ 的`StatefulOp`工序
- `sorted`操作其实就是一个行为是 ”对所有工件进行排序的“ 的`StatefulOp`工序
- `collect`操作其实就是一个行为是 ”将所有工件整合为产品“ 的`TerminalOp`工序

因此我们日常使用的形如：

```java
Stream.of("a", "b", "c", "1", "2", "3").filter(NumUtil::isNum).map(NumUtil::minusOne).limit(1).collect(Collectors.toList);
```

其实就是先创造了一条这样的流水线：

`Head` -> `StatelessOp:filter`  -> `StatelessOp:map`   -> `StatelessOp:filter`  -> `StatefulOp:limit` -> `TerminalOp:collect`

之后输入工件：`"a", "b", "c", "1", "2", "3"`，最后启动流水线产出结果。

### 流水线的组装

#### 入口：从原料到工件

我们知道，在期望对某个或某集合进行 Stream 操作之前，我们都需要使用一种通用的方式，将元素或集合转换为 Stream，这种转换方法通常包括：

- `Stream.of(T value)`：将单个给定元素转换为 Stream
- `Arrays.stream(T[] valueArray)`：将数组转换为 Stream
- `Collection.stream()`：将集合转换为 Stream

深入这些方法后我们发现，实际上他们最终都调用了如下方法：

```java
public static <T> Stream<T> stream(Spliterator<T> spliterator, boolean parallel) {
    Objects.requireNonNull(spliterator);
    return new ReferencePipeline.Head<>(spliterator,
                                        StreamOpFlag.fromCharacteristics(spliterator),
                                        parallel);
}
```

该方法很简单，如前文所述的，构造流水线的前提是构造一个入口，即`Head`，因此不论是从单个对象，还是从数组、集合中构造流水线，我们都将创建一个入口`Head`。

那么入口有了，怎么样将原材料（输入元素）作为工件导入流水线呢？ 就靠 `Spliterator`。

`Spliterator`定义了一类行为：

> 作为 Stream 的输入源，Spiltertor 能够对元素进行遍历（traverse）或拆分（partitioning）。
>
> 可以采用`tryAdvance(Consumer<? super T> action)`方法来将 action 作用于当前元素，可以用
>
> `void forEachRemaining(Consumer<? super T> action)`方法来将 action 作用于所有元素。

在真实的使用中，流水线工序的集合从外部可以看做是一个`Consumer`，而`Spilterator`正是采用`forEachRemainig()`方法遍历每一个元素并执行流水线`Consumer.accept()`的方式来实施流水线操作的。

#### 中间操作：有状态 vs 无状态

**StatelessOp 无状态**

每一个中间操作都会在当前流水线末端增加一道工序。

我们从最简单、最容易理解的操作`map`来看看到底怎么样给流水线增加一道工序：

```java
@Override
@SuppressWarnings("unchecked")
public final <R> Stream<R> map(Function<? super P_OUT, ? extends R> mapper) {
  Objects.requireNonNull(mapper);
  return new StatelessOp<P_OUT, R>(this, StreamShape.REFERENCE,
                                   StreamOpFlag.NOT_SORTED | StreamOpFlag.NOT_DISTINCT) {
    @Override
    Sink<P_OUT> opWrapSink(int flags, Sink<R> sink) {
      return new Sink.ChainedReference<P_OUT, R>(sink) {
        @Override
        public void accept(P_OUT u) {
          downstream.accept(mapper.apply(u));
        }
      };
    }
  };
}
```

我们知道，`map`操作会对流水线上每一个工件进行相同的加工工作，具体的加工方法描述为一个`Function<? super P_OUT, ? extends R> mapper`，所以显然`map`是一种无状态操作。上述代码中也印证了这一点：

对一个`ReferencePipeline`（可以假定当前是一个`Head`）实施`map`操作实际上是返回了一个`StatelessOp`的匿名子类。那么在哪里做挂载工序操作呢？看看`StatelessOp`的构造方法：

```java
// StatelessOp
StatelessOp(AbstractPipeline<?, E_IN, ?> upstream,
            StreamShape inputShape,
            int opFlags) {
  super(upstream, opFlags);
  assert upstream.getOutputShape() == inputShape;
}

// super方法调用了 ReferencePipeline 的构造方法
ReferencePipeline(AbstractPipeline<?, P_IN, ?> upstream, int opFlags) {
  super(upstream, opFlags);
}

// ReferencePipeline 的 super 方法调用了 AbstractPipeline 的构造方法
AbstractPipeline(AbstractPipeline<?, E_IN, ?> previousStage, int opFlags) {
  if (previousStage.linkedOrConsumed)
    throw new IllegalStateException(MSG_STREAM_LINKED);
  previousStage.linkedOrConsumed = true;
  previousStage.nextStage = this;

  this.previousStage = previousStage;
  this.sourceOrOpFlags = opFlags & StreamOpFlag.OP_MASK;
  this.combinedFlags = StreamOpFlag.combineOpFlags(opFlags, previousStage.combinedFlags);
  this.sourceStage = previousStage.sourceStage;
  if (opIsStateful())
    sourceStage.sourceAnyStateful = true;
  this.depth = previousStage.depth + 1;
}
```

根据代码可知，创建的 `StatelessOp`实例，在构造函数中以调用者（刚刚我们假定为`Head`）`this` 作为其`upstream`（也就是 `AbstractPipeline` 中的`previousStage`）这有点像是构造链表时，新增节点将其`previousNode`设置为尾结点，并将尾结点的`nextNode`指向自己一样。

{% asset_img stage-connect.png %}

至于匿名类中复写的`opWrapSink`方法，我们可以暂且认为该方法会在最终流水线启动时对每一个工件调用，因此也就不难理解方法内构造了一个`Sink`（Sink 是什么后文会讲到），在`Sink`中又通过`downstream.accept(mapper.apply(u))`来对工件（也就是输入参数`u`）apply 了 mapper，并继续调用其下游的`accept`方法。

明白了`map`是怎么一回事，`filter`也就显而易见了：

```java
@Override
public final Stream<P_OUT> filter(Predicate<? super P_OUT> predicate) {
  Objects.requireNonNull(predicate);
  return new StatelessOp<P_OUT, P_OUT>(this, StreamShape.REFERENCE,
                                       StreamOpFlag.NOT_SIZED) {
    @Override
    Sink<P_OUT> opWrapSink(int flags, Sink<P_OUT> sink) {
      return new Sink.ChainedReference<P_OUT, P_OUT>(sink) {
        @Override
        public void begin(long size) {
          downstream.begin(-1);
        }

        @Override
        public void accept(P_OUT u) {
          if (predicate.test(u))
            downstream.accept(u);
        }
      };
    }
  };
}
```

`filter`除了具体的操作与`map`不同以外，仍然是构造了一个`StatelessOp`的匿名类。并且结合`filter`的语义，确实是当`predicate.test(u)`为 true 时，才会继续执行下游的操作，否则什么也不做。通过这种方法简单的实现了筛选功能。

**StatefulOp 有状态**

在来对比的看一下有状态操作`limit`:

```java
@Override
public final Stream<P_OUT> limit(long maxSize) {
  if (maxSize < 0)
    throw new IllegalArgumentException(Long.toString(maxSize));
  return SliceOps.makeRef(this, 0, maxSize);
}

// SliceOps
public static <T> Stream<T> makeRef(AbstractPipeline<?, T, ?> upstream,
                                    long skip, long limit) {
  if (skip < 0)
    throw new IllegalArgumentException("Skip must be non-negative: " + skip);

  return new ReferencePipeline.StatefulOp<T, T>(upstream, StreamShape.REFERENCE,
                                                flags(limit)) {
    ... ...
      
    @Override
    Sink<T> opWrapSink(int flags, Sink<T> sink) {
      return new Sink.ChainedReference<T, T>(sink) {
        long n = skip;
        long m = limit >= 0 ? limit : Long.MAX_VALUE;

        @Override
        public void begin(long size) {
          downstream.begin(calcSize(size, skip, m));
        }

        @Override
        public void accept(T t) {
          if (n == 0) {
            if (m > 0) {
              m--;
              downstream.accept(t);
            }
          }
          else {
            n--;
          }
        }

        ... ...
      };
    }
  };
}
```

很直白：将 `limit`值赋给`m`， 每当执行一个工件时`m--`，直到`m<0`后不再继续操作。这里的 "Stateful" 说的就是`m`了。

#### 出口：完成组装

在对流水线添加了一系列工序之后，是时候启动流水线并生产出产品了。

我们通过对流水线挂载最后一个工序：`TerminalOp`来结束流水线的创建，并启动流水线。先来看看`TerminalOp`提供的行为:

```java
interface TerminalOp<E_IN, R> {
    ... ...
    /**
     * Performs a sequential evaluation of the operation using the specified
     * {@code PipelineHelper}, which describes the upstream intermediate
     * operations.
     *
     * @param helper the pipeline helper
     * @param spliterator the source spliterator
     * @return the result of the evaluation
     */
    <P_IN> R evaluateSequential(PipelineHelper<E_IN> helper,
                                Spliterator<P_IN> spliterator);
}

```

除去我们不关心的行为，最重要的行为便是：`evaluateSequential()`了，根据注释可知，实现该方法，来对给定的`spliterator`执行操作，操作是由描述了上游中间操作的`PipelineHelper`来执行。根据前文的继承关系图，我们知道实际上`ReferencePipeling`本身就是一个`PipelineHelper`，再结合前文所述，一条流水线被创建时，`spliterator`已经与`Head`一起被作为流水线入口的一部分了，因此直接找一段比较简单的终止操作实现，`reduce`：

```java
// ReferencePipeline
@Override
public final Optional<P_OUT> reduce(BinaryOperator<P_OUT> accumulator) {
  return evaluate(ReduceOps.makeRef(accumulator));
}

// AbstractPipeline
final <R> R evaluate(TerminalOp<E_OUT, R> terminalOp) {
  assert getOutputShape() == terminalOp.inputShape();
  if (linkedOrConsumed)
    throw new IllegalStateException(MSG_STREAM_LINKED);
  linkedOrConsumed = true;

  return isParallel()
    ? terminalOp.evaluateParallel(this, sourceSpliterator(terminalOp.getOpFlags()))
    : terminalOp.evaluateSequential(this, sourceSpliterator(terminalOp.getOpFlags()));
}

// ReduceOps 
public static <T> TerminalOp<T, Optional<T>>
  makeRef(BinaryOperator<T> operator) {
  Objects.requireNonNull(operator);
  class ReducingSink
    implements AccumulatingSink<T, Optional<T>, ReducingSink> {
    private boolean empty;
    private T state;

    public void begin(long size) {
      empty = true;
      state = null;
    }

    @Override
    public void accept(T t) {
      if (empty) {
        empty = false;
        state = t;
      } else {
        state = operator.apply(state, t);
      }
    }

    @Override
    public Optional<T> get() {
      return empty ? Optional.empty() : Optional.of(state);
    }

    @Override
    public void combine(ReducingSink other) {
      if (!other.empty)
        accept(other.state);
    }
  }
  return new ReduceOp<T, Optional<T>, ReducingSink>(StreamShape.REFERENCE) {
    @Override
    public ReducingSink makeSink() {
      return new ReducingSink();
    }
  };
}
```

可以发现，当我们对一个 Stream 挂载`reduce`操作时，实际上先构造了一个`ReduceOp`（实现了 `TerminalOp`）之后通过语句`terminalOp.evaluateSequential(this, sourceSpliterator(terminalOp.getOpFlags()))` 触发`TerminalOp`中的`evaluateSequential()`方法，其两个入参正好是当前组装好的流水线`this`以及输入源`sourceSpliterator()`（该方法返回一个`Spliterator`）。

那么我们具体来看一看`ReduceOp`中定义的``evaluateSequential()`：

```java
private abstract static class ReduceOp<T, R, S extends AccumulatingSink<T, R, S>>
  implements TerminalOp<T, R> {

  ... ...
		public abstract S makeSink();
    
    @Override
    public <P_IN> R evaluateSequential(PipelineHelper<T> helper,
                                       Spliterator<P_IN> spliterator) {
    return helper.wrapAndCopyInto(makeSink(), spliterator).get();
  }
}
```

结合前面`reduce`方法构造的匿名类中实现的`makeSink()`方法 `return new ReducingSink();`，我们得知其实`reduce`匿名类的真实作用是调用`makeSink`构造了一个`Sink`之后将之传入原 Stream 中，并调用流水线的`wrapAndCopyInto()`方法来启动整个流水线。

**Sink**

前文我们在构建中间操作时遇到过`Sink`但没有细说，终于，要到了解释`Sink`的时刻了。

从流水线的角度讲 ，不论是 `map`、`filter`、 `limit` 还是 `reduce`，都属于高层抽象概念，主要用于定义通用的工序行为。在 Stream 的实现中，采用`Head`、`StatelessOp`、`StatefulOp`和`TerminalOp`作为中间层抽象，主要用于对工序行为进行归纳分类。而最终落到低层实现上，靠的只有一种东西：`Sink`。

从`Sink`的定义来看，不论是什么样的流水线工序，Steam 都将其低层定义成了一个`Sink`，用来在流水线的各个阶段（stage）传递值。从名称中就很明确，`Sink`就像是一个个的水槽，互相之间首尾相连，数据流从第一个水槽漏到第二个，再漏到第三个，以此类推，只能向下漏，不能向上返。

{% asset_img sink.png %}

`Sink`具有两个状态：

- 初始态
- 激活态

三种基本的行为：

- begin: 在接收数据之前调用，并将`Sink`置为激活态
- accept：开始接收数据，并对数据进行处理
- end：数据处理完成后调用，并将`Sink`置为初始态

有了以上知识，结合前文`map`操作的实现：

```java
@Override
Sink<P_OUT> opWrapSink(int flags, Sink<R> sink) {
  return new Sink.ChainedReference<P_OUT, R>(sink) {
    @Override
    public void accept(P_OUT u) {
      downstream.accept(mapper.apply(u));
    }
  };
}
```

就很清晰了，就像前文图里画的一样，通过`opWrapSink`方法，把传入的`sink`进行一层包装，创建一个新的`Sink.ChainedReference`（有点像装饰器模式）。来看看`Sink.ChainedReference`的实现：

```java
abstract static class ChainedReference<T, E_OUT> implements Sink<T> {
  protected final Sink<? super E_OUT> downstream;

  public ChainedReference(Sink<? super E_OUT> downstream) {
    this.downstream = Objects.requireNonNull(downstream);
  }

  @Override
  public void begin(long size) {
    downstream.begin(size);
  }

  @Override
  public void end() {
    downstream.end();
  }

  @Override
  public boolean cancellationRequested() {
    return downstream.cancellationRequested();
  }
}
```

基于传入的`Sink`构造的`Sink.ChainedReference`，顾名思义就像一根链条一样，其`begin`和`end`方法都直接调用传入`Sink`的方法，在`map`的`opWrapSink`中，覆写了`accept()`方法，先执行`mapper`的逻辑，之后执行传入`Sink`的逻辑。

有趣的是，在`Sink.ChainedReference`中将传入的`Sink`命名为"downstram"，也就是下游，那么从实现上，`Sink`是从最底层开始，层层包装，层层向上构建的。而前文中我们提到，`Sink`就像是首尾相连的水槽，水（数据）只能向下流，不能向上流。正因为这一点，实际上`Sink`在组装时，是从下往上，才能满足运行时数据从上往下流动。

基于此，我们是否可以假设，类似`Stream.of("a", "b", "c", "1", "2", "3").filter(NumUtil::isNum).map(NumUtil::minusOne).limit(1).collect(Collectors.toList);`流水线的组装，从实现上其实是先从`collect`这一道`TerminlOp`工序为起点的？

从代码我们可以得知：确实是这样的，回顾一下`ReduceOp`的实现：

```java
private abstract static class ReduceOp<T, R, S extends AccumulatingSink<T, R, S>>
  implements TerminalOp<T, R> {

  ... ...
		public abstract S makeSink();
    
    @Override
    public <P_IN> R evaluateSequential(PipelineHelper<T> helper,
                                       Spliterator<P_IN> spliterator) {
    return helper.wrapAndCopyInto(makeSink(), spliterator).get();
  }
}

public static <T> TerminalOp<T, Optional<T>>
  makeRef(BinaryOperator<T> operator) {
  Objects.requireNonNull(operator);
  class ReducingSink
    implements AccumulatingSink<T, Optional<T>, ReducingSink> {
    ... ...
  }
  return new ReduceOp<T, Optional<T>, ReducingSink>(StreamShape.REFERENCE) {
    @Override
    public ReducingSink makeSink() {
      return new ReducingSink();
    }
  };
}
```

在`helper.wrapAndCopyInto(makeSink(), spliterator)`中第一次被`wrap`的`Sink`就是通过`makeSink()`方法生成出来的。

### 流水线的执行

