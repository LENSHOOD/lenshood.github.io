---
title: 译文: 一种易于理解的共识算法的研究（In Search of an Understandable Consensus Algorithm）
date: 2021-09-13 17:57:50
tags: 
- raft
categories:
- Software Engineering
---

> 本文是对 raft paper 的翻译，原文请见 https://raft.github.io/raft.pdf

## 摘要

Raft 是一种用于管理 log 复制的共识算法。它采用了与 Paxos 不同的算法结构，但能产出与 (multi-)Paxos 等价的结果，并且与 Paxos 一样高效；这令 Raft 成为了比 Paxos 更易于理解且能为构建实际系统提供更好的基础的一种算法。为了更加易懂，Raft 将共识算法中的几个关键元素（比如 leader 选举，log 复制，安全等）互相分离，并通过使用更强的一致性要求来减少必须要考虑的状态（state）的数量。用户调研的结果显示 Raft 相比 Paxos 而言更容易学习。同时，Raft 还包含了一种新的通过覆盖大多数来确保安全的机制来修改集群成员。

<!-- more -->

