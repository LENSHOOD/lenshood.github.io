---
title: title: /Xv6 Rust 0x04/ - Easy Piece of Virtualization (virtual cpu)
date: 2024-11-20 22:50:32
tags:
- xv6
- rust
- os
categories:
- Rust
---

In this chapter, we are going to explore the cpu virtualization, also know as process, in the xv6.

I'm really excited for writing this chapter, because process is one of the fundamental concepts of the operating system. Process is very useful for multiple tasks, and in the design wise, its abstraction is also very elegant.

<!-- more -->

## 1. Overview of Virtual CPU

We all known that with the concept if process, we could run multiple programs on one or a few cpu cores, that makes the process es and the cpu cores present as a many-to-many mappings.

But the key point here is, a process will not need to know how many cpu cores and how many memory it can have. Through some sort of abstraction, a process can freely using all cpu and memory resources to running its program, any resource management should have done by the kernel.

So before we take a look at the design of xv6 process, let's think about what elements a process should have, to ensure the cpu virtualization works.

We can simply recap the machine code we introduced in the first two chapters as an analogy, when we built the "xv6-rust-sample", there are few things that need to be taken care of:

- We should understand the address space then link the program to right places
- There should be a way to load the program and running from the entry point
- Some necessary registers should be initialized, such as stack pointer
- The error handling is also required, like panic

  Additionally, stand in the kernel's shoes, more points turn out:

- There are more processes than cpu cores, how to switch different processes on one cpu?
- Where to save process status when it's switched?
- How to manage the process lifecycle, including create, destroy, and error handling?
- How to allocate and restore memory to and from processes

Actually, if a kernel implements all above points, then it has all elements to run multiple processes. So let's take a first look at the design of xv6 process:

{% asset_img 1.png %}

For the elements contained inside the process: 

The "PID / Name" identifies a specific process.

The "State" field records the current state of the process, the common state are Running / Runnable / Sleeping, which indicates running on cpu, waiting to be scheduled, waiting to be waken up respectively.

The "Open Files" tracks any files that opened by the process, we haven't talked about file system before, but at least we can realize the basic three files that every process will open are STDIN, STDOUT and STDERR.

The "Parent" to track the process parent, like linux, the xv6 also create a new process by `fork()`, therefore, every process should have a parent process.

The "Kernel Stack" allows running kernel code on the address space of a process. After all, for safety purpose, user process cannot share a same stack with kernel.

The "Trap Frame" stores user space process data, this kind of data will be saved and restored when switching between user space and kernel space.

The "Page Table" records the mapping between virtual memory and physical memory. We have described virtual memory in the previous chapter, actually every process should have its own page table.

The "Context" records the basic registers a process is using. When a process needs to be paused and run another process, the current states in the registers should be saved, and once the process can re-run, the registers should be restored.



## 2. Process Memory



## 3. Concurrency



## 4. Scheduling



## 5. Process Lifecycle

