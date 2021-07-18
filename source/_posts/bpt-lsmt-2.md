---
title: 对比 B+ Tree 文件组织 / LSM Tree 文件组织（第二篇：LSM Tree）
date: 2021-07-18 22:54:52
tags:
- b+ tree
- lsm tree
categories:
- DB
---

{% asset_img 1.jpg %}

B+ Tree 与 LSM Tree 是现今各类数据库中使用的比较多的两种数据结构，它们都可以作为数据库的文件组织形式，用于以相对高效的形式来执行数据库的读写。

本文简述了这两种数据结构的操作方式与操作开销，并对比了其自身的优缺点。

<!-- more -->
