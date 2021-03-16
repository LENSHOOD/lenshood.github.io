---
title: Richardson 成熟度模型（通往 REST 的荣光之路）
date: 2021-03-16 21:07:34
tags:
- RESTful
- maturity model
categories:
- Software Engineering
---

> 原文请见：https://martinfowler.com/articles/richardsonMaturityModel.html

Recently I've been reading drafts of [Rest In Practice](https://www.amazon.com/gp/product/0596805829?ie=UTF8&tag=martinfowlerc-20&linkCode=as2&camp=1789&creative=9325&creativeASIN=0596805829): a book that a couple of my colleagues have been working on. Their aim is to explain how to use Restful web services to handle many of the integration problems that enterprises face. At the heart of the book is the notion that the web is an existence proof of a massively scalable distributed system that works really well, and we can take ideas from that to build integrated systems more easily.

最近我正在阅读一本我的几个同事一直在写的书的草稿，书名叫 [Rest In Practice](https://www.amazon.com/gp/product/0596805829?ie=UTF8&tag=martinfowlerc-20&linkCode=as2&camp=1789&creative=9325&creativeASIN=0596805829)。他们写这本书的目的，是为了解释怎么样用 Restful web 服务来处理企业中经常面对的许多集成问题。这本书的核心概念是，web 是一种对大规模可扩展分布式系统能够工作良好的一个真实证明，并且，我们能够从中总结出一些如何更简单的构建集成系统的想法。

![img](https://martinfowler.com/articles/images/richardsonMaturityModel/overview.png)

*图 1：走向 REST*



To help explain the specific properties of a web-style system, the authors use a model of restful maturity that was developed by [Leonard Richardson](http://www.crummy.com/) and [explained](http://www.crummy.com/writing/speaking/2008-QCon/act3.html) at a QCon talk. The model is nice way to think about using these techniques, so I thought I'd take a stab of my own explanation of it. (The protocol examples here are only illustrative, I didn't feel it was worthwhile to code and test them up, so there may be problems in the detail.)

作者们使用了一种由 [Leonard Richardson](http://www.crummy.com/) 在 [QCon 大会上介绍 ](http://www.crummy.com/writing/speaking/2008-QCon/act3.html)的“restful 成熟度模型” ，来帮助解释 web-style 系统的特定属性。该模型是理解如何使用这类技术的一个好办法，因此我想尝试用自己的方式来解释它。（这里用到的协议示例仅仅用于展示，我并不觉得它值得用代码和测试来表述，因此在细节上可能存在些许问题。）

### Level 0

The starting point for the model is using HTTP as a transport system for remote interactions, but without using any of the mechanisms of the web. Essentially what you are doing here is using HTTP as a tunneling mechanism for your own remote interaction mechanism, usually based on [Remote Procedure Invocation](http://www.eaipatterns.com/EncapsulatedSynchronousIntegration.html).

成熟度模型的起点，是将 HTTP 用作远程交互的手段，而并不引入任何 web 机制。本质上，在这一层你只是把 HTTP 当做一个管道用在自己的远程交互机制中（通常基于 [Remote Procedure Invocation](http://www.eaipatterns.com/EncapsulatedSynchronousIntegration.html)）。

![img](https://martinfowler.com/articles/images/richardsonMaturityModel/level0.png)

*图 2：Level 0 的一个例子*

Let's assume I want to book an appointment with my doctor. My appointment software first needs to know what open slots my doctor has on a given date, so it makes a request of the hospital appointment system to obtain that information. In a level 0 scenario, the hospital will expose a service endpoint at some URI. I then post to that endpoint a document containing the details of my request.

假设我想预约我的医生。那么我的预约软件首先需要了解该医生在我指定的日期内还有没有空闲的时间段，所以它向医院的预约系统发了一个请求来获取这一信息。在 level 0 的场景下，医院会在某个 URL 上暴露一个服务 endpoint。首先我会向该端点发起一个 post 请求，请求体是一个包含了请求详情的文档。

```http
POST /appointmentService HTTP/1.1
[various other headers]

<openSlotRequest date = "2010-01-04" doctor = "mjones"/>
```

The server then will return a document giving me this information

接着服务器会返回一份文档来向我说明我请求的信息

```http
HTTP/1.1 200 OK
[various headers]

<openSlotList>
  <slot start = "1400" end = "1450">
    <doctor id = "mjones"/>
  </slot>
  <slot start = "1600" end = "1650">
    <doctor id = "mjones"/>
  </slot>
</openSlotList>
```

I'm using XML here for the example, but the content can actually be anything: JSON, YAML, key-value pairs, or any custom format.

这里我用 XML 来距离，但实际的内容形式可以多种多样：JSON, YAML, 键值对，或任意自定义格式。

My next step is to book an appointment, which I can again do by posting a document to the endpoint.

我的下一步动作就是做预约，我仍然可以通过给前面提到的endpoint 发送一个 post 请求来实现。

```http
POST /appointmentService HTTP/1.1
[various other headers]

<appointmentRequest>
  <slot doctor = "mjones" start = "1400" end = "1450"/>
  <patient id = "jsmith"/>
</appointmentRequest>
```

If all is well I get a response saying my appointment is booked.

如果一切顺利的话，我会得到服务器的返回，告诉我预约成功了。

```http
HTTP/1.1 200 OK
[various headers]

<appointment>
  <slot doctor = "mjones" start = "1400" end = "1450"/>
  <patient id = "jsmith"/>
</appointment>
```

If there is a problem, say someone else got in before me, then I'll get some kind of error message in the reply body.

但假如有个人先于我完成了这一时间段的预定，我就会得到一个包含相应的错误信息的返回。

```http
HTTP/1.1 200 OK
[various headers]

<appointmentRequestFailure>
  <slot doctor = "mjones" start = "1400" end = "1450"/>
  <patient id = "jsmith"/>
  <reason>Slot not available</reason>
</appointmentRequestFailure>
```

So far this is a straightforward RPC style system. It's simple as it's just slinging plain old XML (POX) back and forth. If you use SOAP or XML-RPC it's basically the same mechanism, the only difference is that you wrap the XML messages in some kind of envelope.

到现在为止，这纯粹是一个 RPC 式的系统。它很简单，因为他只是来回的投递 Plain Old XML（POX）。假如你使用过 SOAP 或 XML-RPC，那么这和它们是一个类似的机制，惟一的区别在于你将 XML 信息封装成了某种不同的信封（见[SOAP Envelope](https://www.w3.org/TR/2000/NOTE-SOAP-20000508/#_Toc478383494)）。

### Level 1 - Resources

The first step towards the Glory of Rest in the RMM is to introduce resources. So now rather than making all our requests to a singular service endpoint, we now start talking to individual resources.

在 RMM（Richardson Maturity Model）通过 REST 荣光之路的第一步，是引入资源（resources）。因此相较于之前我们将所有的请求都发往一个相同的 endpoint，现在我们开始讨论单独的资源。

![img](https://martinfowler.com/articles/images/richardsonMaturityModel/level1.png)

*图 3：Level 1 添加资源*

So with our initial query, we might have a resource for given doctor.

所以对于我们的初始查询，我们可能会指定一个 “医生资源”。

```http
POST /doctors/mjones HTTP/1.1
[various other headers]

<openSlotRequest date = "2010-01-04"/>
```

The reply carries the same basic information, but each slot is now a resource that can be addressed individually.

服务器的返回中包含了类似的基础信息，但这一次每一个时间段都变成了一个能够单独寻址的资源。

```http
HTTP/1.1 200 OK
[various headers]

<openSlotList>
  <slot id = "1234" doctor = "mjones" start = "1400" end = "1450"/>
  <slot id = "5678" doctor = "mjones" start = "1600" end = "1650"/>
</openSlotList>
```

With specific resources booking an appointment means posting to a particular slot.

对于特定的资源，预约操作就是向特定的时间段发送 post 请求。

```http
POST /slots/1234 HTTP/1.1
[various other headers]

<appointmentRequest>
  <patient id = "jsmith"/>
</appointmentRequest>
```

If all goes well I get a similar reply to before.

如果一切顺利，我们会得到一个与前文类似的返回。

```http
HTTP/1.1 200 OK
[various headers]

<appointment>
  <slot id = "1234" doctor = "mjones" start = "1400" end = "1450"/>
  <patient id = "jsmith"/>
</appointment>
```

The difference now is that if anyone needs to do anything about the appointment, like book some tests, they first get hold of the appointment resource, which might have a URI like `http://royalhope.nhs.uk/slots/1234/appointment`, and post to that resource.

现在的区别就是假如某人需要对该预约进行一些操作，比如测试预定，他们会先通过类似这种 URL：`http://royalhope.nhs.uk/slots/1234/appointment` 来持有该预约资源，并向该资源发送 post 请求。

To an object guy like me this is like the notion of object identity. Rather than calling some function in the ether and passing arguments, we call a method on one particular object providing arguments for the other information.

对于一个像我这样的对象小子（object guy），这就像是对象 id 的概念一样。我们不是调用网络上的函数并传送参数，而是在一个特定的对象上调用一个方法，提供参数来获得其他信息。



### Level 2 - HTTP Verbs

I've used HTTP POST verbs for all my interactions here in level 0 and 1, but some people use GETs instead or in addition. At these levels it doesn't make much difference, they are both being used as tunneling mechanisms allowing you to tunnel your interactions through HTTP. Level 2 moves away from this, using the HTTP verbs as closely as possible to how they are used in HTTP itself.

在 level 0 和 1 的交互过程当中，我已经用到了 HTTP POST 动词，但是有些人会代替或额外的使用 GET。 在这些层里用哪一种动词无关紧要，因为他们都是被当做隧道机制通过 HTTP 来承载你的交互动作的。Level 2 更进一步，将 HTTP 动词以尽可能接近于 HTTP 本身的用法来使用。

![img](https://martinfowler.com/articles/images/richardsonMaturityModel/level2.png)

*图 4：Level 2 添加 HTTP 动词*

For our the list of slots, this means we want to use GET.

对于获取可用时间段列表，这意味着我们应当使用 GET。

```http
GET /doctors/mjones/slots?date=20100104&status=open HTTP/1.1
Host: royalhope.nhs.uk
```

The reply is the same as it would have been with the POST

对方返回与先前 post 请求一样的内容

```http
HTTP/1.1 200 OK
[various headers]

<openSlotList>
  <slot id = "1234" doctor = "mjones" start = "1400" end = "1450"/>
  <slot id = "5678" doctor = "mjones" start = "1600" end = "1650"/>
</openSlotList>
```

At Level 2, the use of GET for a request like this is crucial. HTTP defines GET as a safe operation, that is it doesn't make any significant changes to the state of anything. This allows us to invoke GETs safely any number of times in any order and get the same results each time. An important consequence of this is that it allows any participant in the routing of requests to use caching, which is a key element in making the web perform as well as it does. HTTP includes various measures to support caching, which can be used by all participants in the communication. By following the rules of HTTP we're able to take advantage of that capability.

在 Level 2，对这种请求使用 GET 会非常关键。HTTP 将 GET 定义为安全操作，即它不会对任何事物的状态产生任何明显的改变。这允许我们安全的调用任意多次 GET，并且每一次都能够得到相同的结果。这样做的一个重要的结果，就是它允许请求路由中的任意参与者都能使用缓存，而缓存是让 web 发挥其应用性能的一个关键因素。HTTP 包括各种用于支持使用缓存的措施，在通信中的所有参与者都可以使用缓存。通过遵循 HTTP 的规则，我们能够从它所提供的能力中获益。

To book an appointment we need an HTTP verb that does change state, a POST or a PUT. I'll use the same POST that I did earlier.

为了进行预约，我们需要一个能修改状态的 HTTP 动词，即 POST 或 PUT。我将会使用与前文类似的 POST。

```http
POST /slots/1234 HTTP/1.1
[various other headers]

<appointmentRequest>
  <patient id = "jsmith"/>
</appointmentRequest>
```

The trade-offs between using POST and PUT here are more than I want to go into here, maybe I'll do a separate article on them some day. But I do want to point out that some people incorrectly make a correspondence between POST/PUT and create/update. The choice between them is rather different to that.

POST 和 PUT 之间的权衡已经超出了本文的范畴，也许某日我会单独写一篇文章来讨论它们。但我的确期望指出，部分人错误的将 POST/PUT 映射为 create/update。这与如何选择它们毫无关系。

Even if I use the same post as level 1, there's another significant difference in how the remote service responds. If all goes well, the service replies with a response code of 201 to indicate that there's a new resource in the world.

即使我像在 level 1 一样使用 post，在远程服务的响应上，它们仍旧存在显著的区别。如果一切顺利，服务器会返回一个 201 响应码来指明这世界上多了一个新的资源。

```http
HTTP/1.1 201 Created
Location: slots/1234/appointment
[various headers]

<appointment>
  <slot id = "1234" doctor = "mjones" start = "1400" end = "1450"/>
  <patient id = "jsmith"/>
</appointment>
```

The 201 response includes a location attribute with a URI that the client can use to GET the current state of that resource in the future. The response here also includes a representation of that resource to save the client an extra call right now.

这个 201 响应在 URI 中包含了一个位置属性，通过这个属性，将来客户端可以通过 GET 来获取该资源当前的状态。该响应还包含了该资源本身，省得客户端再单独发送一个请求来获取。

There is another difference if something goes wrong, such as someone else booking the session.

在出错时（比如另一个人同时也在预约）这一层的表现也会与先前不同。

```http
HTTP/1.1 409 Conflict
[various headers]

<openSlotList>
  <slot id = "5678" doctor = "mjones" start = "1600" end = "1650"/>
</openSlotList>
```

The important part of this response is the use of an HTTP response code to indicate something has gone wrong. In this case a 409 seems a good choice to indicate that someone else has already updated the resource in an incompatible way. Rather than using a return code of 200 but including an error response, at level 2 we explicitly use some kind of error response like this. It's up to the protocol designer to decide what codes to use, but there should be a non-2xx response if an error crops up. Level 2 introduces using HTTP verbs and HTTP response codes.

这一响应中重要的部分就是通过使用 HTTP 响应码来指明发生了错误。在这个例子中，用 409 来指示有人已经通过一个不兼容的方式更新了该资源似乎是个不错的选择。相比于返回一个响应码 200 但包含了错误信息的方式，level 2 中我们显式的使用这样的错误响应。到底使用什么响应码，是由协议设计者来决定的，但只有当错误突然出现时，我们都应该返回一个 非 2xx 的响应。Level 2 引入了 HTTP 动词和 HTTP 响应码。

There is an inconsistency creeping in here. REST advocates talk about using all the HTTP verbs. They also justify their approach by saying that REST is attempting to learn from the practical success of the web. But the world-wide web doesn't use PUT or DELETE much in practice. There are sensible reasons for using PUT and DELETE more, but the existence proof of the web isn't one of them.

这里有个不一致的地方。REST 的鼓吹者会说要使用所有的 HTTP 动词。他们还辩称 REST 正尝试从 web 的实际成功中学习。但实际上万维网并不经常使用到 PUT 或 DELETE。PUT 和 DELETE 有其合理的使用理由，但目前已存在的 web 并不能证明这一点。

The key elements that are supported by the existence of the web are the strong separation between safe (eg GET) and non-safe operations, together with using status codes to help communicate the kinds of errors you run into.

Web 所能支持的关键元素是安全操作（如 GET）和非安全操作之间的强分离，以及使用状态码来帮助传达遇到的各种错误。



## Level 3 - Hypermedia Controls

The final level introduces something that you often hear referred to under the ugly acronym of HATEOAS (Hypertext As The Engine Of Application State). It addresses the question of how to get from a list open slots to knowing what to do to book an appointment.

最后一层介绍了一些你经常会听到的东西，其丑陋的缩写是 HATEOAS（作为应用状态引擎的超文本）。它解决了一个问题，即如何从一个时间段列表中获悉如何去预定时间段。

![img](https://martinfowler.com/articles/images/richardsonMaturityModel/level3.png)

*图5：Level 3 添加超媒体控件*

We begin with the same initial GET that we sent in level 2

我们以一个与 level 2 相同的初始 GET 请求为开始

```http
GET /doctors/mjones/slots?date=20100104&status=open HTTP/1.1
Host: royalhope.nhs.uk
```

But the response has a new element

但这次返回体中多了一个元素

```http
HTTP/1.1 200 OK
[various headers]

<openSlotList>
  <slot id = "1234" doctor = "mjones" start = "1400" end = "1450">
     <link rel = "/linkrels/slot/book" 
           uri = "/slots/1234"/>
  </slot>
  <slot id = "5678" doctor = "mjones" start = "1600" end = "1650">
     <link rel = "/linkrels/slot/book" 
           uri = "/slots/5678"/>
  </slot>
</openSlotList>
```

Each slot now has a link element which contains a URI to tell us how to book an appointment.

现在每个时间段都多了一个包含 URI 的链接元素，它能告诉我们如何发起一次预定。

The point of hypermedia controls is that they tell us what we can do next, and the URI of the resource we need to manipulate to do it. Rather than us having to know where to post our appointment request, the hypermedia controls in the response tell us how to do it.

超媒体控件的意义在于它能告知我们下一步可以干什么，并且包含我们需要操作的资源的 URI。我们不需要提前了解去哪里发送我们的预约请求，而是在响应中的超媒体控件直接告诉我们如何去做。

The POST would again copy that of level 2

POST 请求也还和 level 2 中的一样

```http
POST /slots/1234 HTTP/1.1
[various other headers]

<appointmentRequest>
  <patient id = "jsmith"/>
</appointmentRequest>
```

And the reply contains a number of hypermedia controls for different things to do next.

同样的，返回中包含了一系列接下来可以干的不同事情的超媒体控件。

```http
HTTP/1.1 201 Created
Location: http://royalhope.nhs.uk/slots/1234/appointment
[various headers]

<appointment>
  <slot id = "1234" doctor = "mjones" start = "1400" end = "1450"/>
  <patient id = "jsmith"/>
  <link rel = "/linkrels/appointment/cancel"
        uri = "/slots/1234/appointment"/>
  <link rel = "/linkrels/appointment/addTest"
        uri = "/slots/1234/appointment/tests"/>
  <link rel = "self"
        uri = "/slots/1234/appointment"/>
  <link rel = "/linkrels/appointment/changeTime"
        uri = "/doctors/mjones/slots?date=20100104&status=open"/>
  <link rel = "/linkrels/appointment/updateContactInfo"
        uri = "/patients/jsmith/contactInfo"/>
  <link rel = "/linkrels/help"
        uri = "/help/appointment"/>
</appointment>
```

One obvious benefit of hypermedia controls is that it allows the server to change its URI scheme without breaking clients. As long as clients look up the "addTest" link URI then the server team can juggle all URIs other than the initial entry points.

超媒体控件的一个明显的好处就是允许服务器在不影响客户端的前提下改变其 URI 设计。只要客户端找的到 “addTest” 的链接，那么服务器团队就能改变除了其初始查询入口外的一切 URI。

A further benefit is that it helps client developers explore the protocol. The links give client developers a hint as to what may be possible next. It doesn't give all the information: both the "self" and "cancel" controls point to the same URI - they need to figure out that one is a GET and the other a DELETE. But at least it gives them a starting point as to what to think about for more information and to look for a similar URI in the protocol documentation.

另一个好处是这能帮助客户端来探索整个协议。这些链接能够给客户端开发者一个提示，指出接下来可能会发生什么。但它也并不会给出所有的信息：虽然 “self” 和 “cancel” 使用了相同的 URI，但客户端开发者需要事先知道其中一个是 GET 请求，另一个是 DELETE 请求。但至少它为客户端开发者提供了一个起点，让他们思考如何获取更多信息，以及在协议文档中寻找类似的URI。

Similarly it allows the server team to advertise new capabilities by putting new links in the responses. If the client developers are keeping an eye out for unknown links these links can be a trigger for further exploration.

同样的它允许服务器团队通过放置新的链接在返回中来宣布新功能。如果客户端开发者持续的关注未知的链接，那么这可能能够触发进一步的探索。

There's no absolute standard as to how to represent hypermedia controls. What I've done here is to use the current recommendations of the REST in Practice team, which is to follow ATOM ([RFC 4287](https://tools.ietf.org/html/rfc4287)) I use a `<link>` element with a `uri` attribute for the target URI and a `rel` attribute for to describe the kind of relationship. A well known relationship (such as `self` for a reference to the element itself) is bare, any specific to that server is a fully qualified URI. ATOM states that the definition for well-known linkrels is the [Registry of Link Relations ](http://www.iana.org/assignments/link-relations.html). As I write these are confined to what's done by ATOM, which is generally seen as a leader in level 3 restfulness.

对于超媒体控件，并没有一个标准来规定以何种形式展示。在这里我能做的是引用 “REST in Practice” 团队目前给出的建议，即遵循 ATOM ([RFC 4287](https://tools.ietf.org/html/rfc4287)) ，用一个 `<link>` 元素结合一个目标 URI 的 `uri` 属性和一个描述其关系的 `rel` 属性来表示。一些著名的关联（如用 `self` 来关联该元素自身）并不存在，任何特定于该服务器的关系都是全限定的 URI。ATOM 对众所周知的 linkrel 的声明是  [Registry of Link Relations ](http://www.iana.org/assignments/link-relations.html)。正如我所写的，这些都局限于 ATOM 所做的事情，ATOM 通常被视为 level 3 层的领导者。



## The Meaning of the Levels

I should stress that the RMM, while a good way to think about what the elements of REST, is not a definition of levels of REST itself. Roy Fielding has made it clear that [level 3 RMM is a pre-condition of REST](http://roy.gbiv.com/untangled/2008/rest-apis-must-be-hypertext-driven). Like many terms in software, REST gets lots of definitions, but since Roy Fielding coined the term, his definition should carry more weight than most.

What I find useful about this RMM is that it provides a good step by step way to understand the basic ideas behind restful thinking. As such I see it as tool to help us learn about the concepts and not something that should be used in some kind of assessment mechanism. I don't think we have enough examples yet to be really sure that the restful approach is the right way to integrate systems, I do think it's a very attractive approach and the one that I would recommend in most situations.

Talking about this with Ian Robinson, he stressed that something he found attractive about this model when Leonard Richardson first presented it was its relationship to common design techniques.

- Level 1 tackles the question of handling complexity by using divide and conquer, breaking a large service endpoint down into multiple resources.
- Level 2 introduces a standard set of verbs so that we handle similar situations in the same way, removing unnecessary variation.
- Level 3 introduces discoverability, providing a way of making a protocol more self-documenting.

The result is a model that helps us think about the kind of HTTP service we want to provide and frame the expectations of people looking to interact with it.