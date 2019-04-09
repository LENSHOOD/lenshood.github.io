---
title: 如何保持公共开发分支的整洁
date: 2019-04-09 00:54:50
tags: Git Rebase Squash-Merge
---

### 糟糕而难看的公共分支
在日常开发中，我经常会见到某些项目的 master 或 dev 等公共分支，其 commit log 混乱，并且存在各种无意义的 branch merge，举个例子，见如下提交记录：
```
commit 950c55ba3652d3d7d704ad349fc95d6edb62570b (HEAD -> dev, origin/dev)
Merge: e32f3f3 f566f4f
Author: ****** <******@example.com>
Date:   Thu May 17 11:50:46 2018 +0800

    Merge branch 'exception' into 'dev'

    异常分类统计bug

    See merge request !83

commit f566f4f21e39ee673eacf6a4f50b1f83593a3d99 (origin/exception)
Author: ****** <******@example.com>
Date:   Thu May 17 11:49:45 2018 +0800

    异常分类统计bug

commit e32f3f372a2fb1468294048848bcb935a6fe1648
Merge: 6e5d8fd 6f0b8ae
Author: ****** <******@example.com>
Date:   Wed May 16 17:09:31 2018 +0800

    Merge branch 'exception' into 'dev'

    异常分类统计

    See merge request !81

commit 6f0b8ae9e9e705a9e0e9e8fcb3fe4596f32f3f49
Author: ****** <******@example.com>
Date:   Wed May 16 17:07:50 2018 +0800

    异常分类统计

commit 6e5d8fd148d12947538630cb33f5b4f5689c215c
Merge: c7d4f71 806f4bf
Author: ****** <******@example.com>
Date:   Wed May 16 16:03:06 2018 +0800
Merge branch 'song-new' into 'dev'

    消息中心及分组

    See merge request !80

commit ae697ed784c7d272e9db16129c98bc93e32bfdb0
Author: ****** <songzhengfu@example.com>
Date:   Thu May 10 16:28:55 2018 +0800

    消息中心及分组

commit c7d4f71a03a76a7f5f5eb21ae1829469fb75c492
Author: ****** <******@example.com>
Date:   Wed Apr 25 15:13:09 2018 +0800

    fix bug

commit a627d344b47231515c7900f65fcdd0492039ba42
Author: ****** <******@example.com>
Date:   Tue Apr 24 20:00:22 2018 +0800

    fix bug：未完成
```

例子是截取的某真实项目的 commit log，可以发现在提交记录中存在两个问题：
1. 满篇的 “Merge Branch XX to XXX”
2. 开发时的临时性 commit 被一并提交至公共分支，多数 commit message 是无效信息(参见最后两条 commit)

以上的两个问题大量的出现在各种项目中，导致公共分支的 commit log 非常不雅观，且难以 revert 到适当的位置，随着项目的不断开发，持续恶化，最终彻底搞不懂过去都提交了些什么。

### 我在工作中常用的分支开发实践
针对以上两个问题，我在工作中通常采用对应的两种办法解决，都非常简单，也容易记忆，下面一一展示。

**1. 大量无意义的 merge branch 的问题**

