---
title: title: /Xv6 Rust 0x04/ - Easy Piece of Virtualization (virtual cpu)
date: 2024-11-20 22:50:32
tags:
- xv6
- rust
- os
categories:
- Rust
---

In this chapter, we are going to explore the cpu virtualization, also know as process, in the xv6.

I'm really excited for writing this chapter, because process is one of the fundamental concepts of the operating system. Process is very useful for multiple tasks, and in the design wise, its abstraction is also very elegant.

<!-- more -->

## 1. Overview of Virtual CPU

We all known that with the concept if process, we could run multiple programs on one or a few cpu cores, that makes the process es and the cpu cores present as a many-to-many mappings.

But the key point here is, a process will not need to know how many cpu cores and how many memory it can have. Through some sort of abstraction, a process can freely using all cpu and memory resources to running its program, any resource management should have done by the kernel.

So before we take a look at the design of xv6 process, let's think about what elements a process should have, to ensure the cpu virtualization works.

We can simply recap the machine code we introduced in the first two chapters as an analogy, when we built the "xv6-rust-sample", there are few things that need to be taken care of:

- We should understand the address space then link the program to right places
- There should be a way to load the program and running from the entry point
- Some necessary registers should be initialized, such as stack pointer
- The error handling is also required, like panic

  Additionally, stand in the kernel's shoes, more points turn out:

- There are more processes than cpu cores, how to switch different processes on one cpu?
- Where to save process status when it's switched?
- How to manage the process lifecycle, including create, destroy, and error handling?
- How to allocate and restore memory to and from processes

Actually, if a kernel implements all above points, then it has all elements to run multiple processes. So let's take a first look at the design of xv6 process (you can also check the real code at [here](https://github.com/LENSHOOD/xv6-rust/blob/569774eeff135ebc877bd25a4b283d75ad62d35c/kernel/src/proc.rs#L168)):

{% asset_img 1.png %}

For the elements contained inside the process: 

The "PID / Name" identifies a specific process.

The "State" field records the current state of the process, the common state are Running / Runnable / Sleeping, which indicates running on cpu, waiting to be scheduled, waiting to be waken up respectively.

The "Open Files" tracks any files that opened by the process, we haven't talked about file system before, but at least we can realize the basic three files that every process will open are STDIN, STDOUT and STDERR.

The "Parent" to track the process parent, like linux, the xv6 also create a new process by `fork()`, therefore, every process should have a parent process.

The "Kernel Stack" allows running kernel code on the address space of a process. After all, every process needs to interact with kernel through different types of syscalls, for safety purpose, user process cannot share a same stack with kernel.

The "Trap Frame" stores user space process data, this kind of data will be saved and restored when switching between user space and kernel space. We'll cover this part in the following chapter about interrupt and syscall.

The "Page Table" records the mapping between virtual memory and physical memory. We have described virtual memory in the previous chapter, actually every process should have its own page table.

The "Context" records the basic registers a process is using. When a process needs to be paused and run another process, the current states in the registers should be saved, and once the process can re-run, the registers should be restored.



## 2. Process Memory

The previous code `xv6-rust-sample` set it address space starts from `0x80000000`, and put its text section at the beginning of the address, then set the code entry at there. So where should a process starts from?

