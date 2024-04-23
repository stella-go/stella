// Copyright 2010-2024 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"testing"

	"github.com/stella-go/stella/generator/parser"
)

func TestGenerate(t *testing.T) {
	sql := `
	DROP TABLE IF EXISTS tb_students;
	CREATE TABLE tb_students (
		id INT NOT NULL AUTO_INCREMENT COMMENT 'ROW ID',
		no VARCHAR (32) COMMENT 'STUDENT NUMBER',
		name VARCHAR (64) COMMENT 'STUDENT NAME',
		age INT COMMENT 'STUDENT AGE',
		gender VARCHAR (1) DEFAULT NULL COMMENT 'STUDENT GENDER',
		create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'CREATE TIME',
		update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'UPDATE TIME',
		PRIMARY KEY (id)
	) ENGINE = INNODB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = 'STUDENT RECORDS';
`

	s := parser.Parse(sql)
	file := Generate("model", s, true, true)
	t.Log(file)
	file = Generate("model", s, true, false)
	t.Log(file)
}
