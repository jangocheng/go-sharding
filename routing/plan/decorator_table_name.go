// Copyright 2019 The Gaea Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plan

import (
	"fmt"
	"github.com/XiaoMi/Gaea/core"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"

	"github.com/XiaoMi/Gaea/proxy/router"
	"github.com/pingcap/errors"
)

// TableNameDecorator decorate TableName
type TableNameDecorator struct {
	origin   *ast.TableName
	sharding *core.ShardingTable
	result   *RouteResult
}

// CreateTableNameDecorator create TableNameDecorator
// the table has been checked before
func NewTableNameDecorator(n *ast.TableName, sharding *core.ShardingTable) (*TableNameDecorator, error) {
	if len(n.PartitionNames) != 0 {
		return nil, fmt.Errorf("TableName does not support PartitionNames in sharding")
	}

	ret := &TableNameDecorator{
		origin:   n,
		sharding: sharding,
	}

	return ret, nil
}

// Restore implement ast.Node
func (t *TableNameDecorator) Restore(ctx *format.RestoreCtx) error {
	tableIndex, err := t.result.GetCurrentTableIndex()
	if err != nil {
		return err
	}

	ruleType := t.rule.GetType()

	if t.origin.Schema.String() != "" {
		if ruleType == router.GlobalTableRuleType {
			dbName, err := t.rule.GetDatabaseNameByTableIndex(tableIndex)
			if err != nil {
				return fmt.Errorf("get mycat database name error: %v", err)
			}
			ctx.WriteName(dbName)
			ctx.WritePlain(".")
		} else if router.IsMycatShardingRule(ruleType) {
			dbName, err := t.rule.GetDatabaseNameByTableIndex(tableIndex)
			if err != nil {
				return fmt.Errorf("get mycat database name error: %v", err)
			}
			ctx.WriteName(dbName)
			ctx.WritePlain(".")
		} else {
			ctx.WriteName(t.origin.Schema.String())
			ctx.WritePlain(".")
		}
	}

	// kingshard需要改写表名, mycat不需要改写, 全局表需要改写
	if ruleType == router.GlobalTableRuleType {
		ctx.WriteName(t.origin.Name.String())
	} else if router.IsMycatShardingRule(ruleType) {
		ctx.WriteName(t.origin.Name.String())
	} else {
		ctx.WriteName(fmt.Sprintf("%s_%04d", t.origin.Name.String(), tableIndex))
	}

	for _, value := range t.origin.IndexHints {
		ctx.WritePlain(" ")
		if err := value.Restore(ctx); err != nil {
			return errors.Annotate(err, "An error occurred while splicing IndexHints")
		}
	}
	return nil
}

// Accept implement ast.Node
// do nothing and return current decorator
func (t *TableNameDecorator) Accept(v ast.Visitor) (ast.Node, bool) {
	return t, true
}

// Text implement ast.Node
func (t *TableNameDecorator) Text() string {
	return t.origin.Text()
}

// SetText implement ast.Node
func (t *TableNameDecorator) SetText(text string) {
	t.origin.SetText(text)
}