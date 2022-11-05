// Licensed to LinDB under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. LinDB licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package sql

import (
	"errors"
	"fmt"
	"strconv"

	commonconstants "github.com/lindb/common/constants"

	"github.com/lindb/lindb/aggregation/function"
	"github.com/lindb/lindb/pkg/collections"
	"github.com/lindb/lindb/pkg/strutil"
	"github.com/lindb/lindb/pkg/timeutil"
	"github.com/lindb/lindb/sql/grammar"
	"github.com/lindb/lindb/sql/stmt"
)

// queryStmtParser represents query statement parser using visitor
type queryStmtParser struct {
	baseStmtParser
	explain bool

	selectItems []stmt.Expr
	fieldNames  map[string]struct{} // cache field name include alias

	startTime int64
	endTime   int64

	groupBy  []string
	interval int64
	orderBy  []stmt.Expr

	curOrderByExpr *stmt.OrderByExpr
	hasOrderBy     bool
}

// newQueryStmtParse create a query statement parser
func newQueryStmtParse(explain bool) *queryStmtParser {
	return &queryStmtParser{
		explain:    explain,
		fieldNames: make(map[string]struct{}),
		baseStmtParser: baseStmtParser{
			exprStack: collections.NewStack(),
			namespace: commonconstants.DefaultNamespace,
			limit:     20,
		},
	}
}

// build query statement based on parse result
func (q *queryStmtParser) build() (stmt.Statement, error) {
	if err := q.validation(); err != nil {
		return nil, err
	}

	query := &stmt.Query{}
	query.Explain = q.explain
	query.Namespace = q.namespace
	query.MetricName = q.metricName
	query.SelectItems = q.selectItems
	query.Condition = q.condition

	now := timeutil.Now()
	query.TimeRange = timeutil.TimeRange{Start: q.startTime, End: q.endTime}
	if query.TimeRange.Start <= 0 {
		query.TimeRange.Start = now - timeutil.OneHour
	}
	if query.TimeRange.End <= 0 {
		query.TimeRange.End = now
	}
	if query.TimeRange.End < query.TimeRange.Start {
		return nil, fmt.Errorf("start time cannot be larger than end time")
	}

	query.Interval = timeutil.Interval(q.interval)
	query.GroupBy = q.groupBy
	query.OrderByItems = q.orderBy
	query.Limit = q.limit
	return query, nil
}

// validation tests data if invalid
func (q *queryStmtParser) validation() error {
	if q.err != nil {
		return q.err
	}
	if q.metricName == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if len(q.selectItems) == 0 {
		return fmt.Errorf("select fields cannbe be empty")
	}
	return nil
}

// resetExprStack resets expr stack for next parse fragment
func (q *queryStmtParser) resetExprStack() {
	q.exprStack = collections.NewStack()
}

// visitGroupByKey visits when production groupBy key expression is entered
func (q *queryStmtParser) visitGroupByKey(ctx *grammar.GroupByKeyContext) {
	switch {
	case ctx.Ident() != nil:
		tagKey := strutil.GetStringValue(ctx.Ident().GetText())
		q.groupBy = append(q.groupBy, tagKey)
	case ctx.DurationLit() != nil:
		q.interval = q.parseDuration(ctx.DurationLit())
	}
}

// visitSortField visits when production sort field expression is entered.
func (q *queryStmtParser) visitSortField(ctx *grammar.SortFieldContext) {
	q.hasOrderBy = true
	q.curOrderByExpr = &stmt.OrderByExpr{Desc: len(ctx.AllT_DESC()) > 0}
}

// completeSortField compelted prase order by field.
func (q *queryStmtParser) completeSortField(_ *grammar.SortFieldContext) {
	if q.curOrderByExpr != nil {
		if err := q.check(); err != nil {
			q.err = err
			return
		}
		q.orderBy = append(q.orderBy, q.curOrderByExpr)
	}
	q.hasOrderBy = true
	q.curOrderByExpr = nil
}

// check order by expr if valid, returns err when invalid.
func (q *queryStmtParser) check() error {
	var fieldName string
	switch e := q.curOrderByExpr.Expr.(type) {
	case *stmt.CallExpr:
		if !function.IsSupportOrderBy(e.FuncType) {
			return fmt.Errorf("[%s] function not support order by", e.FuncType)
		}
		if len(e.Params) != 1 {
			return errors.New("order by function params length invalid")
		}
		fieldName = e.Params[0].Rewrite()
	case *stmt.FieldExpr:
		fieldName = e.Name
	}
	_, ok := q.fieldNames[fieldName]
	if !ok {
		return fmt.Errorf("order by field not in select fields, order by field: %s", fieldName)
	}
	return nil
}

