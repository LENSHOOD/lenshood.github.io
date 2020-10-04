---
title: TiDB 学习课程 Lesson-6
date: 2020-10-03 23:06:46
tags:
- tidb
categories:
- TiDB
---

本节课程主要学习的是 TiDB 的 planer 模块，planer 模块的主要功能是将 AST 转化为实际的执行计划，在这当中，包含了两个阶段的优化过程：逻辑优化、物理优化。优化过后的执行计划可以直接构造对应的 Executor 来执行（回忆 Executor 章节的火山模型）。

## Planer 概览

Planer 构造执行计划的入口在 `compiler.go` 的 `Compile()` 函数，函数对传入的 `ast.StmtNode` 进行一番预处理、优化过后，会生成一个 `ExecStmt` 供下一步操作继续。

有趣的是，在整个 planer 的操作中，所有的动作都采用 Visitor 模式实现，通过访问 `ast.StmtNode` 来获取信息进行操作。

在`ast.StmtNode` 中继承了 `ast.Node` ，在 `ast.Node` 中定义了接受 `Visitor` 的入口：

```go
Accept(v Visitor) (node Node, ok bool)
```

再反过来看看 `Visitor`：

```go
type Visitor interface {
	// Enter is called before children nodes are visited.
	// The returned node must be the same type as the input node n.
	// skipChildren returns true means children nodes should be skipped,
	// this is useful when work is done in Enter and there is no need to visit children.
	Enter(n Node) (node Node, skipChildren bool)
	// Leave is called after children nodes have been visited.
	// The returned node's type can be different from the input node if it is a ExprNode,
	// Non-expression node must be the same type as the input node n.
	// ok returns false to stop visiting.
	Leave(n Node) (node Node, ok bool)
}
```

之所以这么做，是因 parser 根据不同语法生成的 AST 有各种各样的形式，对于某种语法节点，可能会存在多种多样的操作（包括预处理、优化等），为了分离关注点而采用 Visitor 模式来讲对同一组数据的不同处理进行划分。

回到 `Compile()`，函数大致如下：

```go
func (c *Compiler) Compile(ctx context.Context, stmtNode ast.StmtNode) (*ExecStmt, error) {
  ... ..
	if err := plannercore.Preprocess(c.Ctx, stmtNode, infoSchema); err != nil {
		return nil, err
	}
	stmtNode = plannercore.TryAddExtraLimit(c.Ctx, stmtNode)

	finalPlan, names, err := planner.Optimize(ctx, c.Ctx, stmtNode, infoSchema)
	... ...
	var lowerPriority bool
	if c.Ctx.GetSessionVars().StmtCtx.Priority == mysql.NoPriority {
		lowerPriority = needLowerPriority(finalPlan)
	}
	... ...
}
```

其中：

1. `Preprocess()` 主要对 AST 中各种节点进行检查，校验等。
2. `TryAddExtraLimit()`在[系统变量](https://docs.pingcap.com/zh/tidb/stable/system-variables)中`sql_select_limit` 置位时添加对应的 `LIMIT` 语句。
3. `Optimize()` 即执行逻辑优化和物理优化（亦是本文的重点）。
4. `needLowerPriority() ` 会统计预测查询结果行数大于某个门限时降低其执优先级。

## 逻辑优化


## 物理优化

