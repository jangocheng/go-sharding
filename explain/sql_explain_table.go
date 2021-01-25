/*
 * Copyright 2021. Go-Sharding Author All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *  File author: Anders Xiao
 */

package explain

import (
	"errors"
	"fmt"
	"github.com/XiaoMi/Gaea/core"
	"github.com/pingcap/parser/ast"
)

func (s *SqlExplain) ExplainTables(sel *ast.SelectStmt, rewriter Rewriter) error {
	if sel.From == nil {
		return errors.New("select 'from' statement is missing")
	}

	join := sel.From.TableRefs
	if join == nil {
		return errors.New("there is an unknown syntax in the select 'from' statement")
	}
	if err := s.explainJoin(join, rewriter); err != nil {
		return err
	}
	return nil
}

func (s *SqlExplain) explainJoin(join *ast.Join, rewriter Rewriter) error {
	if err := checkLimitJoinClause(join); err != nil {
		return fmt.Errorf("invalid join statement: %v", err)
	}

	// 只允许最多两个表的JOIN
	if join.Left != nil {
		err := s.explainJoinSide(join.Left, rewriter, true)
		if err != nil {
			return err
		}
	}
	if join.Right != nil {
		err := s.explainJoinSide(join.Right, rewriter, false)
		if err != nil {
			return err
		}
	}

	// 改写ON条件
	if join.On != nil {
		err := s.explainJoinOn(join.On, rewriter)
		if err != nil {
			return fmt.Errorf("rewrite on condition error: %v", err)
		}
	}

	return nil
}

func (s *SqlExplain) explainJoinSide(joinSide ast.ResultSetNode, rewriter Rewriter, allowNestedJoin bool) error {
	switch sideNode := joinSide.(type) {
	case *ast.TableSource:
		// 改写两个表的node
		err := s.rewriteTableSource(sideNode, rewriter)
		if err != nil {
			return err
		}
	case *ast.Join:
		if allowNestedJoin {
			if err := s.explainJoin(sideNode, rewriter); err != nil {
				return fmt.Errorf("explain nested join statement error: %v", err)
			}
		} else {
			return fmt.Errorf("one side of the join statement is not TableSource, type: %T", joinSide)
		}
	default:
		return fmt.Errorf("invalid sideNode type: %T", joinSide)
	}
	return nil
}

func (s *SqlExplain) explainJoinOn(on *ast.OnCondition, rewriter Rewriter) error {
	newExpr, err := s.explainCondition(on.Expr, rewriter, core.LogicAnd)
	if err != nil {
		return err
	}
	on.Expr = newExpr
	return nil
}

func (s *SqlExplain) rewriteTableSource(table *ast.TableSource, rewriter Rewriter) error {
	lookup := s.CurrentContext().TableLookup()
	switch name := table.Source.(type) {
	case *ast.TableName:
		if s.subQueryDepth.Get() == 0 {
			err := lookup.addTable(table, s.shardingProvider)
			if err != nil {
				return err
			}
		}
		result, e := rewriter.RewriteTable(name, s.CurrentContext())
		if e != nil {
			return fmt.Errorf("rewrite left TableSource error: %v", e)
		}
		if result.IsRewrote() {
			table.Source = result.GetNewNode()
		}
		return nil
	case *ast.SelectStmt:
		if s.maxSubQueryDepth == int32(0) {
			return errors.New("sub query is not supported")
		}
		if s.subQueryDepth.Add(1) > s.maxSubQueryDepth {
			return errors.New("too many subquery")
		}
		return s.explainSubQuery(name, rewriter)
	default:
		return fmt.Errorf("field Source cannot handle, type: %T", table.Source)
	}
}

// 检查TableRefs中存在的不允许在分表中执行的语法
func checkLimitJoinClause(join *ast.Join) error {
	// 不允许USING的列名中出现DB名和表名, 因为目前Join子句的TableName不方便加装饰器
	for _, c := range join.Using {
		if c.Schema.String() != "" {
			return fmt.Errorf("JOIN does not support USING column with schema")
		}
		if c.Table.String() != "" {
			return fmt.Errorf("JOIN does not support USING column with table")
		}
	}
	return nil
}
