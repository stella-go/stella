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

func GenerateGorm(pkg string, filename string, statements []*parser.Statement, banner bool) string {
	serviceName := ""
	if filename != "service" {
		serviceName = generator.FirstUpperCamelCase(filename)
	}

	importsMap := make(map[string]common.Void)
	importsMap["errors"] = common.Null
	importsMap["gorm.io/gorm"] = common.Null
	functions := make([]string, 0)

	for _, statement := range statements {
		functions = append(functions, "// ==================== "+generator.FirstUpperCamelCase(statement.TableName.Name)+" ====================")
		function, imports := c_gorm(serviceName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = u_gorm(serviceName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = r_gorm(serviceName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		function, imports = d_gorm(serviceName, statement)
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
    DB *gorm.DB ` + "`" + `@siu:""` + "`" + `
}`
	bannerS := ""
	if banner {
		bannerS = fmt.Sprintf("\n/**\n * Auto Generate by github.com/stella-go/stella %s on %s.\n */\n", version.VERSION, time.Now().Format("2006/01/02"))

	}
	return fmt.Sprintf("package %s\n%s\nimport (\n%s\n)\n\n%s\n\n%s", pkg, bannerS, strings.Join(importsLines, "\n"), fmt.Sprintf(typeLines, serviceName), strings.Join(functions, "\n"))
}

func c_gorm(serviceName string, statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	funcLines := fmt.Sprintf(`func (p *%sService) Create%s(s *model.%s) error {
    r := p.DB.Model(s).Create(s)
    return r.Error
}
`, serviceName, modelName, modelName)
	return funcLines, nil
}

func u_gorm(serviceName string, statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	primaryKeys := getPrimaryKeyPairs(statement)

	if len(primaryKeys) != 0 {
		funcLines := fmt.Sprintf(`func (p *%sService) Update%s(s *model.%s) error {
    r := p.DB.Model(s).Updates(s)
    return r.Error
}
`, serviceName, modelName, modelName)
		return funcLines, nil
	}
	return "", nil
}

func r_gorm(serviceName string, statement *parser.Statement) (string, []string) {
	funcLines := ""
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines += fmt.Sprintf(`func (p *%sService) QueryMany%s(s *model.%s, page int, size int) (int, []*model.%s, error) {
    stmt := p.DB.Model(s).Where(s)
    var count int64
    many := make([]*model.%s, 0)
    r := stmt.Count(&count)
    if r.Error != nil {
        if errors.Is(r.Error, gorm.ErrRecordNotFound) {
            return 0, many, nil
        } else {
            return 0, nil, r.Error
        }
    }
    r = stmt.Offset((page - 1) * size).Limit(size).Find(&many)
    if r.Error != nil {
        if errors.Is(r.Error, gorm.ErrRecordNotFound) {
            return 0, many, nil
        } else {
            return 0, nil, r.Error
        }
    }
    return int(count), many, nil
}
`, serviceName, modelName, modelName, modelName, modelName)
	primaryKeyNames := make([]string, 0)
	if len(statement.PrimaryKeyPairs) > 0 {
		keys := statement.PrimaryKeyPairs[0]
		for _, k := range keys {
			primaryKeyNames = append(primaryKeyNames, generator.FirstUpperCamelCase(k.Name))
		}
	}
	if len(primaryKeyNames) > 0 {
		funcLines += fmt.Sprintf(`func (p *%sService) Query%s(s *model.%s) (*model.%s, error) {
    ss := &model.%s{}
    r := p.DB.Model(s).Where(s).Take(&ss)
    if r.Error != nil {
        if errors.Is(r.Error, gorm.ErrRecordNotFound) {
            return nil, nil
        } else {
            return nil, r.Error
        }
    }
    return ss, nil
}
`, serviceName, modelName, modelName, modelName, modelName)
	}
	return funcLines, nil
}

func d_gorm(serviceName string, statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	primaryKeys := getPrimaryKeyPairs(statement)

	if len(primaryKeys) != 0 {
		funcLines := fmt.Sprintf(`func (p *%sService) Delete%s(s *model.%s) error {
    r := p.DB.Model(s).Delete(s, s)
    return r.Error
}
`, serviceName, modelName, modelName)
		return funcLines, nil
	}
	return "", nil
}
