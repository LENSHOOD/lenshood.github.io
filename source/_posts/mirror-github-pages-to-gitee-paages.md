---
title: Git Pages 镜像到 Gitee Pages
date: 2020-04-06 21:21:26
tags:
- github pages
- gitee pages
categories:
- Others
---

## Git Pages 镜像到 Gitee Pages
在本站早期的文章[Build your own blog site by using GithubPages + Hexo + Travis CI](https://lenshood.github.io/2019/04/02/Build-your-own-blog-site-by-using-GithubPages-Hexo-Travis/)中描述了如何使用 Hexo 和 Travis 来在 Git Pages 服务上搭建自己的个人博客。

但使用了一段时间之后，发现 Git  Pages 的方案存在一些问题：
- 访问速度慢。不走 VPN 访问 Github 速度感人，Git Pages 也一样，各种资源、图片加载缓慢。
- 难以被 Baidu 收录。网友提供了各种办法来解决，但都挺麻烦。

所以当我了解到 Gitee 也提供了免费的静态网页服务后，决定试试看，但毕竟一边在 github 上更新，一边在 gitee 更新，既麻烦又容易出错，所以期望能够直接把 github 上的更新自动同步到 gitee 上。

经过一些实验后，我成功的实现了把 Git Pages 镜像到 Gitee Pages 的需求。

### Hexo 的 deploy 功能
Hexo 的 Deployment 配置简单而强大：
对于部署到 Git 仓库，直接配置：
```
deploy:
  type: git
  repo: <repository url> # https://bitbucket.org/JohnSmith/johnsmith.bitbucket.io
  branch: [branch]
  message: [message]
```
就能在执行`deploy`命令后自动将生成的静态网页提交至 git 仓库。

结合我们的诉求，我们期望每次执行 `deploy` 时，同时将生成的页面提交至多个仓库（github 和 gitee），可以这样配置：
```
deploy:
- type: git
  repo:
- type: heroku
  repo:
```

### Gitee 创建仓库，生成 Access Token
与 github 一样，gitee 也需要创建一个仓库，然后把静态页面文件 push 上去。

Gitee Pages 略有不同，任何仓库都可以尝试部署为 Pages，不过需要依赖 Gitee 提供 Pages 服务，如下图所示：

{% asset_img gitee_pages_service.png %}

部署后的访问地址规则如下：
- 若仓库名直接设置为用户名，则访问地址为：`https://{用户名}.gitee.io`
- 其他类型的仓库名，访问地址为：`https://gitee.io/{用户名}/{仓库名}`

仓库创建完成后，就可以尝试通过 Hexo 将静态页面部署至 Gitee Pages 了。在这之前，我们还需要创建一个 **Access Token**。

点击 `Gitee 头像 -> settings -> 左边栏 personal access token`, 创建一个 token，该 token 只用于 hexo push 代码，因此只勾选`Full control of your projects`即可。

创建完毕后注意保存待用。

### Travis
最后一步，配置 Travis 自动触发构建。

先前，我们在 github 仓库里已经完整配置了 `.travis.yml`，这一次我们也完全不用改动它，需要改动的是 Hexo 的`_config.yml`。

为了防止 Access Token 泄露，先前的 `_config.yml` 中配置的 deploy 隐藏了 Access Token，通过 travis 在执行时来填入：
```yml
deploy:
- type: git
  repo: https://git_access_token@github.com/LENSHOOD/lenshood.github.io.git
  branch: master
```
这里的`git_access_token`被记录在 Travis 里：

{% asset_img git_access_token.png %}

现在我们在 `_config.yml`里增加 gitee 的配置，完整如下：
```yml
deploy:
- type: git
  repo: https://git_access_token@github.com/LENSHOOD/lenshood.github.io.git
  branch: master
- type: git
  repo: https://gitee_access_token@gitee.com/lenshood/lenshood.git
  branch: master
```

同样的我们把 `gitee_access_token` 也添加在 Travis 里面就好了：

{% asset_img gitee_access_token.png %}

> gitee 的 access token 在使用上与 github 略有不同，github 中 repo 链接直接配置为 `https://{access_token}@{仓库地址}` 即可，但 gitee 需要配置为 `https://{用户名}:{access_token}@{仓库地址}`

### 部署 Gitee Pages
Gitee Pages 比较 low 的一点是，每次有更新 push 到仓库，必须要手动点击 `Services -> Gitee Pages` 进入后点击 `update` 才可以，自动部署需要付费版才可以。

{% asset_img gitee_update.png %}

不过不管如何，用 Gitee Pages，总算可以被 Baidu 收录了！