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
对于 SELECT...FOR UPDATE（FOR SHARE)、UPDATE 、DELETE 语句，InnoDB 只锁索引记录，不使用 gap 锁。由于不使用 gap 锁，可能会出现幻读（phantoms）现象。

  额外影响：
  - 对于 UPDATE 、DELETE 语句，InnoDB 只持有被修改或删除的行的锁。对于未匹配到的记录的锁，会在执行完 WHERE 条件后释放。这极大地降低了（但未消除）死锁发生的概率。
  - 对于 UPDATE 语句，如果行已经被锁，InnoDB 执行 “半一致（semi-consistent）” 读，返回最新提交的版本给 MySQL，以便 MySQL 判断该行是否符合更新的 WHERE 条件。如果某一行成功匹配，MySQL再一次读取该行，之后 InnoDB 给该行加锁或等待行上已有的锁。

因此，从 MySQL Reference Manual 的角度看，MySQL 能够支持 [ANSI SQL-92](https://datacadamia.com/_media/data/type/relation/sql/sql1992.txt) 中定义的完整隔离级别，并且默认处于 ANSI SQL-92 REPEATABLE READ。


## ANSI Isolation Level 的扩展
ANSI SQL 标准的第一版发布于 1986 年，之后又陆续发布了多个主版本和修订版本，最新的修订版是 SQL-2019。不过，其最新的主版本仍然是 1992 年发布的 SQL-92。

如同 MySQL 文档中提到的，SQL-92 定义了四种隔离级别：READ UNCOMMITTED，READ COMMITTED，REPEATABLE READ，和 SERIALIZABLE。SQL-92 通过三个异象（phenomena）对这四种隔离级别做了描述。这三个异象分别是:
1. 脏读（Dirty Read）
2. 不可重复读（Non-repetable Read）
3. 幻象（Phantoms）

上面三种异象也是我们耳熟能详的数据库基础知识。不过，在 ANSI SQL 发布三年后的 1995 年，一篇文章对这种隔离级别的划分与描述方法提出了质疑：

在文章 [*A Critique of ANSI SQL Isolation Levels*](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/tr-95-51.pdf) 中，作者认为仅通过上述三种异象还不足以清晰的定义隔离级别。相应的，作者对 ANSI SQL 的异象与隔离级别做了扩展，重新定义了 8 种异象和 6 种隔离级别。后来，该文成为了理解数据库隔离性的重要论文之一。

下文简要阐述论文中定义的几种异象与隔离级别描述。

### 定义操作与历史

为了能更清晰的表述事务之间的操作关系，将操作简化为 $w$ (write)，$r$ (read) ，每个操作的下标 $n$ 代表执行操作的事务，例如 $r_1$ 代表事务 1 读，$w_2$ 代表事务 2 写。

紧跟着操作的中括号 $[]$ 的内容代表当前操作所涉及的资源，例如 $w_1[x]$ 代表事务 1 写入了资源 $x$，$r_2[P]$ 代表事务 2 读取了满足谓词 $P$ 的资源。

最后，使用 $c$ (commit) 和 $a$ (abort) 来表示提交与回滚。

用一连串的操作来表示一段操作历史，即

 $w_1[x]...r_2[x]...$ ($a_1$ and $c_2$ in any order) 

可以表述事务 2 先写 $x$ ，之后事务 1  读 $x$，最后事务 1 回滚或事务 2 提交。

### 基于锁的隔离

https://zhuanlan.zhihu.com/p/38334464

https://zhuanlan.zhihu.com/p/187597966

https://zhuanlan.zhihu.com/p/107659876

https://zhuanlan.zhihu.com/p/38214642

https://zhuanlan.zhihu.com/p/43621009

### 基于快照（版本）的隔离

### 汇总图

## MySQL 属于 Snapshot 还是 Repeatable？