// visitTimeRangeExpr visits when production timeRange expression is entered.
func (q *queryStmtParser) visitTimeRangeExpr(ctx *grammar.TimeRangeExprContext) {
	timeExprCtxList := ctx.AllTimeExpr()
	for _, timeExpr := range timeExprCtxList {
		timeExprCtx, ok := timeExpr.(*grammar.TimeExprContext)
		if !ok {
			continue
		}
		var timestamp int64
		var err error
		switch {
		case timeExprCtx.Ident() != nil:
			timestamp, err = timeutil.ParseTimestamp(strutil.GetStringValue(timeExprCtx.Ident().GetText()))
		case timeExprCtx.NowExpr() != nil:
			timestamp = timeutil.Now()
			durationExpr, durationExist := timeExprCtx.NowExpr().(*grammar.NowExprContext)
			if durationExist {
				timestamp += q.parseDuration(durationExpr.DurationLit())
			}
		}
		if err != nil {
			q.err = err
			continue
		}
		binaryOp := timeExprCtx.BinaryOperator()
		if binaryOp == nil {
			continue
		}
		binaryOpCtx, ok := binaryOp.(*grammar.BinaryOperatorContext)
		if !ok {
			continue
		}
		if binaryOpCtx.T_GREATER() != nil || binaryOpCtx.T_GREATEREQUAL() != nil {
			q.startTime = timestamp
		}
		if binaryOpCtx.T_LESS() != nil || binaryOpCtx.T_LESSEQUAL() != nil {
			q.endTime = timestamp
		}
	}
}

// parseDuration parses time duration from duration string
func (q *queryStmtParser) parseDuration(ctx grammar.IDurationLitContext) int64 {
	if ctx == nil {
		return 0
	}
	durationCtx, ok := ctx.(*grammar.DurationLitContext)
	if !ok {
		return 0
	}

	duration, err := strconv.ParseInt(durationCtx.IntNumber().GetText(), 10, 64)
	if err != nil {
		q.err = err
		return 0
	}
	var result int64
	if durationCtx.IntervalItem() == nil {
		return result
	}
	unit, ok := durationCtx.IntervalItem().(*grammar.IntervalItemContext)
	if !ok {
		return result
	}
	switch {
	case unit.T_SECOND() != nil:
		result = duration * timeutil.OneSecond
	case unit.T_MINUTE() != nil:
		result = duration * timeutil.OneMinute
	case unit.T_HOUR() != nil:
		result = duration * timeutil.OneHour
	case unit.T_DAY() != nil:
		result = duration * timeutil.OneDay
	case unit.T_WEEK() != nil:
		result = duration * timeutil.OneWeek
	case unit.T_MONTH() != nil:
		result = duration * timeutil.OneMonth
	case unit.T_YEAR() != nil:
		result = duration * timeutil.OneYear
	}
	return result
}

// visitFieldExpr visits when production field expression is entered
func (q *queryStmtParser) visitFieldExpr(ctx *grammar.FieldExprContext) {
	switch {
	case ctx.ExprFunc() != nil:
		q.exprStack.Push(&stmt.CallExpr{})
	case ctx.T_OPEN_P() != nil:
		q.exprStack.Push(&stmt.ParenExpr{})
	case ctx.T_MUL() != nil:
		q.exprStack.Push(&stmt.BinaryExpr{Operator: stmt.MUL})
	case ctx.T_DIV() != nil:
		q.exprStack.Push(&stmt.BinaryExpr{Operator: stmt.DIV})
	case ctx.T_ADD() != nil:
		q.exprStack.Push(&stmt.BinaryExpr{Operator: stmt.ADD})
	case ctx.T_SUB() != nil:
		q.exprStack.Push(&stmt.BinaryExpr{Operator: stmt.SUB})
	}
}

// visitAlias visits when production alias expression is entered
func (q *queryStmtParser) visitAlias(ctx *grammar.AliasContext) {
	if len(q.selectItems) == 0 {
		return
	}
	if selectItem, ok := (q.selectItems[len(q.selectItems)-1]).(*stmt.SelectItem); ok {
		alias := strutil.GetStringValue(ctx.Ident().GetText())
		selectItem.Alias = alias
		q.fieldNames[alias] = struct{}{}
	}
}

