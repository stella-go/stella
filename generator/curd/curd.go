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
)

var (
	typeMapping = map[string]string{
		"TINYINT":   "int",
		"INT":       "int",
		"BIGINT":    "int64",
		"FLOAT":     "float64",
		"CHAR":      "string",
		"VARCHAR":   "string",
		"TEXT":      "string",
		"DATE":      "Time",
		"DATETIME":  "Time",
		"TIMESTAMP": "Time",
		"default":   "interface{}",
	}
	nullTypesMapping = map[string]string{
		"bool":    "sql.NullBool",
		"int":     "sql.NullInt32",
		"int64":   "sql.NullInt64",
		"float64": "sql.NullFloat64",
		"string":  "sql.NullString",
		"Time":    "sql.NullTime",
	}
	nullTypesValueMapping = map[string]string{
		"sql.NullBool":    "Bool",
		"sql.NullInt32":   "Int32",
		"sql.NullInt64":   "Int64",
		"sql.NullFloat64": "Float64",
		"sql.NullString":  "String",
		"sql.NullTime":    "Time",
	}
	notZeroValueMapping = map[string]string{
		"bool":    "!%s",
		"int":     "%s != 0",
		"int64":   "%s != 0",
		"float64": "math.Float64bits(%s) != 0",
		"string":  "%s != \"\"",
		"Time":    "!time.Time(%s).IsZero()",
	}
	nullTypesSort = []string{"sql.NullBool", "sql.NullInt32", "sql.NullInt64", "sql.NullFloat64", "sql.NullString", "sql.NullTime"}
	unDeleteMap   = map[interface{}]interface{}{
		`\"0\"`: `\"1\"`,
		`\"1\"`: `\"0\"`,
		`1`:     `0`,
		`0`:     `1`,
	}
)

func Generate(pkg string, statements []*parser.Statement, logic string) string {
	importsMap := make(map[string]common.Void)
	importsMap["database/sql"] = common.Null
	importsMap["fmt"] = common.Null
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
		function, imports = d(statement, logic)
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
	return fmt.Sprintf("package %s\n\n/**\n * Auto Generate by github.com/stella-go/stella on %s.\n */\n\nimport (\n%s\n)\n\n%s\n\n%s", pkg, time.Now().Format("2006/01/02"), strings.Join(importsLines, "\n"), datasourceLines, strings.Join(functions, "\n"))
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
    if s == nil {
        return fmt.Errorf("pointer can not be nil")
    }
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
}
`, modelName, modelName, SQL, strings.Join(args, ", "))
	return funcLines, nil
}

func u(statement *parser.Statement) (string, []string) {
	importMath := false
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
			fieldType := typeMapping[col.Type]
			if fieldType == "float64" {
				importMath = true
			}
			arg := "s." + fieldName
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
			}
			set += fmt.Sprintf(`if %s {
        set += ", `+"`%s`"+` = ? "
        args = append(args, %s)
    }
    `, fmt.Sprintf(notZeroValueMapping[fieldType], "s."+fieldName), col.ColumnName, arg)
		}
		set += `set = strings.TrimLeft(set, ",")
    set = strings.TrimSpace(set)
    if set == "" {
        return fmt.Errorf("all field is zero")
    }
    SQL = fmt.Sprintf(SQL, set)`
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
		SQL := fmt.Sprintf("update `%s` set %%s where %s", statement.TableName.Name, strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func Update%sBy%s (db DataSource, s *%s) error {
    if s == nil {
        return fmt.Errorf("pointer can not be nil")
    }
    SQL := "%s"
    %s
	args = append(args, %s)
    ret, err := db.Exec(SQL, args...)
    if err != nil {
        return err
    }
    _, err = ret.RowsAffected()
    if err != nil {
        return err
    }
    return nil
}
`, modelName, strings.Join(fields, ""), modelName, SQL, set, strings.Join(args, ", "))
	}
	if importMath {
		return funcLines, []string{"math"}
	}
	return funcLines, nil
}

