---
title: (译文)响应式编程笔记： Part 1 响应式介绍
date: 2019-05-23 21:16:24
tags:
- reactive
categories:
- Reactive
- Software Engineering
---

> 本文非原创，是对英文原文的译文，原文请见：[Notes on Reactive Programming Part I: The Reactive Landscape](https://spring.io/blog/2016/06/07/notes-on-reactive-programming-part-i-the-reactive-landscape#reactive-use-cases)

响应式编程是一种非常有趣的编程思想，目前对于响应式编程，存在诸多的文章、杂谈，然而对于局外人或是简单企业项目的 Java 开发者而言（比如笔者），这些内容并不都容易理解。本文（系列第一篇）没准能帮你理清这些杂乱，文章的内容已经尽可能的具体，绝不会出现指代语义的情况。当然，如果你想要的是更加学术的论述以及 Haskell 语言的代码示例，那就 Google 一下吧，本文并不涉及这些。

响应式编程经常与并发编程、高性能等概念相混淆，以至于（to such an extent）难以将这些概念分清，实际上在原理上他们完全不同。不可避免的，这肯定会导致混乱。响应式编程经常与函数式反应编程（FRP）相互交融（或直接就被称作是 FRP）。一些人认为响应式编程没什么新奇的，他们每天都这么干（这些人通常都使用 JavaScript 进行开发）。另一些人认为，响应式编程是微软带给人间的礼物（在先前他们发布了一些 C# 的 extension 时引起了巨大的轰动）。而在 Java 企业应用领域，近期已经有了一些躁动（见http://www.reactive-streams.org/）就像任何其他的新生事物一样，在何时何地使用的问题上，还存在许多容易犯的错误。

###What is it?
响应式编程是一种将智能路由和事件消费相结合来改变行为的微架构设计。以上是其中一种定义，互联网上还有许多种其他的定义，我们试图构造一些更加具体的理解来解释响应式的概念，以及响应式为何重要。

最初的响应式编程也许会追溯到 1970 年代或更早，因此在概念上没什么新奇，但它的确与现代企业的一些东西产生了共鸣。这种共鸣与微服务的兴起、多处理器的普及共同产生。

以下是一些来自其他地方的有用定义：

> 响应式编程的背后思想是，存在一类随时间变化类型的数据值。计算这类随时间变化的值本身也具有随时间变化的特点。

> 认识响应式编程的第一直觉，可以通过一个简单的办法 - 设想程序是一张表单，而每一个单元格都是一个变量。任一单元格发生了变化，与之相关的单元格的值也会随即变化。这一点与 FRP 是相同的。接下来设想其中的一些单元格会自己发生改变（更甚者可以说，是受到外部世界的改变）：在 GUI 环境中，鼠标的位置变化就是一个很好的例子。

（以上来自 Terminology Question on Stackoverflow https://stackoverflow.com/questions/1028250/what-is-functional-reactive-programming）

FRP 与高性能、并发、异步操作和非阻塞 IO 有很强的相关性。然而，从以 FRP 与上述任一概念都无关作为假设来开始学习可能会更有帮助。诚然，采用响应式模型时，以上的问题都可以自然且透明的被处理，用户无需过多关心。但是讨论如何有效或高效的处理这些问题更为有益（因此这部分也更应该仔细对待）。以同步、单线程的方式实现来一个 FRP 框架也是完全可行的，但这在尝试任何新工具或库时都不太可能有帮助。

###反应式使用场景
对一个新手而言，最难以找到答案的问题应该是：这是做什么用的？以下的一些例子都是在企业环境下的一些通用模式：
- 外部服务调用
目前许多后端服务都是 RESTFful 的（基于 Http 协议），因此从基础上即是阻塞且同步的。这看起开可能没有给 FRP 留下多少发挥空间，实际上，由于采用 RESTful 的服务通常都会调用其他服务，并且会有更多的服务依赖第一个服务调用的结果，因此 FRP 大有可为。在存在大量 IO 操作的场景下，在等待一个请求返回从而能够发送下一个请求的时候，你可怜的客户端很可能在你将所有请求结果组装并返回之前，就因为阻塞而中断了。因此，外部服务调用，尤其是存在复杂依赖关系的调用，有很大的优化空间。FRP 提供了最终可组合性承诺的逻辑来驱动这些操作，因此对于开发者，编写调用服务的代码会变得更容易。

- 高并发消息消费者
消息处理，尤其是具有高并发特点的消息处理是企业应用的一个常见场景。响应式框架都喜欢通过微基准测量来计算并吹嘘采用其方案的 JVM 每秒处理消息的数量。通常这一结果令人震惊（每秒数千万的处理量很常见），但是这一结果也可能是人造的，如果他们说测量结果是通过运行简单的 for 循环来实现的，你就不会感到震惊了。然而我们也不应该过早的放弃这类尝试，毕竟当性能很重要的时候，任何可能的尝试都会被接受。响应式模型天然就适合做消息处理（因为事件很容易转换成消息），因此如果有更快的方法来处理消息，就应该引起关注。

- 电子表格
这可能不算是一个企业应用，但是在企业活动中与每个人都相关。电子表格应用很好的抓住了 FRP 原理和实现上的困境。如果单元格 B依赖单元格 A，而单元格 C 同时依赖 B 和 A，那么如何才能确保 A 的变化能在 B 收到任何改变事件之前传递到 C？如果采用真正的响应式框架来构建，那么答案就是：你无需关心，你只需要声明出这种依赖即可。简而言之，这正是电子表格真正的威力。这种方式也指明了 FRP 与一般的事件驱动框架的区别：FRP 将智能应用于智能路由的概念中。

- 同步/异步处理的抽象
这更像是一种抽象的场景，因此我们应避免误入歧途。这种场景也与前述的更具体的场景存在较大差异，但我们也期望它在一些讨论中有价值。基本上，只要开发人员愿意接受一个额外的抽象层，那么他们就可以忘记他们调用的代码是同步还是异步的。由于处理异步编程需要耗费宝贵的脑细胞，因此这里提供一些有用的想法。响应式编程并不是处理此类问题的唯一方法，但是一些 FRP 的实现者已经足够认真的考虑了这一问题，并且他们提供的工具很有用。

Netflix 的技术博客里有很多具体且有用的场景实例，详见：Netflix Tech Blog: Functional Reactive in the Netflix API with RxJava https://medium.com/netflix-techblog/reactive-programming-in-the-netflix-api-with-rxjava-7811c3a1496a

### 对比
只要你不是生活在上世纪 70 年代的洞穴人，那么一定遇到过与响应式编程相关的概念以及人们人们期望用他们来解决的一些问题。下面就是一些我个人认为与之相关的一些方案。

- Ruby Event-Machine
Event Machine（事件机 https://github.com/eventmachine/eventmachine）是对并发编程的一类抽象（通常涉及非阻塞 IO）。Ruby 开发者们长期被一个很头疼的问题所困扰：如何将一个被设计成是单线程脚本语言的编程语言转换为某种能够开发可用、性能好、负载下持续稳定的服务器应用的语言。很长一段时间 Ruby 都支持线程，但是由于性能差的坏名声，很少被采用。取而代之的是在目前已经成为语言核心代码，普遍可用 （Ruby 1.9 以后的新功能）的 Fibers 功能（https://www.ruby-doc.org/core-1.9.3/Fiber.html）。Fiber 编程模型有点类似于 coroutines 的意思，单个核心线程能够处理大量的并发请求（与 IO 相关）。这种编程模型比较抽象且难以理解，因此开发者更愿意使用它的 Wrapper 来进行开发，而 Event-Machine 正是最常见的一种。Event-Machine 并不需要使用 Fiber（只是抽象了核心关注点），但在 Ruby Web App 中很容易找到采用 Event-Machine 和 Fiber 开发的例子（见 https://www.igvita.com/2009/05/13/fibers-cooperative-scheduling-in-ruby， https://github.com/igrigorik/em-http-request/blob/master/examples/fibered-http.rb）。人们采用 Event-Machine 来优化 IO 密集型应用的可伸缩性，来替代存在大量嵌套回调的丑陋编程模型。

- Actor Model
与 OOP 类似，演员模型来自上世纪 70 年代 CS 领域的概念。Actors 提供了对于计算（computation）的一层抽象（与之对应的是数据与行为层），这层抽象能够将并发性作为一个通用结果，因此在实际操作时，这就能够形成并发系统的基础。Actors 给彼此之间发送消息，因此在某些情况下他们是响应式的，同时对于一些类似 Actors 或是响应式的系统他们在自己的设计上也有许多重叠。通常只有在实现上才会存在差别（例如 Akka（https://doc.akka.io/docs/akka/current/?language=java） 的一个显著的特征是其 Actors 能够跨进程分发）

- Deferred results (Futures)
Java 1.5 版本引入了大量新库，包括 Doug Lea 的 java.util.concurrent 并发包，这其中的一部分引入了延迟结果的概念，并被封装为 `Future`。这是一个在异步模型上进行简单抽象的好例子，他不强制要求相应的实现也必须为异步或特定的异步模型。在 [Netflix Tech Blog: Functional Reactive in the Netflix API with RxJava](https://medium.com/netflix-techblog/reactive-programming-in-the-netflix-api-with-rxjava-7811c3a1496a) 中很好的展示了，`Future`在处理相似任务的并发时很好用，然而一旦这些任务之间存在任何依赖或是条件执行的情况，那么开发者就很容易落入"嵌套回调地狱"。而响应式编程是解决这类问题的一剂良方。

- Map-reduce and fork-join
对并行处理的抽象非常有用，并且能找到许多实践供选择。受到大规模并行分布式处理和 JDK 1.7 自身的驱动，[Map-Reduce](http://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf) （[Hadoop](https://wiki.apache.org/hadoop/MapReduce)）和 [Fork-join](http://gee.cs.oswego.edu/dl/papers/fj.pdf) 最近在 Java 世界里不断的演进。以上抽象都很有用，但是相比于即可被用于简单并行处理抽象，又可被扩展为可组合、可声明通信的 FRP，还是略为浅显。

- Coroutines(协程)
[Coroutine](https://en.wikipedia.org/wiki/Coroutines) 是 Subroutine 的一种概括 - 它像 Subroutine 一样包含入口点和出口点，但当退出时，它将控制权传递给另一个 Coroutine (不一定是传给调用者)，同时，不论 Coroutine 积累了怎样的状态，都会持续保存到下一次调用时。Coroutine 可以作为更高级功能（Actor 或 Stream）的基础模块。响应式编程的其中一个目标就是并行处理 Agent 的通信上提供相同的抽象，因此 Coroutine (如果可以使用的话) 将是一个有用的基础模块。Coroutine 存在各种形式，有一些相比一般情况下更严格，但仍然比普通的 Subroutine 更灵活。Fiber(参见 Event-Machine 的讨论) 是一种风格，而 Generator（在 Python 和 Scala 中更常见）则是另一种风格。

### Java 中的响应式编程
Java 本身并不是一种"响应式的语言"，因为其原生并不支持 Coroutine。有一些运行在 JVM 上的语言（Scala，Clojure）原生支持响应式模型，但 Java 不是，至少在 Java 9 之前不是。然而，响应式编程在企业开发中需求旺盛，因此已经有一些在 JDK 之上提供响应式层的尝试非常活跃。以下我们简略的看一看其中的几种。

[Reactive Streams](http://www.reactive-streams.org/) 是一种非常底层的契约，表现为少量的 Java interface(加上 TCK)，但也适用于其他语言。这些 interface 表示为背压式的 `Publisher` 和`Subscriber` 的基本构造块，形成通用语言可调用的库。Reactive Stream 已经在 Java 9 中被纳入 JDK，名为`java.util.concurrent.Flow`。该项目由来自 Kaazing, Netflix, Pivotal, Red Hat, Twitter, Typesafe 等多个组织的工程师共同合作维护。

[RxJava](https://github.com/ReactiveX/RxJava/wiki)：Netflix 曾经在内部使用的响应式模型，后来他们将之发布为基于开源许可的 [Netflix/RxJava](https://github.com/ReactiveX/RxJava/wiki) (随后被重命名为 "ReactiveX/RxJava")。Netflix 在 RxJava 上开发了大量基于 Groovy 的代码，但他对 Java 使用开放，非常适合采用 Java 8 的 Lambda 进行开发。这里有一种 [对 Reactive Stream 的适配方案](https://github.com/ReactiveX/RxJavaReactiveStreams)。根据 David Karnok 的[响应式代际分类](https://akarnokd.blogspot.co.uk/2016/03/operator-fusion-part-1.html)，RxJava 属于第二代响应式库。

[Reactor](https://projectreactor.io/) 是一种来自 Pivotal 开源团队（Spring 的创造者）的 Java 框架。因为他直接基于 Reactive Stream 开发，因此无需任何适配。Reactor IO 项目还提供了对底层网络运行时(如 Netty 或 Aeron) 的包装。根据 David Karnok 的[响应式代际分类](https://akarnokd.blogspot.co.uk/2016/03/operator-fusion-part-1.html)，Reactor 属于第四代响应式库。

[Spring Framework 5.0](https://projects.spring.io/spring-framework/)(在 2016 年 6 月发布了第一个里程碑) 内建了响应式特性，其中包括构建 Http 服务端和客户端的工具。在 Web 层中已经使用 Spring 的用户会发现，由于大部分的对响应式请求进行分发及背压的工作都交给了框架，因此他们可以直接通过对 `controller`进行注解装饰这种熟悉的编程模型来处理 Http 请求。虽然基于 Reactor 构建，Spring 仍然开放了相关 API 来允许其特性采用其他可选的库来进行开发(例如 Reactor 和 RxJava)。用户可以选择从 Tomcat，Jetty， Netty(通过 Reactor IO) 以及 UnderTow来作为服务端的网络栈。

[Ratpack](https://ratpack.io/) 是一系列用于构建高性能 Http 服务的库。他基于 Netty 构建并在内部采用Reactive Stream 实现(所以你可以在更高层使用其他的 Reactive Streams 实现)。Spring 作为原生组件被支持，同时可以通过几个简单的工具类来提供依赖注入。同时，因为还包含了 autoconfiguration，因此 Spring Boot 用户可以直接将 Ratpack 内嵌于 Spring 应用中，作为 Http 端点进行监听，来替换 Spring Boot 默认使用的内嵌服务器。

[Akka](http://akka.io/) 是一套用于通过 Java 或 Scala 实现 Actor 模式来开发应用程序的工具套件，它使用 Akka Stream 来进行进程间通信，并内建了 Reactive Streams。根据 David Karnok 的[响应式代际分类](https://akarnokd.blogspot.co.uk/2016/03/operator-fusion-part-1.html)，Reactor 属于第三代响应式库。


