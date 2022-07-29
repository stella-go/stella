// Copyright 2010-2022 the original author or authors.

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
	"strings"

	"github.com/stella-go/stella/antlr4/antlr"
	"github.com/stella-go/stella/antlr4/mysql"
)

type MysqlVisitor struct {
	mysql.BaseMySqlParserVisitor
}

func (v *MysqlVisitor) VisitRoot(ctx *mysql.RootContext) interface{} {
	return ctx.SqlStatements().Accept(v)
}

func (v *MysqlVisitor) VisitSqlStatements(ctx *mysql.SqlStatementsContext) interface{} {
	statements := make([]*Statement, 0)
	for _, sqlstatement := range ctx.AllSqlStatement() {
		obj := sqlstatement.Accept(v)
		statement, ok := obj.(*Statement)
		if !ok {
			continue
		}
		statements = append(statements, statement)
	}
	return statements
}

func (v *MysqlVisitor) VisitSqlStatement(ctx *mysql.SqlStatementContext) interface{} {
	if c := ctx.DdlStatement(); c != nil {
		return c.Accept(v)
	}
	return nil
}

func (v *MysqlVisitor) VisitDdlStatement(ctx *mysql.DdlStatementContext) interface{} {
	if c := ctx.CreateTable(); c != nil {
		return c.Accept(v)
	}
	return nil
}

func (v *MysqlVisitor) VisitColumnCreateTable(ctx *mysql.ColumnCreateTableContext) interface{} {
	statement := ctx.CreateDefinitions().Accept(v).(*Statement)
	statement.TableName = ctx.TableName().Accept(v).(*NameDefinition)
	return statement
}

func (v *MysqlVisitor) VisitTableName(ctx *mysql.TableNameContext) interface{} {
	return &NameDefinition{name: ctx.GetText(), start: ctx.GetStart().GetStart(), stop: ctx.GetStop().GetStop()}
}

func (v *MysqlVisitor) VisitCreateDefinitions(ctx *mysql.CreateDefinitionsContext) interface{} {
	statement := &Statement{
		Columns:         make([]*ColumnDefinition, 0),
		PrimaryKeyPairs: make([]PrimaryKeyPair, 0),
		UniqKeyPairs:    make([]UniqueKeyPair, 0),
		IndexKeyPairs:   make([]IndexKeyPair, 0),
	}
	definitions := ctx.AllCreateDefinition()
	for _, definition := range definitions {
		obj := definition.Accept(v)
		switch obj := obj.(type) {
		case *ColumnDefinition:
			statement.Columns = append(statement.Columns, obj)
		case PrimaryKeyPair:
			statement.PrimaryKeyPairs = append(statement.PrimaryKeyPairs, obj)
		case UniqueKeyPair:
			statement.UniqKeyPairs = append(statement.UniqKeyPairs, obj)
		case IndexKeyPair:
			statement.IndexKeyPairs = append(statement.IndexKeyPairs, obj)
		}
	}
	return statement
}

func (v *MysqlVisitor) VisitColumnDeclaration(ctx *mysql.ColumnDeclarationContext) interface{} {
	uid := ctx.Uid()
	columnName := &NameDefinition{name: uid.GetText(), start: uid.GetStart().GetStart(), stop: uid.GetStop().GetStop()}

	column := ctx.ColumnDefinition().Accept(v).(*ColumnDefinition)
	column.ColumnName = columnName
	return column
}

func (v *MysqlVisitor) VisitColumnDefinition(ctx *mysql.ColumnDefinitionContext) interface{} {
	column := &ColumnDefinition{Type: ctx.DataType().Accept(v).(string)}
	for _, constraint := range ctx.AllColumnConstraint() {
		obj := constraint.Accept(v)
		switch obj.(type) {
		case *AutoIncrement:
			column.AutoIncrement = true
		case *PrimaryKey:
			column.PrimaryKey = true
		case *UniqKey:
			column.UniqueKey = true
		case *DefaultValue:
			column.DefaultValue = true
		case *CurrentTimestamp:
			column.DefaultValue = true
			column.CurrentTimestamp = true
		case *NotNull:
			column.NotNull = true
		}
	}
	return column
}
func (v *MysqlVisitor) VisitStringDataType(ctx *mysql.StringDataTypeContext) interface{} {
	return ctx.GetTypeName().GetText()
}

func (v *MysqlVisitor) VisitNationalStringDataType(ctx *mysql.NationalStringDataTypeContext) interface{} {
	return ctx.GetTypeName().GetText()
}

func (v *MysqlVisitor) VisitNationalVaryingStringDataType(ctx *mysql.NationalVaryingStringDataTypeContext) interface{} {
	return ctx.GetTypeName().GetText()
}

func (v *MysqlVisitor) VisitDimensionDataType(ctx *mysql.DimensionDataTypeContext) interface{} {
	return ctx.GetTypeName().GetText()
}

func (v *MysqlVisitor) VisitSimpleDataType(ctx *mysql.SimpleDataTypeContext) interface{} {
	return ctx.GetTypeName().GetText()
}

