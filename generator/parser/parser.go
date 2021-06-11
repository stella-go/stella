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

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/stella-go/stella/generator/parser/mysql"
)

type MysqlVisitor struct {
	mysql.BaseMySqlParserVisitor
}

// visit create table sql
func (v *MysqlVisitor) VisitColumnCreateTable(ctx *mysql.ColumnCreateTableContext) interface{} {
	tableName := ctx.TableName()
	upperName := tableName.Accept(v)
	start := tableName.GetStart().GetStart()
	stop := tableName.GetStop().GetStop()
	properties := ctx.CreateDefinitions().Accept(v)
	return &Statement{upperName: upperName.(string), start: start, stop: stop, Properties: properties.([]*Property)}
}

// Visit table name
func (v *MysqlVisitor) VisitTableName(ctx *mysql.TableNameContext) interface{} {
	return ctx.GetText()
}

// Visit every line in create definitions
func (v *MysqlVisitor) VisitCreateDefinitions(ctx *mysql.CreateDefinitionsContext) interface{} {
	properties := make([]*Property, 0)
	var primaryNames []string
	var uniqueNames []string
	for i := range ctx.AllCreateDefinition() {
		definition := ctx.CreateDefinition(i)
		p := definition.Accept(v)
		switch p := p.(type) {
		case *Property:
			properties = append(properties, p)
		case PrimaryKey:
			primaryNames = []string(p)
		case UniqKey:
			uniqueNames = []string(p)
		}
	}
	for i := range properties {
		for j := range primaryNames {
			if strings.Trim(properties[i].upperName, "`") == strings.Trim(primaryNames[j], "`") {
				properties[i].Primary = true
			}
		}
		for j := range uniqueNames {
			if strings.Trim(properties[i].upperName, "`") == strings.Trim(uniqueNames[j], "`") {
				properties[i].Uniq = true
			}
		}
	}
	return properties
}

// Visit column name and data type
func (v *MysqlVisitor) VisitColumnDeclaration(ctx *mysql.ColumnDeclarationContext) interface{} {
	uid := ctx.Uid()
	upperName := uid.GetText()
	start := uid.GetStart().GetStart()
	stop := uid.GetStop().GetStop()

	columsType := ctx.ColumnDefinition().Accept(v).(*ColumnType)
	property := &Property{upperName: upperName, start: start, stop: stop, DataBaseType: columsType.typeName, Primary: columsType.Primary, Uniq: columsType.Uniq}
	return property
}

// Visit dataType and unique(primary)
func (v *MysqlVisitor) VisitColumnDefinition(ctx *mysql.ColumnDefinitionContext) interface{} {
	typeName := ctx.DataType().GetText()
	index := strings.Index(typeName, "(")
	if index != -1 {
		typeName = typeName[0:index]
	}
	primary := false
	uniq := false
	for i := range ctx.AllColumnConstraint() {
		p := ctx.ColumnConstraint(i).Accept(v)
		switch p.(type) {
		case *PrimaryKey:
			primary = true
		case *UniqKey:
			uniq = true
		}
	}
	return &ColumnType{Primary: primary, Uniq: uniq, typeName: typeName}
}

func (v *MysqlVisitor) VisitPrimaryKeyColumnConstraint(ctx *mysql.PrimaryKeyColumnConstraintContext) interface{} {
	return &PrimaryKey{}
}

func (v *MysqlVisitor) VisitUniqueKeyColumnConstraint(ctx *mysql.UniqueKeyColumnConstraintContext) interface{} {
	return &UniqKey{}
}

// Visit constraint declaration
func (v *MysqlVisitor) VisitConstraintDeclaration(ctx *mysql.ConstraintDeclarationContext) interface{} {
	return ctx.TableConstraint().Accept(v)
}

// Visit primary key table constraint
func (v *MysqlVisitor) VisitPrimaryKeyTableConstraint(ctx *mysql.PrimaryKeyTableConstraintContext) interface{} {
	names := ctx.IndexColumnNames().Accept(v)
	return PrimaryKey(names.([]string))
}

