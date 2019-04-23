---
title: Good Commit Message
date: 2019-04-21 22:33:04
tags:
- git
- commit
categories:
- Git
---

在先前的文章[Good Git Commit](https://lenshood.github.io/2019/04/21/good-git-commit/)中我主要描述了什么样的提交是好的提交，也介绍了好的提交该如何做的几点原则。其中，有一点提到好的提交需要正确的编写 commit message。

考虑到 commit message 是每一个开发者一定会做的事，同时我也的确写过、见过很多不规范的 commit message，这些不规范的 message 经常会影响项目的质量。

本文从此处展开，具体的介绍什么样的 commit message 才是好的、能够提升项目质量的，同时也介绍了如今很流行的 Conventional Commits。

### Commit Message 真的有这么重要吗？
以下是一段真实项目中的代码提交记录：
``` text
commit 2c6ed029f7b8cd9c83031176015d4ab56a2bac9f
Author: xxx <xxxx@xxx.com>
Date:   Thu Dec 13 17:47:27 2018 +0800

    合并代码

commit 2438ce29f37e4b3bf8856a5514d76c858b73636f
Author: xxx <xxxx@xxx.com>
Date:   Thu Dec 13 16:44:44 2018 +0800

    合并

commit 52511f2cf072fba3efcad657cb97dcf802a17714
Author: xxx <xxxx@xxx.com>
Date:   Thu Dec 13 14:36:57 2018 +0800

    解决冲突

commit 2ec266667a83bc5fc3a08562dc10db28468b159c
Author: xxx <xxxx@xxx.com>
Date:   Thu Dec 13 14:19:08 2018 +0800

    salary

commit 9633c04a93a555ac98f38d3437bfced99f781b2d
Author: xxx <xxxx@xxx.com>
Date:   Thu Dec 13 14:17:50 2018 +0800

    user

commit 8ec8246decc0975b1747709275639534bd60df7d
Author: xxx <xxxx@xxx.com>
Date:   Thu Dec 13 14:08:46 2018 +0800

    fix bug
```
显然，浏览该 log 的人并不能通过 commit message 得到任何有效的信息。

能看出提交者的每一个提交的确做了一些事情，包括新的 feature，修复 bug，合并分支等操作，但是这些想要知道具体是做了什么样的 feature，修复了什么样的 bug，解决了哪些冲突？只能通过代码对比来低效的获取。

所以，commit message 真的很重要。

一般而言，不合适的 commit message 通常存在两方面的问题：
1. 信息量太少
	- 为了提交而提交，并没有考虑读者的感受
	- 认为代码即文档，不愿意在 commit message 里面再次描述做了什么
2. 格式不规范
	- 随心所欲，写到哪里算哪里
	- 搞不清楚是 feature 还是 bug fix， 还是 refactor

### 内容应包含什么？
对于 commit message 的内容，应清晰、准确、详细，以下仍旧引用 OpenStack Wiki 的相关内容：
优秀的 commit message 应包含完整而清晰的理解和检验正确性的信息，具体来说：
1. 对于 bug fix, 不要假设 reviewer 能理解最初的问题是什么
	- 在阅读 bug report 并经过多个相关提交中来回跳转后，reviewer 很可能已经搞不清楚最初的问题是什么了。在 commit message 中包含问题描述可以方便读者进行 review。
2. 不要假设代码是不言而喻的并且能够自注释
	- 每个人对代码的理解都是不同的，确保清晰地描述功能点、bug 的问题是什么、你是如何进行修复的。
3. 描述做这样的修改的原因
	- 一个常见的错误是只说清楚了代码是怎么做的，而没有提及为什么这么做。
4. 提交记录的第一行是最重要的
	- 第一行可以起到提纲挈领的作用，使任何读者能够最快的了解的该提交做的事情。
7. 在提交记录中包含任何可能的限制情况
	- 如果该提交可能还会有后续的进一步优化，应在 message 中体现。
8. 不仅包含适合人类阅读的内容，也应包含用于机器快速检索的内容
	- 可以方便检索和归档，常见的有版本号、issue id、bug id 等等

### 格式应规范什么？
对于格式通常是灵活的，每个项目可以有不同的规定，只要在项目内保持统一即可。

通常，规范的格式会包括
- 提交类型：feature，bug fix， refactor 等
- 排版规范：标题、内容、每行长度等
- 其他信息：包括各种编号等等

以下举两个开源项目的规范，作为例子：

1. Spring Boot
> Capitalized, short (50 chars or less) summary
> 摘要部分大写，尽量短(50 个字符以内)
>
> More detailed explanatory text, if necessary.  Wrap it to about 72
> characters or so.  In some contexts, the first line is treated as the
> subject of an email and the rest of the text as the body.  The blank
> line separating the summary from the body is critical (unless you omit
> the body entirely); tools like rebase can get confused if you run the
> two together.
> 如必要提供能详细的解释性文字，并最多 72 个字符换行。在某些情况下，第一行会被
> 自动识别为题目，之后才识别为主题，因此用空行将摘要和详细分开很重要。如果混在
> 一起，类似 rebase 之类的工具可能会产生混淆。
>
> Write your commit message in the imperative: "Fix bug" and not "Fixed bug"
> or "Fixes bug."  This convention matches up with commit messages generated
> by commands like git merge and git revert.
> 采用祈使句来描述，即使用“Fix bug” 而不是 “Fixed bug” 或 “Fixes bug”。此约定和命令
> 自动生成的文字保持一致(类似 git merge， git revert 等命令)。
>
> Further paragraphs come after blank lines.
> 更多的段落一样需要用空行隔开。
>
> - Bullet points are okay, too
> - 可以使用圆点符
> 
> - Typically a hyphen or asterisk is used for the bullet, followed by a
> single space, with blank lines in between, but conventions vary here
> - 用连字符和星号来表示项目（和 markdown 一样），后面跟单空格，中间有空行，不过该约定可能会有差异
>
> - Use a hanging indent
> - 使用悬挂缩进(一种缩进方式，段落除第一行外其余行都缩进)


2. OpenStack
> Provide a brief description of the change in the first line.
> 在第一行提供对修改的简短描述。
> 
> Insert a single blank line after the first line.
> 第一行和后文之间插入空行。
> 
> Provide a detailed description of the change in the following lines, breaking paragraphs where needed.
> 在后续行中提供对修改的详细描述，必要时可以有多个段落。
> 
> The first line should be limited to 50 characters and should not end with a period.
> 第一行应限制在 50 个字符以内，不要以句号结尾。
> 
> Subsequent lines should be wrapped at 72 characters.
> 后文应在 72 个字符后换行。
> 
> Put the 'Change-id', 'Story: xxxx', and 'Task: yyyy' (or 'Closes-Bug #NNNNN' and 'blueprint NNNNNNNNNNN' if the project still uses Launchpad) lines at the very end.
> 将 'Change-id', 'Story: xxxx', 和 'Task: yyyy' 等信息放在最后面。

### Conventional Commits
前文已经提到，对于规范，不同项目可能会根据其项目特点以及习惯，提出不同的要求。对大型的开源项目，其生命周期长，版本发布多，一旦规范确定下来就很难再更改了(例如 Spring Boot 的规范显示发布时间是 2008 年)。

然而基于对规范和约束的大量需求，逐步出现了一种通用的方案，适应性强，使用和记忆都很简单，这就是接下来要提到的：[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0-beta.4/)。

Conventional Commits 称自己为：

**A specification for adding human and machine readable meaning to commit messages** 

#### 规范格式
``` text
<type>[optional scope]: <description>

[optional body]

[optional footer]
```

符合 Conventional Commits 规范的 commit message 应符合以上格式。

简而言之，在 commit message 第一行摘要中应首先包含此次提交的类型type(详见下述)，可选择在 scope 中描述额外的上下文信息(根据项目的不同，scope 会有不同的定义与约束)，最后包含简短的描述。

如有需要，可以在 body 处添加更详细的描述段落。允许存在多个详述段落。

最后，如有需要，可以在 footer 处添加与此次提交相关的元数据，例如关联的 request-id，bug-id 等。

#### type
Conventional Commits 并没有限制 type 的范围，但限定了至少两种 type：
- feat：提交内容包含了新功能
- fix：提交内容包含了 bug 修复

除此之外，通常还存在以下几种常用的 type（被多个开源项目采用）：
- docs：对文档的修改
- style：对格式的调整以及符号的修改等
- refactor： 不改变外部行为的重构
- test：增加测试，重构测试等，不改变产品代码
- chore：零星工作，不改变产品代码
- build：对编译系统的修改，例如 npm，maven 等
- ci：对 CI 配置的修改，如 Travis，Jenkins 等
- perf：对性能提升的修改(可简略的归类至 refactor)

#### BREAKING CHANGE
可选择在 body 的最前面或是 footer 的最前面增加 BREAKING CHANGE 以示该提交的修改非常重要，可能会影响到开发者或使用者。

BREAKING CHANGE 可以与任何 type 相配合，此外还允许在包含 BREAKING CHANGE 的提交 type 后加感叹号(！) 以便于识别。

#### Examples
1. Commit message with description and breaking change in body
``` text
feat: allow provided config object to extend other configs

BREAKING CHANGE: `extends` key in config file is now used for extending other config files
```
2. Commit message with optional ! to draw attention to breaking change
``` text
chore!: drop Node 6 from testing matrix

BREAKING CHANGE: dropping Node 6 which hits end of life in April
```
3. Commit message with no body
``` text
docs: correct spelling of CHANGELOG
```
4. Commit message with scope
``` test
feat(lang): add polish language
```
5. Commit message for a fix using an (optional) issue number.
``` text
fix: correct minor typos in code

see the issue for details on the typos fixed

closes issue #12
```

> 其他支持 Conventional Commits 的开源项目的具体约束：
> angular：https://github.com/angular/angular/blob/22b96b9/CONTRIBUTING.md#-commit-message-guidelines
> karma：http://karma-runner.github.io/3.0/dev/git-commit-msg.html

### 参考
1. [Git Commit Good Practice](https://wiki.openstack.org/wiki/GitCommitMessages)
2. [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)
3. [A Note About Git Commit Messages](https://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html)
4. [KARMA Git Commit Msg](http://karma-runner.github.io/3.0/dev/git-commit-msg.html)
5. [Angular Commit Message Guidelines](https://github.com/angular/angular/blob/22b96b9/CONTRIBUTING.md#-commit-message-guidelines)
6. [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0-beta.4/) 
