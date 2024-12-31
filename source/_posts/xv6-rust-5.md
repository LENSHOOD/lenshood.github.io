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

While `File` act as a high level abstraction object that exists in most of syscalls related to file systems. Here are some typical syscalls:

```rust
// user/ulib/stubs.rs

// Create a pipe, put read/write file descriptors in p[0] and p[1].
pub fn pipe(fdarray: *const i32) -> i32;

// Write n bytes from buf to file descriptor fd; returns n.
pub fn write(fd: i32, addr: *const u8, n: i32) -> i32;

// Read n bytes into buf; returns number read; or 0 if end of file.
pub fn read(fd: i32, addr: *mut u8, n: i32) -> i32;

// Release open file fd.
pub fn close(fd: i32);

// Load a file and execute it with arguments; only returns if error.
pub fn exec(path: *const u8, argv: *const *const u8) -> i32;

// Open a file; flags indicate read/write; returns an fd (file descriptor).
pub fn open(path: *const u8, omode: u64) -> i32;

// Create a device file.
pub fn mknod(path: *const u8, major: u16, minior: u16) -> i32;

// Return a new file descriptor referring to the same file as fd.
pub fn dup(fd: i32) -> i32;
```

Apparently, some of the syscalls using `fd` as their argument to locate a `File`, but `open`, `exec` and `mknod` using `path` instead of `fd`, since these syscalls are only applicable to files that backed by inodes, and each inode has a `path`.

As an example, the following code piece is taken from the implementation of the syscall `read`:

```rust
pub(crate) fn fileread(f: &mut File, addr: usize, n: i32) -> i32 {
    ... ...
    match f.file_type {
        FD_PIPE => {
            let pipe = unsafe { f.pipe.unwrap().as_mut().unwrap() };
            return pipe.read(addr, n);
        }
        FD_INODE => {
            let ip = unsafe { f.ip.unwrap().as_mut().unwrap() };
            let r = ip.readi(true, addr as *mut u8, f.off, n as usize) as u32;
            ... ...
        }
        FD_DEVICE => {
            if f.major < 0 || f.major >= NDEV || unsafe { DEVSW[f.major as usize].is_none() } {
                return -1;
            }
            unsafe {
                DEVSW[f.major as usize]
                    .unwrap()
                    .as_mut()
                    .unwrap()
                    .read(true, addr, n as usize)
            }
        }
        ... ...
    }
}
```

The structure is quite clear. Based on the three different types of file, it reads data from a pipe, an inode or a device respectively.



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

### 2.1 INode

We have been talking about the inode even we don't know what exactly the inode is, that because the inode is a really important concept in the file system and we can barely avoid it. Now let's deep dive into the "inode layer" to find out why it's so important.

First, we should have a look at the `INode` structure:

```rust
// file.rs
// in-memory copy of an inode
#[derive(Copy, Clone)]
pub struct INode {
    pub(crate) dev: u32,        // Device number
    pub(crate) inum: u32,       // Inode number
    pub(crate) ref_cnt: i32,    // Reference count
    pub(crate) lock: Sleeplock, // protects everything below here
    pub(crate) valid: bool,     // inode has been read from disk?

    pub(crate) file_type: FileType, // copy of disk inode
    pub(crate) major: i16,
    pub(crate) minor: i16,
    pub(crate) nlink: i16,
    pub(crate) size: u32,
    pub(crate) addrs: [u32; NDIRECT + 1],
}

// fs.rs
// On-disk inode structure
#[repr(C)]
#[derive(Copy, Clone)]
pub struct DINode {
    pub file_type: FileType,       // File type
    pub major: i16,                // Major device number (T_DEVICE only)
    pub minor: i16,                // Minor device number (T_DEVICE only)
    pub nlink: i16,                // Number of links to inode in file system
    pub size: u32,                 // Size of file (bytes)
    pub addrs: [u32; NDIRECT + 1], // Data block addresses
}

// stat.rs
pub enum FileType {
    NO_TYPE,
    T_DIR,    // Directory
    T_FILE,   // File
    T_DEVICE, // Device
}
```

