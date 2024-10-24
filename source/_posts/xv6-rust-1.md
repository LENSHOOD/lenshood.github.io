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

Originally, xv6 was written in C, which is awesome for students to get hands-on experience with such a classic programming language. But now that Rust is gaining traction—especially since rust-for-linux is part of the main Linux line—wouldn’t it be fun to run xv6 using Rust? 

As a perfect way to kill the time, I have migrated most of xv6 from C to Rust. You can check it out [here](https://github.com/LENSHOOD/xv6-rust). During this migration process, I encountered many sorts of issues and tricky stuff, nothing brings me more satisfaction than successfully resolving a problem!

Therefore, I believe it would be cool for me to share my experiences through a series of articles detailing how I did this, complete with a more structured approach and clear procedures.

All right, let's get started...

## 1. Rust on risc-v

In some previous versions of the xv6, it running on x86 arch, however, currently the xv6 has fully migrated to the risc-v arch. 

As we are going to port the xv6 to rust, in the very first, we better take a look about how to run rust on risc-v.

> Please note that in these series of articles, I will assume the reader has basic knowledge about rust, and knows how to setup the local environment such as rustup, cargo, and IDE.
>
> In the following articles, I will use my local machine as the demo env, and the that includes:
>
> - MBP 2019, Intel
> - cargo 1.75.0-nightly (b4d18d4bd 2023-10-31), some features rely on nightly build
> - ustup 1.26.0 (5af9b9484 2023-04-05)
> - CLion 2024.1, (I didn't choose RustRover, because after gave it a try I found it still not stable)

Here we go! Let's create a new rust project, and name it as "xv6-rust-sampe".

Now, there is one and only one `main.rs` file lies in the `src/` directory, leave it for a sec, we don't need that file right now.

Remember, we are going to run rust code on risc-v arch, before any coding, we should deal with the toolchain in the first place, after all I bet our code cannot run correctly when it been compiled as x86 right?

To choose the correct toolchain, let's create a `.cargo/config.toml`  in the project root, which is the cargo configuration file of our project, we could set our building target here.

Add just two lines in the toml file, like this:

``` toml
[build]
target = "riscv64gc-unknown-none-elf"
```

This will tell our rust toolchain that in this project, we would like to have a riscv program as the output.

So far so good.

In the next step, we go back to the `main.rs`, and there is only one very simple function:

```rust
fn main() {
    println!("Hello, world!");
}
```

However, if you type and execute `cargo run` with confidently, then you would probably got this:

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

Surprise! You have met the first issue in our journey, riscv toolchain doesn't support the `std` lib!

Follow the error hints, we better add the `#![no_std]` to our code. And our second challenge just right behind: no `std` no `println!()`, WTF?

Unfortunately, yes. Actually we will take one whole chapter to implement `printf!()` macro in the next article, which also means, we are not gonna have it today.

It's a bit of awkward for us to not able print a simple "hello world", then we can only take one step back, change to this:

```rust
#![no_std]
fn main() {
    let mut i = 999;
    i = i + 1;
}
```

Hopefully, at least we can check the value of `i` in debugger, and if the value is equal to 1000, then it can prove we successfully run rust code on riscv as well.

The third monster shows up:

```shell
error: `#[panic_handler]` function required, but not found
```

In the toolchain of riscv, it even doesn't have its builtin panic handler!

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

Essentially, we put a "wfi" into a loop, means if any panic happens, instead of report some error, we just let the cpu stall. Besides, the macro "core::arch::asm!()" is a wrapper that will let us easily run assembly in rust code, since there is no `std` lib here, we replace it as `core`(not surprisingly, it doesn't contain a `println!()`), for more details about `core`, check [here](https://doc.rust-lang.org/core/#).

All right, after added the panic handler, and re-run `cargo run`, we will get our final issue: 

```shell
error: requires `start` lang_item
```

[`lang_item`](https://doc.rust-lang.org/beta/unstable-book/language-features/lang-items.html) is a set of items, defined by compiler, to implement special features for the language, for example memory management, exception management, etc., the above error is like a "side effect" of `no_std`. 

Generally the `std` lib by default takes care of all of the special cases related to `lang_item`, once we set `no_std`, many language items need to be provided by ourselves.

The `start` language item is to define the entry point of the program. Since `std` did great job to link the program entry to `main()`, but as the same, no std, no main.

To solve the issue, in the first step we should add `#![no_main]` to our code, that will let the compiler realize we will define our own program entry, hence the compiler will no longer report above error then.

The second step, we need let risc-v understand where to start run our code. And that requires a linker script `.ld`.







## 2. Running on virtual hardware



## 3. Debugger is our closest friend



## 4. What the xv6 is all about?



