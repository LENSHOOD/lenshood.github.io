---
title: （Guava 译文系列）缓存
date: 2019-12-21 23:31:25
tags:
- guava
- translation
categories:
- Guava
---

# Caches

## 举例

```java
LoadingCache<Key, Graph> graphs = CacheBuilder.newBuilder()
       .maximumSize(1000)
       .expireAfterWrite(10, TimeUnit.MINUTES)
       .removalListener(MY_LISTENER)
       .build(
           new CacheLoader<Key, Graph>() {
             @Override
             public Graph load(Key key) throws AnyException {
               return createExpensiveGraph(key);
             }
           });
```

## 适用性

缓存在许多场景下都有巨大的用处。比如，当计算或取值的操作比较耗费资源，或你需要多次使用对于给定输入的输出值时，我们都应该使用缓存。

`Cache`与`ConcurrentMap`比较相似，但却并不完全一样。最显著的区别在于`ConcurrentMap`会将所有加入的元素持久化直至它们被显式的移除。而`Cache`通常可以被配置为自动移除 entries，来节省内存占用。某些情况下，由于`LoadingCache`的自动装载功能，即使不会自动清除entries，它也非常有用。

通常，Guava 的缓存工具在下述情况下比较适用：

*   你期望用一些消耗内存的代价来换取速度提升。
*   你期望一些 key 会被查询多次。
*   你的缓存不需要存储超过 RAM 容量大小的数据。
    （Guava 缓存对于你的应用而言是**本地**唯一的。它不会将数据存入文件或外部服务器。假如这不满足你的需求，可以考虑其他实现，例如 [Memcached](http://memcached.org/)。）

假如上面每一条都满足你的使用场景，那么 Guava 的缓存工具正适合你！

你当然可以像上述例子一样通过 builder模式`CacheBuilder` 来获取 `Cache`，然而更有趣的部分却是自定义你的缓存。

_注意：_假如你不需要`Cache`的特性，实际上 `ConcurrentHashMap`的存储效率更高 -- 可是采用古老的`ConcurrentMap`会非常难以甚至不可能复制大多数`Cache`提供的特性。


## Population

你需要问自己的第一个关于缓存的问题是：是否存在某个 _sensible default_ 函数，通过某个 key 来加载或计算与之关联的 value？如果是，你应该使用 `CacheLoader`。如果不是，或者如果你想要覆盖默认的函数，但是你仍然需要类似 “有则取值，无则计算” 的模式，那么你应该在调用 `get` 时传递一个 `Callable`。我们当然可以先用 `Cache.put` 来设置缓存，但是显然原子的缓存加载操作会更好，因为这会使我们更容易对所有缓存内容的一致性进行推理解释。

#### From a CacheLoader

`LoadingCache`是通过[`CacheLoader`]来构造的`Cache`。创建一个`CacheLoader`简单来说只需实现`V load(K key) throws Exception`。所以举例来说，你可以通过下述代码来创建`CacheLoader`：

```java
LoadingCache<Key, Graph> graphs = CacheBuilder.newBuilder()
       .maximumSize(1000)
       .build(
           new CacheLoader<Key, Graph>() {
             public Graph load(Key key) throws AnyException {
               return createExpensiveGraph(key);
             }
           });

...
try {
  return graphs.get(key);
} catch (ExecutionException e) {
  throw new OtherException(e.getCause());
}
```

查询`LoadingCache`的规范操作是调用[`get(K)`]方法。它将返回一个已经缓存的值，或是使用预设的`CacheLoader`自动加载一个新值并缓存。由于`CacheLoader`可能会抛出一个`Exception`，`LoadingCache.get(K)`会抛出`ExecutionException`。（假如 cache loader 抛出 _非检查_ 异常，`get(K)` 则会抛出一个`UncheckedExecutionException`来包装它。）你也可以选择使用`getUnchecked(K)`来将所有的异常包装为`UncheckedExecutionException`，然而假如底层的`CacheLoader`正常抛出检查异常时，该方法可能会导致某些令人惊讶的行为。

```java
LoadingCache<Key, Graph> graphs = CacheBuilder.newBuilder()
       .expireAfterAccess(10, TimeUnit.MINUTES)
       .build(
           new CacheLoader<Key, Graph>() {
             public Graph load(Key key) { // no checked exception
               return createExpensiveGraph(key);
             }
           });

...
return graphs.getUnchecked(key);
```

`getAll(Iterable<? extends K>)` 能实现批量查找。默认的，`getAll` 会对每一个在缓存中未找到的 key 调用`CacheLoader.load`。当批量检索的效率高于大多单独查找时，你可以通过覆盖[`CacheLoader.loadAll`]方法来使他被 `getAll(Iterable)`调用，以此提升整体性能。

请注意，你可以编写一个`CacheLoader.loadAll`的实现来加载并未明确被请求的 key 所对应的 value。例如，假设在某个组中，计算任意 key 返回的 value 正好是该组中所有的 key，那么用 `loadAll` 便可能同时将组内余下的 value 一并加载。

#### From a Callable

所有的 Guava 缓存，不论 loading 与否，都支持 [`get(K, Callable<V>)`]方法。该方法返回与给定 key 相关联的缓存 value，或者从给定的`Callable`中计算出 value，并添加进缓存。任何与缓存相关联的可观察状态都不会在加载完成前发生改变。该方法为传统的 “有则取值，无则计算” 模式提供了一个简单的替代。

```java
Cache<Key, Value> cache = CacheBuilder.newBuilder()
    .maximumSize(1000)
    .build(); // look Ma, no CacheLoader
...
try {
  // If the key wasn't in the "easy to compute" group, we need to
  // do things the hard way.
  cache.get(key, new Callable<Value>() {
    @Override
    public Value call() throws AnyException {
      return doThingsTheHardWay(key);
    }
  });
} catch (ExecutionException e) {
  throw new OtherException(e.getCause());
}
```

#### Inserted Directly

value 可以直接通过  [`cache.put(key, value)`] 方法直接插入进缓存。该方法会覆盖所有对应 key 先前的缓存值。也可以通过`Cache.asMap()`暴露出的`ConcurrentMap`行为来实现修改。请注意，`asMap`中暴露的任何方法都不会使 entries 自动加载到缓存中。而且，该视图上的原子操作超出了自动缓存加载的范围，所以`Cache.get(K, Callable<V>)` 总是应该优先于`Cache.asMap().putIfAbsent`来通过`CacheLoader` 或 `Callable`加载数据。

## Eviction

一个冰冷的现实是，我们 _从来_ 没有过够用的内存来缓存所有我们想缓存的东西。你得决定：在什么时候缓存的 entry 不值得再存下去了？Guava 提供了三种基础的失效类型：基于容量失效，基于时间失效，基于引用失效。

### 基于容量失效

假如你不想让缓存超出某个固定的容量，可以用[`CacheBuilder.maximumSize(long)`]设置。缓存会尝试使某些不常用或最近未使用的 entries 失效。_注意_：缓存可能会在还未超出容量限制时就开始失效 entries -- 通常是当缓存容量即将达到限额时。

此外，假如不同的 entries 有不同的“权重” -- 例如，假如你的缓存 value 有着完全不同的内存占用 -- 你也许可以通过 [`CacheBuilder.weigher(Weigher)`] 来指定一个权重，并且用 [`CacheBuilder.maximumWeight(long)`]来指定最大权重。另外，与`maximumSize`一样，请注意权重会在 entry 创建时计算，之后以 static 形式存在。

```java
LoadingCache<Key, Graph> graphs = CacheBuilder.newBuilder()
       .maximumWeight(100000)
       .weigher(new Weigher<Key, Graph>() {
          public int weigh(Key k, Graph g) {
            return g.vertices().size();
          }
        })
       .build(
           new CacheLoader<Key, Graph>() {
             public Graph load(Key key) { // no checked exception
               return createExpensiveGraph(key);
             }
           });
```

### 基于时间失效

`CacheBuilder` 对基于时间失效提供了两种方式：

*   [`expireAfterAccess(long, TimeUnit)`] 只会在距离最后一次读写的时间超出了指定的时间后失效。要注意的是，entries 的失效顺序与 [基于容量失效] 一致。
*   [`expireAfterWrite(long, TimeUnit)`]会在距离 entry 被创建，或 value 最近一次被替换的时间超出了指定的时间后失效。假如数据会在一段时间之后变为脏数据，那么正好可以使用这种方式。

定时过期是在写入、偶尔读取的期间执行定期维护的，具体见下述。

#### 测试

测试基于时间失效并不该痛苦... 而且事实上也不用真的等待两秒钟来测试一个两秒失效的缓存。可以使用 [Ticker] 接口和 [`CacheBuilder.ticker(Ticker)`] 方法来给 cache builder 指定一个时间源，而不需要等待系统时钟。

### 基于引用失效

Guava 可以让你将缓存配置为允许对 entries 进行垃圾收集，这可以通过 [weak references] 的 key 和 value，以及[soft references] 的 value 来实现。

*   [`CacheBuilder.weakKeys()`] 通过弱引用来存储 key。这允许该 entries 在其 key 没有被其他（强、软）引用时，可以被垃圾收集。由于垃圾收集只依赖与标识相等性，这使得整个缓存都使用标识相等 (`==`) 来比较 key，而不是 `equals()`。
*   [`CacheBuilder.weakValues()`] 通过弱引用来存储 value。这允许该 entries 在其 value 没有被其他（强、软）引用时，可以被垃圾收集。由于垃圾收集只依赖与标识相等性，这使得整个缓存都使用标识相等 (`==`) 来比较 key，而不是 `equals()`。
*   [`CacheBuilder.softValues()`] 将 value 包装为软引用。软引用对象以全局最近使用最少的方式进行垃圾收集，_来响应内存需求_ 。由于使用软引用对性能可能的影响，我们通常建议使用能加具有可预测性的 [最大缓存容量][size-based eviction] 来代替。使用`softValues()`可能会使 value 的比较使用标识相等 (`==`) ，而不是 `equals()`。

### 显式移除

任何时候，你想要显式的使缓存失效，而不是等待它被失效，可以通过下述方式实现：

*   单个失效 [`Cache.invalidate(key)`]
*   批量失效 [`Cache.invalidateAll(keys)`]
*   全部失效 [`Cache.invalidateAll()`]

### Removal Listeners

You may specify a removal listener for your cache to perform some operation when
an entry is removed, via [`CacheBuilder.removalListener(RemovalListener)`]. The
[`RemovalListener`] gets passed a [`RemovalNotification`], which specifies the
[`RemovalCause`], key, and value.

Note that any exceptions thrown by the `RemovalListener` are logged (using
`Logger`) and swallowed.

```java
CacheLoader<Key, DatabaseConnection> loader = new CacheLoader<Key, DatabaseConnection> () {
  public DatabaseConnection load(Key key) throws Exception {
    return openConnection(key);
  }
};
RemovalListener<Key, DatabaseConnection> removalListener = new RemovalListener<Key, DatabaseConnection>() {
  public void onRemoval(RemovalNotification<Key, DatabaseConnection> removal) {
    DatabaseConnection conn = removal.getValue();
    conn.close(); // tear down properly
  }
};

return CacheBuilder.newBuilder()
  .expireAfterWrite(2, TimeUnit.MINUTES)
  .removalListener(removalListener)
  .build(loader);
```

**Warning**: removal listener operations are executed synchronously by default,
and since cache maintenance is normally performed during normal cache
operations, expensive removal listeners can slow down normal cache function! If
you have an expensive removal listener, use
[`RemovalListeners.asynchronous(RemovalListener, Executor)`] to decorate a
`RemovalListener` to operate asynchronously.

### When Does Cleanup Happen?

Caches built with `CacheBuilder` do _not_ perform cleanup and evict values
"automatically," or instantly after a value expires, or anything of the sort.
Instead, it performs small amounts of maintenance during write operations, or
during occasional read operations if writes are rare.

The reason for this is as follows: if we wanted to perform `Cache` maintenance
continuously, we would need to create a thread, and its operations would be
competing with user operations for shared locks. Additionally, some environments
restrict the creation of threads, which would make `CacheBuilder` unusable in
that environment.

Instead, we put the choice in your hands. If your cache is high-throughput, then
you don't have to worry about performing cache maintenance to clean up expired
entries and the like. If your cache does writes only rarely and you don't want
cleanup to block cache reads, you may wish to create your own maintenance thread
that calls [`Cache.cleanUp()`] at regular intervals.

If you want to schedule regular cache maintenance for a cache which only rarely
has writes, just schedule the maintenance using [`ScheduledExecutorService`].

### Refresh

Refreshing is not quite the same as eviction. As specified in
[`LoadingCache.refresh(K)`], refreshing a key loads a new value for the key,
possibly asynchronously. The old value (if any) is still returned while the key
is being refreshed, in contrast to eviction, which forces retrievals to wait
until the value is loaded anew.

If an exception is thrown while refreshing, the old value is kept, and the
exception is logged and swallowed.

A `CacheLoader` may specify smart behavior to use on a refresh by overriding
[`CacheLoader.reload(K, V)`], which allows you to use the old value in computing
the new value.

```java
// Some keys don't need refreshing, and we want refreshes to be done asynchronously.
LoadingCache<Key, Graph> graphs = CacheBuilder.newBuilder()
       .maximumSize(1000)
       .refreshAfterWrite(1, TimeUnit.MINUTES)
       .build(
           new CacheLoader<Key, Graph>() {
             public Graph load(Key key) { // no checked exception
               return getGraphFromDatabase(key);
             }

             public ListenableFuture<Graph> reload(final Key key, Graph prevGraph) {
               if (neverNeedsRefresh(key)) {
                 return Futures.immediateFuture(prevGraph);
               } else {
                 // asynchronous!
                 ListenableFutureTask<Graph> task = ListenableFutureTask.create(new Callable<Graph>() {
                   public Graph call() {
                     return getGraphFromDatabase(key);
                   }
                 });
                 executor.execute(task);
                 return task;
               }
             }
           });
```

Automatically timed refreshing can be added to a cache using
[`CacheBuilder.refreshAfterWrite(long, TimeUnit)`]. In contrast to
`expireAfterWrite`, `refreshAfterWrite` will make a key _eligible_ for refresh
after the specified duration, but a refresh will only be actually initiated when
the entry is queried. (If `CacheLoader.reload` is implemented to be
asynchronous, then the query will not be slowed down by the refresh.) So, for
example, you can specify both `refreshAfterWrite` and `expireAfterWrite` on the
same cache, so that the expiration timer on an entry isn't blindly reset
whenever an entry becomes eligible for a refresh, so if an entry isn't queried
after it comes eligible for refreshing, it is allowed to expire.

## Features

### Statistics

By using [`CacheBuilder.recordStats()`], you can turn on statistics collection
for Guava caches. The [`Cache.stats()`] method returns a [`CacheStats`] object,
which provides statistics such as

*   [`hitRate()`], which returns the ratio of hits to requests
*   [`averageLoadPenalty()`], the average time spent loading new values, in
    nanoseconds
*   [`evictionCount()`], the number of cache evictions

and many more statistics besides. These statistics are critical in cache tuning,
and we advise keeping an eye on these statistics in performance-critical
applications.

### `asMap`

You can view any `Cache` as a `ConcurrentMap` using its `asMap` view, but how
the `asMap` view interacts with the `Cache` requires some explanation.

*   `cache.asMap()` contains all entries that are _currently loaded_ in the
    cache. So, for example, `cache.asMap().keySet()` contains all the currently
    loaded keys.
*   `asMap().get(key)` is essentially equivalent to `cache.getIfPresent(key)`,
    and never causes values to be loaded. This is consistent with the `Map`
    contract.
*   Access time is reset by all cache read and write operations (including
    `Cache.asMap().get(Object)` and `Cache.asMap().put(K, V)`), but not by
    `containsKey(Object)`, nor by operations on the collection-views of
    `Cache.asMap()`. So, for example, iterating through
    `cache.asMap().entrySet()` does not reset access time for the entries you
    retrieve.

## Interruption

Loading methods (like `get`) never throw `InterruptedException`. We could have
designed these methods to support `InterruptedException`, but our support would
have been incomplete, forcing its costs on all users but its benefits on only
some. For details, read on.

`get` calls that request uncached values fall into two broad categories: those
that load the value and those that await another thread's in-progress load. The
two differ in our ability to support interruption. The easy case is waiting for
another thread's in-progress load: Here we could enter an interruptible wait.
The hard case is loading the value ourselves. Here we're at the mercy of the
user-supplied `CacheLoader`. If it happens to support interruption, we can
support interruption; if not, we can't.

So why not support interruption when the supplied `CacheLoader` does? In a
sense, we do (but see below): If the `CacheLoader` throws
`InterruptedException`, all `get` calls for the key will return promptly (just
as with any other exception). Plus, `get` will restore the interrupt bit in the
loading thread. The surprising part is that the `InterruptedException` is
wrapped in an `ExecutionException`.

In principle, we could unwrap this exception for you. However, this forces all
`LoadingCache` users to handle `InterruptedException`, even though the majority
of `CacheLoader` implementations never throw it. Maybe that's still worthwhile
when you consider that all _non-loading_ threads' waits could still be
interrupted. But many caches are used only in a single thread. Their users must
still catch the impossible `InterruptedException`. And even those users who
share their caches across threads will be able to interrupt their `get` calls
only _sometimes_, based on which thread happens to make a request first.

Our guiding principle in this decision is for the cache to behave as though all
values are loaded in the calling thread. This principle makes it easy to
introduce caching into code that previously recomputed its values on each call.
And if the old code wasn't interruptible, then it's probably OK for the new code
not to be, either.

I said that we support interruption "in a sense." There's another sense in which
we don't, making `LoadingCache` a leaky abstraction. If the loading thread is
interrupted, we treat this much like any other exception. That's fine in many
cases, but it's not the right thing when multiple `get` calls are waiting for
the value. Although the operation that happened to be computing the value was
interrupted, the other operations that need the value might not have been. Yet
all of these callers receive the `InterruptedException` (wrapped in an
`ExecutionException`), even though the load didn't so much "fail" as "abort."
The right behavior would be for one of the remaining threads to retry the load.
We have [a bug filed for this](https://github.com/google/guava/issues/1122).
However, a fix could be risky. Instead of fixing the problem, we may put
additional effort into a proposed `AsyncLoadingCache`, which would return
`Future` objects with correct interruption behavior.

[`CacheLoader`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheLoader.html
[`get(K)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/LoadingCache.html#get-K-
[`CacheLoader.loadAll`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheLoader.html#loadAll-java.lang.Iterable-
[`get(K, Callable<V>)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/Cache.html#get-java.lang.Object-java.util.concurrent.Callable-
[`cache.put(key, value)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/Cache.html#put-K-V-
[`CacheBuilder.maximumSize(long)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#maximumSize-long-
[`CacheBuilder.weigher(Weigher)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#weigher-com.google.common.cache.Weigher-
[`CacheBuilder.maximumWeight(long)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#maximumWeight-long-
[`expireAfterAccess(long, TimeUnit)`]: https://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#expireAfterAccess-long-java.util.concurrent.TimeUnit-
[size-based eviction]: #基于容量失效
[`expireAfterWrite(long, TimeUnit)`]: https://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#expireAfterWrite-long-java.util.concurrent.TimeUnit-
[Ticker]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/base/Ticker.html
[`CacheBuilder.ticker(Ticker)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#ticker-com.google.common.base.Ticker-
[weak references]: http://docs.oracle.com/javase/6/docs/api/java/lang/ref/WeakReference.html
[soft references]: http://docs.oracle.com/javase/6/docs/api/java/lang/ref/SoftReference.html
[`CacheBuilder.weakKeys()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#weakKeys--
[`CacheBuilder.weakValues()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#weakValues--
[`CacheBuilder.softValues()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#softValues--
[`Cache.invalidate(key)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/Cache.html#invalidate-java.lang.Object-
[`Cache.invalidateAll(keys)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/Cache.html#invalidateAll-java.lang.Iterable-
[`Cache.invalidateAll()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/Cache.html#invalidateAll--
[`CacheBuilder.removalListener(RemovalListener)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#removalListener-com.google.common.cache.RemovalListener-
[`RemovalListener`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/RemovalListener.html
[`RemovalNotification`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/RemovalNotification.html
[`RemovalCause`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/RemovalCause.html
[`RemovalListeners.asynchronous(RemovalListener, Executor)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/RemovalListeners.html#asynchronous-com.google.common.cache.RemovalListener-java.util.concurrent.Executor-
[`Cache.cleanUp()`]: http://google.github.io/guava/releases/11.0.1/api/docs/com/google/common/cache/Cache.html#cleanUp--
[`ScheduledExecutorService`]: http://docs.oracle.com/javase/8/docs/api/java/util/concurrent/ScheduledExecutorService.html
[`LoadingCache.refresh(K)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/LoadingCache.html#refresh-K-
[`CacheLoader.reload(K, V)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheLoader.html#reload-K-V-
[`CacheBuilder.refreshAfterWrite(long, TimeUnit)`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheBuilder.html#refreshAfterWrite-long-java.util.concurrent.TimeUnit-
[`CacheBuilder.recordStats()`]: http://google.github.io/guava/releases/12.0/api/docs/com/google/common/cache/CacheBuilder.html#recordStats--
[`Cache.stats()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/Cache.html#stats--
[`CacheStats`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheStats.html
[`hitRate()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheStats.html#hitRate--
[`averageLoadPenalty()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheStats.html#averageLoadPenalty--
[`evictionCount()`]: http://google.github.io/guava/releases/snapshot/api/docs/com/google/common/cache/CacheStats.html#evictionCount--
