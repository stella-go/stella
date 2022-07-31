// Copyright 2010-2022 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

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
	"github.com/stella-go/stella/version"
)

var (
	unDeleteMap = map[interface{}]interface{}{
		`\"0\"`: `\"1\"`,
		`\"1\"`: `\"0\"`,
		`1`:     `0`,
		`0`:     `1`,
	}
)

func Generate(pkg string, statements []*parser.Statement, banner bool, logic string, asc string, desc string, round string) string {
	importsMap := make(map[string]common.Void)
	importsMap["database/sql"] = common.Null
	importsMap["fmt"] = common.Null
	importsMap["strings"] = common.Null
	importsMap["github.com/stella-go/siu/t"] = common.Null
	functions := make([]string, 0)
	switch round {
	case "s":
		round = "time.Second"
	case "ms", "milli":
		round = "time.Millisecond"
	case "μs", "us", "micro":
		round = "time.Microsecond"
	default:
		round = ""
	}
	if round != "" {
		importsMap["time"] = common.Null
	}
	for _, statement := range statements {
		functions = append(functions, "// ==================== "+generator.FirstUpperCamelCase(statement.TableName.Name)+" ====================")
		function, imports := c(statement, round)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = u(statement, round)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}

		function, imports = r(statement, asc, desc, round)
		functions = append(functions, function)
		for _, i := range imports {
			importsMap[i] = common.Null
		}
		function, imports = d(statement, logic, round)
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
	bannerS := ""
	if banner {
		bannerS = fmt.Sprintf("\n/**\n * Auto Generate by github.com/stella-go/stella %s on %s.\n */\n", version.VERSION, time.Now().Format("2006/01/02"))

	}
	return fmt.Sprintf("package %s\n%s\nimport (\n%s\n)\n\n%s\n\n%s", pkg, bannerS, strings.Join(importsLines, "\n"), datasourceLines, strings.Join(functions, "\n"))
}

func c(statement *parser.Statement, round string) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	columns := make([]string, 0)
	values := make([]string, 0)
	args := make([]string, 0)
	for _, col := range statement.Columns {
		if col.AutoIncrement || col.CurrentTimestamp || col.DefaultValue {
			continue
		}
		fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
		columns = append(columns, "\"`"+col.ColumnName.Name+"`\"")
		values = append(values, "\"?\"")
		arg := "s." + fieldName
		if (col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP") && round != "" {
			arg = arg + ".Round(" + round + ")"
		}
		args = append(args, arg)
	}
	insert := fmt.Sprintf(`columns := []string{%s}
    values := []string{%s}
    args := []interface{}{%s}
`, strings.Join(columns, ", "), strings.Join(values, ", "), strings.Join(args, ", "))
	for _, col := range statement.Columns {
		if col.AutoIncrement || col.CurrentTimestamp {
			continue
		}
		if col.DefaultValue {
			fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
			arg := "s." + fieldName
			if (col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP") && round != "" {
				arg = arg + ".Round(" + round + ")"
			}
			insert += fmt.Sprintf(`    if s.%s != nil {
        columns = append(columns, "%s")
        values = append(values, "?")
        args = append(args, %s)
    }
`, fieldName, "`"+col.ColumnName.Name+"`", arg)
		}
	}
	insert += `    SQL = fmt.Sprintf(SQL, strings.Join(columns, ", "), strings.Join(values, ", "))`

	SQL := fmt.Sprintf("insert into `%s` (%%s) values (%%s)", statement.TableName.Name)
	funcLines := fmt.Sprintf(`func Create%s(db DataSource, s *%s) (int64, error) {
    if s == nil {
        return 0, t.Error(fmt.Errorf("pointer can not be nil"))
    }
    SQL := "%s"
    %s
    ret, err := db.Exec(SQL, args...)
    if err != nil {
        return 0, t.Error(err)
    }
    _, err = ret.RowsAffected()
    if err != nil {
        return 0, t.Error(err)
    }
    return ret.LastInsertId()
}
`, modelName, modelName, SQL, insert)
	return funcLines, nil
}

