---
title: 在 IntelliJ Idea 中使用 Vim
date: 2019-06-28 23:44:05
tags:
- vim
- IntelliJ idea
categories:
- Others
---

### 码字利器 提升快感
Intellij Idea 作为写 Java 最爽、最快、最智能的 IDE，其丰富的功能和完善的快捷键让 Java Coder 可以完全采用键盘流的方式写代码，并且做各种额外的事情(包括在终端跑命令、VCS、文件操作等等)都不用切出 IDE。

在使用的时候完全感觉不到它的存在，才是最棒的工具，Idea 可以做到这一点，Vim 也能做到这一点，这两种完全不同的 editor 用习惯了后都能手随心动，行云流水，优秀的 editor 更容易让使用者进入 [心流(flow)](https://en.wikipedia.org/wiki/Flow_(psychology)) 的工作状态。

相比之下，Idea 的优势在于能够快速的在 interface、实现类、方法等等之间跳转，能够方便快速的重构，及通过代码模板自动生成样板代码；Vim 的优势在于高效的文字编辑、处理能力。在Idea 编码的时候经常会不由自主的想要按一下 gg 或是 jk 等等 Vim 的按键来快速的回到页首或是上下移动。

实际上，IdeaVim 作为 IntellliJ Idea 的插件，能实现快速的在 Idea 中使用 Vim 的能力，结合二者的优势，让编码更顺滑。

<!-- more -->

### 使用 IdeaVim

1. 通常初次安装 idea 时推荐插件中就有 ideaVim，可见官方对其的认可，若初次安装时没有装，那么也可以进入 settings -> plugins 搜索ideaVim 下载安装重启后即可使用了。

2. 开启/关闭 ideaVim 的按钮在 Tools -> Vim Emulator，默认情况下其快捷键为 option+command+v，如果你对 idea 很熟悉马上就会发现，这快捷键和 refactor 中的 extract variable 冲突了，那么我们就要更改一下 ideaVim 的开关快捷键。

> 进入 keymap -> 搜索 Vim Emulator，先取消原先的 option+command+v，再重新输入一个即可，我选择的是 ctrl+; 因为键数少，且不冲突，非常方便。

3. 尝试进入 vim 模式，可以发现久违的 vim 风格来了！

### 等等！
那原来 idea 的快捷键怎么办？还能用吗？这是一个复杂的问题。
首先，在 Settings -> Editor -> Vim Emulation标签下，可以看到 IdeaVim 已经自动检测出了与当前 keymap 冲突的快捷键，Windows 下会更多一些 (因为很多在 Windows 下与 Ctrl 相关的快捷键在 Mac 下都变成了 command)。IdeaVim 对这些冲突的快捷键提供了三种选择：
- IDE：该快捷键使用 IDE 的定义
- Vim：该快捷键使用 Vim 的定义
- Undefined：暂未定义，在按下此快捷键时会跳出提示

其次，部分 Idea 快捷键会被 IdeaVim 覆盖，例如 ctrl + left/right 会变成在 word 间跳而不是切换 Tabs 等等。有一些方法能够配置.ideavim 文件来做映射，然而我认为使用IdeaVim 就是取 Vim 之长，补 Idea 之短，这种 Idea 更擅长的工作就还是让 Idea 来做吧！毕竟我们能方便的通过 ctrl+;在 Vim 与 Idea 之间切换。