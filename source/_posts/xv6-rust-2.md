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

{% asset_img header.jpg 500 %}

With the foundation established in the previous article, we now have a working Rust environment on RISC-V.

In this second installment, we'll examine the actual xv6 code, handling the initialization from machine level to supervisor level, and ultimately implementing the `printf!()` macro.

<!-- more -->

## 1. Short but important assembly code

In our current implementation, we designated `main()` as the entry point, performing just one crucial operation before executing test code: initializing the stack pointer.

However, xv6 performs additional initialization steps by placing the earliest boot code in a dedicated assembly file named "[entry.S](https://github.com/LENSHOOD/xv6-rust/blob/master/kernel/src/asm/entry.S)" (some code/comment truncation may occur; see the linked file for complete content):

```assembly
### entry.S

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
// start.rs
... ...
#[repr(C, align(16))]
struct Stack0Aligned([u8; 4096 * NCPU]);
#[no_mangle]
static stack0: Stack0Aligned = Stack0Aligned([0; 4096 * NCPU]);
... ...
```

The `stack0` is defined as a `u8` array sized `4096*NCPU`, ensuring sufficient kernel stack space for each hardware thread. During compilation, this array is placed in the `.rodata` section with its address mapped within the accessible memory range.

We can verify this memory layout by examining the `kernel` binary (compiled output of [xv6-rust](https://github.com/LENSHOOD/xv6-rust)):

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

Finally, please don't forget to update the program entry in the `entry.ld`, as well as the text section:

```ld
... ...
ENTRY( _entry )
... ...
/* As we declared ".section .text" in the beginning of entry.S, 
 * we shoud put the "*(.text)" here with the first order.
 * This will ensure the entry.S is at the beginning of the binary.
 */
.text : {
  *(.text) *(.text*)
}
... ...
```

And declare the `entry.S` in `main.rs`:

```rust
// main.rs
... ...
global_asm!(include_str!("entry.S"));
... ...
```



## 2. Machine -> Supervisor

No doubt that the last line of ASM `call start` will bring us to the `start()`, and here is the core part of the [`start()`](https://github.com/LENSHOOD/xv6-rust/blob/b3a46d46d1b8196b5194eca670f835e476823088/kernel/src/start.rs#L16):

```rust
// start.rs

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

Technically, this barely qualifies as Rust code since most functions (like `r_mstatus()` or `w_mepc()`) are thin wrappers around inline assembly instructions, as seen in the repository's implementation.

These functions primarily interact with RISC-V CSRs (control and status registers). While the complete [Privileged Specification](https://drive.google.com/file/d/17GeetSnT5wW3xNuAHI95-SI1gPGd5sJ_/view) (166 pages) covers CSRs in detail, the following table summarizes their roles in the `start()` function:

CSRs (Control and Status Registers) are privileged-mode registers (machine/supervisor) that:<br>- Store system state information<br>- Control hardware configurations<br>- Are accessed via specialized CSR instructions

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

With the `mret` is executed, the program is running into a new file: [`main.rs`](https://github.com/LENSHOOD/xv6-rust/blob/b3a46d46d1b8196b5194eca670f835e476823088/kernel/src/main.rs#L97), which is hard to tell if it's new, because we already have one, one not exactly since we will introduce a new function `kmain` to replace our previous `main`.

Don't be frightened by a lot of new functions that are called within `kmain`, we are not gonna need them currently, the only functions we should pay our attention to are the `Uart::init()` and `Console::init()`:

```rust
// main.rs

#[no_mangle]
pub extern "C" fn kmain() {
    // The cpuid() returns the hart id, according to the last article 
    // we run qemu with only one cpu, so we could just comment it
    // if cpuid() == 0 {
        Uart::init();
        Console::init();
      ... ...
}
```

QEMU generic [virtual platform](https://www.qemu.org/docs/master/system/riscv/virt.html) for risc-v supports a "NS16550 compatible UART". According to the memory address mapping we talked about in the last chapter:

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

UART address starts from `0x1000000`. And there are about 10 registers to config and control the UART (for more details refer to the [16550 specification](http://byterunner.com/16550.html)).

Let's go back to code. We can find all UART related code in the file [`uart.rs`](https://github.com/LENSHOOD/xv6-rust/blob/master/kernel/src/uart.rs). And basically `Uart::init()` initializes the UART in the mode of 8 bits + 38.4k baud rate + FIFO with interrupt.

In fact, after initialize, we could directly put or get chars by the following code:

```rust
// uart.rs
// Please note that there are many more functions and lines of code in the original uart.rs, 
// but we won't need those right now, only the code here is necessary.
pub(crate) static mut UART_INSTANCE: Uart = Uart::create();

pub fn putc_sync(self: &mut Self, c: u8) {  
  // wait for Transmit Holding Empty to be set in LSR.
  while (ReadReg!(LSR) & LSR_TX_IDLE) == 0 {}
  WriteReg!(THR, c);
}

fn getc(self: &Self) -> i8 {
  return if ReadReg!(LSR) & 0x01 != 0 {
    // input data is ready.
    ReadReg!(RHR) as i8
  } else {
    -1
  };
}
```

Let's have a quick test to print a "A" to the console:

```rust
// main.rs

#[no_mangle]
extern "C" fn kmain() {
  Uart::init();
  unsafe { UART_INSTANCE.putc_sync('A' as u8) }
}
```

```shell
... ...
     Finished dev [unoptimized + debuginfo] target(s) in 0.07s
     Running `qemu-system-riscv64 -s -S -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel target/riscv64gc-unknown-none-elf/debug/xv6-rust-sample`
A
```

Awesome, we have printed the first letter! Since we can print a letter, the `printf!()` is around the corner.



## 4. printf!()

At last, we got here. So far we already output a letter "A" through UART, the next we simply need to create a printer and call UART inside to print.

Generally speaking, the only difference between UART with a printer is that the printer takes a format string rather than a character, which means the printer is on a higher abstraction level, and needs to conduct the preprocess of format string, to parse the format string to a standard string, and then crack down the string to characters. 

Refer to the [`print.rs`](https://github.com/LENSHOOD/xv6-rust/blob/b3a46d46d1b8196b5194eca670f835e476823088/kernel/src/printf.rs#L12), the macro `printf!()` receives the input arguments as the "format_args":

```rust
// print.rs

#[macro_export]
macro_rules! printf
{
        ($($arg:tt)*) => {
        unsafe {
            crate::printf::PRINTER.printf(core::format_args!($($arg)*))
        }
    };
}

pub fn printf(self: &mut Self, args: Arguments<'_>) {
  // Like before, we won't need any locks here, the logic with locks could be commentted safely
  let _ = unsafe { CONSOLE_INSTANCE.write_fmt(args).unwrap() };
}
```

"format_args" allow us to print a string with params, such as `printf!("This is a {}", "param")`.

The best part here is we don't need to do anything by ourselves to parse the relatively complex arguments: `"This is a {}"` and `"param"`.  There is a rust trait `core::fmt::Write` takes care of all that stuff!

Let's go to the [`console.rs`](https://github.com/LENSHOOD/xv6-rust/blob/b3a46d46d1b8196b5194eca670f835e476823088/kernel/src/console.rs#L107):

```rust
// console.rs
pub(crate) static mut CONSOLE_INSTANCE: Console = Console::create();

impl Write for Console {
    // The trait Write expects us to write the function write_str
    // which looks like:
    fn write_str(&mut self, s: &str) -> Result<(), Error> {
        for c in s.bytes() {
            self.putc(c as u16);
        }
        // Return that we succeeded.
        Ok(())
    }
}

pub fn putc(self: &mut Self, c: u16) {
    unsafe {
      if c == BACKSPACE {
        // if the user typed backspace, overwrite with a space.
        UART_INSTANCE.putc_sync(0x08); // ascii \b char
        UART_INSTANCE.putc_sync(0x20); // ascii space char
        UART_INSTANCE.putc_sync(0x08); // ascii \b char
      } else {
        UART_INSTANCE.putc_sync(c as u8);
      }
    }
}
```

The `Write` trait implemented the function `write_fmt` by default, we only need to implement the `write_str` here and output the string that has already been parsed correctly. The string can be outputted by calling the UART function `putc_sync`.

Finally, we could print something with `printf!()`!

```rust
// main.rs

#[no_mangle]
extern "C" fn kmain() {
  Uart::init();
  Console::init();
  printf!("\nHello xv6-rust!\n");
}
```

And the output:

```shell
... ...
     Finished dev [unoptimized + debuginfo] target(s) in 1.60s
     Running `qemu-system-riscv64 -s -S -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel target/riscv64gc-unknown-none-elf/debug/xv6-rust-sample`

Hello xv6-rust!

```

It works!
