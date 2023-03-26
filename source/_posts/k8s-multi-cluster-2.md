---
title: Kubernetes 多集群（下）：方案与演进
date: 2023-03-26 13:19:34
tags:
- multi-cluster
- multi-cloud
- k8s
categories:
- Kubernetes
---

{% asset_img header.jpg 500 %}

本文（分上下两部分）介绍了 K8s 多集群的由来以及实现多集群所面临的核心问题，之后分析并探讨了现有的 K8s 多集群方案，最后根据目前实现方案的痛点与挑战，设想了未来的演进趋势。

本篇是下半部分，主要讨论目前实现 K8s 多集群的开源方案、对现状问题的讨论以及可能的演进方向。

<!-- more -->

## 1. 几种方案对比

### 1.1 KubeFed

### 1.2 Karmada

### 1.3 OCM

### 1.4 Rancher

## 2. 演进趋势

1. 真的需要扁平网络吗？（跨集群 pod 同一个网络）
2. 跨集群方案怎么解决数据同步问题
3. 如何统一管理传统云资源，如 VM，块存储，VPC 等
