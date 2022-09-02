---
title: 探索引入泛型对 lfring 产生的性能影响
date: 2022-08-01 23:31:32
tags: 
- lock free ring buffer
- go generics
- performance optimization
categories:
- Go
---

{% asset_img ring.png 300 %}

本文是对我自己的一个库 [go-lock-free-ring-buffer](https://github.com/LENSHOOD/go-lock-free-ring-buffer) （简称 lfring）改造泛型而产生性能影响的讨论。

介绍该 lfring 的文章可见[这里](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)。

<!-- more -->

### 1. 引入泛型

随着 go1.18 的发布，等待了 10 年的泛型终于发布了。想想去年写的 [lfring](https://github.com/LENSHOOD/go-lock-free-ring-buffer) 的库，正是因为 go 没有泛型的支持，对数据的存取全部都用了 `interface{}`，导致整个流程中反复的进行类型转换，用户在使用的时候也需要大量类型推断。

虽然在实际当中，可以采用单态化的方式简化编程，改善性能，但作为一个通用的无锁队列，用`interface{}`来传递参数实属没有办法的办法。

泛型发布之后，各种评论褒贬不一，但是从 lfring 的角度看，我觉得引入泛型一定是利大于弊的。那么唯一的问题就是，引入了泛型之后，会不会真的如一些文章的测试结果所展示的，会产生性能劣化。lfring 本身是想通过 lock-free 的方式来改善性能从而在某些特定场景下替代 `channel` 的，假如引入泛型会导致性能出现劣化，那就需要多多斟酌了。

#### 1.1 性能测试

性能测试沿用了[之前文章](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)里提到的测试办法：

代码片段：

```go
...
b.RunParallel(func(pb *testing.PB) {
  ...
  for i := 1; pb.Next(); i++ {
    // producer 分支
    if producer {
      buffer.Offer(ints[(i & (len(ints) - 1))])
    } else {
      // consumer 分支
      if _, success := buffer.Poll(); success {
        // handover counter 递增
        atomic.AddInt32(&counter, 1)
      }
    }
  }
})
...
// 取 counter 作为性能指标
b.ReportMetric(float64(counter), "handovers")
```

- 数据源：构造 64 个元素的 int 类型数组，`Offer` 时按顺序从中取出
- 对照组：直接与原先 `interface{}` 方案的代码作为对照

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

  除去容量 2 和 4 后，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 5% 到最多13.8%，平均提升约 8.4%，而对`Hybrid`类型的 lfring，泛型比 `interface{}` 性能提升度不大，从最低 -4% 到最高 7.5%，平均提升约 3.1%。

2. 生产者与消费者比例变化
  {% iframe with-producer.html 100% 500 %}

  在容量固定为 32 时，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 5.2% 到最多 27.2%，平均提升约 17.4%，整体上看在生产者比例较低时，性能提升更明显；而对`Hybrid`类型的 lfring，泛型比 `interface{}` 性能提升度同样稍逊，从最低 -0.1% 到最高 15.3%，平均提升约 5%。

3. 总线程数变化
  {% iframe with-thread.html 100% 500 %}

  在容量固定为 32 时，对于 `NodeBased` 类型的 lfring，泛型比 `interface{}` 性能提升从最少 3% 到最多 35.9%，平均提升约 16.9%；而对`Hybrid`类型的 lfring，除了 2 线程（一个生产者，一个消费者）性能提升较为明显，其他提升度相对较小，改善不明显。

从测试结果上看，改造了泛型以后，性能非但没有劣化，反倒有所提升，而`NodeBased` 类型的 ring 要比  `Hybrid` 类型的性能提升更明显。后文将具体分析其原因。



### 2. 关于泛型的讨论

#### 2.1 泛型概览

Golang 在 2009 年公开发布的时候，就没有包含泛型特性，从 2009 年开始，一直到 2020 年之间，不断地有人提出泛型的 feature request，有关泛型的各种讨论也层出不穷（参考有关泛型的 [issue](https://github.com/golang/go/issues?q=label%3Agenerics) ）。

为什么 Golang 一直没有引入泛型呢？其作者之一 rsc 在 2009 年的讨论 [The Generic Dilemma](https://research.swtch.com/generic) 提到了一个”泛型困境“：

对于泛型，市面上可见的三种处理方式：

1. 不处理（C ），不会对语言添加任何复杂度，但会让编程的过程更慢（slows programers）
2. 在编译期进行单态化或宏扩展（C++），生成大量代码。不产生运行时开销，但由于单态化的代码可能导致 icache miss 增多从而影响部分性能。由于会在编译器生成代码，这种方式会让编译更慢（slows compilation）
3. 隐式装箱/拆箱（Java），可以共用一套泛型函数，实例对象在运行时做强制转换。虽然共享代码可以提高 icache 效率，但运行时的类型转换、运行时通过查表进行函数调用都会限制编译优化、降低运行速度，因此这种方式会让执行更慢（slows execution）

rsc 的文章并没有归纳所有的泛型实践，比如[C# 的实现](https://docs.microsoft.com/en-us/previous-versions/ms379564(v=vs.80)?redirectedfrom=MSDN#generics-implementation)：对于值类型，采用类似 C++ 的方式用实际类型替换，而对于引用类型，则直接改写为 `Object`，当前 go1.18 的泛型实现，和 C# 有点类似，不过，C# 中泛型实例化的操作都是在运行时的，所以相对来说也可以归在 ”slow execution“。此外，在一篇专门记录 go 泛型讨论的文章 [Summary of Go Generics Discussions](https://docs.google.com/document/d/1vrAy9gMpMoS3uaVphB32uVXX4pi-HnNjkMEgyAHX4N4/edit#heading=h.5nkda67to6u0) 里面总结了十几种实现泛型的方法。总之，各种实现也大都是在上面的三个”slow“之间抉择。

总之，go 团队一方面认为直接使用 `interface{}` 的折中办法能解决大多数问题，因此引入泛型并不着急，另一方面也自认为还没有找到能让泛型实现的与 go 语言其他部分紧密配合的方法。结果泛型就一拖再拖，拖了 10 年终于在 1.18 版本发布了。

在 go1.18 中，泛型的实现可以用一句话来概括：*GCShape Stenciling with Dictionaries*。

- [Stenciling](https://go.googlesource.com/proposal/+/refs/heads/master/design/generics-implementation-stenciling.md) 是一种泛型实现方式，指的就是类似 monomorphizing 的办法，对同一个泛型函数，给每个实际类型都生成一个实现，在编译之后，所有对泛型函数的调用，都会被替换为生成的实例类型函数的调用。这是类似 C++的实现。然而由于 Go 的 Type Alias 特性，相同底层类型的多个别称类型，就会生成多个实例函数，所以链接时去重是一个大工程。
- [Dictionaries](https://go.googlesource.com/proposal/+/refs/heads/master/design/generics-implementation-dictionaries.md) 恰好相反，在编译期，对每个泛型函数，只会生成单个汇编代码块，它将作为一个参数传入泛型函数中。dictionary 中主要包含的就是所有可能涉及到的实例类型的 `runtime._type` 引用。在泛型函数被实际调用时，如果函数逻辑涉及到对泛型参数的方法调用，就可以通过 dictionary + 偏移量来获得实例类型的 `runtime._type` ，而后再进一步获得 `itab` 信息并找到函数地址。

Go 的泛型，在 Stenciling 和 Dictionaries 之间，做了一个折中，既不是给所有泛型类型都生成自己独一无二的实现，又不是一种泛型类型全都用一套实现，而是按照 GCShape 的维度来划分：同一个 GCShape 下共享一套实现，不同的 GCShape，生成各自的实现。

所以，GCShape 是什么？

在 [go1.18 的泛型实现文档](https://github.com/golang/proposal/blob/master/design/generics-implementation-dictionaries-go1.18.md)中用一句话对其进行了概括：*Two concrete types are in the same gcshape grouping if and only if they have the same underlying type or they are both pointer types*。即当且仅当两个具体类型，其底层类型完全一致，或是这两个类型都是指针类型时，它们的 GCShape 才是一致的。



#### 2.2 不同的 GCShape

为了更具象化的了解 go 泛型的底层实现，我们通过几个例子来尝试泛型，并看一看 go 的编译器是怎么处理它们的：

```go
type Number interface { ~int64 | ~float32 | ~uint8 }

type u8 uint8

// 数值类型约束的泛型
func numeric[T Number](left, right T) T {
	if left > right {
		return (left - right) * 8
	}
	return (right - left) * 4
}

type Animal interface { say() string }

// 接口类型约束的泛型
func doSay[T Animal](e T) string { return e.say() }

type Cat struct { name string }
func (c *Cat) say() string { return c.name }

type Dog struct { name string }
func (c *Dog) say() string { return c.name }

// 实例类型约束的泛型
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
	s6 := doSay(&cat)
	s7 := doSay(&dog)
	fmt.Printf("s4: %v, s5: %v, s6: %v, s7: %v\n", s4, s5, s6, s7)

	var animal = Animal(&cat)
	s8 := doSay(animal)
	fmt.Printf("s8: %v\n", s8)
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
	TEXT	main.numeric[go.shape.int64_0](SB), DUPOK|NOSPLIT|ABIInternal, $16-24
	...
	MOVQ	AX, main..dict+24(SP)
	MOVQ	BX, main.left+32(SP)
	MOVQ	CX, main.right+40(SP)
	MOVQ	$0, main.~r0(SP)
	MOVQ	main.left+32(SP), CX
	CMPQ	main.right+40(SP), CX ## if left > right
	JLT	51
	JMP	79
	MOVQ	main.left+32(SP), AX
	SUBQ	main.right+40(SP), AX ## AX = left - right
	SHLQ	$3, AX                ## AX << 3 as AX = AX*8
	MOVQ	AX, main.~r0(SP)
	MOVQ	8(SP), BP
	ADDQ	$16, SP
	RET
	MOVQ	main.right+40(SP), AX
	SUBQ	main.left+32(SP), AX  ## AX = right - left
	SHLQ	$2, AX                ## AX << 2 as AX = AX*4
	MOVQ	AX, main.~r0(SP)
	MOVQ	8(SP), BP
	ADDQ	$16, SP
	RET
	...
main.numeric[go.shape.float32_0] STEXT dupok nosplit size=139 args=0x10 locals=0x10 funcid=0x0 align=0x0
	...
main.numeric[go.shape.uint8_0] STEXT dupok nosplit size=103 args=0x10 locals=0x10 funcid=0x0 align=0x0
	...
```

我们最先发现的就是，函数真的生成了三个版本！它们的名字分别是 `main.numeric[go.shape.int64_0]`，`main.numeric[go.shape.float32_0]`， `main.numeric[go.shape.uint8_0]`，正好对应了泛型约束中的 `~int64`、`~float32`、`~uint8`，而我们定义的别名 `u8` 没有出现输出代码中，而 `u8` 被替换为了其底层类型 `uint8`。

由于三个不同 GCShape 所生成的代码，几乎一致（除了`float32` 中用 X0 寄存器替换 BX，用 MOVSS 指令替换 MOVQ 这类整数和浮点数的指令差异），因此只保留了`int64` 的代码。 三种数值类型的 GCShape，生成了三个函数，代表了 Stenciling。

在程序开头我们发现，除了按照 calling convention 将传入参数 `left` 和 `right` 从栈上拷贝到 BX 和 CX 寄存器，还拷贝了一个名为 `main..dict` 的入参，放入 AX（并且按顺序来看，该参数还是第一个入参）。这其实就是 Dictionaries：对于相同的 GCShape，可能映射到不同的具体类型，那么当调用方法时，到底调用哪一个呢？这就需要在传入的 `main..dict` 结构中进行二次检索才能找到。不过这里的例子中并没有用到 `main..dict` 的场景。

再来看一看开启了编译优化的结果：

```assembly
main.numeric[go.shape.int64_0] STEXT dupok nosplit size=27 args=0x18 locals=0x0 funcid=0x0 align=0x0
	TEXT	main.numeric[go.shape.int64_0](SB), DUPOK|NOSPLIT|ABIInternal, $0-24
	...
	CMPQ	CX, BX  ## register based argument
	JGE	16
	SUBQ	CX, BX
	SHLQ	$3, BX
	MOVQ	BX, AX
	RET
	SUBQ	BX, CX
	SHLQ	$2, CX
	MOVQ	CX, AX
	RET
```

显而易见，通过栈传参没了，直接替换为了寄存器传参（[proposal](https://go.googlesource.com/proposal/+/refs/changes/78/248178/1/design/40724-register-calling.md)），另外 `mian..dict` 既然没用，也不必出现了。

另外，非优化版本中的各种来回在寄存器和栈直接传值，也全部优化成了直接在寄存器中运算。

##### Struct 类型泛型 ~ 1：putToArray

由于编译优化和上文类似，因此直接展示开启编译优化后的 `putToArray()` 的汇编结果：

```assembly
main.putToArr[go.shape.struct { <unlinkable>.name string }_0] STEXT dupok size=159 args=0x18 locals=0x30 funcid=0x0 align=0x0
	TEXT	main.putToArr[go.shape.struct { <unlinkable>.name string }_0](SB), DUPOK|ABIInternal, $48-24
	...
	MOVQ	CX, main..autotmp_7+24(SP)
	MOVQ	BX, main..autotmp_8+32(SP)
	## AX，BX，CX 存放 makeslice 需要的三个入参，type，len，cap
	LEAQ	type.go.shape.struct { <unlinkable>.name string }_0(SB), AX
	MOVL	$1, BX
	MOVQ	BX, CX
	...
	CALL	runtime.makeslice(SB)       ## 堆上构造并初始化容量为 1 的 slice，首地址存入 AX
	## 入参的底层类型是 string，包含一个 8 字节的 "str unsafe.Pointer" 和 8 字节的 "len int"
	MOVQ	main..autotmp_7+24(SP), DX  ## (AX) + 8 <= string.len
	MOVQ	DX, 8(AX)
	...
	MOVQ	main..autotmp_8+32(SP), DX  ## (AX) <= string.str
	MOVQ	DX, (AX)
	JMP	101
	...
	## 函数返回的 slice，AX 已经存储了地址，再向 BX，CX 中存放 len == cap == 1
	MOVL	$1, BX
	MOVQ	BX, CX
	MOVQ	40(SP), BP
	ADDQ	$48, SP
	RET
	...
```

生成的函数  `main.putToArr[go.shape.struct { <unlinkable>.name string }_0]` 再次证明了，相同底层类型的 struct，共享同一个GCShape，因此没有出现专为`Cat` 和 `Dog` 生成的代码。

##### Struct 类型泛型 - 2：doSay

`doSay()` 在汇编结果上看，生成了两个函数：

```assembly
main.doSay[go.shape.*uint8_0] STEXT dupok size=118 args=0x10 locals=0x30 funcid=0x0 align=0x0
	TEXT	main.doSay[go.shape.*uint8_0](SB), DUPOK|ABIInternal, $48-16
	...
	MOVQ	AX, main..dict+56(SP)
	MOVQ	BX, main.e+64(SP)     ## 入参 e
	MOVUPS	X15, main.~r0+8(SP) 
	MOVQ	main..dict+56(SP), DX ## DX 存放 dict 地址
	...
	TESTB	AL, (DX)
	LEAQ	16(DX), CX            ## dict+16 存入 CX
	MOVQ	main.e+64(SP), AX     ## 方法隐式参数：receiver
	MOVQ	16(DX), BX            ## dict+16 指向的数据即 (*Dog).say 方法地址
	MOVQ	CX, DX                ## 按目前的 calling convention，DX 存放 closure ctx
	...
	CALL	BX                    ## 跳转到 (*Dog).say
	MOVQ	AX, main..autotmp_3+24(SP)
	MOVQ	BX, main..autotmp_3+32(SP)
	MOVQ	AX, main.~r0+8(SP)
	MOVQ	BX, main.~r0+16(SP)
	MOVQ	40(SP), BP
	ADDQ	$48, SP
	RET
	...
main.doSay[go.shape.interface { <unlinkable>.say() string }_0] STEXT dupok size=141 args=0x18 locals=0x38 funcid=0x0 align=0x0
	...
	LEAQ	16(DX), CX
	MOVQ	main.e+72(SP), AX
	MOVQ	main.e+80(SP), BX
	MOVQ	16(DX), SI             ## dict+16 存入 SI
	MOVQ	CX, DX
	PCDATA	$1, $1
	CALL	SI                     ## 跳转到 Animal.say
	...
<unlinkable>..dict.doSay[*<unlinkable>.Dog] SRODATA dupok size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 type.*<unlinkable>.Dog+0
	rel 0+0 t=23 type.*<unlinkable>.Dog+0
	rel 16+8 t=1 <unlinkable>.(*Dog).say+0
<unlinkable>..dict.doSay[<unlinkable>.Animal] SRODATA dupok size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 type.<unlinkable>.Animal+0
	rel 16+8 t=1 <unlinkable>.Animal.say+0
	
"".doSay[go.shape.*uint8_0] STEXT dupok size=124 args=0x10 locals=0x38 funcid=0x0 align=0x0
	...
	MOVQ	AX, ""..dict+64(SP)
	MOVQ	BX, "".e+72(SP)         ## 入参 e
	MOVUPS	X15, "".~r0+16(SP)
	MOVQ	"".e+72(SP), AX
	MOVQ	AX, ""..autotmp_3+8(SP)
	MOVQ	""..dict+64(SP), CX     ## CX 存放 dict 地址
	...
	TESTB	AL, (CX)
	MOVQ	16(CX), CX              ## dict+16 指向 Cat/Dog 的 itab
	TESTB	AL, (CX)
	MOVQ	24(CX), CX              ## itab+24 即为实际方法列表地址
	...
	CALL	CX                      ## 方法列表中只有一个方法，直接调用就是 say()
	MOVQ	AX, ""..autotmp_4+32(SP)
	MOVQ	BX, ""..autotmp_4+40(SP)
	MOVQ	AX, "".~r0+16(SP)
	MOVQ	BX, "".~r0+24(SP)
	MOVQ	48(SP), BP
	ADDQ	$56, SP
	NOP
	RET
	...
"".doSay[go.shape.interface { "".say() string }_0] STEXT dupok size=170 args=0x18 locals=0x58 funcid=0x0 align=0x0
	...
	MOVQ	"".e+112(SP), CX         ## 入参 e.data (参见 iface)
	MOVQ	CX, ""..autotmp_5+24(SP)
	MOVQ	"".e+104(SP), BX         ## 入参 e.tab (参见 iface)
	LEAQ	type."".Animal(SB), AX   ## Animal type
	PCDATA	$1, $1
	CALL	runtime.assertI2I(SB)    ## 校验传入的 e 是否是 Animal，返回值 itab 存入 AX
	MOVQ	AX, ""..autotmp_3+64(SP)
	MOVQ	""..autotmp_5+24(SP), CX
	MOVQ	CX, ""..autotmp_3+72(SP)
	TESTB	AL, (AX)
	MOVQ	24(AX), DX               ## itab+24 即为实际方法列表地址
	MOVQ	CX, AX
	...
	CALL	DX                       ## 方法列表中只有一个方法，直接调用就是 say()
	...
""..dict.doSay[*"".Cat] SRODATA dupok size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 type.*"".Cat+0
	rel 0+0 t=23 type.*"".Cat+0
	rel 0+0 t=23 type.*"".Cat+0
	rel 16+8 t=1 go.itab.*"".Cat,"".Animal+0
""..dict.doSay[*"".Dog] SRODATA dupok size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 type.*"".Dog+0
	rel 0+0 t=23 type.*"".Dog+0
	rel 0+0 t=23 type.*"".Dog+0
	rel 16+8 t=1 go.itab.*"".Dog,"".Animal+0
""..dict.doSay["".Animal] SRODATA dupok size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 type."".Animal+0
```

对于`doSay()` 分别基于在`main()` 中直接传入的 `*Cat/*Dog` 和传入被转换为 `Animal` 的 Dog，生成了两个类似的函数。`*Cat/*Dog`对应的函数其 GCShape 是 `*uint8_0`，而`Animal` 对应的 GCShape 则是 `interface { <unlinkable>.say() string }_0`，符合 GCShape 的定义（interface 视为单独的 GCShape，而指针类型全部视为`*unit8`）。

这里我们会发现，由于指针类型的 GCShape 是完全相同的，那么当传入的参数分别是 `*Cat/*Dog` 时，如何才能找到正确的 `say()` 呢？

通过汇编代码可知，泛型函数需要先到`dict` 中找到对应类型的 interface `itab`，再通过`itab` 里面的方法列表找到方法进行调用。相比非泛型下的 interface 接口调用，多了一次寻址 dict 的步骤，可以预见这种方式一定更慢。而对 `Animal` 的泛型调用更夸张，中间竟然还穿插了一次运行时的类型校验函数调用。

上述测试主要参考了 Vicent Marti 的文章 [Generics can make your Go code slower](https://planetscale.com/blog/generics-can-make-your-go-code-slower) 中设计的测试，更详细的解读参见该文。



### 3. 回到 lfring

#### 3.1 简单看看汇编

基于前一节的介绍，我们已经对 go 泛型所生成的汇编代码有了一定的认识，那么现在我们直接将两个版本（interface{} 和 generics）的 lfring 汇编结果进行比较，来直观的进行体会：

```assembly
###### interface{} 代码经过精简，省略了无关本文的部分
<unlinkable>.(*nodeBased).Offer STEXT size=293 args=0x18 locals=0x30 funcid=0x0 align=0x0
	MOVQ	AX, <unlinkable>.r+56(SP)         ## receiver
	MOVQ	BX, <unlinkable>.value+64(SP)     ## eface._type
	MOVQ	CX, <unlinkable>.value+72(SP)     ## eface.data
	LEAQ	type.interface {}(SB), AX
	CALL	runtime.newobject(SB)             ## 堆上创建 interface{} 类型对象
	MOVUPS	<unlinkable>.value+64(SP), X0   ## 将 eface 赋值到创建的对象上
	MOVUPS	X0, (AX)                        ## MOVUPS SSE 指令一次性移动 128bit
  ... 之后完全相同

###### generics 代码经过精简，省略了无关本文的部分
<unlinkable>.(*nodeBased[go.shape.int_0]).Offer STEXT dupok size=249 args=0x18 locals=0x20 funcid=0x0 align=0x0
	MOVQ	BX, <unlinkable>.r+48(SP)         ## receiver
	MOVQ	CX, <unlinkable>.value+56(SP)     ## value
	LEAQ	type.go.shape.int_0(SB), AX
	CALL	runtime.newobject(SB)             ## 堆上创建 go.shape.int_0 类型对象
	MOVQ	<unlinkable>.value+56(SP), CX
	MOVQ	CX, (AX)                          ## 为创建的对象赋值为 value
  ... 之后完全相同
```

以`Offer`为例，我们发现 `interface{}` 和 `generics` 两种区别不大，基本都是将传入的`value` 在堆上创建对象，唯一区别是创建对象的类型不同。

再看一看实际调用的地方：

```assembly
###### interface{}
	...
	MOVQ	(DX)(BX*8), AX       ## 入参放到 AX
	CALL	runtime.convT64(SB)  ## 入参是 64bit int，convT64 将其分配在堆上
	MOVQ	<unlinkable>.buffer.itab+48(SP), CX
	MOVQ	24(CX), DX           ## Buffer.iface.fun 得到接口第一个方法地址
	LEAQ	type.int(SB), BX     ## interface{} 类型的 value：eface._type
	MOVQ	AX, SI
	MOVQ	<unlinkable>.buffer.data+64(SP), AX ## receiver
	MOVQ	SI, CX               ## interface{} 类型的 value：eface.data
	CALL	DX                   ## Offer(receiver, eface._type, eface.data)
	...
	
###### generics
	...
	MOVQ	<unlinkable>.buffer.itab+40(SP), R8
	MOVQ	24(R8), CX           ## CX = Buffer.iface.fun
	MOVQ	(DI)(R10*8), BX      ## 入参放到 BX
	MOVQ	SI, AX               ## receiver
	CALL	CX                   ## Offer(receiver, int value)
	...
```

这里就能看出区别了，对于接收泛型类型参数的函数，在编译期其泛型就已经确定被替换为了 `go.shape.int_0`，该 GCShape 不包含指针，因此传参时直接传值。

但对于接收 `interface{}` 类型的参数，虽然实际上传入的参数仍旧是 int，但由于强行转换成了 `interface{}`，因此编译器在进行逃逸分析时，难以确定该 `interface{}` 里面的 data 指针，传入函数体内后会如何被使用（栈上分配要求[堆上指针不可指向栈](https://github.com/golang/go/blob/64b260dbdefcd2205e74d236a7f33d0e6b8f48cb/src/cmd/compile/internal/escape/escape.go#L22)），因此保守的做法就是产生一次额外的 `runtime.convT64` 内存分配动作。

经过对比，我们可以假设，就是因为这一次额外的堆内存分配，导致了性能的差异。

#### 3.2 改变数据类型

根据上述推断，如果在转换 `interface{}` 时不再需要进行堆内存分配，也许两者的差异就会消失。设想，假如传入的参数已经在堆上了，不就不需要额外分配一次了吗？（这里主要为了测试 buffer 本身的性能，把堆上分配的开销排除在了测试外）。

一次，直接给`Offer()` 传入一个已经在堆上分配过的对象指针，我们得到了如下的测试结果：

{% iframe with-capacity-pointer.html 100% 500 %}

看到上图，不需要任何计算，肉眼也能看出来二者几乎一致。既然从测试结果能印证我们的假设，我们再看看汇编：

```assembly
###### interface{}
	...
	MOVQ	24(R8), DX           ## Buffer.iface.fun 得到接口第一个方法地址
	LEAQ	(DI)(R10*8), CX      ## interface{} 类型的 value：eface.data
	MOVQ	SI, AX               ## receiver
	LEAQ	type.*int(SB), BX    ## interface{} 类型的 value：eface._type
	CALL	DX                   ## Offer(receiver, eface._type, eface.data)
	...
	
###### generics
	...
	MOVQ	24(R8), CX           ## CX = Buffer.iface.fun
	LEAQ	(DI)(R10*8), BX      ## 入参放到 BX
	MOVQ	SI, AX               ## receiver
	CALL	CX                   ## Offer(receiver, int value)
	...
```

可以看到在传入了指针后，`interface{}` 不再调用  `runtime.convT64(SB)` 了。另外，从第 6 行也可以发现，type 变成了 `*int`。

另外，我们省略了对 `Hybrid` 类型汇编结果的展示。考虑到第一节的对比中，`Hybrid` 类型的性能提升相比 `NodeBased` 更不明显，我认为主要原因是 `Hybrid` 本来就更慢，因此减少一次堆内存分配所能产生的效应就更小（[Amdahl's Law](https://en.wikipedia.org/wiki/Amdahl%27s_law)）。



### 4. 总结

在前面的文章中，我们先通过测试发现，将 lfring 改造为泛型后，性能有大约 5%~10% 左右的提升。之后分析了 go 泛型的原理以及对几种泛型编译结果的解读。最后通过对 lfring 的泛型、`interface{}` 两种形式的性能、编译结果的对比，得出了性能提升的主要原因是：对非指针类型参数，`interface{}` 在转换过程中需要一次额外的堆内存分配，而泛型不需要。

但由于性能测试中的数据类型恰好选取了值类型而非指针，因此使测试得出了性能更优的片面结论，在全面了解了泛型的实现原理并做了测试后，我们最终可知：只在实际类型是值类型时才会提升性能，而指针类型并不行。

基于第二节对于泛型的分析，我们还发现了 go 当前泛型实现的缺陷：对于传入接口类型的泛型参数，调用泛型函数需要进行两次内存寻址才能找到正确的方法地址，在这种场景下，泛型反倒会降低性能。

最后，引用 Vicent Marti 的文章 [Generics can make your Go code slower](https://planetscale.com/blog/generics-can-make-your-go-code-slower) 中最后对于 go 泛型使用场景的总结：

- **建议用泛型** 的`ByteSeq` 约束，来消除接收 `string` 和 `[]byte` 参数的相同行为方法。泛型生成的 GCShape 和手写的非常相近。
- **建议用泛型** 在数据结构中。这种方式是目前为止最合适的：先前在数据结构中用 `interface{}` 来实现泛型的方式又复杂又不友好（un-ergonomic）。用了泛型之后，可以在类型安全的方式下存储拆箱类型，因此类型推断不再需要了。这样既简单，又高效。（本文讨论的例子，实际上就是应用了这一条建议）
- **建议用泛型** 来约束回调函数参数的类型。在某些情况下 Go 编译器还能将回调函数直接内联拍平。
- **不建议用泛型** 来尝试去虚化（de-virtualize）或内涵方法调用，这根本没用。这是因为所有指针类型的 GCShape 都一样，所以真正的方法信息还是存放在运行时 dictionary 中的。
- **不建议用泛型** 在需要给泛型函数传入一个接口的任何场景。因为实例化 GCShape 时，对于接口类型参数，泛型并没有降低虚化程度，反倒还引入了额外的一层来在全局哈希表中查找方法。只要在性能敏感的场景，都不要传入接口，而是传入实例指针。
- **不建议用泛型** 重写基于接口的 API。由于目前泛型实现的限制，使用接口（除了 `interface{}`）的行为要比用泛型更容易预测。泛型在进行方法调用时有可能会产生的两次间接查找，坦率的讲这很可怕。
- **也别忧伤**，因为 Go 泛型在实现中并未存在任何技术上的限制来阻止它最终发展成以真正的单态化（monomorphization）方式来内联和去虚化方法调用的实现。 



### Reference

1. [Go 1.18 Implementation of Generics](https://github.com/golang/proposal/blob/master/design/generics-implementation-dictionaries-go1.18.md)
2. [Generics implementation - GC Shape Stenciling](https://github.com/golang/proposal/blob/master/design/generics-implementation-gcshape.md)
3. [Generics implementation - Dictionaries](https://github.com/golang/proposal/blob/master/design/generics-implementation-dictionaries.md)
4. [Generics implementation - Stenciling](https://github.com/golang/proposal/blob/master/design/generics-implementation-stenciling.md)
5. [Generics can make your Go code slower](https://planetscale.com/blog/generics-can-make-your-go-code-slower)
6. [Type Parameters Proposal](https://go.googlesource.com/proposal/+/HEAD/design/43651-type-parameters.md)
6. [Go Data Structures: Interfaces](https://research.swtch.com/interfaces)
