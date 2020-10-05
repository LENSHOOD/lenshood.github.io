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
3. `Optimize()` 即生成执行计划，并进行逻辑优化和物理优化（亦是本文的重点）。
4. `needLowerPriority() ` 会统计预测查询结果行数大于某个门限时降低其执优先级。

## 生成执行计划
生存执行计划的主要入口是 `PlanBuilder.Build()` 方法。

`Build()`方法实际上也只起分发的作用，根据`ast.Node` 的类型，将之分发至不同的分支下：

```go
func (b *PlanBuilder) Build(ctx context.Context, node ast.Node) (Plan, error) {
	b.optFlag |= flagPrunColumns
	switch x := node.(type) {
	case *ast.AdminStmt:
		return b.buildAdmin(ctx, x)
	... ...
	case *ast.SelectStmt:
		if x.SelectIntoOpt != nil {
			return b.buildSelectInto(ctx, x)
		}
		return b.buildSelect(ctx, x)
	... ...
	}
	return nil, ErrUnsupportedType.GenWithStack("Unsupported type %T", node)
}
```

我们以 `SelectStmt`为例，进入 `buildSelect()` 方法后，我们能看到整个方法体非常长，主要原因是需要对整个 select 语句的各个部分进行逻辑计划的转换，包括 join、group by、limit 等等。

整个方法会返回一个 `LogicalPlan`，里面包含了大量的信息：

```go
type LogicalPlan interface {
	Plan

	// HashCode encodes a LogicalPlan to fast compare whether a LogicalPlan equals to another.
	// We use a strict encode method here which ensures there is no conflict.
	HashCode() []byte

	// PredicatePushDown pushes down the predicates in the where/on/having clauses as deeply as possible.
	// It will accept a predicate that is an expression slice, and return the expressions that can't be pushed.
	// Because it might change the root if the having clause exists, we need to return a plan that represents a new root.
	PredicatePushDown([]expression.Expression) ([]expression.Expression, LogicalPlan)

	// PruneColumns prunes the unused columns.
	PruneColumns([]*expression.Column) error

	// findBestTask converts the logical plan to the physical plan. It's a new interface.
	// It is called recursively from the parent to the children to create the result physical plan.
	// Some logical plans will convert the children to the physical plans in different ways, and return the one
	// With the lowest cost and how many plans are found in this function.
	// planCounter is a counter for planner to force a plan.
	// If planCounter > 0, the clock_th plan generated in this function will be returned.
	// If planCounter = 0, the plan generated in this function will not be considered.
	// If planCounter = -1, then we will not force plan.
	findBestTask(prop *property.PhysicalProperty, planCounter *PlanCounterTp) (task, int64, error)

	// BuildKeyInfo will collect the information of unique keys into schema.
	// Because this method is also used in cascades planner, we cannot use
	// things like `p.schema` or `p.children` inside it. We should use the `selfSchema`
	// and `childSchema` instead.
	BuildKeyInfo(selfSchema *expression.Schema, childSchema []*expression.Schema)

	// pushDownTopN will push down the topN or limit operator during logical optimization.
	pushDownTopN(topN *LogicalTopN) LogicalPlan

	// recursiveDeriveStats derives statistic info between plans.
	recursiveDeriveStats(colGroups [][]*expression.Column) (*property.StatsInfo, error)

	// DeriveStats derives statistic info for current plan node given child stats.
	// We need selfSchema, childSchema here because it makes this method can be used in
	// cascades planner, where LogicalPlan might not record its children or schema.
	DeriveStats(childStats []*property.StatsInfo, selfSchema *expression.Schema, childSchema []*expression.Schema, colGroups [][]*expression.Column) (*property.StatsInfo, error)

	// ExtractColGroups extracts column groups from child operator whose DNVs are required by the current operator.
	// For example, if current operator is LogicalAggregation of `Group By a, b`, we indicate the child operators to maintain
	// and propagate the NDV info of column group (a, b), to improve the row count estimation of current LogicalAggregation.
	// The parameter colGroups are column groups required by upper operators, besides from the column groups derived from
	// current operator, we should pass down parent colGroups to child operator as many as possible.
	ExtractColGroups(colGroups [][]*expression.Column) [][]*expression.Column

	// PreparePossibleProperties is only used for join and aggregation. Like group by a,b,c, all permutation of (a,b,c) is
	// valid, but the ordered indices in leaf plan is limited. So we can get all possible order properties by a pre-walking.
	PreparePossibleProperties(schema *expression.Schema, childrenProperties ...[][]*expression.Column) [][]*expression.Column

	// exhaustPhysicalPlans generates all possible plans that can match the required property.
	// It will return:
	// 1. All possible plans that can match the required property.
	// 2. Whether the SQL hint can work. Return true if there is no hint.
	exhaustPhysicalPlans(*property.PhysicalProperty) (physicalPlans []PhysicalPlan, hintCanWork bool)

	// ExtractCorrelatedCols extracts correlated columns inside the LogicalPlan.
	ExtractCorrelatedCols() []*expression.CorrelatedColumn

	// MaxOneRow means whether this operator only returns max one row.
	MaxOneRow() bool

	// Get all the children.
	Children() []LogicalPlan

	// SetChildren sets the children for the plan.
	SetChildren(...LogicalPlan)

	// SetChild sets the ith child for the plan.
	SetChild(i int, child LogicalPlan)

	// rollBackTaskMap roll back all taskMap's logs after TimeStamp TS.
	rollBackTaskMap(TS uint64)
}
```

