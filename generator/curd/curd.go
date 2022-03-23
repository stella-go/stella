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
	"github.com/stella-go/stella/generator"
	"github.com/stella-go/stella/generator/parser"
)

func Generate(pkg string, statements []*parser.Statement) string {
	importsMap := make(map[string]common.Void)
	importsMap["database/sql"] = common.Null
	importsMap["fmt"] = common.Null
	importsMap["reflect"] = common.Null
	importsMap["strings"] = common.Null
	importsMap["time"] = common.Null
	functions := make([]string, 0)
	for _, statement := range statements {
		functions = append(functions, "// ==================== "+generator.FirstUpperCamelCase(statement.TableName.Name)+" ====================")
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

	datasourceLines := `type DataSource interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		QueryRow(query string, args ...interface{}) *sql.Row
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}`
	return fmt.Sprintf("package %s\n\n/**\n * Auto Generate by github.com/stella-go/stella on %s.\n */\n\nimport (\n%s\n)\n\n%s\n\n%s", pkg, time.Now().Format("2006/01/02"), strings.Join(importsLines, "\n"), datasourceLines, strings.Join(functions, "\n\n"))
}

func c(statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	columnNames := make([]string, 0)
	placeHolder := make([]string, 0)
	args := make([]string, 0)
	for _, col := range statement.Columns {
		if col.AutoIncrement || col.CurrentTimestamp {
			continue
		}
		columnNames = append(columnNames, "`"+col.ColumnName.Name+"`")
		placeHolder = append(placeHolder, "?")
		arg := "s." + generator.FirstUpperCamelCase(col.ColumnName.Name)
		if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
			arg = "time.Time(" + arg + ")"
		}
		args = append(args, arg)
	}
	SQL := fmt.Sprintf("insert into `%s` (%s) values (%s)", statement.TableName.Name, strings.Join(columnNames, ", "), strings.Join(placeHolder, ", "))
	funcLines := fmt.Sprintf(`func Create%s (db DataSource, s *%s) error {
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
}`+"\n\n", modelName, modelName, SQL, strings.Join(args, ", "))
	return funcLines, nil
}

