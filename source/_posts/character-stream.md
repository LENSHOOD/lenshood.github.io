---
title: Java IO - 字符流
date: 2019-05-08 00:31:14
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

### Reader/Writer
在前文[Java IO - Stream 输入输出流](https://lenshood.github.io/2019/04/28/java-io-stream/)中，介绍了 Java IO 包内相关类对字节流的定义与处理。

通常字节并不适合人类阅读，为了方便阅读，计算机领域采用字符集（[Character encoding](https://en.wikipedia.org/wiki/Character_encoding)）来定义不同字节组合的含义。

因此，在 IO 包中也提供了一组专门用于处理字符流的类，即 Reader/Writer。

Reader/Writer 的设计思想、实现与 IOStream 完全一致，主要区别在于 IOStream 处理字节，Reader/Writer 处理字符。具体的用法可以参考源码实现。

### Encoding/Decoding
本节主要讨论字节-字符转换对性能的影响。 

在 Reader/Writer 中，除了装饰器实现外，InputStreamReader 与 OutputStreamWriter 专用于将字节流转换为字符流。

具体实现中，转换通过 StreamDecoder/StreamEncoder 实现，StreamDecoder 继承自 Reader，因此实现了 read 方法。在其内部，真正用于转码的是 CharsetDecoder，这是一个抽象类，不同的编码格式实现了不同的 decoder。

encode/decode 本质上是一种查表操作，根据不同的字符集规则，将字节映射为字符。查表映射本身是一项较为耗时的工作，若在大量数据条件下，无意义的字符转换很容易拖慢系统。

以下通过简单的例子展示编解码操作对性能的影响：
``` java
public class ByteStreamAndCharCompareDemo {
    
    public static void main(String[] args) throws NoSuchAlgorithmException {
        ByteStreamAndCharCompareDemo demo = new ByteStreamAndCharCompareDemo();

        // 1MB data
        int capacity = 1024 * 1024 * 1024;
        byte[] bytes = demo.buildBigByteArray(capacity);

        // read byte stream cost
        long timeStart = System.currentTimeMillis();
        demo.consumeInputStreamFromArray(bytes);
        long timeEnd = System.currentTimeMillis();
        System.out.println("Build byte stream and read out cost: " + (timeEnd - timeStart));

        // read char stream cost
        timeStart = System.currentTimeMillis();
        demo.consumeReaderFromArray(bytes);
        timeEnd = System.currentTimeMillis();
        System.out.println("Build char stream and read out cost: " + (timeEnd - timeStart));
    }

    public byte[] buildBigByteArray(int capacity) throws NoSuchAlgorithmException {
        byte[] bytes = new byte[capacity];

        Random random = SecureRandom.getInstanceStrong();
        for (int i=0; i<capacity; i++) {
            int v = random.nextInt(255);
            bytes[i] = ((byte)v);
        }

        return bytes;
    }

    public void consumeInputStreamFromArray(byte[] bytes) {
        try(ByteArrayInputStream byteArrayInputStream = new ByteArrayInputStream(bytes)) {
            while (byteArrayInputStream.read() != -1);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    public void consumeReaderFromArray(byte[] bytes) {
        try(InputStreamReader inputStreamReader = new InputStreamReader(new ByteArrayInputStream(bytes))) {
            while (inputStreamReader.read() != -1);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
```

以上简单的例子通过构造一个 1MB 大小的字节数组，分别将之转换为 `ByteArrayInputStream` 与 `InputStreamReader` 并全部读出，计算其花费时间，程序输出可见：

``` shell
Build byte stream and read out cost: 2211
Build char stream and read out cost: 113993
```

显然，转换为字符后读取的时间是直接读取字节的 51.6 倍。

在性能优化中，编解码、序列化都属于很耗时的操作通常会被优先考虑优化。很多时候，REST API 响应中会提供对状态、信息等的描述，描述大都采用 String 类型，因此在调用的序列化过程中，势必存在多次的编码/解码操作。因此，若这类 REST API 用于微服务内部调用，而非与人交互，则相关的描述完全可以采用错误码替代，这样就减少了一定的编解码开销，对于高频调用效果较为显著。

---

Java IO 系列：
[第一篇： Java IO - Stream 输入输出流](https://lenshood.github.io/2019/04/28/java-io-stream/)
[第二篇： Java IO - 字符流](https://lenshood.github.io/2019/05/07/character-stream/)
[第三篇： Java NIO - 基本概念](https://lenshood.github.io/2019/05/18/java-nio-basic-concept/)