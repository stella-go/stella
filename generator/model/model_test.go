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

package model

import (
	"testing"

	"github.com/stella-go/stella/generator/parser"
)

func TestGenerate(t *testing.T) {
	sql := `
create table tb_dept(
	/* tb_dept
	 * dept table
     */
	-- abc
    id int auto_increment,#部门编号 整形 主键 自增长
    DePt_name varchar(18),#部门名称
    description varchar(100),#描述 /*  xxx */
	status tinyint,
primary key(Id)
);

create table tb_dept2(
	/* tb_dept
	 * dept table
     */
	-- abc
    Id int primary key auto_increment,#部门编号 整形 主键 自增长''""
    Name varchar(18),#部门名称
    description varchar(100)#描述
);
`

	s := parser.Parse(sql)
	file := Generate("model", s, true)
	t.Log(file)
}
