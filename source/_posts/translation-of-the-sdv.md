---
title: （翻译）软件定义汽车
date: 2024-04-29 21:22:05
tags:
- sdv 
categories:
- Software Defined Vehicle
---

{% asset_img header.jpg 500 %}

本文翻译自 Gregor Resing（IBM）的文章 [The Software Defined Vehicle](https://www.ibm.com/blogs/digitale-perspektive/2023/06/the-software-defined-vehicle/)。

<!-- more -->

# 摘要

汽车行业正从内燃机向电动汽车转型，同时从以硬件为中心的产品向以支持自动驾驶和 OTA 功能的软件为主导的产品转型。这一革命性变化需要基于硬件与软件分离的新型电气/电子（E/E）和软件架构。

汽车开发和工程流程中的组织变革需要从孤立的组织转变为敏捷和基于 DevOps 的流程，以解决跨职能协作并提高开发速度和灵活性。

这种革命的基础是软件定义汽车（SDV）和集成 DevOps 方法，其中包括将软件更新到车辆中。

本文介绍了 SDV 的原理和它可以解决的挑战，最重要的是，它涵盖了真正的端到端流程的 DevOps 方法。这不仅可以为 OEM 带来标准车辆方面的新商机，还可以为共享出行和更智能环境的整个生态系统带来新商机。

# 关于本文

在本文中，我们介绍了 SDV 硬件和软件概念以及相关的架构变化。我们将汽车行业中的软件定义变化与其他行业进行了比较。

该文件的重点是汽车，主要是乘用车，但也适用于卡车、公共汽车和摩托车。当使用汽车一词时，重点是乘用车。

借助 SDV，汽车 OEM 将成为软件公司。开发和运营流程将转向基于 DevOps 原则的更敏捷的流程。我们描述了如何将 DevOps 原则应用于整个车辆生态系统和基于云的服务，以开发和运营未来的车辆。汽车 DevOps 包括安全性、后端车辆服务和无线更新、持续集成和部署流程、容器化、混合集成测试的使用以及软件和配置管理。

向 SDV 的过渡需要 OEM 组织进行根本性变革，从孤立的面向领域的部门转变为更具跨职能和敏捷性的组织。这些变革以及 OEM 和供应商协作方面的变革简要总结如下。

SDV 需要不同的架构变革。我们为车载架构、车辆的边缘连接、OTA，以及对人工智能（AI）和机器学习 ML 解决方案的容器化测试及包括混合测试的持续集成/部署提供了一些参考架构。

在文档的最后，我们评估了 SDV 的市场机会并就 IBM 如何应对这个不断增长的市场提出了一些建议。

# 介绍

在过去的几十年中，软件与硬件的分离颠覆了多个行业，例如个人电脑行业、移动和电信行业，这使得可更新和升级的产品以及新的商业模式成为可能。除了更灵活的基于软件的产品外，这些产品还始终连接到互联网以提供在线信息，并消费和生成大量数据，这些数据可以通过新的商业模式货币化。

在汽车行业，功能安全至关重要。制动系统需要最高的可靠性才能确保乘客安全。与消费者希望频繁更新和添加新功能汽车其他区域（如信息娱乐系统）相比，更换制动系统的灵活性较低。汽车行业面临着不同新要求的挑战：

- 客户希望在车辆上获得与移动设备相同的用户体验。这需要将应用程序和 Android 和 iOS 等移动操作系统集成到汽车中。
- 汽车需要提供不同级别的自动驾驶，这需要人工智能（AI）和机器学习技术（ML）。
- 环境法规和气候变化导致汽车走向电动汽车。各国的法规各不相同。
- 共享出行概念以及私人交通和公共交通的结合提供了新的商业模式。
- 始终互联的车辆可以与基础设施保持持续通信，并为汽车 OEM 带来新的售后机会。

汽车行业的这些变化通常被称为 ACES（自动驾驶、网联汽车、电动汽车和共享出行），需要对车辆和后端架构进行根本性的变革。这些创新将全部基于提供强大计算资源的电气/电子架构的软件中实现。软件定义了未来的汽车。商业价值的创造从硬件转变为软件。

*软件定义车辆（SDV）将硬件与软件分离，可更新和升级、自主、学习、始终连接、与环境交互，并支持基于服务的商业模式。*

## 其他行业中的软件定义（Software-defined）

在电信行业，软件定义网络（SDN）的引入将数据平面与控制平面分离，从而实现了灵活的应用程序。通过这种分离，网络管理通过软件配置而不是硬件配置（如防火墙规则、网络地址等配置）得到简化。

SDN 提供可编程的中央网络流量控制，无需任何手动配置。这种硬件和软件的分离以及灵活服务的配置与汽车行业对软件定义汽车的努力相媲美。

纵观个人电脑和智能手机行业，可以发现一些相似之处，这些相似之处可以转移到汽车行业。在个人电脑和智能手机行业，以应用程序形式出现的软件正在基于软件平台创建一个生态系统，该平台将硬件与底层软件分离开来。软件平台从硬件细节中抽象出来，并提供一组一致的 SDK 或 API，可供应用程序开发人员使用。

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-1.png" alt="img" style="zoom:150%;" />

图 1：解耦、抽象、标准化和开源为应用程序提供了生态系统

*硬件与软件的分离以及软件抽象层和平台的引入是电信、个人电脑和智能手机行业的基本概念，可应用于汽车行业。*

成功的软件平台适用于多代硬件。较新的软件平台版本可以安装在较旧的硬件上。在这种情况下，一些需要新硬件的功能将无法在安装了新软件平台的旧硬件上使用，但最新的应用程序仍可在旧硬件上运行。

Android Automotive 为汽车信息娱乐领域提供了一个软件平台。在其他汽车领域，如 ADAS/AD、车身、动力系统等，汽车 OEM 利用 AUTOSAR 等中间件组件构建其定制平台。但是，抽象和标准化程度较低。由于复杂性的原因，在这些领域中很难实现同时支持新旧硬件的功能更新。对于 PC（和 Mac）和智能手机（iOS、Android），都有全面的应用商店，使软件安装和更新变得简单。借助 Android Automotive 和 Apple CarPlay，我们看到了汽车信息娱乐领域应用商店的首次实现。PC 软件和智能手机应用程序都在很大程度上基于开源软件（OSS)。OSS 的使用大大提高了软件开发的效率。借助 SDV，汽车应用中 OSS 的使用比例也将增加。

# 软件定义车辆的基本概念

*需要对车辆 E/E 和硬件架构进行根本性变革，以实现硬件与软件的分离，这是软件定义车辆的核心概念。基于这种分离，进一步的软件架构变革提供了所需的灵活性。*

## 硬件架构的演变

当前，汽车的车载E/E架构正由面向领域的分布式架构向区域化、集中式架构转变（见图2）。

从分布式 E/E 架构开始，附加电子控制单元（ECU）被引入来提供新的车辆功能。随着时间的推移，ECU 的数量以及接线和重量也显著增加。

