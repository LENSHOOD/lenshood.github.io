---
title: MySQL(InnoDB) 独特的 Repeatable Read 隔离级别
date: 2020-11-30 10:05:35
mathjax: true
tags:
- database
- isolation level
- mysql
categories:
- MySQL
---

## MySQL 的事务隔离级别
在 [MySQL 8.0 Reference Manual](https://dev.mysql.com/doc/refman/8.0/en/innodb-transaction-isolation-levels.html) 中描述了 InnoDB 的事务隔离介绍：
> InnoDB 提供了 SQL-92 标准中描述的所有 4 种隔离级别的支持，它们是：READ UNCOMMITTED，READ COMMITTED，REPEATABLE READ，和 SERIALIZABLE。InnoDB 的默认隔离级别为 REPEATABLE READ。

<!-- more -->

对于 REPEATABLE READ 和 READ COMMITTED，Reference Manual 中描述到：

- REPEATABLE READ
对于同一个事务中的[一致性读（即读取事务开始时的数据库快照，是 MySQL 读的默认行为）](https://dev.mysql.com/doc/refman/8.0/en/glossary.html#glos_consistent_read)，MySQL 会基于事务开始的快照。就是说如果在同一个事务中发起了多个非阻塞 SELECT，这些 SELECT 彼此之间是保持一致的。

  对于 SELECT...FOR UPDATE（FOR SHARE)、UPDATE 、DELETE 语句，锁取决于语句是使用了具有唯一搜索条件的唯一索引，还是使用范围类型的搜索条件。
  - 对于唯一搜索条件的唯一索引，InnoDB 只锁住索引找到的记录，而不会包含 gap 锁
  - 对于其他的搜索条件，InnoDB 会锁住整个范围，使用 gap 锁或 next-key 锁来阻塞其他会话对被覆盖范围数据的插入。

- READ COMMITTED
对于一致性读，设置或读取记录的最新版本快照。
对于 SELECT...FOR UPDATE（FOR SHARE)、UPDATE 、DELETE 语句，InnoDB 只锁索引记录，不使用 gap 锁。由于不使用 gap 锁，可能会出现幻读（phantom）现象。

  额外影响：
  - 对于 UPDATE 、DELETE 语句，InnoDB 只持有被修改或删除的行的锁。对于未匹配到的记录的锁，会在执行完 WHERE 条件后释放。这极大地降低了（但未消除）死锁发生的概率。
  - 对于 UPDATE 语句，如果行已经被锁，InnoDB 执行 “半一致（semi-consistent）” 读，返回最新提交的版本给 MySQL，以便 MySQL 判断该行是否符合更新的 WHERE 条件。如果某一行成功匹配，MySQL再一次读取该行，之后 InnoDB 给该行加锁或等待行上已有的锁。

