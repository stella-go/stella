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

	"github.com/stella-go/stella/generator"
	"github.com/stella-go/stella/generator/parser"
)

func Generate(pkg string, filename string, statements []*parser.Statement, banner bool) string {
	return generateRouter(pkg, filename, statements, banner, false)
}

func GeneratePanic(pkg string, filename string, statements []*parser.Statement, banner bool) string {
	return generateRouter(pkg, filename, statements, banner, true)
}

func generateRouter(pkg string, filename string, statements []*parser.Statement, banner bool, panicStyle bool) string {
	routerName := ""
	if filename != "router" {
		routerName = generator.FirstUpperCamelCase(filename)
	}

	importsMap := generator.NewImportsSet("github.com/gin-gonic/gin", "github.com/stella-go/siu", "github.com/stella-go/siu/t")
	functions := make([]string, 0)
	routers := make([]string, 0)

	for _, statement := range statements {
		functions = append(functions, "// ==================== "+generator.FirstUpperCamelCase(statement.TableName.Name)+" ====================")
		function, imports, router := c(routerName, statement, panicStyle)
		functions = append(functions, function)
		importsMap.Add(imports...)
		routers = append(routers, router)

		function, imports, router = u(routerName, statement, panicStyle)
		functions = append(functions, function)
		importsMap.Add(imports...)
		routers = append(routers, router)

		function, imports, router = r(routerName, statement, panicStyle)
		functions = append(functions, function)
		importsMap.Add(imports...)
		routers = append(routers, router)

		function, imports, router = d(routerName, statement, panicStyle)
		functions = append(functions, function)
		importsMap.Add(imports...)
		routers = append(routers, router)
	}

	typeLines := `type %sRouter struct {
    Service *service.%sService ` + "`" + `@siu:""` + "`" + `
}

func (p *%sRouter) Router() map[string]gin.HandlerFunc {
    return map[string]gin.HandlerFunc{
%s
    }
}`
	bannerS := generator.Banner(banner)
	return fmt.Sprintf("package %s\n%s\nimport (\n%s\n)\n\n%s\n\n%s", pkg, bannerS, strings.Join(importsMap.Lines(), "\n"), fmt.Sprintf(typeLines, routerName, routerName, routerName, strings.Join(routers, "\n")), strings.Join(functions, "\n"))
}

// sqlTypeToParamType maps SQL types to JSON Schema types for siu.ParamDef.
func sqlTypeToParamType(sqlType string) string {
	switch sqlType {
	case "TINYINT", "INT", "BIGINT":
		return "integer"
	case "FLOAT":
		return "number"
	case "CHAR", "VARCHAR", "TEXT", "DATE", "DATETIME", "TIMESTAMP":
		return "string"
	default:
		return "string"
	}
}

