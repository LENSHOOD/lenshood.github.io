---
title: /Xv6 Rust 0x01/ - Getting Started is the Hardest Part
date: 2024-10-23 22:27:59
tags:
- xv6
- rust
- os
categories:
- Rust
---

{% asset_img header.jpg 500 %}

[Xv6](https://github.com/mit-pdos/xv6-public) is one of the best operating systems for teaching. It’s a great way to learn about how an OS works with basic functions and few lines of code. 

Originally, xv6 was written in C, which is excellent for students to get hands-on experience with such a classic programming language. But now that Rust is gaining traction—especially since rust-for-linux is becoming a part of the main line Linux—wouldn’t it be fun to run xv6 using Rust? 

As an engaging side project, I have migrated most of xv6 from C to Rust. You can check it out [here](https://github.com/LENSHOOD/xv6-rust). During this migration process, I encountered numerous challenges and complex issues, but nothing was more rewarding than successfully resolving them!

Therefore, I'm sharing my experiences through this series of articles that detail the migration process, with a structured approach and clear explanations.

Let's begin our exploration of the migration process.

<!-- more -->

## 1. Rust on risc-v

In some previous versions of the xv6, it running on x86 arch, however, currently the xv6 has fully migrated to the risc-v arch. 

As we are going to port the xv6 to rust, at the very first, we better take a look about how to run rust on risc-v.

> Please note that in these series of articles, I will assume the reader has basic knowledge about rust, and knows how to setup the local environment such as rustup, cargo, and IDE.
>
> In the following articles, I will use my local machine as the demo env, and that includes:
>
> - MBP 2019, Intel
> - cargo 1.75.0-nightly (b4d18d4bd 2023-10-31), some features rely on nightly build
> - rustup 1.26.0 (5af9b9484 2023-04-05)
> - CLion 2024.1, (I didn't choose RustRover, because after giving it a try I found it still not stable)

Here we go! Let's create a new rust project, and name it "xv6-rust-sample".

Now, there is one and only one `main.rs` file residing in the `src/` directory. We'll leave it for now as we don't need that file immediately.

Since we're targeting RISC-V architecture, we must first configure the appropriate toolchain - our code won't run correctly if compiled for x86 architecture.

We'll configure the toolchain by creating a `.cargo/config.toml` file in the project root, where we can specify our build target.

Add these two lines to the configuration file:

``` toml
[build]
target = "riscv64gc-unknown-none-elf"
```

This configuration specifies that we want to compile our project for RISC-V architecture.

With this configuration in place, we can proceed to the next step.

Returning to `main.rs`, we find a simple initial function:

```rust
fn main() {
    println!("Hello, world!");
}
```

However, if you type and execute `cargo run` with confidence, then you would probably get this:

```shell
error[E0463]: can't find crate for `std`
  |
  = note: the `riscv64gc-unknown-none-elf` target may not support the standard library
  = note: `std` is required by `xv6_rust_sample` because it does not declare `#![no_std]`
  = help: consider building the standard library from source with `cargo build -Zbuild-std`
```

Surprise! You have met the first issue in our journey, risc-v toolchain doesn't support the `std` lib!

Follow the error hints, we better add the `#![no_std]` to our code. And our second challenge is just right behind: 

```shell
error: cannot find macro `println` in this scope
 --> src/main.rs:2:5
  |
2 |     println!("Hello, world!");
  |     ^^^^^^^
```

Without the standard library, we lose access to `println!()` functionality.

This is indeed the case. Implementing a `printf!()` macro will require significant work that we'll cover in a future article, meaning we won't have this functionality available today.

Since we can't print output yet, we'll need to modify our approach and use this simpler test case instead:

```rust
#![no_std]
fn main() {
    let mut i = 999;
    i = i + 1;
}
```

We can verify successful execution by checking the value of `i` in the debugger - if it equals 1000, this confirms our Rust code is running correctly on RISC-V.

Next, we encounter our third challenge:

```shell
error: `#[panic_handler]` function required, but not found
```

The RISC-V toolchain doesn't include a built-in panic handler implementation.

Let's implement a basic panic handler:

```rust
#[panic_handler]
pub fn panic(info: &core::panic::PanicInfo) -> ! {
    loop {
        unsafe { core::arch::asm!("wfi") }
    }
}
```

The code includes an unfamiliar instruction - "wfi". This stands for "Wait for Interrupt" (not to be confused with WiFi). 

As defined in the [RISC-V ISA specification](https://riscv.org/wp-content/uploads/2017/05/riscv-privileged-v1.10.pdf) section 3.2.3: "The Wait for Interrupt instruction (WFI) provides a hint to the implementation that the current hart (hardware thread) can be stalled until an interrupt might need servicing."

This implementation creates an infinite loop with WFI instructions, effectively stalling the CPU on panic. The `core::arch::asm!()` macro enables inline assembly in our no_std environment. Note that the `core` library (unlike `std`) doesn't include printing functionality - see the [core documentation](https://doc.rust-lang.org/core/#) for details.

After implementing the panic handler and running `cargo run` again, we encounter our final challenge: 

```shell
error: requires `start` lang_item
```

The [`lang_item`](https://doc.rust-lang.org/beta/unstable-book/language-features/lang-items.html) attribute marks functions that implement special compiler functionality like memory management or exception handling. This error occurs because `no_std` requires us to explicitly define these compiler hooks. 

The standard library normally handles all `lang_item` requirements automatically, but with `no_std` we must implement them ourselves.

The `start` language item defines the program entry point. The standard library normally links this to `main()`, but without `std` this connection is broken.

We resolve this by adding `#![no_main]`, informing the compiler that we'll define our own program entry point, which eliminates this error.

With these changes in place, let's test if the code executes correctly. While Rust's compilation checks catch many issues, runtime verification remains important.

Here's our current implementation:

```rust
#![no_std]
#![no_main]
fn main() {
    let mut i = 999;
    i = i + 1;
}

#[panic_handler]
pub fn panic(info: &core::panic::PanicInfo) -> ! {
    loop {
        unsafe { core::arch::asm!("wfi") }
    }
}
```

We run it, then we'll get:

```shell
target/riscv64gc-unknown-none-elf/debug/xv6-rust-sample: target/riscv64gc-unknown-none-elf/debug/xv6-rust-sample: cannot execute binary file
```

This occurs because we're attempting to run a RISC-V binary on an x86 platform. Our compiled output targets RISC-V architecture and requires the appropriate execution environment.

To execute this binary, we need a RISC-V environment. A virtual machine provides an efficient solution for cross-architecture execution during development.



## 2. Configuring a RISC-V Environment Using QEMU

Having successfully compiled our Rust code for RISC-V, we now require an execution environment. While real hardware like Raspberry PI could be used, a virtual machine offers rapid setup and significant time savings during development.

We selected QEMU for its ease of use, open-source nature, and seamless Rust integration.

Configuring QEMU integration requires just two additions to our existing `.cargo/config.toml`:

``` toml
[build]
target = "riscv64gc-unknown-none-elf"

### The following lines are newly added
[target.riscv64gc-unknown-none-elf]
runner = "qemu-system-riscv64 -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel "
```

I'm not gonna describe much detail of QEMU in this article, please check [here](https://www.qemu.org/docs/master/system/invocation.html#hxtool-0) to see the usage of QEMU if you needed.

In short, in the above lines, we set the "runner" of target "riscv64gc-unknown-none-elf" as a command line that can bring up QEMU. Note that we're using "qemu-system-riscv64" - QEMU provides separate binaries for different target architectures.

The [runner configuration](https://doc.rust-lang.org/cargo/reference/config.html#targettriplerunner) automatically passes our compiled binary to the QEMU command, which explains why we placed the `-kernel` parameter at the end.

With everything configured, we're ready to test the implementation. 

```shell
Finished dev [unoptimized + debuginfo] target(s) in 0.99s
     Running `qemu-system-riscv64 -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel target/riscv64gc-unknown-none-elf/debug/xv6-rust-sample`
```

The execution succeeded without the previous architecture error, and QEMU is now running our RISC-V binary.

Without output capabilities, we must verify program execution through debug mode by inspecting the value of `i` in memory.



## 3. Debugging as a Primary Development Tool

Debugging tools remain essential throughout development. Without them, we'd be limited to basic logging and unable to thoroughly inspect program state.

This guide uses GDB for debugging, though alternatives like LLDB or Rust-GDB offer similar functionality.

To integrate GDB debugging into our workflow:

First, we configure QEMU to accept GDB connections by adding two parameters to our runner command: `-s -S`. These options make QEMU listen for a debugger connection and pause execution until connected.

```toml
[target.riscv64gc-unknown-none-elf]
runner = "qemu-system-riscv64 -s -S -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel "
```

The `-s` parameter is equivalent to `-gdb tcp::1234`, configuring QEMU to listen for GDB connections on TCP port 1234.

The `-S` parameter instructs QEMU to pause execution until a GDB connection is established.

Next, we configure remote debugging in GDB. Using CLion as our IDE, we create a remote debug configuration specifying `localhost:1234` as the target and `target/riscv64gc-unknown-none-elf/debug/xv6-rust-sample` as the symbol file.

With this setup, running `cargo run` and initiating debugging in CLion (with a breakpoint at `main()`) should show a successful GDB connection. Terminating the debugger will also stop QEMU, indicated by the message: `qemu-system-riscv64: QEMU: Terminated via GDBstub`.

However, despite the debugger connection, the program doesn't execute as expected. The reason becomes clear when we examine our implementation:

The issue stems from our `#![no_main]` declaration. While this tells the Rust compiler we'll handle the program entry point ourselves, we haven't actually implemented one yet.

To resolve this, we must explicitly define the program entry point using a [linker script](https://sourceware.org/binutils/docs/ld/Scripts.html) (`.ld` file) that specifies the memory layout:

```ld
/* entry.ld */
OUTPUT_ARCH( "riscv" )
ENTRY( main )

SECTIONS
{
  . = 0x80000000;

  .text : {
    *(.text*)
  }

  .rodata : {
    *(.rodata*)
  }

  .data : {
    *(.data*)
  }

  .bss : {
    *(.bss*)
  }
}
```

The linker script defines the memory layout of our ELF-format binary, following standard RISC-V conventions.

This linker script follows standard ELF section conventions, with the key configuration elements at the beginning:

`OUTPUT_ARCH( "riscv" )` indicates the target file is for the risc-v platform. And `ENTRY( main )` points out our program entry is a symbol called `main`, which is our main function indeed.

The `. = 0x80000000` directive sets the entry point address to match RISC-V's memory layout, where RAM begins at 0x80000000. We can verify this memory mapping in QEMU:

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

Launching QEMU with `-monitor stdio` provides an interactive interface where we can inspect memory regions using `info mtree`, confirming RAM begins at `0x80000000`.

With the linker script created, we must configure Cargo to use it by modifying `.cargo/config.toml`:

```toml
[build]
target = "riscv64gc-unknown-none-elf"
### The following line is newly added
rustflags = ['-Clink-arg=-Tsrc/entry.ld']

[target.riscv64gc-unknown-none-elf]
runner = "qemu-system-riscv64 -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel "
```

And one another step is, let the entry symbol be recognized. Generally, rust would mangle most of the symbols, to make sure all symbols have their own unique name. But as we have set the `ENTRY( main )`, we need to let the `main` stays "main", not other mangled names. To achieve that, we have to change the function signature of `main()` like this:

```rust
... ...
#[no_mangle]
extern "C" fn main() {
... ...
```

The `#[no_mangle]` attribute prevents name mangling, while `extern "C"` ensures the function uses the C calling convention for interoperability.

With these modifications complete, the debugger should now successfully pause execution at the first line of `main()` when a breakpoint is set.



## 4. Debugging Challenges Emerge

While we've successfully set breakpoints at `main()`, debugging reveals unexpected behavior - the program counter shows 0x0 when paused, and stepping through instructions fails. This indicates a deeper issue with our execution environment.

RISC-V's `mcause` CSR (Control and Status Register) records trap causes, which we can inspect to diagnose the execution failure.

Executing `info all-registers` in GDB displays all register values:

```gdb
(gdb) info all-registers
zero           0x0      0
ra             0x0      0x0
sp             0xfffffffffffffff0       0xfffffffffffffff0
gp             0x0      0x0
tp             0x0      0x0

... ...

mscratch       0x0      0
mepc           0x0      0
mcause         0x1      1
mtval          0x0      0

... ...
```

The `mcause` value of `0x1` indicates an "Instruction access fault" according to the [RISC-V privileged specification](https://drive.google.com/file/d/17GeetSnT5wW3xNuAHI95-SI1gPGd5sJ_/view), meaning our program failed to execute its first instruction.

This fault is unexpected since we're running in machine mode (the highest privilege level) without any memory protection mechanisms enabled. The program should have complete system access.

To investigate further, we'll examine the disassembled program using objdump:

```assembly
0000000080000000 <main>:
    80000000:   1141                    addi    sp,sp,-16
    80000002:   3e700513                li      a0,999
    80000006:   c62a                    sw      a0,12(sp)
    80000008:   45b2                    lw      a1,12(sp)
    8000000a:   0015851b                addiw   a0,a1,1
    8000000e:   e02a                    sd      a0,0(sp)
    80000010:   00b54763                blt     a0,a1,8000001e <.Lpcrel_hi0>
    80000014:   a009                    j       80000016 <main+0x16>
    80000016:   6502                    ld      a0,0(sp)
    80000018:   c62a                    sw      a0,12(sp)
    8000001a:   0141                    addi    sp,sp,16
    8000001c:   8082                    ret
```

The issue stems from the stack pointer initialization - starting at 0x0, the first instruction sets `sp` to `0xfffffffffffffff0` (64-bit wraparound from 0x0 - 0x10), which is outside our allocated 128MB memory range (0x80000000 ~ 0x88000000). 

This causes the subsequent store instruction (80000006) to attempt accessing `0xfffffffffffffffc`, far outside our valid memory range of 0x80000000-0x88000000.

To resolve this, we need to properly initialize the stack pointer before execution:

```rust
#![no_std]
#![no_main]
#[no_mangle]
extern "C" fn main() {
    unsafe { core::arch::asm!("la sp, 0x80001000") }
    
    let mut i = 999;
    i = i + 1;
}
```

We add one line of asm to set the `sp` equals to `0x80001000`, since our program is quite simple and will not grow to even `0x800000ff`, so our code section is safe and has no chance to be overridden. 

Finally, the program can be run correctly, and if you like, add a `panic!()` at the end of the program, otherwise when `main()` is return, the program will fail again because we didn't tell it what to do next after `main()` returned.



## 5. What the xv6 is all about?

After all the above sections, now we can get back to talking more about xv6.

Quote from the name of the xv6 book, *xv6: a simple, Unix-like teaching operating system*. Yes, xv6 was inspired by Unix v6, and since the Unix needs to run on specific hardware like PDP-11, and with many low-level details, in 2006, MIT decided to modeled on Unix v6, rewrite it by ANSI C, with multiprocessor support been added, at last created xv6.

As we mentioned at the beginning of this article, the xv6 was running on x86 at first, but then they ported it to risc-v. That's why we provide this entire article to discuss how to run rust on risc-v platform, with the knowledge in this article, I believe we could get our local environment ready to go, and get to know some basic low-level information about risc-v instructions, linker script and ASM in rust.

Basically, although it does not contain many lines of code, xv6 is still a full functional operating system, it has virtualized CPU and memory as process and virtual memory, it supports concurrency, and contains an Unix-like file system to implement persistent. It has user space and kernel space, with a group of system calls (but not compliant with POSIX for clarity and simplicity). Like Unix, xv6 remains macro kernel concept, so it has only one kernel binary.

In the next articles, we will take a close look at the detailed components design of xv6, and then try to port each one of the components to rust...

