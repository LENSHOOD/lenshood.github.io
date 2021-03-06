---
title: 对一次发布失败的反思
date: 2020-07-10 15:21:05
tags: 
- rethink
- release
categories:
- Software Engineering
---

近期所在的项目，在进行近 4 个月的紧张交付后，与客户一起安排并进行了一次发布上线。在上线过程中出现了一个较为严重的事故，好在没有对客户造成太大的损失，本文期望针对完整的事件进行复盘与反思，以加强自己对线上环境的敬畏。

<!-- more -->

### 事发缘由

此次发布所涵盖的功能点中，有一个功能。会定时获取上游服务数据，并根据其中的相关干系人信息，给他们发送提醒邮件。正是这样一个简单且经过多次验证的功能，导致出了事故：

- 由于首次上线，第一次获取到了大量的用户数据（后续才会增量获取），导致发送了大量邮件给相关干系人（试想早上来上班发现自己邮箱里多了几千封莫名其妙的邮件的感受），结果产生了多个相关的投诉单。

在发现这些问题后，我们先紧急关闭了邮件服务（但其实该发的都发完了），之后了解了相关上下文以及问题的直接原因后，向相关受到骚扰的用户发出了致歉信。

### 问题所在

只从前文对事故的描述上看，首先以下几点是无可辩驳的失误：

1. 为什么在告知相关用户新系统上线并可能会发送邮件之前，就把邮件服务打开，并开始给真实用户发送邮件（用户未被告知，一脸懵逼）
2. 为什么会在短时间内（一天内）给用户发送超过数十封邮件的情况下，代码仍然继续执行发送逻辑（完全没有设置限流、报警等措施，当问题发现时已经不可挽回）
3. 第一次同步了大量数据，为什么没有对这些数据进行筛选，难道同步的所有数据都是业务需要的吗？（很多历史数据生命周期早已结束只做留存用，根本不应该再发邮件提醒干系人）

4. 本身业务上存在给相同用户发送数百封邮件的设计，是否合理？（完全没有从用户体验角度考虑）

### 深入探讨

从我们整体的交付流程上看，通常情况下的交付流程是这样的：

1. 由客户给出需求说明方案，我们的 BA + UX + TL 会客户一起进行需求澄清，并给出大致的工作量估计。
2. UX 根据要求给出高保真稿，与客户确认，BA 将原始需求进行拆解，划分迭代，并将最近迭代的细化需求转化为用户故事。
3. 全员一起开 IPM 会议，对迭代内的用户故事进行估点。
4. 进入迭代开发流程，包括研发 -> DC -> QA。
5. 对近期 1-2 个迭代的产出物与客户进行 showcase。
6. 继续下一个迭代

回到事故本身，发送大量邮件的问题上：

1. 在需求澄清阶段，我们完全没有对发送邮件需求中 -- 可能给相同的人发送大量邮件 -- 这种 case 做任何讨论，我们作为实现方没能发现业务需求中的漏洞，默默地接收了这样的需求。
2. 需求澄清阶段拿到的需求可能的确比较粗，忽略了一些边界条件情有可原，但 BA 在进行需求转化为用户故事的过程中，会细致的对需求进行梳理，然而也没能发现这样的问题。（实际情况是发现了并和客户进行了沟通，但没有说服客户修改业务，遂放弃）
3. 所有成员一起 IPM 时，也没能发现这个潜在的故障需求，大家都默认客户已经知道这种风险，且没有再深究（潜意识里将风险与责任甩锅给客户，而忽视了自己的专业性判断）
4. 研发全流程中，开发同学没有考虑给邮件服务增加限流、报警等功能的意识，认为实现邮件服务就是发送邮件，能收到，开发完成。QA 同学也没能覆盖到大批量发送邮件的场景。
5. showcase 时我们与客户又一次全体 blind，忽略了这个问题。

可以看到，经过分析发现，并不仅仅是由于上线准备不足，或是功能开发不完善而导致事故的发生，真实的情况是只要在交付的任何一个流程我们发现了可能的问题并对其进行限制，就一定不会出现问题，然而并没有。

所以实际上，问题的根因在于：`作为交付团队，实际上只是在机械的实现客户的需要，简单而直接的满足需求的实现，但却没有拿出自身的专业性，从根本上对需求本身进行质疑和深入思考，甲乙双方甚至略有陷入互相对立的局面。`

如果不解决根因，再进行安全学习、流程优化，这种事故仍然会一而再而三的发生。

### 解决方案

1. 修复邮件服务风险，增加相关保护逻辑
2. 与客户重新沟通业务，确保达成一致
3. 提高团队服务意识，增进客户合作。