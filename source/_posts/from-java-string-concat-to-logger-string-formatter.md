---
title: 从 Java 字符串连接看 Logger 的 String formatter
date: 2019-02-20 20:57:19
tags: java string-concat logger formatter
---

### Logger.info() 在 debug level 下会对代码效率产生怎样的影响？
在代码中打 log 是开发人员用于获取运行状态信息的常用手段，回顾一下 log level：
>- Trace - Only when I would be "tracing" the code and trying to find one part of a function specifically.
>- Debug - Information that is diagnostically helpful to people more than just developers (IT, sysadmins, etc.).
>- Info - Generally useful information to log (service start/stop, configuration assumptions, etc). Info I want to always have available but usually don't care about under normal circumstances. This is my out-of-the-box config level.
>- Warn - Anything that can potentially cause application oddities, but for which I am automatically recovering. (Such as switching from a primary to backup server, retrying an operation, missing secondary data, etc.)
>- Error - Any error which is fatal to the operation, but not the service or application (can't open a required file, missing data, etc.). These errors will force user (administrator, or direct user) intervention. These are usually reserved (in my apps) for incorrect connection strings, missing services, etc.
>- Fatal - Any error that is forcing a shutdown of the service or application to prevent data loss (or further data loss). I reserve these only for the most heinous errors and situations where there is guaranteed to have been data corruption or loss.

以上 Level 根据级别依次向下包含，即 Trace Level 可打印包括 Trace 在内的所有级别的信息，而 Fatal Level 只打印 Fatal 级别的信息。

那么，若有以下代码：
``` java
public void printWithPlusOperation(String msg) {
    logger.info("Got input message: " + msg);
}

public void printWithLoggerFormat(String msg) {
    logger.info("Got input message: {}", msg);
}
```

哪一个效率更高？
以下为反编译的 ByteCode：
```
public void printWithPlusOperation(java.lang.String);
    Code:
       0: aload_0
       1: getfield      #4                  // Field logger:Lorg/slf4j/Logger;
       4: new           #5                  // class java/lang/StringBuilder
       7: dup
       8: invokespecial #6                  // Method java/lang/StringBuilder."<init>":()V
      11: ldc           #7                  // String Got input message:
      13: invokevirtual #8                  // Method java/lang/StringBuilder.append:(Ljava/lang/String;)Ljava/lang/StringBuilder;
      16: aload_1
      17: invokevirtual #8                  // Method java/lang/StringBuilder.append:(Ljava/lang/String;)Ljava/lang/StringBuilder;
      20: invokevirtual #9                  // Method java/lang/StringBuilder.toString:()Ljava/lang/String;
      23: invokeinterface #10,  2           // InterfaceMethod org/slf4j/Logger.info:(Ljava/lang/String;)V
      28: return

  public void printWithLoggerFormat(java.lang.String);
    Code:
       0: aload_0
       1: getfield      #4                  // Field logger:Lorg/slf4j/Logger;
       4: ldc           #11                 // String Got input message: {}
       6: aload_1
       7: invokeinterface #12,  3           // InterfaceMethod org/slf4j/Logger.info:(Ljava/lang/String;Ljava/lang/Object;)V
      12: return
```
结果清晰可见：使用“+”操作符进行字符串拼接的 `printWithPlusOperation`方法在`logger.info()`之前先创建了 StringBuilder 实例，并将 "Got input message: " 与 msg 参数进行了 append，这是典型的 Java 编译优化。

对于`printWithLoggerFormat`方法，只是简单的将字符串常量传入`logger.info()`。

实际上，以 logback 为例，若当前日志级别较打印语句指定级别更高（在 Error Level 下打印 Info），则该字符常量不会与参数进行拼接动作，而是 `logger.info()`直接返回。

因此，在日志打印被屏蔽的情况下，采用日志拼接字符串的方式，更节省资源，相较之下效率也更高。

#### "Preconditions" and logging arguments should not require evaluation

以上所述内容，正是 Sonar 检查中的一个 issue(squid: S3457)，原文请见：`https://rules.sonarsource.com/java/RSPEC-2629`
