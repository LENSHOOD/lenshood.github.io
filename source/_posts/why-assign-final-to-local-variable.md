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

{% asset_img pic-1.png %}

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

Here is some possible hypothesis:

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





## Reference

1. [In ArrayBlockingQueue, why copy final member field into local final variable?](https://stackoverflow.com/questions/2785964/in-arrayblockingqueue-why-copy-final-member-field-into-local-final-variable)
2. [Performance of locally copied members ?](http://mail.openjdk.java.net/pipermail/core-libs-dev/2010-May/004165.html)
3. [How to Show the Assembly Code Generated by the JVM](https://www.beyondjava.net/show-assembly-code-generated-jvm)
4. [Building hsdis for OpenJDK 15](https://www.morling.dev/blog/building-hsdis-for-openjdk-15/)