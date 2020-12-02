---
title: Isolation Level 与 MySQL
date: 2020-11-30 10:05:35
tags:
- database
- isolation level
- mysql
categories:
- MySQL

---
One of the foundations of database processing. Isolation is the I in the acronym ACID; the isolation level is the setting that fine-tunes the balance between performance and reliability, consistency, and reproducibility of results when multiple transactions are making changes and performing queries at the same time.


## MySQL 的事务隔离级别
在 [MySQL 8.0 Reference Manual](https://dev.mysql.com/doc/refman/8.0/en/innodb-transaction-isolation-levels.html) 中描述了 InnoDB 的隔离事务隔离介绍：
> InnoDB 提供了 SQL-92 标准中描述的所有 4 种隔离级别的支持，它们是：READ UNCOMMITTED，READ COMMITTED，REPEATABLE READ，和 SERIALIZABLE。InnoDB 的默认隔离级别为 REPEATABLE READ。

其中对于 REPEATABLE READ 和 READ COMMITTED，Reference Manual 中描述：
- REPEATABLE READ
对于同一个事务中的[一致性读（即读取事务开始时的数据库快照，是 MySQL 读的默认行为）](https://dev.mysql.com/doc/refman/8.0/en/glossary.html#glos_consistent_read)会读取建立在第一次读的快照上。就是说如果在同一个事务中发起了多个非阻塞 SELECT，这些 SELECT 彼此之间是保持一致的。

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

如同 MySQL 文档中提到的，SQL-92 定义了四种隔离级别：READ UNCOMMITTED，READ COMMITTED，REPEATABLE READ，和 SERIALIZABLE。SQL-92 通过三个异象（phenomena）对这四种隔离级别做了描述。这三个异象分别是:
1. 脏读（Dirty Read）
2. 不可重复读（Non-repetable Read）
3. 幻象（Phantom）

上面三种异象也是我们耳熟能详的数据库基础知识。不过，在 ANSI SQL 发布三年后的 1995 年，一篇文章对这种隔离级别的划分与描述方法提出了质疑：

在文章 [*A Critique of ANSI SQL Isolation Levels*](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/tr-95-51.pdf) 中，作者认为仅通过上述三种异象还不足以清晰的定义隔离级别。相应的，作者对 ANSI SQL 的异象与隔离级别做了扩展，重新定义了 8 种异象和 6 种隔离级别。后来，该文成为了理解数据库隔离性的重要论文之一。

为了能更清晰的表述事务之间的操作关系，我们将操作简化为 $w$ (write)，$r$ (read) ，每个操作的下标 $n$ 代表执行操作的事务，例如 $r_1$ 代表事务 1 读，$w_2$ 代表事务 2 写。紧跟着操作的中括号 $[]$ 的内容代表当前操作所涉及的资源，例如 $w_1[x]$ 代表事务 1 写入了资源 $x$，$r_2[P]$ 代表事务 2 读取了满足谓词 $P$ 的资源。最后，使用 $c$ (commit) 和 $a$ (abort) 来表示提交与回滚。

用一连串的操作来表示一段操作历史，即：

 $w_1[x]...r_2[x]...$ ($a_1$ and $c_2$ in any order) 

可以表述事务 2 先写 $x$ ，之后事务 1  读 $x$，最后事务 1 回滚或事务 2 提交。

### 基于锁的隔离

为了实现事务之间的隔离，可以通过给读写分别加锁的方式（读共享锁、写独占锁）进行并发控制，加锁又可以分为以下三种粒度：

1. 不加锁
2. 使用前加锁，使用完立即释放锁，称为 Short duration 锁
3. 使用前加锁，直到事务提交后释放锁，称为 Long duration 锁

在使用锁隔离的场景下，能够定义如下异象与隔离级别的关系：

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

$ r_1[x=50]..r_2[x=50]..w_2[x=10]..r_2[y=50]..w_2[y=90]..c_2..r_1[y=90]..c_1$

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

对于这种基于版本的隔离方式，作者提出了一种新的隔离级别 **SNAPSHOT ISOLATION**。在这样的隔离下，读、写操作都会基于事务开始时选择的一个快照，在事务提交前，获取一个提交时间戳 *Commit-Timestamp*，假如提交时间戳比当前系统内所有存在的  *Start-Timestamp* 或  *Commit-Timestamp* 都要新，则事务提交成功，否则失败，这种机制称为 *First-committer-wins*，从而避免了 P4。进一步的，由于事务内读取当前快照，而不会读取到其他事务新提交的快照，因此 P2 也能够避免。从这个角度讲，SNAPSHOT ISOLATION 是比 READ COMMITTED 更高级别的隔离。

为了将 SNAPSHOT ISOLATION 与 REPEATABLE READ 进行对比，引入了如下几种异象：

#### A5A：$r_1[x]...w_2[x]...w_2[y]...c_2...r_1[y]...(c_1\ or\ a_1)$

#### A5B：$r_1[x]...r_2[y]...w_1[y]...w_2[x]...(c_1\ and\ c_2\ occur)$

#### A3：$r_1[P]...w_2[y\ in\ P]...c_2...r_1[P]...c_1$



https://zhuanlan.zhihu.com/p/38334464

https://zhuanlan.zhihu.com/p/187597966

https://zhuanlan.zhihu.com/p/107659876

https://zhuanlan.zhihu.com/p/38214642

https://zhuanlan.zhihu.com/p/43621009

### 汇总图

## MySQL 属于 Snapshot 还是 Repeatable？

