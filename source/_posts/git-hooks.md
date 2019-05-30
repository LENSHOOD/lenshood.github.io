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

