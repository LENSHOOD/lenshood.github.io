---
title: 关于浮点数与 IEEE 754
date: 2019-06-20 23:20:43
tags:
	- float
	- ieee754
category:
	- Java
---

由于某些神秘的原因，某些理所当然的数值计算，通过编程语言操作时，会让人匪夷所思。也是因为这些神秘的原因，业务中常见的集星星、代币值、金额计算等场景中，有可能会出现一长串和期望值有微小偏差的数值（尤其是前后端传递数值的时候..）

来看一个 Java 的例子：

``` java
@Test
public void floatCalculationTest() {
    System.out.println("a=" + 1.0f);
    System.out.println("b=" + 0.9f);
    System.out.println("c=" + 0.8f);
    System.out.println("a-b=" + (1.0f - 0.9f));
    System.out.println("b-c=" + (0.9f - 0.8f));
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

我们大都能想到，这是由于小数在十进制与二进制的转换过程中可能会出现无法收敛的情况，因此转换时必须要进行截断，截断代表精度的丢失，因此就会有微小的误差。

有句知乎名言：先问是不是，再问为什么。那么，上述现象到底是不是这样的原因？如果是，具体是怎么产生的？

这一切的一切，都应该从 [IEEE 754](https://en.wikipedia.org/wiki/IEEE_754) 说起。

### 什么是 IEEE 754
在 long long ago，计算机还没有普及的年代，采用计算机进行浮点数运算，由于没有统一的标准，各家都有自己的实现方法，导致兼容性差，也不可靠。鉴于此，IEEE (读作 I triple E) 在 1985 年，提出了一种浮点数存储、运算的标准，来规范计算机的浮点运算，这项标准的编号即为 IEEE 754。后来 IEEE 754 被写入了 ANSI 标准，因此绝大多数计算机都支持该标准。

那么 IEEE 754 究竟定义了些什么呢？

最简单的一句话，IEEE 754 定义了：**在计算机中浮点数应采用科学计数法的形式，存储于固定长度的存储单元中。**

#### 存储格式
IEEE 754 中规定了三种二进制浮点数的样式，分别对应了 32bit、64bit、128bit 长度的编码。目前最常用的两种即：
- single float 单精度浮点数，32bit
- double float 双精度浮点数，64bit
这两种格式的编码结构如下：
{% asset_img float-format.png %}

可见，在存储格式上，规范采用了 \[符号位 sign\] \[指数位 exponent\] \[小数位 fraction\] 三部分来表示，其中
- single float 包括 { 1bit sign |  8bit exponent | 23bit fraction }
- double float 包括 { 1bit sign |  11bit exponent | 52bit fraction }

#### 换算方法
IEEE 754 采用科学计数法来表示浮点数，其换算过程通常分为两步：
1. 将十进制数转换为二进制数：
	- 整数部分直接转为二进制
	- 小数部分在转换无法收敛时选取适当长度进行截断
2. 对上一步的二进制数进行规范化(Normalized)
	- 先根据正负确定符号位
	- 对小数点进行移位，以确保二进制数按照科学计数法：`1.xxxxx * 2^n` 的形式表示。
	- 将科学计数中的指数位 n 取出，作为存储格式中第二部分，为了避免考虑 n 存在正负值的情况，实际写入时给 n 加一个偏移量，单精度浮点数偏移量为 127，双精度浮点数偏移量为 1023。
	- 将科学技术中的小数位取出，单精度取 23bit，双精度取 52bit，并根据第 24bit(53bit) 的值进行舍入，舍入规则为：若第 24bit(53bit) 为 0，则直接舍弃；若为 1，假如之后位全为零则舍弃，不全为零则进位)

以下举例说明上述转换过程：
**eg. 754.01321**
```
1. 转换二进制数（整数与2取余倒序排列；小数与2取整正序排列。更具体的转换方法不在本文范围，请自行查阅资料）
	754 = 1011 1100 10
	0.01321 = 0000 0011 0110 0001 1011 1011 0000 0101 1111 1010 1110 1011 1100 0100 0000 1......(不收敛)
	因此，754.01321 = 1011110010.0000001101100001101110110000010111111010111010111100010000001......
	