Surprise! There are two inode structures actually, and that makes total sense. Because the inode would not only exist in memory, but also persistent on disk.

Comparing the two forms of inodes, the memory version of `INode` has a few more fields than the disk version, such as `lock` and `ref_cnt` for concurrently accessing, and `inum` for identification.

Besides, the disk version of `DINode` is very compact in its size, that's good for store in block and easy to calculate the location in the disk. Let's calculate the real size of one `DINode` by its fields:

- file_type: 1 byte. Rust allocate enumeration's size dynamically [based on the number of elements](https://rust-lang.github.io/unsafe-code-guidelines/layout/enums.html) defined in the `enum`, since the `FileType` has only 4 elements, 1 byte of space is enough for them.
- major: `i16` == 2 bytes
- minor: `i16` == 2 bytes
- nlink: `i16` == 2 bytes
- size: `u32` == 4 bytes
- addrs: `NDIRECT` == 12, so the size of addrs is 4 * (12 + 1) = 52 bytes

Therefore, if we simply add them together, then one `DINode` would take 63 bytes of space on disk. However, like other programs languages, values in rust also have alignment beside size. In [the rust representation](https://doc.rust-lang.org/reference/type-layout.html#representations), the fields in a struct can be reordered, and aligned, so we cannot confidently say the size of `DINode` is 63 bytes.

Additionally, you may noticed that the [`#[repr(C)]`](https://doc.rust-lang.org/reference/type-layout.html#reprc-structs) has been put on the `DINode`, which makes sure the compiler keeps the memory layout of the type exactly the same as the type defined in C language. According to the size and alignment principle of `#[repr(C)]`, we can have the following image to help us calculate the real size:

{% asset_img 3.png %}

And based on the above image, the real size of `DINode` is 64 bytes.

So why is it so important to calculate the real size of `DINode`? Because next we'll talk about how the inodes stay in the disk.

We have known that there is a section in the disk which takes 4 blocks to hold the inode, but only inode metadata, the data of inode stores in another location. Refer to the `DINode`, expect for the last field `addrs`, all other fields can be seen as metadata of a inode, the last field `addrs` stores the reference of the data blocks, which is block number.

Hence, I guess you already figured out how the metadata and real data of inode store in the disk: the inode section stores `DINode`, while the data section stores the data blocks, which numbers are stored in the `DINode` as references.

So far, I suppose we can do a simple math to prove why we should 4 blocks for the inode section:

First, `NINODES` in the super block defines the max number of inodes the file system can have, the value by default is 200. Then we already know one `DINode` takes 64 bytes of space, so 200 inodes will need 12800 bytes to be saved. A block can store 4096 bytes data, since `12800 / 4096 = 3.125`, which means the inode section will need at least 4 blocks.

The following images shows an inode and the placement of its metadata and real data:

{% asset_img 4.png %}

In the above example, the `DINode` has 10354 bytes of data, and that will need 3 blocks to store. But what if the size of data is too large to be stored within 13 blocks? Actually, in the `addrs[]`, only first 12 blocks are using for store data, the number 13th block is specifically designed for more data. If the first 12 blocks are full, but there is still some data needs to be stored, at that time, block 13 will become a big array dedicated to store block references.

In the `addrs[]`, the first 12 blocks are called direct data blocks, which can be referenced from `addrs[]` directly, and the data blocks that exceeds 12, are called indirect data blocks, all of the block number of indirect blocks are stored in the 13th block.

Let's see if the data size exceeds 12 blocks:

{% asset_img 5.png %}

So this time we assume the size is 1639184 bytes, which will need 401 blocks, with the indirect blocks, they can be easily stored. Since the size of a block is 4096 bytes, and a `u32` block number takes 4 bytes to store, in theory, the maximum size of a single file in xv6 would be 4MiB.

### 2.2 Pathname & Directory

With the knowledge of inode, now we can have a look at the "directory". In the file system, a directory is also considered an inode file, no matter the path of itself, or the paths of its sub directories or sub files, are all treated as plain data.

This is the definition of directory entry:

```rust
// fs/mod.rs
pub const DIRSIZ: usize = 14;

pub struct Dirent {
    pub inum: u16,    // inode number
    pub name: [u8; DIRSIZ],
}
```

Directory is a file containing a sequence of `Dirent` structures. And the interesting thing is that a directory file only contains the paths of its subdirectories and subfiles, its own path, on the other hand, is contained in its parent's file.

The following image shows an example of ``/home/lenshood`:



So far we've seen the INode related structures and designs, which help us understand how INode level works. Thus, we are going to omit the INode operations, please check them at the `fs.rs` to see more details.

### 2.3 Log

Under the INode level, there is another important abstraction level called "log". Generally speaking, log level doesn't responsible for any complex logic, it only care about how to reduce the risk of disk write interruption if a crash occurs. Basically, the log level is building on the concept of ["redo log"](https://dev.mysql.com/doc/refman/8.2/en/innodb-redo-log.html#:~:text=The%20redo%20log%20is%20a,or%20low%2Dlevel%20API%20calls.). For the system that running in the real world, we must assume crash will definitely happen in someday. 

So if there is no mechanism to take care of the crash, then it has possibility that some disk write operations are on going while a power off crash occurs, lead to some blocks were written but some other blocks were not. Since we don't have any idea about what kind of data was in the block, the crash may cause a block has been assigned to an INode (meta data saved), but the assigned flag failed to save in the block. That's a quite serious problem because the kernel may reassign the block that it believes is a free block, to another INode.

But don't give me wrong, the key thing that the log level should keep, is not "recovery", is "consistency", because data inconsistent is way more serious than data loss. The problem we just described above can turn the system become a totally disaster because no one can simply tell which INode the block should belong to, comparing that, a data loss is more controllable as we should only revise the data in a few minutes before crash. There's another reason that we just care about consistency not recovery, that is, keep no data loss is too costly, the performance should also be considered otherwise nobody would use it.

To learn how log level works, let's look deeper into the log blocks:

{% asset_img 7.png %}

We've known there are 30 blocks reserved for log, but only last 29 blocks are using for save data, the first log block is for log header:

```rust
struct LogHeader {
    n: u32,
    block: [u32; LOGSIZE],
}
```

LogHeader is very simple, `n` represents how many blocks are saved in log currently, and `block` records the block number where the log blocks map to the real blocks.

When a fs syscall is executing to write data into disk, it first call `begin_op()` indicating a disk write transaction has began. Next, rather than directly call block write method, it just call `log_write()` then return, actually the writing operation is not executed at that time, the log related code will make sure the data is eventually written. After all things done, the syscall will call `end_op()` to finish the transaction.

The following diagram may describe the data write process clearer:

{% asset_img 8.png %}

The `log_write()` will only record the block number of the block that has updates. BTW, if a block was updated, before it goes to disk, there is a block cache to hold the block data temporary in the memory, we'll talk about the block cache later.

When the `end_op()` is called, and if there is no other log write operation on going, the data will be written. First the data will be  written into log blocks, then writes the log header. After that, commit has been successfully executed, and if crash occurs at this point, all of the data saved in log can be recovered, but before that, all data will lost, even if the log blocks are already written. This is how the log mechanism keeps the data writing atomically and consistency. 

After log header is saved, the log blocks will be moved to real blocks, this process is exactly the same as the recovery process. When system started, it will check the log header first, if any log blocks are found, the recovery process will be executed to recover data into real blocks.



## 3. Block Operation



## 4. VirtIO Device

