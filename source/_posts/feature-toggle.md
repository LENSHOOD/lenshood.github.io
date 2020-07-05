---
title: Feature Toggles (aka Feature Flags)
date: 2020-07-04 23:37:54
tags:
- feature toggle
- truck based
categories:
- Software Engineering
---

> 本文是对 Pete Hodgson 的文章 [Feature Toggles (aka Feature Flags)](https://martinfowler.com/articles/feature-toggles.html#CategoriesOfToggles) 的全文翻译，一切版权归原作者所有。

*Feature Toggles（也经常被称为 Feature Flags）是一项强大的技术，它允许团队不改动代码就能改变系统的行为。Feature Toggles 分为了许多种使用类别，当实现和管理他们时，考虑这些类别是十分重要的。Toggle 会引入复杂性。我们可以通过更聪明的实现方法与合适的工具来管理 Toggle 配置，以此来使 Toggle 带来的复杂性可控，但是我们也应该限制系统中 Toggle 的总量。*

“Feature Toggling” 是一组模式，它能帮助团队更快且更安全的将新功能交付给用户。在下文中，我们会以一个小故事开头，来展现一些 Feature Toggling 适用的典型场景。之后我们会深入细节，这包括有助于团队成功实施 Feature Toggle 的特定模式与实践。

Feature Toggles 也被称为 Feature Flags， Feature Bits， 或 Feature Flippers。他们都是同一类技术的同义词。 在下文中我会交替使用 feature toggles 和 feature flags。

## Toggling 小故事

想象这样的场景。有一个复杂的城市规划仿真游戏项目，你供职于其多个项目组中的一个。你的团队负责核心仿真引擎。而你的任务是优化提升 Spline Reticulation 算法的效率。你心里清楚这种优化需要对现有实现进行相当大的改造，而这需要花费数周时间。同时，其他团队成员仍旧需要在与该算法相关的代码基础上继续一些正在进行中的工作。

基于以往合并长寿分支（long-lived branches）的痛苦体验，如果可能的话，这次你想要避免将这项工作进行分支。相反，你决定整个团队仍旧会基于主干进行工作，但对 Spline Reticulation 算法进行优化的开发者们将会使用 Feature Toggle 来防止他们的工作影响到其他成员，并防止对代码库产生不稳定。

### Feature Flag 的诞生

以下是算法优化小组对其进行的第一个修改：

修改前

```javascript
  function reticulateSplines(){
    // current implementation lives here
  }
```

这些示例代码全部使用 JavaScript ES2015

修改后

```javascript
  function reticulateSplines(){
    var useNewAlgorithm = false;
    // useNewAlgorithm = true; // UNCOMMENT IF YOU ARE WORKING ON THE NEW SR ALGORITHM
  
    if( useNewAlgorithm ){
      return enhancedSplineReticulation();
    }else{
      return oldFashionedSplineReticulation();
    }
  }
  
  function oldFashionedSplineReticulation(){
    // current implementation lives here
  }
  
  function enhancedSplineReticulation(){
    // TODO: implement better SR algorithm
  }
```

小组成员将现有的算法实现挪动到`oldFashionedSplineReticulation`函数中，且将`reticulateSplines`变为一个 **Toggle Point**。现在加入某人需要基于新算法工作，那么他可以通过将 `useNewAlgorithm = true` 这行的注释删掉来打开 “使用新算法” **Feature**。

### 让 Toggle Flag 变得更加动态

几个小时过去了，算法优化小组已经准备好在仿真引擎的一些集成测试上跑一跑他们的新算法了。同时，他们还想让这些集成测试能测试旧的算法。因此他们需要能让 Feature 动态的开启或关闭，这意味着是时候将这种对 `useNewAlgorithm = true` 这一行进行“注释”、“反注释”的笨重机制淘汰掉了：

```javascript
function reticulateSplines(){
  if( featureIsEnabled("use-new-SR-algorithm") ){
    return enhancedSplineReticulation();
  }else{
    return oldFashionedSplineReticulation();
  }
}
```

解下来我们引入`featureIsEnabled`函数，这是一个 **Toggle Router**，能用于动态的控制哪一条代码路径是畅通的。有很多种方式来实现一个 Toggle Router，其范围从最简单的内存存储到配有精致 UI 页面的更复杂的分布式系统实现。当下我们采用一个最简单的实现：

```javascript
function createToggleRouter(featureConfig){
  return {
    setFeature(featureName,isEnabled){
      featureConfig[featureName] = isEnabled;
    },
    featureIsEnabled(featureName){
      return featureConfig[featureName];
    }
  };
}
```

注意我们使用了 ES2015 的 [method shorthand](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Object_initializer#Method_definitions)。

我们可以基于一些默认配置（也许是读取自配置文件）来创建一个新的 toggle router，但我们也能动态的对一个功能进行开闭。这就使得自动化测试能同时验证一个 toggle feature 的两面：

```javascript
describe( 'spline reticulation', function(){
  let toggleRouter;
  let simulationEngine;

  beforeEach(function(){
    toggleRouter = createToggleRouter();
    simulationEngine = createSimulationEngine({toggleRouter:toggleRouter});
  });

  it('works correctly with old algorithm', function(){
    // Given
    toggleRouter.setFeature("use-new-SR-algorithm",false);

    // When
    const result = simulationEngine.doSomethingWhichInvolvesSplineReticulation();

    // Then
    verifySplineReticulation(result);
  });

  it('works correctly with new algorithm', function(){
    // Given
    toggleRouter.setFeature("use-new-SR-algorithm",true);

    // When
    const result = simulationEngine.doSomethingWhichInvolvesSplineReticulation();

    // Then
    verifySplineReticulation(result);
  });
});
```

### 准备好要发布了

更多的时间过去了，现在团队相信新的算法已经完成了功能。为了确认这一点，他们已经修改了高层的自动化测试，使得系统能在包含新功能与未包含新功能两种条件下受到验证。另外团队也想要做一些人工的试验性测试来确保所有功能都运行的与期望保持一致，毕竟，Spline Reticulation 是系统行为中非常关键的一部分。

为了对一个尚未被验证为可供一般使用的功能进行人工测试，我们需要能让该功能在生产环境上对一般大众用户关闭，而对内部用户开启。为了实现这一目标，有很多种办法：

- 让 Toggle Router 来基于 **Toggle Configuration** 做出决策，并且让这一 configuration 作用于特定的环境。只在预生产环境开启新功能。
- 通过某种类型的管理界面来允许 Toggle Configuration 能被实时修改。使用管理界面来在测试环境开启新功能。
- 教会 Toggle Router 如何动态的对每一次请求做出决策。这种决策方式引入了 **Toggle Context** 的概念，例如寻找某个特定的 Cookie 或者 Http Header。通常 Toggle Context 会被用作一个代理，来识别发出请求的用户。

（我们将会在之后更深入的探讨上述实现的细节，所以如果你不熟悉这些概念，也别担心。）

![](https://martinfowler.com/articles/feature-toggles/overview-diagram.png)

团队决定使用基于单请求的 Toggle Router，因为这种方式非常灵活。特别令人欣赏的是，这种方案允许他们不需要独立的测试环境，就能测试新算法了。取而代之的，他们只要简单的将新算法在生产环境打开，但只对内部用户开启（通过探测特定的 cookie）。现在团队能够将这种 cookie 加在他们自己身上，之后对新功能进行验证来确定其是否符合预期了。

### 金丝雀发布

新的 Spline Reticulation 算法基于试验性测试完成后，目前看起来工作良好。然而由于这个算法属于仿真游戏引擎中非常重要的一部分，所以他们对将新算法开放给所有用户仍然有些不情愿。团队决定使用 Feature Flags 基础设施来实施[**金丝雀发布**](https://martinfowler.com/bliki/CanaryRelease.html)，只将新功能开放给占总用户数很少量百分比的用户 -- 一个“金丝雀”用户群。

团队增强了 Toggle Router，教给他用户群的概念 -- 一个用户组，他们始终体验到某个功能始终处于打开或关闭态。一个金丝雀用户群，是通过随机采样所有用户中 1% 的用户来创建的（也许可以用对用户 ID 取模的办法）。金丝雀用户群将会持续的体验新功能开启，而其他 99% 的用户仍旧会使用旧的算法。核心业务指标（用户参与度，总收入等等）会同时在两个用户组中被监控，以此来确定新算法不会对用户行为产生负面影响。一旦团队确信新功能不会产生任何不良影响，他们就会将 Toggle Configuration 修改为对整个用户群打开新功能。

### A/B 测试

团队的产品经理在得知新功能顺利发布后很兴奋。她建议团队使用类似的机制来实施 A/B 测试。关于修改犯罪率算法以考虑污染程度是否会增加或降低游戏可玩性的争论由来已久。他们现在有能力用数据来结束这场争论了。他们计划推出一个抓住了这个想法本质的简单实现，用 Feature Toggle 来控制。他们将对一个相当大的用户群体打开这一新功能，然后研究这些用户相比“控制”组用户的行为。这一实践能允许团队基于数据而不是 [HiPPOs](http://www.forbes.com/sites/derosetichy/2013/04/15/what-happens-when-a-hippo-runs-your-company/) 来解决持久的产品辩论。

上述简单场景不仅是为了展现 Feature Toggle 的基本概念，更说明了这一核心能力可以有多少种不同的应用。现在我们已经接触了一些应用的例子了，让我们更进一步。我们将探索不同种类的 toggle，并且观察他们的不同点。我们会涉及如何编写易于维护的 toggle 代码，最后会分享一些实践，来帮助你避免 feature-toggle 系统的一些陷阱。

## Toggle 的分类

我们已经看到了 Feature Toggle 提供的基础功能 -- 在同一个可部署单元中实时的切换代码路径。以上场景同时也展示了这些功能在不同上下文下的不同使用方法。在同一个桶中放置所有 toggle 是一种诱人的选择，但同时也是危险的。不同类型的 toggle 在设计上起到不同的作用，如果以相同的方式来管理他们将会导致痛苦的事情发生。

Feature Toggle 可以采用如下两个维度来分类：toggle 的存活时间与 toggle 决策的动态程度。当然也有一些其他的因素要考虑（例如，由谁来管理这些 toggle），不过我认为寿命与动态度是能指导我们如何管理 toggle 的最重要的两大因素。

让我们通过这两个维度来考虑几种不同的 toggle 类别，并看看他们的适用场景。

### 发布 Toggle

有一类 Feature Flag 可以用于帮助主干开发的团队来实践持续交付（Continuous Delivery）。它允许未完成的功能被切入共享集成分支（例如 master 或 trunk）且允许该分支在任何时候被部署上生产环境。发布 Toggle 允许未完成和未被测试的代码作为[潜在代码（latent code）](http://www.infoq.com/news/2009/08/enabling-lrm)被送入生产中，而这些代码可能永远不会被打开。

产品经理也会使用同样的方法，来防止未完成的产品功能在一个以产品为中心的版本中被暴露给端用户。例如，一个电子商务网站的产品经理也许不想在只有部分送货供应商支持的情况下，让用户看到新的 “预估送货日期” 功能，而更希望直到所有送货供应商都支持这一功能时，才真正上线。产品经理同样也可能会有其他的理由，让一个已经实现完整，且经过全面测试的功能不暴露给用户。比如某功能的发布需要与某些营销活动相协调。发布 Toggle 是持续交付中 “将【功能】发布与【代码】部署分开” 原则的一种最常见的实现。

![](https://martinfowler.com/articles/feature-toggles/chart-1.png)

发布 Toggle 本质上是一种过渡方案。即使以产品为中心的 toggle 可能需要保持较长时间，他们也不应该持续超过一到两周。发布 Toggle 的决策通常是非常静态的。给定发布版的每个 toggle 决策都是固定的，通过发布一个新版本来改变这种 toggle 配置通常是完全可接受的。

### 试验 Toggles

试验 Toggle 通常用于实施多元化发布或 A/B 测试。系统内的每一个用户都会被置入某个群组，之后 Toggle Router 将会基于一个用户的群组，而持续的将这个用户实时的送入某个代码路径下。通过跟踪不同群组的聚合行为，我们能比较不同代码路径产生的影响。这项技术通常用于实现数据驱动优化，例如对一个电商系统的购买流程，或是CTA（Call To Action）按钮上文案选择的优化。

![](https://martinfowler.com/articles/feature-toggles/chart-2.png)

一个试验 Toggle 需要在同一处位置持续保持同样的配置，直到产生了足够显著的统计结果。取决于不同的业务模式，这可能意味着该 Toggle 存在的时间从几小时到几周不等。更长的时间就不太有效果了，因为系统的其他修改有可能会导致试验结果无效。试验 Toggle 的本质决定了它是高度动态的 -- 每一个到来的请求都可能代表了不同用户，所以路由的结果也会不同。

### 运维 Toggles

这一类 toggle 用于对运维层面的系统行为进行控制。当我们要推出一个新功能，但我们对其可能造成的性能影响还不清楚时，会引入运维 Toggle，这样运维管理员就能在需要时快速的禁用或降级生产环境的该功能。

大多数运维 Toggle 都相对短命 -- 一旦新功能在运维层面得到信任，那么该 toggle 就应该退休了。然而给一个系统增加少量长期存在的 “切断开关” 的实践也并不少见，这种开关能允许生产环境的运维人员在系统遭受不寻常的高负载时优雅的降级非关键系统功能。例如，当我们的系统处于重度负载时，我们也许想要禁用首页中生成起来相对昂贵的推荐面板功能。我咨询了一家在线零售商，该公司维护了运维 Toggle 功能，在关键需求产品发布之前，该公司可以故意禁用其网站主要采购流程中的许多非关键功能。这类长寿的运维 Toggle 可以被看做是一种人工控制的[断路器（Circuit Breaker）](https://martinfowler.com/bliki/CircuitBreaker.html)。

![](https://martinfowler.com/articles/feature-toggles/chart-3.png)

我们前面提到过，很多这类 toggle 都只会持续少量时间，但一些关键的控制可能会被保留下来，几乎无限期的留给运维人员。因为这些 toggle 的目的是为了让运维人员能对生产事件进行快速的响应，所以他们需要能够极其快速的被重配置 -- 为了修改运维 Toggle 的状态而需要推出一个新的发布可能不太会让一个运维人员感到快乐。

### 权限 Toggles

这类 toggle 用于修改对某些用户收到的功能或产品体验。例如我们可能有一个 “高级（premium）” 功能但只给付费用户开放。获取也许我们有一组 “alpha” 功能，只对内部用户开放，以及一组 “beta” 功能只对内部用户加 beta 用户开放。我把这种只将新功能开放给内部或 beta 用户的方式称为香槟早午餐（Champagne Brunch）-- 一个 “[drink your own champagne（译者注：类似于“吃自己的狗粮”）](http://www.cio.com/article/122351/Pegasystems_CIO_Tells_Colleagues_Drink_Your_Own_Champagne)” 更早的机会。

一个 Champagne Brunch 在很多地方都与金丝雀发布类似。他们之间的不同在于金丝雀发布是将新功能暴露给一个随机选择的群组，而 Champagne Brunch 是暴露给一些选定的用户们。

![](https://martinfowler.com/articles/feature-toggles/chart-4.png)

### 管理不同类型的 Toggle

到现在为止我们有了一个 toggle 分类的方案，因此我们能够讨论存活时间与动态程度这两个维度是如何影响我们处理不同类别的 Feature Toggle。

#### 静态 vs 动态 Toggle

![](https://martinfowler.com/articles/feature-toggles/chart-6.png)

需要实现实时路由决策的 toggle 要求更加复杂的 Toggle Router，以及对这些 Toggle Router 更复杂的配置。

对于简单的静态路由决策，其 toggle 配置可以简单为每一个功能设置 On/Off，其 toggle router 也只负责将静态的 On/Off 状态转发至 Toggle Point。就像我们先前讨论的，其他类型的 toggle 更加动态化，也就需要更复杂的 toggle router。例如对试验 Toggle 的 router，需要对给定的用户做出动态的路由决策，这可能会通过某种基于用户 id 的一致性分群组算法来实现。与从配置中读取静态的 toggle 状态不同，这类 toggle router 将会需要读取某些分群组配置的定义，例如试验群组与控制群组的规模应该多大。这类配置将会被用作分群组算法的输入。

我们将会在之后深入讨论更多的 Toggle 管理细节。

#### Long-lived toggles vs transient toggles

![](https://martinfowler.com/articles/feature-toggles/chart-5.png)

我们也可以将 toggle 类型分为本质上是临时的 vs. 长寿且可能会持续数年的。这一区别将会强烈的影响到我们对功能 Toggle Point 的实现方法。假如我们添加了一个将在几天后被移除的发布 Toggle，那么我们可能就完全抛弃 Toggle Point 而采用简单对 Toggle Router 进行 if/else 判断。这正是我们在前文中 spline reticulation 例子的做法。

```javascript
function reticulateSplines(){
  if( featureIsEnabled("use-new-SR-algorithm") ){
    return enhancedSplineReticulation();
  }else{
    return oldFashionedSplineReticulation();
  }
}
```

然而假如我们创建了一个新的权限 Toggle，我们期望其 Toggle Point 存活非常长的时间，那么我们当然不想随意的将 Toggle Point 实现为少量 if/else 检查。我们需要使用更加易于维护的实现技术。

## 实现技术

Feature Flags 似乎产生了相当混乱的 Toggle Point 代码，而这些 Toggle Point 也有在整个代码库中扩散的趋势。确保这一趋势对任何 feature flags 都可控则非常重要，尤其是对于那些长寿的 flag。以下有一些实现模式与实践能帮助减少这类问题。

### 决策点与决策逻辑解耦

一个 Feature Toggle 的常见错误就是将 toggle 决策发生的地方（即 Toggle Point）与决策背后的逻辑（即 Toggle Router）耦合在一起。来看一个例子。我们目前正在开发下一代电商系统。我们的其中一个新功能可以让用户便捷的通过点击他们的订单确认邮件（即清单邮件）中的一个链接，就能取消该订单。我们用 Feature Flags 来管理所有下一代新功能的推出。我们初始的 feature flag 实现看起来是这样的：

invoiceEmailer.js

```javascript
  const features = fetchFeatureTogglesFromSomewhere();

  function generateInvoiceEmail(){
    const baseEmail = buildEmailForInvoice(this.invoice);
    if( features.isEnabled("next-gen-ecomm") ){ 
      return addOrderCancellationContentToEmail(baseEmail);
    }else{
      return baseEmail;
    }
  }
```

当生成清单邮件时我们的 InvoiceEmailler 检查 `next-gen-ecomm `功能是否启用。如果是，则邮件发送器会增加一些附加的订单取消内容至邮件中，

这看起来是一个合理的做法，不过非常脆弱。关于是否在清单邮件中包含订单取消功能的相关内容直接和一个广泛的`next-gen-ecomm（下一代 ecomm）`功能开关相关联 -- 而且居然使用了一个魔数字符串。为什么发清单邮件的代码需要知晓订单取消功能是下一代功能集的一部分呢？如果我们想要暴露下一代功能中的一部分，而不包含订单取消呢？或者反之亦然？如果我们只想将订单取消功能暴露给一部分用户呢？在特性开发中，这种 “切换范围” 的更改很常见。还需要牢记在心的就是，这种 toggle point 会有蔓延至整个代码库的趋势。以我们现在的方法，因为 toggle 决策逻辑是 toggle point 的一部分，任何对该决策逻辑的修改都需要搜索所有这些被蔓延至代码库的 toggle point。

令人欣喜的是，[软件领域的任何问题都能通过增加一个中间层来解决（any problem in software can be solved by adding a layer of indirection](https://en.wikipedia.org/wiki/Fundamental_theorem_of_software_engineering)。我们可以用以下方式来将 toggle point 从决策逻辑中解耦：

featureDecisions.js

```javascript
  function createFeatureDecisions(features){
    return {
      includeOrderCancellationInEmail(){
        return features.isEnabled("next-gen-ecomm");
      }
      // ... additional decision functions also live here ...
    };
  }
```

invoiceEmailer.js

```javascript
  const features = fetchFeatureTogglesFromSomewhere();
  const featureDecisions = createFeatureDecisions(features);

  function generateInvoiceEmail(){
    const baseEmail = buildEmailForInvoice(this.invoice);
    if( featureDecisions.includeOrderCancellationInEmail() ){
      return addOrderCancellationContentToEmail(baseEmail);
    }else{
      return baseEmail;
    }
  }
```

我们引入了一个 `FeatureDecisions` 对象，作为一个所有 feature toggle 决策逻辑的集合点。我们在该对象上为每一个特定的 toggle 决策创建了一个决策方法 -- 在我们的 “我们是否应该在清单邮件中包含订单取消功能” 例子中，其决策被`includeOrderCancellationInEmail` 方法代表。至此，决策的 “逻辑” 已经变成检查`next-gen-ecomm` 特性状态的一个简单过程，但随着逻辑的更新发展，我们有了一个单独的地方来管理它。无论何时我们想要修改这个特定 toggle 决策的逻辑，我们都只要找到这个单一的地方即可。我们也许想要修改该决策的范围 -- 例如哪个特定的 feature flag 来控制该决策。或者，我们可能需要修改产生决策的原因 -- 想要从静态 toggle 配置驱动转为 A/B 试验驱动，或者任何由于操作上的问题，例如订单取消基础设施出现故障时。在所有的场景下，我们的清单邮件发送器都能幸福的对 toggle 决策是如何或为何产生保持不知情。

### 决策倒置

在之前的例子中，我们的清单邮件发送器需要询问 feature flags 基础设施功能应该如何执行。这意味着我们的清单邮件发送器需要知道一个额外的概念 -- feature flaging， 同时也就有一个额外的模块与他耦合。这使得清单邮件发送器更难单独工作和思考，也更难测试。随着 feature flaging 在我们的系统中逐渐流行的趋势，我们会看到更多的模块与成为一个全局依赖项的 feature flaging 耦合。这并不是一个理想的场景。

在软件设计中我们总能使用控制反转来解决这类耦合问题。在我们的例子里也一样。下面是我们如何将 feature flaging 基础设施与清单邮件发送器解耦的：

invoiceEmailer.js

```javascript
  function createInvoiceEmailler(config){
    return {
      generateInvoiceEmail(){
        const baseEmail = buildEmailForInvoice(this.invoice);
        if( config.includeOrderCancellationInEmail ){
          return addOrderCancellationContentToEmail(email);
        }else{
          return baseEmail;
        }
      },
  
      // ... other invoice emailer methods ...
    };
  }
```

featureAwareFactory.js

```javascript
  function createFeatureAwareFactoryBasedOn(featureDecisions){
    return {
      invoiceEmailler(){
        return createInvoiceEmailler({
          includeOrderCancellationInEmail: featureDecisions.includeOrderCancellationInEmail()
        });
      },
  
      // ... other factory methods ...
    };
  }
```

现在，与 `InvoiceEmailler` 直接获取 `FeatureDecisions` 不同，这些决策以一个 `config` 对象的形式，在构造时期被注入。`InvoiceEmailler` 现在对什么 feature flaging 已经完全不知情了。他只知道一些行为面能够被实时的配置。这种方式也让对 `InvoiceEmailler` 行为的测试变得容易 -- 我们能通过在测试时传入不同的配置选项，来将生成邮件中包含或不包含订单取消内容的两条路径都测试到：

```javascript
describe( 'invoice emailling', function(){
  it( 'includes order cancellation content when configured to do so', function(){
    // Given 
    const emailler = createInvoiceEmailler({includeOrderCancellationInEmail:true});

    // When
    const email = emailler.generateInvoiceEmail();

    // Then
    verifyEmailContainsOrderCancellationContent(email);
  };

  it( 'does not includes order cancellation content when configured to not do so', function(){
    // Given 
    const emailler = createInvoiceEmailler({includeOrderCancellationInEmail:false});

    // When
    const email = emailler.generateInvoiceEmail();

    // Then
    verifyEmailDoesNotContainOrderCancellationContent(email);
  };
});
```

我们同时还引入了一个 `FeatureAwareFactory` 来将这类需要 “决策注入” 的对象集中创建。这是通用依赖注入模式的一种应用。如果我们的代码库中已经配置了 DI 系统，那我们也许能直接使用它来完成我们的实现。

### 避免条件判断

到目前为止，我们例子中的 Toggle Point 都是以 if 语句来实现的。这在构建简单、短命的 Toggle Point 上还说得过去。但我们并不建议在需要过个 Toggle Point 的地方使用条件判断式的 Toggle Point，也不建议在期望 Toggle Point 长期存活的场景中使用。一个更易于维护的替代方法是采用某种策略模式来实现：

invoiceEmailler.js

```javascript
  function createInvoiceEmailler(additionalContentEnhancer){
    return {
      generateInvoiceEmail(){
        const baseEmail = buildEmailForInvoice(this.invoice);
        return additionalContentEnhancer(baseEmail);
      },
      // ... other invoice emailer methods ...
  
    };
  }
```

featureAwareFactory.js

```javascript
  function identityFn(x){ return x; }
  
  function createFeatureAwareFactoryBasedOn(featureDecisions){
    return {
      invoiceEmailler(){
        if( featureDecisions.includeOrderCancellationInEmail() ){
          return createInvoiceEmailler(addOrderCancellationContentToEmail);
        }else{
          return createInvoiceEmailler(identityFn);
        }
      },
  
      // ... other factory methods ...
    };
  }
```

这里我们通过给清单邮件发送器配置一个内容增强函数来实现策略模式。`FeatureAwareFactory`在创建清单邮件发送器时通过 `FeatureDecision` 的指导来选择一个策略。如果订单取消应该包含在邮件中，那么它会传入一个添加邮件内容的增强函数。否则他就传入一个`identityFn` 函数 -- 这个函数没有任何修改的作用，只是简单的将邮件返回。

## Toggle 配置

### 动态路由 vs 静态配置

先前我们将 feature flags 分成了两类：在给定代码部署中 toggle 的路由决策本质是静态的 vs. 决策在运行时是动态变化的。需要注意的一个重要的点是，两种方式可以实时改变 flag 的决策。首先，一些运维 Toggle 可能会由于需要响应系统故障而被动态的从 On *重配置* 为 Off。其次，一些 Toggle 类型，例如权限 Toggle 和试验 Toggle 基于一些请求上下文例如用户标记等来为每一个请求配置动态的路由决策。前者通过重配置来实现动态，而后者则是内在的动态。这种内在的动态 toggle 可能会做出高度动态的**决策（decision）**但依然存在一个相当静态的 **配置（configuration）**，这种配置可能只有通过重新部署来改变。试验 Toggle 就是这类 feature flag 的一个典型 -- 我们并不需要在运行时去修改其参数。事实上这样做很可能会导致该试验在统计上失效。

### 更偏好静态配置

如果 Toggle 实现允许的话，采用源代码控制或重新部署来管理 toggle 配置是最好的。通过源代码控制来管理 toggle 配置带给我们的好处与通过源代码控制来实现基础设施即代码的好处一样。他允许 toggle 配置与被 toggle 的代码库共存，这提供了一个巨大的好处：toggle 配置会随着你的持续交付流水线移动，就像代码修改或者基础设施修改一样。这充分发挥了 CD 的优点 -- 以一致的方式且跨环境验证的可重复构建。这也极大的降低了 feature flag 的测试负担。我们并不需要分别验证该一个发布在 Toggle On 或 Off 时的表现，因为其状态已经写入了发布且不会被改变（至少对低动态度的 flag 而言）。另一个 toggle 配置与源代码控制并存的好处是我们能简单的看到上一个发布中 toggle 的状态，并且能在需要时简单的重建上一个发布。

### 管理 Toggle 配置的方法

虽然我们更喜欢静态配置，但仍然有很多场景例如运维 Toggle 等，要求更动态的配置办法。一起来看看一些管理 toggle 配置的选项，范围从简单但低动态的方法到一些高度先进但会增加额外复杂度的方法。

### 硬编码 Toggle 配置

最基本的技术 -- 也许基础到不被认为是一个 Feature Flag -- 就是简单地将代码块注释或反注释。例如：

```javascript
function reticulateSplines(){
  //return oldFashionedSplineReticulation();
  return enhancedSplineReticulation();
}
```

比注释代码稍微高级一点的方法是当可用的时候使用预处理器的 `#ifdef` 特性。

由于这种硬编码并不能允许对 toggle 动态的重配置，因此它只适用于我们期望通过部署代码来重配置 flag 的场景。

### 参数化的 Toggle 配置

通过硬编码提供的编译时配置在许多场景，包括许多测试场景下，都不够灵活。一个至少允许 feature flags 在不需要重新编译应用或服务就能重配置的简单的方法是通过命令行参数或环境变量来指定 Toggle 配置。这是一种简单且历史悠久的 Feature Toggle 或 Feature Flaging。然而他也存在限制。在大量进程之间协调配置也许会变的笨拙，且对 toggle 配置的修改需要重新部署或至少需要重启进程（可能还存在重配置 toggle 的人对服务器的访问特权问题）。

### Toggle 配置文件

另一个选择是从一些结构化文件中读取 Toggle 配置。将 Toggle 配置放入更加通用的应用配置文件中的办法非常常见。

有了 Toggle 配置文件，现在你可以通过简单的修改文件来重配置，而不是重新编译整个应用代码。然而，即使你不需要重新编译应用来 toggle 一个特性，在多数情况下你仍旧需要为了重配置一个 flag 而实施重新部署。

### Toggle 配置在应用 DB 中

当你的系统达到一定规模后，使用文件来管理 toggle 配合会变得很麻烦。通过文件来修改配置相对麻烦。在一对服务器之间保持配置的一致性是一项挑战，进行一致性修改则更难。为了响应这一点，许多组织都将 Toggle 配置移动至了某种中心化存储中，通常是应用现存的 DB 中。这经常伴随着某种形式的管理界面的构建，这样能使系统运维、测试和产品经理能查看和修改 Feature Flags 及其配置。

### 分布式 Toggle 配置

使用已经是系统架构组成部分的通用目的 DB 来存储 toggle 配置非常常见；因为这是在引入 Feature Flags 且开始获得关注后，Toggle 配置最明显的去处了。然而如今有了一类专用目的的分层 Key-Value 存储更加适合管理应用配置 -- 像 ZooKeeper、etcd 或 Consul。这些来自于分布式集群的服务提供了一个为集群所有节点共享的环境配置源。配置能在任何需要修改的时候被修改，且集群内所有节点都能自动的获悉这一修改 -- 一个非常趁手的奖励特性。使用这些系统来管理 Toggle 配置意味着我们能在集群任何节点上拥有 Toggle Router，且基于整个集群的配置来协调给出最终的决策。

其中一些系统（例如 Consul）还带有管理界面，提供基本的管理 Toggle 配置的方式。不过，在某些情况下，我们通常会创建一个定制的用于管理 toggle 配置的小应用。

### 覆盖配置

到现在为止我们的讨论都假定所有的配置都采用同一种机制来提供。而许多系统的实际情况会更复杂，可能需要对不同来源的配置进行分层覆盖。对 Toggle 配置而言，给定一个默认配置和特定环境下的一个覆盖配置是很常见的。这些配置覆盖可能来自于简单的附加配置文件，也可能来自于复杂的 ZooKeeper 集群。请注意，任何特定于环境的配置覆盖，都与持续交付的理想状态 -- 即在交付流水线中保持完全相同的位和配置流 -- 背道而驰（runs counter）。通常在实用主义角度，可以使用一些环境特定的配置覆盖，但努力保持你的可部署单元与你的配置不受环境影响将会得到一个更简单、安全的流水线。当讨论测试一个 feature toggle 的系统的时候，我们会很快的重新讨论这个话题。

#### 基于请求的覆盖

一个实施特定环境配置覆盖的替代方法是让 toggle 的 On/Off 状态能基于单个请求来覆盖，方式包括指定的 cookie、查询参数、或 HTTP header。这与全量配置覆盖相比有一点优势。假如服务是负载均衡的，那么你仍旧能确信该配置覆盖会被实施在无论是哪一个被命中的服务实例上。你也可以在不影响其他用户使用的情况下，覆盖生产环境的 toggle 配置，并且不太可能会意外的将配置覆盖遗留。假如基于请求的配置覆盖机制使用持久的 cookie，那么某个测试你的系统的人就能定制他自己的 toggle 集合来覆盖配置，且能在浏览器中持续生效。

基于请求的配置覆盖方法的缺点在于，这可能引入好奇的或是恶意的终端用户来由他们自己修改 feature toggle。一些组织可能会对某些未发布特性可能会被某个足够坚定的组织访问到而感到难受。对配置覆盖进行加密签名是缓解这一问题的一种选择，但无论如何这种方法都会对你的 feature toggle 系统增加复杂度 -- 和攻击面。

我在[这篇文章](http://blog.thepete.net/blog/2012/11/06/cookie-based-feature-flag-overrides/)中详细介绍了这种基于 cookie 配置覆盖的方法，同时还描述了一个由我和 ThoughtWorks 同事一起开源的[一个 ruby 的实现](http://blog.thepete.net/blog/2013/08/24/introducing-rack-flags/)。

## Working with feature-flagged systems

While feature toggling is absolutely a helpful technique it does also bring additional complexity. There are a few techniques which can help make life easier when working with a feature-flagged system.

### Expose current feature toggle configuration

It's always been a helpful practice to embed build/version numbers into a deployed artifact and expose that metadata somewhere so that a dev, tester or operator can find out what specific code is running in a given environment. The same idea should be applied with feature flags. Any system using feature flags should expose some way for an operator to discover the current state of the toggle configuration. In an HTTP-oriented SOA system this is often accomplished via some sort of metadata API endpoint or endpoints. See for example Spring Boot's [Actuator endpoints](http://docs.spring.io/spring-boot/docs/current/reference/html/production-ready-endpoints.html).

### Take advantage of structured Toggle Configuration files

It's typical to store base Toggle Configuration in some sort of structured, human-readable file (often in YAML format) managed via source-control. There are some additional benefits we can derive from this file. Including a human-readable description for each toggle is surprisingly useful, particularly for toggles managed by folks other than the core delivery team. What would you prefer to see when trying to decide whether to enable an Ops toggle during a production outage event: **basic-rec-algo** or **"Use a simplistic recommendation algorithm. This is fast and produces less load on backend systems, but is way less accurate than our standard algorithm."**? Some teams also opt to include additional metadata in their toggle configuration files such as a creation date, a primary developer contact, or even an expiration date for toggles which are intended to be short lived.

### Manage different toggles differently

As discussed earlier, there are various categories of Feature Toggles with different characteristics. These differences should be embraced, and different toggles managed in different ways, even if all the various toggles might be controlled using the same technical machinery.

Let's revisit our previous example of an ecommerce site which has a Recommended Products section on the homepage. Initially we might have placed that section behind a Release Toggle while it was under development. We might then have moved it to being behind an Experiment Toggle to validate that it was helping drive revenue. Finally we might move it behind an Ops Toggle so that we can turn it off when we're under extreme load. If we've followed the earlier advice around de-coupling decision logic from Toggle Points then these differences in toggle category should have had no impact on the Toggle Point code at all.

However from a feature flag management perspective these transitions absolutely should have an impact. As part of transitioning from Release Toggle to an Experiment Toggle the way the toggle is configured will change, and likely move to a different area - perhaps into an Admin UI rather than a yaml file in source control. Product folks will likely now manage the configuration rather than developers. Likewise, the transition from Experiment Toggle to Ops Toggle will mean another change in how the toggle is configured, where that configuration lives, and who manages the configuration.

### Feature Toggles introduce validation complexity

With feature-flagged systems our Continuous Delivery process becomes more complex, particularly in regard to testing. We'll often need to test multiple codepaths for the same artifact as it moves through a CD pipeline. To illustrate why, imagine we are shipping a system which can either use a new optimized tax calculation algorithm if a toggle is on, or otherwise continue to use our existing algorithm. At the time that a given deployable artifact is moving through our CD pipeline we can't know whether the toggle will at some point be turned On or Off in production - that's the whole point of feature flags after all. Therefore in order to validate all codepaths which may end up live in production we must perform test our artifact in **both** states: with the toggle flipped On and flipped Off.

![](https://martinfowler.com/articles/feature-toggles/feature-toggles-testing.png)

We can see that with a single toggle in play this introduces a requirement to double up on at least some of our testing. With multiple toggles in play we have a combinatoric explosion of possible toggle states. Validating behavior for each of these states would be a monumental task. This can lead to some healthy skepticism towards Feature Flags from folks with a testing focus.

Happily, the situation isn't as bad as some testers might initially imagine. While a feature-flagged release candidate does need testing with a few toggle configurations, it is not necessary to test *every* possible combination. Most feature flags will not interact with each other, and most releases will not involve a change to the configuration of more than one feature flag.

So, which feature toggle configurations should a team test? It's most important to test the toggle configuration which you expect to become live in production, which means the current production toggle configuration plus any toggles which you intend to release flipped On. It's also wise to test the fall-back configuration where those toggles you intend to release are also flipped Off. To avoid any surprise regressions in a future release many teams also perform some tests with all toggles flipped On. Note that this advice only makes sense if you stick to a convention of toggle semantics where existing or legacy behavior is enabled when a feature is Off and new or future behavior is enabled when a feature is On.

If your feature flag system doesn't support runtime configuration then you may have to restart the process you're testing in order to flip a toggle, or worse re-deploy an artifact into a testing environment. This can have a very detrimental effect on the cycle time of your validation process, which in turn impacts the all important feedback loop that CI/CD provides. To avoid this issue consider exposing an endpoint which allows for dynamic in-memory re-configuration of a feature flag. These types of override becomes even more necessary when you are using things like Experiment Toggles where it's even more fiddly to exercise both paths of a toggle.

This ability to dynamically re-configure specific service instances is a very sharp tool. If used inappropriately it can cause a lot of pain and confusion in a shared environment. This facility should only ever be used by automated tests, and possibly as part of manual exploratory testing and debugging. If there is a need for a more general-purpose toggle control mechanism for use in production environments it would be best built out using a real distributed configuration system as discussed in the Toggle Configuration section above.

### Where to place your toggle

#### Toggles at the edge

For categories of toggle which need per-request context (Experiment Toggles, Permissioning Toggles) it makes sense to place Toggle Points in the edge services of your system - i.e. the publicly exposed web apps that present functionality to end users. This is where your user's individual requests first enter your domain and thus where your Toggle Router has the most context available to make toggling decisions based on the user and their request. A side-benefit of placing Toggle Points at the edge of your system is that it keeps fiddly conditional toggling logic out of the core of your system. In many cases you can place your Toggle Point right where you're rendering HTML, as in this Rails example:

someFile.erb

```erb
  <%= if featureDecisions.showRecommendationsSection? %>
    <%= render 'recommendations_section' %>
  <% end %>
```

Placing Toggle Points at the edges also makes sense when you are controlling access to new user-facing features which aren't yet ready for launch. In this context you can again control access using a toggle which simply shows or hides UI elements. As an example, perhaps you are building the ability to [log in to your application using Facebook](https://developers.facebook.com/docs/facebook-login) but aren't ready to roll it out to users just yet. The implementation of this feature may involve changes in various parts of your architecture, but you can control exposure of the feature with a simple feature toggle at the UI layer which hides the "Log in with Facebook" button.

It's interesting to note that with some of these types of feature flag the bulk of the unreleased functionality itself might actually be publicly exposed, but sitting at a url which is not discoverable by users.

#### Toggles in the core

There are other types of lower-level toggle which must be placed deeper within your architecture. These toggles are usually technical in nature, and control how some functionality is implemented internally. An example would be a Release Toggle which controls whether to use a new piece of caching infrastructure in front of a third-party API or just route requests directly to that API. Localizing these toggling decisions within the service whose functionality is being toggled is the only sensible option in these cases.

### Managing the carrying cost of Feature Toggles

Feature Flags have a tendency to multiply rapidly, particularly when first introduced. They are useful and cheap to create and so often a lot are created. However toggles do come with a carrying cost. They require you to introduce new abstractions or conditional logic into your code. They also introduce a significant testing burden. Knight Capital Group's [$460 million dollar mistake](http://dougseven.com/2014/04/17/knightmare-a-devops-cautionary-tale/) serves as a cautionary tale on what can go wrong when you don't manage your feature flags correctly (amongst other things).

Savvy teams view the Feature Toggles in their codebase as inventory which comes with a carrying cost and seek to keep that inventory as low as possible. In order to keep the number of feature flags manageable a team must be proactive in removing feature flags that are no longer needed. Some teams have a rule of always adding a toggle removal task onto the team's backlog whenever a Release Toggle is first introduced. Other teams put "expiration dates" on their toggles. Some go as far as creating "time bombs" which will fail a test (or even refuse to start an application!) if a feature flag is still around after its expiration date. We can also apply a Lean approach to reducing inventory, placing a limit on the number of feature flags a system is allowed to have at any one time. Once that limit is reached if someone wants to add a new toggle they will first need to do the work to remove an existing flag.

------

## Acknowledgements

Thanks to Brandon Byars and Max Lincoln for providing detailed feedback and suggestions to early drafts of this article. Many thanks to Martin Fowler for support, advice and encouragement. Thanks to my colleagues Michael Wongwaisayawan and Leo Shaw for editorial review, and to Fernanda Alcocer for making my diagrams look less ugly.