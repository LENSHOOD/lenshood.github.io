---
title: 震惊！一行注释竟能让 lfring 性能提升一倍！
date: 2022-08-01 23:31:32
tags: 
- lock free
- ring buffer
- performance optimization
categories:
- Go
---

{% asset_img ring.png 500 %}

本文介绍了对笔者先前的 go 库 [go-lock-free-ring-buffer](https://github.com/LENSHOOD/go-lock-free-ring-buffer) （简称 lfring）的性能优化。
介绍该 lfring 的文章可见[这里](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/)。

<!-- more -->