func r(statement *parser.Statement) (string, []string) {
	importMath := false
	modelName := generator.FirstUpperCamelCase(statement.TableName.Name)
	funcLines := ""
	names := make([]string, 0)
	nullable := make(map[string][]string)
	binds := make([]string, 0)

	for _, col := range statement.Columns {
		names = append(names, "`"+col.ColumnName.Name+"`")
		fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
		if !col.NotNull {
			typ, ok := typeMapping[col.Type]
			if !ok {
				typ = typeMapping["default"]
			}
			nullType := nullTypesMapping[typ]
			if list, ok := nullable[nullType]; ok {
				nullable[nullType] = append(list, fieldName)
			} else {
				nullable[nullType] = []string{fieldName}
			}
			binds = append(binds, "&"+fieldName)
		} else {
			binds = append(binds, "&ret."+fieldName)
		}
	}

	nullableDefinitions := make([]string, 0)
	for _, v := range nullTypesSort {
		if names, ok := nullable[v]; ok {
			nullableDefinitions = append(nullableDefinitions, fmt.Sprintf("        var %s %s", strings.Join(names, ", "), v))
		}
	}
	nullableDefinition := strings.Join(nullableDefinitions, "\n")
	if nullableDefinition != "" {
		nullableDefinition = "\n" + nullableDefinition
	}
	nullableAssignments := make([]string, 0)
	for _, v := range nullTypesSort {
		if names, ok := nullable[v]; ok {
			for _, name := range names {
				nullValue := name + "." + nullTypesValueMapping[v]
				if v == "sql.NullInt32" {
					nullValue = "int(" + nullValue + ")"
				}
				if v == "sql.NullTime" {
					nullValue = "Time(" + nullValue + ")"
				}
				nullableAssignments = append(nullableAssignments, fmt.Sprintf(`    if %s.Valid {
            ret.%s = %s
        }`, name, name, nullValue))
			}
		}
	}
	nullableAssignment := strings.Join(nullableAssignments, "\n    ")
	if nullableAssignment != "" {
		nullableAssignment = "\n    " + nullableAssignment
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
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
			}
			args = append(args, arg)
			fields = append(fields, fieldName)
		}
		SQL := fmt.Sprintf("select %s from `%s` where %s", strings.Join(names, ", "), statement.TableName, strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func Query%sBy%s (db DataSource, s *%s) (*%s, error) {
    if s == nil {
        return nil, fmt.Errorf("pointer can not be nil")
    }
    SQL := "%s"
    ret := &%s{}%s
    err := db.QueryRow(SQL, %s).Scan(%s)
    if err != nil {
        if err != sql.ErrNoRows {
            return nil, err
        }
        return nil, nil
    }%s
    return ret, nil
}
`, modelName, strings.Join(fields, ""), modelName, modelName, SQL, modelName, nullableDefinition, strings.Join(args, ", "), strings.Join(binds, ", "), nullableAssignment)
	}

	indexKeyPairs := getIndexKeyPairs(statement)
	for _, keys := range indexKeyPairs {
		fields := make([]string, 0)
		conditions := make([]string, 0)
		args := make([]string, 0)
		for _, col := range keys {
			conditions = append(conditions, "`"+col.ColumnName.Name+"` = ?")
			fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
			arg := "s." + fieldName
			if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
				arg = "time.Time(" + arg + ")"
			}
			args = append(args, arg)
			fields = append(fields, fieldName)
		}
		SQL1 := fmt.Sprintf("select count(*) from `%s` where %s", statement.TableName.Name, strings.Join(conditions, " and "))
		SQL2 := fmt.Sprintf("select %s from `%s` where %s limit ?, ?", strings.Join(names, ", "), statement.TableName.Name, strings.Join(conditions, " and "))
		funcLines += fmt.Sprintf(`func QueryMany%sBy%s (db DataSource, s *%s, page int, size int) (int, []*%s, error) {
    if s == nil {
        return 0, nil, fmt.Errorf("pointer can not be nil")
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
        ret := &%s{}%s
        err = rows.Scan(%s)
        if err != nil {
            return 0, nil, err
        }%s
        results = append(results, ret)
    }
    return count, results, nil
}
`, modelName, strings.Join(fields, ""), modelName, modelName, SQL1, strings.Join(args, ", "), SQL2, strings.Join(args, ", "), modelName, modelName, nullableDefinition, strings.Join(binds, ", "), nullableAssignment)
	}

	where := `where := ""
    args := make([]interface{}, 0)
    if s != nil {
`
	for _, col := range statement.Columns {
		fieldName := generator.FirstUpperCamelCase(col.ColumnName.Name)
		fieldType := typeMapping[col.Type]
		if fieldType == "float64" {
			importMath = true
		}
		arg := "s." + fieldName
		if col.Type == "DATE" || col.Type == "DATETIME" || col.Type == "TIMESTAMP" {
			arg = "time.Time(" + arg + ")"
		}
		where += fmt.Sprintf(`        if %s {
            where += "and `+"`%s`"+` = ? "
            args = append(args, %s)
        }
`, fmt.Sprintf(notZeroValueMapping[fieldType], "s."+fieldName), col.ColumnName, arg)
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
        ret := &%s{}%s
        err = rows.Scan(%s)
        if err != nil {
            return 0, nil, err
        }%s
        results = append(results, ret)
    }
    return count, results, nil
}
`, modelName, modelName, modelName, SQL1, SQL2, where, modelName, modelName, nullableDefinition, strings.Join(binds, ", "), nullableAssignment)

	if importMath {
		return funcLines, []string{"math"}
	}
	return funcLines, nil
}

func d(statement *parser.Statement, logic string) (string, []string) {
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
		funcTemplate := `func %sDelete%sBy%s (db DataSource, s *%s) error {
    if s == nil {
        return fmt.Errorf("pointer can not be nil")
    }
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