// Visit unique key table constraint
func (v *MysqlVisitor) VisitUniqueKeyTableConstraint(ctx *mysql.UniqueKeyTableConstraintContext) interface{} {
	names := ctx.IndexColumnNames().Accept(v)
	return UniqKey(names.([]string))
}

// Visit index column names
func (v *MysqlVisitor) VisitIndexColumnNames(ctx *mysql.IndexColumnNamesContext) interface{} {
	names := make([]string, 0)
	for i := range ctx.AllIndexColumnName() {
		name := ctx.IndexColumnName(i).Accept(v)
		names = append(names, name.(string))
	}
	return names
}

// Visit index column name
func (v *MysqlVisitor) VisitIndexColumnName(ctx *mysql.IndexColumnNameContext) interface{} {
	return ctx.GetText()
}

func Parse(sql string) []*Statement {
	sqls := SplitSQL(sql)
	statements := make([]*Statement, 0)
	re := regexp.MustCompile(`CREATE +TABLE.*`)
	for _, s := range sqls {
		SQL := strings.ToUpper(s)
		trimSQL := strings.TrimSpace(SQL)
		if re.MatchString(trimSQL) {
			is := antlr.NewInputStream(SQL)
			lexer := mysql.NewMySqlLexer(is)
			stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
			parser := mysql.NewMySqlParser(stream)
			obj := (parser.CreateTable().Accept(&MysqlVisitor{}))
			statement, ok := obj.(*Statement)
			if !ok {
				fmt.Println("SQL has error: ", SQL)
				continue
			}
			statement.Fill(s)
			statements = append(statements, statement)
		}
	}
	return statements
}

func SplitSQL(sql string) []string {
	sql = strings.ReplaceAll(sql, "\r", "\n")

	// process /* */ multi-comment
	if index := strings.Index(sql, "/*"); index != -1 {
		re := regexp.MustCompile(`/\*(?s).*?\*/`)
		sql = re.ReplaceAllString(sql, "")
	}

	// process -- comment
	if index := strings.Index(sql, "-- "); index != -1 {
		re := regexp.MustCompile("-- .*\n")
		sql = re.ReplaceAllString(sql, "\n")
	}
	// process # comment
	if index := strings.Index(sql, "#"); index != -1 {
		re := regexp.MustCompile("#.*\n")
		sql = re.ReplaceAllString(sql, "\n")
	}

	sharp := 0
	backtick := 0
	quotes := 0
	doubleQuotes := 0
	points := make([]int, 0)
	for i := range sql {
		char := sql[i : i+1]
		switch char {
		case "#":
			if backtick&1 == 0 && quotes&1 == 0 && doubleQuotes&1 == 0 {
				sharp++
			}
		case "`":
			if sharp&1 == 0 && quotes&1 == 0 && doubleQuotes&1 == 0 {
				backtick++
			}
		case "'":
			if sharp&1 == 0 && backtick&1 == 0 && doubleQuotes&1 == 0 {
				quotes++
			}
		case "\"":
			if sharp&1 == 0 && backtick&1 == 0 && quotes&1 == 0 {
				doubleQuotes++
			}
		case "\n":
			if sharp&1 != 0 {
				sharp++
			}
		case ";":
			if sharp&1 == 0 && backtick&1 == 0 && quotes&1 == 0 && doubleQuotes&1 == 0 {
				points = append(points, i)
			}
		}
	}
	splits := make([]string, 0)
	if len(points) != 0 {
		for i := range points {
			if i == 0 {
				splits = append(splits, sql[:points[i]+1])
			} else {
				splits = append(splits, sql[points[i-1]+1:points[i]+1])
			}
		}
		splits = append(splits, sql[points[len(points)-1]+1:])
	} else {
		splits = append(splits, sql)
	}
	return splits
}
