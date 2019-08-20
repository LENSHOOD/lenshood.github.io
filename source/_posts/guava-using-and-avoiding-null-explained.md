---
title: 使用和避免使用 null
date: 2019-08-20 23:05:17
tags:
- guava
- translation
categories:
- Guava
---

## 使用和避免使用 null
> "null 烂透了“ - [Doug Lea](http://en.wikipedia.org/wiki/Doug_Lea)
> "这是我犯得值 10 亿刀的错误" - [Sir C. A. R. Hoare](http://en.wikipedia.org/wiki/C._A._R._Hoare) 提到他发明的 null 时如是说

对`null`的粗心使用会导致各种各样令人难以置信的错误。对 Google 的基础代码进行研究后，我们发现大约 95% 的集合中都不应有任何 null 值，如果对 `null` 快速失败而不是默默地接受，便会对开发者有所帮助。

此外，`null` 也会产生令人不悦的模糊情况。通常很少有能准确的揣摩到返回`null`值原本意义场景，例如，当 `Map.get(key)`返回`null` 时，一种可能是 key 对应的值本来就是`null`，而另一种情况则是该 key 并不存在。`Null` 能代表失败，能代表成功，能代表几乎任何事。如果能用其他的值来代替`null`，会使表意更加清晰。

