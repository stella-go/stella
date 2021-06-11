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
	"regexp"
	"strings"
)

type Statement struct {
	upperName  string
	start      int
	stop       int
	TableName  string
	ModelName  string
	Properties []*Property
}

func (s *Statement) Fill(sql string) {
	s.upperName = strings.Trim(s.upperName, "`")
	s.TableName = strings.Trim(cut(sql, s.start, s.stop), "`")
	s.ModelName = firstUpperCamelCase(s.TableName)
	for _, property := range s.Properties {
		property.Fill(sql)
	}
}

func (s *Statement) String() string {
	return fmt.Sprintf(`Statement{upperName: %s, start: %d, stop: %d, TableName: %s, ModelName: %s, Properties: %v}`,
		s.upperName, s.start, s.stop, s.TableName, s.ModelName, s.Properties,
	)
}

type Property struct {
	upperName    string
	start        int
	stop         int
	ColumnName   string
	PropertyName string
	SnakeName    string
	DataBaseType string
	Primary      bool
	Uniq         bool
}

func (p *Property) String() string {
	return fmt.Sprintf(`Property{upperName: %s, start: %d, stop: %d, PropertyName: %s, ColumnName: %s, DataBaseType: %v, Primary: %t, Uniq: %t}`,
		p.upperName, p.start, p.stop, p.PropertyName, p.ColumnName, p.DataBaseType, p.Primary, p.Uniq,
	)
}

func (p *Property) Fill(sql string) {
	p.upperName = strings.Trim(p.upperName, "`")
	p.ColumnName = strings.Trim(cut(sql, p.start, p.stop), "`")
	p.PropertyName = firstUpperCamelCase(p.ColumnName)
	p.SnakeName = toSnakeCase(p.PropertyName)
}

func firstUpperCamelCase(s string) string {
	s = toCamelCase(s)
	s = strings.ToUpper(s[0:1]) + s[1:]
	return s
}

func toCamelCase(s string) string {
	re := regexp.MustCompile(`_(\w)`)
	return re.ReplaceAllStringFunc(s, toUpper)
}

func toSnakeCase(s string) string {
	re := regexp.MustCompile(`[A-Z]`)
	snake := re.ReplaceAllStringFunc(s, toSnake)
	return strings.Trim(snake, "_")
}

func toUpper(s string) string {
	return strings.ToUpper(s[1:])
}

func toSnake(s string) string {
	return "_" + strings.ToLower(s[:1])
}

func cut(s string, start int, stop int) string {
	rns := []rune(s)
	return string(rns[start : stop+1])
}

type PrimaryKey []string

type UniqKey []string

type ColumnType struct {
	typeName string
	Primary  bool
	Uniq     bool
}
