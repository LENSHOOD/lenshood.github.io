---
title: /Xv6 Rust 0x05/ - Persistence
date: 2024-12-10 22:25:29
tags:
- xv6
- rust
- os
categories:
- Rust
---

With all of the content in the previous chapters, we have known how a process to be initialized and run, but before the kernel runs the first line of a process's code, a question still remained, how does my user program stored in the disk and to be loaded?

In this chapter, we are going to discover the disk management and file system in the xv6, we'll see the persistence stack from the top to the bottom, let's get started!

<!-- more -->

## 1. Overview of Persistence

The xv6 designs a very clear persistence system, which is designed by layers. The [xv6-book](https://pdos.csail.mit.edu/6.828/2024/xv6/book-riscv-rev4.pdf) demonstrates a diagram(in chapter 8) to show such layers as follows:

{% asset_img 1.png %}

Actually, the first three layers are all related to the concept of "File". The xv6 inherits the unix philosophy of ["everything is a file"](https://en.wikipedia.org/wiki/Everything_is_a_file), that is a high level abstraction that regard all devices, pipes, directories, and disk files as "Files", to simplify access and management. Every file has a unique file descriptor to identify it conveniently, file descriptor is also the name of first layer.

Let's have a first look at the structure `File`:

```rust
pub struct File {
    pub(crate) file_type: FDType,
    ref_cnt: i32, // reference count
    pub(crate) readable: bool,
    pub(crate) writable: bool,
    pub(crate) pipe: Option<*mut Pipe>, // FD_PIPE
    pub(crate) ip: Option<*mut INode>,  // FD_INODE and FD_DEVICE
    pub(crate) off: u32,                // FD_INODE
    pub(crate) major: i16,              // FD_DEVICE
}

pub(crate) enum FDType {
    FD_NONE,
    FD_PIPE,
    FD_INODE,
    FD_DEVICE,
}
```

The FDType stands for "file descriptor type", apparently there are three different types of file descriptors except for the "NONE". The "PIPE" indicates a communication channel between two processes, and a pipe is completely lives in memory; the "DEVICE" indicates some kind of devices like the UART console, which can be operated by read and write data; the "INODE" indicates a kind of object that holds some data blocks in the disk.

Looking into the `File` structure, it holds some fields that represent different type of files. For example, both the fields `pipe` and `ip` are "Option" type, which means they may exist on a file or not. Say a pipe file, it would only need the `pipe` field, all other fields corresponding to inode or device can be empty.

The other two layers, "pathname" and "directory" represent the way to access and manage file. Path name provides a hierarchical  way to locate a specific file, like "/home/lenshood/video.mkv", while the directory is dedicated for a special type of file that holds all sub-directory files as its data, a directory is like "/home/lenshood/".



## 2. File Organization



## 3. Block Operation



## 4. VirtIO Device