为了降低布线复杂性和重量，引入了域控制器。对于不同的域，如 ADAS/AD、车身和底盘、能源和动力系统以及信息娱乐，功能域中的 ECU 通过不同的总线系统连接到该域的其他 ECU。域控制器用于整合先前在域的不同 ECU 上实现的域功能。

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-2.png" alt="img" style="zoom:150%;" />

图 2：车载 E/E 架构的演变（概念)

*随着高性能计算机（HPC）和千兆以太网等车载高带宽网络的引入，E/E 架构可以进一步整合和集中。HPC 是实现软件定义车辆的关键要素。HPC 允许在强大的计算平台上执行不同领域的车辆功能，并允许更轻松地更新和升级车载软件。*

HPC 可用于整合不同领域的软件。例如，ADAS/AD、信息娱乐和车身控制可以在单个 HPC 上执行。跨域用例可以在 HPC 上实现，利用对域控制器的访问和对车辆总线的直接访问。最新的 E/E 架构使用区域（Zone）控制器 ECU 而不是域（Domain）控制器 ECU。区域控制器位于车辆的不同物理区域，并连接到该区域的智能传感器和执行器。区域控制器通过以太网连接到 HPC。这种区域架构可减少和简化布线和接线，并减轻重量。区域控制器可以提供智能电源管理、用作电源集线器并提供智能保险丝。区域控制器处理输入/输出（I/O）并从高阶计算中抽象出 I/O。

HPC 不会取代其他类型的 ECU。现在和未来的 E/E 架构将结合基于 µController 和 µProcessor 的 ECU，并由不同的总线系统组成。

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-3.png" alt="img" style="zoom:150%;" />

图 3：ECU 架构

HPC 通常连接到互联网并提供网关功能。HPC 软件的动态性和灵活性非常高。对于 SDV，HPC 是实施新车辆功能的主要部署单元。因此，无线更新（OTA）概念是任何基于 HPC 的架构不可或缺的一部分。

当今的 E/E 架构包含不同的总线系统，如以太网、CAN、LIN 和 Flexray。每种总线在实时能力、带宽、速度和成本等方面都有明显的优缺点。不同总线之间的主要区别在于所使用的通信原理。以太网允许在 PC、服务器和移动设备上进行 IT 行业中使用的面向服务的通信，从而实现车辆和后端的微服务软件架构。CAN 和 Flexray 等总线使用面向信号的通信。确定性的面向信号的通信对于硬实时用例是必要的。SDV 需要满足安全性和实时性要求，但也应通过提供面向服务的软件设计来提供灵活性。这导致了一种结合信号和服务通信的混合或分层通信架构。

借助 ADAS/AD 更智能的传感器，传感器融合和人工智能（AI）是 SDV 不可或缺的一部分。这种以数据为中心的处理需要足够的带宽（千兆以太网）和具有强大 GPU 或 NPU 的 ECU。低延迟和高带宽的云连接和高清地图也是实现 3 至 5 级自动驾驶的必备条件。

防抱死制动系统（ABS)、安全气囊控制或电动转向和发动机管理等车辆功能具有严格的实时要求和苛刻的功能安全要求（ASIL D)。这些功能由 µController 控制，通常连接到 Flexray 或 CAN 总线。通信基于信号。智能传感器和执行器也使用 µController。AUTOSAR Classic 通常用于此类 ECU。AUTOSAR Classic 基础软件（BSW）包括 OSEK 实时操作系统（见图 3）。在 AUTOSAR Classic 中，应用程序被实现为软件组件（SWC)，它们通过 AUTOSAR 运行时环境（RTE）提供的虚拟功能总线进行通信。

µController 应用程序提供基本的安全相关和实时约束车辆功能。与其他车辆软件相比，µController 应用程序相对静态。尽管如此，SDV 将提供通过无线（OTA）更新此软件的方法。

ASIL B 级的典型车辆功能包括控制前灯和刹车灯。域 ECU 或区域 ECU 提供 ASIL-B 功能。对于这些功能，使用 µProcessor。除了 CAN 和/或 Flexray 连接外，域/区域 ECU 还可以连接到汽车以太网并通过 SOME/IP 等协议与其他 ECU 通信。通信是面向服务的。

网关 ECU 提供总线间通信，连接到两个或更多车辆总线（以太网、CAN、Flexray、LIN），并可在不同的总线之间转换消息或信号。对于以太网到 CAN、Flexray、LIN 的通信，它们将面向服务的消息转换为信号，反之亦然。对于远程信息处理/连接 ECU（TCU)，ECU 可能具有互联网连接，需要加密和其他安全功能，并提供对其他 ECU 的访问权限以进行 OTA。

µProcessor 应用程序通常在 AUTOSAR Adaptive 堆栈上运行，这需要基于 POSIX 的操作系统，如 QNX 或 Linux。应用程序通常用 C++ 编写，与 µController 应用程序相比，其功能更新更具动态性。新车辆功能至少需要对这些高级应用程序进行更改。

多个 µController 或者 µProcessor + µController 通常会组合在片上系统（SoC）ECU 中。在这种 ECU 中，ECU 内的通信是通过共享内存等 IPC 方法进行的。

ADAS/AD 功能（如雷达和基于摄像头的巡航控制）也主要属于 ASIL-B 应用。这些要求苛刻的用例需要功能强大的 ECU。这些应用类型的 ECU 基于高性能 µProcessor 架构，并在 SoC 上包含 GPU 和 NPU，以执行基于传感器融合的人工智能算法。

在 HPC 硬件之上，虚拟化层（Hypervisor）用于并行运行多个操作系统。例如，Linux 操作系统可以与 AUTOSAR Adaptive 堆栈一起在同一硬件上执行，Android Automotive 操作系统也可以在同一硬件上执行。同时，Android Automotive 可用于为信息娱乐提供环境并运行仪表盘 UI。

例如，AUTOSAR Adaptive 堆栈可用于实现车身控制功能或 ADAS/AD 功能。AUTOSAR Adaptive 需要 Linux 等 POSIX 操作系统。要运行安全关键型应用程序，Linux 操作系统需要通过功能安全认证（ISO 26262)。Linux 的目标 ASIL 级别为 ASIL-B。AUTOSAR Adaptive 将应用程序作为 POSIX 操作系统的进程执行。这提供了灵活性，因为应用程序可以在运行时更新。

OpenJ9 等标准 Java VM 未通过功能安全认证，因此只能用于 ASIL-QM 车辆功能，例如云集成，以及作为 Android Automotive 的替代方案来实现用户界面。

## 软件架构的演变

将硬件与软件分离是实现 SDV 的关键概念。自 2002 年以来，汽车行业在 AUTOSAR 等标准上投入了大量资金。AUTOSAR Classic 标准化了 ECU 的软件架构，并通过提高 OEM 和供应商之间软件模块的可重用性和可交换性来改善集成 E/E 架构的复杂性管理。它提供了标准化的基础软件（实时操作系统）和中间件（虚拟功能总线）。AUTOSAR classic 支持 CAN、LIN 总线等基于信号的通信，它用于功能安全级别高达 ASIL-D 的深度嵌入式系统，如发动机控制系统或制动系统。AUTOSAR classic 的应用程序是用 C 编程语言编写的。

