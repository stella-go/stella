// Copyright 2010-2024 the original author or authors.

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
	OnUpdate         bool
	NotNull          bool
	DefaultValue     *DefaultValue
	CurrentTimestamp bool
	Comment          *Comment
}

func (p *ColumnDefinition) Fill(sql string) {
	if p.ColumnName != nil {
		p.ColumnName.Fill(sql)
	}
	if p.DefaultValue != nil && p.DefaultValue.start != 0 && p.DefaultValue.stop != 0 {
		p.DefaultValue.Fill(sql)
	}
	if p.Comment != nil {
		p.Comment.Fill(sql)
	}
}

func (p *ColumnDefinition) String() string {
	return fmt.Sprintf("ColumnDefinition{ ColumnName: %v, Type: %v, PrimaryKey: %v, UniqKey: %v, AutoIncrement: %v, OnUpdate: %v, NotNull: %v, DefaultValue: %v, CurrentTimestamp: %v, Comment: %s}",
		p.ColumnName,
		p.Type,
		p.PrimaryKey,
		p.UniqueKey,
		p.AutoIncrement,
		p.OnUpdate,
		p.NotNull,
		p.DefaultValue,
		p.CurrentTimestamp,
		p.Comment,
	)
}

type PrimaryKeyPair []*NameDefinition

type UniqueKeyPair []*NameDefinition

type IndexKeyPair []*NameDefinition

type AutoIncrement struct {
	autoIncrement bool
	onUpdate      bool
}

type PrimaryKey struct{}

type UniqKey struct{}

type IndexKey struct{}

type NotNull struct{}

type DefaultValue struct {
	currentTimestamp bool
	onUpdate         bool
	DefaultValue     bool
	Value            string
	start            int
	stop             int
}

func (p *DefaultValue) Fill(sql string) {
	value := cut(sql, p.start, p.stop)
	value = strings.TrimLeft(value, "(")
	value = strings.TrimRight(value, ")")
	p.Value = value
}

type CurrentTimestamp struct{}

type Statement struct {
	TableName       *NameDefinition
	Columns         []*ColumnDefinition
	PrimaryKeyPairs []PrimaryKeyPair
	UniqKeyPairs    []UniqueKeyPair
	IndexKeyPairs   []IndexKeyPair
	Comment         *Comment
}

func (p *Statement) Fill(sql string) {
	p.TableName.Fill(sql)
	for _, column := range p.Columns {
		column.Fill(sql)
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
	if p.Comment != nil {
		p.Comment.Fill(sql)
	}
}

func (p *Statement) String() string {
	return fmt.Sprintf("Statement{TableName: %v, Columns: %v, PrimaryKeyPairs: %v, UniqKeyPairs: %v, IndexKeyPairs: %v, Comment: %s}",
		p.TableName,
		p.Columns,
		p.PrimaryKeyPairs,
		p.UniqKeyPairs,
		p.IndexKeyPairs,
		p.Comment,
	)
}

type Comment struct {
	Comment string
	comment string
	start   int
	stop    int
}

func (p *Comment) Fill(sql string) {
	comment := cut(sql, p.start, p.stop)
	quote := '\x00'
	start, end := 0, 0
	for i, c := range []rune(comment) {
		switch c {
		case '\'':
			if start == 0 {
				start = i
				quote = c
			} else if c == quote {
				end = i
			}
		case '"':
			if start == 0 {
				start = i
				quote = c
			} else if c == quote {
				end = i
			}
		}
	}
	p.Comment = comment[start+1 : end]
}

func (p *Comment) String() string {
	return p.Comment
}