因此，从 MySQL Reference Manual 的角度看，MySQL 能够支持 [ANSI SQL-92](https://datacadamia.com/_media/data/type/relation/sql/sql1992.txt) 中定义的完整隔离级别，并且默认处于 ANSI SQL-92 REPEATABLE READ。


## ANSI Isolation Level 的扩展
ANSI SQL 标准的第一版发布于 1986 年，之后又陆续发布了多个主版本和修订版本，最新的修订版是 SQL-2019。不过，其最新的主版本仍然是 1992 年发布的 SQL-92。

如同 MySQL 文档中提到的，SQL-92 定义了四种隔离级别：READ UNCOMMITTED，READ COMMITTED，REPEATABLE READ，和 SERIALIZABLE。SQL-92 主要通过三种异象（phenomena）来对这四种隔离级别进行描述。这三种异象分别是:
1. 脏读（Dirty Read）
2. 不可重复读（Non-repetable Read）
3. 幻象（Phantom）

上面三种异象是我们耳熟能详的数据库基础知识。不过，在 ANSI SQL 发布三年后的 1995 年，一篇文章对这种隔离级别的划分与描述方法提出了质疑：

在文章 [*A Critique of ANSI SQL Isolation Levels*](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/tr-95-51.pdf) 中，作者认为仅通过上述三种异象还不足以清晰的定义隔离级别。相应的，作者对 ANSI SQL 的异象与隔离级别做了扩展，重新定义了 8 种异象和 6 种隔离级别。后来，该文成为了理解数据库隔离性的重要论文之一。

这一节我们将简单整理作者的理论。

为了能更清晰的表述事务之间的操作关系，我们将操作简化为 $w$ (write)，$r$ (read) ，每个操作的下标 $n$ 代表执行操作的事务，例如 $r_1$ 代表事务 1 读，$w_2$ 代表事务 2 写。紧跟着操作的中括号 $[]$ 的内容代表当前操作所涉及的资源，例如 $w_1[x]$ 代表事务 1 写入了资源 $x$，$r_2[P]$ 代表事务 2 读取了满足谓词 $P$ 的资源。最后，使用 $c$ (commit) 和 $a$ (abort) 来表示提交与回滚。

因此我们就可以用一连串的操作来表示一段操作历史：

 $w_1[x]...r_2[x]...$ ($a_1$ and $c_2$ in any order) 

可以表述事务 2 先写 $x$ ，之后事务 1  读 $x$，最后事务 1 回滚或事务 2 提交。

### 基于锁的隔离

为了实现事务之间的隔离，可以通过给读写分别加锁的方式（读共享锁、写独占锁）进行并发控制，加锁又可以分为以下三种粒度：

1. 不加锁
2. 使用前加锁，使用完立即释放锁，称为 Short duration 锁
3. 使用前加锁，直到事务提交后释放锁，称为 Long duration 锁 （类似两阶段锁的实现）

在使用锁隔离的事务控制实现下，能够定义如下异象与隔离级别的关系：

#### P0： $w_1[x]...w_2[x]...\ ((c_1\ or\ a_1)\  and\ (c_2\ or\ a_2)\ in\ any\ order)$ 

P0 称为 **Dirty Write**，在读写均不加锁时可能发生 P0。这种异象会破坏数据库的一致性，因此是任何隔离级别都不可容忍的。

禁止 P0 发生，隔离级别就能达到 READ UNCOMMITTED。

#### P1：$w_1[x]...r_2[x]...\ ((c_1\ or\ a_1)\ and\ (c_2\ or\ a_2)\ in\ any\ order) $

P1 称为 **Dirty Read**，在读不加锁，写加 Long duration 锁时可能发生 P1。这种情况可能导致事务读取到未提交的脏数据。如下事务执行历史模拟了 P1 发生的情况：

$r_1[x=50]..w_1[x=10]..r_2[x=10]..r_2[y=50]..c_2..r_1[y=50]..w_1[y=90]..c_1$

银行维护了账户$x$ 与 $y$，它们在银行的存款总余额为 $x+y=100$ 元。事务 1 中，$x$ 向 $y$ 转账 40 元，先修改 $x$ 为 10元，接着修改 $y$ 为 90 元。但事务 2 在事务 1 修改 $y$ 之前开始，它读取了账户余额，得取到 $x=10$，$y=50$，因此事务 2 中 $x+y=60$ 元，产生了不一致状态。

禁止 P1 发生，隔离级别就能达到 READ COMMITTED。

#### P2：$r_1[x]...w_2[x]...\ ((c_1\ or\ a_1)\ and\ (c_2\ or\ a_2)\ in\ any\ order)$

P2 称为 **Non-repeatable Read (Fuzzy Read)**，在读加 Short duration 锁，写加 Long duration 锁时可能发生 P2。这种情况可能导致同一事务中两次读取到的数据不一致。如下事务执行历史模拟了 P2 发生的情况：

$r_1[x=50]..r_2[x=50]..w_2[x=10]..r_2[y=50]..w_2[y=90]..c_2..r_1[y=90]..c_1$

与前面的例子类似，能够看到在事务 2 提交之后，事务 1 对 $y$ 的第二次读与第一次不同，此时，在事务 1 中， $x+y=140$，产生了不一致状态。

禁止 P2 发生，隔离级别就能达到 REPEATABLE READ。

#### P3： $r_1[P]...w_2[y \ in \ P]...\ ((c_1\ or\ a_1)\ and\ (c_2\ or\ a_2)\ in\ any\ order)$

P3 称为 **Phantom**，在读数据项加 Long duration 锁，读谓词（Predicate）加 Short duration 锁，写加 Long duration 锁时可能发生 P3。这种情况可能导致同一事务中两次读取到的数据量不一致。如下事务执行历史模拟了 P3 发生的情况：

$r_1[P]..w_2[insert\ y\ to\ P]..r_2[z]..w_2[z]..c_2..r_1[z]..c_1$

上述执行历史描述了事务 1 基于某种条件谓词 $P$ 来查找在职员工列表，这时事务 2 向在职员工列表插入了一个新员工，并更新了代表员工总数的数据项 $z$，在事务 2 提交后，事务 1 检查数据项 $z$，会发现与读取的在职员工列表不符。

禁止 P3 发生，隔离级别就能达到最高级 SERIALIZABLE。

#### CURSOR STABILITY

以上 P0 - P3 的描述相对比较符合 ANSI SQL-92 中对异象与隔离级别的描述。但论文作者认为 P0 - P3 还不足以更细致、完整的的描述异象与隔离级别的关系，因此文中继续定义了：

#### P4：$r_1[x]...w_2[x]...w_1[x]...c1$

P4 称为 **Lost Update**，稍加观察就可以发现，如果禁止 P2 发生，那么 P4 也一定不会发生。因此可以得出 P4 也是在 READ COMMITTED 级别下可能发生的异象。如下事务执行历史模拟了 P4 发生的情况：

$r_1[x=100]..r_2[x=100]..w_2[x=120]..c_2..w_1[x=130]..c_1$

显而易见，事务 2 对 $x$ 进行的更新丢失了。

现在，让我们引入游标：定义 $rc$ 为读游标， $wc$ 为写游标指向的数据项，那么存在游标时，P4 限定为 P4C：

#### P4C：$rc_1[x]...w_2[x]...w_1[x]...c1$

假如在读数据项时加基于游标（Cursor）的 Short duration 锁，读谓词（Predicate）加 Short duration 锁，写加 Long duration 锁（相比读完数据立即释放的 Short duration 读锁，基于游标的锁会将持有时间扩展到游标移动到下一个位置前）。

由于游标锁的存在，在 $rc_1$ 和 $wc_1$ 之间，一定不会插入 $wc_2$ （游标没有移动，锁一直存在），所以游标锁可以避免 P4C 的发生。

基于上述讨论，引入一个新的隔离级别 **CURSOR STABILITY**，当禁止 P4C 发生时，隔离级别就能达到 CURSOR STABILITY。

### 基于快照（版本）的隔离

为了在性能和一致性之间找到更好的平衡，许多数据库选择使用快照版本来进行并发控制（即 MVCC）。在事务开始时获取  *Start-Timestamp*，依据该时间戳读取最新的快照，读、写都基于快照进行，因此只读事务可以不被阻塞。

对于这种基于版本的隔离方式，作者提出了一种新的隔离级别 **SNAPSHOT ISOLATION**。在这样的隔离下，读、写操作都会基于事务开始时选择的一个快照，在事务提交前，获取一个提交时间戳 *Commit-Timestamp*，假如提交时间戳比当前系统内所有存在的  *Start-Timestamp* 或  *Commit-Timestamp* 都要新，则事务提交成功，否则失败，这种称为 *First-committer-wins* 的机制避免了 P4。进一步的，由于事务内读取当前快照，而不会读取到其他事务新提交的快照，因此 P2 也能够避免。从这个角度讲，SNAPSHOT ISOLATION 是比 READ COMMITTED 更高级别的隔离。

为了将 SNAPSHOT ISOLATION 与 REPEATABLE READ 进行对比，引入了如下几种异象：

#### A5A：$r_1[x]...w_2[x]...w_2[y]...c_2...r_1[y]...(c_1\ or\ a_1)$

#### A5B：$r_1[x]...r_2[y]...w_1[y]...w_2[x]...(c_1\ and\ c_2\ occur)$

#### A3：$r_1[P]...w_2[y\ in\ P]...c_2...r_1[P]...c_1$

其中 A5A 与 A5B 合称 A5（数据项约束冲突），A5 属于 P2 的一个子集，区别是在 A5 下 $x$ 与 $y$ 存在约束关系。

因此对于 A5A：$x$ 与 $y$ 存在约束关系，事务 1 中 $x$ 先被读取，由于事务 2 的提交，在事务 1 中随后读取的 $y$ 有可能已经不满足  $x$ 与 $y$ 的约束。A5A 又被称为 **Read Skew**。显然 A5A 在 SNAPSHOT ISOLATION 与 REPEATABLE READ 下都不会发生。

对于 A5B：$x$ 与 $y$ 存在约束关系，事务 1 和事务 2 分别读取了 $x$ 和  $y$ ，之后事务 1 更新了 $y$，最后事务 2 更新了 $x$。由于事务 2 更新 $x$ 时参考的 $y$ 已经不是最新值，因此 $x$ 与 $y$  的约束可能会被打破，这种异象称 **Write Skew**。显然，REPEATABLE READ 由于 Long duration 的读锁限制了 A5B 不可能发生，但在 SNAPSHOT ISOLATION 下由于事务 1、2 都读取满足各自时间戳的快照，所以 A5B 可能会发生。

A3 属于 P3 的一个子集，与 P3 的区别在于对事务的行为做了更多的限定。所以 A3 也属于一种 **Phantom**。 事务 1 基于条件谓词 $P$ 读取到的结果，由于事务 2 对符合 $P$ 的集合中新插入了数据，导致事务 1 再次按 $P$ 读取时读到了不同的结果集。由于 SNAPSHOT ISOLATION 读快照的特性，事务 2 版本的快照中对 $P$ 产生的影响，不会反映在事务 1 的快照中，因此  SNAPSHOT ISOLATION 下 A3 不可能发生，但显然 REPEATABLE READ 下由于对条件谓词的 Short duration 锁，A3 可能会发生。

基于上述分析，可以得出结论：SNAPSHOT ISOLATION 与 REPEATABLE READ 各有千秋，不分伯仲，无法简单比较高低。

### 完整的隔离级别与异象的关系

综合前文的所有分析，论文作者总结了下图：

{% asset_img summary.png %}

## MySQL 的 REPEATABLE READ 能够避免 Write Skew 吗？

根据第一节的描述，MySQL 的默认隔离级别是 REPEATABLE READ （实际是 InnoDB 的隔离级别，考虑到 InnoDB 是 MySQL 默认的存储引擎，后文不再区分）。同时，MySQL Reference Manual 中讲到 [InnoDB 使用了锁技术（行级锁）与 MVCC 技术（非阻塞读）来共同实现其事务模型](https://dev.mysql.com/doc/refman/8.0/en/innodb-transaction-model.html)，这是为了将传统的两阶段锁与多版本数据库的优势相结合以提升整体性能。

由于文档中对于隔离级别的描述是以 ANSI SQL-92 的四级隔离级别来描述的，那么结合第二节的内容，以及 MySQL 基于多版本的实现，我们是否可以假设，文档所述 MySQL 默认的 REPEATABLE READ，实际上是 SNAPSHOT ISOLATION？或者说，由于结合了锁与多版本，MySQL 既支持了 REPEATABLE READ 同时也支持了 SNAPSHOT ISOLATION？

以下我们用具体的示例来一一验证。

### P2 Non-repeatable Read

我们先从较低的异象开始，看看在默认隔离级别下，P2 会不会发生：

{% asset_img p2.png %}

上图的执行顺序是：$r_1[a]...w_2[a]...c_2...r_1[a]...c_1...r_1[a]$ 。

由于多版本的特性，事务 1 中读不阻塞，在事务 2 更新行并提交后，再次读取数据不变，直到事务 1 提交后，才会读取到最新值。

再以前文 P2 中的例子来测试：

$r_1[x=50]..r_2[x=50]..w_2[x=10]..r_2[y=50]..w_2[y=90]..c_2..r_1[y=90]..c_1$ 

{% asset_img p2_2.png %}

我们能看到在整个事务 1 中，始终保持 $x+y=100$ ，并没有出现 $x+y=140$ 的情况。

依据上述测试我们能够确信，默认情况下，MySQL 的“只读事务”（为什么是只读事务见下文）能够避免 Non-repeatable Read。

### P4 Lost Update

借用第二节 P4 的例子来测试：

$r_1[x=100]..r_2[x=100]..w_2[x=120]..c_2..w_1[x=130]..c_1$

{% asset_img p4.png %}

神奇的是，我们发现事务 1 先开始，且在事务 2 已经修改了 x 的情况下仍能正常提交（不满足 *First-committer-wins*），也就是说 MySQL 在默认 

REPEATABLE READ 的隔离级别下发生了 Lost Update。可以分析：

1. 假如 MySQL 是以 Long Duration 数据项读锁 + Long Duration 写锁来实现 REPEATABLE READ，那么在事务 1 正在读 $x$ 时，事务 2 对 $x$ 只可读而不可写；
2. 假如 MySQL 是以 MVCC 来实现 SNAPSHOT ISOLATION，那么根据 *First-committer-wins* 原则，事务 1 先开始，后提交，提交时会发现 $x$ 已经被修改而提交失败。

看来都不是，测试表明，默认情况下，MySQL 不能避免 Lost Update。

### A5A Read Skew

对于 Read Skew 异象的测试，仍然取先前转账的例子，其中有约束 $x+y=100$：

{% asset_img a5a.png %}

上图的执行顺序是：$r_1[x]..w_2[x]..w_2[y]..c_2..r_1[y]..c_1$

与 P2 的例子类似，由于多版本的特性，事务 1 的两次读取，第一次读 $x$，第二次读 $y$，虽然两次读取了不同的内容，但其仍然保持在同一个版本下，满足了 $x+y=100$ 的约束。因此默认情况下，MySQL 的只读事务能够避免 Read Skew。

### A5B Write Skew

对于 Write Skew 异象的测试，我们举一个直观的例子来说明：

> 医院急诊科 24 小时都有医生值班，为此医院会安排医生轮班。为了防止值班医生临时有事，通常轮班计划会尽量多安排几位医生同时在岗。此外，无论如何排班，都需要确保同一时刻至少有一位医生在岗。
>
> 值班医生可以通过电脑系统请假，当请假时段内有除了该医生以外的其他医生在岗时，就可以请假成功，否则失败。这能确保 “同一时刻至少一人在岗” 的约束条件。
>
> 急诊科共有 a, b, c 三位医生。某时刻，医生 a 和 b 被安排值班，这时 a 医生有事需要请假，不巧 b 医生也想请假，此时系统的操作历史如下...

{% asset_img a5b.png %}

上图的执行顺序是：$r_1[P:on\_call=true]..r_2[P:on\_call=true]..w_1[on\_call_a=false]..c_1..w_2[on\_call_b=false]..c_2$，结果是操作过后没有人值班，打破了约束。

以上的例子是一个很常见的并发故障，由于没有对条件谓词查询加 Long Duration 锁，导致不同事务因为旧的条件查询结果执行了错误的操作。据此，默认情况下，MySQL 不能避免 Write Skew。

### A3 Phantom

Phantom 的测试可以使用一个简单的例子：

{% asset_img a3.png %}

上图的执行顺序是：$r_1[P:a=1]..w_2[insert\ (1, 50)]..c_2..r_1[P:a=1]$，结果表明事务 1 的两次条件查询结果一致，符合多版本的特性，因此 MySQL 只读事务默认情况下能够避免 A3 （Phantom 的一种）。

但当我们继续做如下尝试：

{% asset_img p3.png %}

事务 1 对条件 $a=1$ 的所有值 +1 时，事务 2 插入的数据突然出现，并且也被 +1 了，这表明在读写事务中发生了 Phantom。

因此综合来看，默认情况下，MySQL 仍然不能避免 Phantom。

### 更低的隔离级别

依据上述例子，我们发现，MySQL 所谓默认的 REPEATABLE READ 级别，确实符合 ANSI SQL-92 中所定义的 “不会发生 Fuzzy Read，但可能发生 Phantom” 的描述。但实际上比 A Critique of ANSI SQL Isolation Levels 文中整理的  SNAPSHOT ISOLATION 和  REPEATABLE READ 都要低，因为 Lost Update、Write Skew、Phantom 都可能会发生。

根据 MySQL 文档中关于[一致性读](https://dev.mysql.com/doc/refman/8.0/en/innodb-consistent-read.html)的描述，发生上述现象的一个原因是，当在事务内对数据进行 update 操作后，当 select 语句读取受 update 影响的数据行时会返回最新版本（也即 update 是对最新版本而不是快照版本进行的），因此如果同时有另一个事务也在操作时原事务中可能看到原先不存在的数据。

## MySQL 的实现不好吗？

根据上一节的测试，似乎 MySQL 对事务隔离的实现并不怎么好，它在默认隔离级别下会出现各种异象，实现也难以准确的归类在  A Critique of ANSI SQL Isolation Levels  所述的几种隔离级别内。那么 MySQL 的事务隔离实现的不好吗？

想想也不一定，事务隔离本身就是对一致性与性能的平衡。在 MySQL 的实现下，虽然会出现多种异象，但事务提交失败回滚的几率相比标准的隔离级别实现要小得多，因此也就提升了并发性。对于相对简单的只读事务 MySQL 采用一致性读来支持更高的吞吐，而对于需要同时查询并更新数据的读写事务，常规的 select 并不能提供足够的防护，因此 MySQL 建议使用[上锁读](https://dev.mysql.com/doc/refman/8.0/en/innodb-locking-reads.html)（即我们熟悉的 `select ... for update` 或 `select for share`）来确保符合用户操作的真实意图。

InnoDB 在其事务模型实现中提到：`the goal is to combine the best properties of a multi-versioning database with traditional two-phase locking`，所以结合前文看， MySQL 的选择是默认采用多版本实现非阻塞读，同时提供了用户可选的两阶段锁实现更强的隔离性。这种实现直到最新的 MySQL 8.0 都没有改变过，从运行在 MySQL 上的多种应用，和其庞大的用户群来看，这种选择应该是合理的。

