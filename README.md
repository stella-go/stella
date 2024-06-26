# `stella` An efficient development tool
`stella` provides functions such as quickly creation of projects, conversion of SQL into structures and database operation templates, line number macros, etc.

## Installation
```bash
# GO111MODULE=off go get -u github.com/stella-go/stella
go get -u github.com/stella-go/stella
```
or
```bash
go install github.com/stella-go/stella
```

## Command Line
```bash
Usage: 

	sub-commands:
		generate	Generate template code.
		create		Create template project.
		line		Fill __LINE__ symbol.

	stella <command> -h for more info.
```

### Generate
Conversion of SQL into structures and database operation templates

```bash
Usage: 
        stella generate -i init.sql -o model 

  -asc string
        order by
  -banner
        output banner (default true)
  -curd
        generate curd
  -desc string
        reverse order by
  -f string
        output file name
  -h    print help info
  -help
        print help info
  -i string
        input sql file
  -logic string
        logic delete
  -m    generate models (default true)
  -o string
        output dictionary
  -p string
        package name
  -panic
        panic style
  -round string
        round time [s/ms/μs] (default "s")
  -router
        generate router
  -service
        generate service
  -std
        stdout print
  -sub string
        sql subset
```

For example,
```sql
DROP TABLE IF EXISTS `tb_students`;
CREATE TABLE `tb_students` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'ROW ID',
    `no` VARCHAR (32) COMMENT 'STUDENT NUMBER',
    `name` VARCHAR (64) COMMENT 'STUDENT NAME',
    `age` INT COMMENT 'STUDENT AGE',
    `gender` VARCHAR (1) DEFAULT NULL COMMENT 'STUDENT GENDER',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'CREATE TIME',
    `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'UPDATE TIME',
    PRIMARY KEY (`id`)
) ENGINE = INNODB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = 'STUDENT RECORDS';
```

Run command `stella generate -p model -i init.sql -o model`, Will generate two files `model_auto.go` and `model_curd_auto.go`
```go
package model

import (
	"fmt"
	"github.com/stella-go/siu/t/n"
)

type TbStudents struct {
	Id         *n.Int    `form:"id" json:"id,omitempty"`
	No         *n.String `form:"no" json:"no,omitempty"`
	Name       *n.String `form:"name" json:"name,omitempty"`
	Age        *n.Int    `form:"age" json:"age,omitempty"`
	Gender     *n.String `form:"gender" json:"gender,omitempty"`
	CreateTime *n.Time   `form:"create_time" json:"create_time,omitempty"`
	UpdateTime *n.Time   `form:"update_time" json:"update_time,omitempty"`
}

func (s *TbStudents) String() string {
	return fmt.Sprintf("TbStudents{Id: %s, No: %s, Name: %s, Age: %s, Gender: %s, CreateTime: %s, UpdateTime: %s}", s.Id, s.No, s.Name, s.Age, s.Gender, s.CreateTime, s.UpdateTime)
}
```
```go
package model

import (
	"database/sql"
	"fmt"
	"github.com/stella-go/siu/t"
	"strings"
)

