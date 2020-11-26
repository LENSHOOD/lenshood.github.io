---
title: Git Hooks
date: 2019-05-30 23:11:58
tags:
- git
- hooks
categories:
- Git
---

规范 Git 系列：
[第一篇：如何保持公共开发分支的整洁](https://lenshood.github.io/2019/04/08/keep-git-branch-clean/)
[第二篇：Good Git Commit](https://lenshood.github.io/2019/04/21/good-git-commit/)
[第三篇：Good Commit Message](https://lenshood.github.io/2019/04/21/conventional-commit-message/)
[第四篇：Git Hooks](https://lenshood.github.io/2019/05/30/git-hooks/)

<!-- more -->

---

本系列前几篇文章讲了许多理论，如何保持分支整洁，如何撰写合理的 commit message 等等。本文不再多谈理论，而是将引入一项 git built-in 的强大功能 - **hooks**。

与我们所知的其他软件或系统的 hooks 一致，git hooks 也是一种类似钩子函数的脚本，可以在执行某些 git 命令的前后自动触发。

上面提到了 1.脚本 2.自动，有了这两点 git 的功能将被极大的延伸，因为有无数的用户可以根据自己的喜好在 hooks 定义的范围内进行自己的创作，来提升工作效率。正因此，github 上有非常多与 git hooks 相关的优秀项目。

## What exactly the GIT-HOOKS is ?
正如前文所述，hooks 是 git 内置的功能，能够允许用户定义脚本并在重要操作发生时被触发。hooks 分为两部分，client-side 和 server-side。client-side 主要在 git 命令操作时被触发，例如 commit、merge 等。server-side 主要在 git 服务端起作用，例如在收到 push commit 的时候被触发(本文仅涉及 client-side)。

git hooks 作为 git 的内置功能，无需额外安装。具体的脚本存放在 .git/hooks/ 下，git 默认在该目录下放置了一些示例脚本，都以 .sample 作为后缀名，这些示例不会被运行，git hooks 只会尝试运行该目录下没有后缀名的文件。此外，在运行 `git init` 命令时，hooks 会自动被创建在 .git 下。

## What kinds of HOOKS dose it provided ?
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
	
## Any examples ?
1. 提交前静态检查

git-hooks 最常用的场景应属提交前的代码静态检查了，由于 git-hooks 本身类似于给 git 命令增加了生命周期钩子，同时支持执行脚本，因此我们能够在 git-hooks 里面触发各式各样的外部工具。

以下以 checkstyle 为例，结合 gradle 来展示如何在执行 git commit 之前自动进行 checkstyle。

- 创建一个名为 git-hooks-demo 的 gradle 项目，执行 `git init` 初始化为 git 项目
- 在 gradle 中引入 checkstyle 插件
``` groovy
    apply plugin: 'checkstyle'
    
    checkstyle {
        toolVersion = '8.21'
    }
    
    tasks.withType(Checkstyle).each { checkstyleTask ->
        checkstyleTask.doLast {
            reports.all { report ->
                def outputFile = report.destination
                if (outputFile.exists() && outputFile.text.contains("<error ")) {
                    throw new GradleException("There were checkstyle warnings! For more info check $outputFile")
                }
            }
        }
    }
```
    以上定义了 checkstyle 的 gradle 插件，并对其结果进行了处理， 一旦发现 error 就抛异常中断流程。
- 在代码目录下创建 git-hooks 目录，用于存放 hooks 文件。同时，在 build.gradle 中增加一个 task 用于关联 git-hooks
``` groovy
	task installGitHooks() {
    	"git config core.hooksPath ./git-hooks".execute()
	}
```
	> 为什么要这么做呢？ 根据上文，默认情况下 hooks 文件是存放于 .git/hooks 下的，因此存在一个严重的问题，他不会随代码一同提交至远程仓库，因此我们采用改变 hooks 文件目录的形式用于提交。
- 在 git-hooks 目录下创建新文件：pre-commit
``` shell
    #!/bin/sh
    set -x

    ./gradlew checkstyleMain

    RESULT=$?
    exit $RESULT
```
    pre-commit(注意没有任何后缀名)的内容即执行 `./gradlew checkstyleMain` 之后exit，任何返回不为零的 exit 将会打断提交的流程。

试验一下，对当前代码进行提交，可得到如下结果：

``` shell
+ ./gradlew checkstyleMain

... omit checkstyle error and warning ...

FAILURE: Build failed with an exception.

* Where:
Build file '.../git-hooks-demo/build.gradle' line: 18

* What went wrong:
Execution failed for task ':checkstyleMain'.
> There were checkstyle warnings! For more info check .../git-hooks-demo/build/reports/checkstyle/main.xml

* Try:
Run with --stacktrace option to get the stack trace. Run with --info or --debug option to get more log output. Run with --scan to get full insights.

* Get more help at https://help.gradle.org

BUILD FAILED in 1s
2 actionable tasks: 2 executed
+ RESULT=1
+ exit 1

```

可见 checkstyle 对当前的代码进行检查后发现了错误，并由 pre-commit 终止了此次提交。

2. Conventional Commit 检查

前文[Good Commit Message](https://lenshood.github.io/2019/04/21/conventional-commit-message/) 中提到了一种 commit message 的编写规范。

规范固然好，然而如果能在每次提交之前，对已经写好的 message 进行检查，并对不符合规范的地方进行提醒，则能够降低错误提交的概率，对新人也更为友好。

我们拥有强大的 gradle，可以直接执行 groovy 脚本，因此我们可以自己编写一个简单的 gradle task，结合 git-hooks 中的 prepare-commit-msg，即可对 commit message 进行检查了。

- gradle task:
``` groovy
task checkCommitMsgByConventionalCommit() {
    String[] types = ["feat", "fix", "docs", "style", "refactor",
            "test", "chore", "build", "ci", "perf"]

    // if no params then directly return
    if (!project.hasProperty("commitMsg")) {
        return
    }

    // check the three section: header, detail, footer
    String[] msgArray = ((String)commitMsg).split('\n\n')
    if (msgArray.length > 3) {
        throw new GradleException("[COMMIT MESSAGE CHECKER] \n" +
                "Too many message sections! (Header, Detail, Footer allowed)" +
                "\nAbout conventional commit, see: https://www.conventionalcommits.org")
    }

    // check header format
    String header = msgArray[0]
    String[] headers = header.split(": ")
    if (headers.length != 2) {
        throw new GradleException("[COMMIT MESSAGE CHECKER] \n" +
                "Wrong header format, which should be: {type}({scope}): {description}" +
                "\nAbout conventional commit, see: https://www.conventionalcommits.org")
    }

    // check type
    for (int i=0; i<types.length; i++) {
        if (headers[0].startsWith(types[i])) {
            break
        }

        if (i == types.length - 1) {
            throw new GradleException("[COMMIT MESSAGE CHECKER] \n"
                    + "Wrong type detected! "
                    + "Allowed: feat, fix, docs, style, refactor, test, chore, build, ci, perf."
                    + "\nAbout conventional commit, see: https://www.conventionalcommits.org")
        }
    }

    // check header length
    if (header.length() > 72) {
        throw new GradleException("[COMMIT MESSAGE CHECKER] \n" 
                + "Too long header! Header length should be shorter than 72 characters."
                + "\nAbout conventional commit, see: https://www.conventionalcommits.org")
    }
}

```

- prepare-commit-msg
``` shell
	#!/bin/bash
	
	COMMIT_MSG_FILE=$1
	COMMIT_SOURCE=$2
	SHA1=$3
	
	MSG=$(cat $COMMIT_MSG_FILE)
	./gradlew checkCommitMsgByConventionalCommit -PcommitMsg="$MSG"
	
	RESULT=$?
	exit 1
```

因此，当执行 `git commit -m 'xxx'` 时，如果出现检测不通过的情况，则提交会失败。
示例显示：

``` shell
FAILURE: Build failed with an exception.

* Where:
Build file '/Users/alexanderzhang/Documents/personal/code/java_project/git-hooks-demo/build.gradle' line: 45

* What went wrong:
A problem occurred evaluating root project 'git-hooks'.
> [COMMIT MESSAGE CHECKER] 
  Wrong header format, which should be: {type}({scope}): {description}
  About conventional commit, see: https://www.conventionalcommits.org

* Try:
Run with --stacktrace option to get the stack trace. Run with --info or --debug option to get more log output. Run with --scan to get full insights.

* Get more help at https://help.gradle.org

BUILD FAILED in 0s
```

## Reference
https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks
https://githooks.com/

---

规范 Git 系列：
[第一篇：如何保持公共开发分支的整洁](https://lenshood.github.io/2019/04/08/keep-git-branch-clean/)
[第二篇：Good Git Commit](https://lenshood.github.io/2019/04/21/good-git-commit/)
[第三篇：Good Commit Message](https://lenshood.github.io/2019/04/21/conventional-commit-message/)
[第四篇：Git Hooks](https://lenshood.github.io/2019/05/30/git-hooks/)

