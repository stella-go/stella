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

package router

import (
	"fmt"
	"strings"
	"time"

	"github.com/stella-go/stella/common"
	"github.com/stella-go/stella/generator"
	"github.com/stella-go/stella/generator/parser"
	"github.com/stella-go/stella/version"
)

func Generate(pkg string, filename string, statements []*parser.Statement, banner bool) string {
	routerName := ""
	if filename != "router" {
		routerName = generator.FirstUpperCamelCase(filename)
	}

	importsMap := make(map[string]common.Void)
	functions := make([]string, 0)
	routers := make([]string, 0)

	importsMap["github.com/gin-gonic/gin"] = common.Null
	importsMap["github.com/stella-go/siu"] = common.Null
	importsMap["github.com/stella-go/siu/t"] = common.Null
	for _, statement := range statements {
		functions = append(functions, "// ==================== "+generator.FirstUpperCamelCase(statement.TableName.Name)+" ====================")
		function, imports, router := c(routerName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		routers = append(routers, router)

		function, imports, router = u(routerName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		routers = append(routers, router)

		function, imports, router = r(routerName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		routers = append(routers, router)
		function, imports, router = d(routerName, statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		routers = append(routers, router)
	}

	importsLines := make([]string, 0)
	for i := range importsMap {
		if i == "" {
			continue
		}
		importsLines = append(importsLines, "\t\""+i+"\"")
	}

	typeLines := `type %sRouter struct {
    Service *service.%sService ` + "`" + `@siu:""` + "`" + `
}

func (p *%sRouter) Router() map[string]gin.HandlerFunc {
    return map[string]gin.HandlerFunc{
%s
    }
}`
	bannerS := ""
	if banner {
		bannerS = fmt.Sprintf("\n/**\n * Auto Generate by github.com/stella-go/stella %s on %s.\n */\n", version.VERSION, time.Now().Format("2006/01/02"))
	}
	return fmt.Sprintf("package %s\n%s\nimport (\n%s\n)\n\n%s\n\n%s", pkg, bannerS, strings.Join(importsLines, "\n"), fmt.Sprintf(typeLines, routerName, routerName, routerName, strings.Join(routers, "\n")), strings.Join(functions, "\n"))
}

func c(routerName string, statement *parser.Statement) (string, []string, string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	funcLines := fmt.Sprintf(`func (p *%sRouter) Create%s(c *gin.Context) {
    request := &t.RequestBean[*model.%s]{}
    err := c.ShouldBind(request)
    if err != nil {
        siu.ERROR("__LINE__ bad request:", err)
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    s := request.Data
    if s == nil {
        siu.ERROR("__LINE__ bad request: empty data")
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    err = p.Service.Create%s(s)
    if err != nil {
        siu.ERROR("__LINE__ create %s error:", err)
        c.JSON(200, t.FailWith(500, "system error"))
    } else {
        c.JSON(200, t.Success())
    }
}
`, routerName, modelName, modelName, modelName, modelName)
	return funcLines, nil, fmt.Sprintf(`        "POST /api/%s": p.Create%s,`, generator.ToStrikeCase(statement.TableName.Name), modelName)
}

func u(routerName string, statement *parser.Statement) (string, []string, string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	funcLines := fmt.Sprintf(`func (p *%sRouter) Update%s(c *gin.Context) {
    request := &t.RequestBean[*model.%s]{}
    err := c.ShouldBind(request)
    if err != nil {
        siu.ERROR("__LINE__ bad request:", err)
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    s := request.Data
    if s == nil {
        siu.ERROR("__LINE__ bad request: empty data")
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    err = p.Service.Update%s(s)
    if err != nil {
        siu.ERROR("__LINE__ update %s error:", err)
        c.JSON(200, t.FailWith(500, "system error"))
    } else {
        c.JSON(200, t.Success())
    }
}
`, routerName, modelName, modelName, modelName, modelName)
	return funcLines, nil, fmt.Sprintf(`        "PUT /api/%s": p.Update%s,`, generator.ToStrikeCase(statement.TableName.Name), modelName)
}

func r(routerName string, statement *parser.Statement) (string, []string, string) {
	funcLines := ""
	routers := make([]string, 0)
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	funcLines += fmt.Sprintf(`func (p *%sRouter) QueryMany%s(c *gin.Context) {
    type Pageable struct {
        *model.%s
        Page int `+"`form:\"page\" json:\"page\"`"+`
        Size int `+"`form:\"size\" json:\"size\"`"+`
    }
    request := &t.RequestBean[*Pageable]{}
    err := c.ShouldBind(request)
    if err != nil {
        siu.ERROR("__LINE__ bad request:", err)
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    data := request.Data
    var s *model.%s
    var page, size int
    if data != nil {
        s = data.%s
        page = data.Page
        size = data.Size
    }
    if page <= 0 {
        page = 1
    }
    if size <= 0 {
        size = 10
    }
    type PageableResult struct {
        Count int `+"`json:\"count\"`"+`
        List []*model.%s `+"`json:\"list\"`"+`
    }
    count, list, err := p.Service.QueryMany%s(s, page, size)
    if err != nil {
        siu.ERROR("__LINE__ query %s error:", err)
        c.JSON(200, t.FailWith(500, "system error"))
    } else {
        c.JSON(200, t.SuccessWith(&PageableResult{Count: count, List: list}))
    }
}
`, routerName, modelName, modelName, modelName, modelName, modelName, modelName, modelName)
	routers = append(routers, fmt.Sprintf(`        "POST /api/%s/many": p.QueryMany%s,`, generator.ToStrikeCase(statement.TableName.Name), modelName))
	primaryKeyNames := make([]string, 0)
	if len(statement.PrimaryKeyPairs) > 0 {
		keys := statement.PrimaryKeyPairs[0]
		for _, k := range keys {
			primaryKeyNames = append(primaryKeyNames, generator.FirstUpperCamelCase(k.Name))
		}
	}
	if len(primaryKeyNames) > 0 {
		funcLines += fmt.Sprintf(`func (p *%sRouter) Query%s(c *gin.Context) {
    request := &t.RequestBean[*model.%s]{}
    err := c.ShouldBind(request)
    if err != nil {
        siu.ERROR("__LINE__ bad request:", err)
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    s := request.Data
    if s == nil {
        siu.ERROR("__LINE__ bad request: empty data")
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    one, err := p.Service.Query%s(s)
    if err != nil {
        siu.ERROR("__LINE__ query %s error:", err)
        c.JSON(200, t.FailWith(500, "system error"))
    } else {
        c.JSON(200, t.SuccessWith(one))
    }
}
`, routerName, modelName, modelName, modelName, modelName)
		routers = append(routers, fmt.Sprintf(`        "POST /api/%s/one": p.Query%s,`, generator.ToStrikeCase(statement.TableName.Name), modelName))
	}
	return funcLines, nil, strings.Join(routers, "\n")
}

func d(routerName string, statement *parser.Statement) (string, []string, string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	funcLines := fmt.Sprintf(`func (p *%sRouter) Delete%s(c *gin.Context) {
    request := &t.RequestBean[*model.%s]{}
    err := c.ShouldBind(request)
    if err != nil {
        siu.ERROR("__LINE__ bad request:", err)
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    s := request.Data
    if s == nil {
        siu.ERROR("__LINE__ bad request: empty data")
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    err = p.Service.Delete%s(s)
    if err != nil {
        siu.ERROR("__LINE__ delete %s error:", err)
        c.JSON(200, t.FailWith(500, "system error"))
    } else {
        c.JSON(200, t.Success())
    }
}
`, routerName, modelName, modelName, modelName, modelName)
	return funcLines, nil, fmt.Sprintf(`        "DELETE /api/%s": p.Delete%s,`, generator.ToStrikeCase(statement.TableName.Name), modelName)
}
