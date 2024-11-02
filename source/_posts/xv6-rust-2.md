---
title: /Xv6 Rust 0x02/ - printf!("Hello xv6-rust!")
date: 2024-11-01 14:36:49
tags:
- xv6
- rust
- os
categories:
- Rust
---

With the help of the previous article, right now we have a good foundation about run rust on risc-v platform.

In the second episode, we are going to jump into some real code of xv6, and take care of the initialize from machine level to supervisor level, and finally, make the `printf!()` macro available in our code!

## 1. Short but important assembly code

In our latest code, we set the entry of our code as `main()`, and after that, we only did one thing before running the test code, which is set the stack pointer.

However, xv6 will do more in the stage of initialization, it chooses to have an individual ASM file to put the very initial code in it, and the ASM file named "[entry.S](https://github.com/LENSHOOD/xv6-rust/blob/master/kernel/src/asm/entry.S)"(some part of code or comment will be truncated, please click the link attached to the file to see the full code):

```assembly
.section .text
.global _entry
_entry:
        la sp, stack0    # load the addr of "stack0" in to sp, "stack0" located in start.rs
        li a0, 1024*4    # load immediate 4096 to a0
        csrr a1, mhartid # read csr, load mhartid to a1 (store the id of current hart, start from 0)
        addi a1, a1, 1   # a1 = a1 + 1
        mul a0, a0, a1   # a0 = a0 * a1
        add sp, sp, a0   # sp = sp + a0
        call start
spin:
        j spin

```

The above code basically set up a 4096-bytes stack, for every hart, and the start address of stack, which we have set to `0x80001000` in our previous code, comes from a constant `stack0` that located in the "[start.rs]((https://github.com/LENSHOOD/xv6-rust/blob/b3a46d46d1b8196b5194eca670f835e476823088/kernel/src/start.rs#L13))".

```rust
### start.rs
... ...
#[repr(C, align(16))]
struct Stack0Aligned([u8; 4096 * NCPU]);
#[no_mangle]
static stack0: Stack0Aligned = Stack0Aligned([0; 4096 * NCPU]);
... ...
```

Define the `stack0` as a `u8` array with length of `4096*NCPU` can safely reserve enough space for kernel stack in each hart, after compiled, the `stack0` will be settled in the `.rodata` section, with its address available in the memory range.

Let's take a look about it (the binary `kernel` is the kernel output of [xv6-rust](https://github.com/LENSHOOD/xv6-rust)):

```shell
$ readelf -s kernel | grep stack0
Num:    Value            Size   Type    Bind    Vis      Ndx  Name
... ...
33873:  0000000080015800 32768  OBJECT  GLOBAL  DEFAULT    2  stack0
... ...

$ readelf -S kernel              
There are 21 section headers, starting at offset 0x2b58f8:

Section Headers:
  [Nr] Name              Type             Address           Offset
       Size              EntSize          Flags  Link  Info  Align
  [ 0]                   NULL             0000000000000000  00000000
       0000000000000000  0000000000000000           0     0     0
  [ 1] .text             PROGBITS         0000000080000000  00001000
       0000000000015000  0000000000000000  AX       0     0     16
  [ 2] .rodata           PROGBITS         0000000080015000  00016000
       000000000000d950  0000000000000000  AM       0     0     16
  [ 3] .eh_frame         PROGBITS         0000000080022950  00023950
       00000000000004b8  0000000000000000   A       0     0     8
  [ 4] .data             PROGBITS         0000000080022e08  00023e08
       000000000002b788  0000000000000000  WA       0     0     8
  [ 5] .bss              NOBITS           000000008004e590  0004f590
       0000000000000660  0000000000000000  WA       0     0     8
  ... ...
```

The above content shows the `stack0` has address `0x80015800`, with `Ndx = 2` means the `stack0` is located at `.rodata` section.

Basically the `entry.S` only responsible for initialized the stack pointer, and then jump to rust code directly.

At last, please don't forget to update the program entry in the `entry.ld`:

```ld
... ...
ENTRY( _entry )
... ...
```



## 2. Machine -> Supervisor

