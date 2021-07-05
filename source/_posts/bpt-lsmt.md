---
title: 对比 B+ Tree 与 LSM Tree
date: 2021-07-05 22:42:59
tags:
- b+ tree
- lsm tree
categories:
- DB
---

{% asset_img 1.png %}

B+ Tree 与 LSM Tree 是现今各类数据库中使用的比较多的两种数据结构，它们都可以作为数据库的文件组织形式，用于以相对高效的形式来执行数据库的读写。

本文简述了这两种数据结构的操作方式与操作开销，并对比了其自身的优缺点。

<!-- more -->

## B+ Tree

B+ Tree 是我们比较熟悉的一种数据结构，它以节点（node）为单位来存储数据，每个 node 都可以存放多个 k-v 键值对，多个 node 以类似一棵树的形式组成完整的 B+ Tree，node 与 node 之间由指针来连接。

首先我们应当知道的是：B+ Tree 中保存的数据是有序的，我们通过 B+ Tree，可以实施查询、顺序访问、插入、删除的操作。

下图所示的就是一颗 B+ Tree （省略了具体的 value 值，下同）：

{% asset_img 2.jpg %}

对于上图，我们有如下解释：

1. 关于 node：

   - node 中的 key 保持有序，从左至右依次增加。node 之间的 key 也保持有序，左边 node 的 key 小于右边 node

   - 红色 node 代表 internal node，即内部节点。internal node 包含的 k-v 键值对中，value 是指向下一个 node 的指针。在每一个 internal node 的最左边，还有一个指针，指向所有 key 都比本节点中最小 key 还要小的子节点。而 key 右边指针指向的子节点，其最小的 key 一定大于等于当前 key。
   - 蓝色 node 代表 leaf node，即叶子节点。leaf node 包含的 k-v 键值对中，根据具体实现的不同 value 可能代表存放实际数据的文件的偏移量，也可能直接就是实际数据（clustering index）。另外我们会发现，leaf node 本身还是一个双向链表，他们包含了前后指针。
   - B+Tree 是完全平衡的，即所有的 leaf node 都处于同一层级

2. 图上部的 `N=4` 代表了这一棵 B+ Tree 的度（degree）是 4（也称 N-way B+ Tree）。

   - degree 的意思是，这棵树中任意一个 internal node，它最多能指向多少个子节点，由于 internal node 最左边还有一个指针，因此一个 internal node 中最多只能包含 `N-1` 个 key
   - node 中不仅限制了最大 key 数量，同时也限制了最少的 key 应大于 `N/2 - 1`。即每个 node 都至少是 “半满” 的
   - 假如由于插入、删除等操作导致 node 中 key 的数量不满足要求，则必须对 node 进行 分裂（insert） 合并（delete）
   - root node 既可以是 leaf，也可以是 internal。root node 的节点数不受 N 的限制

### B+Tree 的操作

一切对 B+ Tree 的操作，无论如何都需要在操作完成后仍然满足 B+ Tree 的定义要求。

#### Search

由于 B+ Tree 的有序性，search 操作非常简单，只要从 root 开始，根据需要查询的 key 来比较大小，一层层查找，直到找到 leaf node，之后在 leaf node 中按序找到对应的 k-v pair。

{% asset_img 3.jpg %}

如上图紫色箭头所示的查找路径，我们期望找到 `key == 6` 的值，那么需要：

1. `6 > 5 (root_key)`，进入右边 pointer 指向的子节点
2. `6 < 7 (far left of internal node)`，进入中间层 internal node 最左边 pointer 指向的子节点
3. 当前子节点已经是 leaf node，进入其中，依次对比，最终找到 `k == 6` 的节点

我们发现，对于一颗高为 `h` 的 B+ Tree，最多只需要 `(h - 1) + (N - 1)` 次对比，就能找到（或确定不存在）值。

其中，

- `h - 1` 代表了查找 node 的次数，由于实际中磁盘存储的单位即是 node，因此 `h - 1` 可以看做是 `I/O` 次数
- `N - 1` 代表了在 node 内部查找，由于实际中 node 都会被读入内存，因此 node 内的查找开销几乎可以忽略不计

#### Sequential Access

由于 leaf node 中包含了指向下一个 leaf node 的指针，因此对于类似 `select * from tbl where i > 5`这样的 顺序访问，只要找到第一个满足条件的 leaf node 后，就可以直接通过尾指针来定位下一个 node。

假如是 `clustering index` （聚集索引）的实现，由于 node 之间的有序性，最佳情况下其数据在磁盘中的实际存储位置也是顺序存储的，那么顺序访问就会非常高效。

#### Insertion

#### Deletion

### B+ Tree 与 Buffer Pool

