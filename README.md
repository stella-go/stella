# `stella` An efficient development tool
`stella` provides functions such as quickly creation of projects, conversion of SQL into structures and database operation templates, line number macros, etc.

## Installation
```bash
go get -u github.com/stella-go/stella
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
        stella generate -p model -i init.sql -o model -f model

  -banner
        output banner (default true)
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
  -m    only generate models
  -o string
        output dictionary
  -p string
        package name (default "model")
  -round string
        round time [s/ms/μs]
```

For example,
```sql
DROP TABLE IF EXISTS `tb_students`;
CREATE TABLE `tb_students` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'ROW ID',
    `no` VARCHAR (32) COMMENT 'STUDENT NUMBER',
    `name` VARCHAR (64) COMMENT 'STUDENT NAME',
    `age` INT COMMENT 'STUDENT AGE',
    `gender` VARCHAR (1) COMMENT 'STUDENT GENDER',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'CREATE TIME',
    `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'UPDATE TIME',
    PRIMARY KEY (`id`)
) ENGINE = INNODB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = 'STUDENT RECORDS';
```

Run command `stella generate -p model -i init.sql -o model`, Will generate two files `model_auto.go` and `model_curd_auto.go`
```go
package model

/**
 * Auto Generate by github.com/stella-go/stella on 2022/06/10.
 */

import (
	"fmt"
	"time"
)

type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	tm, err := time.ParseInLocation("\"2006-01-02 15:04:05\"", string(data), time.Local)
	if err != nil {
		tm, err := time.ParseInLocation("\"2006-01-02\"", string(data), time.Local)
		if err != nil {
			return err
		}
		*t = Time(tm)
		return err
	}
	*t = Time(tm)
	return err
}
func (t Time) String() string {
	return time.Time(t).String()
}

type TbStudents struct {
	Id         int    `json:"id"`
	No         string `json:"no"`
	Name       string `json:"name"`
	Age        int    `json:"age"`
	Gender     string `json:"gender"`
	CreateTime Time   `json:"create_time"`
	UpdateTime Time   `json:"update_time"`
}

func (s *TbStudents) String() string {
	return fmt.Sprintf("TbStudents{Id: %d, No: %s, Name: %s, Age: %d, Gender: %s, CreateTime: %v, UpdateTime: %v}", s.Id, s.No, s.Name, s.Age, s.Gender, s.CreateTime, s.UpdateTime)
}
```
```go
package model

/**
 * Auto Generate by github.com/stella-go/stella on 2022/06/10.
 */

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type DataSource interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// ==================== TbStudents ====================
func CreateTbStudents(db DataSource, s *TbStudents) error {
	if s == nil {
		return fmt.Errorf("pointer can not be nil")
	}
	SQL := "insert into `tb_students` (`no`, `name`, `age`, `gender`) values (?, ?, ?, ?)"
	ret, err := db.Exec(SQL, s.No, s.Name, s.Age, s.Gender)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func UpdateTbStudentsById(db DataSource, s *TbStudents) error {
	if s == nil {
		return fmt.Errorf("pointer can not be nil")
	}
	SQL := "update `tb_students` set %s where `id` = ?"
	set := ""
	args := make([]interface{}, 0)
	if s.No != "" {
		set += ", `no` = ? "
		args = append(args, s.No)
	}
	if s.Name != "" {
		set += ", `name` = ? "
		args = append(args, s.Name)
	}
	if s.Age != 0 {
		set += ", `age` = ? "
		args = append(args, s.Age)
	}
	if s.Gender != "" {
		set += ", `gender` = ? "
		args = append(args, s.Gender)
	}
	set = strings.TrimLeft(set, ",")
	set = strings.TrimSpace(set)
	if set == "" {
		return fmt.Errorf("all field is zero")
	}
	SQL = fmt.Sprintf(SQL, set)
	args = append(args, s.Id)
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

func QueryTbStudentsById(db DataSource, s *TbStudents) (*TbStudents, error) {
	if s == nil {
		return nil, fmt.Errorf("pointer can not be nil")
	}
	SQL := "select `id`, `no`, `name`, `age`, `gender`, `create_time`, `update_time` from `tb_students` where `id` = ?"
	ret := &TbStudents{}
	var Age sql.NullInt32
	var No, Name, Gender sql.NullString
	err := db.QueryRow(SQL, s.Id).Scan(&ret.Id, &No, &Name, &Age, &Gender, &ret.CreateTime, &ret.UpdateTime)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}
	if Age.Valid {
		ret.Age = int(Age.Int32)
	}
	if No.Valid {
		ret.No = No.String
	}
	if Name.Valid {
		ret.Name = Name.String
	}
	if Gender.Valid {
		ret.Gender = Gender.String
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
		if s.Id != 0 {
			where += "and `id` = ? "
			args = append(args, s.Id)
		}
		if s.No != "" {
			where += "and `no` = ? "
			args = append(args, s.No)
		}
		if s.Name != "" {
			where += "and `name` = ? "
			args = append(args, s.Name)
		}
		if s.Age != 0 {
			where += "and `age` = ? "
			args = append(args, s.Age)
		}
		if s.Gender != "" {
			where += "and `gender` = ? "
			args = append(args, s.Gender)
		}
		if !time.Time(s.CreateTime).IsZero() {
			where += "and `create_time` = ? "
			args = append(args, time.Time(s.CreateTime))
		}
		if !time.Time(s.UpdateTime).IsZero() {
			where += "and `update_time` = ? "
			args = append(args, time.Time(s.UpdateTime))
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

	results := make([]*TbStudents, 0)
	for rows.Next() {
		ret := &TbStudents{}
		var Age sql.NullInt32
		var No, Name, Gender sql.NullString
		err = rows.Scan(&ret.Id, &No, &Name, &Age, &Gender, &ret.CreateTime, &ret.UpdateTime)
		if err != nil {
			return 0, nil, err
		}
		if Age.Valid {
			ret.Age = int(Age.Int32)
		}
		if No.Valid {
			ret.No = No.String
		}
		if Name.Valid {
			ret.Name = Name.String
		}
		if Gender.Valid {
			ret.Gender = Gender.String
		}
		results = append(results, ret)
	}
	return count, results, nil
}

func DeleteTbStudentsById(db DataSource, s *TbStudents) error {
	if s == nil {
		return fmt.Errorf("pointer can not be nil")
	}
	SQL := "delete from `tb_students` where `id` = ?"
	ret, err := db.Exec(SQL, s.Id)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
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