在指定 AUTOSAR classic 时，最初并未考虑更新这些系统。随着 ADAS 系统的引入及其扩展的计算能力和 OTA 更新，需要一种比 AUTOSAR classic 更灵活的解决方案，该解决方案被指定为 AUTOSAR adaptive。AUTOSAR adaptive 支持基于以太网和 SOME/IP 的服务通信。AUTOSAR adaptive 通常用于高性能系统，如 ADAS/AD 系统，可通过 OTA 更新。AUTOSAR adaptive 的应用程序用 C++ 编程语言编写。AUTOSAR adaptive 软件堆栈需要在 POSIX 操作系统上运行。

尽管 AUTOSAR 将硬件与软件分离，并提供了标准化和中间件，但与 PC 和智能手机应用程序相比，汽车应用程序的应用程序开发仍然复杂且缓慢。功能安全要求需要进行大量测试和验证。

汽车电子电气、硬件和软件是一个基于多系统的大系统。开发软件中的新功能需要考虑整个系统。对于 SDV，OEM 需要提供软件平台，包括用于特定电子电气硬件配置的完整软件堆栈。OEM 通常将其称为操作系统（例如 VW.OS、MB.OS），不应将其与 Windows 或 Linux 等操作系统混淆。相反，这些汽车操作系统包括引导加载程序、虚拟机管理程序、设备驱动程序、操作系统（Linux、实时操作系统、Android）、中间件（AUTOSAR 经典版、自适应等）、虚拟机/容器等的完整软件堆栈。

*AUTOSAR classic 和 adaptive 是汽车软件的关键软件标准。要实现软件定义汽车的车辆操作系统愿景，还需要 Eclipse SDV 和 SOAFEE 计划等其他标准和计划。*

## 车辆使用的容器技术

为了支持灵活性和简单、一致的部署，容器在 IT 行业和云环境中变得流行起来。容器是包含代码及其依赖项（如库、工具、设置和所需的运行时组件）的软件包。容器镜像可以分层，这意味着它们可以依赖于已经存在的镜像，而这些镜像本身又依赖于其他镜像等等。这种分层方法是定义清晰依赖关系而无需开销的好方法。因此，容器在管理软件组件相互依赖的复杂性方面具有巨大优势。在汽车行业，不同车型和配置之间的版本和配置管理是一项重大挑战，随着越来越多的车辆功能以软件形式实现，这一挑战将进一步加剧。容器重量轻、便携、安全，并提供隔离空间，从而提高系统的稳定性。通过将软件与操作系统分离，可以轻松传输容器，这对 OTA 来说是一个很大的优势。

尽管容器技术需要更多的硬件资源，但它在车辆中的应用却越来越多。要执行容器，需要在 HPC 上的 Linux 操作系统之上使用 Podman 等容器运行时或 Kubernetes 等更复杂的平台。中间件将集成 V2X 对云的访问，提供面向服务的车载通信，并利用容器技术进行 OTA。

*IT行业成功运用的容器技术可以应用于汽车行业，以简化无线更新、增强可扩展性和稳健性，并管理硬件和软件配置的依赖关系。*

## 平台和 API 管理

应用程序编程接口（API）定义云后端向联网汽车提供的服务。SDV 将使用多种服务来满足不同的车辆域。随着时间的推移，这些服务及其 API 将发生变化。车载软件将在车辆的使用寿命内发生变化，这是由于 OEM 或外部组织提供的功能更新和新功能，包括具有智能基础设施的 V2X 场景。随着新一代车辆的推出，功能集将会增加，需要额外的后端服务。其他后端服务将被更新，从而产生新版本的 API，或者如果不再使用则被删除。在这种复杂的软件和服务配置和依赖关系管理场景中，需要管理车载软件平台和服务 API。在 IT 行业中，基于 http 的 API 通常用于管理基于微服务的解决方案。微服务架构（包括保护车辆与后端通信的概念）可用于网联汽车。API 管理是解决方案的重要组成部分，用于管理后端提供的服务，这些服务依赖于车辆硬件和软件配置以及对所提供功能集的更改。

*随着从车辆到云后端的基于微服务的通信的使用越来越多，以及软件定义功能的不断更新，软件平台和 API 管理成为管理日益复杂的 API 和软件依赖关系的必需品。*

### 信息娱乐

车主希望智能手机上运行的应用程序（如 Spotify 或 Netflix）可以在汽车的主机和后座娱乐系统上使用。考虑到避免驾驶员分心、不同屏幕尺寸等特殊要求，Android Automotive 瞄准这个市场，提供专为车辆使用而定制的 Android 应用程序。一些 OEM 已决定将 Android Automotive 集成到他们的汽车中。其他 OEM 仅支持 Android Auto 或 Apple Carplay，它们将应用程序内容镜像到主机显示屏上，但不会深入集成到汽车中。考虑到车辆数据的价值以及 OEM 在 Android Automotive 方面对 Google 的依赖，OEM 正在考虑 Android Automotive 的替代方案。一种可能的解决方案是基于 Android Automotive 开源（AOSP）开发定制的 Android Automotive 解决方案。在这种情况下，OEM 需要提供定制的应用商店，因为定制的 Android Automotive 设备未经 Google 认证，因此不允许访问 Google Playstore。需要考虑互联网服务的区域可用性。例如，中国需要与欧盟或美国不同的服务提供商。 Google Play 服务仅适用于经过认证的 Android Automotive 设备，且不在中国提供。

苹果于 2022 年 6 月宣布升级其 CarPlay 解决方案，不仅与车辆信息娱乐系统集成，还与仪表盘用户界面集成。这种更深层次的车辆集成和功能安全要求需要 OEM 与苹果更紧密的合作。

*用户希望智能手机应用程序能够无缝集成到汽车的信息娱乐系统中。Android 用作在信息娱乐 ECU 上运行的操作系统和软件平台。如果 OEM 决定使用 Android 的开源版本（AOSP)，则 OEM 需要提供与 Google Play 服务和 Google Play Store 相当的服务和应用商店。*

# 车辆 DevOps

*SDV 需要集成的后端和车载架构（见图 4）。为了支持功能的持续开发和推出，DevOps方法和工具可以应用于汽车软件开发和运营*。

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-4.png" alt="img" style="zoom:150%;" />

图 4：车辆的 DevOps

SDV 始终连接到边缘服务提供商（telco)，以访问后端提供的服务。5G 网络等新技术提供高带宽和低延迟。这些移动网络为车辆及其乘客提供新的移动服务。SDV 连接到内容交付网络（CDN)，以接收其固件和软件以及多媒体或流媒体内容的 OTA 更新。所有与云的通信都受到网络安全组件的严格保护。OTA 更新包括车辆软件的所有层，包括

