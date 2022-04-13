---
title: Go Runtime 设计：存储资源管理
mathjax: true
date: 2022-04-06 21:12:05
tags: 
- source
- go
categories:
- Golang
---

<img src="https://raw.githubusercontent.com/MariaLetta/free-gophers-pack/8f7fbe7906dd4433a5719df73d3dde6f481b459f/goroutines/svg/15.svg" width="500;" />

本文介绍了 Golang Runtime 中关于内存管理的设计。

<!-- more -->



## 1. Go 的内存划分

### 1.1 进程内存

不论 Go 在 Runtime 中如何组织 goroutine 的内存布局，从操作系统的视角来看，加载一个 Go 可执行文件，与加载其他可执行文件没什么区别，最终都会以进程的形式运行在操作系统上。

因此在具体分析 Go 自己的内存布局和管理之前，我们先来回顾一下 Linux 的内存布局（图源：*CSAPP Figure 9.26*）：

{% asset_img mem-layout.jpg %}

Linux 进程的整个虚拟地址空间，从 0x400000 开始，到栈区高地址结束，更高的地址空间留给内核。

进程内存段从低地址开始分别被划分为：

- 代码段 `.text`
- 已初始化数据段 `.data`
- 未初始化数据段 `.bss`
- 运行时堆 `heap`
- 共享库内存映射区 `mmap`
- 运行时栈 `stack`

上述各分段的意义不再赘述，我们任意找到一个运行中的进程，通过访问 `/proc/{pid}/maps` 来看一看实际的内存布局：

```shell
[root@xxx ~]# pmap -X 1410
### 可以发现我们执行的是 sleep 1000
1410:   sleep 1000
         Address Perm   Offset Device  Inode   Size Rss Pss Referenced Anonymous Swap Locked Mapping
### .text
        00400000 r-xp 00000000  fd:01 789707     24  16  16         16         0    0      0 sleep
### .data
        00606000 r--p 00006000  fd:01 789707      4   4   4          4         4    0      0 sleep
### .bss
        00607000 rw-p 00007000  fd:01 789707      4   4   4          4         4    0      0 sleep
### heap，ptmalloc 堆初始空间为 132KiB
        01963000 rw-p 00000000  00:00      0    132  12  12         12        12    0      0 [heap]
### mmap，下同
    7f81ea7b0000 r--p 00000000  fd:01 801455 103692  48  10         48         0    0      0 locale-archive
    7f81f0cf3000 r-xp 00000000  fd:01 788272   1808 416  41        416         0    0      0 libc-2.17.so
    7f81f0eb7000 ---p 001c4000  fd:01 788272   2044   0   0          0         0    0      0 libc-2.17.so
    7f81f10b6000 r--p 001c3000  fd:01 788272     16  16  16         16        16    0      0 libc-2.17.so
    7f81f10ba000 rw-p 001c7000  fd:01 788272      8   8   8          8         8    0      0 libc-2.17.so
    7f81f10bc000 rw-p 00000000  00:00      0     20  12  12         12        12    0      0
    7f81f10c1000 r-xp 00000000  fd:01 786459    136 108   9        108         0    0      0 ld-2.17.so
    7f81f12d7000 rw-p 00000000  00:00      0     12  12  12         12        12    0      0
    7f81f12e1000 rw-p 00000000  00:00      0      4   4   4          4         4    0      0
    7f81f12e2000 r--p 00021000  fd:01 786459      4   4   4          4         4    0      0 ld-2.17.so
    7f81f12e3000 rw-p 00022000  fd:01 786459      4   4   4          4         4    0      0 ld-2.17.so
    7f81f12e4000 rw-p 00000000  00:00      0      4   4   4          4         4    0      0
### stack，默认容量 132KiB    
    7fff33464000 rw-p 00000000  00:00      0    132  12  12         12        12    0      0 [stack]
    7fff335a7000 r-xp 00000000  00:00      0      8   4   0          4         0    0      0 [vdso]
ffffffffff600000 r-xp 00000000  00:00      0      4   0   0          0         0    0      0 [vsyscall]
                                             ====== === === ========== ========= ==== ======
                                             108060 688 172        688        96    0      0 KB
```

前面是一个简单的 sleep 进程的内存布局，主要作为一个基准，用于和 Go 进程作比较。

现在来看一看一个 Go 进程的内存布局：

