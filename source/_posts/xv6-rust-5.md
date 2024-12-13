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
// file.rs
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

The other two layers, "pathname" and "directory" represent the way to access and manage file. Path name provides a hierarchical  way to locate a specific file, like `/home/lenshood/video.mkv`, while the directory is dedicated for a special type of file that holds all sub-directory files as its data, a directory is like `/home/lenshood/`.



## 2. File Organization

Because the file system on the disk is quite important and complicated, now we are focusing on the file and dir that organized by the inode system. And we'll leave the device and pipe part at the end.

As mentioned before, inode is an object that holds the data of file or directory, and is placed on the disk. So basically all of the files and directories are stored as form of inode on the disk.

The problem is, we see the disk as a very big array (like memory but commonly slower and bigger), so how to effectively put the file and directory on this array, organized as the path name hierarchy, at the same time, easy to create, update and delete?

Generally, most of the file systems are implemented by splitting the entire disk as multiple sections, some section records the metadata, some sections store file, and some sections just stays empty for future use. The file system in the xv6 also designed like that, but much simpler.

The following figure describes the disk section structure for the xv6:

{% asset_img 2.png %}

We can see from the above figure that the entire disk has been divided as numerous same sized blocks, which size is 4096 bytes per block. That gives the file system the minimum control unit, that is one block. And each block has its own number, starts from 0.

Based on block, a disk image can be divided into these six sections:

- Boot Section(block 0): size of 1 block, no particular purpose, it stays empty since the xv6 defines meaningful block starts from No.1
- Super Section(block 1): size of 1 block, it contains the metadata of the file system, including magic number, the size of the entire image and the size of other sections, we'll check it soon after.
- Log Section(block 2~31): size of 30 blocks, it mainly records the disk operation logs for recovery purpose, we'll check it soon after
- INode Section(block 32~35): size of 4 blocks, the inode was designed to store its metadata and real data separately, these 4 blocks only store the metadata parts of all inodes.
- Bitmap Section(block 36): size of 1 block, as we know the bitmap is a very efficiency data structure to indicate huge statuses by a few space, here is the same, bitmap section using 1 block to indicate the occupied blocks and empty blocks in the entire disk
- Data Section(block 37~2000): size of 1963 blocks, the real files data store in this section, which you can notice it takes the most blocks of the disk. The end block 2000 was defined by a constant, and can be changed, but more data blocks also means more inodes, so extend the size of data section may need to extend inode section as well.

Firstly, let's take a look at the super block:

```rust
// fs.rs
pub struct SuperBlock {
    pub(crate) magic: u32,      // Must be FSMAGIC
    pub(crate) size: u32,       // Size of file system image (blocks)
    pub(crate) nblocks: u32,    // Number of data blocks
    pub(crate) ninodes: u32,    // Number of inodes.
    pub(crate) nlog: u32,       // Number of log blocks
    pub(crate) logstart: u32,   // Block number of first log block
    pub(crate) inodestart: u32, // Block number of first inode block
    pub(crate) bmapstart: u32,  // Block number of first free map block
}

// mkfs/main.rs
const SB: SuperBlock = SuperBlock {
    magic: FSMAGIC,
    size: (FSSIZE as u32).to_le(),
    nblocks: NBLOCKS.to_le(),
    ninodes: NINODES.to_le(),
    nlog: NLOG.to_le(),
    logstart: 2u32.to_le(),
    inodestart: (2 + NLOG).to_le(),
    bmapstart: (2 + NLOG + NINODEBLOCKS).to_le(),
};
```

As a place to hold the metadata of the file system, `SuperBlock` holds very essential data, if we see the initialize of `SuperBlock`  in the mkfs, there are `FSMAGIC`, `FSSIZE`, `NBLOCKS`, `NINODES`, `NLOG` and other constants are assigned, these constants define the block size of several sections.





## 3. Block Operation



## 4. VirtIO Device