- 板级支持包、驱动程序等
- 用于 HPC 虚拟化的虚拟机管理程序
- 不同的操作系统（Linux、ROS、Android Automotive、AUTOSAR Classic），包括容器运行时和平台
- 中间件组件（Adaptive AUTOSAR、Java JRE）
- 应用软件（AUTOSAR Classic Apps / SWC、AUTOSAR Adaptive Apps、Android Apps、Java Apps）

## 内置安全功能

*安全性是 SDV 所有组件中不可或缺的一部分。在后端，车队的安全运营中心（V-SOC）必不可少。*

随着车辆始终连接到互联网并使用越来越多不同的在线服务，网络安全攻击者对车辆的攻击媒介正在增加。网络安全威胁需要谨慎应对。

安全运营中心（SOC）是一个集中式组织，负责在组织和技术层面处理安全问题。通过使用分析和车辆远程访问，V-SOC 可以分析网络安全威胁并提供应对措施，例如为特定车辆组提供 OTA 更新。所有 OTA 更新均在车辆运营中心进行管理。需要准确且最新的配置和版本信息以及软件组件的兼容性才能向特定车辆发布和分发 OTA 更新。

所有车辆及其用户均经过识别，车辆访问由车辆运营中心控制。身份和访问管理支持钥匙丢失或车辆被盗等用例。身份、访问和车辆数据都是敏感信息。车辆运营中心提供手段来为车辆及其用户执行特定国家/地区的数据安全和隐私规则。

## 后端车辆服务

云后端为 SDV 提供服务，并允许 OEM 灵活快速地实现新的收入来源。一些车辆服务包括：

- 车载信息娱乐系统中的应用程序和媒体内容流媒体。
- 在某些地区，紧急呼叫等紧急服务已成为强制性要求。可以收集碰撞和碰撞前数据，以改进自动驾驶的机器学习算法。
- 远程服务提供已经常见的功能，包括检测车辆位置、控制访问以及提供剩余里程、下次预定维护等信息。
- 未来的车辆与基础设施和其他外部信息（车辆与一切，V2X）将扩展导航、停车和天气服务。车辆将深度集成到基础设施中，并获取有关其环境的在线信息。当前具有链接上下文的高清地图对于自动驾驶（AD）至关重要。
- 维护和维修服务基于远程诊断，并结合分析和人工智能来支持维护和持续反馈以改进产品。
- 车队管理服务向租赁公司或其他车队运营商提供有关车队和个别车辆的最新信息。车队管理是共享出行的基石。
- 合作伙伴门户允许外部公司（例如保险公司）访问车队和个人车辆信息。
- 电动汽车服务根据可用的充电站和车辆电池状态提供路线计算。

要开发和运营 SDV，需要一种新的集成方法。在 IT 行业，基于云基础设施开发和运营产品的 DevOps 流程已成功建立。随着软件成为汽车 OEM 的重点，DevOps 原则可以应用于车辆开发、维护和运营。

云基础设施承载着所有 IT 流程。将使用不同的超大规模云以及公共和 OEM 私有云。因此，需要多云管理和通用云安全。不同的 OEM 流程将通过 API 交换数据并进行连接，由 API 管理解决方案进行管理。OEM 将连接到电信公司和其他边缘提供商，还将连接到外部环境，例如内容提供商（如 Netflix）、服务提供商（如天气信息提供商）、OEM 的供应商和公共基础设施。

## 无线更新（OTA)

*随着网络连接变得无处不在，无线更新或 OTA 成为车辆更新的自然选择。*

OTA 引人注目的两个关键方面包括：

1. 定期更新关键和非关键软件
2. 车主可以在汽车的整个生命周期内购买或订阅新车辆功能，这些功能可通过 OTA 推送，并有可能以服务的形式提供。因此，为 OEM 提供新的收入来源

软件定义汽车是整个OTA的核心框架，从技术上可以分为两部分：

1. 软件无线更新（SOTA)，顾名思义，包括更新车辆内的软件组件
2. 固件无线更新（FOTA）包括通过无线方式更新固件的过程，即控制底层硬件的系统软件。

根据功能，OTA可进一步分为以下几类：

1. 信息娱乐和导航更新——包括 IVI 和导航更新，旨在提升用户体验并保持地图时效性
2. 能源管理更新——对于电动汽车来说至关重要，更新可能包括更新加热/冷却系统控制、扩展范围等。
3. 驾驶控制更新——包括所有驾驶性能改进和更新，例如 ADAS、自适应巡航控制、电动机软件、车身控制模块、动力传动系统控制、智能制动、车道辅助更新等。通常，这些更新具有普遍性，因此只能在车辆静止时更新
4. 安全更新——包括动力系统、安全气囊、刹车、稳定性控制等的关键安全更新。
5. 设备和其他功能更新——包括远程信息处理、摄像头、激光雷达等设备的更新，以及语音辅助、气候控制等其他功能的更新。

在设计端到端、有效的 OTA 解决方案时，必须考虑几个因素。可扩展性是一个关键问题，它与车辆数量、车型、车队规模、地理覆盖范围、OTA 类型、目标功能/设备数量等有关。顶级 Tier 1 供应商和主要 OEM 拥有多种 OTA 更新和远程数据收集解决方案，确保与 OTA 的兼容性是一项挑战。从标准方面来看，选择专有平台还是标准化开放平台将是关键。合规性和政策是必须考虑的另外两个因素，例如 GDPR、WP29、ISO22900 等法规合规性或其他企业合规性。同样，必须支持配置多个自定义策略的能力，包括通知、活动部署、回滚、安装、依赖关系和变体管理。

虽然大多数汽车 OEM 目前主要处理信息娱乐领域，但随着 SDV 成为中心点，扩展上述所有功能的可能性和需求将是必要的。

## 持续集成和部署

基于云基础设施，常见的持续集成和部署（CI/CD）服务是执行不同 OEM DevOps 流程的基础。CI/CD 是快速开发、集成和测试新车辆功能的关键。常见的 CI/CD 服务是所有开发制品和源代码的版本和配置管理。打包的可执行软件存储和管理在制品库中。CI/CD 环境为软件提供构建和测试流水线。构建、集成和测试在不同阶段执行，从开发阶段、模块、ECU 到整车集成阶段。软件持续部署到模拟、虚拟测试、软件在环（SiL）和硬件在环（HiL）环境中。提供有关成功/失败和质量以及 KPI 的准确状态信息，以使流程尽可能透明。

在CI/CD工具链之上，实现了不同的工程和软件开发、维护和运营流程：

