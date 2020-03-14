---
title: Rust 怎么写测试
date: 2020-03-14 14:59:28
tags: 
- rust
- test
categories:
- Rust
---

## Rust 怎么写测试
### 先跑起来
不说什么具体的知识，我们先一步步的，来写个最简单的测试，并且让他啊跑起来，看看 rust 下的测试是什么样子的：
1. 创建个 lib 工程：`cargo new simplest-test --lib`
2. 在 `src/lib.rs` 里面，rust 已经自动帮我们写下了如下代码：
    ```rust
    #[cfg(test)]
    mod test {
        #[test]
        fn it_works() {
            assert_eq!(2 + 2, 4);
        }
    }
    ```
3. 运行一下：`cargo test`
    ```
       Compiling simplest-test v0.1.0 (/User/xxx/simplest-test)
            Finished dev [unoptimized + debuginfo] target(s) in 5.91s
                Running target/debug/deps/simplest_test-c430fbaec5f55b85

    running 1 test
    test tests::it_works ... ok

    test result: ok. 1 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out

        Doc-tests simplest-test

    running 0 tests

    test result: ok. 0 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out
    ```

经过上述三步，我们已经创建了一个最简单的测试工程，并且运行了自带的测试。

在代码层面，测试本身无需多说，结构上我们看到，与业务代码不同，测试在 module 上增加了 attribute: `#[cfg(test)]`，在测试方法上增加了 attribute: `#[test]`。
- `#[cfg(test)]`：配置编译条件，只有在 test 模式下，被标记的代码块才会被编译（换句话说，它确保 release 中不包含测试代码）
- `#[test]`：被标记的方法将被视为测试来执行

在 output 中，还包含了两部分“running x test”，第一部分是我们已有的测试，第二部分为文档测试，本文暂不涉及。

### 断言
rust 原生提供了几种简单的测试断言，能够满足基本的测试需求，以下是 rust 的测试断言与 junit 测试断言的对应表：

rust | junit
---|---
assert!() | assertTrue()
assert_eq!() | assertEqual()
assert_ne!() | assertNotEqual()
#[should_panic(expected = "{error message}")] | assertThrows()

> 对 panic 的断言使用的是 attribute 而不是 macro

### 单元测试
rust 的封装性与 java 略有不同，只有 default(private) 与 pub(public)，那么有个问题：我想写个单元测试，难道必须要把需要测试的函数都 pub 出来吗？（其实 java 里面也有要不要测试 private 方法的讨论，参见《修改代码的艺术》）

而在 rust 中，我们不用过多担心测试私有函数的问题，因为通常 rust 的单元测试会直接与被测代码放置在一起，见以下代码：
```rust
// src/biz.rs
fn int_adder(op1:i32, op2:i32) -> i32 {
    op1 + op2
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn should_return_4_when_2_add_2() {
        assert_eq!(int_adder(2, 2), 4);
    }
}
```
按照上述方式，在 biz 中定义了 child module: `tests`，child module 能够合法的访问他的 parent，因此能够直接调用其 parent 的私有方法。

同时，`#[cfg(test)]`确保了该单元测试不会被编译（除非在 test 模式下）。

### 集成测试
与 java 类似，rust 也能识别特定的测试目录 tests, tests 目录与 src 目录并列，作为专门存放集成测试的地方。

显然，放在 tests 中的测试代码不再能访问到私有函数了。不过集成测试本来也不应该测试细节。另外，放置在 tests 目录下的测试代码，不再需要`#[cfg(test)]` attribute 了，rust 编译器会自动识别所有 tests 目录下的代码为测试。

对上一节中的例子进行修改，得到如下集成测试：
```rust
-----------------------------------------------
// src/biz.rs
-----------------------------------------------
pub fn int_adder(op1:i32, op2:i32) -> i32 {
    op1 + op2
}

-----------------------------------------------
// src/lib.rs
-----------------------------------------------
pub mod biz;

-----------------------------------------------
// tests/biz_test.rs
-----------------------------------------------
use simplest_test::biz;

#[test]
fn should_return_4_when_2_add_2() {
    assert_eq!(biz::int_adder(2, 2), 4);
}
```

再看看测试输出：
```
   Compiling simplest-test v0.1.0 (/Users/xxx/simplest-test)
    Finished dev [unoptimized + debuginfo] target(s) in 0.55s
     Running target/debug/deps/simplest_test-c430fbaec5f55b85

running 0 tests

test result: ok. 0 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out

     Running target/debug/deps/biz_test-c9cf61d397408434

running 1 test
test should_return_4_when_2_add_2 ... ok

test result: ok. 1 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out

   Doc-tests simplest-test

running 0 tests
```
可以看到，与前文相比，一共有三部分 `running x tests`，比之前多的一部分，就是我们新增的集成测试了。tests 目录下的测试会单独占据一块输出。

### 测试运行控制
1. 默认情况下，rust 采用多线程并行执行所有测试，当有串行需要时可以执行：`cargo test -- --test-threads={thread_numbers}`来控制执行测试的线程数。
2. rust 默认不打印 passed test 的任何输出，当有需要打印输出时，执行：`cargo test -- --show-output`
3. 当期望只运行某个特定测试时，执行：`cargo test {test_function_name}`
4. 当期望只运行某一类测试时，执行：`cargo test {test_function_name_matcher}`
5. 与 junit 类似，当期望 ignore 测试时，在测试函数上添加： `#[ignore]` attribute
