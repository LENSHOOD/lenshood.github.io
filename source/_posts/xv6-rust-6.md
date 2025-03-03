---
title: /Xv6 Rust 0x06/ - User Space
date: 2025-02-24 14:35:03
tags:
- xv6
- rust
- os
categories:
- Rust
---

Based on disk and file management, now we are able to store the user space program on the disk, and let them running after kernel started. But before that, there is still a topic we haven't covered: how does xv6 jump from kernel space to user space? 

After all, the content we talked in previous chapters is only limited in the supervisor level, even machine level, where the code has full control of hardware. However, the user space program cannot be granted such huge scope of control, then we should know how to jump from kernel space to user space, so that we could provide a safer environment for the user program.

In this chapter, we are going to find it out.

<!-- more -->

## 1. Jumping in the CPU Perspective

If you remember, there was a table in our second chapter, describes several CSRs that risc-v provides to user, some of them are responsible for mode switching.

Speaking of how to jumping from supervisor mode to user mode, there would be the following questions come up with in your mind:

- What kind of instruction is able to trigger the switching?
- After jumping to user mode, where exactly the program will go to?
- How to deal with the context and different memory space between two modes?

At first, let's recap the privilege mode switch that we've mentioned in second chapter:

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

Let's take a close look at the `sret` instruction:

{% asset_img 1.png %}

Apparently, `SRET` doesn't rely any source or destination register, so when using the `SRET`, we only need to call the bare instruction.

According to the specification, *`xRET` sets the `pc` to the value stored in the `xepc` register.* Hence, before `SRET` is called, we could set the address into the `sepc`, then once it called, the program will be jump into the address.

So far, it looks `SRET` does a lot of things for us, so that we'll no longer need to concern about the first two questions. However, in risc-v architecture, no more support will be provided. Now, for the question of context and memory space switch, we are on our own.

Imagine the kernel is about to complete initialization, and program is running on the supervisor mode. Now, the kernel should start creating the very first process in the whole system, we call it `init`. Assuming that a few milliseconds later, the process structure has been created and all of the importance fields have been set, next the kernel must think about runs the `SRET`  instruction, and hands the control of CPU to `init`.

But before calling the `SRET`, both the context and memory space should be replaced as well, because:

- Context Switch: the context here means the general purpose registers, there are two main reasons that the context switch should be done; first, the value of registers in supervisor mode must not be leaked to user mode for safety; second, in other cases like syscall or interrupt handling, we need to make sure when go back to user mode, the user process can resume correctly with all registers still store the origin values, that requires properly context switch too.
- Memory Space Switch: we have known that there is a kernel page table dedicated for kernel code, if we don't set the user process page table after switch mode, then the user process is able to access kernel memory space, which is extremely dangerous; besides, kernel page table does not hold user code in the text section, makes the user process unable to get its code.

Therefore, it's essential for kernel to take care of the context and memory space switch. The following is a diagram that shows the process of switching from supervisor mode to user mode:

{% asset_img 2.png %}

