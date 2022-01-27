---
title: Summary of 2021
date: 2022-01-01 23:37:52
tags:
- summary
categories:
- Others
---

## Farewell 2021

It seems like time elapse way faster than my memory and mind. It's already 2022!

If use some phrases to summaries my 2021, I think I'll take leisure, self promotion and anxiety. 

I spend such a good leisure time almost whole year. Due to lack of project in my department, I'm stay on the beach about 8 weeks totally. Even most of my 15 days annual vacations were took in that beach time, to lower my cost. Hence, I have a plenty of spare time to do what I love to, include coding, speak sessions, learning and traveling!

It's no big deal for my "self promotion" part. The key words maybe contains read some books, dabble some techs, open source some code, write some articles, get real "promotion"(only title but maybe not pay) and try some new job contents.

I think it's quite common to say I'm anxious about my life while 2021 it's my first year over 30-year-old. The more deep dive to software development field, I can feel more regret to go into this industry in my 25 rather than 18! If I choose CS in my collage, I'm sure I can do better than today's me. I spent 7 years at my EE major, I feel good about it, but I don't love it. After graduation I meet many excellent people, they're good at coding, have passion in software technology, and love to share. But the most important thing is they're all so young and still have a lot of years to explore and improve.

<!-- more -->



## Skills

### Golang

Golang's easy-to-learn and simple-to-use attributes let me prefer to use it to build my own projects.

Early this year I was plan to write a simple local cache as a golang learning project. However, after compared the performance of my implementation with the famous Java cache *caffeine*, I realized that there's so many tech points to care about inside a simple local cache. So I stopped develop my go cache, turned to write a ring buffer by golang. It was so much fun to try write a high performance ring buffer, I learned a lot, include lock-free programing, cache line optimization, performance test in go, etc. Now my implementation already got a dozen of stars! :)

The ring buffer implementation strengthen my confidence, so after that I began to take a further step: implement the raft algorithm. At first I translate the whole raft paper and preliminary understand the concept of raft, then I spent few hours to made a quick design of the whole architecture. After that, coding part takes me several days, then I finished the first version (thanks to the good beach time). Then the testing and correctness verification spend me a lot of time, and I still working on this part.

The latest go project of mine is a task control and monitor agent, which included in my recent consulting work. The whole design is like some sort of very lite version of *kubelet*, we call it *vmlet*, haha.

### Database

After some superficial contribution at TiDB, I realized that this kind of contribute cannot make me truly understand database technology. So this year I start to learn CMU 15-445, a series of database courses. Before that, I somehow have lower acceptance of video courses. I rather like reading books because I think video course maybe too boring to keep focus. But that courses changed my mind, the content is very logical and the lecture taught by professor Andy Pavlo is easy to understand.

With the CMU 15-445, I also try to do the course project, a simple storage engine called *bustub*. I choose rust lang on this project, and I have to say rust do have steep learning curve! This project is not finished and I haven't spend my time on that in recent few couple of months, but I believe I can keep working on it.

### Books & papers & writings

1. [Book] DDIA: it's a very good book on distributed system field. Not finished, still reading.
2. [Book] The art of multi-processors programing: As the name, it's a book of multi-processors programing. First six chapters are too academic to understand, the rest parts are better. Still reading.
3. [Book] Modern operating systems: basic course in CS, but I never systematically learned it. Still reading.
4. [Book] CSAPP: another CS basic course, some part not easy, but very useful. Still reading.
5. [Book] Innodb storage engine: Good book, but maybe outdated. Still Reading.
6. [Paper] Google Spanner
7. [Paper] Raft Algorithm
8. [Paper] LSM Tree
9. [Writing] 浅谈对开发者友好的软件设计



## Financial

2021 I've got -5% of total earned rate. That's not good.

The main reason of the loss is that I put about 10+% of my money into Chinese internet industry, and that part fell about 20%. The government supervision lead to internet industry going down, but I still believe it do have chance to recover, after all the profitability and business are all healthy.

I'm not professional, but I know in the long term, the return is definitely considerable. So I'm going to keep high position of fund investment in 2022, be patient, be confidence.

There's another windfall, I got a little options from company IPO, better than nothing.



## Others

I'm gonna left this part to say something about epidemic.

In the end of 2021, Xi'an, the city I'm living, occurred covid-19 epidemic. The reaction of city gov is a piece of shit, they acted like  unprepared, unprofessional and stupid. They leave people to starve, to died of untimely treatment, to force abortion because of absent negative nucleic acid test result. Keeping their position seems to be the only thing they care about.

I'm so disappointment with that, especially the epidemic is covid-19 not covid-21. They got two years to be prepared, but they failed all the people live in Xi'an. 

Every one would hope their home town become better. But I have lived here for decades, and I have seen this again and again: people try really hard to do there best, but the bureaucratic ruined it. Today I'm really need to consider is there any good reason to still live here?

Shame on you, Xi'an gov.