2. 规范化
	以单精度浮点数为例，
	a. 754.01321 是正数，sign = 0
	b. 对小数点进行移位，可以得到 
		754.01321 = 1.0111100100000001101100001101110110000010111111010111010111100010000001...... * 2^9
		将指数位 e=9 取出，加偏移量 127 得：exponent = 9 + 127 = 136 = 1000 1000
	c. 将科学计数法表示的数字小数部分取 23bit 作为规范化数的小数部分，注意第 24bit 是 0，因此舍弃不进位，得到：0111 1001 0000 0001 1011 000
	
	最后拼装在一起，754.01321 遵循 IEEE 754 规范的单精度值为：
	0 | 1000 1000 | 0111 1001 0000 0001 1011 000
	即：443c80d8(hex)
	
	类似的算法得到双精度值为：
	0 | 1000 0001 000 | 0111 1001 0000 0001 1011 0000 1101 1101 1000 0010 1111 1101 0111
	即：4087901b0dd82fd7(hex)
	
在 Java 中简单验证一下：
	Integer.toHexString(Float.floatToIntBits(754.01321f))
	Long.toHexString(Double.doubleToLongBits(754.01321))
得到： 
	443c80d8
	4087901b0dd82fd7 
再次转换回十进制数后：
	754.01318359375
	754.0132099999999581996235065162181854248046875
可见确实存在由于小数转换导致的误差
```

**eg. 0.1072**
```
1. 转换二进制数
	0 = 0
	0.1072 = 0001 1011 0111 0001 0111 0101 1000 1110 0010 0001 1001 0110 0101 0010 1011...
	
2. 规范化
	a. sign = 0
	b. 对小数点进行移位后，可以得到 
		1.1011 0111 0001 0111 0101 1000 1110 0010 0001 1001 0110 0101 0010 1011... * 2^-4
		exponent = -4 + 127 = 123 = 0111 1011
	c. 第 24bit 是 0，不进位，得到：1011 0111 0001 0111 0101 100
	
	最后，0.1072 遵循 IEEE 754 规范的单精度值为：
	0 | 0111 1011 | 1011 0111 0001 0111 0101 100
	即：3ddb8bac(hex)
	
	类似的算法得到双精度值为：
	0 | 0111 1111 011 | 1011 0111 0001 0111 0101 1000 1110 0010 0001 1001 0110 0101 0011
	即：3fbb71758e219653(hex)
	
在 Java 中简单验证一下：
	Integer.toHexString(Float.floatToIntBits(0.1072f))
	Long.toHexString(Double.doubleToLongBits(0.1072))
得到： 
	3ddb8bac(hex) 0.1071999967098236083984375(dec)
	3fbb71758e219653(hex) 0.10720000000000000361932706027801032178103923797607421875(dec)
```

此外，IEEE 754 还规定了一些特殊情况：

- Zero
Sign bit = 0; biased exponent = all 0 bits; and the fraction = all 0 bits;

- Positive and Negative Infinity
Sign bit = 0 for positive infinity, 1 for negative infinity; biased exponent = all 1 bits; and the fraction = all 0 bits;

- NaN (Not-A-Number)
Sign bit = 0 or 1; biased exponent = all 1 bits; and the fraction is anything but all 0 bits. 

### 回到最初的问题
Java 当中浮点数也是[采用 IEEE 754 来存储的](https://docs.oracle.com/javase/tutorial/java/nutsandbolts/datatypes.html)，根据最初的问题，我们先把 1.0f， 0.9f， 0.8f 三个数字进行转换，得到：
- 1.0f -> 3f800000 -> 0 | 0111 1111 | 0000 0000 0000 0000 0000 000
- 0.9f -> 3f666666 -> 0 | 0111 1110 | 1100 1100 1100 1100 1100 110
- 0.8f -> 3f4ccccd -> 0 | 0111 1110 | 1001 1001 1001 1001 1001 101

按规则反向转换为十进制数后可以得到：
- 1.0f -> 3f800000 -> 1.0
- 0.9f -> 3f666666 -> 0.89999997615814208984375
- 0.8f -> 3f4ccccd -> 0.800000011920928955078125
- 1.0f - 0.9f = 0.10000002384185791015625
- 0.9f - 0.8f = 0.099999964237213134765625

原问题的结果：
``` shell
a=1.0
b=0.9
c=0.8
a-b=0.100000024
b-c=0.099999964
```

根据上述计算结果，与最初问题中显示的结果相比，有两个问题：

1. 0.8f 与 0.9f，按道理是有误差的，但是打印出来之后却是准确的 0.8 和 0.9
2. 1.0f - 0.9f 和 0.9f - 0.8f 的打印结果有偏差

显然 Java 帮我们做了舍入，逻辑就在 `Float.toString()`中
#### Float.toString()
``` java
// Float.java
public static String toString(float f) {
    return FloatingDecimal.toJavaFormatString(f);
}

// FloatingDecimal.java
public static String toJavaFormatString(float f) {
    return getBinaryToASCIIConverter(f).toJavaFormatString();
}

