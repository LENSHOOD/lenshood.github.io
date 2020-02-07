---
title:  Java 基于 Redis 实现分布式锁需要注意什么？
date: 2020-02-04 23:13:16
tags: 
- distributed lock
- redis
categories:
- Java
---

在如今这样一个张口分布式，闭口微服务的软件开发趋势下，多实例似乎已经不是某种选择而是一个无需多说的基本技术要求了。

多实例为我们带来稳定性提升的同时，也伴随着更复杂的技术要求，原先在本地即可处理的问题，全部扩展为分布式问题，其中就包含我们今天会聊到的多实例同步即分布式锁问题。JDK 提供的锁实现已经能够非常好的解决本地同步问题，而扩展到多实例环境下，Redis、ZooKeeper 等优秀的实现也使得我们使用分布式锁变得更加简单。

其实对于分布式锁的原理、分布式锁的 Redis 实现、ZK 实现等等各类文章不计其数，然而只要简单一搜就会发现，大多数文章都在教大家 Redis 分布式锁的原理和实现方法，但却没有几篇会写什么实现是好的实现，是适合用于生产环境，高效而考虑全面的实现。这将是本文讨论的内容。

### 分布式锁的要求
本节简单阐述分布式锁的基本要求，通常满足下述要求便可以说是比较完整的实现了。

1. 操作原子性
    - 与本地锁一样，加锁的过程必须保证原子性，否则失去锁的意义
    - Redis 的单线程模型帮我们解决了大部分原子性的问题，但仍然要考虑客户端代码的原子性
2. 可重入性
    - 分布式锁一样要考虑可重入的问题
    - Redis 通常能解决实例间的可重入问题，那么实例内线程间的可重入怎么办？
3. 效率
    - Redis 作为通过 TCP 通信的外部服务，网络延迟不可避免，因此相比本地锁操作时间更久
    - 分布式锁获取失败的通常做法是线程休眠一段时间
    - 如何才能尽可能减少不必要的通信与休眠？

### Local + Remote 结合实现分布式锁
正如上一节所述，采用 Redis，我们能很好的实现实例间的原子性（单线程模型），可重入性（各实例分配 UUID）。

而 JDK 的本地锁（如 ReentrantLock）又能非常完善的解决线程间同步的原子性、可重入性。

此外，对于实例内不同线程间的同步，JDK 通过 AQS 中一系列的方法确保高效稳定，因此省去了与 Redis 通信的消耗。

综上，如果将本地锁与远程锁结合在一起，便可以分别实现分布式锁在实例内与实例间的各项要求了。

### 代码实现
> 下文代码中，本地锁使用 ReentrantLock， Redis client 使用 Jedis。如替换其他方案，按照流程也很简单。

#### 整体架构
1. 初始化锁
```txet
.
└── 初始化锁
    └── new instance       
```      

2. 获取锁
```txet
.
└── 获取锁
    └── 尝试获取本地锁
        ├── 成功
        │   └── 尝试获取远程锁
        │       ├── 成功
        │       │   └── 加锁完成
        │       ├── 失败
        │       │   └── 轮询远程锁
        │       └── 超时
        │           ├── 释放本地锁
        │           └── 退出
        ├── 失败
        │   └── 阻塞等待
        └── 超时
            └── 退出
```

3. 释放锁
```txet
.
└── 释放锁
    ├── 当前线程持有本地锁？
    │   ├── 是重入状态？（hold count > 1）
    │   │   └── 释放本地锁
    │   └── 非重入状态
    │       ├── 释放远程锁
    │       └── 释放本地锁
    └── 未持有本地锁
        └── 无法释放，抛出错误
```

#### 代码框架
```java
public class RedisDistributedLock implements Lock {
    private static final String OBTAIN_LOCK_SCRIPT = ...
    private final ReentrantLock localLock = new ReentrantLock();

    @Override
    public void lock() {
        ...
    }

    @Override
    public void lockInterruptibly() throws InterruptedException {
        throw new UnsupportedOperationException();
    }

    @Override
    public boolean tryLock() {
        throw new UnsupportedOperationException();
    }

    @Override
    public boolean tryLock(long time, TimeUnit unit) throws InterruptedException {
        ...
    }

    @Override
    public void unlock() {
        ...
    }

    @Override
    public Condition newCondition() {
        throw new UnsupportedOperationException();
    }
}

```
由上述代码可见，我们的分布式锁实现了 `Lock` 接口，来确保依赖倒置，用户可以方便的在本地锁与分布式锁之前切换而无需改动逻辑。

在 class field 中，`OBTAIN_LOCK_SCRIPT` 是用于执行 redis 获取锁操作的 lua script，详情见后文。`localLock` 即本地锁，代码中采用 `ReentrantLock` 用作本地锁。

出于演示性质考虑，只实现了 `Lock` 中定义的三个方法：`lock()`, `tryLock(long time, TimeUnit unit)`, `unlock()`，其他方法可以自由发散。

接下来我们将主要介绍获取锁、释放锁这两部分代码。

#### lock

#### tryLock

#### unlock
