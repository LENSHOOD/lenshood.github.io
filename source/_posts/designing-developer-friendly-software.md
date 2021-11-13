---
title: 浅谈对开发者友好的（developer-friendly）软件设计
date: 2021-11-07 17:00:01
tags:
- user interface
- developer-friendly
- software design
categories:
- Software Engineering
---

{% asset_img 8.png 200 200 %}

面向开发者用户的软件，相比普通用户仅在限定的场景下使用外，还可能会被集成、扩展、二次开发等等，因此在代码或设计层面也应该尽可能的考虑如何对开发者更友好。

本文从：

- Keep It Simple, Stupid
- Least Surprise
- Guide, Not Blame

三个不同的角度，结合实际案例，尝试阐述和讨论哪些设计是对开发者友好的。

<!-- more -->

## Keep It Simple, Stupid

用户想要我们的软件易用，易懂，易扩展。

开发者就需要从 API、设计、协作等多个方面确保简单，而简单很难。

{% asset_img 7.jpeg 400 300 %}



### 耐心与好奇心成反比

当我们尝试使用一种新的包、工具等等时，首先面临的就是如何引用、安装的问题。

我们会去主页看 README，但人的耐心通常很有限...

{% asset_img 1.png 439 463 %} 

上图是 Prometheus 的安装页，它不仅存在大段的文字，甚至还有配置文件，这潜在的给用户施加了不小的心理负担。

如果用户想要尝试，可能要专门找半小时空闲，鼓起勇气、正襟危坐的开始依照文档试验。

{% asset_img 2.png 490 381 %}

上图是 rustup 的安装页。

相比起来，一眼就能看到深色背景的命令，30 秒就可以在 shell 里面执行，那么任何人都可以近乎零负担的在本地快速搭建 rust 环境。

README 或者 Home 页是通常是用户第一次接触我们的软件的地方，怎么样抓住用户的好奇心的确需要仔细研究。



### 简洁就是美

简洁之美，体现在如何优雅的解决问题。

Golang 中启动一个 go-routine 的操作可谓极致简洁：

{% asset_img code-0.png %}

不需要 import 任何包，没有其他与之相关的 key word 要理解和记忆，甚至连对 go-routine 本身的引用都不给返回（怎么管理 go-routine 是另一个故事了）。正是这种简单易用的设计，使程序员想要启动一个 go-routine 时毫无负担。



### 我不需要我不需要的

>  *C++ implementations obey the zero-overhead principle: What you don't use, you don't pay for [Stroustrup, 1994]. And further: What you do use, you couldn't hand code any better.*
>
> *-- Stroustrup*

在 C++ 中，用户用不到的功能一定不会产生任何开销，而使用到的功能所产生的开销，也一定不会比用户自己去手写更高。

这一原则直截了当的给出用户十分明确的选择，零成本原则也是 C++能作为系统级语言的一个重要原因。

类似 C++、Rust 语言所提供的零成本抽象的特性（Trait、Future 等等），让对性能敏感的用户无需担心为了提升代码设计引入的抽象可能会导致额外的开销，这让用户可以更加有信心的进行代码抽象而不用担心性能问题。



### 约定大于配置

将环境、配置，以约定默认的方式自动设置，这样就减少使用者在最开始需要做出决定的数量，也就降低了上手难度和用户的心理负担。

Ruby on Rails 相对较早的实践了这一概念，并在其框架内应用了大量约定，来降低初学者的使用门槛以及提升专家的生产效率。

Spring Boot 甚至完全就是为了方便用户使用 Spring 框架而创造的。通过一系列的自动化配置、条件配置等方法，让用户只需要非常少量的配置（甚至零配置）就可以 “Just Run”。

而对于不同的使用场景下用户可能会选择不同的额外自定义配置项，这时候如何优雅的让用户只关心自己想要的配置呢？

##### Functional Options

当构建某个实体需要许多必选、可选的参数时，传统的两种办法：

- 全部作为传入函数，或每种参数写一个包装函数
- 传入一个配置类（结构）

上述方法都存在一些[问题](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)，更好的办法是以可变参数的形式进行配置。以创建 grpc server 为例：