- E/E 工程流程始于车辆特性和要求的定义。这些是所有后续开发、维护和操作流程的锚点。DevOps 流程需要考虑功能安全（ISO 26262）以及车辆安全（ISO 21434)，并基于意向性、可重复性和可追溯性的原则。E/E 工程软件相关流程包括逻辑功能架构、软件和服务设计、硬件和网络拓扑设计以及信号和服务导向通信设计的子流程。面向硬件的流程包括硬件组件架构、电子电路设计、布线设计、几何拓扑和线束设计。
- 车辆软件是自上而下开发的，以实现车辆功能。车载软件由敏捷团队开发，并基于 DevOps 原则。目标是完全自动化软件开发过程并减少当前使用的不同工具链的数量。尽管如此，Android Automotive 和 AUTOSAR Classic 和 Adaptive 需要不同的工具链。开源工具避免了供应商锁定，将更频繁地用于微服务应用程序和用户体验应用程序以及其他高级软件。车辆软件需要打包在不同的容器中，以提供给 OTA 和制造流程。
- 自动驾驶软件基于人工智能（AI）和机器学习（ML）算法以及深度神经网络（DNN）训练，这需要大量数据。数据被采集到数据湖中，然后被标记、过滤和整理为训练数据。需要进行分析才能深入了解大数据集。训练后的模型需要通过模拟进行测试和验证。已批准的软件被打包并存储在制品库中。
- 网联汽车要求同时开发车载软件和后端服务。对于后端服务，可以使用不同的云原生技术，例如微服务。后端服务基于敏捷工具和云工具链开发。后端服务提供车载软件或其他后端服务使用的 API。

## 容器化混合集成测试

如前几章所述，汽车车载系统开发过程包括各种异构硬件和软件组件。对于敏捷流程，频繁构建以及早期测试和验证（也称为左移）是必需的。从设计到实际实施的需求可追溯性是满足汽车标准合规性的另一个关键要求。车辆软件由多个团队开发，团队之间的密切合作非常重要。需要概述不同团队和软件应用程序和模块的状态，以跟踪整个软件开发过程的进度。

现代 CI/CD 系统提供了多个阶段来提高软件从设计到部署的成熟度。这些阶段使用的流程和工具对于不同的组织是不同的，并且取决于开发的不同类型的软件。例如，带有 AUTOSAR classic 的深度嵌入式微控制器需要与 Android 汽车信息娱乐解决方案完全不同的流程和工具。因此，CI/CD 系统需要提供高度的灵活性，并且需要根据特定团队的要求进行量身定制。尽管工具和流程是根据特定团队的需求而采用的，但软件集成和进度报告的总体流程需要在整个组织内保持一致和标准化。

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-5.png" alt="img" style="zoom:150%;" />

图 5：多阶段构建、集成和测试

在 IT 行业，使用容器技术且基于云的 CI/CD 系统正在成为标准。如图 5 所示，云原生 CI/CD 系统的典型堆栈由以下组件组成：

嵌入式软件开发的特点是工具环境非常异构。无法在云中执行多个嵌入式工具。为了将这些工具集成到通用 CI/CD 流程中，可以实现特殊的插件或容器用于实现对安装物理工具的适配。

车载软件在软件在环（SIL）和硬件在环（HIL）环境中进行测试，使用特殊的模拟、测试工具和硬件设备。随着容器技术在车辆中的引入，利用云原生 CI/CD 系统成为可能。在 Linux 上运行的 AUTOSAR Adaptive 应用程序可以在云原生 CI/CD 系统中执行。可以使用容器之间的虚拟以太网测试这些虚拟 ECU 与其他虚拟 ECU 的交互。从开发过程早期的虚拟 ECU 开始，这些 ECU 可以被连接到 CI/CD 工具链的物理 ECU 取代。这可以通过在物理设备上实施插件或容器适配器来实现。这种方法可以用“现实滑块（reality slider，如图 5）”来可视化：从软件的虚拟执行开始，到混合虚拟和物理设备的混合场景，软件最终部署到物理测试设备，最后部署到真实车辆上。

*基于容器化的 CI/CD 系统，集成虚拟和物理测试可以尽早、快速地集成和测试车辆软件和后端服务。*

## 软件版本和配置管理

软件版本和配置管理（SCM）是支持 OEM 作为软件公司的一个重要主题。随着以软件、API 和服务形式实现的功能日益复杂，必须仔细管理这些组件根据特定配置和版本的兼容性。如果出现问题，出于监管原因，必须追溯到根本原因，并提供细粒度的更新来解决问题，即使是十年前开发的汽车也是如此。

高效的软件公司会促进软件组件、库等的重用。因此，可供重用的可用制品需要易于开发人员找到且易于使用。一些软件公司（例如 Google 和 Facebook）使用所谓的单一存储库。在单一存储库中，公司的所有软件制品都存储在单个存储库中。所有开发人员都可以访问该公司的整个源代码。其他公司使用多存储库。不同项目或域的源代码存储在不同的存储库中。这两种方法各有利弊，本文无法讨论，且特定于组织。单一和多存储库也可以组合使用。例如，汽车 OEM 可能决定将某个域（例如 ADAS/AD）的所有源存储在单一存储库中，并为其他域（如信息娱乐域）提供另一个存储库。

*软件版本和配置管理是成功的软件公司促进跨团队协作和软件重用以及管理软件组件和服务之间的复杂性和依赖性的基础。*

# 自动驾驶 - 大循环

自动驾驶的机器学习基于传感器数据。神经网络的训练取决于训练数据的质量。对于更高级别的自动驾驶，即使是避免与鹿相撞等不寻常事件也需要进行训练。对于这些和其他非标准驾驶情况，训练数据很少。一种可能性是模拟这些事件，另一种可能性是从驾驶车辆收集数据。从车队收集数据的问题是从驾驶过程中产生的大量数据中选择或过滤正确的数据。不可能将每个传感器信息都传输到每个驾驶情况的后端。重要的是从某些驾驶情况下收集传感器信息，这可以改进训练数据集。一种识别和选择有趣的驾驶场景以改进训练数据集的方法称为大循环（见图 6）。在驾驶过程中控制车辆的机器人循环通过车辆外部循环扩展到后端，以通过改进训练数据集来改进机器学习算法。大循环由以下过程步骤组成：

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-6.png" alt="img" style="zoom:150%;" />

图 6：自动驾驶的大循环

车辆和软件之间的这种闭环实现了机器学习算法的持续改进过程。

自动驾驶汽车对架构有着独特的要求，更多地依赖于“故障操作”策略，而不是“故障安全”或“故障软化”。以下是设计自动驾驶汽车软件定义车辆架构时需要考虑的重要因素：

因此，人们高度重视车载计算机系统来处理实时传感器数据，需要一系列传感器和车辆模块数据处理，与云无缝集成以将重新训练的模型刷新到车辆中。

*通过建立大循环，可以不断改进 ADAS/AD 的 AI 模型和 ML 算法，大循环是一个与后端共享车辆数据以更新 AI 模型和 ML 算法的持续过程。*

# 组织变革

*软件定义汽车需要汽车行业进行组织变革并引入开源、敏捷和 DevOps 方法和实践。*

汽车 OEM 还可以通过利用和采用开源软件（OSS）和实践向 IT 行业学习。公司使用开源实践称为内部源。软件公司可以从内部源实践中受益，以改善开放式沟通和协作，并通过基于绩效的决策和适当的质量保证。这是组织文化的根本性变化。

