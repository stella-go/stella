// Copyright 2010-2023 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/stella-go/stella/common"
	"github.com/stella-go/stella/generator"
	"github.com/stella-go/stella/generator/parser"
	"github.com/stella-go/stella/version"
)

func GeneratePanic(pkg string, statements []*parser.Statement, banner bool) string {
	importsMap := make(map[string]common.Void)
	importsMap["database/sql"] = common.Null
	functions := make([]string, 0)

	for _, statement := range statements {
		functions = append(functions, "// ==================== "+generator.FirstUpperCamelCase(statement.TableName.Name)+" ====================")
		function, imports := c_panic(statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = u_panic(statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = r_panic(statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		function, imports = d_panic(statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
	}

	importsLines := make([]string, 0)
	for i := range importsMap {
		if i == "" {
			continue
		}
		importsLines = append(importsLines, "\t\""+i+"\"")
	}

	typeLines := `type Service struct {
    DB *sql.DB ` + "`" + `@siu:""` + "`" + `
}`
	bannerS := ""
	if banner {
		bannerS = fmt.Sprintf("\n/**\n * Auto Generate by github.com/stella-go/stella %s on %s.\n */\n", version.VERSION, time.Now().Format("2006/01/02"))

	}
	return fmt.Sprintf("package %s\n%s\nimport (\n%s\n)\n\n%s\n\n%s", pkg, bannerS, strings.Join(importsLines, "\n"), typeLines, strings.Join(functions, "\n"))
}

func c_panic(statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	funcLines := fmt.Sprintf(`func (p *Service) Create%s(s *model.%s) {
    model.Create%s(p.DB, s)
}
`, modelName, modelName, modelName)
	return funcLines, nil
}

func u_panic(statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	primaryKeys := getPrimaryKeyPairs(statement)

	if len(primaryKeys) != 0 {
		fields := make([]string, 0)
		keys := primaryKeys[0]
		for _, k := range keys {
			fields = append(fields, generator.FirstUpperCamelCase(k.ColumnName.Name))
		}
		funcLines := fmt.Sprintf(`func (p *Service) Update%s(s *model.%s) {
    model.Update%sBy%s(p.DB, s)
}
`, modelName, modelName, modelName, strings.Join(fields, ", "))
		return funcLines, nil
	}
	return "", nil
}

func r_panic(statement *parser.Statement) (string, []string) {
	funcLines := ""
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines += fmt.Sprintf(`func (p *Service) Query%s(s *model.%s, page int, size int) (int, []*model.%s) {
    return model.QueryMany%s(p.DB, s, page, size)
}
`, modelName, modelName, modelName, modelName)
	primaryKeyNames := make([]string, 0)
	if len(statement.PrimaryKeyPairs) > 0 {
		keys := statement.PrimaryKeyPairs[0]
		for _, k := range keys {
			primaryKeyNames = append(primaryKeyNames, generator.FirstUpperCamelCase(k.Name))
		}
	}
	if len(primaryKeyNames) > 0 {
		sName := strings.Join(primaryKeyNames, "")
		funcLines += fmt.Sprintf(`func (p *Service) Query%sBy%s(s *model.%s) *model.%s {
    return model.Query%sBy%s(p.DB, s)
}
`, modelName, sName, modelName, modelName, modelName, sName)
	}
	return funcLines, nil
}

func d_panic(statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	primaryKeys := getPrimaryKeyPairs(statement)

	if len(primaryKeys) != 0 {
		fields := make([]string, 0)
		keys := primaryKeys[0]
		for _, k := range keys {
			fields = append(fields, generator.FirstUpperCamelCase(k.ColumnName.Name))
		}
		funcLines := fmt.Sprintf(`func (p *Service) Delete%s(s *model.%s) {
    model.Delete%sBy%s(p.DB, s)
}
`, modelName, modelName, modelName, strings.Join(fields, ", "))
		return funcLines, nil
	}
	return "", nil
}
