---
title: Easy piece 1: Virtualization - virtual memory
date: 2024-11-08 23:29:43
tags:
- xv6
- rust
- os
categories:
- Rust
---

We have learnt how to setup risc-v in rust, and also initialized risc-v to be able to print format strings, in this chapter we are taking the first look of the OS kernel, and will try to figure out the memory management in xv6.

<!-- more -->

## 1. Overview of physical memory

Before we get any further, let's recall the memory mapping in the QEMU that we mentioned in previous chapters:

```shell
$ qemu-system-riscv64 -monitor stdio
QEMU 9.1.0 monitor - type 'help' for more information
(qemu) info mtree
address-space: cpu-memory-0
address-space: memory
  0000000000000000-ffffffffffffffff (prio 0, i/o): system
    0000000000001000-000000000000ffff (prio 0, rom): riscv.spike.mrom
    0000000001000000-000000000100000f (prio 1, i/o): riscv.htif.uart
    0000000002000000-0000000002003fff (prio 0, i/o): riscv.aclint.swi
    0000000002004000-000000000200bfff (prio 0, i/o): riscv.aclint.mtimer
    0000000080000000-0000000087ffffff (prio 0, ram): riscv.spike.ram

address-space: I/O
  0000000000000000-000000000000ffff (prio 0, i/o): io
```

As we can see, the range of RAM is from `0x80000000` to `0x87ffffff`, which exactly to be 128MiB, and that because we set it to be 128MiB in the runner command.

And let's take a look to the [`kernel.ld`](https://github.com/LENSHOOD/xv6-rust/blob/master/kernel/src/ld/kernel.ld) in the xv6-rust repo:

```ld
OUTPUT_ARCH( "riscv" )
ENTRY( _entry )

SECTIONS
{
  /*
   * ensure that entry.S / _entry is at 0x80000000,
   * where qemu's -kernel jumps.
   */
  . = 0x80000000;

  .text : {
    /* no idea why the previous form *(.text .text.*) not working, maybe is relate to ld version */
    *(.text) *(.text.*)
    . = ALIGN(0x1000);
    _trampoline = .;
    *(trampsec)
    . = ALIGN(0x1000);
    ASSERT(. - _trampoline == 0x1000, "error: trampoline larger than one page");
    PROVIDE(etext = .);
  }

  .rodata : {
    . = ALIGN(16);
    *(.srodata .srodata.*) /* do not need to distinguish this from .rodata */
    . = ALIGN(16);
    *(.rodata .rodata.*)
  }

  .data : {
    . = ALIGN(16);
    *(.sdata .sdata.*) /* do not need to distinguish this from .data */
    . = ALIGN(16);
    *(.data .data.*)
  }

  .bss : {
    . = ALIGN(16);
    *(.sbss .sbss.*) /* do not need to distinguish this from .bss */
    . = ALIGN(16);
    *(.bss .bss.*)
  }

  PROVIDE(end = .);
}
```

It's a little bit more complicated than the one we wrote in the previous chapters, but still quite clear.

Let's ignore the complicated part in the middle of the file, only focus on the `PROVIDE(***)` lines:

```ld
/* define a symbol named as "etext", and set it value 
 * to the current location address. At here, means the
 * address where end of text section.
 */
PROVIDE(etext = .);

/* define a symbol named as "end", and set it value
 * to the address where all kernel data ends.
 */
PROVIDE(end = .);
```

We can imagine, before `etext`, the program lies there and cannot be changed, between `etext` and `end`, any read-only data, writable data and bss data are put there, after `end`, the rest of the RAM space is available for stack or heap. Here are the real symbol addresses:

```shell
$ readelf -s kernel | grep -E "etext|stack0|end"
 33873: 00000000800162c0 32768 OBJECT  GLOBAL DEFAULT    2 stack0
 33892: 000000008004eb30     0 NOTYPE  GLOBAL DEFAULT    5 end
 33893: 0000000080015000     0 NOTYPE  GLOBAL DEFAULT    1 etext
```

Hence, we could get the simple RAM map as follows:

{% asset_img 1.png %}

And the interesting thing is, if you take one more step to check the section addresses by `readelf`, you'll find the `stack0` was actually located in the `rodata` section, because we define it as a `static` flied in the [`start.rs`](https://github.com/LENSHOOD/xv6-rust/blob/9cd275a5591956c8c16103acf177c057e485c600/kernel/src/start.rs#L13), but we use the `stack0` as the kernel stack, which is writable. 

The reason why we can write a read-only flied without an exception, is at the beginning of `start()` we access it in the machine mode, and afterward we set the Physical Memory Protection to RWX across the range of `0~0x3ffffffffffff`, according to the last chapter, so that in the supervisor mode, the `stack0` can also be written.



## 2. Memory allocator

So far we have known the physical RAM space, next let's see how to manage the RAM space so that it can easily be allocated and returned back.

