---
title: Git Hooks
date: 2019-05-30 23:11:58
tags:
- git
- hooks
categories:
- Git
---

本系列前几篇文章讲了许多理论，如何保持分支整洁，如何撰写合理的commit message等等。本文不再多谈理论，而是将引入一项 git built-in 的强大功能 - **hooks**。

与我们所知的其他软件或系统的 hooks 一致，git hooks 也是一种类似钩子函数的脚本，可以在执行某些 git 命令的前后自动触发。

上面提到了 1.脚本 2.自动，有了这两点 git 的功能就被极大的延伸了，因为有无数的用户可以根据自己的喜好在 hooks 定义的范围内进行自己的创作，来提升工作效率。正因此，github 上有非常多与 git hooks 相关的优秀项目。

### What exactly the GIT-HOOKS is ?
正如前文所述，hooks 是 git 内置的功能，能够允许用户定义脚本并在重要操作发生时被触发。hooks 分为两部分，client-side 和 server-side。client-side 主要在 git 命令操作时被触发，例如 commit、merge 等。server-side 主要在 git 服务端起作用，例如在收到 push commit 的时候被触发(本文仅涉及 client-side)。

git hooks 作为 git 的内置功能，无需额外安装。具体的脚本存放在 .git/hooks/ 下，git 默认在该目录下放置了一些示例脚本，都以 .sample 作为后缀名，这些示例不会被运行，git hooks 只会尝试运行该目录下没有后缀名的文件。此外，在运行 `git init` 命令时，hooks 会自动被创建在 .git 下。

### What kinds of HOOKS dose it provided ?
#### commit workflow 相关
1. pre-commit：
	在 `git commit` 执行前被执行（这里的执行前是指还没有进入到撰写 commit message 的阶段），通常可以用此 hook 来做一些提交前的工作，比如静态检查、运行测试等等，任何无法通过的情况都会打断 commit 命令，并给出错误原因。
  
2. prepare-commit-message
	与上一条不同，这个 hook 在已生成默认 commit message 之后，进入 message 编辑之前执行。以上解释不是很清晰，实际上这个 hook 很少会在普通提交时使用，它主要用于 merge、squash、amend 等等场景下使用，可以看到这些场景的特点是在用户输入自定义 message 之前都会默认创建 message，使用本 hook 即可对这些默认的 message 进行修改。
	
3. commit-msg
	这个 hook 在用户写完 commit message 之后触发，他可以拿到即将被提交的 message。因此，我们可以用它来对用户提交的 commit message 进行审查、编辑、处理等任何操作，任何原因无法通过的情况都会打断提交流程。

#### 其他
1. pre-rebase
	在任何 rebase 操作之前被触发，主要用于对 rebase 进行检查、控制，例如不允许 rebase 任何已经 push 过的提交等。
	
2. post-rewrite，post-checkout，post-merge
	以上三个 post-xxx 的 hook 分别会在1. 对 message 进行修改后 2. checkout 后 3. merge 后 被触发。通常都用于做一些命令完成后的工作，例如设置环境，移动文件，清空目录等。
	
### Any examples ?




### Reference
https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks
https://githooks.com/