实现了 `LogicalPlan` 类型的结构有很多种，包括`LogicalJoin`、`LogicalAggregation`、`LogicalSelection` 等等，最终返回的逻辑计划，实际上是一个嵌套结构，包含了层层计划的嵌套。

`Build()` 方法根据规则，将会把 AST 生成为一个基础的、未经过优化的执行计划，以下述语句为例：

```sql
select b from t1, t2 where t1.c = t2.c and t1.a > 5
```

经过`Build()`后会生成如下图的执行计划：

{% asset_img original-plan.png %}

其中，数据（DataSource）从 t1，t2 两张表中被获取，之后以 `t1.c = t2.c` 作为条件进行 join，之后使用一个 selection 来处理 `t1.a > 5` 的筛选，最终，对返回结果进行投影（Projection），将需要的列 b 投影出来。

## 逻辑优化

在得到了原始的执行计划后，接下来就会对原始的计划进行逐级优化，优化逻辑的入口是`plannercore.DoOptimize()`函数：

```go
func DoOptimize(ctx context.Context, sctx sessionctx.Context, flag uint64, logic LogicalPlan) (PhysicalPlan, float64, error) {
	... ...
	logic, err := logicalOptimize(ctx, flag, logic)
	... ...
	physical, cost, err := physicalOptimize(logic, &planCounter)
	... ...
	finalPlan := postOptimize(sctx, physical)
	return finalPlan, cost, nil
}
```

可以看到，整个优化过程很明确：

1. `logicalOptimize()` 进行逻辑优化
2. `physicalOptimize()`进行物理优化
3. `postOptimize()` 进行后优化

最后生成`finalPlan`并返回。

### 逻辑优化种类

我们在`logicalOptimize()`中发现，整个优化过程实际上就是将 `logicalPlan` 依次通过所有的逻辑优化器，各种逻辑优化器声明在`optimizer.go` 中定义的一个变量 `optRuleList` 中：

```go
var optRuleList = []logicalOptRule{
	&gcSubstituter{},
	&columnPruner{},
	&buildKeySolver{},
	&decorrelateSolver{},
	&aggregationEliminator{},
	&projectionEliminator{},
	&maxMinEliminator{},
	&ppdSolver{},
	&outerJoinEliminator{},
	&partitionProcessor{},
	&aggregationPushDownSolver{},
	&pushDownTopNOptimizer{},
	&joinReOrderSolver{},
	&columnPruner{}, // column pruning again at last, note it will mess up the results of buildKeySolver
}
```

显然，每一行都是一种优化器，例如`gcSubstituter`用于将表达式替换为虚拟生成列，以便于使用索引。`columnPruner`用于对列进行剪裁，即去除用不到的列，避免将他们读取出来，以减小数据读取量。

因此，每一种优化器，都是针对特定场景来进行优化，其手段或降低数据量，或减少计算量，最终目的都是为了提升性能。

### 列剪裁 Column Pruner

列剪裁主要是将算子中用不到的列去掉，以减少读取的总数据量，毕竟用不到的列，读取出来也毫无意义。

比如：

```sql
select a from t where b > 5
```

假设表 t 共有 abcd 四个列，在上述语句中，只用到了展示列 a 与 筛选列 b，因此 c d 根本没必要读取，因此从最根部的 DataSource 处就不需要 c d 两列。

前一节 `LogicalPlan` 的定义中，就定义了 `PruneColumns([]*expression.Column) error` 方法，因此包括 join、aggregation、datasource 在内的多种执行计划都实现了该方法。

### 聚合消除 Aggregation Eliminator

聚合消除能够在 `group by {unique key}` 时将不需要的聚合计算消除掉，以减少计算量。那么，什么样的聚合可以消除掉呢？

```sql
select min(b) from t group by a
```

在上述语句中，假如 a 列存在唯一键，那么上述语句实际上与：

```sql
select b from t group by a
```

是等价的，因此`min()`操作毫无必要。

类似的，`count()`、`sum()`、`avg()` 也可以在 group by 唯一键时消除掉。

### 谓词下推 Predicate Push Down

谓词下推是一个非常常用也非常重要的优化形式，它的核心观点是，将可能的筛选条件，尽可能的下推至执行计划的叶子节点（即最先执行筛选），这样就能从源头减少数据量，从而减少后续所有操作的数据量。

举例说明：

```sql
select * from t1, t2 where t1.a > 3 and t2.b > 5
```

原始的执行计划中，会先将 t1 t2 进行 join，之后对 join 的结果再进行 `t2.b > 5`的筛选，然而，如果我们能够先将 `t2.b > 5`作用在 DataSource 上，读取出的 t2 本身就已经少了很多，这时再进行 join，数据量就会明显的减少。

不过谓词下推也存在局限，谓词下推不能推过 MaxOneRow 和 Limit 节点，毕竟先进行筛选，再 limit，和先 limit，再筛选是两个概念。

## 物理优化

