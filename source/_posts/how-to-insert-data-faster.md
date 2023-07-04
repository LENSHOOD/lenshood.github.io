---
title: 如何让数据库插入速度提升200倍
date: 2023-06-11 14:50:45
tags:
- prepare data
- SQL
categories:
- DB
---

{% asset_img header.jpg 300 %}

本文介绍了在一次构造测试数据的活动中，我们如何通过优化手段将数据插入速度提升200倍的经历。

<!-- more -->

在帮助某客户构建业务系统过程中，为了测试数据库表性能，我们需要构造 1 亿条模拟业务数据，这些数据在关键业务字段上需要能反映实际业务特点，包括数据类型和数据分布。

客户先前曾经构造过同样规模的测试数据，通过在代码中循环批次插入，构造一亿条数据大约需要 6 小时。据此我们认为这种方案引入了不必要的网络 IO，效率太差，通过存储过程构造肯定会更快。

### 第一个版本

在对构造数据的需求进行分析后，我们计划通过两步来实现功能：

1. 构造业务数据集
   - 预置业务数据集，按数据分布的实际比例进行推算，预置的数据集的规模在数十万条左右
   - 对于唯一性业务数据（主要是流水号类）创建序列生成器生成唯一数值
2. 编写存储过程
   - 通过随机函数从业务数据集中选取数据
   - 拼接 INSERT 语句，批次插入
   - 循环上述过程

搞清楚需求之后，我们（~~ChatGPT~~）很快就写出了如下脚本：

```sql
-- 数据库：openGauss

-- 假设：
-- 被插入数据表名 t0，包含三个字段 A(bigint), B(varchar), C(datetime)
-- A 数据来自 Sequence s0，B 数据随机选取业务数据集 t1，C 数据随机选取 2014~2023 年中任意一天
CREATE OR REPLACE PROCEDURE P0(IN n INTEGER, IN m INTEGER)
DECLARE
    i INTEGER;
    a_value BIGINT;
    b_value VARCHAR;
    c_value TIMESTAMP;
    sql_insert TEXT := '';
BEGIN
    FOR i IN 1..n LOOP
        sql_insert := '';
        FOR j IN 1..m LOOP
            SELECT nextval('s0') INTO a_value;
            -- 这里由于数据库原因，无法使用 TABLESAMPLE
            SELECT b_column FROM t1 ORDER BY random() LIMIT 1 INTO b_value;
            SELECT TIMESTAMP '2014-01-01' + random() * (TIMESTAMP '2023-06-30' - TIMESTAMP '2014-01-01') INTO c_value;  
            -- 拼接 insert statement
            sql_insert := sql_insert || format('(%s, %s, %s),', a_value, quote_literal(b_value), quote_literal(c_value));
        END LOOP;
        
        sql_insert := rtrim(sql_insert, ',');
        EXECUTE format('INSERT INTO t0 (A, B, C) VALUES %s;', sql_insert);
        
    END LOOP;
END;
```

然而执行后发现，这个存储过程慢的令人发指，估算插入一亿条数据耗时 180 小时，比代码插入反倒慢了 30 倍。

### 性能优化

为了挽回顾问（有但不多）的专业度，我们决定对存储过程进行性能优化。

一般来说，实施性能优化之前，可以先进行分析以寻找可能的性能劣化点，再决定从哪里开始。由于脚本比较简单，我们很容易发现一些可能的劣化点：

- `nextval()` 函数及 Sequence 的性能
- `random()` 函数性能
- `ORDER BY random()` 的执行效率
- FOR 循环

只要依次测试就能发现最耗时之处。

#### 优化 SQL 语句

分析执行计划是对 SQL 进行优化时最直接的尝试。通过执行计划，我们快速发现了第一处问题点：

```SQL
SELECT b_column FROM t1 ORDER BY random() LIMIT 1
Limit
  -> Sort
    Sort Key: (random())
      -> Seq Scan on xxxxx
```

通过 `order by random()` 来从数据集 `t1` 中选取随机值，实际上是为 `t1` 中每一行数据生成随机值，之后按每行随机值进行排序得来的。由于我们只选取排序后的第一条数据，因此排序工作并没有意义。

