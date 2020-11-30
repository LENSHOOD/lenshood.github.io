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
> InnoDB 提供了 SQL-92 标准中描述的所有 4 种隔离级别的支持，它们是： READ UNCOMMITTED，READ COMMITTED，REPEATABLE READ，和 SERIALIZABLE。InnoDB 的默认隔离级别为 REPEATABLE READ。

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

从 MySQL Reference Manual 的角度看，MySQL 能够支持 [ANSI SQL-92](https://datacadamia.com/_media/data/type/relation/sql/sql1992.txt) 中定义的完整隔离级别，并且默认处于 ANSI SQL-92 REPEATABLE READ。


## ANSI Isolation 的扩展

## MySQL 属于 Snapshot 还是 Repeatable？