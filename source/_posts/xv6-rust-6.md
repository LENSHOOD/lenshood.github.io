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

After save the user program counter into trap frame, in the next it mainly dealing with the three trap reasons: syscalls, interrupts and exceptions. We'll cover these parts in next section, now we are going to the final line directly: [`usertrapret()`](https://github.com/LENSHOOD/xv6-rust/blob/5654d2a13560a47a5aa5505a0a9fd36bdf0274cf/kernel/src/trap.rs#L103):

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



## 3. Syscalls



## 4. Init Process
