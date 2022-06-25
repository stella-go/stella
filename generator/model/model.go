// Copyright 2010-2022 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

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
	"github.com/stella-go/stella/generator"
	"github.com/stella-go/stella/generator/parser"
)

var (
	typeMapping = map[string]string{
		"TINYINT":   "*n.Int",
		"INT":       "*n.Int",
		"BIGINT":    "*n.Int64",
		"FLOAT":     "*n.Float64",
		"CHAR":      "*n.String",
		"VARCHAR":   "*n.String",
		"TEXT":      "*n.String",
		"DATE":      "*n.Time",
		"DATETIME":  "*n.Time",
		"TIMESTAMP": "*n.Time",
		"default":   "interface{}",
	}
	typeImportsMapping = map[string]string{
		"*n.Bool":    "github.com/stella-go/siu/t/n",
		"*n.Int":     "github.com/stella-go/siu/t/n",
		"*n.Int64":   "github.com/stella-go/siu/t/n",
		"*n.Float64": "github.com/stella-go/siu/t/n",
		"*n.String":  "github.com/stella-go/siu/t/n",
		"*n.Time":    "github.com/stella-go/siu/t/n",
	}
)

type Field struct {
	name string
	typ  string
	tag  string
}

func (f *Field) String() string {
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
		line := f.name + ": %s"
		formats = append(formats, line)
		arg := "s." + f.name
		args = append(args, arg)
	}

	return "func (s *" + s.name + ") String() string {\n\treturn fmt.Sprintf(\"" + s.name + "{" + strings.Join(formats, ", ") + "}\", " + strings.Join(args, ", ") + ")\n}\n"
}

func Generate(pkg string, statements []*parser.Statement, banner bool) string {
	importsMap := make(map[string]common.Void)
	importsMap["fmt"] = common.Null
	structs := make([]string, 0)
	for _, statement := range statements {
		fields := make([]*Field, 0)
		for _, col := range statement.Columns {
			typ, ok := typeMapping[col.Type]
			if !ok {
				typ = typeMapping["default"]
			}
			importsMap[typeImportsMapping[typ]] = common.Null

			tag := fmt.Sprintf("form:\"%s\" json:\"%s,omitempty\"", generator.ToSnakeCase(col.ColumnName.Name), generator.ToSnakeCase(col.ColumnName.Name))
			field := &Field{generator.FirstUpperCamelCase(col.ColumnName.Name), typ, tag}
			fields = append(fields, field)
		}
		struc := &Struct{generator.FirstUpperCamelCase(statement.TableName.Name), fields}
		structs = append(structs, struc.String())
	}

	importsLines := make([]string, 0)
	for i := range importsMap {
		if i == "" {
			continue
		}
		importsLines = append(importsLines, "\t\""+i+"\"")
	}
	bannerS := ""
	if banner {
		bannerS = fmt.Sprintf("\n/**\n * Auto Generate by github.com/stella-go/stella on %s.\n */\n", time.Now().Format("2006/01/02"))

	}
	return fmt.Sprintf("package %s\n%s\nimport (\n%s\n)\n\n%s", pkg, bannerS, strings.Join(importsLines, "\n"), strings.Join(structs, "\n"))
}