Firstly, there should be some memory spaces allocated to hold the pre-stored registers, additionally, the page table is created along with creating of `Proc` structure (see [`inner_alloc()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L474)).

Secondly, the address of user process page table should be set into `satp`, and the value of general purpose registers should also be restored.

Finally, put the user space address (virtual address) into the `sepc`, and call `SRET` at the end of the program. After that, everything is changed to user space!



## 2. Trap and Trampoline

As there are some extra work needs to be done before switching into user space, where does that need to happen?

We haven't mentioned the full address layput of xv6 in previous chapters, now it's time to show both kernel address layout and process address layout, these are very helpful to understanding the concept of "trampoline". Let's have a look! (The following diagrams are taken from [the xv6 book](https://pdos.csail.mit.edu/6.828/2024/xv6/book-riscv-rev4.pdf), figure 3.3 and figure 3.4)

{% asset_img 3.png %}

Above is the kernel address layout that includes virtual address space on left and physical address space on right. Follow the sequence of low address to high address, which is also bottom to top, the kernel address space can be divided into several parts (please note that the mappings between virtual space and physical space in kernel is a little complicated, we'll introduce them together, hopefully, the what we have learnt in previous chapters can help us for better understanding):

- (Physical) boot ROM: qemu actually provide this
- (Physical) core local interrupter: it contains a timer
- (Physical + Virtual) PLIC, UART0, VIRTIO disk: we have talked about them before
- (Physical + Virtual) Kernel memory: 
  - Text section contains all kernel code
  - Data section stores some constants and statics
  - Free memory holds all other data, including kernel objects and process data (refer to [`kalloc`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/kalloc.rs#L80))

- (Virtual) Process stacks: each process has it own stack, which is allocated here, actually they are allocated from the "Free memory" section
- (Virtual) Trampoline: interesting section, according to above diagram, it maps to the address near the [KERNBASE](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/memlayout.rs#L84), which is also the same address of text section. Is this a coincidence?

As the kernel vm init code shows:

```rust
// vm.rs
fn kvmmake<'a>() -> &'a PageTable {
    ... ...

    let trapoline_addr = (unsafe { &trampoline } as *const u8).addr();
    // map the trampoline for trap entry/exit to
    // the highest virtual address in the kernel.
    kvmmap(kpgtbl, TRAMPOLINE, trapoline_addr, PGSIZE, PTE_R | PTE_X);
    // printf!("TRAMPOLINE Mapped.\n");

		... ...
}

// memlayout.rs
pub const TRAMPOLINE: usize = MAXVA - PGSIZE;
```

The value of virtual address Trampoline is `MAXVA - PGSIZE`, which means the trampoline section is located in the top address and takes one page of space. This is consistent with the above diagram.

However, it maps to the physical address: trapoline_addr, which is actually a label "trampoline" that is defined in the [trampoline.S](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L17):

```assembly
### trampoline.S

.section trampsec
.globl trampoline
trampoline:
.align 4
.globl uservec
uservec:    
	    ... ...

.globl userret
userret:
      ... ...
```

Now I suppose you already know why it maps to the read-only text section, because the `trampoline` points to some code that used to handle the trap and trap return.

The location of the trampoline is intentional. Let's see the process address layout:

{% asset_img 4.png %}

I guess most of the sections in above diagram are very familiar to you, because they are no difference from other modern operating systems, except for the trampoline.

The most obvious similarity is the address of trampoline in process address space is exactly the same as it in kernel address space. Why? Because each time a trap happens in user mode, risc-v switching to supervisor mode, and then redirect the program to [`uservec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L20), which is the trap handler address, we'll see the registration of the [`uservec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L20) afterward.

First, let's take a close look at it:

```assembly
### trampoline.S
uservec:    
	    #
      # trap.c sets stvec to point here, so
      # traps from user space start here,
      # in supervisor mode, but with a
      # user page table.
      #

      # save user a0 in sscratch so
      # a0 can be used to get at TRAPFRAME.
      csrw sscratch, a0

      # each process has a separate p->trapframe memory area,
      # but it's mapped to the same virtual address
      # (TRAPFRAME) in every process's user page table.
      # there is no "#define" in rust, so directly copy the TRAPFRAME value here
      # li a0, TRAPFRAME
      li a0, 274877898752

      # save the user registers in TRAPFRAME
      sd ra, 40(a0)
      ... ...
      sd t6, 280(a0)

      # save the user a0 in p->trapframe->a0
      csrr t0, sscratch
      sd t0, 112(a0)

      # initialize kernel stack pointer, from p->trapframe->kernel_sp
      ld sp, 8(a0)

      # make tp hold the current hartid, from p->trapframe->kernel_hartid
      ld tp, 32(a0)

      # load the address of usertrap(), from p->trapframe->kernel_trap
      ld t0, 16(a0)


      # fetch the kernel page table address, from p->trapframe->kernel_satp.
      ld t1, 0(a0)

      # wait for any previous memory operations to complete, so that
      # they use the user page table.
      sfence.vma zero, zero

      # install the kernel page table.
      csrw satp, t1

      # flush now-stale user entries from the TLB.
      sfence.vma zero, zero

      # jump to usertrap(), which does not return
      jr t0
```

Apparently, once the program goes to it, value of many registers are saved into [`Trapframe`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L112), which is allocated in the [`inner_alloc()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L465). Basically this process is the context switching that we discussed before. All user registers are saved.

But the most important line is `csrw satp, t1`, which installs the kernel page table, you may ask a question at this stage: does that mean, before this line, although the risc-v has been switched to supervisor mode, the xv6 still running on user address space? 

Exactly! That's the essential reason of trampoline section should share the same address between the kernel space and user space. Otherwise the [`uservec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L20) cannot be correctly located if trap happens. Because there is no place that allows page table switching before trap.

Additionally, after install the kernel page table, what if an external interrupt happens? At this moment, the trap vector is still set to [`uservec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L20), if the kernel space and user space don't share the same trampoline address, there would be some chance to jump into an undefined address that is translated by kernel page table.

### 2.1 trap handler

After done the context and memory space switching, at the end of the [`uservec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L20), it jumps to the real trap handler function, which is called [`usertrap()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L41):

```rust
// trap.rs

//
// handle an interrupt, exception, or system call from user space.
// called from trampoline.S
//
fn usertrap() {
    if (r_sstatus() & SSTATUS_SPP) != 0 {
        panic!("usertrap: not from user mode");
    }

    // send interrupts and exceptions to kerneltrap(),
    // since we're now in the kernel.
    w_stvec((unsafe { &kernelvec } as *const u8).addr());

    let p = myproc();

    // save user program counter.
    let tf = unsafe { p.trapframe.unwrap().as_mut().unwrap() };
    tf.epc = r_sepc() as u64;

    let mut which_dev = 0;
    if r_scause() == 8 {
        // system call

        if p.killed() != 0 {
            exit(-1);
        }

        // sepc points to the ecall instruction,
        // but we want to return to the next instruction.
        tf.epc += 4;

        // an interrupt will change sepc, scause, and sstatus,
        // so enable only now that we're done with those registers.
        intr_on();

        syscall();
    } else {
        which_dev = devintr();
        if which_dev != 0 {
            // ok
        } else {
            printf!(
                "usertrap(): unexpected scause {:x} pid={}\n",
                r_scause(),
                p.pid
            );
            printf!("            sepc={:x} stval={:x}\n", r_sepc(), r_stval());
            p.setkilled();
        }
    }

    if p.killed() != 0 {
        exit(-1);
    }

    // give up the CPU if this is a timer interrupt.
    if which_dev == 2 {
        yield_curr_proc();
    }

    usertrapret();
}

```

Basically, the [`usertrap()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L41) handles interrupts, exceptions and syscalls. But before recognizing any of them, it first set the `stvec` to the [`kernelvec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/kernelvec.S#L12) which is the trap handler specific for kernel space. 

```assembly
### kernelvec.S
kernelvec:
    # make room to save registers.
    addi sp, sp, -256

    # save the registers.
    sd ra, 0(sp)
    ... ...
    sd t6, 240(sp)

    # call the C trap handler in trap.c
    call kerneltrap

    # restore registers.
    ld ra, 0(sp)
    ... ...
    ld t6, 240(sp)

    addi sp, sp, 256

    # return to whatever we were doing in the kernel.
    sret
```

In the same way, there are also context save and restore in it, the only difference is the kernel has its own trap handler: 

```rust
// trap.rs

// interrupts and exceptions from kernel code go here via kernelvec,
// on whatever the current kernel stack is.
#[no_mangle]
extern "C" fn kerneltrap() {
    let mut which_dev = 0;
    let sepc = r_sepc();
    let sstatus = r_sstatus();
    let scause = r_scause();

    if (sstatus & SSTATUS_SPP) == 0 {
        panic!("kerneltrap: not from supervisor mode");
    }
    if intr_get() {
        panic!("kerneltrap: interrupts enabled");
    }

    which_dev = devintr();
    if which_dev == 0 {
        printf!("scause {:x}\n", scause);
        printf!("sepc={:x} stval={:x}\n", r_sepc(), r_stval());
        panic!("kerneltrap");
    }

    let p = myproc();
    // give up the CPU if this is a timer interrupt.
    if which_dev == 2 && p.state == RUNNING {
        p.proc_yield();
    }

    // the yield() may have caused some traps to occur,
    // so restore trap registers for use by kernelvec.S's sepc instruction.
    w_sepc(sepc);
    w_sstatus(sstatus);
}
```

Since there is no syscalls in kernel space, the kernel trap handler only handles interrupts(external, software and timer) and exceptions, which makes it simpler than user trap handler.

Let's go back to the user trap handler, after all we just explained the first line of it.

After save the user program counter into trap frame, in the next it mainly deals with the three trap reasons: syscalls, interrupts and exceptions. We'll cover these parts in next section, now we are going to the final call: [`usertrapret()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L103):

```rust
// trap.rs

//
// return to user space
//
pub fn usertrapret() {
    let p = myproc();

    // we're about to switch the destination of traps from
    // kerneltrap() to usertrap(), so turn off interrupts until
    // we're back in user space, where usertrap() is correct.
    intr_off();

    // send syscalls, interrupts, and exceptions to uservec in trampoline.S
    let uservec_addr = (unsafe { &uservec } as *const u8).addr();
    let trampoline_addr = (unsafe { &trampoline } as *const u8).addr();
    let trampoline_uservec = TRAMPOLINE + uservec_addr - trampoline_addr;
    w_stvec(trampoline_uservec);

    // set up trapframe values that uservec will need when
    // the process next traps into the kernel.

    let trapframe = unsafe { p.trapframe.unwrap().as_mut().unwrap() };
    trapframe.kernel_satp = r_satp() as u64; // kernel page table
    trapframe.kernel_sp = (p.kstack + 2 * PGSIZE) as u64; // process's kernel stack
    trapframe.kernel_trap = usertrap as u64;
    trapframe.kernel_hartid = r_tp(); // hartid for cpuid()

    // set up the registers that trampoline.S's sret will use
    // to get to user space.

    // set S Previous Privilege mode to User.
    let mut x = r_sstatus();
    x &= !SSTATUS_SPP; // clear SPP to 0 for user mode
    x |= SSTATUS_SPIE; // enable interrupts in user mode
    w_sstatus(x);

    // set S Exception Program Counter to the saved user pc.
    w_sepc(trapframe.epc as usize);

    // tell trampoline.S the user page table to switch to.
    let satp = MAKE_SATP!((p.pagetable.unwrap() as *const PageTable).addr());

    // jump to userret in trampoline.S at the top of memory, which
    // switches to the user page table, restores user registers,
    // and switches to user mode with sret.
    let userret_addr = (unsafe { &userret } as *const u8).addr();
    let trampoline_userret = TRAMPOLINE + userret_addr - trampoline_addr;

    type UserRetFn = unsafe extern "C" fn(stap: usize);
    unsafe {
        let userret_fn: UserRetFn = core::mem::transmute(trampoline_userret);
        userret_fn(satp);
    };
}

```

In this function, the [`uservec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L20) is set to the `stvec`, and here is the only place the user trap vector is set. Next, saving the kernel space context to the trap frame so that they can be resumed in the following trap. After that, set some registers for preparation to make sure the user mode can be correctly switched afterward. At last, the address of [`userret`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L103) is called along with the page table address, the [`userret`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L103) does the reverse operation compare to the [`uservec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/trampoline.S#L20).



## 3. Syscalls, Interrupts and Exceptions

We have taken a glance to the trap handler previously, now we are going to look deeper into the handlers.

There are three types of trap, and will happen in the following circumstance respectively:

-  Syscalls: they are the bridges between user program and kernel, since there are a plenty of operations that a user program will be needed for implementing some logic, however the operating system cannot trust the user program to do so because those operations are dangerous running in user mode. So kernel provides a interface layer so that user program can just call it to get what it needs, and delegates the job to kernel.
- Interrupts: we have learnt that the effective interactive method between the peripherals and the OS is through the external interrupt, like UART and VIRTIO, and also a hardware timer we haven't talked about. The CPU needs to handle these interrupts immediately because interrupts usually don't wait in a line, that requires a trap to suspend whatever is running currently, and turn to handle the interrupt.
- Exceptions: please imagine if a user program accesses an address that is out of the range that allows it to access? No matter accidentally or maliciously, this action needs to be stopped. Therefore, if a program(both kernel code and user code) does some disallowed operation, the CPU trapped, and let the trap handler to deal with what to do next.

### 3.1 Syscalls

According to previous code in the [`usertrap()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L41):

```rust
//trap.rs

... ...
if r_scause() == 8 {
    // system call

    if p.killed() != 0 {
        exit(-1);
    }

    // sepc points to the ecall instruction,
    // but we want to return to the next instruction.
    tf.epc += 4;

    // an interrupt will change sepc, scause, and sstatus,
    // so enable only now that we're done with those registers.
    intr_on();

    syscall();
}
... ...
```

When the value of `scause` is 8, indicates a syscall happened.

{% asset_img 5.png %}

As we can see, the `scause` has two parts, interrupt and exception code. Since there is only 1 highest bit to indicate the interrupt status, as long as the bit equals to 1, a interrupt happened.

Look into the [`syscall()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/syscall/syscall.rs#L112), we will found how xv6 handles syscalls:

```rust
// syscall.rs
const SYSCALL: [Option<fn() -> u64>; 22] = {
    let mut arr: [Option<fn() -> u64>; 22] = [None; 22];
    arr[0] = None;
    arr[SYS_fork] = Some(sys_fork);
    ... ...
    arr[SYS_close] = Some(sys_close);
    arr
};

pub fn syscall() {
    let p = myproc();

    let tf = unsafe { p.trapframe.unwrap().as_mut().unwrap() };
    let num = tf.a7 as usize;

    if num > 0 && num < SYSCALL.len() && SYSCALL[num].is_some() {
        // Use num to lookup the system call function for num, call it,
        // and store its return value in p->trapframe->a0
        tf.a0 = SYSCALL[num].unwrap()();
    } else {
        printf!(
            "{} {}: unknown sys call {}\n",
            p.pid,
            core::str::from_utf8(&p.name).unwrap(),
            num
        );
        tf.a0 = u64::MAX;
    }
}
```

Basically, it retrieve the syscall number from trap frame, and then using the number maps the real syscall function from a constant array. And look at the array you may find many familiar functions such as `sys_open`, `sys_fork` and `sys_read`.

The above is how kernel handles syscalls. But you may curious how the user code trigger the syscall in user space? Let's move on to user code:

```rust
// user/src/ulib/stubs.rs
extern "C" {
    // system calls

    // Create a process, return child’s PID.
    pub fn fork() -> i32;

    ... ...
}
```

```assembly
// user/src/ulib/usys.S
.global fork
fork:
 li a7, 1 # SYS_fork
 ecall
 ret
... ...
```

Actually, there are several stub functions located in user code, and each stub relates to a few lines of assembly code, which do only one thing: call `ECALL` by syscall number as a parameter.

{% asset_img 6.png %}

In above diagram, the instruction `ECALL` and `EBREAK` share the same structure. And they behave similar as well. Beneath the `ECALL`, it actually generates an "environment-call-from-U-mode" exception if it is called in user mode and performs no other operation. So we can regard the `ECALL` as a special exception. Similarly, `EBREAK` generates a breakpoint exception and performs no other operation. It's usually used by a debugger. 

Essentially these two instructions only switch user mode to supervisor mode then do nothing further, this simple behavior leaves the operating system enough space to do whatever it wants, such as syscall or debug.

> ECALL and EBREAK cause the receiving privilege mode’s `epc` register to be set to the address of the ECALL or EBREAK instruction itself, not the address of the following instruction. 

Refer to above quote, the `epc` will be set to the address of `ECALL` itself, that's why in the [`usertrap()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L41), it runs [`tf.epc += 4;`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L66) before call the [`syscall()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/syscall/syscall.rs#L112).

### 3.2 Interrupts

To handle interrupt, the [`devintr`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L204) needs to be called:

```rust
// trap.rs
... ...
} else {
    which_dev = devintr();
    if which_dev != 0 {
        // ok
    } else {
        printf!(
            "usertrap(): unexpected scause {:x} pid={}\n",
            r_scause(),
            p.pid
        );
        printf!("            sepc={:x} stval={:x}\n", r_sepc(), r_stval());
        p.setkilled();
    }
}
... ...
```

Inside the [`devintr`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L204), there are still a few branches to differentiate more cases:

```rust
// trap.rs

// check if it's an external interrupt or software interrupt,
// and handle it.
// returns 2 if timer interrupt,
// 1 if other device,
// 0 if not recognized.
fn devintr() -> u8 {
    let scause = r_scause();

    if (scause & 0x8000000000000000) != 0 && (scause & 0xff) == 9 {
        // this is a supervisor external interrupt, via PLIC.

        // irq indicates which device interrupted.
        let irq = plic_claim();

        if irq == UART0_IRQ as u32 {
            unsafe {
                UART_INSTANCE.intr();
            }
        } else if irq == VIRTIO0_IRQ as u32 {
            unsafe {
                virtio_disk_intr();
            }
        } else if irq != 0 {
            printf!("unexpected interrupt irq={}\n", irq);
        }

        // the PLIC allows each device to raise at most one
        // interrupt at a time; tell the PLIC the device is
        // now allowed to interrupt again.
        if irq != 0 {
            plic_complete(irq);
        }

        return 1;
    }

    if scause == 0x8000000000000001 {
        // software interrupt from a machine-mode timer interrupt,
        // forwarded by timervec in kernelvec.S.

        if cpuid() == 0 {
            clockintr();
        }

        // acknowledge the software interrupt by clearing
        // the SSIP bit in sip.
        w_sip(r_sip() & !2);

        return 2;
    }

    0
}
```

To understand the branches, we may need to investigate how many reasons the `scause` stands for:

| Interrupt | Exception Code | Description                    |
| :-------: | :------------: | ------------------------------ |
|     1     |       0        | Reserved                       |
|     1     |       1        | Supervisor software interrupt  |
|     1     |      2-4       | Reserved                       |
|     1     |       5        | Supervisor timer interrupt     |
|     1     |      6-8       | Reserved                       |
|     1     |       9        | Supervisor external interrupt  |
|     1     |     10-12      | Reserved                       |
|     1     |       13       | Counter-overflow interrupt     |
|     1     |     14-15      | Reserved                       |
|     0     |     >= 16      | Designated for platform use    |
|     0     |       0        | Instruction address misaligned |
|     0     |       1        | Instruction access fault       |
|     0     |       2        | Illegal instruction            |
|     0     |       3        | Breakpoint                     |
|     0     |       4        | Load address misaligned        |
|     0     |       5        | Load access fault              |
|     0     |       6        | Store/AMO address misaligned   |
|     0     |       7        | Store/AMO access fault         |
|     0     |       8        | Environment call from U-mode   |
|     0     |       9        | Environment call from S-mode   |
|     0     |     10-11      | Reserved                       |
|     0     |       12       | Instruction page fault         |
|     0     |       13       | Load page fault                |
|     0     |       14       | Reserved                       |
|     0     |       15       | Store/AMO page fault           |
|     0     |     16-17      | Reserved                       |
|     0     |       18       | Software check                 |
|     0     |       19       | Hardware error                 |
|     0     |     20-23      | Designated for custom use      |
|     0     |     24-31      | Designated for custom use      |
|           |     32-47      | Reserved                       |
|           |     48-63      | Designated for custom use      |
|           |     >= 64      | Reserved                       |

It looks quite complicated, but in this stage we only need to care two scenarios:

1. `(scause & 0x8000000000000000) != 0 && (scause & 0xff) == 9`

   Above condition filters out the external interrupt so that once the program goes into this branch, it means there was an external interrupt triggered by some device.

   In current xv6 source code there are only UART and VIRTIO supported, but it won't be too hard to add other devices to the system. In real operating system the interrupt handlers for specific devices are often put in some software package called driver. 

   After handles the external interrupt, `plic_complete(irq)` should be called to reset the PLIC so that new interrupts can be triggered again, otherwise the PLIC will keep waiting for handler to do its work.

2. `scause == 0x8000000000000001`

   This condition only filters software interrupt. But why does it handle the timer interrupt? 

   In the [`timerinit()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/start.rs#L57), the timer interrupt handler is set to [`timervec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/kernelvec.S#L95), hence, once timer ticked, only [`timervec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/kernelvec.S#L95) can handle the interrupt. And if you see it closely, the [`timervec`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/kernelvec.S#L95) triggers a software interrupt in the end. The user trap handler wouldn't handle the timer interrupt until now.

   Why bother with all this? Because in previous versions of risc-v, the time compare registers can only be accessed in machine mode, so that there's no way for kernel to reset the timer registers in supervisor mode.

   Refer to the xv6 book version 3 (page 56):

   > A timer interrupt can occur at any point when user or kernel code is executing; there’s no way for the kernel to disable timer interrupts during critical operations. Thus the timer interrupt handler must do its job in a way guaranteed not to disturb interrupted kernel code. The basic strategy is for the handler to ask the RISC-V to raise a “software interrupt” and immediately return. The RISC-V delivers software interrupts to the kernel with the ordinary trap mechanism, and allows the kernel to disable them.

   Fortunately, risc-v now supports the "SSTC" extension. [Here](https://drive.google.com/file/d/1O0ogDHijAc7gM58Byb0BRqIRGYsdOt2D/view) is the documentation about the "SSTC", which is ratified in 2021. The SSTC extension *"provide supervisor mode with its own CSR-based timer interrupt facility that it can directly manage to provide its own timer service."* 

   [QEMU has supported this extension](https://lists.gnu.org/archive/html/qemu-riscv/2022-05/msg00063.html) back to 2022. And the newer version of xv6 has changed to SSTC, see [here](https://github.com/mit-pdos/xv6-riscv/blob/de247db5e6384b138f270e0a7c745989b5a9c23b/kernel/trap.c#L210).

### 3.3 Exceptions

The exceptions are mainly related to the `scause`. And as long as the returns 0, means an exception happened:

```rust
// trap.rs
... ...
if which_dev == 0 {
    printf!("scause {:x}\n", scause);
    printf!("sepc={:x} stval={:x}\n", r_sepc(), r_stval());
    panic!("kerneltrap");
}
... ...
```

The handling of exceptions is quite simple and straightforward: it panics. The value of `scause`, `sepc` and `stval` will be printed along with panic. Those values are really useful to help investigate the root cause of exceptions. The `scause` records the exception reason, the `sepc` holds the virtual address of instruction that cause trap, while the `stval` is written to different useful information based on different value of `scause`.

We have reviewed the detail of exception types before, and the followings are what kinds of value will be written into the `stval`:

| Exceptions                                                   | Value of `stval`                                             |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| Breakpoint, address-misaligned, access-fault, page-fault     | Faulting virtual address                                     |
| Access-fault or page-fault caused by misaligned load or store | The virtual address of the portion of the access that caused the fault. |
| Instruction access-fault or page-fault                       | The virtual address of the portion of the instruction that caused the fault |
| Illegal instruction exception                                | Faulting instruction bits                                    |
| Other traps                                                  | Zero                                                         |

These three registers are very helpful when kernel crashes. Especially in the debugging process of xv6, it would be common to encounter the stack overflow problem, since there is always a "guard" area between two stacks, as long as stack overflows, there would be a access fault exception thrown, in this moment, checking the value of `sepc` and `stval` will help to find the code position.

## 4. Init Process

With all previous content as the foundation, now we can finally discover how the xv6 starts its first process: init process.

The following diagram shows the main sequence of kernel builds up init process and then executes it as the first user process:

{% asset_img 6.png %}

Some parts like switching between user mode and supervisor mode has been covered before, next we are going to focus on the other parts.

### 4.1 User Init

First, let's see how the init process is created:

```rust
// proc.rs

const INIT_CODE: [u8; 52] = [
    0x17, 0x05, 0x00, 0x00, 0x13, 0x05, 0x45, 0x02, 0x97, 0x05, 0x00, 0x00, 0x93, 0x85, 0x35, 0x02,
    0x93, 0x08, 0x70, 0x00, 0x73, 0x00, 0x00, 0x00, 0x93, 0x08, 0x20, 0x00, 0x73, 0x00, 0x00, 0x00,
    0xef, 0xf0, 0x9f, 0xff, 0x2f, 0x69, 0x6e, 0x69, 0x74, 0x00, 0x00, 0x24, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00,
];

// Set up first user process.
pub fn userinit() {
    let p = allocproc().unwrap();
    // allocate one user page and copy initcode's instructions
    // and data into it.
    uvmfirst(
        unsafe { p.pagetable.unwrap().as_mut().unwrap() },
        &INIT_CODE as *const u8,
        mem::size_of_val(&INIT_CODE),
    );
    p.sz = PGSIZE;

    // prepare for the very first "return" from kernel to user.
    unsafe {
        p.trapframe.unwrap().as_mut().unwrap().epc = 0; // user program counter
        p.trapframe.unwrap().as_mut().unwrap().sp = PGSIZE as u64; // user stack pointer
    }

    let mut name = [0; 16];
    name.copy_from_slice("initcode\0\0\0\0\0\0\0\0".as_bytes());
    p.name = name;
    p.cwd = namei(&[b'/']).map(|inner| inner as *mut INode);

    p.state = RUNNABLE;

    p.lock.release();

    unsafe {
        INIT_PROC = Some(p);
    }
}
```

The [`userinit()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L307) function does three important things:

1. Allocate a process structure that holds stack, page table and trap frame. The allocation mainly covered by the [`allocproc()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L446), which picks an unused `Proc` structure to hold init.
2. Load code into memory at user space, which is performed by the [`uvmfirst()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/vm.rs#L273) that maps the [`INIT_CODE`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L299) into page table as the text section.
3. Set entry point as 0 in virtual address, and then set status to `RUNNABLE`.

Looking at the [`INIT_CODE`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L299) you'll find there are only binaries. In fact, these binaries come from the compile result of [`initcode.S`](https://github.com/LENSHOOD/xv6-rust/blob/master/user/initcode/initcode.S):

```assembly
### user/initcode/initcode.S

# Initial process that execs /init.
# This code runs in user space.

# exec(init, argv)
.globl start
start:
        la a0, init
        la a1, argv
        li a7, 7 # SYS_exec
        ecall

# for(;;) exit();
exit:
        li a7, 2 # SYS_exit
        ecall
        jal exit

# char init[] = "/init\0";
init:
  .string "/init\0"

# char *argv[] = { init, 0 };
.p2align 2
argv:
  .long init
  .long 0
```

The above code is quite simple, it only calls `SYS_exec` syscall with `/init\0` string as the argument. Since the compiled binaries are very simple the xv6 can even hardcoded it as a constant, this will omit the step to load it from file system.

Apparently, the [`INIT_CODE`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L299) is like a step-stone for initialization, the only thing it performs is to call `SYS_exec` syscall to replace its code text as the program `/init`. I'm sure you already knew the responsibility of the `exec` syscall in POSIX, the `SYS_exec` is just like that. Will talk the `/init` soon later.

Besides, if you look into the implementation of the [`uvmfirst()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/vm.rs#L273), it loads the [`INIT_CODE`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L299) at virtual address 0x0, and since the `trapframe.epc` is also set to 0, once the init process is put on cpu, the first line of code it would run is `start` in the [`INIT_CODE`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L299).

### 4.2 Running On CPU

Once the init process is well prepared, then the kernel continues its init procedure, and runs the last step: [`scheduler()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/main.rs#L164).

We have learnt in the fourth chapter what the scheduler will do: traverse the proc list to find if there is any process is on `RUNNABLE` state. Here we only have one process that is runnable: init process. So the scheduler will call [`swtch`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/asm/switch.S#L9) to switch the context and return. But where will it be returned?

If you remember, we have mentioned this back in chapter-4, the return address is set in the [`inner_alloc()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L460):

```rust
//proc.rs

fn inner_alloc<'a>(p: &'a mut Proc<'a>) -> Option<&'a mut Proc<'a>> {
    ... ...
    p.context.ra = forkret as u64;
    ... ...
}
```

Since the [`scheduler()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/main.rs#L164) is running on kernel space, we can't return to user space directly after switch, that what the [`forkret`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L426) is responsible:

```rust
const FIRST: AtomicBool = AtomicBool::new(true);
fn forkret() {
    // Still holding p->lock from scheduler.
    let my_proc = myproc();
    my_proc.lock.release();

    if FIRST.load(Ordering::Relaxed) {
        // File system initialization must be run in the context of a
        // regular process (e.g., because it calls sleep), and thus cannot
        // be run from main().
        FIRST.store(false, Ordering::Relaxed);
        fs::fsinit(ROOTDEV);
    }

    usertrapret();
}
```

It's simple and clear, at the first time the [`forkret`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L426) is called, the file system needs to be initialized, this step mainly reads the `SuperBlock` from disk, please refer to chapter-5 for more details. 

After that, [`usertrapret()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L103) will be called, we have seen that in previous sections, at this very step, the `epc` will be set to `trapframe.epc`, which is 0, and once `SRET` is called, the init process will finally start to run.

### 4.3 Syscall Exec

Following the previous sequence diagram, the [`INIT_CODE`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L299) only execute `ecall` to get in trap again, and trigger the `SYS_exec`.

As we already knew how syscall is handled, let's go checking the implementation of [`sys_exec()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/syscall/sysfile.rs#L18):

```rust
pub(crate) fn sys_exec() -> u64 {
    ... ...
    let mut ret = -1;
    if !bad {
        ret = exec(path, &argv);
    }

    ... ...

    return ret as u64;
}
```

Most of its jobs are fetch the argument along with `ecall`, it will need some effort to do so is because the argument is at user space and needs to be copy into kernel space. But these parts are not very important, now we deep dive into the [`exec()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/exec.rs#L25):

```rust
pub fn exec(path: [u8; MAXPATH], argv: &[Option<*mut u8>; MAXARG]) -> i32 {
    // find the inode of /init
    let ip_op = namei(&path);
    ... ...

    ip.ilock();

    // Check ELF header and copy the text section into memory
    let mut elf = ElfHeader::create();
    let tot = ip.readi(false, &mut elf, 0, mem::size_of::<ElfHeader>());
    ... ...

    for _i in 0..elf.phnum {
        let tot = ip.readi(false, &mut ph, off, ph_sz);
        ... ...
        if loadseg(page_table, ph.vaddr, ip, ph.off, ph.filesz) < 0 {
            return goto_bad(Some(page_table), sz, Some(ip));
        }
    }
    ... ...
    let p = myproc();
    let oldsz = p.sz;
    ... ...
    // Push argument strings, prepare rest of stack in ustack.
    loop {
        ... ...
        if copyout(page_table, sp, curr_argv, strlen(curr_argv) + 1) < 0 {
            return goto_bad(Some(page_table), sz, Some(ip));
        }
        ... ...
    }

    ... ...

    // Commit to the user image.
    let oldpagetable = unsafe { p.pagetable.unwrap().as_mut().unwrap() };
    p.pagetable = Some(page_table as *mut PageTable);
    p.sz = sz;
    tf.epc = elf.entry; // initial program counter = main
    tf.sp = sp as u64; // initial stack pointer
    proc_freepagetable(oldpagetable, oldsz);

    return argc as i32; // this ends up in a0, the first argument to main(argc, argv)
}
```

Since this function is very long, about code pieces only contain a few main steps. For more information please directly see the raw code.

In short, the final target of the [`exec()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/exec.rs#L25) is replacing the current process into a new one, but keeps the current process structure and pid. Through this procedure, many things will get replaced, such as program code, constants, variables.

In order to achieve that, first we need to load the `/init` into the memory, that's what the `namei(&path)` does. Once we have the inode points to `/init` in our hand, we can read the content of `/init`.

In the above code, it first reads the ELF header as the xv6 file follows the ELF format. The ELF header records the section information, including program segments that are described in the program header. These segments are text, data and others, which are necessary for the program to be executed. 

Through the `loadseg()` function, all program segments are loaded into memory and mapped into a newly create page table, this page table will be the new page table that replaces the old one. Expect for the program segment loading, the arguments passed along with the `SYS_exec` are also copied into stack.

At last, the page table is replaced, and the old one is released, the `epc` is set to `elf.entry` which points to the first line of code of `/init`. At this moment, a new process is finally born.

### 4.4 Init

We haven't seen how the init looks like, at the end of this article, let's have a look:

```rust
#[start]
fn main(_argc: isize, _argv: *const *const u8) -> isize {
    unsafe {
        // let mut console_slice: [u8; MAXPATH] = [b'\0'; MAXPATH];
        // console_slice.copy_from_slice("console".as_bytes());
        let console_slice = "console\0".as_bytes();
        if open(console_slice.as_ptr(), O_RDWR) < 0 {
            mknod(console_slice.as_ptr(), CONSOLE as u16, 0);
            open(console_slice.as_ptr(), O_RDWR);
        }
        dup(0); // stdout fd=1
        dup(0); // stderr fd=2

        let mut pid;
        let mut wpid;
        loop {
            printf!("init: starting sh\n");
            pid = fork();
            if pid < 0 {
                printf!("init: fork failed\n");
                exit(1);
            }
            if pid == 0 {
                let argv: *const *const u8 =
                    (&["sh\0".as_bytes().as_ptr(), "".as_bytes().as_ptr()]).as_ptr();
                exec("sh\0".as_bytes().as_ptr(), argv);
                printf!("init: exec sh failed\n");
                exit(1);
            }

            loop {
                // this call to wait() returns if the shell exits,
                // or if a parentless process exits.
                wpid = wait(0 as *const u8);
                if wpid == pid {
                    // the shell exited; restart it.
                    break;
                } else if wpid < 0 {
                    printf!("init: wait returned an error\n");
                    exit(1);
                } else {
                    // it was a parentless process; do nothing.
                }
            }
        }
    }
}

```

There isn't too much code in it, the whole logic can be split into two simple parts:

- Open Console as stdin, stdout and stderr
- Fork itself
  - For the child process, call `SYS_exec` again to replace itself as the shell program
  - For the parent process, which is also the init process it self, wait for it child to be exited, and if it happens that the child process also has its children, then the grandchildren processes will be regarded as orphans. Therefore, along with the exit of the child process, all orphan processes will be reparented to init. (See [`exit()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/proc.rs#L704) for details)

 





