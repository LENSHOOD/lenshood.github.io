---
title: Why Assign Final Field to Local Variable?
date: 2021-06-21 21:48:13
tags:
- java
- performance optimization
categories:
- Java
---

Recently there's a friend ask a question in a tech group chat, he said that: 

> In the implementation of `CopyOnWriteArrayList.add(E e)`,  why the writer assign the final field `lock` to a local variable ?

Then he posted a picture like this:

{% asset_img pic-1.png %}

When I open my local JDK source and get `CopyOnWriteArrayList.add(E e)`, I found that the implementation of `add(E e)` in my version of JDK (jdk-15) has already refactored to just use `synchronized` key word (since now the performance is better than `ReentrantLock`) .

Actually the picture's version of `CopyOnWriteArrayList.add(E e)` is contained in JDK 1.8, so I switch my jdk version, and found the code, then I fell into thought...

<!-- more -->

### It's useless?

Why Doug Lea(the code writer) did like that? It make no sense!

1. The `lock` field is defined as `final`, no one can change it
2. Won't it be optimized by compiler?

After some Google, there's one guy said at [StackOverflow](https://stackoverflow.com/questions/2785964/in-arrayblockingqueue-why-copy-final-member-field-into-local-final-variable):

{% asset_img pic-2.png %}

And open the [thread](http://mail.openjdk.java.net/pipermail/core-libs-dev/2010-May/004165.html), we can see it says it's an "extreme optimization" and can make the compiler to "produces the smallest bytecode".

WOW, That's amazing! I never thought that would come!

So now I wander: it that real?



### Let's Find Out

According to the content of that thread post, the optimization is act on bytecode even machine code. So I wrote such simplified test code to simulate the circumstance:

```java
package lenshood.demo;

import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;

public class FinalTest {
    private final Lock fLock = new ReentrantLock();
    private int i;

    public void useLocal() {
        final Lock lLock = this.fLock;

        lLock.lock();
        try {
            i++;
        } finally {
            lLock.unlock();
        }
    }

    public void useField() {
        fLock.lock();
        try {
            i++;
        } finally {
            fLock.unlock();
        }
    }

    public static void main(String[] args) {
        FinalTest finalTest = new FinalTest();
        for (int i = 0; i < 10_000_000; i++) {
            finalTest.useLocal();
            finalTest.useField();
        }

        System.out.println(finalTest.i);
    }
}
```

There's two different methods to demonstrate the two coding style of use local variable or directly use final field.

And let's see the bytecode of the two methods:

```bytecode
### useLocal()
 0 aload_0
##############
 1 getfield #10 <lenshood/demo/FinalTest.fLock>
############## 
 4 astore_1
 5 aload_1
 6 invokeinterface #16 <java/util/concurrent/locks/Lock.lock> count 1
11 aload_0
12 dup
13 getfield #21 <lenshood/demo/FinalTest.i>
16 iconst_1
17 iadd
18 putfield #21 <lenshood/demo/FinalTest.i>
21 aload_1
22 invokeinterface #25 <java/util/concurrent/locks/Lock.unlock> count 1
27 goto 39 (+12)
30 astore_2
31 aload_1
32 invokeinterface #25 <java/util/concurrent/locks/Lock.unlock> count 1
37 aload_2
38 athrow
39 return

-------------------------------------------------------------------------------

### useField()
 0 aload_0
##############
 1 getfield #10 <zxh/demo/FinalTest.fLock>
##############
 4 invokeinterface #16 <java/util/concurrent/locks/Lock.lock> count 1
 9 aload_0
10 dup
11 getfield #21 <zxh/demo/FinalTest.i>
14 iconst_1
15 iadd
16 putfield #21 <zxh/demo/FinalTest.i>
19 aload_0
##############
20 getfield #10 <zxh/demo/FinalTest.fLock>
##############
23 invokeinterface #25 <java/util/concurrent/locks/Lock.unlock> count 1
28 goto 43 (+15)
31 astore_1
32 aload_0
33 getfield #10 <zxh/demo/FinalTest.fLock>
36 invokeinterface #25 <java/util/concurrent/locks/Lock.unlock> count 1
41 aload_1
42 athrow
43 return
```

Compare the two copy of bytecodes, it's obvious to find that:

1. In the `useLocal()`, there's one "getfield" and one "astore_1" + two "aload_1" to assign/load local variables from final field "fLock".
2. In the `useField()`, there's two "getfield".

Hence, we found the bytecodes do have difference, but why `1*getfiled + 1*astore + 2*aload` is better than `2*getfield` ?

Here is some possible hypotheses:

- Local variable can store at registers, but field can only get from memory, which is slower
- Final field has the semantics of `happens-before`, and JVM may insert load barriers before get final field

But how to prove them? We better go deeper: pass through bytecode and go to asm!



### Get ASM from JIT

Firstly we may need to install a plugin for HotSpot VM to do disassembling.

`hsdis` is contained in the jdk source code, we can find it from openjdk at GitHub.

To jdk-15, the `hsdis` is located in: `src/utils/hsdis`

##### Install `hsdis` to MacOS (for JDK-15)

1. `binutils` is needed:
   - Download `binutils` from: https://www.gnu.org/software/binutils/
   - `tar -xvf binutils-xxx.tar.bz2`
2. Get `hsdis` source, then build it
   - Assume we're in the `hsdis` dir, put `binutils` we just downloaded in it.
   - `make BINUTILS=binutils-xxx ARCH=amd64`

3. Put plugin into jdk
   - `sudo cp build/macosx-amd64/hsdis-amd64.dylib $JAVA_HOME/lib/server`

##### Get ASM

1. `javac FinalTest.java`
2. `java -Xbatch -XX:-TieredCompilation -XX:+UnlockDiagnosticVMOptions -XX:+PrintAssembly`

Then we can get ASM code output to shell, select the section related to our methods:

```assembly
############ useLocal() ############
[Entry Point]
  # {method} {0x000000011c4003f0} 'useLocal' '()V' in 'FinalTest'
  #           [sp+0x40]  (sp of caller)
  
  … …

  0x0000000117530070:   mov    0x10(%rsi),%r11d             ;*getfield fLock {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - FinalTest::useLocal@1 (line 9)
  0x0000000117530074:   mov    0x8(%r12,%r11,8),%r10d       ; implicit exception: dispatches to 0x0000000117530330
  0x0000000117530079:   nopl   0x0(%rax)
  0x0000000117530080:   cmp    $0x3446b,%r10d               ;   {metadata('java/util/concurrent/locks/ReentrantLock')}
  0x0000000117530087:   jne    0x00000001175302a0
  0x000000011753008d:   lea    (%r12,%r11,8),%rbx           ;*invokeinterface lock {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - FinalTest::useLocal@6 (line 11)
  0x0000000117530091:   mov    0xc(%rbx),%r14d              ;*getfield sync {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - java.util.concurrent.locks.ReentrantLock::lock@1 (line 322)
                                                            ; - FinalTest::useLocal@6 (line 11)
  … …

  0x0000000117530110:   incl   0xc(%r10)                    ;*putfield i {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - FinalTest::useLocal@18 (line 13)
  0x0000000117530114:   mov    0xc(%rbx),%ebx               ;*getfield sync {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - java.util.concurrent.locks.ReentrantLock::unlock@1 (line 494)
                                                            ; - FinalTest::useLocal@22 (line 15)
  … …
  
```

In the ASM of `useLocal()`, we can simply find it first get the final field `fLock` and put it to `r11` register as a local variable (`0x0000000117530070`), after that, when find the `ReentrantLock` instant fron dynamic table ( 0x000000011753008d  ), the code directly use `r11` to get the address.

Down to `0x0000000117530110` we know it's the `i++` operation, and then at the next address `0x0000000117530114` -- when do `unlock()` -- the `sync` field (inner field in `ReentrantLock`) are just addressing from `rbx`, which contained calculated result from `r11` (0x000000011753008d).

```assembly
############ useFeild() ############
[Entry Point]
  # {method} {0x000000011c4004e0} 'useField' '()V' in 'FinalTest'
  #           [sp+0x40]  (sp of caller)

  … …

  0x000000011752f5ef:   mov    0x10(%rsi),%r10d             ;*getfield fLock {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - FinalTest::useField@1 (line 20)
  0x000000011752f5f3:   mov    0x8(%r12,%r10,8),%r8d        ; implicit exception: dispatches to 0x000000011752f8d0
  0x000000011752f5f8:   nopl   0x0(%rax,%rax,1)
  0x000000011752f600:   cmp    $0x3446b,%r8d                ;   {metadata('java/util/concurrent/locks/ReentrantLock')}
  0x000000011752f607:   jne    0x000000011752f834
  0x000000011752f60d:   shl    $0x3,%r10                    ;*invokeinterface lock {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - FinalTest::useField@4 (line 20)
  0x000000011752f611:   mov    0xc(%r10),%r13d              ;*getfield sync {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - java.util.concurrent.locks.ReentrantLock::lock@1 (line 322)
                                                            ; - FinalTest::useField@4 (line 20)
  
  … …

  0x000000011752f68c:   incl   0xc(%rbx)                    ;*putfield i {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - FinalTest::useField@16 (line 22)
  0x000000011752f68f:   mov    0x10(%rbx),%ebp              ;*getfield fLock {reexecute=0 rethrow=0 return_oop=0}
                                                            ; - FinalTest::useField@20 (line 24)
```

The `useField()` is even simplier, at `0x000000011752f5ef` and `0x000000011752f68f`, it just read `fLock` twice from memory.

### So the Performance indeed better
Go back to our two hypotheses:
1. Register: yes, it use register to hold local variable and avoid twice load from memory(cache)
2. Load barrier: there's no explicit barriers we can find, however, due to the [strong memory model of x86](https://www.asrivas.me/blog/memory-barriers-on-x86/)([TSO](https://www.cl.cam.ac.uk/~pes20/weakmemory/cacm.pdf)), `mov` already implied the LoadLoad barrier semantics.

### Conclusion
After our study, now we can explain why assign a final field to local variable can get better performance, we can also know that why it's an "extreme optimization".

Hence, put this optimization to everywhere maybe not a good idea, but the spirit of pursue the ultimate performance it's really admirable.

## Reference

1. [In ArrayBlockingQueue, why copy final member field into local final variable?](https://stackoverflow.com/questions/2785964/in-arrayblockingqueue-why-copy-final-member-field-into-local-final-variable)
2. [Performance of locally copied members ?](http://mail.openjdk.java.net/pipermail/core-libs-dev/2010-May/004165.html)
3. [How to Show the Assembly Code Generated by the JVM](https://www.beyondjava.net/show-assembly-code-generated-jvm)
4. [Building hsdis for OpenJDK 15](https://www.morling.dev/blog/building-hsdis-for-openjdk-15/)
5. [Memory Barriers on x86](https://www.asrivas.me/blog/memory-barriers-on-x86/)