func (v *MysqlVisitor) VisitCollectionDataType(ctx *mysql.CollectionDataTypeContext) interface{} {
	return ctx.GetTypeName().GetText()
}

func (v *MysqlVisitor) VisitSpatialDataType(ctx *mysql.SpatialDataTypeContext) interface{} {
	return ctx.GetTypeName().GetText()
}

func (v *MysqlVisitor) VisitLongVarcharDataType(ctx *mysql.LongVarcharDataTypeContext) interface{} {
	return ctx.GetTypeName().GetText()
}

func (v *MysqlVisitor) VisitLongVarbinaryDataType(ctx *mysql.LongVarbinaryDataTypeContext) interface{} {
	return "LONG VARBINARY"
}

func (v *MysqlVisitor) VisitAutoIncrementColumnConstraint(ctx *mysql.AutoIncrementColumnConstraintContext) interface{} {
	return &AutoIncrement{}
}

func (v *MysqlVisitor) VisitPrimaryKeyColumnConstraint(ctx *mysql.PrimaryKeyColumnConstraintContext) interface{} {
	return &PrimaryKey{}
}

func (v *MysqlVisitor) VisitUniqueKeyColumnConstraint(ctx *mysql.UniqueKeyColumnConstraintContext) interface{} {
	return &UniqKey{}
}

func (v *MysqlVisitor) VisitDefaultColumnConstraint(ctx *mysql.DefaultColumnConstraintContext) interface{} {
	if c := ctx.DefaultValue().Accept(v); c != nil {
		return c
	} else {
		return &DefaultValue{}
	}
}

func (v *MysqlVisitor) VisitDefaultValue(ctx *mysql.DefaultValueContext) interface{} {
	for _, current := range ctx.AllCurrentTimestamp() {
		if obj := current.Accept(v); obj != nil {
			return obj
		}
	}
	return nil
}

func (v *MysqlVisitor) VisitNullColumnConstraint(ctx *mysql.NullColumnConstraintContext) interface{} {
	return ctx.NullNotnull().Accept(v)
}

func (v *MysqlVisitor) VisitNullNotnull(ctx *mysql.NullNotnullContext) interface{} {
	if !ctx.IsEmpty() && ctx.GetText() == "NOTNULL" {
		return &NotNull{}
	}
	return nil
}

func (v *MysqlVisitor) VisitCurrentTimestamp(ctx *mysql.CurrentTimestampContext) interface{} {
	return &CurrentTimestamp{}
}

func (v *MysqlVisitor) VisitConstraintDeclaration(ctx *mysql.ConstraintDeclarationContext) interface{} {
	return ctx.TableConstraint().Accept(v)
}

func (v *MysqlVisitor) VisitPrimaryKeyTableConstraint(ctx *mysql.PrimaryKeyTableConstraintContext) interface{} {
	names := ctx.IndexColumnNames().Accept(v).([]*NameDefinition)
	return PrimaryKeyPair(names)
}

func (v *MysqlVisitor) VisitUniqueKeyTableConstraint(ctx *mysql.UniqueKeyTableConstraintContext) interface{} {
	names := ctx.IndexColumnNames().Accept(v).([]*NameDefinition)
	return UniqueKeyPair(names)
}

func (v *MysqlVisitor) VisitIndexDeclaration(ctx *mysql.IndexDeclarationContext) interface{} {
	return ctx.IndexColumnDefinition().Accept(v)
}

func (v *MysqlVisitor) VisitSimpleIndexDeclaration(ctx *mysql.SimpleIndexDeclarationContext) interface{} {
	names := ctx.IndexColumnNames().Accept(v).([]*NameDefinition)
	return IndexKeyPair(names)
}

func (v *MysqlVisitor) VisitSpecialIndexDeclaration(ctx *mysql.SpecialIndexDeclarationContext) interface{} {
	names := ctx.IndexColumnNames().Accept(v).([]*NameDefinition)
	return IndexKeyPair(names)
}

func (v *MysqlVisitor) VisitIndexColumnNames(ctx *mysql.IndexColumnNamesContext) interface{} {
	names := make([]*NameDefinition, 0)
	for i := range ctx.AllIndexColumnName() {
		name := ctx.IndexColumnName(i).Accept(v)
		names = append(names, name.(*NameDefinition))
	}
	return names
}

func (v *MysqlVisitor) VisitIndexColumnName(ctx *mysql.IndexColumnNameContext) interface{} {
	return &NameDefinition{name: ctx.GetText(), start: ctx.GetStart().GetStart(), stop: ctx.GetStop().GetStop()}
}

func Parse(sql string) []*Statement {
	SQL := strings.ToUpper(sql)
	is := antlr.NewInputStream(SQL)
	lexer := mysql.NewMySqlLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := mysql.NewMySqlParser(stream)
	obj := parser.Root().Accept(&MysqlVisitor{})
	statements, ok := obj.([]*Statement)
	if !ok {
		fmt.Println("SQL has error: ", sql)
		return nil
	}
	for _, statement := range statements {
		statement.Fill(sql)
	}
	return statements
}
