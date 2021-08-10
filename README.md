# `stella` An efficient development tool
`stella` provides functions such as quickly creation of projects, conversion of SQL into structures and database operation templates, line number macros, etc.

## Installation
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
        stella generate -p model -i init.sql -o model

  -h    print help info
  -i string
        input sql file
  -o string
        output dictionary
  -p string
        package name (default "model")
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

Run command `stella generate -p model -i init.sql -o model`, Will generate two files `model.go` and `model_curd.go`
```go
package model

/**
 * Auto Generate by github.com/stella-go/stella on 2021/08/10.
 */

import (
	"fmt"
	"time"
)

type TbStudents struct {
	Id         int       `json:"id"`
	No         string    `json:"no"`
	Name       string    `json:"name"`
	Age        int       `json:"age"`
	Gender     string    `json:"gender"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (s *TbStudents) String() string {
	return fmt.Sprintf("TbStudents{Id: %d, No: %s, Name: %s, Age: %d, Gender: %s, CreateTime: %v, UpdateTime: %v}", s.Id, s.No, s.Name, s.Age, s.Gender, s.CreateTime, s.UpdateTime)
}
```
```go
package model

/**
 * Auto Generate by github.com/stella-go/stella on 2021/08/10.
 */

import (
	"database/sql"
)

// ==================== TbStudents ====================

func CreateTbStudents(db *sql.DB, s *TbStudents) error {
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

func UpdateTbStudents(db *sql.DB, s *TbStudents) error {
	SQL := "update `tb_students` set `no` = ?, `name` = ?, `age` = ?, `gender` = ? where `id` = ?"
	ret, err := db.Exec(SQL, s.No, s.Name, s.Age, s.Gender, s.Id)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func QueryTbStudents(db *sql.DB, s *TbStudents) (*TbStudents, error) {
	SQL := "select `id`, `no`, `name`, `age`, `gender`, `create_time`, `update_time` from `tb_students` where `id` = ?"
	ret := &TbStudents{}
	err := db.QueryRow(SQL, s.Id).Scan(&ret.Id, &ret.No, &ret.Name, &ret.Age, &ret.Gender, &ret.CreateTime, &ret.UpdateTime)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func QueryManyTbStudents(db *sql.DB, page int, size int) (int, []*TbStudents, error) {
	SQL1 := "select count(*) from `tb_students`"
	count := 0
	err := db.QueryRow(SQL1).Scan(&count)
	if err != nil {
		return 0, nil, err
	}

	SQL2 := "select `id`, `no`, `name`, `age`, `gender`, `create_time`, `update_time` from `tb_students` limit ?, ?"
	rows, err := db.Query(SQL2, (page-1)*size, size)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, nil, err
		}
	}
	defer rows.Close()

	results := make([]*TbStudents, 0)
	for rows.Next() {
		ret := &TbStudents{}
		rows.Scan(&ret.Id, &ret.No, &ret.Name, &ret.Age, &ret.Gender, &ret.CreateTime, &ret.UpdateTime)
		results = append(results, ret)
	}
	return count, results, nil
}

func DeleteTbStudents(db *sql.DB, s *TbStudents) error {
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
***NOTICE*: The java and node templates have not yet been implemented, And go only have a server template.

```bash
Usage: 
        stella create -n my-project

  -h    print help info
  -l string
        projcet language [go/java/node] (default "go")
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
```

For example,
```go
//go:generate stella line
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
//go:generate stella line
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
//go:generate stella line
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
//go:generate stella line
package main

import (
	"fmt"
)

func main() {
	fmt.Println("__LINE:main.go:9__")
	fmt.Println("__LINE:main.go:10__")
}
```