---
title: Rust Module
date: 2020-03-17 00:11:16
tags: 
- rust module
categories:
- Rust
---

## Rust Modules
rust 的 module system，类似 java 的 package，可以用于将代码分别放置在适合他们的单元内。同时 rust 还允许用户控制 module 之间的可见性（public/private）。

在 module 内，rust 允许用户放置 function，struct，trait，struct implement 以及 child module。

rust 对 module 的要求是任何 module 都应处在一颗以 root module 为根的 module tree 中。期望访问某个 module  时，能够找到一条通路从root一路向下直到该 module。

<!-- more -->

### module 的定义

module 通常可以有以下两种形式来定义：
1. `mod {模块代码}`
    - 声明与实现放在一起
2. `mod 模块名` + `独立的模块代码文件`
    - 在文件 A 中声明，在文件 B 中实现

#### 声明与实现放在一起

`mod {模块代码}` 这种方式是比较简单的一种定义方法，举例如下：
```rust
// src/lib.rs
mod demo {
    fn say_hello() {
        print!("hello");
    }
}
```
如此，一个简单的包含单个函数的 module 就定义好了。

#### 在文件 A 中声明，在文件 B 中实现

假如整个系统都使用第一种定义方式，即所有 module 都定义在同一个文件里，那么这个文件可能会无比巨大，失去任何可读性。

所以，对于略复杂的系统而言，模块的定义更多会使用 `mod 模块名` + `独立的模块代码文件` 这种形式。

**对前文例子进行简单的扩展：** 我们期望把 `demo` 定义在一个独立的文件 demo.rs 中。

```rust
---------------------------
// src/demo.rs
---------------------------
fn say_hello() {
    print!("hello");
}

---------------------------
// src/lib.rs
---------------------------
mod demo;
```

果然，这一次 mod 的声明与其包含的函数 `say_hello()` 分别被放置在了 `lib.rs` 与 `demo.rs` 中，其中：
- `lib.rs` 中仅包含声明语句 `mod demo`
- 声明的模块通过相同的文件名 `demo.rs` 找到了对应的实现。

再进一步，可能有人会问：*我的项目很复杂，需要包含多层级目录的形式，怎么办？*

**再次对上述例子进行扩展：** 除 root 以外，我们的系统中包含三个 module，分别为 demo1，demo2，demo3，且其层级关系为：
```shell
.
├── lib.rs
├── demo1.rs
└── sub
    ├── mod.rs (module 声明文件)
    ├── demo2.rs
    └── demo3.rs
```

这种情况下，module 的定义如下：
```rust
---------------------------
// src/demo2.rs
---------------------------
fn say_my_name() {
    print!("I'm demo2");
}

---------------------------
// src/demo3.rs
---------------------------
fn say_my_name() {
    print!("I'm demo3");
}

---------------------------
// src/mod.rs
---------------------------
mod demo2;
mod demo3;

---------------------------
// src/demo1.rs
---------------------------
fn say_my_name() {
    print!("I'm demo1");
}

---------------------------
// src/lib.rs
---------------------------
mod demo1;
mod sub;
```

显然，与单层级的 module 分离定义方式相比，多层级略有差异，其中最重要的差异是：在 sub 子目录下多了一个 `mod.rs` 文件。正式该文件对其目录下的两个 module `demo2` `demo3` 进行了定义。

此外，在 root module 中，`mod sub` 语句直接将 sub 目录定义成了一个 module，正是这种 `目录名` + `mod.rs` 的定义方式支撑了多目录的需求。

### module 的访问
至此，我们已经成功的定义了 module，接下来看一看怎么样访问 module 以及其包含的内容。

需要访问 module，首先要在代码中引用 module 或其内容的路径，有两种方式声明引用路径：
1. 绝对路径：以 crate* 名或者直接写作 `crate`
2. 相对路径: 以当前 module 开始，结合 `self`, `super` 或其他标识。

`* crate 即当前工程包，类似 java 的 jar`

不论是绝对路径还是相对路径，层与层之间，都以`::`分隔。

仍旧采用前文的例子，我们期望在`lib.rs`中分别访问 `demo1`, `demo2`, `demo3` 中的`say_my_name()`, 则代码如下：

```rust
mod demo1;
mod sub;

use crate::sub::demo2;
use sub::demo3;

fn main() {
    print!(demo1::say_hello());
    
    print!(demo2::say_hello());

    print!(demo3::say_hello());
}
```

由于 `demo1` 本身已经在 `lib.rs` 中声明，因此可以直接引用到，对于 `demo2` 采用绝对路径，以 `crate` 起始，对应的，`demo3` 采用相对路径。

> 注意，为了确保 module 以及 say_hello() 能够被正确的访问到，他们都被声明为 pub
