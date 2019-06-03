---
title: Good Git Commit
date: 2019-04-21 13:16:47
tags: 
- git
- commit
categories:
- Git
---

规范 Git 系列：
[第一篇：如何保持公共开发分支的整洁](https://lenshood.github.io/2019/04/08/keep-git-branch-clean/)
[第二篇：Good Git Commit](https://lenshood.github.io/2019/04/21/good-git-commit/)
[第三篇：Good Commit Message](https://lenshood.github.io/2019/04/21/conventional-commit-message/)
[第四篇：Git Hooks](https://lenshood.github.io/2019/05/30/git-hooks/)

---

在前一篇文章 [如何保持公共开发分支的整洁](https://lenshood.github.io/2019/04/08/keep-git-branch-clean/)中，提到了一些不规范的 git 提交，本文将进一步介绍什么样的提交才是优秀的 git 提交，以及怎么样规范 git 提交。

### 为什么要保持规范的 git 提交？
主要和项目规模相关，与上古时代一个程序员单挑整个软件项目不同，随着软件规模的急剧扩大，现代的软件都是以团队为单位开发的。考虑到团队协作的效率，如何将自己所承担的部分工作独立的、简介的、易于插拔的合并入软件主干代码，是团队中每个程序员都应该考虑的问题。

杂乱无章且关联错乱的代码提交不仅会影响他人对开发进度的了解，也会导致版本发布、bug 修复、revert 等工作难以进行。

以数人或数十人小团队为单位的开发团队，commit 对效率与质量的影响已经如此敏感，更别说更松散的、数百上千人参与的开源项目了。因此大多数大型的开源项目，都会规定一套代码规范(Code of Conduct)，其中包括如何 pull request，如何提交 patch，当然也包括 -- 很重要的 -- 如何创建合适的提交。

### 什么样的提交是优秀的提交
简而言之一句话：
**优秀的 commit 要保证每一次提交只有一个逻辑更改**

> (good commits is to ensure there is only one "logical change" per commit，来自 OpenStack 的 wiki：[Git Commit Good Practice](https://wiki.openstack.org/wiki/GitCommitMessages))

逻辑更改可以是某个 functional change，也可以是某个小的 feature，其核心就是在保证功能完整性的前提下，不可再次拆分，类似原子操作的概念。

保证只包含一个逻辑提交，是因为：
- 代码修改量越小，越快且越容易对其进行 review 并识别潜在缺陷。
- 假如代码修改确实导致了缺陷，则在对这次有缺陷的提交进行 revert 的时候，其中没有代码与其他提交混杂也会使 revert 更容易。
- 当使用 [git bisect](https://git-scm.com/docs/git-bisect) 功能进行故障排除时，定义良好的小更改有助于准确的隔离引入的代码问题。
- 当使用 [git annotate/blame](https://git-scm.com/docs/git-annotate) 时，定义良好的小更改有助于准确的隔离一段代码的来源和原因。

### 怎么做？
- 经常提交
	- 经常性的提交能够不时地让我们思考当前的修改是否一个可以完成独立的功能，也能帮助我们的提交符合 logical change 的要求。
	- 如果习惯酣畅淋漓的狂写一整天代码，等到 feature 完成后才进行提交，那就设定个番茄钟，每隔一小时让自己休息一下，顺便整理代码看看是否能做一次提交吧！
- 不要把两个不相关的 functional change 混在一起
	- 违反了 logical change 的原则，相关度低的修改混在一起提交会使得 review 时不容易发现缺陷， revert 也会非常麻烦。
- 不要将较大的 feature 修改作为单个提交
	- 我们经常会遇到这样的现象：对某个 feature 只有所有的代码都存在时才可用，否则提交上去的代码会使整个软件不可用，然而这并不意味着所有的代码应该在一次提交中提供。
	- 首先，通常新的 feature 都会重构原有的代码，将重构的部分独立提交将非常有利于 review 并测试该重构是否对原有功能有影响。
	- 此外，将新的 feature 拆分为多个提交也便于进行独立的 review，同时在该 feature 未合并之前，还便于其他开发人员选取出其中最佳的（cherry-pick）一部分进行合入。
	- 最后，如果是一个大的 feature，为什么不能拆分为多个独立的功能块呢？是不是代码耦合太高？看看是不是代码已经违反了 [SOLID 原则](https://en.wikipedia.org/wiki/SOLID)。
- 不要提交做了一半的工作
	- WTF？这条是对上一条的打脸操作吗？高中政治课本说过：矛盾都是对立与统一的。
	- 不要提交过多的工作，也不要提交做了一半的工作。那么如何平衡呢？还是它：logical change。确保每个提交都是一个 logical change，就可确保既不是未完成的工作，也不是单个大提交。
- 确保该提交进行了完全的测试
	- 测试不全的提交的结果很可能是返工修改 + `git commit —amend`。
	- 如果这条提交已经 push 了，那么等待着的就只有新增一个 bug fix commit 了。
	- 我不止一次看到过 commit log 里面写到：“fix stupid bug of forget to add/delete xxx”。 (谁说我自己就没这么干过？？)
- 正确的编写 commit message
	- 除了阅读代码，对某个修改的第一印象及描述通常都会通过 commit message 体现，模糊不清或是不规范的 message 会使后续回溯变得困难。
	- 对 commit message 的要求通常是清晰、准确、包含上下文的说明。
	- 对于成熟或大型的项目，会有专门的 commit convention 来具体的规范 message。
- push 到远程分支时应慎重
	- 一旦 push 原则上无法回退，因为这会影响到所有 pull 了该提交的开发者。
	- 试问在群里被所有人 diss 是什么感觉XD。

### Example
这部分还是引用 OpenStack wiki 中所示的示例来进行说明：
#### 不好的实践
``` text
commit ae878fc8b9761d099a4145617e4a48cbeb390623
  Author: [removed]
  Date:   Fri Jun 1 01:44:02 2012 +0000

    Refactor libvirt create calls

     * minimizes duplicated code for create
     * makes wait_for_destroy happen on shutdown instead of undefine
     * allows for destruction of an instance while leaving the domain
     * uses reset for hard reboot instead of create/destroy
     * makes resume_host_state use new methods instead of hard_reboot
     * makes rescue/unrescue not use hard reboot to recreate domain

    Change-Id: I2072f93ad6c889d534b04009671147af653048e7
```
首先，以上提交显然存在两方面的修改：
1. The switch to use the new "reset" API for the "hard_reboot" method (message 第四条)
2. The adjustment to internal driver methods to not use "hard_reboot" (message 第五、六条)
因此存在以下的问题：
1. 首先并没有什么有说服力的原因需要将以上两方面的修改放在一起，应该改为两个提交，第一个提交将多处名为 hard_reboot 的调用替换，第二个提交重写 hard_reboot 方法。
2. 其次，使用 libvirt reset 方法的相关内容隐藏在了大的 “Refactor” 信息中，对于 reviewer 很可能忽略了实际上这个提交引入了一个新的 libvert API 的依赖。如果这个提交包含缺陷，然而由于包含了各种不相关的修改，因此无法进行简单的 revert。

``` text
commit e0540dfed1c1276106105aea8d5765356961ef3d
  Author: [removed]
  Date:   Wed May 16 15:17:53 2012 +0400

    blueprint lvm-disk-images

    Add ability to use LVM volumes for VM disks.

    Implements LVM disks support for libvirt driver.

    VM disks will be stored on LVM volumes in volume group
     specified by `libvirt_images_volume_group` option.
     Another option `libvirt_local_images_type` specify which storage
     type will be used. Supported values are `raw`, `lvm`, `qcow2`,
     `default`. If `libvirt_local_images_type` = `default`, usual
     logic with `use_cow_images` flag is used.
    Boolean option `libvirt_sparse_logical_volumes` controls which type
     of logical volumes will be created (sparsed with virtualsize or
     usual logical volumes with full space allocation). Default value
     for this option is `False`.
    Commit introduce three classes: `Raw`, `Qcow2` and `Lvm`. They contain
     image creation logic, that was stored in
     `LibvirtConnection._cache_image` and `libvirt_info` methods,
     that produce right `LibvirtGuestConfigDisk` configurations for
     libvirt. `Backend` class choose which image type to use.

    Change-Id: I0d01cb7d2fd67de2565b8d45d34f7846ad4112c2
```
上述提交引入了一个主要的特性，所以表面上看起来采用一个 commit 很合理，但是看看补丁就会发现，提交者将大量的代码重构工作与新的 LVM 特性揉在了一起。这使得很难去回溯整个支持QCow2/Raw 镜像特性的过程。该过程至少可以分为四个小提交：
1. Replace the 'use_cow_images' config FLAG with the new FLAG 'libvirt_local_images_type', with back-compat code for support of legacy 'use_cow_images' FLAG
2. Creation of internal "Image" class and subclasses for Raw & QCow2 image type impls
3. Refactor libvirt driver to replace raw/qcow2 image management code, with calls to the new "Image" class APIs
4. Introduce the new "LVM" Image class implementation


#### 好的实践
``` text
commit 3114a97ba188895daff4a3d337b2c73855d4632d
  Author: [removed]
  Date:   Mon Jun 11 17:16:10 2012 +0100

    Update default policies for KVM guest PIT & RTC timers

  commit 573ada525b8a7384398a8d7d5f094f343555df56
  Author: [removed]
  Date:   Tue May 1 17:09:32 2012 +0100

    Add support for configuring libvirt VM clock and timers
```
以上两条提交共同实现了对 KVM guest timer 提供配置支持。引入新增 API 创建 libvert XML 配置的功能与通过新增 API 修改 KVM guest 创建策略的功能清晰的分离开了。

``` text
commit 62bea64940cf629829e2945255cc34903f310115
  Author: [removed]
  Date:   Fri Jun 1 14:49:42 2012 -0400

    Add a comment to rpc.queue_get_for().
    Change-Id: Ifa7d648e9b33ad2416236dc6966527c257baaf88

  commit cf2b87347cd801112f89552a78efabb92a63bac6
  Author: [removed]
  Date:   Wed May 30 14:57:03 2012 -0400

    Add shared_storage_test methods to compute rpcapi.
...snip...
    Add get_instance_disk_info to the compute rpcapi.
...snip...
    Add remove_volume_connection to the compute rpcapi.
...snip...
    Add compare_cpu to the compute rpcapi.
...snip...
    Add get_console_topic() to the compute rpcapi.
...snip...
    Add refresh_provider_fw_rules() to compute rpcapi.
...many more commits...
```
以上一系列提交将整个 nova 内的 RPC API 层重构，使之允许采用可插拔的消息传递实现。对于这种对核心功能有重大变化的修改，将工作分为一个大的提交序列的方式，是能够对每一个步骤进行有意义的 code review，track/identify regression 的关键。



### 参考
[OpenStack Wiki：Git Commit Good Practice](https://wiki.openstack.org/wiki/GitCommitMessages)
[Git Commit Best Practices](https://github.com/trein/dev-best-practices/wiki/Git-Commit-Best-Practices)

---

规范 Git 系列：
[第一篇：如何保持公共开发分支的整洁](https://lenshood.github.io/2019/04/08/keep-git-branch-clean/)
[第二篇：Good Git Commit](https://lenshood.github.io/2019/04/21/good-git-commit/)
[第三篇：Good Commit Message](https://lenshood.github.io/2019/04/21/conventional-commit-message/)
[第四篇：Git Hooks](https://lenshood.github.io/2019/05/30/git-hooks/)