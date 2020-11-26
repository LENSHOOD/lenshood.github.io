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

<!-- more -->

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
    private static final String clientId = UUID.randomUUID().toString();
    private static final int EXPIRE_SECONDS = ...;

    private final String lockKey;
    private final ReentrantLock localLock = new ReentrantLock();

    public RedisDistributedLock(String lockKey) {
        this.lockKey = lockKey;
    }

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

在 class field 中，
- `OBTAIN_LOCK_SCRIPT` 是用于执行 redis 获取锁操作的 lua script，详情见后文。
- `clientId`用于唯一标识当前所在实例，是分布式锁进行重入的重要属性，注意该 field 为 static，因此仅此一份。
-  `lockKey` 为锁 key，用于标识一个锁，在构造函数中初始化。
- `localLock` 即本地锁，代码中采用 `ReentrantLock` 用作本地锁。

出于演示性质考虑，只实现了 `Lock` 中定义的三个方法：`lock()`, `tryLock(long time, TimeUnit unit)`, `unlock()`，其他方法可以自由发散。

接下来我们将主要介绍获取锁、释放锁这两部分代码。

#### lock
```java
private static final String OBTAIN_LOCK_SCRIPT =
            "local lockClientId = redis.call('GET', KEYS[1])\n" +
            "if lockClientId == ARGV[1] then\n" +
            "  redis.call('EXPIRE', ARGV[2])\n" +
            "  return true\n" +
            "else if not lockClientId then\n" +
            "  redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])\n" +
            "  return true\n" +
            "end\n" +
            "return false";

@Override
public void lock() {
    localLock.lock();
    boolean acquired = false;
    try {
        while (!(acquired = obtainRemoteLock())) {
            sleep();
        }
    } finally {
        if (!acquired) {
            localLock.unlock();
        }
    }
}

private boolean obtainRemoteLock() {
    return Boolean.parseBoolean((String) getJedis().eval(
            OBTAIN_LOCK_SCRIPT, 1, lockKey, clientId, String.valueOf(EXPIRE_SECONDS)));
}

private void sleep() {
    try {
        Thread.sleep(100);
    } catch (InterruptedException e) {
        // do not response interrupt
    }
}
```

`lock()`中包含了绝大多数的核心逻辑，可以看到其主要流程如下：
- 获取本地锁
- 循环调用 `obtainRemoteLock()` 直至其返回 true，或抛出异常
- 假如跳出循环后仍未能获取到锁，则释放本地锁

以上流程中，需要细说的正是 `obtainRemoteLock()`：
该方法直接通过 `eval` 来执行了前面提到的 lua 脚本，我们来看看脚本的内容：
1. `local lockClientId = redis.call('GET', KEYS[1])`
    - 此处是通过 get 获取到了 key 值，并赋值为 lockClientId，其中 `KEYS[1]` 是 eval 传入的 key 参数
2. `if lockClientId == ARGV[1] then`
    - 这里将拿到的值与参数 ARGV[1] 进行判断，结合 `obtainRemoteLock()`的逻辑我们发现 ARGV[1] 其实是 `clientId`，所以假如获取的值与 clientId 相等，则代表一种情况：获取锁的线程与锁处于同一个实例
    - 又因为：每次获取远程锁之前需要先获取本地锁，在同一实例下，本地锁确保了同一时间只能有一个线程尝试获取远程锁
    - 结合上述两点，可以确定：**当 lockClientId 等于 clientId 的时候，是同一实例下的同一线程重入了代码段。**
    - `redis.call('EXPIRE', ARGV[2])` 在重入之后刷新锁超时时间，ARGV[2] 即我们传入的 `EXPIRE_SECONDS`
    - 最后直接返回 true，结束逻辑
3. `else if not lockClientId then`
    - 假如 get 的结果为 null(nil) 表明锁还没有被任何人获取，直接获取后返回 true
    - 这里用到了 redis 的 set 命令 `redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])`
4. `return false`
    - 即不是重入，锁又存在，证明锁被其他实例持有了，返回 false

上述一连串判断逻辑，因为全部都是在 Redis 内执行的，我们完全不用考虑原子性问题，因此可以放心大胆的相信执行结果。
#### tryLock
```java
@Override
public boolean tryLock(long time, TimeUnit unit) throws InterruptedException {
    if ( !localLock.tryLock(time, unit)) {
        return false;
    }

    boolean acquired = false;
    try {
        long expire = System.currentTimeMillis() + TimeUnit.MILLISECONDS.convert(time, unit);
        while (!(acquired = obtainRemoteLock()) && System.currentTimeMillis() < expire) {
            sleep();
        }
        return acquired;
    } finally {
        if (!acquired) {
            localLock.unlock();
        }
    }
}
```
结合 `lock()`的逻辑，`tryLock()`看起来只是增加了超时逻辑，并没有本质的区别。

