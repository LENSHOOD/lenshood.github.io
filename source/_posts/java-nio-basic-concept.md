---
title: Java NIO - 基本概念
date: 2019-05-19 01:14:10
tags: 
- java
- nio
categories:
- Java
---

Java IO 系列：
[第一篇： Java IO - Stream 输入输出流](https://lenshood.github.io/2019/04/28/java-io-stream/)
[第二篇： Java IO - 字符流](https://lenshood.github.io/2019/05/07/character-stream/)
[第三篇： Java NIO - 基本概念](https://lenshood.github.io/2019/05/18/java-nio-basic-concept/)

<!-- more -->

---

### IO ？ NIO ？
在 Java 的早些年，IO 操作围绕着流 (Stream) 来实现，基于最基础的字节流，提供了字节流的输入、输出，并在此基础上提供了文件读写，Socket 网络，序列化等功能。

Java 1.4 以后，提供了 NIO，来对原先的 IO 进行更新。NIO 即 new io，是一组 IO 操作的全新实现。

相较而言，当我们操作流时，通常会与单个字节，或字节数组打交道。对读出的字节，Stream 相关的代码并未帮助我们做任何的存储，我们需要自行处理。同时，我们无法在流中前后移动，除非对读出的字节进行缓存。(BufferedXXX 相关类通过装饰器的形式替我们实现了缓存的功能)。

当我们采用 NIO 相关类操作时，Java 向我们提供了 Buffer 实现，所有对数据的读写全部都通过 Buffer 来进行，而不是直接与数据源进行交互。对 Buffer 的支持一直延伸到 Native 方法内，而不仅仅是通过某种装饰器来增强功能。此外，NIO 改善了 Socket，使之与 Selector 结合，支持非阻塞的操作形式，很大程度上提升了单机的服务能力。

### NIO 三大基础
1. Buffer
	顾名思义，Buffer 提供存储的能力，他对基础类型数组做了一层封装，提供了对数组进行读写、移动等能力。Buffer 定义了几种属性来支持实现其能力：
	- capacity: the number of elements it contains, >=0 , never change
	- limit: the index of the first element that should not be read or written, >=0, <= capacity
	- position: the index of the next element to be read or written, >=0, <= limit
	- mark: the index to which its position will be reset when call reset()
2. Channel
	Channel 作为替代 Stream 的一类概念，是由一组 Channel 接口，定义了 Channel 的各种不同行为。正如 Think in Java 中所述，Channel 与 Buffer 的关系，如同煤矿与矿车，Channel 就像是一座数据的煤矿，而我们想要从这座煤矿中取出原煤，需要将 Buffer 当做是矿车，空载而入，满载而出。
	相比采用 Stream 直接进行数据存取，Channel 的数据存取，都需要 Buffer 来配合实施。
3. Selector
	提供了与 Linux 的 select() 类似的多路选择器实现，用以进行非阻塞 IO 的支持。简要来讲，可以将某些处于非阻塞模式下的 Channel 注册到 Selector 中，并声明这个 Channel 所关心的事件，包括 READ，WRITE，ACCEPT，CONNECT。调用 Selector 的 select() 方法，Selector 会阻塞，同时一旦有任何被注册的 Channel 产生了其注册的事件，select() 返回，并给出相关的 SelectionKey，我们通过该 Key，即可获得对应的事件、Channel 等信息，并进行处理。
	通过 Selector 我们可以将传统的多线程处理阻塞 IO 的方式，转变为高效的单线程处理非阻塞 IO，这里的高效，建立在：
	- IO 操作等待的时间远大于 CPU 时间
	- 创建与销毁线程非常占用系统资源

借用 Think of Java 中的一张类图，可以对 NIO 更加清晰(图片非原创，若侵权请联系我删除)：
{% asset_img relationship-between-nio-classes.png %}

---

Java IO 系列：
[第一篇： Java IO - Stream 输入输出流](https://lenshood.github.io/2019/04/28/java-io-stream/)
[第二篇： Java IO - 字符流](https://lenshood.github.io/2019/05/07/character-stream/)
[第三篇： Java NIO - 基本概念](https://lenshood.github.io/2019/05/18/java-nio-basic-concept/)