/*
 *
 *  * Copyright 2021. Go-Sharding Author All Rights Reserved.
 *  *
 *  *  Licensed under the Apache License, Version 2.0 (the "License");
 *  *  you may not use this file except in compliance with the License.
 *  *  You may obtain a copy of the License at
 *  *
 *  *      http://www.apache.org/licenses/LICENSE-2.0
 *  *
 *  *  Unless required by applicable law or agreed to in writing, software
 *  *  distributed under the License is distributed on an "AS IS" BASIS,
 *  *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  *  See the License for the specific language governing permissions and
 *  *  limitations under the License.
 *  *
 *  *  File author: Anders Xiao
 *
 */

package config

import "github.com/XiaoMi/Gaea/core"

type Settings struct {
	DataSources  map[string]*core.DataSource
	ShardingRule *ShardingRule
}

type ShardingRule struct {
	Tables map[string]*core.ShardingTable
}

func NewSettings() *Settings {
	s := &Settings{
		DataSources: make(map[string]*core.DataSource),
		ShardingRule: &ShardingRule{
			Tables: make(map[string]*core.ShardingTable),
		},
	}
	return s
}