---
title: Linux Tracing - Kprobe
date: 2025-06-29 23:40:44
tags:
- linux
- os
- kprobe
categories:
- Linux
---

自各类计算机程序开始被编写、运行开始，我们就一直想通过各种方式来了解它的执行过程和状态从而判断计算机程序运行的行为和效率。作为被使用最广泛的操作系统，Linux 经过多年发展，拥有了各类工具和组件来实现对用户程序以及内核程序的追踪，这些组件组成了 Linux 的追踪（Tracing）系统，它的魔力令人着迷。

本文将从最基础也是最灵活的 Kprobes（Kernel Probes） 入手，了解 Linux Tracing 系统的设计（本文基于 linux kernel v6.15.4）。

<!-- more -->

## 1. 设计尝试

在了解 kprobes 之前，我们也许可以先设想一下，对于一段程序，如果我们想要了解其执行过程中的每一步，其运行前后程序状态的变化，需要怎么做？这个问题值得思考，事实上在明确更具体的限定条件以前，我们无法作答，因为该程序的类型并不明确。

我们可以分析如下几种情况：

- 假如是裸机程序，由于常见的 CPU 都提供了调试中断或调试异常机制，由此我们可以让程序单步执行，但同时也需要类似硬件调试器的工具来获取暂停后各寄存器的状态
- 假如是运行在操作系统（如 Linux）中的程序，不论是用户态或内核态程序，我们都可以在每一条或需要的任一条指令前后插入软件中断指令使 CPU 陷入中断，进而我们可以在中断处理程序中获取当前程序状态
- 假如是运行在虚拟机（例如 JVM）中的程序，则需要借助特定虚拟机的能力，实现虚拟环境下的 “软件中断”，从而暂停程序运行，这非常依赖具体的虚拟机实现

不过，基于上述分析，我们发现不论是哪种场景下，为了设计一个能了解任意一段程序的执行过程和状态的工具，都需要实现以下两个核心设计要求：

1. 该工具必须拥有随时暂停以及恢复程序运行的能力
2. 该工具可以作为外部观察者观察当前系统的状态

现在，我们对前面的问题加以限定：我们期望能在观测在 Linux 下运行的任意用户态或内核态程序。映射到设计要求：

1. 暂停/恢复程序运行：需要能动态的在程序的代码段特定位置中插入软件中断指令，使 CPU 陷入中断
2. 观察系统状态：在中断处理程序中读取并保存当前的寄存器上下文

显然，为了满足上述设计要求，我们的工具需要有极高的权限从而实现对正在运行的程序进行修改和探测，这要求工具必须运行在内核态。一旦满足了上述基础要求，对工具进行不同方向上的扩展，就能实现差异化能力：

- 为程序创建用户态接口，可支持对程序的运行状态和寄存器值进行控制，这就实现了调试器的功能
- 为程序预留编程扩展接口，可同时对多个程序进行大规模持续观测，这能实现对任意程序的追踪和剖析功能



## 2. Kprobes 案例