如今，许多汽车 OEM 都是按不同的车辆领域甚至特定的 ECU 进行组织的。根据康威定律，这些组织将设计反映其通信结构的产品。即使引入了 SDV，不同的汽车领域仍将是一个硬件/软件主干，以提供客户功能。在此主干之上，随着集中式架构和面向服务架构的兴起，跨功能、跨领域软件的重要性将要求 OEM 的组织结构发生变化。简单地在现有领域之上添加一层将导致沟通不畅，无法实现预期的灵活性和开发速度。因此，必须建立新的组织结构和敏捷原则。Spotify 等软件公司使用部落和行会等敏捷组织结构以及 Scrum 和看板等方法取得了巨大成功。汽车行业面临着硬件工程和功能安全要求等特定挑战。将敏捷方法和组织模式应用于汽车组织需要仔细调查。

软件的复杂性不断增加，导致对汽车软件开发人员的需求量很大。汽车软件包括嵌入式软件，大部分用 C/C++ 编写。拥有这些汽车软件技能的人才在市场上很难找到。人才争夺战才刚刚开始。为了解决这个问题，使用开源软件、在开源社区、供应商、中间件提供商的生态系统中进行协作以及 OEM 之间在非竞争性软件方面的合作是关键。

*SDV 将引领 OEM、供应商和云提供商之间的新合作模式。*

随着信息娱乐和自动驾驶领域对高性能计算单元的需求，汽车 OEM 和供应商之间的传统合作模式将发生变化。OEM 直接与高通（例如宝马、大众）或 Nvidia（例如梅赛德斯）等芯片制造商合作。Tier 1 供应商也试图建立自己的软件平台并与云提供商（例如博世、微软）合作。

# 关键要点

- 软件定义汽车（SDV）将硬件与软件分离，可更新和升级、自主、始终学习、始终连接、与环境交互并支持基于服务的商业模式。
- 硬件与软件的分离以及软件抽象层和平台的引入是电信、个人电脑和智能手机行业的基本概念，可应用于汽车行业。
- 需要对车辆 E/E 和硬件架构进行根本性变革，以实现硬件与软件的分离，这是软件定义车辆的核心概念。基于这种分离，进一步的软件架构变革提供了所需的灵活性。
- 随着高性能计算机（HPC）和千兆以太网等车载高带宽网络的引入，E/E 架构可以进一步整合和集中。HPC 是实现软件定义汽车的关键要素。它们允许在强大的计算平台上执行不同领域的车辆功能，并允许更轻松地更新和升级车载软件。
- AUTOSAR classic 和 adaptive 是汽车软件的关键软件标准。要实现软件定义汽车的车辆操作系统愿景，还需要 Eclipse SDV 和 SOAFEE 计划等其他标准和计划。
- 在 IT 行业成功应用的容器技术可以应用于汽车行业的车辆和云后端，以简化无线更新、增强可扩展性和稳健性，并管理硬件和软件配置的依赖关系。
- 随着从汽车到云后端的基于微服务的通信越来越多被使用，以及软件定义功能的不断更新，软件平台和 API 管理成为管理日益复杂的 API 和软件依赖关系的必需品。
- 用户希望智能手机应用程序能够无缝集成到汽车的信息娱乐系统中。Android 用作在信息娱乐 ECU 上运行的操作系统和软件平台。如果 OEM 决定使用 Android 的开源版本（AOSP)，则 OEM 需要提供与 Google Play 服务和 Google Play Store 相当的服务和应用商店。
- SDV 需要集成后端和车载架构（见图 4）。为了支持功能的持续开发和推出，DevOps 方法和工具可应用于汽车软件开发和运营。
- 安全性是 SDV 所有组件中不可或缺的一部分。在后端，车队的安全运营中心（V-SOC）必不可少。
- 随着网络连接变得无处不在，无线更新或 OTA 成为车辆更新的自然选择。
- 基于云基础设施，通用的持续集成和部署（CI/CD）服务是执行不同 OEM DevOps 流程的基础。CI/CD 是快速开发、集成和测试新车辆功能的关键。
- 基于容器化的CI/CD系统，集成虚拟和物理测试可以尽早、快速地集成和测试车辆软件和后端服务。
- 通过建立大循环，可以不断改进 ADAS/AD 的 AI 模型和 ML 算法，大循环是一个与后端共享车辆数据以更新 AI 模型和 ML 算法的持续过程。
- 软件定义汽车需要汽车行业进行组织变革并引入开源、敏捷和 DevOps 方法和实践。
- SDV 将引领 OEM、供应商和云提供商之间的新合作模式。

# 参考架构

以下参考架构仅供参考，旨在展示如何解决 SDV 问题。参考架构仅提供一些示例，需要在特定场景中进一步讨论以满足个人需求。

## 车载架构

车载参考架构基本上根据其关键性来划分工作负载（图 7）。

- ASIL-D 应用程序将作为 AUTOSAR  classic 应用程序运行，其集成的安全 RTOS 位于微控制器之上。这些应用程序的后续更新尚未计划。
- 安全级别达到 ASIL-B 的应用程序可以在 x86 或 ARM 微处理器架构或 GPU 上运行。这些应用程序是 AUTOSAR adaptive、自定义 POSIX / Linux 应用程序或 Java 或 Android 应用程序。这些应用程序可以通过无线（OTA）更新。

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-7.png" alt="img" style="zoom:150%;" />

图7：车载架构

车载通信通过车辆网络/总线使用 SOME/IP 或其他汽车标准等通信协议进行。

Red Hat 正在开发一款具有功能安全认证的车载 Linux 操作系统。该 Linux 将在基于 Arm、x86 或 GPU 的 ECU 上运行。

典型的 HPC 软件架构包括一个虚拟化层，用于并行运行多个操作系统。要运行 ASIL-B 应用程序，此虚拟机管理程序也需要通过 ASIL-B 认证。虚拟化是可选的，Linux 也可以直接在 ECU 硬件上运行。

容器技术可帮助更新应用程序并管理应用程序和中间件组件的依赖关系。要运行容器化的车载软件，可以将 Podman 等容器运行时集成到 Linux 操作系统中。

对于 OTA，需要在车辆中运行代理软件来连接基于云的后端以接收更新包。

## IBM IEAP Edge AI 和中间件技术

在 2021 年慕尼黑国际汽车博览会（IAA）上，IBM 展示了我们的车载软件解决方案愿景，该解决方案用于管理软件容器、Java 应用程序和 AI 模型，并结合了我们的边缘和中间件技术。该演示（请参见图 8）展示了以下方面：

- 如何有效管理车载生命周期

  - 软件容器

  - Java 应用程序和

  - 人工智能模型

- 企业级车辆边缘设备管理

- 管理所有以车辆为中心的应用程序的平台；可用作 OEM 特定的应用商店

- 灵活地协调从车辆/边缘到云的工作负载

