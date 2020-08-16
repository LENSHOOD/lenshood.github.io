---
title: TiDB 学习课程 Lesson-1
date: 2020-08-14 20:51:26
tags:
- tidb
categories:
- TiDB
---

今年了解到 TiDB 这样一个 HTAP 数据库，在 Github 上面热度超高，再加上国人开发与维护，因此产生了很大的兴趣。

正好社区搞了一个《高性能 TiDB 课程学习计划》通过 12 节课来使学员能达到 Committer 的水平，因此就报名参加啦。

第一节课的课程作业是*尝试在本地启动一个最小集群，并在开启事务时打印 “hello transaction” 并将这一过程写成一片文章*，因此本文将会主要描述这一过程。

### 简单介绍 TiDB 架构

TiDB 是一款支持 SQL 语义并直接兼容 MySQL 通信协议的分布式关系型数据库。可以看到这句话里面包含了三个很重要的信息：

- 兼容 MySQL
- 关系型数据库
- 分布式

假如作为一个业务开发者，从外部来看，TiDB 在绝大多数情况下直接表现为一个 MySQL，业务中的各种 Select 查询，表关联，以及对事务、ACID 的要求，它都能满足。在此基础上，通过分布式实现了方便的动态扩缩容，以及高性能。

那么 `MySQL + 分布式` 这样一个在传统互联网应用中想要而不可得的东西，TiDB 是如何实现的呢？请看下图（转载自 PingCAP 官网，一切版权归其所有）：

![](https://download.pingcap.com/images/docs-cn/tidb-architecture-1.png)

整个数据库集群分为三类模块：

- tikv：分布式 K-V 存储，通过 Raft 协议保证强一致性实现多副本的高可用，是数据的实际存储位置
- tidb：接受客户端连接，解析 SQL 语句，并转换为访问 tikv 的计算，最后收集数据并返回
- pd：placement driver，管理整个集群的元信息，并负责 tikv 节点的调度。

上述模块全都都可以多节点部署。

了解了总体架构后，我们发现，外表上看似乎是一台 MySQL，实际上除了对 MySQL 协议的支持以外，其内部结构已经和 MySQL 没有太大关系了。相比之下，TiDB 可能更像一个 Spark 集群，因此其实 TiDB 就是一个根正苗红的分布式数据库方案。

### 从源码开始部署

如果只谈部署的话，其实 TiDB 提供了一个非常方便的集群部署工具：TiUP，在[《TiDB in action》的安装部署章节中](https://book.tidb.io/session2/chapter1/tiup-tiops.html)，非常清楚的描述了 TiUP 的使：不论是部署单机测试环境还是部署集群，几条命令就能全部搞定了。

不过在我们当前的上下文中，是想要从源码开始部署的，所以，接下来我会暂时先抛开 TiUP，直接从源码开始。

##### 1. 安装 Go 环境

三大件中，tidb 和 pd 都是基于 go 开发的，所以我们首先需要安装 Go 语言的环境：

- 方法一 ：[官网](https://golang.org/)直接下载安装包安装即可。
- 方法二：`brew install golang`，但 brew 目前安装的版本是 1.14.7，实际上 1.15 已经发布了。
- 方法三：安装一个 GoLand， 在里面直接安装

安装完成后需要配置 `GOROOT` 和 `GOPATH`两个环境变量，详情可以见[官网](https://golang.org/doc/install)。

##### 2. 学习 Golang

因为先前没有接触过 Go 语言，所以先花了两天时间简单学习了 Go，从 java 切换过来入门不算太难。推荐[官网教程](https://tour.golang.org/welcome/1)。

##### 3. 编译运行 tidb

直接 clone master 分支到本地，我们可以看到源码中提供了 Makefile，default target 会以 `tidb_server/main.go`作为入口进行编译。编译完成后的 binary 会输出在 `bin` 目录，可以直接通过`./bin/tidb-server`执行。

##### 4. 编译运行 pd

##### 5. 编译运行 tikv

##### 6. 启动一个最小集群：one tidb + one pd + three tikv

### 开启事务时打印 “hello transaction” 

##### 1. tidb 执行 SQL 命令的总体步骤

{% asset_img tidb-execute-cmd.png %}

##### 2. `func (s *session) Txn(active bool) (kv.Transaction, error)`

##### 3. Executor

##### 4. 运行结果