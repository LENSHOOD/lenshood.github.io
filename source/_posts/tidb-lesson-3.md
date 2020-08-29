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
2. **TiDB CPU Profile**
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

   

### TiDB CPU Profile

由于采用多种测试手段（例如上一节课讲到的 sysbench、ycsb、tpcc）对 TiDB + TiKV 进行全方位的 CPU + IO + Memory 的 profiling 并对其结果进行整合分析是一项比较大的工程。

因此，本文对 Profiling 的 scope 做了限定，**只针对在 tpcc 测试方法下的 TiDB 的 CPU 使用情况进行 Profiling，并分析。**

在确定了方向后，我们开始动手：

1. 环境：

   前文已经提到过了，在 docker 虚拟的 centos7 下部署 1 pd + 1 tidb + 3 tikv

2. 测试方法：

   tpcc 100 warehouse，相关命令如下：

   ```shell
   # prepare base data
   > ./bin/go-tpc tpcc --warehouses 100 prepare
   
   # run tpcc
   > ./bin/go-tpc tpcc --warehouses 100 run --time 3m --threads 64
   ```

3. Profiling:

   采用在 TiDB 内已经开启的 go-pprof 进行数据采集，相关命令如下：

   ```shell
   # do 60 seconds profiling at tpcc prepare stage (tidb set pprof port as 10080)
   > curl http://127.0.0.1:10080/debug/zip\?seconds\=60 --output tpcc-w100-prepare.zip
   
   # do 60 seconds profiling at tpcc run stage
   > curl http://127.0.0.1:10080/debug/zip\?seconds\=60 --output tpcc-w100-run.zip
   
   # after unzip the downloaded zip package, use pprof tool to illustrate profile result
   > go tool pprof -http=:8080 {unziped dir}/profile
   ```

经过上述步骤，我们已经能够拿到在测试期间对 TiDB 的 CPU Profiling 数据了：

- prepare 期间的 CPU Profiling：

{% asset_img pprof-cpu-tpcc-prepare.png %}



- run 期间的 CPU Profiling：

{% asset_img pprof-cpu-tpcc-run.png %}



从上述结果中我们可以清晰地看出：

1. prepare 阶段 Parser 的 CPU usage 占比很大
2. run 阶段没有特别明显的 CPU usage 占比大的函数，其对 CPU 的消耗表现为整体平均
3. 底层 gc 相关的逻辑显著的占用了 CPU 资源

