---
title: Java IO - Stream 输入输出流
date: 2019-04-28 23:15:32
tags: 
- java
- io stream
categories:
- Java
---

Java IO 系列：
[第一篇： Java IO - Stream 输入输出流](https://lenshood.github.io/2019/04/28/java-io-stream/)
[第二篇： Java IO - 字符流](https://lenshood.github.io/2019/05/07/character-stream/)
[第三篇： Java NIO - 基本概念](https://lenshood.github.io/2019/05/18/java-nio-basic-concept/)

<!-- more -->

---

Java io 包用于操纵数据的输入输出，其主要采用流的概念（stream）为核心，通过装饰器模式（decorator）实现了对流的多种操作与功能。

在 IO 包中，可以看到里面存放了各种各样的类，命名包括 **xxxInputStream xxxOutputStream xxxReader xxxWriter** 等等，以上四种类正好对应了 io 中的四种基本概念，除此之外，在 io 包中还包含了一些为了配合上述四种类所创建的工具、实体、异常等。

#### InputStream OutputStream

这是 io 中最为基础也是最为底层的两个接口，分别代表输入流，输出流。

请注意流的概念，在 io 中，流可以认为是字节流，即一串有限/无限的字节集合，对于 Java 程序而言，InputStream 即为该字节流的入口，而 OutputStream 则为字节流的出口。

java 8 之前并没有专门的定义流（java 8 的 Stream 也与本文所述的概念不同，而是通过 InputStream 和 OutputStream 两个 interface 定义了流的行为，最基本的 read 和 write 行为，可以对流进行读写，读写的载体是字节（interface 中以 int 表示）因此称字节流。

可以用水流来作一类比，家里的水流来自水龙头，用来洗手、擦地之后，又被倒入下水管，那么在家庭范围内，水龙头就是水流的入口，下水管就是水流的出口。在当今数据时代，流的概念已经被极大的扩展了，无论是字节、字符、事件、消息都可以被抽象为流，流式处理也成为了数据处理的一大重要分支。

#### Stream Family
首先，在 io 包中，对 InputStream 和 OutputStream 的实现有许多个，其中主要包含：
(由于每一种具体实现都对应了 InputStream 和 OutputStream 两个实现类，因此以下仅标明其前缀，后缀的 XXX = InputStream/OutputStream)
- BufferedXXX
- ByteArrayXXX
- DataXXX
- FileXXX
- FilterXXX
- ObjectXXX
- PipedXXX
- PushbackInputStream
- SequenceInputStream

以上 9 种流实现，看着多，实际上，又可以分成两类：

1. 端点适配类：专用于在其他类型的数据与 Java stream 之间进行适配与转换，包括：
	- ByteArrayXXX：将字节数组与流进行相互转换
	- FileXXX：通过 native 方法将文件与流进行相互转换

2. 装饰类：专用于操作流，通过装饰器模式，对 Java stream 进行额外的操作，包括：
	- BufferedXXX：以缓存的方式更高效的读写流
	- DataXXX：采用各类基本类型 (primitive type) 对流进行读写
	- FilterXXX：基础装饰器类
	- ObjectXXX：用于序列化与反序列化
	- PipedXXX：流管道，用于在线程间建立通信管道
	- PushbackInputStream：提供回退机制的流实现，便于在上下文中操作
	- SequenceInputStream：多个流合并为单个流

为何这样区分？

仍旧以水流的例子来说明。上文提到，家庭用水的输入端是水龙头，输出端是下水道，那么假如向上一个层面，对于整个楼栋来说，输入端就变为自来水进水阀，输出端变成了污水下水道。再向上，水汽遇冷凝结成水，顺着河流，进入自来水厂，再进入家庭，由下水管再回到河流，最终再次变成水汽。

相对应的，端点适配类可类比为水汽和水相互转化的部分，前后的状态不一致；装饰类可以类比为河水净化为自来水，自来水被加压泵入楼栋，前后都是水，只是对其做了不同的处理。

#### Decorator Stream
FilterXXX 作为BufferedXXX,DataXXX, PushbackInputStream的父类，为其提供了基本的装饰器封装，并实现了所有接口功能（伪实现，原封不动的调用了被封装流的相关方法）。因此通过多态的方式，其各种子类只需要实现对应的特定功能即可。

以下以 BufferedInputStream 为例，简单介绍装饰器模式的应用。

`public class BufferedInputStream extends FilterInputStream {`

BufferedInputStream 继承自 FilterInputStream，其实现了基本的装饰器模式，将一个 InputStream 作为构造器参数，通过一个 protected field 来持有该输入流。

BufferedInputStream 在调用父类构造器的同时，初始化了内部缓冲区：
``` java
public BufferedInputStream(InputStream in, int size) {
    super(in);
    if (size <= 0) {
        throw new IllegalArgumentException("Buffer size <= 0");
    }
    buf = new byte[size];
}
```

BufferedInputStream 覆写的 read()：
``` java
public synchronized int read() throws IOException {
        if (pos >= count) {
            fill();
            if (pos >= count)
                return -1;
        }
        return getBufIfOpen()[pos++] & 0xff;
    }
```

fill() 中一次性预读取了固定长度的字符存入 buffer，在此之后所有不超出 buffer 所包含范围的字符都从 buffer 中读取，避免了频繁的 io 操作。其持有输入流的真正 read() 正是在 fill() 中才被调用。

对比来看 FilterInputStream 的 read()：
``` java
public int read() throws IOException {
        return in.read();
    }
```
可见只是简单地调用了持有输入流自身的 read()。因此对于扩展了缓存功能的 BufferedInputStream，其通过装饰器模式，对原输入流的 read() 进行了装饰，增加了缓存的能力。

---

Java IO 系列：
[第一篇： Java IO - Stream 输入输出流](https://lenshood.github.io/2019/04/28/java-io-stream/)
[第二篇： Java IO - 字符流](https://lenshood.github.io/2019/05/07/character-stream/)
[第三篇： Java NIO - 基本概念](https://lenshood.github.io/2019/05/18/java-nio-basic-concept/)