---
title: 关于浮点数与 IEEE 754
date: 2019-06-20 23:20:43
tags:
	- float
	- ieee754
category:
	- java
---

由于某些神秘的原因，某些理所当然的数值计算，通过编程语言操作时，会让人匪夷所思。

来看一个 Java 的例子：

``` java
@Test
public void floatCalculationTest() {
    System.out.println("a=" + 1.0f);
    System.out.println("b=" + 0.9f);
    System.out.println("a-b=" + (1.0f - 0.9f));
}
```

执行结束后，控制台会显示什么？

执行结果：

``` shell
a=1.0
b=0.9
c=0.8
a-b=0.100000024
b-c=0.099999964
```

单个数字拎出来打印都正常，但是运算后出现了很小的误差。这种情况在 JavaScript 中也很常见，我们经常会发现在 js 里做一些简单运算的时候不是我们想要的结果。

这一切的锅，都应该由 IEEE 754 来背。

### 什么是 IEEE 754
在 long long ago，计算机还没有普及的年代，如何用计算机