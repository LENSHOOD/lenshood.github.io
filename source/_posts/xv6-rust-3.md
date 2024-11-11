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



## 2. Memory allocator