修改后的语句+执行计划：

```sql
SELECT b_column FROM t1 WHERE id in (SELECT floor(random()*10000)) LIMIT 1
Index Only Scan
  Index Cond
  Init Plan1
    -> Result xxx
```

 `t1` 的主键是自增整数，我们直接随机一个不超限的整数值，通过主键索引快速查找。

另外，考虑到事务提交与 insert 拼接性能损耗的平衡，经测试插入批次规模在 100~1000 之间效率最高。

再进一步，由于数据集和实际生成数据的比例在 1:10000 以上，因此无需每一次都进行随机选取，当内循环批次大小在几百至数千时数据离散程度不会发生明显变化。故将随机选取语句放入外循环，可以大幅降低计算次数。

从结果上看，最明显的劣化点，修改后效果最好：通过上述修改，预估只需要 4 小时就可插入完成，性能提升了 45 倍。

#### 提高资源利用率

基于第一次分析，我们发现除了`ORDER BY`以外，其他几部分的耗时都很短，优化的意义不大。因此我们跳出耗时，转向资源利用率。

对于数据库的插入性能，主要的指标关注点在于 **IO 写入速率**和 **CPU 使用率**。IO 写入速率越高说明单位时间插入的数据越多，CPU 使用率越高证明执行引擎跑的越满（等待 IO 时间少）。

通过监控发现，在存储过程执行期间，IO Write 约 20MiB/s，CPU 使用率 10~20%，基于此我们认为可以尝试加并发。由于我们并不清楚数据库的硬件配置，无法根据 CPU 核数来划定线程数范围，因此进行了多轮并发插入测试，最后得到结论是 8 个线程下加速比最高。

同时，依赖和竞争是限制并发性能的两大障碍。容易发现，这些并发运行的存储过程都依赖同一个 Sequence `s0` 来生成序列数字。

Sequence 通过加锁来实现递增的并发安全，为了提升 `nextval()` 的并发取数性能，为 Sequence 添加 `Cache` 属性，降低锁竞争。经测试 Cache 前后  `nextval()`  性能可以提升一倍以上。

最终，进行并发优化后的插入速度同比又提升了 4 倍，并且 CPU 使用率达到 96%，IO Write 约 115MiB/s。

由于缺少监控数据，我们无法确定 CPU 使用率中 user 和 sys 的比例，但考虑到托管云服务通常会正确设置线程池、绑核等参数，因此我们认为 user 占据主导。

#### 减少冗余操作

经过上一轮优化后，CPU 使用率已达 96%，IO Write 则达到 115MiB/s。我们虽然无法判断 IO 带宽被利用了多少，但可以明确 CPU 已经基本被打满。在这种情况下想要进一步优化，不妨将视线移至计算逻辑。

数据库执行一次操作，在计算层大致需要经过解析 SQL，逻辑+物理优化、形成执行计划这几个过程，之后会在计算层和存储层分别执行各类算子。计算层主要是 CPU 计算，在存储层更多是 CPU 和 IO 操作混合。

我们发现，对变量的赋值操作都是采用 `SELECT xxx INTO var` 的形式，这样每个变量的赋值都要经历一遍解析、优化和执行的过程。因此我们尝试将这种操作修改为 `var := {value conpute}` 的形式，即对变量直接赋值。测试表明，直接赋值比`SELECT xxx INTO var`要快 5 倍左右。

经过以上修改后的存储过程，整体 CPU 利用率下降了 2%，而 IO Write 提升了 17%，达到了135MiB/s 。

通过几轮优化，我们最终在 50 分钟内完成了 1 亿条数据的写入，比原先预估的 180 小时，提速了 216 倍。

这里附上优化后的脚本：

