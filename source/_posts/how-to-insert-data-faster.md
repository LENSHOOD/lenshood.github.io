---
title: 如何让插入数据速度提升200倍
date: 2023-06-11 14:50:45
tags:
- prepare data
- SQL
categories:
- DB
---

本文介绍了在一次构造测试数据的活动中，如何通过优化手段将数据插入速度提升200倍。

<!-- more -->

在帮助某客户构建业务系统过程中，我们需要对关键业务流程涉及的数据库表进行性能预估，并给出构建索引和分区的建议。

通过对历史业务和业务增长率的调查分析，业务表的热数据在一亿条左右。这意味着为了测试性能我们需要构造 1 亿条模拟业务数据，这些数据在关键业务字段上需要能反映实际业务特点，包括数据类型和数据分布。

客户先前曾经构造过同样规模的测试数据，通过在代码中循环批次插入数据，构造一亿条数据大约需要 6 小时。据此我们粗略的认为这种方案引入了不必要的网络 IO，效率太差，也许通过存储过程构造会更快。

### 实现基本功能

对构造数据的需求进行分析后，我们需要完成如下的任务来实现功能：

1. 构造业务数据集
   - 预置业务数据集，按数据分布的实际比例进行推算，预置的数据集的规模远小于一亿
   - 对于唯一性业务数据（主要是流水号类）创建序列生成器生成唯一数值
2. 编写存储过程
   - 通过随机函数从业务数据集中选取数据
   - 拼接 INSERT 语句，批次插入
   - 循环上述过程

据此可以得到如下存储过程示例：

```sql
-- 数据库：OpenGauss

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
            sql_insert := sql_insert || format('INSERT INTO t0 (A, B, C) VALUES (%s, %s, %s);', a_value, quote_literal(b_value), quote_literal(c_value));
        END LOOP;
        
        EXECUTE sql_insert;
        
    END LOOP;
END;
```

执行后我们发现，上述存储过程慢的令人发指，估算插入一亿条数据耗时 180 小时，比代码插入慢了 30 倍。

这显然存在明确的性能优化空间。

### 实施优化

一般在实施性能优化之前，会先进行类似 “性能建模” 的分析过程，寻找可能的性能劣化点，再决定从哪里开始。

由于脚本比较简单，我们很容易发现一些可能的劣化点：

- `nextval()` 函数及 Sequence 的性能
- `random()` 函数
- `ORDER BY random()`
- FOR 循环

很简单，对于函数、语句类，我们只要依次测试就能发现最耗时之处。

#### 优化执行计划

1. order by 执行计划

   ```SQL
   SELECT b_column FROM t1 ORDER BY random() LIMIT 1
   Limit
     -> Sort
       Sort Key: (random())
         -> Seq Scan on xxxxx
     
   SELECT b_column FROM t1 WHERE id in ( SELECT floor(random()*10000) ) LIMIT 1
   Index Only Scan
     Index Cond
     Init Plan1
       -> Result xxx

2. 挪到外循环

   

#### 资源利用最大化

基于第一次分析，我们发现除了`ORDER BY`以外，其他几部分的耗时都很短，优化的意义不大。因此我们跳出耗时，转向资源利用率。

对于数据库的插入性能，主要的指标关注点在于 **IO 写入速率**和 **CPU 使用率**。IO 写入速率越高说明单位时间插入的数据越多，CPU 使用率越高证明执行引擎跑的越满（等待 IO 时间少）。

通过监控发现，在存储过程执行期间，IO Write 约 20 Mbps，CPU 使用率 10~20%，基于此可以尝试并发执行。我们并不清楚数据库的硬件配置，无法根据 CPU 核数来划定线程数范围，因此进行了多轮测试，最后得出结论 8 个线程下加速比最高。

经过测试，8 线程并发插入，插入速度同比又提升了 4 倍，并且 CPU 使用率达到 75%，IO Write 约 85 Mbps。

依赖和竞争是限制并发性能的两大障碍，进一步分析上述脚本，我们会发现多个并发运行的存储过程，都依赖同一个 Sequence `s0` 来生成序列数字。



#### 减少不必要的操作

经过上一轮优化后，CPU 使用率已达 96%，IO Write 则达到 115MiB/s。我们虽然无法判断 IO 带宽被利用了多少，但可以明确 CPU 已经基本被打满。在这种情况下想要进一步优化，不妨将视线移至 CPU 计算。

我们知道，数据库执行一次操作，在计算层大致需要经过解析 SQL，逻辑和物理优化、形成执行计划这几个过程，之后会在存储层执行各类算子。计算层主要是 CPU 计算，在存储层更多是 CPU 和 IO 操作混合。

容易发现，对变量的赋值操作都是采用 `SELECT xxx INTO var` 的模式，每个变量的赋值都要经历一遍解析的过程。因此我们尝试将这种赋值操作修改为 `var := {value conpute}` 的形式，对变量直接赋值。

经过以上修改，CPU 利用率下降了 2%，IO Write 提升了 17%。



## 性能分析的思考

本文所描述场景，上下文清晰，依赖少，很适合用来讨论性能分析和优化的实践方法。

### 了解系统架构

对系统的了解程度直接决定了进行性能分析时定义问题的范畴。

通过系统架构图，我们能了解被分析系统用于解决哪一类问题，是通过怎样的方式实现的，以及其核心业务符合怎样的服务模型。

举例如下：

1. 系统提供的是计算类服务还是存储类服务？据此大致判断需要关注的指标。
   - 计算类：如视频编解码，游戏引擎等，关注 CPU 使用率、线程调度、并行度
   - 存储类：如数据库系统，关注请求吞吐量、网络带宽、IO
2. 各子模块之间的交互关系是什么？据此了解模块间的边界。
   - 函数调用还是IPC 
   - 是否存在缓冲队列



### 建立合理的模型



### 观测洞察的重要性