产生 merge branch 的主要原因是公共分支有了新提交，在 pull 的时候，和本地提交进行了合并，类似下图的过程：
![](https://git-scm.com/book/en/v2/images/basic-merging-2.png)

图中，本地基于 C2 提交了 C3 和 C5， 远端有第三者基于 C2 提交了 C4，因此 pull 的时候，git 将二者合并后自动提交为 C6，C6 的 message 即 “Merge Branch XX to XXX”。这种方式的好处是可以保留 C3 - C5 的分叉-合并记录，合并后可以完全的 revert 回两个分支。

通常，我们鼓励更细粒度的故事卡拆分，反映在代码上，scope 较小的卡可以形成一个或数个提交，因此本地的代码与公共分支一般不会有过大的不同。既然如此，试想我们是否可以将 C3 和 C5 直接挪动到 C4 的后面，并尝试解决挪动过程中产生的 conflict ？其实这便是一个标准的 rebase 的操作。

举例说明：
现有一 demo 项目，lenshood 和 old-wang 共同开发，当前 old-wang 在 master 分支上提交了一条记录：

```
commit 19e0a8cc9921fda34f6a4e6ecd12379e1bf5fc5f (HEAD -> master, old-wang)
Author: old-wang <old-wang@example.com>
Date:   Tue Apr 9 01:38:43 2019 +0800

    Old Wang say hi

commit 8bfc1498829f7abce5818accd0e9257ecc198dd8 (lenshood)
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:36:56 2019 +0800

    Initial commit
```

此时 lenshood 基于 Initial commit，在本地 master 进行了开发，如果采用直接 pull 的方式，则最终产生如下记录：
```
commit 9ab9b0b981bcf02d14746fb70727abbea3822723 (HEAD -> master)
Merge: 19e0a8c 7dd24b2
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:44:29 2019 +0800

    Merge branch 'lenshood' //解决冲突后的提交记录

commit 7dd24b27f7e6330c2886b89eef9982c09a68bcc4 (lenshood)
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:43:47 2019 +0800

    Lenshood say hi

commit 19e0a8cc9921fda34f6a4e6ecd12379e1bf5fc5f (old-wang)
Author: old-wang <old-wang@example.com>
Date:   Tue Apr 9 01:38:43 2019 +0800

    Old Wang say hi

commit 8bfc1498829f7abce5818accd0e9257ecc198dd8
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:36:56 2019 +0800

    Initial commit
```

假如 lenshood 采用另一种方式：`git pull origin master --rebase`，则可得到如下提交记录：
```
commit 472f04f94cf325ab49d06574f27aa0c7385af4ae (HEAD -> master, lenshood)
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:56:19 2019 +0800

    Lenshood say hi

commit 19e0a8cc9921fda34f6a4e6ecd12379e1bf5fc5f (old-wang)
Author: old-wang <old-wang@example.com>
Date:   Tue Apr 9 01:38:43 2019 +0800

    Old Wang say hi

commit 8bfc1498829f7abce5818accd0e9257ecc198dd8
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:36:56 2019 +0800

    Initial commit
```

可见，通过这种方式，lenshood 的提交记录已经被挪动到 master 的头部，不会再出现难看的 “Merge Branch” 了。

> 再次回顾该方法： `git pull origin master --rebase`



**2. 大量无意义的中间提交记录的问题**

产生这种 commit log 的主要原因是将开发中的许多中间提交记录一并合入了主干。在采用 Git 作为版本控制系统的开发实践中，经常性的提交是一件好事，因为在 Git 的提交之间移动、操作非常方便。

比如某 feature 中，一文件可能会被修改两处，我们期望调试时在不修改，只修改一处和同时修改两处这三种情况之间切换，那么就会存在两个中间提交。 有时甚至开发到一半去吃饭，也会顺手加上 WIP 并提交。

这种中间提交，很多时候我们不会过于注重其 commit message 的格式、意义、规范等，毕竟本地开发的 scope 只有自己，只要达到快速创建记录点的目的就足够了。然而假如这种模棱两可的提交污染到了公共分支，则会影响到其他开发者的理解。(就好像把 private 方法全部改成了 public，会对他人在理解上造成误导)。

仍然以上文的提交记录为例：
```
commit 472f04f94cf325ab49d06574f27aa0c7385af4ae (backup)
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:56:19 2019 +0800
  
Lenshood say hi

commit 19e0a8cc9921fda34f6a4e6ecd12379e1bf5fc5f
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:38:43 2019 +0800

    Old Wang say hi

commit 8bfc1498829f7abce5818accd0e9257ecc198dd8
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:36:56 2019 +0800

Initial commit
```

此时 lenshood 做了某个重要的故事卡，并合入了主干：
```
commit 1465d364368772046210ad25ed49d649e07e7494 (HEAD -> master, l)
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 23:13:38 2019 +0800

    lenshood go back to work

commit e241b144629777c0eceb0de01c58d3b9eb9e4ae1
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 23:13:08 2019 +0800

    WIP：lenshood get shit done

commit c1506762ae897c4796dbd362ffa8cb1c2568dcf0
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 23:12:33 2019 +0800

    WIP：lenshood take off his pants

commit 4f79fded6c01b8c538d4d01da0d1fbb4445bdf13
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 23:11:29 2019 +0800

    WIP：lenshood walk into toilet

commit 167ff9e5aeab31659ee8c4b483ba4e12154bb029
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 23:09:35 2019 +0800

    lenshood wants to shit

commit 472f04f94cf325ab49d06574f27aa0c7385af4ae (backup)
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:56:19 2019 +0800

    Lenshood say hi

commit 19e0a8cc9921fda34f6a4e6ecd12379e1bf5fc5f
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:38:43 2019 +0800

    Old Wang say hi

commit 8bfc1498829f7abce5818accd0e9257ecc198dd8
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:36:56 2019 +0800

    Initial commit
```

可见大量的中间工作占据并污染了主干，造成了许多包含 WIP 的提交记录。
尝试采用 `git merge lenshood --squash`，squash 这个参数非常形象，就好像把许多提交挤在了一起。在执行该操作后，代码的修改已经合并，但只处于 stage 状态，并未形成提交，需要手动进行 commit 操作，同时编写合理的 commit message，结果如下：

```
commit fc953f85395b2f78b02d05e48443ab49bd89f3c1 (HEAD -> master)
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 23:21:59 2019 +0800

    Lenshood enjoys shit
    - walk into toilet
    - take off pants
    - get shit done
    - go back to work

commit 472f04f94cf325ab49d06574f27aa0c7385af4ae (backup)
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:56:19 2019 +0800

    Lenshood say hi

commit 19e0a8cc9921fda34f6a4e6ecd12379e1bf5fc5f
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:38:43 2019 +0800

    Old Wang say hi

commit 8bfc1498829f7abce5818accd0e9257ecc198dd8
Author: lenshood <lenshood@example.com>
Date:   Tue Apr 9 01:36:56 2019 +0800

    Initial commit
```
结果一目了然，不再过多赘述。

> 再次回顾该方法： `git merge {branch} --squash`