```sql
CREATE SEQUENCE s0 cache 1000;

CREATE OR REPLACE PROCEDURE P0(IN n INTEGER, IN m INTEGER)
DECLARE
    i INTEGER;
    a_value BIGINT;
    b_value VARCHAR;
    c_value TIMESTAMP;
    sql_insert TEXT := '';
BEGIN
    FOR i IN 1..n LOOP
        SELECT b_column FROM t1 WHERE id in (SELECT floor(random()*10000)) LIMIT 1 INTO b_value;
        c_value := TIMESTAMP '2014-01-01' + random() * (TIMESTAMP '2023-06-30' - TIMESTAMP '2014-01-01');
        
        sql_insert := '';
        FOR j IN 1..m LOOP
            a_value := nextval('s0');
            sql_insert := sql_insert || format('(%s, %s, %s),', a_value, quote_literal(b_value), quote_literal(c_value));
        END LOOP;
        
        sql_insert := rtrim(sql_insert, ',');
        EXECUTE format('INSERT INTO t0 (A, B, C) VALUES %s;', sql_insert);
        
    END LOOP;
END;
```

### 性能优化的思考

前文所述的性能优化过程，场景简单、依赖少，很适合用来作为熟悉性能分析和优化流程的案例。因此我们总结了以下几点对性能分析和优化的思考实践：

#### 了解系统架构

**对系统的了解程度直接决定了进行性能分析时定义问题的范畴。**

通过了解系统架构，我们能回答如下问题，进而帮助分析：

- 系统提供的是计算类服务还是存储类服务？据此大致判断需要关注的指标。如计算类服务关注 CPU 使用率、线程调度、并行度；存储类服务关注请求吞吐量、网络带宽、IO等。
- 各业务模块的核心逻辑是什么？据此筛选可能的性能瓶颈点和劣化点。如数据库的 Parser、Optimizer、Executor 等模块的功能和资源使用特点。
- 子模块之间的交互关系是什么？据此了解模块间调用链和阻塞点。如函数调用和 IPC 的时间差异、缓冲队列的使用率、缓存命中率等。
- 用到了哪些关键算法或数据结构？据此可以更细致的分析系统的工作过程。如分析 HashJoin 和 IndexJoin 的性能差异、B+Tree 与 LSM Tree 对读写的不同表现等。

> 在前面的例子中，对数据库架构有一定程度的认识，帮助了我们更快的进行优化决策。

#### 建立模型

结合对系统架构和业务模型的了解，我们能**更加自信的在高阶层面对系统性能关注点和可能的劣化点进行建模**。

可能的步骤包括：

1. 确定关注对象：根据观测指标和业务模块，框定出大致的关注对象范围
2. 识别性能劣化点：从逻辑、交互、算法和数据结构层面识别可能的劣化点
3. 评估优先级：对劣化点进行评估和排序，先从综合收益大的地方开始

> 在前面的例子中，存储过程脚本比较简单，建模意义不大。但对于稍显复杂的系统，建模能帮助我们厘清组件关系，让劣化点一目了然。

#### 观测洞察的重要性

**没有观测，何谈分析？成熟的可观测能力是深入分析系统的基石。**

上文描述的场景中，我们获取的信息很有限：不清楚硬件环境，可供观测的指标也很少，不了解数据库所在机器的系统配置，对 OpenGauss 和 PostgreSQL 在架构设计上的差异也不够熟悉。

观测不足导致对系统认知模糊，这使我们容易陷入无头苍蝇的境地：用“想到什么就试一试”来代替方法和流程。好在是我们优化的对象非常简单，否则可能没这么顺利。

正如医生在诊断复杂疾病的时候，都会要求进行各项身体检查。全面的观测能力可以提升对系统的洞察深度，也就更易于正确的进行建模、分析和优化。略带夸张的讲，对系统观测的深度，甚至决定了性能优化的成败。

构建成熟而完善的系统观测能力对于性能优化至关重要，然而，这并非一蹴而就的过程。它需要在观测体系、性能工程等领域持续进行投入和建设。

> 在前面的例子中，虽然观测指标少，但至少有 CPU 使用率和 IO 写入速率，这两项监控指标帮助我们完成了后续的关键优化动作。



### 扩展阅读

[PostgreSQL: Populating a Database](https://www.postgresql.org/docs/current/populate.html)

[PostgreSQL: Sequence](https://www.postgresql.org/docs/current/sql-createsequence.html)
