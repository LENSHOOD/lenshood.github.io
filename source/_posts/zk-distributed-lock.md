---
title: Java 基于 ZooKeeper 实现分布式锁需要注意什么
date: 2020-03-25 23:26:16
tags: 
- distributed lock
- zookeeper
categories:
- Java
---

在前一篇有关 Redis 分布式锁的[文章](https://lenshood.github.io/2020/02/04/redis-distributed-lock/)中，我们讨论了几点有关分布式锁的要求：
1. 操作原子性
2. 可重入性
3. 效率

为了满足上述条件，采用 `本地锁 + Redis 锁` 的方式解决了问题。不过在文章末尾提到，Redis 
不保证强一致性，因此对一致性要求很高的场景会存在安全隐患。

本文将讨论使用满足 CP 要求的 ZooKeeper 来实现强一致性的分布式锁。

<!-- more -->

### Zookeeper 分布式锁原理

结合 Redis 的分布式锁实现，我们能够想到最直接的 zk lock 实现方式，可能会是以 `ZNode` 来类比 redis 的 kv pair：创建一个 `ZNode`，通过判断其是否存在、以及其值是否与当前 client id 一致来尝试获取一个锁。

然而，结合 zk 的诸多优秀特性，实际上我们能更优雅的实现这一过程：
1. 创建一个路径为 `locknode/{guid}-lock-` 的 znode，同时将之设置为 `EPHEMERAL_SEQUENTIAL`, 其中的 `guid` 是为了解决一种边缘 case*。因此，我们会创建形如 `locknode/{guid}-lock-0000000012` 的一个节点。
2. 尝试获取 `locknode` 下的所有节点，对其进行排序，若刚刚创建的节点处在第一位，则获取锁成功，退出当前流程。
3. 若不为第一位，则对整个序列中排在自己持有的路径前一位的路径添加一个 watcher，并检查该前一位节点是否存在
4. 若前一位节点不存在，跳转至第二步，否则休眠等待。当被 watch 的路径发生变化时（通常是被删除），等待被唤醒并跳转至第二步。


可以看到，上述实现分布式锁的流程，用到了 zk 的两个特性：
1. sequence node
    - 通过 zk 内部保证的序列来确保获取锁公平（回顾 Redis 的方案，每隔 100ms 重试，是一种抢占式的非公平策略）
    - 每一次获取锁的尝试都会被如实的记录下来，易于观察整个获取锁的过程，也易于 debug
2. watcher
    - watcher 避免了轮询，每个等待中的路径都只观察其前一位路径，确保锁释放时只会有一个等待者（而不是所有）被唤醒，避免了羊群效应 （herd effect）。

> 注* guid 的特殊 case：对于`EPHEMERAL_SEQUENTIAL`节点的创建，假设节点创建成功，但 zk server 在返回创建结果之前 crash，那么在 client 重新连接至 zk 后，其 session 仍然有效，因此节点亦存在。
>
> 这时将出现诡异的一幕：某种情况下，该 client 以为自己没有获取到锁（实际上已经拿到了），这时他会再次创建一个 path，并休眠，而另一个 client 一直在等待第一位 path 被释放，但却永远也等不到（本来持有锁的 client 却休眠了）。
>
> 通过给 path 增加 guid 前缀的办法，当 client 检测到 create 非正常返回时，会启动 retry 流程：获取所有 children，若其中包含有 guid 的节点，则认为节点已经创建成功。

### 代码实现
1. lock()
```java
@Override
public void lock() {
    boolean acquired = false;
    localLock.lock();
    try {
        // reentrant
        acquired = localLock.getHoldCount() > 1;
        if (acquired) {
            return;
        }

        acquire();
        acquired = true;
    } finally {
        if (!acquired) {
            localLock.unlock();
        }
    }
}
```
与[Redis 分布式锁](https://lenshood.github.io/2020/02/04/redis-distributed-lock/)中实现类似，zk 分布式锁的 `lock()` 部分也采用了本地锁+分布式锁结合的方式：首先获取本地锁，之后尝试获取 zk 锁（即`acquire()`）。

这里对于可重入的处理比 Redis 的方案简单一些：
在 Redis 锁中，需要在 Redis 判断当前 client Id 是否与锁中保存的一致。而这里的方案，直接判断本地锁是否重入，若是则直接返回。

之所以能够简化，其原因是 ZooKeeper 锁并没有像 Redis 锁一样给锁加上了超时时间，再结合 ZooKeeper 强一致的特点，因此不会出现本地锁获取到而分布式锁被自动释放的情况。

接下来看看真正获取分布式锁的逻辑：
```java
void acquire() {
    String lockPath = createLockPath();
    if (lockPath.equals(getCurrentFirstPath())) {
        return;
    }

    boolean needDelete = true;
    watcherLock.lock();
    try {
        do {
            Condition condition = watcherLock.newCondition();
            addWatcher(getPreviousPath(lockPath), new LockWatcher(condition));
            condition.await();
        } while ((!lockPath.equals(getCurrentFirstPath())));
        needDelete = false;
    } catch (InterruptedException e) {
        Thread.currentThread().interrupt();
    } catch (Exception e) {
        throw new ZkLockException(e);
    } finally {
        watcherLock.unlock();
        if (needDelete) {
            deletePath(lockPath);
        }
    }
}
```
如上所见，`lock()` 方法实现了前文中描述的加锁过程：先创建锁路径，然后获取目前排序第一位的锁路径，若与创建的路径相同则直接获取锁，否则获取到前一个路径，对其添加 watcher，并进入休眠，直到被唤醒后获得锁。这里采用了一个 `watcherLock` 来控制休眠与唤醒。唤醒机制写在 `LockWatcher` 中：
```java
private class LockWatcher implements Watcher {
    private final Condition currentCondition;

    private LockWatcher(Condition currentCondition) {
        this.currentCondition = currentCondition;
    }

    @Override
    public void process(WatchedEvent event) {
        localLock.lock();
        try {
            currentCondition.signalAll();
        } finally {
            localLock.unlock();
        }
    }
}
```
通过实现 `Watcher` 来实现当被监听 path 有变动时释放 `Condition` 的等待状态

其他逻辑中包含的底层实现如下：
```java
String createLockPath() {
    try {
        return client.create().withProtection().withMode(CreateMode.EPHEMERAL_SEQUENTIAL).forPath(LOCK_PATH);
    } catch (Exception e) {
        throw new ZkLockException(e);
    }
}

private void deletePath(String lockPath) {
    try {
        client.delete().guaranteed().forPath(lockPath);
    } catch (Exception e) {
        // do nothing
    }
}
```
可以看到，底层实现中主要使用 `Curator` 来与 zk 进行交互，其中的 `client` 是 `WatcherRemoveCuratorFramework`。
> 实际上 Curator 本身提供了完整的 zk lock 实现，Spring Integration ZooKeeper 中的 LockRegistry 也直接包装了 Curator 的方案，本文以讨论原理为目的，实际使用中还是采用 Curator 更好。

```java
String getCurrentFirstPath() {
    List<String> allSortedPaths = getAllSortedPaths();
    if (allSortedPaths.isEmpty()) {
        throw new ZkLockException();
    }

    return allSortedPaths.get(0);
}

String getPreviousPath(String lockPath) {
    List<String> allSortedPaths = getAllSortedPaths();
    int previousIndex = allSortedPaths.indexOf(lockPath) - 1;
    if (previousIndex < 0) {
        throw new ZkLockException();
    }

    return allSortedPaths.get(previousIndex);
}

private List<String> getAllSortedPaths() {
    try {
        return client.getChildren()
                .forPath(BASE_LOCK_PATH)
                .stream()
                .sorted(Comparator.comparing(path -> path.split("-")[2]))
                .collect(Collectors.toList());
    } catch (Exception e) {
        throw new ZkLockException(e);
    }
}
```
以上为各种对所有锁路径的排序等操作。

2. unlock()
```java
@Override
public void unlock() {
    if (!localLock.isHeldByCurrentThread()) {
        throw new IllegalStateException("You do not own the lock");
    }

    if (localLock.getHoldCount() > 1) {
        localLock.unlock();
        return;
    }

    try {
        client.delete().guaranteed().forPath(getCurrentFirstPath());
    } catch (Exception e) {
        throw new ZkLockException(e);
    } finally {
        localLock.unlock();
    }
}
```
unlock 的过程简单了很多，首先判断线程是否合法，之后判断是否是重入状态，最后直接删除相关节点即可。
> 在 Curator 的 Lock [实现](https://github.com/apache/curator/blob/master/curator-recipes/src/main/java/org/apache/curator/framework/recipes/locks/LockInternals.java)中（commit f0a09db4423f06455ed93c20778c65aaf7e8b06e 之后的版本），release 锁之前，调用了`client.removeWatchers();`，经过代码分析，实际上对于 foreground 运行的 ZooKeeper 才删除 watcher，background 运行的不会删除。

### 总结
采用 ZooKeeper 实现的分布式锁，在实现原理上与 Redis 有一定的区别，它采用临时序列节点的方式实现公平的分布式锁，并通过 Watcher 机制，避免了释放锁时可能产生的羊群效应。

ZooKeeper 以其强一致性的特点，使得采用它实现的分布式锁安全可靠，不过性能相比 Redis 差一些。

实际使用中可以直接采用 Curator 提供的分布式锁方案，Curator Recipes 库包括了可重入、共享锁、信号量、栅栏等多种实现，方便可靠。