// buildParamDef generates a siu.ParamDef literal string.
func buildParamDef(col *parser.ColumnDefinition, required bool) string {
	paramType := sqlTypeToParamType(col.Type)
	desc := ""
	if col.Comment != nil && col.Comment.Comment != "" {
		desc = col.Comment.Comment
	}
	parts := []string{fmt.Sprintf("Type: %q", paramType)}
	if desc != "" {
		parts = append(parts, fmt.Sprintf("Description: %q", desc))
	}
	if required {
		parts = append(parts, "Required: true")
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// buildMetaRouter returns a router map line wrapping the handler with siu.Meta().
// Parameters reflect the RequestBean[T] structure: timestamp + data{...model fields}.
func buildMetaRouter(route string, handlerExpr string, summary string, columns []*parser.ColumnDefinition, requiredCols []*parser.ColumnDefinition, excludeAutoIncrement bool, extraParams [][2]string) string {
	requiredSet := make(map[string]bool)
	for _, c := range requiredCols {
		requiredSet[c.ColumnName.Name] = true
	}

	dataParams := make([]string, 0)
	for _, col := range columns {
		if excludeAutoIncrement && col.AutoIncrement {
			continue
		}
		jsonName := generator.ToSnakeCase(col.ColumnName.Name)
		dataParams = append(dataParams, fmt.Sprintf("                    %q: %s", jsonName, buildParamDef(col, requiredSet[col.ColumnName.Name])))
	}
	for _, ep := range extraParams {
		dataParams = append(dataParams, fmt.Sprintf("                    %q: {Type: %q}", ep[0], ep[1]))
	}

	return fmt.Sprintf(`        %q: siu.Meta(%s, siu.RouteDef{
            Summary: %q,
            Params: map[string]siu.ParamDef{
                "timestamp": {Type: "integer", Description: "Request timestamp"},
                "data": {Type: "object", Properties: map[string]siu.ParamDef{
%s,
                }},
            },
        }),`, route, handlerExpr, summary, strings.Join(dataParams, ",\n"))
}

func panicRecover() string {
	return `    defer func() {
        if err := recover(); err != nil {
            c.JSON(200, t.FailWith(500, "system error"))
        }
    }()
`
}

func c(routerName string, statement *parser.Statement, panicStyle bool) (string, []string, string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	var funcLines string
	if panicStyle {
		funcLines = fmt.Sprintf(`func (p *%sRouter) Create%s(c *gin.Context) {
%s    request := &t.RequestBean[*model.%s]{}
    err := c.ShouldBind(request)
    t.AssertErrorNil(err)
    s := request.Data
    if s == nil {
        siu.ERROR("__LINE__ bad request: empty data")
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    p.Service.Create%s(s)
    c.JSON(200, t.Success())
}
`, routerName, modelName, panicRecover(), modelName, modelName)
	} else {
		funcLines = fmt.Sprintf(`func (p *%sRouter) Create%s(c *gin.Context) {
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
	}
	route := fmt.Sprintf("POST /api/%s", generator.ToStrikeCase(statement.TableName.Name))
	handler := fmt.Sprintf("p.Create%s", modelName)
	summary := fmt.Sprintf("Create %s", modelName)
	// For create: non-null non-auto-increment columns are required
	requiredCols := make([]*parser.ColumnDefinition, 0)
	for _, col := range statement.Columns {
		if col.AutoIncrement {
			continue
		}
		if col.NotNull && col.DefaultValue == nil && !col.CurrentTimestamp {
			requiredCols = append(requiredCols, col)
		}
	}
	return funcLines, nil, buildMetaRouter(route, handler, summary, statement.Columns, requiredCols, true, nil)
}

func u(routerName string, statement *parser.Statement, panicStyle bool) (string, []string, string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	var funcLines string
	if panicStyle {
		funcLines = fmt.Sprintf(`func (p *%sRouter) Update%s(c *gin.Context) {
%s    request := &t.RequestBean[*model.%s]{}
    err := c.ShouldBind(request)
    t.AssertErrorNil(err)
    s := request.Data
    if s == nil {
        siu.ERROR("__LINE__ bad request: empty data")
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    p.Service.Update%s(s)
    c.JSON(200, t.Success())
}
`, routerName, modelName, panicRecover(), modelName, modelName)
	} else {
		funcLines = fmt.Sprintf(`func (p *%sRouter) Update%s(c *gin.Context) {
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
	}
	route := fmt.Sprintf("PUT /api/%s", generator.ToStrikeCase(statement.TableName.Name))
	handler := fmt.Sprintf("p.Update%s", modelName)
	summary := fmt.Sprintf("Update %s", modelName)
	// For update: primary key columns are required to identify the record
	pkCols := make([]*parser.ColumnDefinition, 0)
	for _, cols := range parser.GetPrimaryKeyPairs(statement) {
		pkCols = append(pkCols, cols...)
	}
	return funcLines, nil, buildMetaRouter(route, handler, summary, statement.Columns, pkCols, false, nil)
}

func r(routerName string, statement *parser.Statement, panicStyle bool) (string, []string, string) {
	funcLines := ""
	routers := make([]string, 0)
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	if panicStyle {
		funcLines += fmt.Sprintf(`func (p *%sRouter) QueryMany%s(c *gin.Context) {
%s    type Pageable struct {
        *model.%s
        Page int `+"`form:\"page\" json:\"page\"`"+`
        Size int `+"`form:\"size\" json:\"size\"`"+`
    }
    request := &t.RequestBean[*Pageable]{}
    err := c.ShouldBind(request)
    t.AssertErrorNil(err)
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
    count, list := p.Service.QueryMany%s(s, page, size)
    c.JSON(200, t.SuccessWith(&PageableResult{Count: count, List: list}))
}
`, routerName, modelName, panicRecover(), modelName, modelName, modelName, modelName, modelName)
	} else {
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
	}
	queryManyRoute := fmt.Sprintf("POST /api/%s/many", generator.ToStrikeCase(statement.TableName.Name))
	queryManyHandler := fmt.Sprintf("p.QueryMany%s", modelName)
	queryManySummary := fmt.Sprintf("Query %s list", modelName)
	extraParams := [][2]string{{"page", "integer"}, {"size", "integer"}}
	routers = append(routers, buildMetaRouter(queryManyRoute, queryManyHandler, queryManySummary, statement.Columns, nil, false, extraParams))

	primaryKeyNames := make([]string, 0)
	if len(statement.PrimaryKeyPairs) > 0 {
		keys := statement.PrimaryKeyPairs[0]
		for _, k := range keys {
			primaryKeyNames = append(primaryKeyNames, generator.FirstUpperCamelCase(k.Name))
		}
	}
	if len(primaryKeyNames) > 0 {
		if panicStyle {
			funcLines += fmt.Sprintf(`func (p *%sRouter) Query%s(c *gin.Context) {
%s    request := &t.RequestBean[*model.%s]{}
    err := c.ShouldBind(request)
    t.AssertErrorNil(err)
    s := request.Data
    if s == nil {
        siu.ERROR("__LINE__ bad request: empty data")
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    one := p.Service.Query%s(s)
    c.JSON(200, t.SuccessWith(one))
}
`, routerName, modelName, panicRecover(), modelName, modelName)
		} else {
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
		}
		queryOneRoute := fmt.Sprintf("POST /api/%s/one", generator.ToStrikeCase(statement.TableName.Name))
		queryOneHandler := fmt.Sprintf("p.Query%s", modelName)
		queryOneSummary := fmt.Sprintf("Query %s by primary key", modelName)
		pkCols := make([]*parser.ColumnDefinition, 0)
		for _, cols := range parser.GetPrimaryKeyPairs(statement) {
			pkCols = append(pkCols, cols...)
		}
		routers = append(routers, buildMetaRouter(queryOneRoute, queryOneHandler, queryOneSummary, pkCols, pkCols, false, nil))
	}
	return funcLines, nil, strings.Join(routers, "\n")
}

func d(routerName string, statement *parser.Statement, panicStyle bool) (string, []string, string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)

	var funcLines string
	if panicStyle {
		funcLines = fmt.Sprintf(`func (p *%sRouter) Delete%s(c *gin.Context) {
%s    request := &t.RequestBean[*model.%s]{}
    err := c.ShouldBind(request)
    t.AssertErrorNil(err)
    s := request.Data
    if s == nil {
        siu.ERROR("__LINE__ bad request: empty data")
        c.JSON(200, t.FailWith(400, "bad request"))
        return
    }
    p.Service.Delete%s(s)
    c.JSON(200, t.Success())
}
`, routerName, modelName, panicRecover(), modelName, modelName)
	} else {
		funcLines = fmt.Sprintf(`func (p *%sRouter) Delete%s(c *gin.Context) {
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
	}
	route := fmt.Sprintf("DELETE /api/%s", generator.ToStrikeCase(statement.TableName.Name))
	handler := fmt.Sprintf("p.Delete%s", modelName)
	summary := fmt.Sprintf("Delete %s", modelName)
	pkCols := make([]*parser.ColumnDefinition, 0)
	for _, cols := range parser.GetPrimaryKeyPairs(statement) {
		pkCols = append(pkCols, cols...)
	}
	return funcLines, nil, buildMetaRouter(route, handler, summary, pkCols, pkCols, false, nil)
}
