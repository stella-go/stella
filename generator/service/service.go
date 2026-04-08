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

	"github.com/stella-go/stella/generator"
	"github.com/stella-go/stella/generator/parser"
)

type ServiceStyle int

const (
	StyleNormal ServiceStyle = iota
	StyleGorm
	StylePanic
)

func Generate(pkg string, filename string, statements []*parser.Statement, banner bool) string {
	return generateService(pkg, filename, statements, banner, StyleNormal)
}

func GenerateGorm(pkg string, filename string, statements []*parser.Statement, banner bool) string {
	return generateService(pkg, filename, statements, banner, StyleGorm)
}

func GeneratePanic(pkg string, filename string, statements []*parser.Statement, banner bool) string {
	return generateService(pkg, filename, statements, banner, StylePanic)
}

func generateService(pkg string, filename string, statements []*parser.Statement, banner bool, style ServiceStyle) string {
	serviceName := ""
	if filename != "service" {
		serviceName = generator.FirstUpperCamelCase(filename)
	}

	var importsMap generator.ImportsSet
	var dbType string
	var dataPrefix string
	switch style {
	case StyleGorm:
		importsMap = generator.NewImportsSet("gorm.io/gorm", "github.com/stella-go/siu/fn/g")
		dbType = "*gorm.DB"
		dataPrefix = "g"
	default:
		importsMap = generator.NewImportsSet("database/sql", "github.com/stella-go/siu/fn/data")
		dbType = "*sql.DB"
		dataPrefix = "data"
	}

	functions := make([]string, 0)
	for _, statement := range statements {
		functions = append(functions, "// ==================== "+generator.FirstUpperCamelCase(statement.TableName.Name)+" ====================")
		function, imports := genCreate(serviceName, statement, style, dataPrefix)
		functions = append(functions, function)
		importsMap.Add(imports...)

		function, imports = genUpdate(serviceName, statement, style, dataPrefix)
		functions = append(functions, function)
		importsMap.Add(imports...)

		function, imports = genQuery(serviceName, statement, style, dataPrefix)
		functions = append(functions, function)
		importsMap.Add(imports...)

		function, imports = genDelete(serviceName, statement, style, dataPrefix)
		functions = append(functions, function)
		importsMap.Add(imports...)
	}

	typeLines := fmt.Sprintf(`type %sService struct {
    DB %s `+"`"+`@siu:""`+"`"+`
}`, serviceName, dbType)

	bannerS := generator.Banner(banner)
	return fmt.Sprintf("package %s\n%s\nimport (\n%s\n)\n\n%s\n\n%s", pkg, bannerS, strings.Join(importsMap.Lines(), "\n"), typeLines, strings.Join(functions, "\n"))
}

func genCreate(serviceName string, statement *parser.Statement, style ServiceStyle, dataPrefix string) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	switch style {
	case StylePanic:
		return fmt.Sprintf(`func (p *%sService) Create%s(s *model.%s) {
    _, err := %s.Create(p.DB, s)
    if err != nil {
        panic(err)
    }
}
`, serviceName, modelName, modelName, dataPrefix), nil
	case StyleGorm:
		return fmt.Sprintf(`func (p *%sService) Create%s(s *model.%s) error {
    return %s.Create(p.DB, s)
}
`, serviceName, modelName, modelName, dataPrefix), nil
	default:
		return fmt.Sprintf(`func (p *%sService) Create%s(s *model.%s) error {
    _, err := %s.Create(p.DB, s)
    return err
}
`, serviceName, modelName, modelName, dataPrefix), nil
	}
}

func genUpdate(serviceName string, statement *parser.Statement, style ServiceStyle, dataPrefix string) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	primaryKeys := parser.GetPrimaryKeyPairs(statement)
	if len(primaryKeys) == 0 {
		return "", nil
	}

	switch style {
	case StylePanic:
		return fmt.Sprintf(`func (p *%sService) Update%s(s *model.%s) {
    _, err := %s.Update(p.DB, s)
    if err != nil {
        panic(err)
    }
}
`, serviceName, modelName, modelName, dataPrefix), nil
	default:
		return fmt.Sprintf(`func (p *%sService) Update%s(s *model.%s) error {
    _, err := %s.Update(p.DB, s)
    return err
}
`, serviceName, modelName, modelName, dataPrefix), nil
	}
}

func genQuery(serviceName string, statement *parser.Statement, style ServiceStyle, dataPrefix string) (string, []string) {
	funcLines := ""
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	switch style {
	case StylePanic:
		funcLines += fmt.Sprintf(`func (p *%sService) QueryMany%s(s *model.%s, page int, size int) (int, []*model.%s) {
    count, many, err := %s.QueryMany(p.DB, s, page, size)
    if err != nil {
        panic(err)
    }
    return count, many
}
`, serviceName, modelName, modelName, modelName, dataPrefix)
	default:
		funcLines += fmt.Sprintf(`func (p *%sService) QueryMany%s(s *model.%s, page int, size int) (int, []*model.%s, error) {
    return %s.QueryMany(p.DB, s, page, size)
}
`, serviceName, modelName, modelName, modelName, dataPrefix)
	}

	primaryKeyNames := make([]string, 0)
	if len(statement.PrimaryKeyPairs) > 0 {
		keys := statement.PrimaryKeyPairs[0]
		for _, k := range keys {
			primaryKeyNames = append(primaryKeyNames, generator.FirstUpperCamelCase(k.Name))
		}
	}
	if len(primaryKeyNames) > 0 {
		switch style {
		case StylePanic:
			funcLines += fmt.Sprintf(`func (p *%sService) Query%s(s *model.%s) *model.%s {
    one, err := %s.Query(p.DB, s)
    if err != nil {
        panic(err)
    }
    return one
}
`, serviceName, modelName, modelName, modelName, dataPrefix)
		default:
			funcLines += fmt.Sprintf(`func (p *%sService) Query%s(s *model.%s) (*model.%s, error) {
    return %s.Query(p.DB, s)
}
`, serviceName, modelName, modelName, modelName, dataPrefix)
		}
	}
	return funcLines, nil
}

func genDelete(serviceName string, statement *parser.Statement, style ServiceStyle, dataPrefix string) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	primaryKeys := parser.GetPrimaryKeyPairs(statement)
	if len(primaryKeys) == 0 {
		return "", nil
	}

	switch style {
	case StylePanic:
		return fmt.Sprintf(`func (p *%sService) Delete%s(s *model.%s) {
    _, err := %s.Delete(p.DB, s)
    if err != nil {
        panic(err)
    }
}
`, serviceName, modelName, modelName, dataPrefix), nil
	default:
		return fmt.Sprintf(`func (p *%sService) Delete%s(s *model.%s) error {
    _, err := %s.Delete(p.DB, s)
    return err
}
`, serviceName, modelName, modelName, dataPrefix), nil
	}
}