No doubt that the last line of ASM `call start` will bring us to the `start()`, and here is the core part of the [`start()`](https://github.com/LENSHOOD/xv6-rust/blob/b3a46d46d1b8196b5194eca670f835e476823088/kernel/src/start.rs#L16):

```rust
#[no_mangle]
extern "C" fn start() {
    // set M Previous Privilege mode to Supervisor, for mret.
    let mut x = r_mstatus();
    x &= !MSTATUS_MPP_MASK;
    x |= MSTATUS_MPP_S;
    w_mstatus(x);

    // set M Exception Program Counter to main, for mret.
    // requires gcc -mcmodel=medany
    w_mepc(kmain as usize);

    // disable paging for now.
    w_satp(0);

    // delegate all interrupts and exceptions to supervisor mode.
    w_medeleg(0xffff);
    w_mideleg(0xffff);
    w_sie(r_sie() | SIE_SEIE | SIE_STIE | SIE_SSIE);

    // configure Physical Memory Protection to give supervisor mode
    // access to all of physical memory.
    w_pmpaddr0(0x3ffffffffffff);
    w_pmpcfg0(0xf);

    // ask for clock interrupts.
    // Note: here we could safely comment the timer init, because it won't be needed for a period 
    // timerinit();

    // keep each CPU's hartid in its tp register, for cpuid().
    let id = r_mhartid();
    w_tp(id);

    // switch to supervisor mode and jump to main().
    unsafe { asm!("mret") }
}
```

Actually I even cannot even call it a piece of "rust" code, because if you clone the repo and go through the related code, you may found almost all functions here (like `r_mstatus()` or `w_mepc()`) are wrappers to ASM code.

Almost all of the above functions are operate risc-v CSRs (control and status registers), of course we could follow the risc-v specification to learn the details about those CSRs (the entire [Privileged Specification](https://drive.google.com/file/d/17GeetSnT5wW3xNuAHI95-SI1gPGd5sJ_/view) with 166 pages only talks about the CSRs), but I'm gonna post the following table to briefly introduce what they do in the `start()`.

CSRs are a group of registers that can only be accessed in privileged mode, such as machine mode or supervisor mode, those registers are capable of store status, or change the system configurations, and can be read or written by CSR instructions.

| Register               | Name                                          | Description                                                  |
| ---------------------- | --------------------------------------------- | ------------------------------------------------------------ |
| **mstatus**            | Machine Status Register                       | The mstatus register keeps track of and controls the hart’s current operating state.<br />Here we only care about the MPP filed, which stores the previous privileged mode:<br />M = 11; S = 01; U = 00;<br />Back to the code, it sets the previous privileged mode from machine to supervisor. (Note that here the MPP filed only store the mode value, the privileged mode won't be switched immediately) |
| **mepc**               | Machine Exception Program Counter             | When a trap is taken into M-mode, mepc is written with the virtual address of the instruction that was interrupted or that encountered the exception, and it may be explicitly written by software.<br />So why here in the code, the function address of `kmain` been written? It's highly connected with the `mret` instruction, we will get back to this afterward. |
| **stap**               | Supervisor Address Translation and Protection | It controls supervisor-mode address translation and protection. And here we just set it to 0 for disable the virtual address translation. |
| **medeleg / mideleg**  | Machine Trap Delegation Registers             | By default, all traps at any privilege level are handled in machine mode. In our code, both these registers are set as 0xffff to indicate that all traps will be delegated to handle on S-mode |
| **sie**                | Supervisor Interrupt Registers                | In the code, the External / Timer / Software interrupts are all enabled on S-mode. |
| **pmpaddr0 / pmpcfg0** | Physical Memory Protection                    | These two register combined controlling the access permission across a specific address range.<br />Here, xv6 allows RWX permission on S-mode, across the range of `0~0x3ffffffffffff`, that range covers almost 1PiB address space. |
| **tp**                 | Thread Pointer                                | `tp` is one of the general purpose registers, not part of CSR. Obviously the name thread pointer indicates this register is a thread local store register.<br />Then it's easy for us to understand the code: store hart id into `tp` for quicker access. |

The instruction `mret` is highly related to the `mepc` register, like we described in the above table. `mret` is called "Trap-Return Instructions", which is to return from the trap. 

Generally speaking, when any trap like interrupt or exception happens, the instruction address where the trigger the trap, will be stored in the `xPC` register(like `mepc` or `sepc`), then the program will be redirect to a trap handler that related to the specific trap. Once the handler done its work, and the program needs to return to the original location, it will need to fetch the address from `xPC`, and set program counter with that, then jump back to the address.

`mret` (and not surprisingly, there is a `sret` too)  does the whole process by only one instruction, besides, it will also trigger the privileged mode switch, to the mode saved in the MPP filed of `mstatus`. 

So I suppose you have understood the code logic here: at first set the `kmain` to `mepc`, then do some work, at last call `mret` so that the program will jump to the `kmain`, while the privileged mode is switched to S-mode as well.

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



## 3. We need UART

With the `mret` is executed, program is running into a new file: [`main.rs`](https://github.com/LENSHOOD/xv6-rust/blob/b3a46d46d1b8196b5194eca670f835e476823088/kernel/src/main.rs#L97), which is hard to tell if it's new, because we already have one, one not exactly since we will introduce a new function `kmain` to replace our previous `main`.

Don't be frighten by a lot of new functions are called within `kmain`, we are not gonna need them currently, the only one function we should pay our attention on is the `Uart::init()`:

```rust
#[no_mangle]
pub extern "C" fn kmain() {
    if cpuid() == 0 {
        Uart::init();
      ... ...
}
```