#### unlock
```java
@Override
public void unlock() {
    if (!localLock.isHeldByCurrentThread()) {
        throw new IllegalStateException("You do not own lock at " + lockKey);
    }

    if (localLock.getHoldCount() > 1) {
        localLock.unlock();
        return;
    }

    try {
        if (!clientId.equals(getJedis().get(lockKey))) {
            throw new IllegalStateException("Lock was released in the store due to expiration. " +
                    "The integrity of data protected by this lock may have been compromised.");
        }
        getJedis().del(lockKey);
    } finally {
        localLock.unlock();
    }
}
```
相比本地锁，分布式锁的解锁过程需要考虑的多一些：
1. 先判断尝试解锁的线程与持有本地锁的线程是否一致，实际上 ReentrantLock.unlock() 原生即有相关判断，但是目前我们还暂时不想让本地锁直接被解锁，因此手动判断一下。
2. 当本地锁重入计数大于 1 时，本地锁解锁后直接返回。由于我们的远程锁并没有记录重入计数这一参数，因此对于重入线程的解锁，只解锁本地。
3. 先判断当前远程锁的值是否与本实例 clientId 相等，如果不等则认为是远程锁超时被释放，因此分布式锁的逻辑已经被破坏，只能抛出异常。
    - 这里涉及到远程锁超时时间的设定问题，设定过长可能会导致死锁时间过长，设定过短则容易在逻辑未执行完便自动释放，因此实际上应该结合业务来设定。
4. 假如一切正常，则释放远程锁，之后再释放本地锁。

### RedisLockRegistry
对 Spring Integration Redis 熟悉的同学，一定已经发现，前面的代码完全就是 RedisLockRegistry 的简化版，许多变量名都没改。

是的，其实前文所述的代码就是 RedisLockRegistry 的核心逻辑。RedisLockRegistry 是 Redis 分布式锁中代码比较简单、功能比较完善的一种实现，可以很好的满足常见的分布式锁要求。（由于采用 sleep-retry  的方式尝试获取锁，在低时延或高并发要求下并不适用）

RedisLockRegistry 对外部库的依赖较少，虽然执行 redis 命令主要使用的 Spring Redis Template，不过也很容易迁移为类似 Jedis 的方案。

不过截至目前 Spring-Integration-Redis 在 github 上面并没有放置任何 licence，按照 github 的规定，没有 licence 的代码版权默认受到保护，因此我们可以学习其设计思想并自己尝试实现，但是最好不要直接移植代码。

### Redis 多实例
通过上述方法，我们似乎可以成功的将实例间同步的问题转交给 Redis 来处理。然而就存在两种情况：
1. 采用单实例 Redis -- Redis 存在单点风险，应用服务都依赖 Redis， 一旦宕机业务全挂
2. 采用 Redis 集群 -- 应用服务实例间的同步问题转化为了 Redis 实例间的同步问题

单实例 Redis 一定是不可接受的，所以似乎允许上生产环境的唯一方案就是 Redis 集群了。那么如何保证 Redis 实例间的同步呢？

我们知道，Redis 集群的数据冗余策略不同于类似 HDFS 的 3 Replica，而是采用一对一主从的形式，每个节点一主一从，主节点宕机备节点上，备节点也宕机就全完。同时，主从之间的数据同步是异步的。以上这些都是为了超高吞吐量而做出的妥协。

所以，设想会有这种情况：

当应用服务节点 App-A 从 Redis 某主节点 R-Master 获取到锁后，R-Master 宕机，此时 R-Master 的数据还没来得及同步到 R-Slave。现在 R-Slave 成为了主节点，这时候 App-B 尝试获取锁，不出意外的也获取成功了。

基于以上问题，Redis 给出了 [RedLock](https://redis.io/topics/distlock) 方案，该方案采用相互孤立的奇数个 Redis 节点来共同存储锁，对于获取锁的操作，只有当 (N-1)/2 + 1 个 Redis 实例都获取成功且获取时间不超过锁失效时间的前提下，才真正被判定为获取到了锁，这种场景下锁的争抢就看谁能先成功操作超过半数的 Redis 实例。Redisson 实现了 [RedLock 的客户端方案](https://github.com/redisson/redisson/wiki/8.-distributed-locks-and-synchronizers)。

当然，在 Redis 官网上也贴出了各方对于 RedLock 方案的争论，这里不再赘述。

总之，对于问题的处理终归是结合实际情况来权衡的，
- 假如小概率（但几乎一定会发生）的 Redis 宕机未同步导致锁失效的问题，业务可以承受，那么 RedisLockRegistry + Redis 集群的方案就没问题
- 对性能和可靠性都有更高要求的情况下，不妨使用 RedLock 方案
- 业务非常关键，一定要求强一致的分布式锁，使用 ZooKeeper 的方案会更好（性能没法和 Redis 比）

### 参考
[RedisLockRegistry at Github](https://github.com/spring-projects/spring-integration/blob/master/spring-integration-redis/src/main/java/org/springframework/integration/redis/util/RedisLockRegistry.java)

[Redis Documentaion](https://redis.io/documentation)
