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

### Toggling 小故事

想象这样的场景。有一个复杂的城市规划仿真游戏项目，你供职于其多个项目组中的一个。你的团队负责核心仿真引擎。而你的任务是优化提升 Spline Reticulation 算法的效率。你心里清楚这种优化需要对现有实现进行相当大的改造，而这需要花费数周时间。同时，其他团队成员仍旧需要在与该算法相关的代码基础上继续一些正在进行中的工作。

基于以往合并长寿分支（long-lived branches）的痛苦体验，如果可能的话，这次你想要避免将这项工作进行分支。相反，你决定整个团队仍旧会基于主干进行工作，但对 Spline Reticulation 算法进行优化的开发者们将会使用 Feature Toggle 来防止他们的工作影响到其他成员，并防止对代码库产生不稳定。

#### Feature Flag 的诞生

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

#### 让 Toggle Flag 变得更加动态

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

#### 准备好要发布了

更多的时间过去了，现在团队相信新的算法已经完成了功能。为了确认这一点，他们已经修改了高层的自动化测试，使得系统能在包含新功能与未包含新功能两种条件下受到验证。另外团队也想要做一些人工的试验性测试来确保所有功能都运行的与期望保持一致，毕竟，Spline Reticulation 是系统行为中非常关键的一部分。

为了对一个尚未被验证为可供一般使用的功能进行人工测试，我们需要能让该功能在生产环境上对一般大众用户关闭，而对内部用户开启。为了实现这一目标，有很多种办法：

- 让 Toggle Router 来基于 **Toggle Configuration** 做出决策，并且让这一 configuration 作用于特定的环境。只在预生产环境开启新功能。
- 通过某种类型的管理界面来允许 Toggle Configuration 能被实时修改。使用管理界面来在测试环境开启新功能。
- 教会 Toggle Router 如何动态的对每一次请求做出决策。这种决策方式引入了 **Toggle Context** 的概念，例如寻找某个特定的 Cookie 或者 Http Header。通常 Toggle Context 会被用作一个代理，来识别发出请求的用户。

（我们将会在之后更深入的探讨上述实现的细节，所以如果你不熟悉这些概念，也别担心。）

![](https://martinfowler.com/articles/feature-toggles/overview-diagram.png)

团队决定使用基于单请求的 Toggle Router，因为这种方式非常灵活。特别令人欣赏的是，这种方案允许他们不需要独立的测试环境，就能测试新算法了。取而代之的，他们只要简单的将新算法在生产环境打开，但只对内部用户开启（通过探测特定的 cookie）。现在团队能够将这种 cookie 加在他们自己身上，之后对新功能进行验证来确定其是否符合预期了。

#### 金丝雀发布

新的 Spline Reticulation 算法基于试验性测试完成后，目前看起来工作良好。然而由于这个算法属于仿真游戏引擎中非常重要的一部分，所以他们对将新算法开放给所有用户仍然有些不情愿。团队决定使用 Feature Flags 基础设施来实施[**金丝雀发布**](https://martinfowler.com/bliki/CanaryRelease.html)，只将新功能开放给占总用户数很少量百分比的用户 -- 一个“金丝雀”用户群。

团队增强了 Toggle Router，教给他用户群的概念 -- 一个用户组，他们始终体验到某个功能始终处于打开或关闭态。一个金丝雀用户群，是通过随机采样所有用户中 1% 的用户来创建的（也许可以用对用户 ID 取模的办法）。金丝雀用户群将会持续的体验新功能开启，而其他 99% 的用户仍旧会使用旧的算法。核心业务指标（用户参与度，总收入等等）会同时在两个用户组中被监控，以此来确定新算法不会对用户行为产生负面影响。一旦团队确信新功能不会产生任何不良影响，他们就会将 Toggle Configuration 修改为对整个用户群打开新功能。

#### A/B 测试

团队的产品经理在得知新功能顺利发布后很兴奋。她建议团队使用类似的机制来实施 A/B 测试。关于修改犯罪率算法以考虑污染程度是否会增加或降低游戏可玩性的争论由来已久。他们现在有能力用数据来结束这场争论了。他们计划推出一个抓住了这个想法本质的简单实现，用 Feature Toggle 来控制。他们将对一个相当大的用户群体打开这一新功能，然后研究这些用户相比“控制”组用户的行为。这一实践能允许团队基于数据而不是 [HiPPOs](http://www.forbes.com/sites/derosetichy/2013/04/15/what-happens-when-a-hippo-runs-your-company/) 来解决持久的产品辩论。

上述简单场景不仅是为了展现 Feature Toggle 的基本概念，更说明了这一核心能力可以有多少种不同的应用。现在我们已经接触了一些应用的例子了，让我们更进一步。我们将探索不同种类的 toggle，并且观察他们的不同点。我们会涉及如何编写易于维护的 toggle 代码，最后会分享一些实践，来帮助你避免 feature-toggle 系统的一些陷阱。