In xv6, the smallest unit of memory management is "page", which is 4096B by default, and we can adjust the page size by setting the [`PGSIZE`](https://github.com/LENSHOOD/xv6-rust/blob/9cd275a5591956c8c16103acf177c057e485c600/kernel/src/riscv.rs#L269). Basically RAM management is divide the RAM into numerous and consistent pages, like this:

{% asset_img 2.png %}

So now we have divided the RAM into 32689 pages, how to track them? How to allocate one or many pages to some code that would need such amount of pages? And how to recycle them in the end?

Xv6 using the implicit list to track them, and each list item called a `Run`:

```rust
// kalloc.rs
struct Run {
    next: *mut Run,
}
```

It's very simple to be understood, the list actually only track the free pages, so the `next` pointer can directly put into the page space itself, avoid to introduce any extra memory space to save the pointer. Once a page been allocated, the whole 4096B can be used to store data, when it return back, we can find the correct position of that page by calculating its address.

{% asset_img 3.png %}

Based on the above image, we can better realize how xv6 manage its RAM. During the kernel starting, the available RAM space will be initialized by divided in to pages, and fill with junk data:

```rust
// kalloc.rs
pub fn kinit() {
  ... ...
  // Range: From "end" to "PHYSTOP"
  KMEM.freerange((&mut end) as *mut u8, PHYSTOP as *mut u8);
  ... ...
}

fn freerange<T: Sized>(self: &mut Self, pa_start: *mut T, pa_end: *mut T) {
  let mut p = PGROUNDUP!(pa_start);
  while p + PGSIZE <= pa_end as usize {
    self.kfree(p as *mut T);
    p += PGSIZE;
  }
}

pub fn kfree<T: Sized>(self: &mut Self, pa: *mut T) {
  ... ...
  // Fill with junk to catch dangling refs.
  memset(pa as *mut u8, 1, PGSIZE);
  // Build "Run" in each pages
  let r = pa as *mut Run;
  (*r).next = self.freelist;
  // "freelist" is the head of the implicit list
  self.freelist = r;
  ... ...
}
```

When the initialize completed, then any kernel code could call `kalloc()` to get one page of available memory:

```rust
// kalloc.rs
pub fn kalloc<T: Sized>(self: &mut Self) -> *mut T {
  ... ...
  // Get the latest free page "r" to be allocated
  let r = self.freelist;
  if !r.is_null() {
    unsafe {
      // Delete the "r" from free list
      self.freelist = (*r).next;
    }
  }
  ... ...
  r as *mut T
}
```

Please note that, til then all of the addresses and pages we have talked about are physical memory, actually before a page been used, its address must be converted to the virtual address. Next we'll go to the virtual memory part.



## 3. Virtual Memory

Virtual memory mechanism is the key to achieve the memory virtualization, modern processes usually support virtual memory in hardware level, which includes address translation, permission control and related interrupts.

In risc-v architecture, virtual memory management is based on page, variety of page table structure modes are supported in risc-v, like Sv32, Sv39 and Sv48, even Sv52, the number behind the "Sv" indicates the address width, for example, "Sv39" is the short for "Supervisor Virtual addressing with 39-bit virtual addresses", which means under the Sv39 mode, the virtual address space that a process can accessing is $2^{39}$ bytes that around 512GiB.

Let's recap the `satp` CSR mentioned in the last chapter, and go one step deeper(more details please see *[The RISC-V Instruction Set Manual: Volume II: 10.1.11.](https://drive.google.com/file/d/17GeetSnT5wW3xNuAHI95-SI1gPGd5sJ_/view)*):

{% asset_img 4.png %}

Above image is the definition of `satp`, it contains three parts, the first part defines the mode, like the following image:

{% asset_img 5.png %}

Xv6 uses risc-v64 and Sv39, so the highest 4bit would be `0x8`.

The second part "ASID" basically provide better performance by allow OS to hint TLB through it, however, xv6 doesn't using such feature.

The third part "PPN" is called physical page number, the `stap.PPN` should be set to the PPN of the root page table, so that the processor can find page table through it.

Talking about page table, what exactly the page table structure being defined in Sv39? Let's have a look, in the meantime, there are much more details in *[The RISC-V Instruction Set Manual: Volume II: 10.3 ~ 10.6.](https://drive.google.com/file/d/17GeetSnT5wW3xNuAHI95-SI1gPGd5sJ_/view)*, please check it if you are interested.

In fact, not only risc-v, almost all page tables share the concept of multi-level page table entries. We can have an example to make it clearer for understanding:

Say we have a variable `a` located in a process's stack, which virtual address is `0x3f_fff7_e000`(you'll know why we choose such a bug number afterward in the following sections). And we assume it's physical address is `0x80050000`, so if we need a page table right now to hold such mapping relationship, the simplest way is having an array to acts as a table, like this:

 {% asset_img 6.png %}

If we don't yet care how we got such a huge array, which size is exactly as same as the 39-bit space, the index of the array represents virtual address, while the value represents physical address. This at least makes our "theoretically" page table available.

But it's no doubt that this won't work. Considering 99.99% of the above page table are not been used, but the space still need to be occupied since array is consistent. Can we use more space efficient way like linked list? Not really, because address translation is a very high frequency operation which requires very low latency.

Then multi-level page table turns out to be the final solution:

 {% asset_img 7.png %}

With the multi-level approach, only a few of page table will be created, like the above image, if we want to store the mapping `0x3f_fff7_e000 -> 0x80050000`, we only need for 512-entries page table, assume an entry store a 64-bit number(the first three store the next page table address, the last one store the physical address), then they only take $512 * 8 * 4 = 16\ KiB$ space in total.

More importantly, the virtual memory structure manages memory by page, one page in Sv39 is 4096-bit, which is $2^{12}$, therefore, actually we don't need the low 12 bits at all, since the last 12 bits in each page is 12 bits 0. risc-v make use of this 12-bit to record permissions.  The following image shows format of virtual address, physical address and PTE(page table entry) in Sv39:

 {% asset_img 8.png %}

The PTE flags `R`, `W`, `X` indicate the read, write and execute permission of the PTE, U indicates user mode, V use to identify if the PTE is valid.



## 4. Initialize kernel virtual memory