- 如何提高应用程序和系统软件的可移植性、生产力和可维护性

- 它利用未来的 Red Hat 车载操作系统，整合所有车辆特定的必需增强功能，包括 ASIL-B 的持续安全认证

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-8.png" alt="img" style="zoom:150%;" />

图 8：IEAP 边缘 AI

IAA 演示基于以下技术：

- IBM 嵌入式汽车平台，一个用于车载应用的 Java Open J9 – Java 运行时环境
- IBM Research 的 Edge AI SDK，可嵌入数据和 AI 算法库
- IBM Edge Application Manager（IEAM)，用于从云端管理边缘设备上的容器工作负载。IEAM 基于 Linux 基金会开源项目 Open Horizon。

## 无线更新和数据管道

无线（OTA）更新对于实现软件定义汽车所需的灵活性至关重要。车辆功能可以通过无线方式更新或升级。OEM 还可以为现有车队实现新功能，并通过车辆软件的 OTA 更新分发这些功能。

汽车供应商联盟 eSync 正在通过定义和标准化 OTA 更新和诊断的功能和 API 来简化 OTA 部署。eSync 无缝跨越所有车载网络的边界，以达到任何符合 eSync 标准的模块（TCU、ECU、IVI、ADAS 等）。它涵盖云到车的连接、车辆网关、数据管理和具有端到端网络安全的中间件（来源：[eSync 联盟](https://esyncalliance.org/)）。

Excelfore 和 Red Hat 基于 eSync 标准创建了参考架构（来源：“ [Excelfore 和 Red Hat 标准化汽车 OTA 更新](https://www.redhat.com/en/resources/excelfore-standardize-automotive-ota-updates-overview)”，见图 9）。

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-9.png" alt="img" style="zoom:150%;" />

图 9：eSync OTA 架构
来源：[Excelfore 和 Red Hat 标准化汽车 OTA 更新](https://www.redhat.com/en/resources/excelfore-standardize-automotive-ota-updates-overview)

该架构的后端由基于在 Red Hat OpenShift 上运行的 OTA 服务器的 eSync 平台组成。在车辆中，eSync 客户端在网关 ECU 上运行，而车辆中的代理在 ECU 或区域/域 ECU 或传感器上运行。容器可用于车载软件，并为车辆的更新和升级提供多种好处。容器是自包含的实体，附带执行所需的所有依赖项。它们是轻量级的，促进了通用组件的重用，因此减少了更新的大小。

为了支持车辆中的容器，需要使用符合 POSIX 标准的操作系统，例如 Red Hat 车载操作系统。此 Linux 包含一个轻量级容器运行时，例如 Podman，用于在 ECU 上运行和停止容器。

## 基于 OpenShift 在 SiL 和 HiL 环境中测试 AI 和 MIL

开发和测试自动驾驶（AD）系统需要分析和存储比以往更多的数据。能够在管理快速基础设施增长的同时更快地提供洞察的客户将成为行业领导者。为了更快地提供这些洞察，底层 IT 技术必须以安全性、可靠性和高性能支持新的大数据和传统应用程序。为了处理大量非结构化数据增长，解决方案必须无缝扩展，同时将数据价值与不同存储层和类型的功能和成本相匹配。IBM 提供人工智能（AI）工作负载解决方案，重点关注支持 NVIDIA DGX 的 IBM Storage for AI。

人工智能（AI）和机器学习（ML）基础设施对于加速自动驾驶的发展至关重要。AI 和 ML 是自动驾驶汽车所有主要部件的底层技术，包括感知和定位、高级路径规划、行为仲裁、运动控制器。需要使用支持整个开发生命周期的集成机器学习解决方案组合来快速构建、验证和管理可扩展的 AI 和 ML 作业。自动驾驶公司需要快速进行实验，实现可重用性，并无缝构建、部署和大规模操作 ML 模型。软件在环（SiL）和硬件在环（HiL）测试是测试过程不可或缺的一部分。

机器学习生命周期是一个多阶段过程，旨在利用大量和各种数据、丰富的计算能力以及开源机器学习（ML）工具来构建智能应用程序。从高层次上讲，生命周期分为四个步骤：

- 数据采集和准备，以确保输入数据完整且高质量
- ML 建模，包括训练、测试和选择具有最高预测准确率的模型
- 应用程序开发过程中的 ML 模型部署和推理
- ML 模型监控和管理，用于衡量业务绩效并解决潜在的生产数据漂移。

机器学习开发人员主要负责 ML 建模，以确保所选的 ML 模型继续实现感兴趣的最高性能指标。ML 开发人员面临的主要挑战是：

- 选择并部署正确的 ML 工具（例如 Apache Spark、TensorFlow、PyTorch 等）
- 训练、测试、选择和重新训练提供最高预测准确度的 ML 模型所需的复杂性和时间
- 由于缺乏硬件加速，ML 建模和推理任务执行缓慢
- 反复依赖 IT 运营来配置和管理基础设施
- 与数据工程师和软件开发人员合作，确保输入数据的卫生，并在应用程序开发过程中成功部署 ML 模型

自动驾驶人工智能和大数据管理解决方案需要以下组件

- 高性能大数据存储解决方案
- 用于托管 OpenShift 或 Kubernetes 等容器平台的混合云基础设施
- 可扩展的高性能 AI/ML 服务器平台
- 与 HiL/SiL 集成

自动驾驶AI与大数据管理的参考架构如下图所示（图10）：

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-10.png" alt="img" style="zoom:150%;" />

图 10：在 SiL 和 HiL 上测试 AI 和 MIL
来源：[自动驾驶的 AI 和大数据管理](https://link.springer.com/chapter/10.1007/978-3-658-33521-2_30)

IBM Storage 使 IBM 的客户能够从小规模开始，采用经济实惠的解决方案进行早期实验，然后扩展性能和容量，以几乎无限的规模支持生产 AI、分析和商业应用程序。

RedHat OpenShift 和 Kubernetes 容器是加速 ML 生命周期的关键，因为这些技术为数据科学家提供了训练、测试和部署 ML 模型所需的敏捷性、灵活性、可移植性和可扩展性。Red Hat OpenShift 是业界领先的容器和 Kubernetes 混合云平台。它提供了容器管理和编排的所有优势。OpenShift 集成了 DevOps 功能并与硬件加速器集成，使数据科学家和软件开发人员之间能够更好地协作，并加速在混合云（数据中心、边缘和公共云）中推出智能应用程序。

SuperPOD 是 NVIDIA 的下一代云原生、多租户 AI 超级计算机。IBM 通过其存储解决方案 ESS 3200 支持 SuperPOD。ESS 3200 具有高度可扩展性。添加系统可以线性增加吞吐量，这与 SuperPOD 架构非常匹配。

ESS 基于 IBM Spectrum Scale，是 IBM 混合云战略的一部分，因此可通过企业存储服务无缝访问全球所有组织的数据。这使企业能够在所有企业数据中充分利用 SuperPOD 的强大功能。

## CI/CD 和混合测试

开发和运营软件定义汽车的组织需要最先进的持续集成和部署（CI/CD）流程，包括快速可靠的质量保证。在汽车行业中，通常使用软件在环（SiL）和硬件在环（HiL）流程。当前的 HiL/SiL 流程通常采用手动步骤执行，从而中断了软件的全自动开发和部署过程。设置和更新 HiL 环境是一项困难且昂贵的任务。无法使用 HiL 环境测试所有可能的车辆模型和变体。这引出了虚拟测试的想法。不是针对物理硬件进行测试，而是可以模拟 ECU 的通信伙伴。在 ECU 开发的早期阶段，当其他 ECU 尚不可用时，就可以使用虚拟测试。在开发的后期阶段，虚拟 ECU 将被物理 ECU 取代，测试将变为混合测试，包括传统的 SiL 和 HiL 测试。

<img src="https://www.ibm.com/blogs/digitale-perspektive/wp-content/uploads/2023/06/grafik-11.png" alt="img" style="zoom:150%;" />

图 11：CI/CD 和混合测试

容器技术提供了轻松管理和启动的机制

- 包括虚拟 ECU 的测试容器
- 连接到 HiL 或 SiL 环境的测试容器
- 工具容器提供编译器、模拟等开发工具。
- 构建管道来执行软件的构建和测试

包括混合测试在内的 CI/CD 工具链的实施是特定于客户端的，并且取决于构建软件所需的可用环境和工具

- 超大规模云提供云基础设施
- 容器平台，例如 Kubernetes 或 OpenShift
- 基于 Tekton、Jenkins、Argo 的 CI/CD 工具链（仅限 CD）

为特定组织实施流程，根据特定流程要求调整可用工具

与制品存储库集成，用于存储构建结果以供测试和部署期间使用

与源代码存储库（如 git、bitbucket 等）集成，其中包含源代码和版本及配置控制

# 标准、倡议和开源项目

| **姓名**            | **描述**                                                     | **SDV 相关性**               | **关联**                                            |
| ------------------- | ------------------------------------------------------------ | ---------------------------- | --------------------------------------------------- |
| **AUTOSAR**         | AUTOSAR（汽车开放系统架构）是汽车制造商、供应商、服务提供商以及汽车、电子、半导体和软件行业公司的全球开发合作伙伴关系。 | 车载软件                     | [https://www.AUTOSAR.org](https://www.autosar.org/) |
| **Autoware**        | Autoware 基金会是一家非营利组织，致力于支持实现自动驾驶出行的开源项目。Autoware 基金会在企业发展和学术研究之间创造协同效应，让自动驾驶技术惠及每个人。您的贡献至关重要。 | 自动驾驶                     | https://www.autoware.org/                           |
| **Covesa / Genevi** | COVESA 是一个开放、协作且具有影响力的技术联盟，致力于加速发挥联网汽车的全部潜力。 | 联网汽车                     | [https://covesa.global](https://covesa.global/)     |
| **Eclipse Hono**    | Eclipse Hono 提供远程服务接口，用于将大量物联网设备连接到后端，并以统一的方式与它们交互，而不管设备通信协议如何。 | 联网汽车远程服务接口         | https://www.eclipse.org/hono/                       |
| **Eclipse Iceoryx** | Eclipse Iceoryx 是一个进程间通信中间件，可以实现几乎无限的恒定时间的数据传输。 | 车载软件通讯中间件           | https://iceoryx.io/v1.0.1/                          |
| **Eclipse Kuksa**   | 为从车辆本身到云端的汽车软件生态系统建立共享标准和软件基础设施，可以提高开发速度，节省成本，并有助于为从原始设备制造商到供应商到第三方服务提供商的各种汽车参与者建立市场和开放平台，同时又不影响安全性。 | 车载和云软件标准             | https://www.eclipse.org/kuksa                       |
| **Eclipse OpenADx** | “开放、自动驾驶加速器”工作组的目标是提供业界广泛接受的自动驾驶工具链定义，这是一种参考架构，定义了范围内感兴趣的技术开源项目的互操作性，以实现现有开发工具更好的互操作性和功能性 | 自动驾驶工具链               | https://openadx.eclipse.org/vision/                 |
| **Eclipse SDV**     | Eclipse SDV 是一个面向未来软件定义汽车的开放技术平台；专注于使用由充满活力的社区开发的开源和开放规范来加速汽车级车载软件堆栈的创新。 | SDV 的 Umbrella Eclipse 项目 | [https://sdv.eclipse.org](https://sdv.eclipse.org/) |
| **SOAFEE**          | 嵌入式边缘可扩展开放架构（SOAFEE）项目是由汽车制造商、半导体供应商、开源和独立软件供应商以及云技术领导者定义的行业主导合作。该计划旨在提供针对混合关键性汽车应用增强的云原生架构，并提供相应的开源参考实现，以支持商业和非商业产品。SOAFEE 以 Project Cassini 和 SystemReady 等技术为基础，这些技术定义了 Arm 架构的标准启动和安全要求，SOAFEE 增加了云原生开发和部署框架，同时引入了汽车工作负载所需的功能安全、安全性和实时功能。 | SDV 的云原生架构             | [https://soafee.io](https://soafee.io/)             |


表1：SDV 标准、计划和开源项目

# 缩略语词汇表

- **ACES**自动驾驶、车联网、电动汽车和共享出行）
- **ADAS / AD**高级驾驶辅助系统 / 自动驾驶
- **AI**人工智能
- **API**应用程序编程接口
- **ASIL**汽车安全完整性等级，ISO 26262 功能安全的组成部分，分为 QM、A、B、C、D 级。其中 ASIL D 具有最高的功能安全要求
- **AUTOSAR**全球汽车制造商开发伙伴关系
- **CAN**控制器局域网络，一种车辆总线系统
- **DevOps**开发和运营
- **DevSecOps**开发、安全和运营
- **E/E**电气/电子
- **ECU**电子控制单元
- **FOTA**固件空中升级
- **Flexray**串行、确定性和容错总线系统
- **GPU**图形处理单元
- **HIL**硬件在环
- **HPC**高性能计算机
- **I/O**输入输出
- **IVI**车载信息娱乐系统
- **LIN**本地互联网络，传感器和执行器的串行通信系统
- **机器**学习
- **NPU**神经处理单元
- **OEM**原始设备制造商
- **OTA**无线更新
- **操作系统**
- **OSS**开源软件
- **Podman 是**一款在 Linux 上运行容器的容器引擎
- **ROS**机器人操作系统
- **TCU**远程信息处理控制单元
- **SCM**软件配置管理
- **SDK**软件开发套件
- **SDN**软件定义网络
- **SDV**软件定义汽车
- **SOAFEE**嵌入式边缘可扩展开放架构（SOAFEE)
- **SOTA**软件的更新
- **SIL**软件在环
- **SOC**安全运营中心
- **UX**用户体验
- **V2X**车辆到电动汽车
