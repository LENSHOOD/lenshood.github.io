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

如今这个开源软件盛行，每个小领域都有好几种不同的解决方案的时代，简洁的设计，不仅能让使用者眼前一亮，迅速被吸引，还能降低认知上的负担，使开发者乐于尝试体验。

##### 零负担使用

Golang 中启动一个 go-routine 的操作可谓极致简洁：

```go
go run()
```

不需要 import 任何包，没有其他与之相关的 key word 要理解，甚至连对 go-routine 本身的引用都不给返回（怎么管理 go-routine 是另一故事了）；正式这种非常简单易用的设计，使程序员在 golang 中启动一个 go-routine 毫无负担。

##### README 的寿命只有 10 秒

当我们尝试使用一种新的包、工具等等时，首先面临的就是如何引用、安装的问题。

我们会去主页看 README，但人的耐心通常很有限...

比如某人想要尝试 Prometheus，TA 会找到如下的页面：

{% asset_img 1.png %}

要说这个页面里虽然字儿多，但几处重点也都用不同的颜色突出了，主要一上来就列这么一篇配置文件让人有点生畏。

而假如某人想要装个 rust？

{% asset_img 2.png %}

只要执行一下一眼就能看到的深色背景的命令，一切就只剩下等待了。

##### 别让我权衡

> C++ implementations obey the zero-overhead principle: What you don't use, you don't pay for [Stroustrup, 1994]. And further: What you do use, you couldn't hand code any better.
>
> -- Stroustrup

类似 C++、Rust 语言所提供的一些零成本抽象的特性（Trait、Future 等等），让对性能敏感的用户无需担心为了提升代码设计引入的抽象可能会导致额外的开销，这让用户可以更加有信心的进行代码抽象而不用担心性能问题。



### 灵活

##### 约定大于配置

将环境、配置，以约定默认的方式自动设置，这样就减少使用者在最开始需要做出决定的数量，也就降低了上手难度和用户的心理负担。

Ruby on Rails 相对较早的实践了这一概念，并在其框架内应用了大量约定来，来降低初学者的使用门槛，和提升专家的生产效率。

Spring Boot 甚至完全就是为了方便用户使用 Spring 框架而创造的。通过一系列的自动化配置、条件配置等方法，让用户只需要非常少量的配置（甚至零配置）就可以 “Just Run”。

##### Functional Options

当构建某个实体需要许多必须、可选、默认参数时，传统的两种办法：

- 全部作为传入函数，或每种参数写一个包装函数
- 传入一个配置类（结构）

上述方法都存在一些[问题](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)，更好的办法是以可变参数的形式进行配置，以创建 grpc server 为例：

```go
// default
svr := grpc.NewServer()

// with configs
s := grpc.NewServer(
		grpc.ConnectionTimeout(30 * time.Second),
		grpc.MaxRecvMsgSize(1024),
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())))

// build new server
func NewServer(opt ...ServerOption) *Server {
	opts := defaultServerOptions
	for _, o := range opt {
		o.apply(&opts)
	}
  ... ...
}
```

这样的设计能够方便使用者灵活的选择想要的配置，甚至是自定义的配置项。

##### 声明式 API

描述你想要的结果，而不是告诉我要怎么做。

```scala
val textFile = sc.textFile("hdfs://...")
val counts = textFile.flatMap(line => line.split(" "))
                 .map(word => (word, 1))
                 .reduceByKey(_ + _)
counts.saveAsTextFile("hdfs://...")
```



```sql
CREATE TABLE FILES (line STRING);

LOAD DATA INPATH 'docs' OVERWRITE INTO TABLE FILES;

CREATE TABLE word_counts AS
SELECT word, count(1) AS count FROM
(SELECT explode(split(line, ' ')) AS word FROM FILES) w
GROUP BY word
ORDER BY word;
```

声明式 API 的抽象层次显然要比命令式 API 要高，但这也意味着声明式 API 通常更难以实现。

常见的声明式 API 的实现大都基于解决特定领域的问题，而不具备图灵完备性。但即使是能解决限定领域内的所有问题，也不容易设计与实现。



### 易懂

我们的软件不只是会被使用，它还会被不断地迭代、修改、更新。因此从设计层面讲，软件的易懂性就很重要，这体现在：

- 清晰的架构设计
- 明确的抽象层级
- 依赖管理



## Least Surprise

不要惊吓用户！

通常在某个特定的领域内，人们会在领域上下文内形成一系列的惯例和常识，比如：

- 走路撞到墙，头会痛，但墙通常不会塌
- 在网页上填完表单按下提交按钮，页面会跳转
- 在命令后面追加 `--help` 通常会返回该命令的使用方法

因此，我们的软件所表现出的行为，应该尽量满足在其领域内具有一致性、显而易见、可预测。



### 单一控制来源

用户通常期望我们的软件能提供来源清晰，行为一致的配置，如果有很多种不同的方式都能达到类似的配置效果，用户就会感到困惑。

Spring 框架在发展了这些年后，因其出色的灵活性，反过来也会导致用户的理解困难。

比如 Spring Security 中想要配置自定义的认证时，可以：

```java
// 1. 自定义一个 UserDetialService 的实现，基本的方式
@Bean
public CustomUserDetailsService getUserDetailsService() {
    return new CustomUserDetailsService();
}

// 2. 自定义一个 AuthenticationProvider 的实现，更灵活
@Bean
public CustomAuthenticationProvider getAuthenticationProvider() {
    return new CustomAuthenticationProvider();
}

// 3. 自定义一个 OncePerRequestFilter 的子类，不太符合 Spring Security 的设计初衷，但也能用
@Bean
public CustomFilter getCustomFilter() {
    return new CustomFilter();
}
```

上面这三种方式都可以满足认证的要求，包括官方文档在内的诸多资料都会尝试使用其中的一种或两种方式来配置认证，如果用户对其设计原理不甚了解（比如刚刚上手），看到多种不同的配置方法，就很容易会产生不解与慌乱。



### 统一语言

1. k8s yaml

### 无二义性

某些情况下，用户在使用我们的软件时必须要对某些配置进行设定。从用户的角度看，对于配置项，用户期望的是最好能一眼就看出来该配置的内涵是什么，假如配置项存在二义性，就会让用户摸不着头脑。

这里引用一个讨论 TiDB 可交互性文章中的例子：

在 TiDB 5.0 版本中引入了一个配置开关：`tidb_allow_mpp = ON|OFF (default=ON)`，该配置的本意是如果设置为 OFF，则禁止优化器使用 TiFlash 来执行查询，但假如设置为 ON，那么优化器会自行选择是否使用 TiFlash 。

所以虽然配置的是 ON，但其实到底有没有用 TiFlash，还得看优化器的判断。就像是房间里控制灯光的开关，关掉灯一定不会亮，而打开后灯却不一定会亮。

引文中给出的修改建议是：`tidb_allow_mpp = ON|OFF|AUTO`，多了的这个 AUTO 让用户一目了然。

### 遵循约定

1. go ctx 并发



## Guide, not Blame

### 错那了，怎么办

1. rust compiler 报错信息
2. hateoas

### 清晰准确的文档

1. 可测试的文档： rust doc

### 帮助识别而不是记忆

1. terraform init plan apply
2. dry-run

#### 交互式体验

1. 反馈
