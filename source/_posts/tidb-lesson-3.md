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
1. 给出 Dockerfile：
  
  ```dockerfile
    ## 使用 centos7 作为基础镜像
    FROM centos:centos7

    ## 安装 tiup
    RUN /bin/bash -c 'curl --proto '=https' --tlsv1.2 -sSf https://tiup-mirrors.pingcap.com/install.sh | sh'

    ## 设定环境变量
    ENV PATH /root/.tiup/bin:$PATH

    ## 安装 tiup playground 所需的所有组件
    RUN /bin/bash -c 'tiup install playground | tiup install prometheus | tiup install pd | tiup install tikv | tiup install tidb | tiup install grafana'

    ## 设定 entrypoint 启动 playground 集群，host 映射到 0.0.0.0
    ENTRYPOINT tiup --tag=local-tidb-cluster playground --db=1 --kv=3 --pd=1 --tiflash=0 --monitor --host=0.0.0.0
  ```

  

2. 启动容器环境：
   docker-compose.yaml 如下所示，

   ```yaml
   version: '2.0'
   services:
     tiup-playground-cluster:
       build:
         context: .
         dockerfile: Dockerfile
       ports:
       - "4000:4000"
       - "2379:2379"
       - "9090:9090"
       - "3000:3000"
       - "10080:10080"
   ```
   ```shell
   # 启动容器
   > docker-compose up -d
   ```

3. 通过[上一次课程作业](https://lenshood.github.io/2020/08/19/tidb-lesson-2/)提到的压测工具，对部署的 TiDB 集群进行测试

   