在了解了上述设计要求和要点后，我们也许能更容易理解 kprobes。首先，kprobes 能够动态的切入几乎任意内核程序（除了包含在[blacklist](https://docs.kernel.org/trace/kprobes.html#kprobes-blacklist)中的那些）并收集信息（甚至修改寄存器值）。

当切入点被调用前后，kprobes 会执行自定义的 handler 程序。通常，kprobes 的注册、注销以及 handler 程序的定义都被包含在内核模块中，这样对定义了 kprobes 的内核模块进行加载时，kprobes 就能被插入内核中了。

如下是一个十分简单的 kprobes 内核模块代码案例，通过该案例我们就能基本了解 kprobes 的使用：

```c
#include ...

// 定义 kprobe 结构以备后用（想想为什么要定义为全局静态变量？）
static struct kprobe kp;

// 实际的自定义 handler，在切入点被命中后被 kprobes 框架调用
static int handler_pre(struct kprobe *p, struct pt_regs *regs)
{
    const char __user *filename = (const char __user *)regs->si;
    char fname[256];

    if (filename && strncpy_from_user(fname, filename, sizeof(fname)) > 0) {
        fname[sizeof(fname) - 1] = '\0';
        printk(KERN_INFO "[kprobe] openat() called with filename: %s\n", fname);
    }
    return 0;
}

// 注册 kprobes
static int __init kprobe_init(void)
{
    kp.symbol_name = "sys_openat";
    kp.pre_handler = handler_pre;

    if (register_kprobe(&kp) < 0) {
        pr_err("Failed to register kprobe\n");
        return -1;
    }
    return 0;
}

// 注销 kprobes
static void __exit kprobe_exit(void)
{
    unregister_kprobe(&kp);
}

// 内核模块加载/卸载
module_init(kprobe_init);
module_exit(kprobe_exit);
```

如上所示是一个简单的 kprobes 内核模块，它能够在 sys_openat 符号（即 open 系统调用）被调用时尝试打印 filename。其核心在如下三部分：

1. 定义 kprobe 结构，并写入 `kp.symbol_name = "sys_openat";` 以及 `kp.pre_handler = handler_pre;`：这定义了 kprobes 的挂载点符号和 handler 程序，pre_handler 代表将在挂载点被调用前执行
2. 注册 kprobes：`register_kprobe(&kp)` 将 kprobe 结构注册进 kprobes 框架中，触发生效
3. 注销 kprobes：`unregister_kprobe(&kp);` 将 kprobe 结构移除 kprobes 框架

此外，handler 函数传入的参数 `pt_regs *regs` 包含了当前的寄存器信息，这是一个平台相关参数，使用者需要根据体系架构和 ABI 的差异来选择从正确的寄存器中获取需要的值。

通过上述代码可以看到 kprobes 框架的抽象程度很高，使用简单。接下来我们从实际设计的角度进一步探寻其原理。



## 3. 设计原理

在深入 kprobes 的设计原理之前，我们不妨再次以主观的视角来进行思考：假如要设计实现上一节所描述的 kprobes 能力，作为kprobes 框架需要解决哪些问题？

首先，最核心的点是，需要将 handler 的调用插入所监控的切入点，在第一节我们已经分析过，可以通过在切入点前后插入软件中断指令使 CPU 陷入中断并在中断处理程序中调用 handler，这样在执行完成后能够随中断机制再次回到切入点位置。此外，可以想一想，除了软件中断我们能不能直接插入一条跳转指令使程序流直接跳转到 handler 呢？这也许省去了中断处理的开销，但也需要考虑如何跳回。

其次，我们发现在前一节的代码中，有一个 `struct kprobe` 结构来组合包括切入点符号、处理程序等等的关键元素，这些关键元素聚合在一起才能完整的描述一个 kprobe。那么，作为一个面向众多使用者的框架，kprobes 必须考虑如何管理这些`kprobe`结构，从而使每一个注册的 kprobe 都能顺利执行，也能在不在需要的时候被注销。这涉及到对`kprobe`结构的访问、存储、和增删。

最后，还需考虑一些看似边缘实则很容易产生的情况，例如多个程序并发的注册、注销 kprobes 时如何确保并发安全？以及如果尝试为同一个符号注册多个不同的 kprobe 会怎么样？还有，作为对系统运行拥有绝对控制权的内核，考虑是否可以不止将符号作为切入点，是否能将程序的任意一行作为切入点？

随着对这些问题的思考，接下来我们进入原理分析。

### 3.1 管理 Kprobes

从最简单的开始，对于注册的多个 `kprobe` 结构，kprobes 框架需要有一套机制来维护和管理这些 `kprobe`，以便于在切入点被命中时能快速检索到并执行实际的自定义操作。

事实上，在 Linux Kprobes 的设计中，是通过一个 Hash Table 来管理所有 `kprobe` 的：

```c
// kernel/kprobes.c
#define KPROBE_HASH_BITS 6
#define KPROBE_TABLE_SIZE (1 << KPROBE_HASH_BITS)
static struct hlist_head kprobe_table[KPROBE_TABLE_SIZE];
```

显然，`kprobe_table` 是一个拥有 64 个 slot 的线性数组，其每一个数组元素是一个 `hlist_head` 类型的值，hlist 是一个常见的内核数据结构，用于构建通用 Hash Table。`hlist_head` 指向一个链表，其链表元素的类型为 `hlist_node`。事实上 `struct hlist_head name[1 << (bits)]` 这类定义就是一个标准 Hash Table 的模版定义（更多内容可参考[这里](https://kernelnewbies.org/FAQ/Hashtables)）。

因此我们可以将 `kprobe_table` 视为 64 个 hash buckets，为了应对 hash collision，每一个 bucket 中实际存放了一个长度不一的链表。那么为了让 `struct kprobe` 顺利成为 hlist 中的一个元素，其结构内部势必需要包含一个 `hlist_node` 来作为连接节点，参考源码定义，的确如此：

```c
// include/linux/kprobes.h
struct kprobe {
	struct hlist_node hlist;
	... ...
};
```

此外，既然 `kprobe_table` 是一个 Hash Table，那么对于一个新的 `kprobe`，基于什么作为 Hash Key 来判断将其插入的位置呢？答案是通过切入点的 `addr`。

```c
// include/linux/kprobes.h
typedef int kprobe_opcode_t;
struct kprobe {
	struct hlist_node hlist;
	... ...
  /* location of the probe point */
	kprobe_opcode_t *addr;
  ... ...
};

// kernel/kprobes.c
static int __register_kprobe(struct kprobe *p)
{
  ... ...
	INIT_HLIST_NODE(&p->hlist);
	hlist_add_head_rcu(&p->hlist, &kprobe_table[hash_ptr(p->addr, KPROBE_HASH_BITS)]);
  ... ...
}
```

回忆第二节的示例代码，我们仅为 `kprobe` 设置了 `kp.symbol_name = "sys_openat";`。实际上在注册的过程中会将符号名替换为相对地址并存入 `kprobe`，这样就可以将 `addr` 作为 key 来计算 slot index 了。

综上，我们可以绘制如下示意图来描述 kprobes 框架对 `kprobe` 的管理结构：

{% asset_img 1.png %}



### 3.2 Kprobe 结构

上一节我们已经部分了解了 `struct kprobe` 中包含了 `hlist` 和 `addr` 两个关键字段，接下来我们一起通过讨论完整的`kprobe` 结构，来进一步认识 kprobes 框架的工作方式。

```c
// include/linux/kprobes.h

struct kprobe {
	struct hlist_node hlist;

	/* list of kprobes for multi-handler support */
	struct list_head list;

	/*count the number of times this probe was temporarily disarmed */
	unsigned long nmissed;

	/* location of the probe point */
	kprobe_opcode_t *addr;

	/* Allow user to indicate symbol name of the probe point */
	const char *symbol_name;

	/* Offset into the symbol */
	unsigned int offset;

	/* Called before addr is executed. */
	kprobe_pre_handler_t pre_handler;

	/* Called after addr is executed, unless... */
	kprobe_post_handler_t post_handler;

	/* Saved opcode (which has been replaced with breakpoint) */
	kprobe_opcode_t opcode;

	/* copy of the original instruction */
	struct arch_specific_insn ainsn;

	/*
	 * Indicates various status flags.
	 * Protected by kprobe_mutex after this kprobe is registered.
	 */
	u32 flags;
};
```

通过每个 field 的注释我们已经能够大致的了解到它们的用途。

#### 3.2.1 切入点相关的 Fields

首先在前面的内容中已经提到过，在 `kprobe` 中可以指定切入点的 `addr`，也可以指定 `symbol_name + offset` 作为切入点，但 `addr` 和 `symbol_name` 是互斥关系，不能同时存在，否则注册时会报错。实际上在 kprobe 最终运作时，都是以 `addr` 作为实际的切入点位置，而假如指定了 `symbol_name` 则会经过一个转换的过程将之转换为 `addr`。

对符号到地址的转换过程依赖子系统提供的能力：根据配置不同，默认情况下内核中已导出的符号和其地址会由 `kallsyms` 记录并提供运行时查询。因此如果 `kprobe` 中通过 `symbol_name + offset` 的形式描述切入点位置，kprobes 框架会通过 `kallsyms` 接口来对其进行验证并转换为实际的 `addr`。简化流程可见：

{% asset_img 2.png %}

其次，`pre_handler` 和 `post_handler` 会在切入点被命中的前后被执行，简单看看它们的函数签名：

```c
// include/linux/kprobes.h

typedef int (*kprobe_pre_handler_t) (struct kprobe *, struct pt_regs *);
typedef void (*kprobe_post_handler_t) (struct kprobe *, struct pt_regs *, unsigned long flags);
```

第一个相同的参数 `struct kprobe *` 不多做赘述，而第二个相同的参数 `struct pt_regs *` 中保存了当前执行 CPU 的寄存器值。此外，根据 [kprobe 文档](https://docs.kernel.org/trace/kprobes.html#register-kprobe)，`post_handler` 的最后一个参数 `flags` 目前总是为零值。

需要注意的是`pre_handler` 还需要一个`int`返回值。由于 kprobe 运行在内核态，handler 程序实际上可以修改任何寄存器的值从而改变程序执行路径（例如修改 program counter），但这非常危险，在 kprobe 文档的[相关章节](https://docs.kernel.org/trace/kprobes.html#changing-execution-path)中建议：“对于要修改程序执行流程的内核老司机，在修改寄存器值之后记得让 `pre_handler` 返回非零值，进而 kprobe 框架会直接跳转到新的地址，而不再执行原切入点的指令，也不再执行 `post_handler`”。

对于 `kprobe_opcode_t opcode;` 则涉及到了 kprobe 的核心原理，即为了正确命中切入点，kprobe 框架会把切入点的原始指令暂存，并替换为 Breakpoint 指令，例如 x86 架构下的 INT3，这里的 `opcode` 就保存了切入点指令。更多细节我们在下一小节讨论。

相对应的，`struct arch_specific_insn ainsn;` 保存了与体系结构相关的额外信息，`opcode` 是与体系结构无关的字段，它直接保存了指令的内容。而 `ainsn` 进一步保存了与体系结构相关的一些信息，不同的体系结构下其实现可能存在差异，例如 [x86 的实现](https://elixir.bootlin.com/linux/v6.15.4/source/arch/x86/include/asm/kprobes.h#L54)以及 [riscv 的实现](https://elixir.bootlin.com/linux/v6.15.4/source/arch/riscv/include/asm/probes.h#L19)。

此外，`unsigned long nmissed;` 存储了当前 kprobe 被触发，但未能成功执行 handler 的次数，通常情况下，这是由于该 kprobe 发生了重入（reenter），即该 kprobe 触发了另一个 kprobe。在这种情况下，handler 不会再被运行，而是记录一次 `nmissed`。

最后，`u32 flags;` 作为一个状态指示器指示了当前的各种状态，其中包括启用、禁用、已注册等等。

#### 3.2.2 相同切入点上的多个 kprobes

回到前面提到过的一个问题：作为一个通用的框架，假如有多个调用方期望在同一个切入点上挂载多个 kprobe，该怎么办呢？

事实上 kprobes 框架依然采用了线性表的方式来确保这种情况能够被满足。`struct list_head list;` 充当了该职责，`list_head` 是内核中提供的双向链表结构，它非常简单的只包含了前后指针两个字段。与 `hlist_head` 类似，实际使用时嵌入到具体的结构中，通过计算偏移量就能取回实际的结构。

因此，`kprobe` 可以通过 `list` 字段将其他与其具有相同切入点的`kropbe` 通过该双向链表串联在了一起，从而当切入点被命中时，每一个 `kprobe` 都会被依次触发而不会被漏掉。

有趣的是，为了实现依次触发的要求，在每一个`kprobe` 链表的头部，会放置一个特殊的 `kprobe` 结构，名为 “aggregator”。它与普通`kprobe` 的唯一区别在于其 `pre_handler` 和 `post_handler` 会被替换为如下的 handler：

```c
// kernel/kprobes.c

static int aggr_pre_handler(struct kprobe *p, struct pt_regs *regs)
{
  ...
	list_for_each_entry_rcu(kp, &p->list, list) {
    ... ...
    if (kp->pre_handler(kp, regs))
      return 1;
    ... ...
	}
  ...
}

static void aggr_post_handler(struct kprobe *p, struct pt_regs *regs, unsigned long flags)
{
  ...
	list_for_each_entry_rcu(kp, &p->list, list) {
		... ...
    kp->post_handler(kp, regs, flags);
    ... ...
	}
  ...
}
```

很简单的，这些 aggrator handler 实际上遍历了所有链表内的 `kprobe` 并实际调用它们的 handler。

同时，我们现在可以更新一下 kprobes 框架的结构图：

{% asset_img 3.png %}



### 3.3 命中





## 4. Kretprobes