{% asset_img code-1.png %}

这样的设计能够方便使用者灵活的选择想要的配置，甚至是自定义的配置项。不同的用户对配置的关注点可能不同，上述方式可以允许用户自由挑选自己想要的配置传入，而不用考虑自己不关心的配置项。



### 关注结果，不关注过程

如果允许用户直接描述 ta 想要的结果，那么用户就不必指定具体的工作过程了。



{% asset_img code-2.png %}



上述代码描述的是用 java 语言来实现 word count，先将单词映射为 (word, count) - pair，之后对相同的 word 进行聚合，最后得到结果。

这是过程式的办法。

而如果用 SQL 这种声明式的实现，可见下图。



{% asset_img code-3.png %}



SQL 语言只描述了用户想要的结果，至于获取这一结果中所要经历的过程，用户无需过问，也不关心。

在 K8S 的声明式 API 设计中，除了能灵活的描述结果状态以外，还能保证操作的幂等性，用户体验非常好。

显然，声明式 API 的抽象层次要比过程式 API 更高，但这也意味着声明式 API 更难实现。常见的声明式 API 的实现大都基于解决特定领域的问题，并不具备图灵完备性。



## Least Surprise

不要惊吓用户！

{% asset_img 6.jpeg %}

通常在某个特定的领域，人们会在领域上下文内形成一系列的惯例和常识，比如：

- 走路撞到墙，头会痛，但墙通常不会塌
- 在网页上填完表单按下提交按钮，页面会跳转
- 在命令后面追加 `--help` 通常会返回该命令的使用方法

因此，我们的软件所表现出的行为，应该尽量满足在其领域内具有一致性、显而易见、可预测。



### 单一控制来源

作为用户，通常期望软件能提供来源清晰，行为一致的配置，而如果有很多种不同的方式都能达到类似的配置效果，用户就会感到困惑，不知道应该用哪一个。

Spring 框架在发展了这些年后，由于其出色的灵活性设计，反过来也导致了一定程度的理解困难。

比如 Spring Security 中想要配置自定义的认证时，可以：



{% asset_img code-4.png %}



上面这三种方式都可以满足认证的要求，包括官方文档在内的诸多资料都会尝试使用其中的一种或两种方式来配置认证，如果用户对其设计原理不甚了解（比如刚刚上手），看到这么多种不同的配置方法，就很容易会产生不解与慌乱。



### 无二义性

某些情况下，用户在使用我们的软件时必须要对某些配置进行设定。从用户的角度看，对于配置项，用户期望的是最好能一眼就看出来该配置的内涵是什么，假如配置项存在二义性，就会让用户摸不着头脑。

