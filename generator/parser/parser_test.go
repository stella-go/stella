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
	"testing"
)

func TestParse2(t *testing.T) {
	sql := `
drop table if exists tb_dept2;
create table tb_dept2(
    Id int auto_increment,#部门编号 整形 主键 自增长''""
    Name varchar(18),#部门名称
    description varchar(100),#描述
primary key (Id,Name)
);
`
	s := Parse(sql)
	t.Log(s)
}

func TestParse3(t *testing.T) {
	sql := "create table tb_dept(`Id` int primary key auto_increment, Name varchar(18), description varchar(100))"
	s := Parse(sql)
	t.Log(s)
}

func TestParse4(t *testing.T) {
	sql := `
create table tb_dept2(
    Id int auto_increment,#部门编号 整形 主键 自增长''""
    Name varchar(18),#部门名称
    description varchar(100),#描述
primary key (Id,Name),
unique key uniq_desc(description)
);
`
	t.Log(sql)
	s := Parse(sql)
	t.Log(s)
}

func TestParse5(t *testing.T) {
	sql := `
create table tb_dept2(
    Id int auto_increment,#部门编号 整形 主键 自增长''""
    Name varchar(18),#部门名称
    description varchar(100) unique key,#描述
primary key (Id,Name)
);
`
	t.Log(sql)
	s := Parse(sql)
	t.Log(s)
}

func TestSplitSQL(t *testing.T) {
	sql := `
create table tb_dept(
	/* tb_dept
	 * dept table
     */
	-- abc
    Id int auto_increment,#部门编号 整形 主键 自增长
    Name varchar(18),#部门名称
    description varchar(100),#描述 /*  xxx */
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
	splits := Parse(sql)
	t.Log(splits)
}
func TestSplitSQL2(t *testing.T) {
	sql := "create table tb_dept(`Id'` int primary key auto_increment, Name varchar(18), description varchar(100))"
	splits := Parse(sql)
	t.Log(splits)
}

func TestLen(t *testing.T) {
	s := "：abc"
	fmt.Println(len(s))
	fmt.Println(len([]byte(s)))
	fmt.Println(s[1])
}
