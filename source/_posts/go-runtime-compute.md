---
title: Go Runtime 设计：计算资源调度
mathjax: true
date: 2022-03-10 00:34:44
tags: 
- source
- go
categories:
- Golang
---

## 1. 为什么需要 GoRoutine？

### 1.1 多少个线程是最优的？

我们知道，计算机执行一项任务，通常情况下需要由计算、存储、I/O 等多个组件配合运转才能完成。由于 CPU 与其他设备之间速度的巨大差异，我们更倾向于利用多任务同时运行的方式，来最大化的利用 CPU 的计算能力，这也就是并发编程的意义所在。

而由于多核处理器的广泛普及，在多个 CPU 核心上 ”并行的“ 进行多任务 ”并发“，是编写高效程序的必经之路。从程序执行层次的角度看，”并行“ 更倾向于在底层语境下，指代多个 CPU 核心能同时执行代码片段，而 ”并发“ 更倾向于在高层语境下，指代多个任务能同时在计算机上运行。

OS 通过调度机制帮我们实现了将用户层面的多任务并发，映射到硬件层面的多核心并行。从最大化资源利用的角度讲（暂时抛开任务执行公平性不谈），其映射机制，是对 CPU 资源的一种 “超卖”：任务可能处于执行和等待（包括阻塞）两种状态，执行状态下需要 CPU 资源而等待状态下则可以出让 CPU 资源给其他任务使用。根据任务类型的不同，通常可能分成 CPU 密集型任务与 I/O 密集型任务。

那么理论上，到底要同时执行多少个任务（线程数），才能最大化的利用计算资源呢？《Java 并发编程实战》中给出了如下公式：
$$
N_{threads}=N_{cpu}*U_{cpu}*(1+\frac{W}{C})
\\
\\
N_{threads}=number\ of\ CPUs
\\
U_{cpu}=target\ CPU\ utilization,\ 0\leqslant U_{cpu} \leqslant 1
\\
\frac{W}{C}=ratio\ of\ wait\ time\ to\ compute\ time
$$
显然，基于资源最大化考虑，我们期望 $U_{cpu} \to 1$。

那么，对于计算密集型任务，随着计算占比的不断提高，其 $\frac{W}{C} \to 0$，因此 $N_{threads} \to N_{cpu}$ ；而对于 I/O 密集型任务，随着 I/O 等待占比的不断提高，其 $\frac{W}{C} \to \infin$ ，因此 $N_{threads} \to \infin$。



### 1.2 线程越多越好吗？

前面我们看到了，对于 I/O 占比较高的 I/O 密集型任务，理论公式中倾向于创建更多的线程来填补 CPU 的空闲，但这并不是零成本的。