func u(statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines := ""
	uniqKeyPairs := getUniqKeyPairs(statement)
	for _, keys := range uniqKeyPairs {
		args := make([]string, 0)
		values := make([]string, 0)

		for _, col := range statement.Columns {
			if col.AutoIncrement || col.CurrentTimestamp {
				continue
			}
			if contains(keys, col) {
				continue
			}
			values = append(values, "`"+col.ColumnName.Name+"` = ?")
			arg := "s." + generator.FirstUpperCamelCase(col.ColumnName.Name)
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
			}
			args = append(args, arg)
		}
		fields := make([]string, 0)
		conditions := make([]string, 0)
		for _, col := range keys {
			conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
			arg := "s." + generator.FirstUpperCamelCase(col.ColumnName.Name)
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
			}
			args = append(args, arg)
			fields = append(fields, generator.FirstUpperCamelCase(col.ColumnName.Name))
		}
		SQL := fmt.Sprintf("update `%s` set %s where %s", statement.TableName.Name, strings.Join(values, ", "), strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func Update%sBy%s (db DataSource, s *%s) error {
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
}`+"\n\n", modelName, strings.Join(fields, ""), modelName, SQL, strings.Join(args, ", "))
	}
	return funcLines, nil
}

func r(statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines := ""
	names := make([]string, 0)
	binds := make([]string, 0)

	for _, col := range statement.Columns {
		names = append(names, "`"+col.ColumnName.Name+"`")
		binds = append(binds, "&ret."+generator.FirstUpperCamelCase(col.ColumnName.Name))
	}

	uniqKeyPairs := getUniqKeyPairs(statement)
	for _, keys := range uniqKeyPairs {
		fields := make([]string, 0)
		conditions := make([]string, 0)
		args := make([]string, 0)
		for _, col := range keys {
			conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
			arg := "s." + generator.FirstUpperCamelCase(col.ColumnName.Name)
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
			}
			args = append(args, arg)
			fields = append(fields, generator.FirstUpperCamelCase(col.ColumnName.Name))
		}
		SQL := fmt.Sprintf("select %s from `%s` where %s", strings.Join(names, ", "), statement.TableName, strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func Query%sBy%s (db DataSource, s *%s) (*%s, error) {
	SQL := "%s"
	ret := &%s{}
	err := db.QueryRow(SQL, %s).Scan(%s)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil,err
		}
		return nil,nil
	}
	return ret,nil
}`+"\n\n", modelName, strings.Join(fields, ""), modelName, modelName, SQL, modelName, strings.Join(args, ", "), strings.Join(binds, ", "))
	}

	indexKeyPairs := getIndexKeyPairs(statement)
	for _, keys := range indexKeyPairs {
		fields := make([]string, 0)
		conditions := make([]string, 0)
		args := make([]string, 0)
		for _, col := range keys {
			conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
			arg := "s." + generator.FirstUpperCamelCase(col.ColumnName.Name)
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
			}
			args = append(args, arg)
			fields = append(fields, generator.FirstUpperCamelCase(col.ColumnName.Name))
		}
		SQL1 := fmt.Sprintf("select count(*) from `%s` where %s", statement.TableName.Name, strings.Join(conditions, " and "))
		SQL2 := fmt.Sprintf("select %s from `%s` where %s limit ?, ?", strings.Join(names, ", "), statement.TableName.Name, strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func QueryMany%sBy%s (db DataSource, s *%s, page int, size int) (int, []*%s, error) {
	SQL1 := "%s"
	count := 0
	err := db.QueryRow(SQL1, %s).Scan(&count)
	if err != nil {
		return 0, nil, err
	}

	SQL2 := "%s"
	rows, err := db.Query(SQL2, %s, (page-1)*size, size)
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
}`+"\n\n", modelName, strings.Join(fields, ""), modelName, modelName, SQL1, strings.Join(args, ", "), SQL2, strings.Join(args, ", "), modelName, modelName, strings.Join(binds, ", "))
	}

	where := `    where := ""
    args := make([]interface{}, 0)
    if s != nil {
`
	for _, col := range statement.Columns {
		arg := "s." + generator.FirstUpperCamelCase(col.ColumnName.Name)
		if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
			arg = "time.Time(" + arg + ")"
		}
		where += fmt.Sprintf(`        if v := reflect.ValueOf(s.%s); !v.IsZero() {
            where += "and `+"`%s`"+` = ? "
            args = append(args, %s)
        }
`, generator.FirstUpperCamelCase(col.ColumnName.Name), col.ColumnName, arg)
	}

	where += `        where = strings.TrimLeft(where, "and")
	    where = strings.TrimSpace(where)
        if where != "" {
			where = "where " + where
        }
    }
	SQL1 = fmt.Sprintf(SQL1, where)
	SQL2 = fmt.Sprintf(SQL2, where)`

	SQL1 := fmt.Sprintf("select count(*) from `%s` %%s", statement.TableName.Name)
	SQL2 := fmt.Sprintf("select %s from `%s` %%s limit ?, ?", strings.Join(names, ", "), statement.TableName.Name)
	funcLines += fmt.Sprintf(`func QueryMany%s (db DataSource, s *%s, page int, size int) (int, []*%s, error) {
	SQL1 := "%s"
	SQL2 := "%s"
	%s
	count := 0
	err := db.QueryRow(SQL1, args...).Scan(&count)
	if err != nil {
		return 0, nil, err
	}
	args = append(args, (page-1)*size, size)
	rows, err := db.Query(SQL2, args...)
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
}`+"\n\n", modelName, modelName, modelName, SQL1, SQL2, where, modelName, modelName, strings.Join(binds, ", "))

	return funcLines, nil
}

func d(statement *parser.Statement) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines := ""
	uniqKeyPairs := getUniqKeyPairs(statement)
	for _, keys := range uniqKeyPairs {
		fields := make([]string, 0)
		conditions := make([]string, 0)
		args := make([]string, 0)
		for _, col := range keys {
			conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
			arg := "s." + generator.FirstUpperCamelCase(col.ColumnName.Name)
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
			}
			args = append(args, arg)
			fields = append(fields, generator.FirstUpperCamelCase(col.ColumnName.Name))
		}
		SQL := fmt.Sprintf("delete from `%s` where %s", statement.TableName, strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func Delete%sBy%s (db DataSource, s *%s) error {
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
}`+"\n\n", modelName, strings.Join(fields, ""), modelName, SQL, strings.Join(args, ", "))
	}
	return funcLines, nil
}

func getIndexKeyPairs(statement *parser.Statement) [][]*parser.ColumnDefinition {
	keyPairs := make([][]*parser.ColumnDefinition, 0)
	for _, pair := range statement.IndexKeyPairs {
		p := make([]*parser.ColumnDefinition, 0)
		for _, k := range pair {
			for _, c := range statement.Columns {
				if c.ColumnName.Name == k.Name {
					p = append(p, c)
					break
				}
			}
		}
		keyPairs = append(keyPairs, p)
	}
	return keyPairs
}

func getUniqKeyPairs(statement *parser.Statement) [][]*parser.ColumnDefinition {
	keyPairs := make([][]*parser.ColumnDefinition, 0)
	for _, col := range statement.Columns {
		if col.PrimaryKey || col.UniqueKey {
			keyPairs = append(keyPairs, []*parser.ColumnDefinition{col})
		}
	}
	for _, pair := range statement.PrimaryKeyPairs {
		p := make([]*parser.ColumnDefinition, 0)
		for _, k := range pair {
			for _, c := range statement.Columns {
				if c.ColumnName.Name == k.Name {
					p = append(p, c)
					break
				}
			}
		}
		keyPairs = append(keyPairs, p)
	}
	for _, pair := range statement.UniqKeyPairs {
		p := make([]*parser.ColumnDefinition, 0)
		for _, k := range pair {
			for _, c := range statement.Columns {
				if c.ColumnName.Name == k.Name {
					p = append(p, c)
					break
				}
			}
		}
		keyPairs = append(keyPairs, p)
	}
	return keyPairs
}

func contains(arr []*parser.ColumnDefinition, s *parser.ColumnDefinition) bool {
	for _, a := range arr {
		if s.ColumnName.Name == a.ColumnName.Name {
			return true
		}
	}
	return false
}