We could refer the process creation syscall [`exec()`](https://github.com/LENSHOOD/xv6-rust/blob/569774eeff135ebc877bd25a4b283d75ad62d35c/kernel/src/exec.rs#L61) to find out (again, we'll cover this in a later chapter):

```rust
// exec.rs
pub fn exec(path: [u8; MAXPATH], argv: &[Option<*mut u8>; MAXARG]) -> i32 {
  ... ...
	for _i in 0..elf.phnum {
        // load program from file system
      	let tot = ip.readi(false, &mut ph, off, ph_sz);
        ... ...
        // allocate enough space and map into process page table
        let sz1 = uvmalloc(
            page_table,
            sz,
            (ph.vaddr + ph.memsz) as usize,
            flags2perm(ph.flags),
        );
        ... ...
    		// the program will be loaded into virtual address ph.vaddr
        if loadseg(page_table, ph.vaddr, ip, ph.off, ph.filesz) < 0 {
            ... ...
        }
    }
    ... ...
    // the entry point will be set into "sepc" csr later
    tf.epc = elf.entry; // initial program counter = main
}
```

From above code, we realized the program will be loaded into `ph.vaddr`, actually `ph` means "ProgramHeader", which is also ELF format, therefore, to find the real entry point, we have to check the [linker script](https://github.com/LENSHOOD/xv6-rust/blob/569774eeff135ebc877bd25a4b283d75ad62d35c/user/src/ld/user.ld#L7) of user program:

```ld
// user/src/ld/user.ld
OUTPUT_ARCH( "riscv" )
ENTRY( main )

SECTIONS
{
 . = 0x0;
 
 ... ...

 PROVIDE(end = .);
}
```

Obviously, a user program starts from `main`, and the entry address is `0x0`. Of cause we can change that to any address, because the entry address will be set into the CSR "sepc", and once the CPU switch to user mode, the user program will start from there.

Besides, like we mentioned before, every processes need a dedicated stack that allows kernel [code](https://github.com/LENSHOOD/xv6-rust/blob/569774eeff135ebc877bd25a4b283d75ad62d35c/kernel/src/proc.rs#L272) running on it. The following code shows the initialization of every kernel stacks:

```rust
// proc.rs
pub fn proc_mapstacks(kpgtbl: &mut PageTable) {
    for idx in 0..NPROC {
        unsafe {
            let pa_0: *mut u8 = KMEM.kalloc();
            let pa_1: *mut u8 = KMEM.kalloc();
            let va = KSTACK!(idx);
            kvmmap(kpgtbl, va, pa_0 as usize, PGSIZE, PTE_R | PTE_W);
            kvmmap(kpgtbl, va + PGSIZE, pa_1 as usize, PGSIZE, PTE_R | PTE_W);
        }
    }
}

// memlayout.rs
#[macro_export]
macro_rules! KSTACK {
    ( $p:expr ) => {
        $crate::memlayout::TRAMPOLINE - (($p) + 1) * 3 * $crate::riscv::PGSIZE
    };
}
```

We have already introduced `proc_mapstacks()` in the previous chapter, but no details about it. Here, apparently this function allocates two pages for each process as their kernel stack.

But there are some interest things here:

1. In the original xv6 implementation with C language, one process only have one page of stack, however, in my rust version, many core lib functions were introduced, cause one page (4096 bytes) of stack is not enough, lead to risc-v throw an invalid access exception. Hence, the kernel stack is extended to two pages, and that's enough at least now. You may wondering, we have set nearly all address space available for RWX permission (refer to chapter 2, pmpaddr0 / pmpcfg0) how can it possible to throw an exception of no access permission? Let's move to the second interesting thing.
2.  The access exception relates to the macro `KSTACK!()`. If you see it carefully, you may find this macro actually makes each process has 3 pages of stack space, but we only allocate 2 pages for it, and leave the last page of stack space with no physical memory mapping to it. If any code allocate stack exceeded the 2 pages stack, the `sp` will point to the third stack page, which is invalid because of no mapped physical memory, then the exception is thrown. This kind of page called "guard page", it can prevent other process's stack from accidentally overwritten by an overflowed stack operation. (There's a defect here that if the applied stack space exceeded more than one page, then it can break the guard in some cases)

{% asset_img 2.png %}

Above image shows the location of the process stacks, it's worth noting that all of these stacks are allocated while kernel is starting, so they occupy physical memory all the time, on the contrary, if a user process needs some memory to store data, they can do that by calling syscall `mmap()`, which can dynamically allocate memory space.



## 3. Concurrency

Once the concept of multiple tasks is introduced,  the data contention issue will be definitely along with too. Additionally, although we only assign one cpu core in the QEMU runner, multiple cpu cores is also allowed for xv6 to run. Therefore, there should be some form of mechanisms to take care of the concurrency and parallelization.

The most fundamental thing of concurrency is lock. I suppose you have noticed that there were many code examples in the previous chapters contain locks, but we just ignored them and said we would cover them later. Here we are going to cover this part.

`SpinLock` is the simplest lock implementation, the definition is as follow:

```rust
pub struct Spinlock {
    locked: u64, // Is the lock held?
    name: &'static str,             // For debugging: Name of lock.
    cpu: Option<*mut Cpu<'static>>, // The cpu holding the lock.
}
```

From the definition, we can learn that it's basically a value holder, `locked == 1` indicates the lock is held, otherwise if `locked == 0` means the lock is released. Besides, there is a reference `cpu` points to current cpu, which also means the `SpinLock` is related to specific cpu.

So how does it work? Let's look closer:

```rust
impl Spinlock {
    pub fn acquire(self: &mut Self) {
        push_off(); // disable interrupts to afn deadlock.

        // On RISC-V, sync_lock_test_and_set turns into an atomic swap:
        //   a5 = 1
        //   s1 = &lk->locked
        //   amoswap.w.aq a5, a5, (s1)
        while __sync_lock_test_and_set(&mut self.locked, 1) != 0 {}
        __sync_synchronize();

        self.cpu = Some(mycpu());
    }

    pub fn release(self: &mut Self) {
        self.cpu = None;

        __sync_synchronize();
        // On RISC-V, sync_lock_release turns into an atomic swap:
        //   s1 = &lk->locked
        //   amoswap.w zero, zero, (s1)
        __sync_lock_release(&self.locked);

        pop_off();
    }

    /// Check whether this cpu is holding the lock.
    /// Interrupts must be off.
    pub fn holding(self: &Self) -> bool {
        self.locked == 1 && self.cpu == Some(mycpu())
    }
}
```

The `SpinLock` essentially using the atomic "test and swap" instruction `amoswap` provided by risc-v, and in the phase of acquire, since the lock might be held by another cpu, so it simply retry forever inside a loop to keep applying the lock (there is no logic such as wait in a queue to let it wait effectively, because `SpinLock` is the simplest implementation here). But when it comes to release the lock, since the lock has already held by current cpu, so try loop is no longer need. 

However, any unexpected interruption is able to break the lock, so we can noticed that once the lock is held, the cpu interrupt will  be disabled. Of cause this will significantly impact the performance, but after all, for a teaching OS, simplicity is more important than performance :) .

At last, the `__sync_synchronize()` internally call `fence` instruction, to keep the memory ordering before and after the "test and swap". And since the “AMO” instructions in risc-v by default not support any memory barrier(need to add extra "aq, rl" after the instruction), and to make sure the compiler can also realize the memory ordering, the `fence` is added before and after the "AMO" instructions to ensure the correct memory ordering in both cpu and compiler. The memory model of risc-v is complicated, here is a [good video](https://youtu.be/QkbWgCSAEoo?si=0cMjiypXe8iTUntZ) to talk about that.





## 4. Scheduling



## 5. Process Lifecycle

