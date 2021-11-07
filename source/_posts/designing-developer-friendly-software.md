---
title: 浅谈程序员友好型（developer-friendly）软件的设计
date: 2021-11-07 17:00:01
tags:
- user interface
- developer-friendly
- software design
categories:
- Software Engineering
---

前言

<!-- more -->

## KISS

### 简洁

1. 设计简单：go routine
2. 引入简单：tiup
3. 使用简单：spring boot build image
4. 应对选择恐惧症：零成本抽象 zero overload abstraction



### 灵活

1. 约定大于配置：尽量默认配置（jeager client，简单的初始化，all in one 服务端）
2. 可变参数化配置：https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
3. 过程控制变声明控制：sql，java lambda



### 易懂

1. api 易懂性：hateoas
2. 与真实世界映射降低认知成本 k8s



## Least Surprise

### 单一的控制来源

1. 反例，spring boot 配置，难用：，很多不同的方式都可以达到同样的结果

### 统一语言

1. k8s yaml

### 无二义性

1. tidb mpp 配置， on off auto，（可观测性与可交互性 by 黄东旭）

### 遵循约定

1. go ctx 并发



## Guide, not Blame

### 错那了，怎么办？

1. rust compiler 报错信息

### 清晰准确的文档

1. 可测试的文档： rust doc

### 协助用户记忆

1. terraform init plan apply
2. dry-run

#### 交互式体验

1. 反馈
