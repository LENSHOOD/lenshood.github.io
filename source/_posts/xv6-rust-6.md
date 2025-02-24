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
- How to deal with the different memory space between two modes?

## 2. Trap and Trampoline



## 3. Init Process



## 4. 