func u(statement *parser.Statement, round string) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines := ""
	uniqKeyPairs := getUniqKeyPairs(statement)
	for _, keys := range uniqKeyPairs {
		args := make([]string, 0)
		set := `set := ""
    args := make([]interface{}, 0)
    `
		for _, col := range statement.Columns {
			if col.AutoIncrement || col.CurrentTimestamp {
				continue
			}
			if contains(keys, col) {
				continue
			}
			fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
			arg := "s." + fieldName
			if (col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP") && round != "" {
				arg = arg + ".Round(" + round + ")"
			}
			set += fmt.Sprintf(`if %s != nil {
        set += ", `+"`%s`"+` = ? "
        args = append(args, %s)
    }
    `, "s."+fieldName, col.ColumnName, arg)
		}
		set += `set = strings.TrimLeft(set, ",")
    set = strings.TrimSpace(set)
    if set == "" {
        return fmt.Errorf("all field is nil")
    }
    SQL = fmt.Sprintf(SQL, set)`
		fields := make([]string, 0)
		conditions := make([]string, 0)
		for _, col := range keys {
			conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
			arg := "s." + generator.FirstUpperCamelCase(col.ColumnName.Name)
			if (col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP") && round != "" {
				arg = arg + ".Round(" + round + ")"
			}
			args = append(args, arg)
			fields = append(fields, generator.FirstUpperCamelCase(col.ColumnName.Name))
		}
		SQL := fmt.Sprintf("update `%s` set %%s where %s", statement.TableName.Name, strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func Update%sBy%s(db DataSource, s *%s) error {
    if s == nil {
        return t.Error(fmt.Errorf("pointer can not be nil"))
    }
    SQL := "%s"
    %s
    args = append(args, %s)
    ret, err := db.Exec(SQL, args...)
    if err != nil {
        return t.Error(err)
    }
    _, err = ret.RowsAffected()
    if err != nil {
        return t.Error(err)
    }
    return nil
}
`, modelName, strings.Join(fields, ""), modelName, SQL, set, strings.Join(args, ", "))
	}
	return funcLines, nil
}

func r(statement *parser.Statement, asc string, desc string, round string) (string, []string) {
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines := ""
	names := make([]string, 0)
	binds := make([]string, 0)

	for _, col := range statement.Columns {
		names = append(names, "`"+col.ColumnName.Name+"`")
		fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
		binds = append(binds, "&ret."+fieldName)
	}

	uniqKeyPairs := getUniqKeyPairs(statement)
	for _, keys := range uniqKeyPairs {
		fields := make([]string, 0)
		conditions := make([]string, 0)
		args := make([]string, 0)
		for _, col := range keys {
			conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
			fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
			arg := "s." + fieldName
			if (col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP") && round != "" {
				arg = arg + ".Round(" + round + ")"
			}
			args = append(args, arg)
			fields = append(fields, fieldName)
		}
		SQL := fmt.Sprintf("select %s from `%s` where %s", strings.Join(names, ", "), statement.TableName, strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func Query%sBy%s(db DataSource, s *%s) (*%s, error) {
    if s == nil {
        return nil, t.Error(fmt.Errorf("pointer can not be nil"))
    }
    SQL := "%s"
    ret := &%s{}
    err := db.QueryRow(SQL, %s).Scan(%s)
    if err != nil {
        if err != sql.ErrNoRows {
            return nil, t.Error(err)
        }
        return nil, nil
    }
    return ret, nil
}
`, modelName, strings.Join(fields, ""), modelName, modelName, SQL, modelName, strings.Join(args, ", "), strings.Join(binds, ", "))
	}
	type Order struct {
		FuncSuffix string
		Statement  string
	}
	orders := []*Order{{}}
	if asc != "" {
		columns := make([]string, 0)
		if (strings.HasPrefix(asc, "\"") && strings.HasSuffix(asc, "\"")) || (strings.HasPrefix(asc, "'") && strings.HasSuffix(asc, "'")) {
			asc = asc[1 : len(asc)-1]
		}
		columnNames := strings.Split(asc, ",")
		for _, name := range columnNames {
			for _, c := range statement.Columns {
				if c.ColumnName.Name != name {
					continue
				}
				columns = append(columns, c.ColumnName.Name)
			}
		}
		if len(columns) != 0 {
			s1 := make([]string, 0)
			s2 := make([]string, 0)
			for _, c := range columns {
				s1 = append(s1, generator.FirstUpperCamelCase(c))
				s2 = append(s2, "`"+c+"`")
			}
			orders = append(orders, &Order{FuncSuffix: fmt.Sprintf("OrderBy%s", strings.Join(s1, "")), Statement: fmt.Sprintf("order by %s ", strings.Join(s2, ", "))})
		}
	}
	if desc != "" {
		columns := make([]string, 0)
		if (strings.HasPrefix(desc, "\"") && strings.HasSuffix(desc, "\"")) || (strings.HasPrefix(desc, "'") && strings.HasSuffix(desc, "'")) {
			desc = desc[1 : len(desc)-1]
		}
		columnNames := strings.Split(desc, ",")
		for _, name := range columnNames {
			for _, c := range statement.Columns {
				if c.ColumnName.Name != name {
					continue
				}
				columns = append(columns, c.ColumnName.Name)
			}
		}
		if len(columns) != 0 {
			s1 := make([]string, 0)
			s2 := make([]string, 0)
			for _, c := range columns {
				s1 = append(s1, generator.FirstUpperCamelCase(c))
				s2 = append(s2, "`"+c+"`")
			}
			orders = append(orders, &Order{FuncSuffix: fmt.Sprintf("OrderBy%sDesc", strings.Join(s1, "")), Statement: fmt.Sprintf("order by %s desc ", strings.Join(s2, ", "))})
		}
	}
	for _, order := range orders {
		indexKeyPairs := getIndexKeyPairs(statement)
		for _, keys := range indexKeyPairs {
			fields := make([]string, 0)
			conditions := make([]string, 0)
			args := make([]string, 0)
			for _, col := range keys {
				conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
				fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
				arg := "s." + fieldName
				if (col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP") && round != "" {
					arg = arg + ".Round(" + round + ")"
				}
				args = append(args, arg)
				fields = append(fields, fieldName)
			}
			SQL1 := fmt.Sprintf("select count(*) from `%s` where %s", statement.TableName.Name, strings.Join(conditions, " and "))
			SQL2 := fmt.Sprintf("select %s from `%s` where %s %slimit ?, ?", strings.Join(names, ", "), statement.TableName.Name, strings.Join(conditions, " and "), order.Statement)
			funcLines += fmt.Sprintf(`func QueryMany%sBy%s%s(db DataSource, s *%s, page int, size int) (int, []*%s, error) {
    if s == nil {
        return 0, nil, t.Error(fmt.Errorf("pointer can not be nil"))
    }
    if page <= 0 {
        page = 1
    }
    if size <= 0 {
        size = 10
    }
    SQL1 := "%s"
    count := 0
    err := db.QueryRow(SQL1, %s).Scan(&count)
    if err != nil {
        return 0, nil, t.Error(err)
    }

    SQL2 := "%s"
    rows, err := db.Query(SQL2, %s, (page-1)*size, size)
    if err != nil {
        if err != sql.ErrNoRows {
            return 0, nil, t.Error(err)
        }
    }
    defer rows.Close()

    results := make([]*%s, 0)
    for rows.Next() {
        ret := &%s{}
        err = rows.Scan(%s)
        if err != nil {
            return 0, nil, t.Error(err)
        }
        results = append(results, ret)
    }
    return count, results, nil
}
`, modelName, strings.Join(fields, ""), order.FuncSuffix, modelName, modelName, SQL1, strings.Join(args, ", "), SQL2, strings.Join(args, ", "), modelName, modelName, strings.Join(binds, ", "))
		}

		where := `where := ""
    args := make([]interface{}, 0)
    if s != nil {
`
		for _, col := range statement.Columns {
			fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
			arg := "s." + fieldName
			if (col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP") && round != "" {
				arg = arg + ".Round(" + round + ")"
			}
			where += fmt.Sprintf(`        if %s != nil {
            where += "and `+"`%s`"+` = ? "
            args = append(args, %s)
        }
`, "s."+fieldName, col.ColumnName, arg)
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
		SQL2 := fmt.Sprintf("select %s from `%s` %%s %slimit ?, ?", strings.Join(names, ", "), statement.TableName.Name, order.Statement)
		funcLines += fmt.Sprintf(`func QueryMany%s%s(db DataSource, s *%s, page int, size int) (int, []*%s, error) {
    if page <= 0 {
        page = 1
    }
    if size <= 0 {
        size = 10
    }
    SQL1 := "%s"
    SQL2 := "%s"
    %s
    count := 0
    err := db.QueryRow(SQL1, args...).Scan(&count)
    if err != nil {
        return 0, nil, t.Error(err)
    }
    args = append(args, (page-1)*size, size)
    rows, err := db.Query(SQL2, args...)
    if err != nil {
        if err != sql.ErrNoRows {
            return 0, nil, t.Error(err)
        }
    }
    defer rows.Close()

    results := make([]*%s, 0)
    for rows.Next() {
        ret := &%s{}
        err = rows.Scan(%s)
        if err != nil {
            return 0, nil, t.Error(err)
        }
        results = append(results, ret)
    }
    return count, results, nil
}
`, modelName, order.FuncSuffix, modelName, modelName, SQL1, SQL2, where, modelName, modelName, strings.Join(binds, ", "))
	}
	return funcLines, nil
}