这里引用一个[讨论 TiDB 可交互性文章](https://mp.weixin.qq.com/s/WEO1y8vg21CXlix8wO28hw)中的例子：

在 TiDB 5.0 版本中引入了一个配置开关：

`tidb_allow_mpp = ON|OFF (default=ON)`

这个开关项的本意是如果设置为 OFF，则禁止优化器使用 TiFlash 来执行查询，而假如设置为 ON，那么优化器会根据实际情况自行选择是否使用 TiFlash 。

所以虽然配置的是 ON，但其实到底有没有用 TiFlash，还得看优化器的判断。*”就像是房间里控制灯光的开关，关掉时灯一定不会亮，而打开后灯却不一定会亮“*。这种二义性开关的存在，容易让用户误解、会错意。

面对上述问题，文中给出的修改建议是，修改为：

 `tidb_allow_mpp = ON|OFF|AUTO`

多了的这个 AUTO 确实能让用户一目了然。



### 遵循惯例

有很多设计上的、语言层面的或是领域方面的惯例和规范，通常软件开发者们都会默认去遵循这些惯例和规范。



{% asset_img code-5.png %}



这里引用了重构 2 中查询和修改分离的例子，某些时候方法命名甚至直接省略了后面，变成 `getTotalOutstanding()`。

通常遇到以 `getXXX` 开头的函数，用户大都会默认该函数具有幂等性，假如使用后发现调用动作竟然产生了某些副作用（比如这里是每调用一次都会发送一次账单），就会让用户费解。（Rust 很棒的一点就是当发现 `get_xxx(&mut self)` 这种方法定义时会自动高亮警告 ）

另有一例：

{% asset_img code-6.png %}

通常类似上述的 “移动” 操作，都是 from / src 在前，to / target 在后，而如果我们的函数是反过来的，就是在坑用户了。

不过，考虑到上述操作的两个参数都属于同一 string 类型，我们没办法限制用户一定会按照先 from 后 to 的形式传参（也许用户钟爱 intel 汇编语法？），那么更好的方式是：



{% asset_img code-7.png %}



## Guide, Not Blame

RTFM，是老人对新人的谆谆教诲？还是软件作者对伸手党的有声控诉？

{% asset_img 5.jpeg %}

每当我们看到用户报告的错误显示 `Http Code 400` 时是否都一阵窃喜？

{% asset_img 4.png %}

“用户错误” 是用户自己的问题，与开发者无关，是这样吗？



### 报错了，然后呢？

当用户执行了误操作后，我们的软件理应将详细的错误信息反馈给用户，但除此之外，能做的还有很多：



{% asset_img code-8.png %}



上面展示的是 Rust 编译器的编译报错，从上到下分别是：

1. 告诉我们错误原因是 “缺少生命周期标志”，错误码是 E0106
2. 指出是 “linear_probe_hash_table.rs” 文件的第 17:26 个字符出错
3. 又用剪头指明了代码错误的位置
4. “help” 部分告诉我们 “可以考虑使用 ``a` 符号”，最后用波浪线给出了改正后的结果

有人说写 Rust 是 “compiler-driven development”，从编译器这种保姆级的报错信息来看，确实所言不虚。



### 帮助用户识别而非记忆

在一些较复杂、步骤较多的配置操作后，最终执行前用户心里可能没底，我们的软件应该帮用户检查并识别问题（即类似 dry-run 的能力），从而降低错误发生的概率。

我们知道 Terraform 的工作流是 `Write -> Plan -> Apply`。

在编写完成 tf 文件和执行操作之前的 `Plan` 阶段就是用于告知客户接下来将要执行操作的执行计划，以及可能产生的影响。

```shell
$ terraform plan
An execution plan has been generated and is shown below.
 
Resource actions are indicated with the following symbols:
  + create
 
Terraform will perform the following actions:
 
  # aws_ebs_volume.iac_in_action will be created
  + resource "aws_ebs_volume" "iac_in_action" {
        + arn               = (known after apply)
        + availability_zone = "us-east-1a"
        + encrypted         = (known after apply)
        + id                = (known after apply)
        + iops              = 1000
        + kms_key_id        = (known after apply)
        + size              = 100
        + snapshot_id       = (known after apply)
        + tags              = {
                + "Name" = "Terraform-managed EBS Volume for IaC in Action"
            }
            + type              = "io1"
    }
 
Plan: 1 to add, 0 to change, 0 to destroy.
```

{% asset_img code-9.png %}

`Plan` 会根据当前资源的状态和用户期望状态作对比，给出执行计划，而不会对系统产生任何实质影响。假如用户发现执行计划中与其预期不符，就可以回过头去重新修正。



### 交互式文档

虽然本文最开始提到，用户可能只会花 30 秒来浏览文档，但真正到深入使用我们的软件时，看文档是必须的。

传统的文档看起来不仅枯燥，而且由于缺少反馈，用户很难记住文档要传达的知识。

{% asset_img 3.png %}

上图展示的是 Arthas 提供的交互式文档（学习课程），通过在线的 ”playground + 引导用户完成任务” 的形式，加强反馈，按阶段给予奖励，可以很好的提升体验。



## 结语

本文主要讨论了构建开发者友好的软件需要包含的三点要素，并通过一些事例佐证了这些要素本身的必要性。

综上来看，我们认为对开发者体验友好的软件：

- 首先，应该在设计和交互上尽量保持简单，做到易用、易懂、易扩展。
- 其次，也应该遵循一些常识和领域内的惯例，从而避免在使用中让用户产生困惑。
- 最后，应该尽量引导用户做出正确的操作，同时降低试错成本改善学习体验。



## Reference

