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
	"github.com/XiaoMi/Gaea/core"
	"github.com/pingcap/parser/ast"
)

func (s *SqlExplain) ExplainWhere(sel *ast.SelectStmt, rewriter Rewriter) error {
	where := sel.Where
	if where != nil {
		expr, err := s.explainCondition(where, rewriter, core.LogicAnd)
		if err != nil {
			sel.Where = expr
		}
	}
	return nil
}