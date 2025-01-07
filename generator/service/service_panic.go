// Copyright 2010-2024 the original author or authors.

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

func GeneratePanic(pkg string, filename string, statements []*parser.Statement, banner bool) string {
	serviceName := ""
	if filename != "service" {
		serviceName = generator.FirstUpperCamelCase(filename)
	}

	importsMap := make(map[string]common.Void)
	importsMap["database/sql"] = common.Null
	importsMap["github.com/stella-go/siu/fn/data"] = common.Null
	functions := make([]string, 0)

	for _, statement := range statements {
		functions = append(functions, "// ==================== "+generator.FirstUpperCamelCase(statement.TableName.Name)+" ====================")
		function, imports := c_panic(serviceName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = u_panic(serviceName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = r_panic(serviceName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		function, imports = d_panic(serviceName, statement)
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

	typeLines := `type %sService struct {
    DB *sql.DB ` + "`" + `@siu:""` + "`" + `
}`
	bannerS := ""
	if banner {
		bannerS = fmt.Sprintf("\n/**\n * Auto Generate by github.com/stella-go/stella %s on %s.\n */\n", version.VERSION, time.Now().Format("2006/01/02"))

	}
	return fmt.Sprintf("package %s\n%s\nimport (\n%s\n)\n\n%s\n\n%s", pkg, bannerS, strings.Join(importsLines, "\n"), fmt.Sprintf(typeLines, serviceName), strings.Join(functions, "\n"))
}

func c_panic(serviceName string, statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	funcLines := fmt.Sprintf(`func (p *%sService) Create%s(s *model.%s) {
    _, err := data.Create(p.DB, s)
    if err != nil {
        panic(err)
    }
}
`, serviceName, modelName, modelName)
	return funcLines, nil
}

func u_panic(serviceName string, statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	primaryKeys := getPrimaryKeyPairs(statement)

	if len(primaryKeys) != 0 {
		funcLines := fmt.Sprintf(`func (p *%sService) Update%s(s *model.%s) {
    _, err := data.Update(p.DB, s)
    if err != nil {
        panic(err)
    }
}
`, serviceName, modelName, modelName)
		return funcLines, nil
	}
	return "", nil
}

func r_panic(serviceName string, statement *parser.Statement) (string, []string) {
	funcLines := ""
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines += fmt.Sprintf(`func (p *%sService) QueryMany%s(s *model.%s, page int, size int) (int, []*model.%s) {
    count, many, err := data.QueryMany(p.DB, s, page, size)
    if err != nil {
        panic(err)
    }
    return count, many
}
`, serviceName, modelName, modelName, modelName)
	primaryKeyNames := make([]string, 0)
	if len(statement.PrimaryKeyPairs) > 0 {
		keys := statement.PrimaryKeyPairs[0]
		for _, k := range keys {
			primaryKeyNames = append(primaryKeyNames, generator.FirstUpperCamelCase(k.Name))
		}
	}
	if len(primaryKeyNames) > 0 {
		funcLines += fmt.Sprintf(`func (p *%sService) Query%s(s *model.%s) *model.%s {
    one, err := data.Query(p.DB, s)
    if err != nil {
        panic(err)
    }
    return one
}
`, serviceName, modelName, modelName, modelName)
	}
	return funcLines, nil
}

func d_panic(serviceName string, statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	primaryKeys := getPrimaryKeyPairs(statement)

	if len(primaryKeys) != 0 {
		funcLines := fmt.Sprintf(`func (p *%sService) Delete%s(s *model.%s) {
    _, err := data.Delete(p.DB, s)
    if err != nil {
        panic(err)
    }
}
`, serviceName, modelName, modelName)
		return funcLines, nil
	}
	return "", nil
}
