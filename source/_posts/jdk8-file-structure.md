---
title: JDK 8 文件目录结构
date: 2019-04-18 01:32:27
tags: java
---

> 下文翻译自 oracle Java SE 8 官方文档， 原文链接：https://docs.oracle.com/javase/8/docs/technotes/tools/windows/jdkfiles.html

### JDK 8 文件目录结构

本文介绍 JDK 的目录及其所包含的文件。JRE 的文件结构与 JDK 的 jre 目录下文件完全一致。

以下包含三个主题
- Demos 和示例
- 开发相关的文件及目录
- 其他文件及目录

#### Demos 和示例
Demos 和示例展示如何在 Java 平台开发程序，相关文件可在 Java SE 下载页面独立下载，详见链接：http://www.oracle.com/technetwork/java/javase/downloads/index.html

对应二进制文件我们提供了 .tar.z 和 .tar.gz 格式的压缩包。与其他的 Oracle Solaris 64位包类似，在 Oracle Solaris 下的 Demos 和示例包需要依赖安装其对应的32位包。

#### 开发相关的文件及目录
这部分主要描述在 Java 平台下开发应用程序所需要的最重要的文件和目录。部分目录下可能并没有包含 Java 源码和 C 头文件，这些目录的相关信息参见：其他文件及目录部分。

``` text
jdk1.8.0
     bin
          java*
          javac*
          javap*
          javah*
          javadoc*
     lib
          tools.jar
          dt.jar
     jre
          bin
               java*
          lib
               applet
               ext
                    jfxrt.jar
                    localdata.jar
               fonts
               security
               sparc
                    server
                    client
               rt.jar
               charsets.jar
```

假设 JDK 软件安装在 `/jdk1.8.0` 目录下，以下即为最重要的目录简介：

**/jdk1.8.0**
> JDK 软件安装的根目录。包括版权、许可证和 README 文件。同时也包含了 Java 平台的源代码副本 - src.zip。

**/jdk1.8.0/bin**
> JDK 中包含的所有开发工具的可执行文件。在 PATH 环境变量中应包含本目录的入口。

**/jdk1.8.0/lib**
> 由开发工具使用的相关文件。本目录中包括 tools.jar - 是用于支持 JDK 中实用工具的非核心类包。以及 dt.jar - 是 BeanInfo 文件的设计时（DesignTime）副本，可用于告知 IDE 如何显示 Java 组件以及开发者如何在应用程序中个性化定制他们。（注：dt.jar 主要用于帮助开发者更简便的开发 Swing 程序，因此就不难理解前文提到的 DeignTime （设计时）和我们熟知的 Runtime（运行时）的区别。）

**/jdk1.8.0/jre**
> JDK 开发工具依赖的 JRE（Java 运行时环境）的根目录。Java 运行时环境是 Java 平台的一种实现。系统属性 java.home 引用的即为本目录。

**/jdk1.8.0/jre/bin**
> Java 平台使用的工具及库的可执行文件。与前文 /jdk1.8.0/bin 目录中的文件完全一致。其中的 Java 启动器工具充当了应用程序启动器（用以替换先前随 JRE 1.1 版本附带的命令行工具）。本目录无需包含在 PATH 环境变量中。

**/jdk1.8.0/jre/lib**
> JRE 所需的代码库、属性配置和资源文件。例如 rt.jar 包含了启动类 - 压缩了 Java 平台核心 API 的运行时类，charsets.jar 包含了字符转换类。除了 ext 子目录外。还有一些额外的资源子目录未在这里描述。

**/jdk1.8.0/jre/lib/ext**
> Java 平台扩展组件的默认安装目录。例如 JavaHelp JAR 文件的安装位置等。该目录包含了 jfxrt.jar， 该 jar 中包含 JavaFX 的运行时库，还有 localedata.jar，包含了 java.text 和 java.util 的本地数据。 扩展机制的相关信息可见 http://docs.oracle.com/javase/8/docs/technotes/guides/extensions/index.html

**/jdk1.8.0/jre/lib/security**
> 安全管理相关文件的目录。包含安全策略文件 java.policy 和安全属性文件 java.security。

**/jdk1.8.0/jre/lib/sparc**
> Oracle Solaris 版本的 Java 平台所需的 .so 文件。

**/jdk1.8.0/jre/lib/sparc/client**
> 实现了 Java HotSpot VM 技术的客户端所需的 .so 文件。 Java HotSpot VM 是默认的 Java 虚拟机（JVM）。

**/jdk1.8.0/jre/lib/sparc/server**
> Java HotSpot VM 服务端所需的 .so 文件。

**/jdk1.8.0/jre/lib/applet**
> 包含 applet 支持类的 Jar 文件可存放于此目录。这种允许 applet 类在本地文件系统预加载的功能不仅能够缩短启动时间，也能像从互联网下载 applet 的方式提供一样的保护机制。

**/jdk1.8.0/jre/lib/fonts**
> 平台的字体文件。

#### 其他文件及目录
这部分描述 Java 源代码、C 头文件和其他目录及文件的目录结构。

``` text
jdk1.8.0
     db
     include
     man
     src.zip
```

**/jdk1.8.0/src.zip**
> Java 平台的源代码

**/jdk1.8.0/db**
> Java DB， 与之相关的技术内容参见http://docs.oracle.com/javadb/

**/jdk1.8.0/include**
> 用于支持 Java Native 接口和 JVM Debugger 接口的 C 头文件。Java Native 详见http://docs.oracle.com/javase/8/docs/technotes/guides/jni/index.html

> Java 平台调试架构 (JPDA) 的信息详见http://docs.oracle.com/javase/8/docs/technotes/guides/jpda/index.html

**/jdk1.8.0/man**

> JDK 工具的说明页。