func d(statement *parser.Statement, logic string, round string) (string, []string) {
	var logicDelete bool
	var logicCol string
	var logicValue string
	if logic != "" {
		if index := strings.Index(logic, "="); index != -1 {
			logicCol, logicValue = logic[:index], logic[index+1:]
			if (strings.HasPrefix(logicValue, "\"") && strings.HasSuffix(logicValue, "\"")) || (strings.HasPrefix(logicValue, "'") && strings.HasSuffix(logicValue, "'")) {
				logicValue = logicValue[1 : len(logicValue)-1]
			}
			for _, c := range statement.Columns {
				if c.ColumnName.Name != logicCol {
					continue
				}
				if c.Type != "TINYINT" && c.Type != "INT" && c.Type != "BIGINT" && c.Type != "FLOAT" {
					logicValue = `\"` + logicValue + `\"`
				}
				logicDelete = true
			}
		}
	}
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines := ""
	uniqKeyPairs := getUniqKeyPairs(statement)
	for _, keys := range uniqKeyPairs {
		fields := make([]string, 0)
		conditions := make([]string, 0)
		args := make([]string, 0)
		for _, col := range keys {
			fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
			conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
			arg := "s." + fieldName
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
				if round != "" {
					arg = arg + ".Round(" + round + ")"
				}
			}
			args = append(args, arg)
			fields = append(fields, fieldName)
		}
		SQL := ""
		if logicDelete {
			SQL = fmt.Sprintf("update `%s` set `%s` = %s where %s", statement.TableName, logicCol, logicValue, strings.Join(conditions, " and "))
		} else {
			SQL = fmt.Sprintf("delete from `%s` where %s", statement.TableName, strings.Join(conditions, " and "))
		}
		funcTemplate := `func %sDelete%sBy%s(db DataSource, s *%s) error {
    if s == nil {
        return t.Error(fmt.Errorf("pointer can not be nil"))
    }
    SQL := "%s"
    ret, err := db.Exec(SQL, %s)
    if err != nil {
        return t.Error(err)
    }
    _, err = ret.RowsAffected()
    if err != nil {
        return t.Error(err)
    }
    return nil
}
`
		funcLines += fmt.Sprintf(funcTemplate, "", modelName, strings.Join(fields, ""), modelName, SQL, strings.Join(args, ", "))
		if logicDelete {
			if unDeleteValue, ok := unDeleteMap[logicValue]; ok {
				UNSQL := fmt.Sprintf("update `%s` set `%s` = %s where %s", statement.TableName, logicCol, unDeleteValue, strings.Join(conditions, " and "))
				funcLines += fmt.Sprintf(funcTemplate, "Un", modelName, strings.Join(fields, ""), modelName, UNSQL, strings.Join(args, ", "))
			}
		}

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
				if strings.EqualFold(c.ColumnName.Name, k.Name) {
					p = append(p, c)
					break
				}
			}
		}
		if len(p) != 0 {
			keyPairs = append(keyPairs, p)
		}
	}
	for _, pair := range statement.UniqKeyPairs {
		p := make([]*parser.ColumnDefinition, 0)
		for _, k := range pair {
			for _, c := range statement.Columns {
				if strings.EqualFold(c.ColumnName.Name, k.Name) {
					p = append(p, c)
					break
				}
			}
		}
		if len(p) != 0 {
			keyPairs = append(keyPairs, p)
		}
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
