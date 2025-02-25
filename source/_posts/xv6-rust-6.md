---
title: /Xv6 Rust 0x06/ - User Space
date: 2025-02-24 14:35:03
tags:
- xv6
- rust
- os
categories:
- Rust
---

Based on disk and file management, now we are able to store the user space program on the disk, and let them running after kernel started. But before that, there is still a topic we haven't covered: how does xv6 jump from kernel space to user space? 

After all, the content we talked in previous chapters is only limited in the supervisor level, even machine level, where the code has full control of hardware. However, the user space program cannot be granted such huge scope of control, then we should know how to jump from kernel space to user space, so that we could provide a safer environment for the user program.

In this chapter, we are going to find it out.

<!-- more -->

## 1. Jumping in the CPU Perspective

If you remember, there was a table in our second chapter, describes several CSRs that risc-v provides to user, some of them are responsible for mode switching.

Speaking of how to jumping from supervisor mode to user mode, there would be the following questions come up with in your mind:

- What kind of instruction is able to trigger the switching?
- After jumping to user mode, where exactly the program will go to?
- How to deal with the context and different memory space between two modes?

At first, let's recap the privilege mode switch that we've mentioned in second chapter:

> **How does risc-v deal with the privileged mode switch?**
>
> *.... RISC-V Privileged Specification Chapter 1.2 ...*
>
> *A hart normally runs application code in U-mode until some trap (e.g., a supervisor call or a timer*
> *interrupt) forces a switch to a trap handler, which usually runs in a more privileged mode. The hart*
> *will then execute the trap handler, which will eventually resume execution at or after the original*
> *trapped instruction in U-mode. Traps that increase privilege level are termed vertical traps, while traps*
> *that remain at the same privilege level are termed horizontal traps. The RISC-V privileged architecture*
> *provides flexible routing of traps to different privilege layers.*
>
> *.... RISC-V Privileged Specification Chapter 1.2 ...*
>
> Generally, when a trap happens, the address of where the cause the trap will be saved in `mepc` or `sepc`, regarding the current privileged mode. After trap handled by specific handler, it should call either `mret` or `sret` to return to the previous mode, which is stored in the `MPP` or `SPP` filed of the `mstatus`.

Let's take a close look at the `sret` instruction:

{% asset_img 1.png %}

Apparently, `SRET` doesn't rely any source or destination register, so when using the `SRET`, we only need to call the bare instruction.

According to the specification, *`xRET` sets the `pc` to the value stored in the `xepc` register.* Hence, before `SRET` is called, we could set the address into the `sepc`, then once it called, the program will be jump into the address.

So far, it looks `SRET` does a lot of things for us, so that we'll no longer need to concern about the first two questions. However, in risc-v architecture, no more support will be provided. Now, for the question of context and memory space switch, we are on our own.

Imagine the kernel is about to complete initialization, and program is running on the supervisor mode. Now, the kernel should start creating the very first process in the whole system, we call it `init`. Assuming that a few milliseconds later, the process structure has been created and all of the importance fields have been set, next the kernel must think about runs the `SRET`  instruction, and hands the control of CPU to `init`.

But before calling the `SRET`, both the context and memory space should be replaced as well, because:

- Context Switch: the context here means the general purpose registers, there are two main reasons that the context switch should be done; first, the value of registers in supervisor mode must not be leaked to user mode for safety; second, in other cases like syscall or interrupt handling, we need to make sure when go back to user mode, the user process can resume correctly with all registers still store the origin values, that requires properly context switch too.
- Memory Space Switch: we have known that there is a kernel page table dedicated for kernel code, if we don't set the user process page table after switch mode, then the user process is able to access kernel memory space, which is extremely dangerous; besides, kernel page table does not hold user code in the text section, makes the user process unable to get its code.

Therefore, it's essential for kernel to take care of the context and memory space switch. The following is a diagram that shows the process of switching from supervisor mode to user mode:

{% asset_img 2.png %}

Firstly, there should be some memory spaces allocated to hold the pre-stored registers, additionally, the page table is created along with creating of `Proc` structure (see [`inner_alloc()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L474)).

Secondly, the address of user process page table should be set into `satp`, and the value of general purpose registers should also be restored.

Finally, put the user space address (virtual address) into the `sepc`, and call `SRET` at the end of the program. After that, everything is changed to user space!



## 2. Trap and Trampoline



## 3. Init Process



## 4. 