```shell
[root@xxx ~]# pmap -X 11564
### vmlet 是一个用 go 实现的 agent 程序，代码规模大约是 3-5k 行
11564:   /opt/vmlet/vmlet run -c ./vmlet.yaml
         Address Perm   Offset Device  Inode   Size   Rss   Pss Referenced Anonymous Swap Locked Mapping
### .text
        00400000 r-xp 00000000  fd:01 524358   6724  2272  2272       2272         0    0      0 vmlet
### .data        
        00a91000 r--p 00691000  fd:01 524358   7176  2644  2644       2620         0    0      0 vmlet
### .bss
        01193000 rw-p 00d93000  fd:01 524358    488   180   180        180       120    0      0 vmlet
### mmap，可以看到下面全部都是匿名的 mmap（没有 fd）      
        0120d000 rw-p 00000000  00:00      0    272   108   108        108       108    0      0
      c000000000 rw-p 00000000  00:00      0  81920  4252  4252       4200      4252    0      0
      c005000000 ---p 00000000  00:00      0  49152     0     0          0         0    0      0
    7f7eb7d55000 rw-p 00000000  00:00      0  40648  4884  4884       4884      4884    0      0
    7f7eba507000 ---p 00000000  00:00      0 263680     0     0          0         0    0      0
    7f7eca687000 rw-p 00000000  00:00      0      4     4     4          4         4    0      0
    7f7eca688000 ---p 00000000  00:00      0 293564     0     0          0         0    0      0
    7f7edc537000 rw-p 00000000  00:00      0      4     4     4          4         4    0      0
    7f7edc538000 ---p 00000000  00:00      0  36692     0     0          0         0    0      0
    7f7ede90d000 rw-p 00000000  00:00      0      4     4     4          4         4    0      0
    7f7ede90e000 ---p 00000000  00:00      0   4580     0     0          0         0    0      0
    7f7eded87000 rw-p 00000000  00:00      0      4     4     4          4         4    0      0
    7f7eded88000 ---p 00000000  00:00      0    508     0     0          0         0    0      0
    7f7edee07000 rw-p 00000000  00:00      0    384    40    40         40        40    0      0
### stack，默认容量 132KiB    
    7ffe4432c000 rw-p 00000000  00:00      0    132    16    16         16        16    0      0 [stack]
    7ffe44353000 r-xp 00000000  00:00      0      8     4     0          4         0    0      0 [vdso]
ffffffffff600000 r-xp 00000000  00:00      0      4     0     0          0         0    0      0 [vsyscall]
                                             ====== ===== ===== ========== ========= ==== ======
                                             785948 14416 14412      14340      9436    0      0 KB
```

从上面两个不同进程的内存布局中，我们能发现几个有趣的地方：

1. sleep 有`[heap]`，go 进程没有`[heap]`？
   - 对于`Mapping = [heap]`，只有当程序使用 `brk()` 分配堆内存后才会显示，以 `mmap()` 的形式分配的内存不显示为 `[heap]`
   - sleep 属于 glibc 库，glibc 默认的 ptmalloc 初始化时会通过 `brk()` 分配 132KiB 的堆空间
   - 后文会提到，go runtime 使用的内存分配器，是完全采用 `mmap()` 来分配内存的，没有调用过`brk()`，也就不显示 `[heap]`
2. 为什么 `[stack]` 的默认容量是 132KiB？
   - Linux 在执行 exec 时，初始化栈空间的逻辑中为栈顶[分配了 `PAGE_SIZE` 的空间](https://github.com/torvalds/linux/blob/1862a69c917417142190bc18c8ce16680598664b/fs/exec.c#L272)，amd64 架构下 linux page size 默认是 4KiB
   - 在之后的 [`setup_arg_pages()`](https://github.com/torvalds/linux/blob/1862a69c917417142190bc18c8ce16680598664b/fs/exec.c#L747) 中，对栈进行扩展，[扩展容量是硬编码的 128KiB](https://github.com/torvalds/linux/blob/1862a69c917417142190bc18c8ce16680598664b/fs/exec.c#L836)，4 + 128 = 132KiB，上面两个进程对栈的使用都没有超过 132KiB，所以 `[stack]` 都是 132KiB
   - 后文会提到，goroutine 的栈内存，也全都是从 `mmap()` 而来，只有主进程所在的 g0 栈直接使用系统栈，但 g0 限制了栈空间最大 8KiB，不会超出 132KiB

### 1.2 栈内存

在[本系列上一篇中](https://lenshood.github.io/2022/03/09/go-runtime-compute/)，我们已经知道，goroutine 持有自己的运行栈。

栈在创建 g 的时候需要一并分配出来，在销毁 g 的时候也需要一并清除掉。那么 go runtime 是从哪里为 g 分配栈内存的呢？

{% asset_img stack.jpg %}

上图展示了 g 的栈分配结构，在分配栈内存时，首先需要根据栈空间的大小来决定是从哪里分配。

1. 每个 P 都会持有一部分称为 `stack cache` 的空间，专门用于分配较小的栈
2. 假如需要分配的栈空间比较大，就会选择直接从全局的 `large stack pool`中进行分配

对于较小的栈，按照尺寸划分成 4 阶（linux 系统，其他系统阶数会有差异），每一阶的内存块分别指定为 2KiB、4KiB、8Kib、16KiB。对于小栈分配，根据所需空间大小向上取二次幂，然后从合适的阶中分配内存。

#### stack cache

由于小栈空间占比小，分配的频次也比较频繁（创建 g 默认栈2KiB，之后逐步扩容），因此在每一个 P 中都存放有一个本地的`stack cache`，从缓存中分配栈，不需要加锁，速度更快。

定义了 `stack cache` 每一阶最大不超过 32KiB，超出后会释放一半，剩 16KiB。而当分配时发现某一阶缓存为空，则会从全局的`stack pool` 中分配总量为 16KiB 的空间放入缓存。

#### stack pool

全局的小栈空间池。按照四个阶存储四条链表，用来保存固定大小的可用内存块。

当 `stack pool` 中可用空间耗尽后，会一次性从全局 heap 中申请 32KiB 的内存，并按阶切分成小块，插入链表。

#### large stack pool

从前面可知，小栈空间单个最大空间块是 16KiB，所以如果需要超过 16KiB 的栈，就需要从 `large stack pool` 中申请。

### 1.3 堆内存



### 1.4 其他内存



## 2. 内存分配器

### 2.1 为什么要有内存分配器？



### 2.2 不同的内存分配器算法



### 2.3 Go 内存分配器设计





## 3. 垃圾收集器

### 3.1 不同的垃圾收集算法



### 3.2 Go 垃圾收集器设计