// FloatingDecimal.java
static BinaryToASCIIConverter getBinaryToASCIIConverter(double d, boolean isCompatibleFormat) {
    ......
    
    BinaryToASCIIBuffer buf = getBinaryToASCIIBuffer();
    buf.setSign(isNegative);
    // call the routine that actually does all the hard work.
    buf.dtoa(binExp, fractBits, nSignificantBits, isCompatibleFormat);
    
    ......
}

// BinaryToASCIIBuffer inner class in FloatingDecimal.java
private void dtoa( int binExp, long fractBits, int nSignificantBits, boolean isCompatibleFormat) {
    ......
    
    while( ! low && ! high ){
        q = b / s;
        b = 10 * ( b % s );
        m *= 10;
        assert q < 10 : q; // excessively large digit
        if ( m > 0L ){
            low  = (b <  m );
            high = (b+m > tens );
        } else {
            // hack -- m might overflow!
            // in this case, it is certainly > b,
            // which won't
            // and b+m > tens, too, since that has overflowed
            // either!
            low = true;
            high = true;
        }
        digits[ndigit++] = (char)('0' + q);
    }
    
    ......
}
```

在 dtoa() 方法中，对小数位进行还原时，做了较为复杂的舍入操作(我还没搞懂怎么做的..)。
将原问题稍作修改后：
``` java
@Test
public void floatCalculationTest() {
	System.out.println("a=" + new BigDecimal(1.0f));
	System.out.println("b=" + new BigDecimal(0.9f));
	System.out.println("c=" + new BigDecimal(0.8f));
	System.out.println("a-b=" + new BigDecimal(1.0f - 0.9f));
	System.out.println("b-c=" + new BigDecimal(0.9f - 0.8f));
}
```
得到：
``` shell
a=1
b=0.89999997615814208984375
c=0.800000011920928955078125
a-b=0.10000002384185791015625
b-c=0.099999964237213134765625
```

可见结果终于与我们计算的结果一致了。

其实正因为 Java 的舍入操作，确保了大部分不复杂的浮点数运算（需要是 double 而不是 float）可以拿到期望的结果（实测在 Java 中大多 double 运算在小数点后八位以内的值并没有出错），然而也正是 Java 贴心的处理，可能会让我们误以为 Java 的浮点数计算是绝对准确的，然后就可能会毫无心理准备的踩坑。

### 其他关于 IEEE 754
1. 除了规范化，还有非规范化的概念
	非规范化即 exponent = all 0 bits，但 fraction 不全为零的情况(exponent 和 fraction 全为零代表 0)。
	规定非规范化的数，其科学计数法表示的首位为 0，即 0.xxxx * 2^n。可见非规范化可以表示比规范化更小的数。
	因此，浮点数可以表示的范围：
	
	| 精度   | 规范化                         | 非规范化                        |
	| ------ | ------------------------------ | ------------------------------- |
	| 单精度 | ±2^-126 ~ (2 - 2^-23) * 2^127  | ±2^-149 ~ (1 - 2^-23) * 2^-126  |
	| 双精度 | ±2^-1022 ~(2 - 2^-52) * 2^1023 | ±2^-1074 ~ (1- 2^-52) * 2^-1022 |
	
2. 计算机中，浮点数在运算时，必须先将小数点调整至相同位置，之后再进行计算，故

    - 非规范化虽然可以表示比规范化更小的数，但由于计算时需要重新调整为规范化，因此会被忽略。
    - 乘法和除法运算对空间要求更高，故 CPU 的浮点运算单元通常会提供多个寄存器来暂存中间数据。

### 总结
根据上文的讨论，我们能知道为什么在 Java 中对浮点数的计算会出现意外的结果，也了解到由于计算机的限制，我们在计算浮点数时无法避免出现这种情况。
那么，在使用的过程中，有什么折中的办法呢？
1. 在没有必要使用浮点数的场景使用整数；
2. 在精度要求不高的场景对结果做长度限制并四舍五入；
3. 使用 Double 可以让精度更高，结果更接近真实值，但空间占用也更大；
4. 计算准确值的方法通常是将数字转换为字符串再模拟十进制进行运算，Java 中的解决方案是 BigDecimal，但是运算速度会大幅降低。

### 参考
1. [IEEE 754](https://en.wikipedia.org/wiki/IEEE_754)
2. [IEEE 754 Format](http://www.oxfordmathcenter.com/drupal7/node/43)
3. [IEEE754表示浮点数](https://www.jianshu.com/p/e5d72d764f2f)
4. [Full Precision Calculator](https://www.mathsisfun.com/calculator-precision.html)
