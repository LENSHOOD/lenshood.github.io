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

### Overview
正如前文所述，hooks 是 git 内置的功能，能够允许用户定义脚本并在重要操作发生时被触发。hooks 分为两部分，client-side 和 server-side。client-side 主要在 git 命令操作时被触发，例如 commit、merge 等。server-side 主要在 git 服务端起作用，例如在收到 push commit 的时候被触发。

git hooks 作为 git 的内置功能，无需额外安装。具体的脚本存放在 .git/hooks/ 下，git 默认在该目录下放置了一些示例脚本，都以 .sample 作为后缀名，这些示例不会被运行，git hooks 只会尝试运行该目录下没有后缀名的文件。此外，在运行 `git init` 命令时，hooks 会自动被创建在 .git 下。

### 

