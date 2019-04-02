---
title: Build_your_own_blog_site_by_using_GithubPages_+_Hexo_+_Travis
date: 2019-04-02 21:52:28
tags: Github Pages, Hexo, Travis
---

### Build your own blog site by using Gitpages + Hexo + Travis
---
#### Overview
Thanks for the github static site service - github pages - it's really easy to build your own blog for free.

At the rest of this article, you could see how to build a personal blog site step by step, using GithubPages + Hexo + Travis.

#### GithubPages
GithubPages is a free static site hosting service for personal/organization/project. In this article we just talk about personal pages.

Each account of github can host only one site on GithubPages, and the address is {username}.github.io.

Let's get started:
1. Create a repository named **{username}.github.io**, please make sure the repo name is exactly match this pattern, or it may not work (mine is lenshood.github.io).
![](create_repo.png)

2. Clone the repo to local.
``` shell
git clone https://github.com/username/username.github.io
```
3. Create a simple index.html to display "Hello World".
``` shell
echo "Hello World" > index.html
```

4. Push it to repo just like ordinary project.
``` shell
git add .
git commit -m 'Demo Site'
git push -u origin master
```

5. Access to {username}.github.io to see what happend.
![](hello_world.png)

> For more details please see: https://pages.github.com/

#### Hexo
