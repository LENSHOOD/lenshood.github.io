---
title: 在 Spring Webflux 使用 Spring Integration Redis 实现分布式锁
date: 2019-09-28 23:17:37
tags:
- java
- webflux
categories：
- Java
---
### 分布式锁
在传统单体应用中，我们用锁来保证非幂等操作在并发调用时不会产生意外的结果。最常见的应用场景应该就是 MySQL 用锁保证写表时不会重复写入或读表时读到脏数据。

进入微服务时代，整个业务系统可能是由数十个不同的服务节点组成，每个服务节点还包括多个实例确保高可用。在这样的环境下，一个写请求可能会由于负载均衡通过不同的服务实例操作数据，大多 NoSQL 实现为了并发性而牺牲了事务，则可能导致数据的正确性被破坏。这时如果有一个全局锁来对不同服务的操作进行限制，那么会一定程度解决上述问题。（对于复杂场景还需要采用分布式事务来处理回滚等等。）

与本地锁类似，分布式锁也是独立的对象，只不过存储在独立的节点上。最朴素的方法是在数据库中存储一段数据，以此为锁对象，存在则表示锁已被其他服务获取，不存在则表示可获取。当然此方案完全没考虑过死锁、可重入性等问题，而且如果是用关系型数据库来实现，则无法支撑高并发的场景。因此通常我们会采用 Redis、ZooKeeper 等方案来实现，并对锁代码进行一定设计，增加超时、重试等等功能。

### Redis 分布式锁
可以直接采用：`set [k] [v] px [milliseconds] nx`来原子的创建一个锁对象，其中 `px` 为超时时间，`nx` 为在 key 不存在时才创建。因此，对于尝试锁的逻辑，只有当上述命令返回`OK`时，才代表获得锁。同时，为了防止各种原因导致的死锁，超时时间过后，锁对象自动释放。

若考虑锁的可重入性，则需要对锁对象的值进行设计，确认不同线程（实例）获取锁时写入的值唯一，因此涉及可重入判断时，先`get [k]`获取值，若与本地唯一值一致，则可重入，重入后重置超时时间。

### Spring Integration Redis
Spring Integration 集成了许多中间件、第三方组件与 Spring 的适配，其中也包括 Redis。由于 Redis 锁简单、可靠的特点而被大规模使用，Spring Integration 索性直接提供了 Redis 锁的实现来简化开发，对应的类名：`RedisLockRegisty`。

`RedisLockRegisty`作为一个锁注册器，主要提供了`obtain(lock key)`和`destroy()`两种方法分别实现注册锁对象以及销毁。在`RedisLockRegisty`的内部实现了`RedisLock`内部类，它继承自 Java `Lock`，因此拥有锁通用的几个方法：
- `lock()`
- `tryLock()`
- `unlock()`
实际上，该实现采用了两层锁的结构，一层本地锁，一层 Redis 锁。这样做的好处是对于单实例内部的并发调用，可以直接走本地锁而不必与 Redis 通信，减少了操作时间，同时也降低了 Redis 的压力。

在`RedisLockRegisty`中，获取锁的操作，采用直接调用 hardcode 在类内的一段 lua 代码：
```lua
local lockClientId = redis.call('GET', KEYS[1])
    if lockClientId == ARGV[1] then
        redis.call('PEXPIRE', KEYS[1], ARGV[2])
        return true
    elseif not lockClientId then
        redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
        return true
    end
    return false
```

其中`KEYS[1]`为 key 值，`ARGV[1]`为 clientId，是在创建类时生成的 UUID，设置为锁的值用以判断可重入 `ARGV[2]`为超时时间。

可见上述 lua 代码与上一节提到的获取锁的流程一致。

### Webflux 结合 RedisLockRegistry
如何在采用 Project Reactor 异步框架为核心的 Spring Webflux 中应用`RedisLockRegistry`来实现原子操作？

Webflux 的编程思想是所有的操作都应在一个 stream 内完成，`RedisLock.lock()`作为一个阻塞操作，会阻塞当前流。那么如何在 Webflux 中使用 Redis 锁？

在 Project Reactor 文档中提到[如何包装一个同步阻塞的调用？](https://projectreactor.io/docs/core/release/reference/#faq.wrap-blocking)简单来讲，为了确保阻塞调用不阻塞整个流，我们需要将之运行在一个独立的线程内，采用`subscribeOn`来实现。

以下为相应实现代码：
```java
@Component
public class TransactionHelper {
    private RedisLockRegistry redisLockRegistry;

    @Autowired
    public TransactionHelper(RedisLockRegistry redisLockRegistry) {
        this.redisLockRegistry = redisLockRegistry;
    }

    /**
     * Do supplier in a transaction protected by a distributed lock, lock key is given by param key.
     */
    public <T> Mono<T> doInTransaction(String transactionKey, Supplier<Mono<T>> supplier) {
        Lock lock = redisLockRegistry.obtain(transactionKey);
        return Mono.just(0)
                .doFirst(lock::lock)
                .doFinally(dummy ->  lock.unlock())
                .flatMap(dummy -> supplier.get())
                .subscribeOn(Schedulers.elastic());
    }
}
```
上述代码通过对 supplier 操作的前后进行加锁，来实现将整个 supplier 的操作放置在同一事务内。

其中：
- supplier 为返回 `Mono<?>` 的的无参数调用（当然也可采用 Function，返回`Flux<?>`等形式更灵活的满足需要）
- `doFirst(Runnable)` 会确保 runnable 操作在执行 supplier 之前执行，此处的 runnable 为加锁，当无法获取锁时阻塞等待
- `doFinally(SignalType)`确保无论流发出任何结束信号（success，fail，cancel）都会在最后调用其设定的逻辑。
- `subscribeOn(Schedulers.elastic())`将上述流的所有操作放入由 Schedulers 创建的新线程中执行，因此不会阻塞主线程。

> 先进行 `doFirst` 和 `doFinally` 的原因是 supplier 中的操作有可能会造成线程切换，导致 `doFinally` 可能与 `doFirst`不在同一线程中执行，这有可能出现 Thread-A 创建的锁最终由 Thread-B 来释放的情况，使得锁报错并无法正确得到释放。

### 最后
上文介绍了用 Redis 构建分布式锁，并在 Webflux 框架下实现的方案。

Redis 分布式锁固然优秀，然而却并不是无懈可击的。试想假如有某个操作在 Redis 集群的某节点上创建了锁，然而在集群同步完成前该节点挂掉，那么锁就失效了。

基于此，Redis 的作者给出了“RedLock”方案，大致来讲是通过构造多个 Redis 集群，并多重上锁的方案，来降低故障的概率。Dr. Martin Kleppmann并不认为 Redlock 能解决故障，并[写了篇文章来论证](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html)，详情不在本文展开，请参考相关资料。
