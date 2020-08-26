---
title: TiDB 学习课程 Lesson-3
date: 2020-08-26 22:59:34
tags:
- tidb
categories:
- TiDB
---

本节课程作业，我们会采用多种 profile 工具来对 TiDB 的性能进行分析，寻找其性能瓶颈，并根据分析结果来给出优化建议。

主要内容如下：

1. **TiUP 部署最小集群**
2. **CPU Profile**
3. **IO Profile**
4. **Mem Profile**
5. **性能瓶颈分析与优化建议**



### TiUP 部署最小集群
为了确保与[官方文档中建议的环境](https://docs.pingcap.com/zh/tidb/stable/hardware-and-software-requirements)一致，我选择在 docker 容器中启动 centos7 环境，再借助 TiUP 部署本地集群。
1. 启动容器环境：
   docker-compose.yaml 如下所示，
   
   ```yaml
   version: '2.0'
   services:
     tiup-playground-cluster:
       image: centos:centos7
       ports:
       - "127.0.0.1:4000:4000"
       - "127.0.0.1:2379:2379"
       - "127.0.0.1:9090:9090"
       - "127.0.0.1:3000:3000"
       volumes:
       - ./inner-tiup-mount:/root/.tiup
       tty: true
   ```
   ```shell
   # 启动容器
   > docker-compose up -d
   
   # 进入容器
   > docker exec -it {container-id} /bin/bash
   ```
2. 在容器内安装 TiUP 组件：

   ```shell
   curl --proto '=https' --tlsv1.2 -sSf https://tiup-mirrors.pingcap.com/install.sh | sh
   ```
   
3. 通过 `tiup playground` 来快速搭建本地集群（注意配置 host）：
	```shell
	tiup --tag=local-tidb-cluster playground --db=1 --kv=3 --pd=1 --tiflash=0 --monitor  --host=0.0.0.0
	```

	
	
	
	