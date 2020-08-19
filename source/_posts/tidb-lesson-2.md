---
title: TiDB 学习课程 Lesson-2
date: 2020-08-19 22:51:46
tags:
- tidb
categories:
- TiDB
---

这一节，我们需要对一个完整的 TiDB 集群进行性能测试，为了实现这一目标，本文会通过以下几个步骤来介绍如何从零开始实现 TiDB 的性能测试：

1. **借助 kind 在单机模拟集群**
2. **通过 TiDB Operator 部署 TiDB 集群到 K8S**
3. **根据机器硬件调整 TiDB 线程池配置**
4. **sysbench 测试**
5. **go-ycsb 测试**
6. **go-tpc 测试**
7. **性能瓶颈分析**

