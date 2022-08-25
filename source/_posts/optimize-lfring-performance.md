---
title: 引入泛型改善 lfring 性能的探索
date: 2022-08-01 23:31:32
tags: 
- lock free ring buffer
- go generics
- performance optimization
categories:
- Go
---

{% asset_img ring.png 300 %}

本文介绍了近期对先前的库 [go-lock-free-ring-buffer](https://github.com/LENSHOOD/go-lock-free-ring-buffer) （简称 lfring）改造泛型而产生的性能优化。
介绍该 lfring 的文章可见[这里](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)。

<!-- more -->

### 1. 引入泛型

随着 go1.18 的发布，等待了 10 年的泛型终于发布了。想想去年写的 [lfring](https://github.com/LENSHOOD/go-lock-free-ring-buffer) 的库，正是因为 go 没有泛型的支持，对数据的存取全部都用了 `interface{}`，导致整个流程中反复的进行类型转换，用户也需要大量类型推断。

虽然在实际当中，可以采用单态化的方式简化编程，改善性能，但作为一个通用的无锁队列，用`interface{}`来传递参数实属没有办法的办法。

泛型发布之后，各种评论褒贬不一，但是从 lfring 的角度看，我觉得引入泛型一定是利大于弊的。那么唯一的问题就是，引入了泛型之后，会不会真的如一些文章的测试结果所展示的，会产生性能劣化。lfring 本身是想通过 lock-free 的方式来改善性能从而在某些特定场景下替代 `channel` 的，假如引入了泛型导致性能出现劣化，就可能需要多多斟酌了。

#### 1.1 性能测试

性能测试沿用了[之前文章](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)里提到的测试办法：

- 数据源：构造 64 个元素的 int 类型数组，`Offer` 时按顺序从中取出
- 对照组：将 `channel` 按 lfring 的 api 进行封装，作为对比

- 测试指标：
  - 不同的 goroutine 分别进行`Offer` 和 `Poll` 的动作，由于生产消费之间的速度差异，以及不同 goroutine 之间竞争的原因，`Poll` 可能会拿不到值（返回 `success = false`，代表这一次取操作失败，需要重试），需要重试
  - 设置一个 counter，每一次成功的 `Poll` 都能让 counter+1，counter+1 代表了某个数成功的完成了一次`Offer` 到  `Poll` 的 handover
  - 显然，在相同的时间限制内，handover 次数越多，代表性能越高（考虑到一定的随机性，实际测试中采取的是多次采样，去除 3σ 区间外的离群值后取均值的办法）
- 测试类型：通过三个不同的维度变量来测试不同场景下的性能表现
  - 环型队列中的元素容量，从 2 到 1024 变化
  - 生产者、消费者的比例，从 11:1 到 1:11 变化（测试机器一共 12 个逻辑核）
  - 总线程数量，从 2 到 48 变化（采用设置 `GOMAXPROCS` 的办法控制底层线程数，P 和 goroutine 仍旧一一对应。让线程竞争更多的表现在OS 调度器上而不是 go 调度器上）

此外，为了防止 go 版本更新而产生的其他优化对数据造成影响，使用 `interface{}` 和泛型的两套代码都在 `go version go1.19 darwin/amd64` 环境下测试。

平台是2019 款 MBP，2.6 GHz 6-Core Intel Core i7 CPU，超线程后共 12 个逻辑核。



#### 1.2 测试结果

注：下图中的 `NodeBased` 和 `Hybrid` 是 lfring 库中队列的两种不同实现，我在[先前的博客](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)中对其进行了介绍，同时在文中也分析了 `NodeBased`的性能比 `Hybrid` 好的原因 。

1. 环形队列元素容量变化
  {% iframe with-capacity.html 100% 500 %}

  从图中可见，不论是哪一种类型，其容量在 2 和 4 时都没有太大的变化，而后才产生了差异，可以暂且假设在容量为 2 和 4 时，因为其他未明因素限制了发挥。

  除去容量 2 和 4 后，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 5% 到最多13.8%，平均提升约 8.4%，而对`Hybrid`类型的 lfring，泛型比 `interface{}` 性能提升并不明显，从最低 -4% 到最高 7.5%，平均提升约 3.1%。

  但令人吃惊的是，随着容量的不断上升，作为对照组的 `Channel` 的性能近乎与线性增长，并且在 64 之后超出了无锁队列的实现，这一现象将在后文中进行分析。单看 `Channel` 本身，在引入泛型后的性能提升巨大，从最小的 10.7% 一直到 77.1%，从趋势上看甚至如果容量更大时性能提升会更多，平均提升约为 50.4%。 
  
2. 生产者与消费者比例变化
  {% iframe with-producer.html 100% 500 %}

  在容量固定为 32 时，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 5.2% 到最多 27.2%，平均提升约 17.4%，整体上看在生产者比例较低时，性能提升更明显；而对`Hybrid`类型的 lfring，泛型比 `interface{}` 性能提升并不明显，从最低 -0.1% 到最高 15.3%，平均提升约 5%；对照组 `Channel` ，对比情况也类似，在生产者比例较低时，性能提升巨大，而在消费者比例较低时区别不大，从最低 -9% 到最高 134%，平均提升约 54.1%

3. 总线程数变化
  {% iframe with-thread.html 100% 500 %}

  在容量固定为 32 时，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 3% 到最多 35.9%，平均提升约 16.9%；而对`Hybrid`类型的 lfring，除了 2 线程（一个生产者，一个消费者）性能提升较为明显，其他几乎一致，可认为没有明显改善；对照组 `Channel` ，在 2 线程时性能有所降低，但其余情况下性能都有大幅提升，从最低 -12.2% 到最高 109%，平均提升约 46%



### 2. 关于泛型的讨论

#### 2.1 泛型概览

Golang 在 2009 年公开发布的时候，就没有包含泛型特性，从 2009 年开始，一直到 2020 年之间，不断地有人提出泛型的 feature request，有关泛型的各种讨论也层出不穷（参考有关泛型的 [issue](https://github.com/golang/go/issues?q=label%3Agenerics) ）。

为什么 Golang 一直没有引入泛型呢？其作者之一 rsc 在 2009 年的讨论 [The Generic Dilemma](https://research.swtch.com/generic) 提到了一个”泛型困境“：

对于泛型，市面上可见的三种处理方式：

1. 不处理（C ），不会对语言添加任何复杂度，但会让编码更慢（slows programers）
2. 在编译期进行单态化或宏扩展（C++），生成大量代码。不产生运行时开销，但由于单态化的代码会导致 icache miss 增多从而影响性能。这种方式会让编译更慢（slows compilation）
3. 隐式装箱/拆箱（Java），可以共用一套泛型函数，实例对象在运行时做强制转换。虽然共享代码可以提高 icache 效率，但运行时的类型转换、运行时通过查表进行函数调用都会限制优化、降低运行速度，因此这种方式会让执行更慢（slows execution）

rsc 的文章并没有归纳所有的泛型实践，比如[C# 的实现](https://docs.microsoft.com/en-us/previous-versions/ms379564(v=vs.80)?redirectedfrom=MSDN#generics-implementation)：对于值类型，采用类似 C++ 的方式用实际类型替换，而对于引用类型，则直接改写为 `Object`，当前 go1.18 的泛型实现，和 C# 有点类似，不过，C# 中泛型实例化的操作都是在运行时的，所以相对来说也可以归在 ”slow execution“。此外，在一篇专门记录 go 泛型讨论的文章 [Summary of Go Generics Discussions](https://docs.google.com/document/d/1vrAy9gMpMoS3uaVphB32uVXX4pi-HnNjkMEgyAHX4N4/edit#heading=h.5nkda67to6u0) 里面总结了十几种实现泛型的方法。总之，各种实现之间也大都是在上面的三个”slow“之间抉择。

总之，go 团队一方面认为直接使用 `interface{}` 等等折中的办法能解决大多数问题，因此引入泛型并不着急，另一方面也认为还没有找到能让泛型实现的与 go 语言其他部分紧密配合的方法。因此泛型便一拖再拖。

在 go1.18 中，泛型的实现可以用一句话来概括：*GCShape Stenciling with Dictionaries*。

- [Stenciling](https://go.googlesource.com/proposal/+/refs/heads/master/design/generics-implementation-stenciling.md) 是一种泛型实现方式，指的就是类似 monomorphizing 的办法，对同一个泛型函数，给每个实际类型都生成一个实现，在编译之后，所有对泛型函数的调用，都会被替换为生成的实例类型函数的调用。这是类似 C++的实现。然而由于 Go 的 Type Alias 特性，相同底层类型的多个别称类型，就会生成多个实例函数。
- [Dictionaries](https://go.googlesource.com/proposal/+/refs/heads/master/design/generics-implementation-dictionaries.md) 恰好相反，在编译期，对每个泛型函数，只会生成单个汇编代码块，它将作为一个参数传入泛型函数中。dictionary 中主要包含的就是所有可能涉及到的实例类型的 `runtime._type` 引用。在泛型函数被实际调用时，如果函数逻辑涉及到对泛型参数的方法调用，就可以通过 dictionary + 偏移量来获得实例类型的 `runtime._type` ，而后再进一步获得 `itab` 信息并找到函数地址（对 interface 约束的泛型参数）。

Go 的泛型，在 Stenciling 和 Dictionaries 之间，做了一个折中，既不是给所有泛型类型都生成自己独一无二的实现，又不是一种泛型类型全都用一套实现，而是按照 GCShape 的维度来划分：同一个 GCShape 下共享一套实现，不同的 GCShape，生成各自的实现。

所以，GCShape 是什么？

在 [go1.18 的泛型实现文档](https://github.com/golang/proposal/blob/master/design/generics-implementation-dictionaries-go1.18.md)中用一句话对其进行了概况：*Two concrete types are in the same gcshape grouping if and only if they have the same underlying type or they are both pointer types*。即当且仅当两个具体类型，其底层类型完全一致，或是这两个类型都是指针类型时，它们的 GCShape 才是一致的。



#### 2.2 不同的 GCShape

为了更具象化的了解 go 泛型的底层实现，我们通过几个例子来尝试泛型，并看一看 go 的编译器是怎么处理它们的：

```go
package main

import "fmt"

type Number interface {
	~int64 | ~float32 | ~uint8
}

type u8 uint8

func numeric[T Number](left, right T) T {
	if left > right {
		return (left - right) * 8
	}
	return (right - left) * 4
}

type Animal interface {
	say() string
}
func doSay[T Animal](e T) string { return e.say() }

type Cat struct { name string }
func (c *Cat) say() string { return c.name }

type Dog struct { name string }
func (c *Dog) say() string { return c.name }

func putToArr[T Dog | Cat](e T) []T {
	arr := make([]T, 1)
	arr[0] = e
	return arr
}

func main() {
	n1 := numeric(int64(17), int64(35))
	n2 := numeric(float32(155.5), float32(233.1))
	n3 := numeric(u8(3), u8(5))
	fmt.Printf("n1: %v, n2, %v, n3 %v\n", n1, n2, n3)

	cat := Cat{"cat"}
	dog := Dog{"dog"}
	s1 := cat.say()
	s2 := dog.say()
	s3 := Animal(&cat).say()
	fmt.Printf("s1: %v, s2, %v, s3: %v\n", s1, s2, s3)

	s4 := putToArr(dog)
	s5 := putToArr(cat)
	s6 := doSay(&dog)
	fmt.Printf("s4: %v, s5: %v, s6: %v\n", s4, s5, s6)

	var animal = Animal(&cat)
	s7 := doSay(animal)
	fmt.Printf("s7: %v\n", s7)
}

```

上述程序大致对泛型进行了如下的两种测试：

- 定义泛型函数 `numeric()`，可接受的底层类型为三种数值类型；在`main()` 测试中，分别测试了`int64`、`float32`、以及重命名的`uint8` 
- 定义具有`say()` 方法的 interface `Animal`，并定义两个底层类型都是 string 的 struct `Cat` 和 `Dog`，之后分别定义了两个泛型函数，一个接受限制为 `Animal` 的泛型类型，另一个接受限制为 `Cat` 或 `Dog` 的泛型类型

接下来我们分别看一看上述两种测试，其实际的编译结果是怎么样的。为了理解上更清晰，我们还会对比编译结果在进行编译优化前/后的不同状态。

##### 数值类型泛型

对于泛型函数 `func numeric[T Number](left, right T) T`，我们得到其未经优化的汇编代码：

```assembly
main.numeric[go.shape.int64_0] STEXT dupok nosplit size=107 args=0x18 locals=0x10 funcid=0x0 align=0x0
	0x0000 00000 (main.go:11)	TEXT	main.numeric[go.shape.int64_0](SB), DUPOK|NOSPLIT|ABIInternal, $16-24
	...
	0x000e 00014 (main.go:11)	MOVQ	AX, main..dict+24(SP)
	0x0013 00019 (main.go:11)	MOVQ	BX, main.left+32(SP)
	0x0018 00024 (main.go:11)	MOVQ	CX, main.right+40(SP)
	0x001d 00029 (main.go:11)	MOVQ	$0, main.~r0(SP)
	0x0025 00037 (main.go:12)	MOVQ	main.left+32(SP), CX
	0x002a 00042 (main.go:12)	CMPQ	main.right+40(SP), CX ## if left > right
	0x002f 00047 (main.go:12)	JLT	51
	0x0031 00049 (main.go:12)	JMP	79
	0x0033 00051 (main.go:13)	MOVQ	main.left+32(SP), AX
	0x0038 00056 (main.go:13)	SUBQ	main.right+40(SP), AX ## AX = left - right
	0x003d 00061 (main.go:13)	SHLQ	$3, AX                ## AX << 3 as AX = AX*8
	0x0041 00065 (main.go:13)	MOVQ	AX, main.~r0(SP)
	0x0045 00069 (main.go:13)	MOVQ	8(SP), BP
	0x004a 00074 (main.go:13)	ADDQ	$16, SP
	0x004e 00078 (main.go:13)	RET
	0x004f 00079 (main.go:15)	MOVQ	main.right+40(SP), AX
	0x0054 00084 (main.go:15)	SUBQ	main.left+32(SP), AX  ## AX = right - left
	0x0059 00089 (main.go:15)	SHLQ	$2, AX                ## AX << 2 as AX = AX*4
	0x005d 00093 (main.go:15)	MOVQ	AX, main.~r0(SP)
	0x0061 00097 (main.go:15)	MOVQ	8(SP), BP
	0x0066 00102 (main.go:15)	ADDQ	$16, SP
	0x006a 00106 (main.go:15)	RET
	...
main.numeric[go.shape.float32_0] STEXT dupok nosplit size=139 args=0x10 locals=0x10 funcid=0x0 align=0x0
	...
main.numeric[go.shape.uint8_0] STEXT dupok nosplit size=103 args=0x10 locals=0x10 funcid=0x0 align=0x0
	...
```

我们最先发现的就是，函数真的生成了三个版本！它们的名字分别是 `main.numeric[go.shape.int64_0]`，`main.numeric[go.shape.float32_0]`， `main.numeric[go.shape.uint8_0]`，正好对应了泛型约束中的 `~int64`、`~float32`、`~uint8`，其中我们定义的别名 `u8` 没有出现输出代码中，而 `u8` 被替换为了其底层类型 `uint8`。

由于三个不同 GCShape 所生成的代码，几乎一致（除了`float32` 中用 X0 寄存器替换 BX，用 MOVSS 指令替换 MOVQ 这类整数和浮点数的指令差异），因此只保留了`int64` 的代码。 三种数值类型的 GCShape，生成了三个函数，代表了 Stenciling。

在程序开头我们发现，除了按照 calling convention 将传入参数 `left` 和 `right` 从栈上拷贝到 BX 和 CX 寄存器，还拷贝了一个名为 `main..dict` 的入参，放入 AX（并且按顺序来看，该参数还是第一个入参）。这实际上代表了 Dictionaries：对于相同的 GCShape，可能映射到不同的具体类型，那么当调用方法时，到底调用哪一个呢？这就需要在传入的这个 `main..dict` 结构中进行二次检索才能找得到。不过这里的例子中并没有用到 `main..dict` 的场景。

再来看一看开启了编译优化的结果：

```assembly
main.numeric[go.shape.int64_0] STEXT dupok nosplit size=27 args=0x18 locals=0x0 funcid=0x0 align=0x0
	0x0000 00000 (main.go:11)	TEXT	main.numeric[go.shape.int64_0](SB), DUPOK|NOSPLIT|ABIInternal, $0-24
	...
	0x0000 00000 (main.go:12)	CMPQ	CX, BX  ## register based argument
	0x0003 00003 (main.go:12)	JGE	16
	0x0005 00005 (main.go:13)	SUBQ	CX, BX
	0x0008 00008 (main.go:13)	SHLQ	$3, BX
	0x000c 00012 (main.go:13)	MOVQ	BX, AX
	0x000f 00015 (main.go:13)	RET
	0x0010 00016 (main.go:15)	SUBQ	BX, CX
	0x0013 00019 (main.go:15)	SHLQ	$2, CX
	0x0017 00023 (main.go:15)	MOVQ	CX, AX
	0x001a 00026 (main.go:15)	RET
```

显而易见，通过栈传参没了，直接替换为了寄存器传参（[proposal](https://go.googlesource.com/proposal/+/refs/changes/78/248178/1/design/40724-register-calling.md)），另外 `mian..dict` 既然没用，也不必出现了。

另外，非优化版本中的各种来回在寄存器和栈直接传值，也全部优化成了直接在寄存器中运算。

##### Struct 类型泛型 - 1：putToArray

先看看未经优化的 `putToArray()` 的汇编结果：

```assembly
main.putToArr[go.shape.struct { <unlinkable>.name string }_0] STEXT dupok size=229 args=0x18 locals=0x50 funcid=0x0 align=0x0
	0x0000 00000 (main.go:32)	TEXT	main.putToArr[go.shape.struct { <unlinkable>.name string }_0](SB), DUPOK|ABIInternal, $80-24
	...
	0x0018 00024 (main.go:32)	MOVQ	AX, main..dict+88(SP)
	0x001d 00029 (main.go:32)	MOVQ	BX, main.e+96(SP)
	0x0022 00034 (main.go:32)	MOVQ	CX, main.e+104(SP)
	0x0027 00039 (main.go:32)	MOVQ	$0, main.~r0+24(SP)
	0x0030 00048 (main.go:32)	MOVUPS	X15, main.~r0+32(SP)
	## AX，BX，CX 存放 makeslice 需要的三个入参，type，len，cap
	0x0036 00054 (main.go:33)	LEAQ	type.go.shape.struct { <unlinkable>.name string }_0(SB), AX
	0x003d 00061 (main.go:33)	MOVL	$1, BX
	0x0042 00066 (main.go:33)	MOVQ	BX, CX
	...
	0x0045 00069 (main.go:33)	CALL	runtime.makeslice(SB) ## 构造并初始化容量为 1 的 slice
	0x004a 00074 (main.go:33)	MOVQ	AX, main.arr+48(SP)   ## AX 保存了返回的 slice 地址
	0x004f 00079 (main.go:33)	MOVQ	$1, main.arr+56(SP)   ## slice len
	0x0058 00088 (main.go:33)	MOVQ	$1, main.arr+64(SP)   ## slice cap
	## 入参的底层类型是 string，包含一个 8 字节的 "str unsafe.Pointer" 和 8 字节的 "len int"
	0x0061 00097 (main.go:34)	MOVQ	main.e+96(SP), DX     ## str
	0x0066 00102 (main.go:34)	MOVQ	main.e+104(SP), SI    ## len
	0x006b 00107 (main.go:34)	JMP	109
	0x006d 00109 (main.go:34)	MOVQ	SI, 8(AX)             ## string.len 存入 arr[0]+8
	...
	0x007c 00124 (main.go:34)	MOVQ	DX, (AX)              ## string.str 存入 arr[0]
	...
	0x008c 00140 (main.go:35)	MOVQ	main.arr+56(SP), BX
	0x0091 00145 (main.go:35)	MOVQ	main.arr+64(SP), CX
	0x0096 00150 (main.go:35)	MOVQ	main.arr+48(SP), DX
	0x009b 00155 (main.go:35)	MOVQ	DX, main.~r0+24(SP)
	0x00a0 00160 (main.go:35)	MOVQ	BX, main.~r0+32(SP)
	0x00a5 00165 (main.go:35)	MOVQ	CX, main.~r0+40(SP)
	0x00aa 00170 (main.go:35)	MOVQ	main.~r0+24(SP), AX
	0x00af 00175 (main.go:35)	MOVQ	72(SP), BP
	0x00b4 00180 (main.go:35)	ADDQ	$80, SP
	0x00b8 00184 (main.go:35)	RET
	...
```

生成的函数  `main.putToArr[go.shape.struct { <unlinkable>.name string }_0]` 再次证明了，相同底层类型的 struct，共享同一个GCShape。

优化后的代码：

```assembly
main.putToArr[go.shape.struct { <unlinkable>.name string }_0] STEXT dupok size=159 args=0x18 locals=0x30 funcid=0x0 align=0x0
	0x0000 00000 (main.go:32)	TEXT	main.putToArr[go.shape.struct { <unlinkable>.name string }_0](SB), DUPOK|ABIInternal, $48-24
	...
	0x0019 00025 (main.go:34)	MOVQ	CX, main..autotmp_7+24(SP)
	0x001e 00030 (main.go:34)	MOVQ	BX, main..autotmp_8+32(SP)
	0x0023 00035 (main.go:33)	LEAQ	type.go.shape.struct { <unlinkable>.name string }_0(SB), AX
	0x002a 00042 (main.go:33)	MOVL	$1, BX
	0x002f 00047 (main.go:33)	MOVQ	BX, CX
	...
	0x0032 00050 (main.go:33)	CALL	runtime.makeslice(SB)
	0x0037 00055 (main.go:34)	MOVQ	main..autotmp_7+24(SP), DX  ## string.len
	0x003c 00060 (main.go:34)	MOVQ	DX, 8(AX)
	...
	0x0049 00073 (main.go:34)	MOVQ	main..autotmp_8+32(SP), DX  ## string.str
	0x004e 00078 (main.go:34)	MOVQ	DX, (AX)
	0x0051 00081 (main.go:34)	JMP	101
	0x0053 00083 (main.go:34)	MOVQ	AX, DI
	...
	0x0065 00101 (main.go:35)	MOVL	$1, BX
	0x006a 00106 (main.go:35)	MOVQ	BX, CX
	0x006d 00109 (main.go:35)	MOVQ	40(SP), BP
	0x0072 00114 (main.go:35)	ADDQ	$48, SP
	0x0076 00118 (main.go:35)	RET
	...
```

同样也是优化掉了许多栈复制动作和 `main..dict` 相关的逻辑。

##### Struct 类型泛型 - 1：putToArray





### 3. 回到 lfring

#### 3.1 简单看看汇编

#### 3.2 改变数据类型



### 4. Channel 迷思