关于线程所带来的开销，Eli Bendersky 在他的博文 [Measuring context switching and memory overheads for Linux threads](https://eli.thegreenplace.net/2018/measuring-context-switching-and-memory-overheads-for-linux-threads/) 中做了一些测量，

1. 上下文切换与启动的开销：

   ![](https://eli.thegreenplace.net/images/2018/plot-launch-switch.png)

​		可以看到在线程绑核切换下，上下文切换的开销每次大约 2us，线程启动开销大约 5us。

2. 内存开销：

   线程的内存开销主要体现在线程栈，他的[代码示例](https://github.com/eliben/code-for-blog/blob/master/2018/threadoverhead/threadspammer.c)表明，10000 个各持有 100KiB 实际栈空间消耗的线程，virtual memory 约为 80GiB（未实际分配），RSS 约为 80MiB。

Eli Bendersky 的文章主要想表达的是现代操作系统的线程开销已经非常小了，很多时候我们并不需要采用事件驱动等方式来增加复杂度，目前的操作系统支持数万个线程绰绰有余。

但假如我们想要数十万、上百万的线程呢？假如在不增加复杂度的前提下，能做到更低的开销呢？golang 用 goroutine 给出了解决方案。

在 Eli Bendersky 的文章中，测试线程切换的用例是让两个线程通过一个管道往复传递数据，结果是在一秒内，大概能来回传递 40 万次。而后他又顺手用 go 重写了[测试代码](https://github.com/eliben/code-for-blog/blob/master/2018/threadoverhead/channel-msgpersec.go)，得到的结果是：每秒 280 万次。

不论如何，无节制的创建新线程，最终一定会产生许多安全性问题，如过多的上下文切换，内存耗尽等等。

> 实际上谈论线程切换的开销时，涉及到的点比较多，也很难给出绝对正确的设计决策。
>
> 在 Google 的[这个视频](https://youtu.be/KXuZi9aeGTw)中，提到了明确的的数据：
>
> 1. 线程切换的开销在 1~6 us 之间，其差异在于同一个 CPU  核心内切换，所花费的时间可以低至 1us，而在不同核心之间切换的开销可能会达到 6us。设置了 CPU Affinity 后的测试结果证明了这一点。
> 2. 然而我们不能简单的将所有线程绑核了事。毕竟虽然绑核可以提升性能，但当存在 idle core 时，其他线程可能无法调度到该  core 上。
> 3. 进一步的数据显示，线程切换中，进入内核态的开销只有最多 50ns （通常我们认为在用户态和内核态之间切换十分耗时），大部分开销都是由于现代调度器复杂的调度决策导致在不同的 core 之间发生线程调度。



### 1.3 限制最大线程数

既然我们不能容忍无限制创建线程，那么最直接想到的自然是设定一个线程数上限，当线程超限后，拒绝再创建新的线程。

线程池是最通用的解决方案。

![](https://upload.wikimedia.org/wikipedia/commons/thumb/0/0c/Thread_pool.svg/2880px-Thread_pool.svg.png)

池化是资源复用的常见方式，线程池可以最多持有 n 个工作线程（当然根据工作负载的变化，n 可以是动态的），同时持有一个任务队列。工作线程执行如下的循环：从队列获取任务 -> 执行任务 -> 再次从队列获取任务，因此如果没有空闲的工作线程，任务就必须在队列等待。

线程池不仅能限制线程的最大数量，同时也能降低线程反复创建、销毁产生的开销。对于突发的大规模任务也能比较优雅的实现降级、削峰填谷等措施。

不过，简单使用线程池，一个任务对应一个线程的这种同步并发编程模型没有改变。

对于同步并发模式，显然有其固有的优势：

- 程序清晰简单，易于实现
- 线程本地变量易于分配和回收

当然除了可能创建过多线程产生的资源问题以外，还有额外的劣势：

- 由于粒度较粗，任务内嵌套的可并行部分（如多个 I/O 操作等），难以并行化。本质上是无法真正将 CPU 操作和 I/O 操作分开，而由于 CPU 操作和 I/O 操作的差异性，频繁在 CPU 操作之间进行上下文切换，有害无益。



### 1.4 换种思路

对于 I/O 密集型的任务，执行过程中有很大一部时间都在等待，当 I/O 返回时任务才能继续工作。也正是这种等待的特点，给了我们创建多线程来提高 CPU 利用率的理由：阻塞等待中的线程不需要 CPU 时间。

设想假如我们将这种机制反过来，线程不是阻塞等待被唤醒，而是主动询问所有正在等待的 I/O，检查某个 I/O 是不是返回了。如果返回了，就处理与之关联的任务，而如果没有返回，线程就继续检查下一个等待中的 I/O，或者创建新的 I/O 调用。

与阻塞唤醒的被动式相比，询问的方式会更加主动。原先的 “执行任务 -> I/O 阻塞 -> 继续执行” 的流程，变成了 “执行任务 -> 注册 I/O 事件 -> 回调任务“。

这种模式称之为事件驱动的并发编程模型，线程进行轮询（poll）的动作，称为事件循环。

![](https://miro.medium.com/max/1400/1*rWGbyCbcJTKI-m3ZEDhCaA.png)

从工作原理上我们就能发觉，事件驱动模型有如下的特点：

- 不需要很多线程：与多线程通过阻塞出让 CPU 相比，主动切换任务继续使用 CPU 资源，实际上是绕过了系统调度器
- 需要通过回调函数来保持事件与任务的关联关系：通过事件回调来继续执行先前被中断的任务
- 主线程不允许存在阻塞：
  - 在多线程模型下对资源的阻塞等待式访问，需要全部替换成非阻塞式访问，否则一旦出现阻塞，将导致事件轮询线程无法继续轮询。
  - 对于阻塞 I/O 实际上还是需要引入线程池，但此时的 I/O 线程只负责 I/O 操作，不再负责处理任务逻辑




### 1.5 Callback Hell

在事件驱动模型里，将耗费时间的 I/O 阻塞调用交给线程池进行异步化，在阻塞调用返回后，通过调用 callback 函数来恢复执行任务逻辑。

这种方式在简单的任务逻辑中运行的很好，然而当存在一个任务，其整个逻辑链条中包含了多个相互依赖的阻塞 I/O，这时 callback 函数的注册链路会不断加深，最后形成难以理解的 ”Callback Hell“。

![](https://miro.medium.com/max/1400/1*zxx4iQAG4HilOIQqDKpxJw.jpeg)

产生 Callback Hell 的本质是什么？需要通过参数传递上下文。

注册回调函数时，将回调函数地址作为参数传递给事件注册器，是为了能够在合适的时机被调用。回调函数内访问的外部变量，是由编译器默默地通过闭包传递（不支持闭包的语言需要在堆上分配对象，并通过参数传递其地址）。

因为没有外部协助，所以我们需要在应用代码中通过回调函数进行上下文传递，随着传递次数的增多，就导致了回调地狱。

那么，假设：

1. 如果能通过一些手段更优雅的维护任务的上下文，就不需要在参数中层层嵌套传递上下文
2. 如果不需要嵌套回调函数，就能像写同步阻塞的多线程代码一样写事件驱动的异步代码，进而方便的实现任务间的交互协作

基于上述讨论，我们自然会发现，通过将应用逻辑拆分成一个个小的异步任务（而不是同步函数调用），并且通过合理的方式维护任务上下文，我们能够实现任务间的切换和调度。

这已经覆盖了系统调度器的绝大部分工作内容（除了抢占，事件循环类似于协作式调度），任务可以类比为线程，不同点在于任务之间切换是协作式的（等待资源时主动出让 CPU），假如一个任务不主动出让线程，他就能永久的拥有该线程。对于这种执行协作式任务的模型，我们可以称之为协程（co-routine）模型。

> 这里需要注意的是，协程模型的提出相比线程模型更早。线程通过抢占式调度解决了协程的协作式调度对资源使用的的非公平性。

对于维护上下文的问题，协程模型的解决方式有两种：

- 有栈式：通过保存、恢复现场，将协程的调用栈保存在协程结构内部
- 无栈式：将协程之间的上下文保存在外部，常见的办法是有限状态机



### 1.6 用户级调度器

前面讨论完后，我们发现通过事件驱动 + 异步 I/O + 优雅切换上下文的办法，可以比较高效且友好的将应用逻辑中的 CPU 处理部分和 I/O 处理部分分开来执行，同时还不降低代码逻辑的完整性。

此时此刻，只剩下如下的两个问题未能解决：

1. 饥饿问题：某些任务由于各种原因，长时间占据 CPU 时间，导致其他任务饥饿，可能产生严重的不公平。
2. 线程管理问题：I/O 线程池如何分配更合理；并行的任务之间，如何通信和处理数据竞争。

对于问题 1，需要引入抢占式调度，在合适的时机对任务触发抢占，强制该任务出让 CPU。对于问题 2，可以抽象线程管理层，向下管理系统线程，向上提供任务之间的并发原语。

最终，我们就能得到一个构建在操作系统线程之上的用户级调度器。

它：

- 将任务作为调度基本单元
- 支持并发的任务协作与抢占，妥善处理数据竞争
- 向任务屏蔽阻塞的系统调用
- 能够基于任务编写同步风格的代码

以上就是 golang 调度器的大致特性，golang 中的任务正是 goroutine。

{% asset_img g-m-p-sched.svg %}

由于引入了完整的调度器抽象，golang 便有能力将 goroutine 与 channel 结合，实现了 CSP 并发模型，将任务之间的通信和数据竞争转化为对象所有权的传递，优雅的解决了并发通信问题（*Do not communicate by sharing memory. Instead, share memory by communicating.*）。



## 2. 什么是 G-P-M 模型？

### 2.1 基本调度理论

[调度](https://en.wikipedia.org/wiki/Scheduling_(computing))，就是分配*资源（resource）*用以执行*任务（task）*的动作。

这里的资源，可以是计算资源如 CPU，存储资源如内存空间，网络资源如带宽、端口等。任务是基于资源，所执行的动作，它依赖资源并通过操作资源来产生价值。

**调度目标**

根据不同的资源、任务以及业务目标，调度器的设计目标是多样的：

- 最大化吞吐量：效率优先，目的是让任务能尽可能充足的利用资源，而不是把资源花费在调度或等待上。
- 最小化等待时间：体验优先，目的是让尽可能多的任务开始执行，效率和任务的实际完成时间次要考虑。
- 最小化延迟和响应时间：体验优先，目的是尽可能让每一个任务都等待相对较少的时间，且能相对较快的执行。
- 最大化公平：公平优先，目的是结合任务的任务优先级，以及单位资源的负载率，尽可能公平的将资源和任务匹配。

显然，上述目标之间非但不相辅相成，反而很可能相互掣肘（比如最大吞吐和最小延迟），因此选定调度器的设计目标必须结合实际的业务目标。

**调度原理**

之所以需要调度，是基于这种假设：资源通常是有限的，而需要执行的任务比资源多得多。假如任务比资源还少，那么就没有调度的必要了。

因此调度器的工作原理就是根据当下任务、资源的状态，基于特定的调度策略，做出调度决策：接下来哪些 task 将会拥有哪些资源。

我们可以得出，调度器在逻辑层面的样子：

{% asset_img logical-sched.png %}

**映射到 Go**

上述调度器原理，如果映射到 Go，显然 goroutine 是 task，操作系统线程是 resource。因此调度过程就是将选中的 goroutine 放到选中的线程上执行。

此外还需要考虑几个细节问题：

1. 如何组织待执行的 goroutine ?
   - 平衡查找树：提高查找效率，适用于经常需要取出特定的 goroutine
   - 堆：用堆来实现优先级队列，适用于 goroutine 区分优先级的场景
   - 链表：存储为普通队列，适用于每一个 goroutine 都是相对平等的
2. 如何组织线程资源？
   - 无界线程池：可能会经常创建或销毁大量线程，类似于 1：1 的映射关系，不适用 M:N 的场景
   - 有界线程池，容量等于 CPU 核数，绑核：线程与 CPU 核数 1：1，可以最大限度降低操作系统的线程切换，但假如 goroutine 触发系统调用，会阻塞整个线程
   - 有界线程池，灵活调整容量：根据 goroutine 数量灵活调整线程数，对于执行 goroutine 的线程保持最多与 CPU 核数一致，当进行系统调用时创建新线程，这样不会阻塞其他 goroutine，但线程管理更复杂
3. 何时触发切换？
   - 系统调用：系统调用会阻塞线程，因此当有任务执行系统调用时，触发切换，并将系统调用的执行放到专门的线程中
   - 协作：多个 goroutine 间协作，由于效率差异导致可能导致等待，出现等待时触发切换
   - 抢占：为了避免单个 goroutine 占据过多的 CPU 时间，需要定期扫描，将执行时间过久的 goroutine 换出
   - 主动触发：将触发调度权交给 goroutine，业务上可以选择主动放弃 CPU
4. 如何实现切换？
   - 保存上下文：保存 PC、SP 以及其他通用寄存器，保存 goroutine 私有栈
   - 恢复上下文：恢复待换入的 goroutine 的 PC、SP 以及通用寄存器，和其私有栈



### 2.2 集中式调度器

基于前文所述，我们应该很容易的就能想象出一个 go 调度器的雏形：

{% asset_img central.jpg %}

显然，在 go 语言演进的初期，其调度器也是类似这个样子的。其特点有：

1. 所有 goroutine 都进入一个全局队列，用 g 表示
2. 线程分执行 goroutine、执行 Syscall、空闲三种，用 m 表示
3. 由 m 进入调度逻辑，触发调度，从全局队列中获取新 g，替换 curg

这种调度方式很直接的反映了调度器需要做的事情：把任务（g）分配到资源（m）上。我们也称这种方式为集中式调度。

如果看 runtime/proc.go 的代码，在文件顶部注释中，引用了 go 调度器的[设计文档](https://golang.org/s/go11sched)，在设计文档中提到了上述集中式调度器存在的问题：

1. 由于中心化存储所有状态，多线程调度时需要抢锁，需要锁保护的操作包括creation、completion、rescheduling 等，文中的测算数据是有大约 14% 的开销花在了对 futex 的锁竞争上
2. 由于调度决策导致的同一个 g 在多个 m 之间往复执行（handoff），产生额外的延迟和开销（回顾 1.2 节的线程切换开销）
3. 在 g 运行过程中的栈、小对象等等，都会存放在 m.mcache 缓存中，每当创建新的 m 时都会分配 mcache，但当 m 在执行 syscall 的时候，并不需要 mcache。在某些情况下执行 g 的 m 与其他 m 的比例可能高达 1：100，这导致：
   - 没用的 mcache 产生额外的资源消耗
   - 当 g 切换到不同的 m 上时，在 mcache 上加载关联的栈、对象等，会降低 data locality
4. 由于系统调用导致 g 频繁的在不同的 m 上切换，产生大量开销



### 2.3 P 来了

根据前面提到的破坏性能的场景，我们期望能做出如下的改变：

- 尽量减少调度器抢锁，改善调度等待
- 尽量降低同一个 g 在不同 m 之间切换的概率，提升 data locality
- 尽量剥离非 m 必须的属性（如 mcache），降低资源浪费

为了达成上述目标，Dmitry Vyukov 在他的[设计文档](https://golang.org/s/go11sched)里，引入了 p 的概念。

P 代表 processor，从 go 调度器的角度看，可以理解为逻辑处理器。即将 m、syscall、I/O 等等概念屏蔽到 p 以下，逻辑上只有 g，g 只在 p 上运行，类比线程在 CPU 上运行。

在 m 的层面，除了原本的 m 执行 g 不变以外，要求 m 想要执行 g，必须先和某一个 p 绑定，g 相关的状态、上下文、对象等等都保存在 p 内。

如此可以引出完整的 G-M-P 模型图：

{% asset_img gmp.png %}

p 的数量默认与 CPU 核数保持一致，每个 p 里面都保存有一个自己的私有 g 队列，当 m 要执行 g 时，需要先绑定 p，并且从 p 的私有队列中获取 g。

这样一来，前面的目标悉数实现：

- 每一个 g 在需要被调度时，m 都会在尝试在绑定的 p 上调度，调度参与方只有单线程 m 和 p 的私有 g 队列，不需要加锁（实际上由于 p 可能会在 m 之间传递，还是需要用 cas 操作队列，但争抢概率大大降低）
- 当 m 与 p 绑定后，调度所依赖的数据和操作大都在当前 m 上（p 的私有队列甚至是一个 ring），这可以有效利用 CPU 的缓存、预取等优化手段
- 原本的 mcache，现在放在了 p 处，这样数据随 p 移动，与 m 彻底脱钩了

那么这里有一个新问题：g 都保存在 p 的本地队列中，由于调度不均衡，导致有的 p 空闲，有的 p 负担过重怎么办？

引入全局队列。

当出现某个 p 的私有队列空/满，导致无法取出/存入 g 时，将从全局队列中批量取/存一部分 g。全局队列用链表实现，无界，一般不会塞满。此外，因为通常 p 不会从全局队列中拿 g，为了保证一定的公平性（不至于全局队列中的 g 饥饿），每经过一定的调度次数后，就会强制从全局队列获取一个 g。

{% asset_img gmp-globrunq.png %}

假如全局队列也空了呢？

工作窃取。（这里省略了一些检查 gc、定时器、netpoll 等等动作，通常工作窃取是最后的选择）

从其他 p 的队列尾部，窃取一半的工作，转移到当前 p。

{% asset_img gmp-steal.png %}

要是实在没有任务了呢？就只能让当前 m 陷入睡眠，p 进入 idle 队列，共同等待新任务到来。



## 3. 如何实现调度？

## 4. 如何实现抢占？