type DataSource interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// ==================== TbStudents ====================
func CreateTbStudents(db DataSource, s *TbStudents) (int64, error) {
	if s == nil {
		return 0, t.Error(fmt.Errorf("pointer can not be nil"))
	}
	SQL := "insert into `tb_students` (%s) values (%s)"
	columns := []string{"`no`", "`name`", "`age`"}
	values := []string{"?", "?", "?"}
	args := []interface{}{s.No, s.Name, s.Age}
	if s.Gender != nil {
		columns = append(columns, "`gender`")
		values = append(values, "?")
		args = append(args, s.Gender)
	}
	SQL = fmt.Sprintf(SQL, strings.Join(columns, ", "), strings.Join(values, ", "))
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

func UpdateTbStudentsById(db DataSource, s *TbStudents) error {
	if s == nil {
		return t.Error(fmt.Errorf("pointer can not be nil"))
	}
	SQL := "update `tb_students` set %s where `id` = ?"
	set := ""
	args := make([]interface{}, 0)
	if s.No != nil {
		set += ", `no` = ? "
		args = append(args, s.No)
	}
	if s.Name != nil {
		set += ", `name` = ? "
		args = append(args, s.Name)
	}
	if s.Age != nil {
		set += ", `age` = ? "
		args = append(args, s.Age)
	}
	if s.Gender != nil {
		set += ", `gender` = ? "
		args = append(args, s.Gender)
	}
	set = strings.TrimLeft(set, ",")
	set = strings.TrimSpace(set)
	if set == "" {
		return fmt.Errorf("all field is nil")
	}
	SQL = fmt.Sprintf(SQL, set)
	args = append(args, s.Id)
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

func QueryTbStudentsById(db DataSource, s *TbStudents) (*TbStudents, error) {
	if s == nil {
		return nil, t.Error(fmt.Errorf("pointer can not be nil"))
	}
	SQL := "select `id`, `no`, `name`, `age`, `gender`, `create_time`, `update_time` from `tb_students` where `id` = ?"
	ret := &TbStudents{}
	err := db.QueryRow(SQL, s.Id).Scan(&ret.Id, &ret.No, &ret.Name, &ret.Age, &ret.Gender, &ret.CreateTime, &ret.UpdateTime)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, t.Error(err)
		}
		return nil, nil
	}
	return ret, nil
}
func QueryManyTbStudents(db DataSource, s *TbStudents, page int, size int) (int, []*TbStudents, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	SQL1 := "select count(*) from `tb_students` %s"
	SQL2 := "select `id`, `no`, `name`, `age`, `gender`, `create_time`, `update_time` from `tb_students` %s limit ?, ?"
	where := ""
	args := make([]interface{}, 0)
	if s != nil {
		if s.Id != nil {
			where += "and `id` = ? "
			args = append(args, s.Id)
		}
		if s.No != nil {
			where += "and `no` = ? "
			args = append(args, s.No)
		}
		if s.Name != nil {
			where += "and `name` = ? "
			args = append(args, s.Name)
		}
		if s.Age != nil {
			where += "and `age` = ? "
			args = append(args, s.Age)
		}
		if s.Gender != nil {
			where += "and `gender` = ? "
			args = append(args, s.Gender)
		}
		if s.CreateTime != nil {
			where += "and `create_time` = ? "
			args = append(args, s.CreateTime)
		}
		if s.UpdateTime != nil {
			where += "and `update_time` = ? "
			args = append(args, s.UpdateTime)
		}
		where = strings.TrimLeft(where, "and")
		where = strings.TrimSpace(where)
		if where != "" {
			where = "where " + where
		}
	}
	SQL1 = fmt.Sprintf(SQL1, where)
	SQL2 = fmt.Sprintf(SQL2, where)
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

	results := make([]*TbStudents, 0)
	for rows.Next() {
		ret := &TbStudents{}
		err = rows.Scan(&ret.Id, &ret.No, &ret.Name, &ret.Age, &ret.Gender, &ret.CreateTime, &ret.UpdateTime)
		if err != nil {
			return 0, nil, t.Error(err)
		}
		results = append(results, ret)
	}
	return count, results, nil
}

func DeleteTbStudentsById(db DataSource, s *TbStudents) error {
	if s == nil {
		return t.Error(fmt.Errorf("pointer can not be nil"))
	}
	SQL := "delete from `tb_students` where `id` = ?"
	ret, err := db.Exec(SQL, s.Id)
	if err != nil {
		return t.Error(err)
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return t.Error(err)
	}
	return nil
}
```

### Create
Create a template project.

**NOTICE**: The java and node templates have not yet been implemented, And go only have a server template.

```bash
Usage: 
        stella create -n my-project

  -h    print help info
  -help
        print help info
  -l string
        projcet language (default "go")
  -n string
        project name (default "demo")
  -o string
        output dictionary (default ".")
  -t string
        project type [server/sdk] (default "server")
```

### Line
Replace the `__LINE__` tag with the line number. It is used to solve the problem that the implementation of golang's built-in callee is expensive to obtain the line number.

```bash
Usage: 
        stella line [path/to [path/to ...]]

        stella line, By default it is equivalent to "stella line ."
  -h    print help info
  -help
        print help info
  -ignore string
        ignore file patterns
  -include string
        include file patterns (default "*.*")
  -s    use file short name
```

For example,
```go
//go:generate stella line -include=*.go -s .
package main

import (
	"fmt"
)

func main() {
	fmt.Println("__LINE__")
}
```

Run command `go generate`
```go
//go:generate stella line -include=*.go -s .
package main

import (
	"fmt"
)

func main() {
	fmt.Println("__LINE:main.go:9__")
}
```

And then insert an new line.
```go
//go:generate stella line -include=*.go -s .
package main

import (
	"fmt"
)

func main() {
	fmt.Println("__LINE__")
	fmt.Println("__LINE:main.go:9__")
}
```

Run command `go generate`
```go
//go:generate stella line -include=*.go -s .
package main

import (
	"fmt"
)

func main() {
	fmt.Println("__LINE:main.go:9__")
	fmt.Println("__LINE:main.go:10__")
}
```