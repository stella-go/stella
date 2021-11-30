// Copyright 2010-2021 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"fmt"
	"strings"
)

func cut(s string, start int, stop int) string {
	rns := []rune(s)
	return string(rns[start : stop+1])
}

type NameDefinition struct {
	Name  string
	name  string
	start int
	stop  int
}

func (p *NameDefinition) Fill(sql string) {
	p.Name = strings.Trim(cut(sql, p.start, p.stop), "`")
}

func (p *NameDefinition) String() string {
	return p.Name
}

type ColumnDefinition struct {
	ColumnName       *NameDefinition
	Type             string
	PrimaryKey       bool
	UniqueKey        bool
	AutoIncrement    bool
	CurrentTimestamp bool
}

func (p *ColumnDefinition) String() string {
	return fmt.Sprintf("ColumnDefinition{ ColumnName: %v, Type: %v, PrimaryKey: %v, UniqKey: %v, AutoIncrement: %v, CurrentTimestamp: %v}",
		p.ColumnName,
		p.Type,
		p.PrimaryKey,
		p.UniqueKey,
		p.AutoIncrement,
		p.CurrentTimestamp,
	)
}

type PrimaryKeyPair []*NameDefinition

type UniqueKeyPair []*NameDefinition

type IndexKeyPair []*NameDefinition

type AutoIncrement struct{}

type PrimaryKey struct{}

type UniqKey struct{}

type IndexKey struct{}

type CurrentTimestamp struct{}

type Statement struct {
	TableName       *NameDefinition
	Columns         []*ColumnDefinition
	PrimaryKeyPairs []PrimaryKeyPair
	UniqKeyPairs    []UniqueKeyPair
	IndexKeyPairs   []IndexKeyPair
}

func (p *Statement) Fill(sql string) {
	p.TableName.Fill(sql)
	for _, column := range p.Columns {
		column.ColumnName.Fill(sql)
	}
	for _, pair := range p.PrimaryKeyPairs {
		for _, name := range pair {
			name.Fill(sql)
		}
	}
	for _, pair := range p.UniqKeyPairs {
		for _, name := range pair {
			name.Fill(sql)
		}
	}
	for _, pair := range p.IndexKeyPairs {
		for _, name := range pair {
			name.Fill(sql)
		}
	}
}

func (p *Statement) String() string {
	return fmt.Sprintf("Statement{TableName: %v,Columns: %v,PrimaryKeyPairs: %v,UniqKeyPairs: %v,IndexKeyPairs: %v}",
		p.TableName,
		p.Columns,
		p.PrimaryKeyPairs,
		p.UniqKeyPairs,
		p.IndexKeyPairs,
	)
}
