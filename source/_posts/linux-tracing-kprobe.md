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

本文将从最基础也是最灵活的 Kprobes（Kernel Probes） 入手，了解 Linux Tracing 系统的设计。

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



## 4. Kretprobes
