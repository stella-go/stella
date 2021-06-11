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

package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/stella-go/stella/common"
	"github.com/stella-go/stella/generator/parser"
)

var (
	fmtPlaceHolderMapping = map[string]string{
		"bool":    "%t",
		"int":     "%d",
		"string":  "%s",
		"default": "%v",
	}
	typeMapping = map[string]string{
		"TINYINT":   "bool",
		"INT":       "int",
		"BIGINT":    "int64",
		"VARCHAR":   "string",
		"TEXT":      "string",
		"DATE":      "time.Time",
		"DATETIME":  "time.Time",
		"TIMESTAMP": "time.Time",
		"default":   "interface{}",
	}
	typeImportsMapping = map[string]string{
		"time.Time": "time",
	}
)

type Field struct {
	name string
	typ  string
	tag  string
}

func (f *Field) String() string {
	time.Now()
	return fmt.Sprintf("%s %s `%s`", f.name, f.typ, f.tag)
}

type Struct struct {
	name   string
	fields []*Field
}

func (s *Struct) String() string {
	lines := make([]string, 0)
	for _, field := range s.fields {
		lines = append(lines, "\t"+field.String())
	}
	return fmt.Sprintf("type %s struct {\n%s\n}\n%s", s.name, strings.Join(lines, "\n"), s.toString())
}

func (s *Struct) toString() string {
	formats := make([]string, 0)
	args := make([]string, 0)
	for _, f := range s.fields {
		placeHolder, ok := fmtPlaceHolderMapping[f.typ]
		if !ok {
			placeHolder = fmtPlaceHolderMapping["default"]
		}
		line := f.name + ": " + placeHolder
		formats = append(formats, line)
		arg := "s." + f.name
		args = append(args, arg)
	}

	return "func (s *" + s.name + ") String() string {\n\treturn fmt.Sprintf(\"" + s.name + "{" + strings.Join(formats, ", ") + "}\", " + strings.Join(args, ", ") + ")\n}\n"
}

func Generate(pkg string, statements []*parser.Statement) string {
	importsMap := make(map[string]common.Void)
	importsMap["fmt"] = common.Null
	structs := make([]string, 0)
	for _, statement := range statements {
		fields := make([]*Field, 0)
		for _, prop := range statement.Properties {
			typ, ok := typeMapping[prop.DataBaseType]
			if !ok {
				typ = typeMapping["default"]
			}
			importsMap[typeImportsMapping[typ]] = common.Null

			tag := fmt.Sprintf("json:\"%s\"", prop.SnakeName)
			field := &Field{prop.PropertyName, typ, tag}
			fields = append(fields, field)
		}
		struc := &Struct{statement.ModelName, fields}
		structs = append(structs, struc.String())
	}

	importsLines := make([]string, 0)
	for i := range importsMap {
		if i == "" {
			continue
		}
		importsLines = append(importsLines, "\t\""+i+"\"")
	}
	return fmt.Sprintf("package %s\n\n/**\n * Auto Generate by github.com/stella-go/stella on %s.\n */\n\nimport (\n%s\n)\n\n%s", pkg, time.Now().Format("2006/01/02"), strings.Join(importsLines, "\n"), strings.Join(structs, "\n"))
}
