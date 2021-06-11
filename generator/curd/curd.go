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

package curd

import (
	"fmt"
	"strings"
	"time"

	"github.com/stella-go/stella/common"
	"github.com/stella-go/stella/generator/parser"
)

func Generate(pkg string, statements []*parser.Statement) string {
	importsMap := make(map[string]common.Void)
	importsMap["database/sql"] = common.Null
	functions := make([]string, 0)
	for _, statement := range statements {
		functions = append(functions, "// ==================== "+statement.ModelName+" ====================")
		function, imports := c(statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = u(statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = r(statement)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		function, imports = d(statement)
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
	return fmt.Sprintf("package %s\n\n/**\n * Auto Generate by github.com/stella-go/stella on %s.\n */\n\nimport (\n%s\n)\n\n%s", pkg, time.Now().Format("2006/01/02"), strings.Join(importsLines, "\n"), strings.Join(functions, "\n\n"))
}

func c(statement *parser.Statement) (string, []string) {
	columnNames := make([]string, 0)
	placeHolder := make([]string, 0)
	args := make([]string, 0)
	for _, prop := range statement.Properties {
		if prop.Primary || prop.DataBaseType == "DATETIME" || prop.DataBaseType == "TIMESTAMP" {
			continue
		}
		columnNames = append(columnNames, "`"+prop.ColumnName+"`")
		placeHolder = append(placeHolder, "?")
		args = append(args, "s."+prop.PropertyName)
	}
	SQL := fmt.Sprintf("insert into `%s` (%s) values (%s)", statement.TableName, strings.Join(columnNames, ", "), strings.Join(placeHolder, ", "))
	funcLines := fmt.Sprintf(`func Create%s (db *sql.DB, s *%s) error {
    SQL := "%s"
    ret, err := db.Exec(SQL, %s)
    if err != nil {
    	return err
    }
    _, err = ret.RowsAffected()
    if err != nil {
    	return err
    }
    return nil
}`, statement.ModelName, statement.ModelName, SQL, strings.Join(args, ", "))
	return funcLines, nil
}

func u(statement *parser.Statement) (string, []string) {
	valuePairs := make([]string, 0)
	args := make([]string, 0)
	keys := make([]*parser.Property, 0)
	for _, prop := range statement.Properties {
		if prop.DataBaseType == "DATETIME" || prop.DataBaseType == "TIMESTAMP" {
			continue
		}

		if prop.Primary || prop.Uniq {
			keys = append(keys, prop)
			continue
		}

		valuePairs = append(valuePairs, "`"+prop.ColumnName+"` = ?")
		args = append(args, "s."+prop.PropertyName)
	}

	keyPairs := make([]string, 0)
	for _, prop := range keys {
		keyPairs = append(keyPairs, "`"+prop.ColumnName+"` = ?")
		args = append(args, "s."+prop.PropertyName)
	}

	SQL := fmt.Sprintf("update `%s` set %s where %s", statement.TableName, strings.Join(valuePairs, ", "), strings.Join(keyPairs, " and "))
	funcLines := fmt.Sprintf(`func Update%s (db *sql.DB, s *%s) error {
	SQL := "%s"
	ret, err := db.Exec(SQL, %s)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}`, statement.ModelName, statement.ModelName, SQL, strings.Join(args, ", "))
	return funcLines, nil
}

func r(statement *parser.Statement) (string, []string) {
	columnNames := make([]string, 0)
	binds := make([]string, 0)
	args := make([]string, 0)
	keys := make([]*parser.Property, 0)
	for _, prop := range statement.Properties {
		columnNames = append(columnNames, "`"+prop.ColumnName+"`")
		binds = append(binds, "&ret."+prop.PropertyName)
		if prop.Primary || prop.Uniq {
			keys = append(keys, prop)
		}
	}

	keyPairs := make([]string, 0)
	for _, prop := range keys {
		keyPairs = append(keyPairs, "`"+prop.ColumnName+"` = ?")
		args = append(args, "s."+prop.PropertyName)
	}

	SQL := fmt.Sprintf("select %s from `%s` where %s", strings.Join(columnNames, ", "), statement.TableName, strings.Join(keyPairs, " and "))
	funcLines := fmt.Sprintf(`func Query%s (db *sql.DB, s *%s) (*%s,error) {
	SQL := "%s"
	ret := &%s{}
	err := db.QueryRow(SQL, %s).Scan(%s)
	if err != nil {
		return nil,err
	}
	return ret,nil
}`, statement.ModelName, statement.ModelName, statement.ModelName, SQL, statement.ModelName, strings.Join(args, ", "), strings.Join(binds, ", "))

	SQL1 := fmt.Sprintf("select count(*) from `%s`", statement.TableName)
	SQL2 := fmt.Sprintf("select %s from `%s` limit ?, ?", strings.Join(columnNames, ", "), statement.TableName)
	manyFuncLines := fmt.Sprintf(`func QueryMany%s (db *sql.DB, page int, size int) (int, []*%s, error) {
	SQL1 := "%s"
	count := 0
	err := db.QueryRow(SQL1).Scan(&count)
	if err != nil {
		return 0, nil, err
	}

	SQL2 := "%s"
	rows, err := db.Query(SQL2, (page-1)*size, size)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, nil, err
		}
	}
	defer rows.Close()

	results := make([]*%s, 0)
	for rows.Next() {
		ret := &%s{}
		rows.Scan(%s)
		results = append(results, ret)
	}
	return count, results, nil
}`, statement.ModelName, statement.ModelName, SQL1, SQL2, statement.ModelName, statement.ModelName, strings.Join(binds, ", "))
	return funcLines + "\n\n" + manyFuncLines, nil
}

func d(statement *parser.Statement) (string, []string) {
	args := make([]string, 0)
	keys := make([]*parser.Property, 0)
	for _, prop := range statement.Properties {
		if prop.Primary || prop.Uniq {
			keys = append(keys, prop)
		}
	}

	keyPairs := make([]string, 0)
	for _, prop := range keys {
		keyPairs = append(keyPairs, "`"+prop.ColumnName+"` = ?")
		args = append(args, "s."+prop.PropertyName)
	}

	SQL := fmt.Sprintf("delete from `%s` where %s", statement.TableName, strings.Join(keyPairs, " and "))
	funcLines := fmt.Sprintf(`func Delete%s (db *sql.DB, s *%s) error {
	SQL := "%s"
	ret, err := db.Exec(SQL, %s)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}`, statement.ModelName, statement.ModelName, SQL, strings.Join(args, ", "))
	return funcLines, nil
}