// visitFuncName visits when production function call expression is entered
func (q *queryStmtParser) visitFuncName(ctx *grammar.FuncNameContext) {
	if q.exprStack.Empty() {
		return
	}
	callExpr, ok := q.exprStack.Peek().(*stmt.CallExpr)
	if !ok {
		return
	}
	switch {
	case ctx.T_SUM() != nil:
		callExpr.FuncType = function.Sum
	case ctx.T_MIN() != nil:
		callExpr.FuncType = function.Min
	case ctx.T_MAX() != nil:
		callExpr.FuncType = function.Max
	case ctx.T_COUNT() != nil:
		callExpr.FuncType = function.Count
	case ctx.T_LAST() != nil:
		callExpr.FuncType = function.Last
	case ctx.T_FIRST() != nil:
		callExpr.FuncType = function.First
	case ctx.T_AVG() != nil:
		callExpr.FuncType = function.Avg
	case ctx.T_STDDEV() != nil:
		callExpr.FuncType = function.Stddev
	case ctx.T_QUANTILE() != nil:
		callExpr.FuncType = function.Quantile
	case ctx.T_RATE() != nil:
		callExpr.FuncType = function.Rate
	}
}

// completeFuncExpr completes a function call expression for select list
func (q *queryStmtParser) completeFuncExpr() {
	cur := q.exprStack.Pop()
	if cur != nil {
		expr, ok := cur.(stmt.Expr)
		if ok {
			q.setExprParam(expr)
		}
		if q.exprStack.Empty() {
			if q.hasOrderBy {
				q.curOrderByExpr.Expr = expr
			} else {
				q.selectItems = append(q.selectItems, &stmt.SelectItem{Expr: expr})

				// select field(func rewrite name)
				q.fieldNames[expr.Rewrite()] = struct{}{}
			}
		}
	}
}

// visitExprAtom visits when production atom expr expression is entered
func (q *queryStmtParser) visitExprAtom(ctx *grammar.ExprAtomContext) {
	switch {
	case ctx.Ident() != nil: // field
		q.parseFieldName(strutil.GetStringValue(ctx.Ident().GetText()))
	case ctx.DecNumber() != nil || ctx.IntNumber() != nil:
		valStr := ""
		switch {
		case ctx.DecNumber() != nil:
			valStr = ctx.DecNumber().GetText()
		case ctx.IntNumber() != nil:
			valStr = ctx.IntNumber().GetText()
		}

		val, _ := strconv.ParseFloat(valStr, 64)
		if !q.exprStack.Empty() {
			q.setExprParam(&stmt.NumberLiteral{Val: val})
		}
	default:
	}
}

// parseFieldName parses field name for select/order by expr.
func (q *queryStmtParser) parseFieldName(fieldName string) {
	fieldExpr := &stmt.FieldExpr{Name: fieldName}

	switch {
	case q.hasOrderBy: // handle order by item
		if q.exprStack.Empty() {
			q.curOrderByExpr.Expr = fieldExpr
		} else {
			q.setExprParam(fieldExpr)
		}
	default: // handle select item
		if q.exprStack.Empty() {
			q.selectItems = append(q.selectItems, &stmt.SelectItem{Expr: fieldExpr})
		} else {
			q.setExprParam(fieldExpr)
		}
		q.fieldNames[fieldName] = struct{}{}
	}
}

// completeFieldExpr completes a field expr,
// only paren and binary expr need to do set expr param,
// set function's param in complete func parse section.
func (q *queryStmtParser) completeFieldExpr(ctx *grammar.FieldExprContext) {
	switch {
	case ctx.T_OPEN_P() != nil:
	case ctx.T_MUL() != nil:
	case ctx.T_DIV() != nil:
	case ctx.T_ADD() != nil:
	case ctx.T_SUB() != nil:
	default:
		return
	}

	cur := q.exprStack.Pop()

	if cur != nil {
		expr, ok := cur.(stmt.Expr)
		if ok {
			q.setExprParam(expr)
		}
		if q.exprStack.Empty() {
			q.selectItems = append(q.selectItems, &stmt.SelectItem{Expr: expr})
		}
	}
}
