---
title: [xv6 rust 0x1] Starting is the hardest part
date: 2024-10-23 22:27:59
tags:
- xv6
- rust
- os
categories:
- Rust
---

[Xv6](https://github.com/mit-pdos/xv6-public) is one of the best operating systems for teaching. It’s a great way to learn about how an OS works with basic functions and few line of code. 

Originally, xv6 was written in C, which is awesome for students to get hands-on experience with such a classic programming language. But now that Rust is gaining traction—especially since rust-for-linux is becoming a part of the main line Linux—wouldn’t it be fun to run xv6 using Rust? 

As a perfect way to kill time, I have migrated most of xv6 from C to Rust. You can check it out [here](https://github.com/LENSHOOD/xv6-rust). During this migration process, I encountered many sorts of issues and tricky stuff, nothing brings me more satisfaction than successfully resolving a problem!

Therefore, I believe it would be cool for me to share my experiences through a series of articles detailing how I did this, complete with a more structured approach and clear procedures.

All right, let's get started...

## 1. Rust on risc-v

In some previous versions of the xv6, it running on x86 arch, however, currently the xv6 has fully migrated to the risc-v arch. 

As we are going to port the xv6 to rust, at the very first, we better take a look about how to run rust on risc-v.

> Please note that in these series of articles, I will assume the reader has basic knowledge about rust, and knows how to setup the local environment such as rustup, cargo, and IDE.
>
> In the following articles, I will use my local machine as the demo env, and that includes:
>
> - MBP 2019, Intel
> - cargo 1.75.0-nightly (b4d18d4bd 2023-10-31), some features rely on nightly build
> - ustup 1.26.0 (5af9b9484 2023-04-05)
> - CLion 2024.1, (I didn't choose RustRover, because after giving it a try I found it still not stable)

Here we go! Let's create a new rust project, and name it "xv6-rust-sampe".

Now, there is one and only one `main.rs` file lies in the `src/` directory, leave it for a sec, we don't need that file right now.

Remember, we are going to run rust code on risc-v arch, before any coding, we should deal with the toolchain in the first place, after all I bet our code cannot run correctly when it has been compiled as x86 right?

To choose the correct toolchain, let's create a `.cargo/config.toml`  in the project root, which is the cargo configuration file of our project, we could set our building target here.

Add just two lines in the toml file, like this:

``` toml
[build]
target = "riscv64gc-unknown-none-elf"
```

This will tell our rust toolchain that in this project, we would like to have a risc-v program as the output.

So far so good.

In the next step, we go back to the `main.rs`, and there is only one very simple function:

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

error: cannot find macro `println` in this scope
 --> src/main.rs:2:5
  |
2 |     println!("Hello, world!");
  |     ^^^^^^^

error: `#[panic_handler]` function required, but not found

For more information about this error, try `rustc --explain E0463`.
error: could not compile `xv6-rust-sample` (bin "xv6-rust-sample") due to 3 previous errors
```

Surprise! You have met the first issue in our journey, risc-v toolchain doesn't support the `std` lib!

Follow the error hints, we better add the `#![no_std]` to our code. And our second challenge is just right behind: no `std` no `println!()`, WTF?

Unfortunately, yes. Actually we will take one whole chapter to implement `printf!()` macro in the next article, which also means, we are not gonna have it today.

It's a bit awkward for us to not be able to print a simple "hello world", then we can only take one step back, change to this:

```rust
#![no_std]
fn main() {
    let mut i = 999;
    i = i + 1;
}
```

Hopefully, at least we can check the value of `i` in debugger, and if the value is equal to 1000, then it can prove we successfully run rust code on risc-v as well.

The third monster shows up:

```shell
error: `#[panic_handler]` function required, but not found
```

In the toolchain of risc-v, it doesn't even have its builtin panic handler!

Fine, let's build one to it:

```rust
#[panic_handler]
pub fn panic(info: &core::panic::PanicInfo) -> ! {
    loop {
        unsafe { core::arch::asm!("wfi") }
    }
}
```

In the above code, some weird stuff shows up. What is "wfi"? (at least I can assure you it's not wifi) 

According to the [risc-v ISA](https://riscv.org/wp-content/uploads/2017/05/riscv-privileged-v1.10.pdf) section 3.2.3: "The Wait for Interrupt instruction (WFI) provides a hint to the implementation that the current hart can be stalled until an interrupt might need servicing', FYI, the word "hart" means hardware thread.

Essentially, we put a "wfi" into a loop, which means if any panic happens, instead of reporting some error, we just let the cpu stall. Besides, the macro "core::arch::asm!()" is a wrapper that will let us easily run assembly in rust code, since there is no `std` lib here, we replace it as `core`(not surprisingly, it doesn't contain a `println!()`), for more details about `core`, check [here](https://doc.rust-lang.org/core/#).

All right, after adding the panic handler, and re-run `cargo run`, we will get our final issue: 

```shell
error: requires `start` lang_item
```

[`lang_item`](https://doc.rust-lang.org/beta/unstable-book/language-features/lang-items.html) is a set of items, defined by compiler, to implement special features for the language, for example memory management, exception management, etc., the above error is like a "side effect" of `no_std`. 

Generally the `std` lib by default takes care of all of the special cases related to `lang_item`, once we set `no_std`, many language items need to be provided by ourselves.

The `start` language item is to define the entry point of the program. Since `std` did a great job to link the program entry to `main()`, so just like the above case, no std, no main.

To solve the issue, we should add the `#![no_main]` to our code, that will let the compiler realize we will define our own program entry, hence the compiler will no longer report above error then.

Up to now, maybe we could try running the code to see if everything goes well? After all, in many cases of rust, pass compile means pass everything.

Let's recap the current code:

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

Basically, that means we run the binary in a wrong arch. As we are aware, the target binary we would like to have is a program that can be run on risc-v platform, not our x86 platform.

In simple terms, we need a risc-v env to run the binary. And in such circumstances, virtual machine is a great choice for us.



## 2. Setup risc-v platform based on QEMU

We have successfully compiled the example rust code with risc-v target. Now we need to have a virtual machine to simulate the risc-v environment, of course you can do it on real hardware like Raspberry PI, but virtual machine can help us setup the target platform in a second, that would incredibly save time in the initial stages of development.

Here, we choose QEMU because it's very easy to use, open soured and could integrate to rust seamlessly.

It's quite simple to setup the QEMU with rust integration, we only need to add two lines in the previous `.cargo/config.toml`:

``` toml
[build]
target = "riscv64gc-unknown-none-elf"

### The following lines are newly added
[target.riscv64gc-unknown-none-elf]
runner = "qemu-system-riscv64 -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel "
```

I'm not gonna describe much detail of QEMU in this article, please check [here](https://www.qemu.org/docs/master/system/invocation.html#hxtool-0) to see the usage of QEMU if you needed.

In short, in the above lines, we set the "runner" of target "riscv64gc-unknown-none-elf" as a command line that can bring up QEMU. You may already noticed, the QEMU binary we execute is "qemu-system-riscv64", which means there are many different binaries that are for other platforms.

Once we set the [runner](https://doc.rust-lang.org/cargo/reference/config.html#targettriplerunner) for some target, then every time we execute cargo `cargo run`, the target file of our program will be passed as an argument to the command we put into "runner" field. That also why we put the `-kernel` param of QEMU in the end.

All set, let's give it a try! 

```shell
Finished dev [unoptimized + debuginfo] target(s) in 0.99s
     Running `qemu-system-riscv64 -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel target/riscv64gc-unknown-none-elf/debug/xv6-rust-sample`
```

It went very well, no "cannot execute binary file" error ever again, and the QEMU seems running.

Because we can't output anything, the only way to verify the correctness of our program is run as debug mode, and check the value of "i" in memory.



## 3. Debugger is our closest friend

No matter now or later, the debugger is always a super important helper to us. Without a debugger, we cannot learn the current program status easily, and can only use the logger to print context with many restrictions.

I'm using the GDB as the debugger in the next series of articles, but you can also choose other debuggers like lldb or rust-gdb, they are quite the same.

So, how to introduce gdb into our project?

Step 1, we need to let QEMU be able to accept a GDB connection, additionally, pause QEMU to wait for a gdb connection. That requires us to add two params: `-s -S`in to the runner command:

```toml
[target.riscv64gc-unknown-none-elf]
runner = "qemu-system-riscv64 -s -S -machine virt -bios none -m 128M -smp 1 -nographic -global virtio-mmio.force-legacy=false -kernel "
```

The `-s` is a shorthand for `-gdb tcp::1234`, which means to listen the GDB connection on tcp port 1234.

The `-S` ask QEMU not start to run until a GDB connection comes in.

Step 2, run GDB in remote debug mode. I use Clion as my local IDE, so I can simply create a remote debug in the "Run/Debug Configuration", with the remote args as `localhost:1234`, and choose the symbol file to `target/riscv64gc-unknown-none-elf/debug/xv6-rust-sample`.

After completing the above two steps, when we run `cargo run` again, set a breakpoint in the first line of `main()`, and click debug on Clion, we should see the `Debugger connected to localhost:1234` in the debug tab. And if we stop the debugger, QEMU will stop too, and shows: `qemu-system-riscv64: QEMU: Terminated via GDBstub`.

But nothing happens except the debugger connected. Why?

Actually, we haven't completed our program when we added `#![no_main]` in the previous content. `#![no_main]` only tells rust compiler "you don't need to worry about the program entry anymore, we the developer will take care of that". But in fact we didn't do anything related to the program entry at all!

Hence, right now we need to let QEMU understand where to start running our code. And that requires a [linker script](https://sourceware.org/binutils/docs/ld/Scripts.html) `.ld`, just like this:

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

So linker script is basically define the memory layout of the output binary, and since the rust risc-v cross toolchain will generate the target file as ELF format, we defined the layout as ELF style.

The above `ld` file is quite simple, the four sections are basic ELF sections and nothing special here. The only fields we need to put our eyes on are the fields on top.

`OUTPUT_ARCH( "riscv" )` indicates the target file is for the risc-v platform. And `ENTRY( main )` points out our program entry is a symbol called `main`, which is our main function indeed.

The `. = 0x80000000` stands for putting the entry onto the address 0x80000000, so that our binary will start from that. QEMU supports many different hardware architectures, particularly in risc-v, the RAM address starts from 0x80000000. We can execute a very simple command to prove that:

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

We start a risc-v virtual machine with `-monitor stdio`, that would not just run a VM instance, but also bring us into an interactive interface, we can check the current memory regions by `info mtree`, apparently the RAM begins at `0000000080000000`.

Now we have our `ld` file, but we still need to activate the script in our program, which need to modify the `.cargo/config.toml`:

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

`#[no_mangle]` force the compiler not mangle this function, and `extern "C" ` is a declaration of FFI, to export the function with C ABI.

After all of the above modifications, I'm sure the program can stop at the first line of main if there is a breakpoint.



## 4. Things are getting complicated

In the previous chapter, I ensured the program could be stopped at the first line, but I bet you have tried, the debugger can no longer step over to the second line. And if you pause the program through gdb, the current memory address turns out to be 0x0. Something's wrong here.

There is a CSR called `mstatus` in risc-v to indicate any event that caused the trap, we could check the value of `mcause` to investigate why our program is in a failure.

Execut `info all-registers` in gdb, will show value of all registers:

```gdb
(gdb) info all-registers
zero           0x0	0
ra             0x0	0x0
sp             0xfffffffffffffff0	0xfffffffffffffff0
gp             0x0	0x0
tp             0x0	0x0

... ...

mscratch       0x0	0
mepc           0x0	0
mcause         0x1	1
mtval          0x0	0

... ...
```

The `mcause` shows value of `0x1`, refer to the risc-v document([Table 14. Machine cause register (mcause) values after trap.](https://drive.google.com/file/d/17GeetSnT5wW3xNuAHI95-SI1gPGd5sJ_/view)), `0x1` means "Instruction access fault".

But how can it be? It won't be insufficient access permission, after all we haven't set any privileged level, so our program is running on the machine mode, which is the highest privileged mode, we can literally do everything.

If we move one step forward, decompile the program with obj-dump, and see the assembly code here:

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

Yes, the stack pointer! `sp` is initially zero, so that after line 80000000, the `sp` will be set to `0x0 - 0xf = `(we are on the 64-bit platform). 

Unfortunately, at line 80000006, the value of a0 will be saved to `sp + 12`, which is `0xfffffffffffffffc`, but obviously this address is illegal. If you remember, we only create a VM with 128MiB memory, which means the available physical address range is `0x80000000 ~ 0x88000000`.

To make it correct, let's set the `sp` in the first place:

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



