package service

import (
	"testing"

	"github.com/stella-go/stella/generator/parser"
)

const sql = `
create table tb_dept(
	/* tb_dept
	 * dept table
     */
	-- abc
    id int auto_increment,#部门编号 整形 主键 自增长
    DePt_name varchar(18),#部门名称
    description varchar(100) DEFAULT NULL,#描述 /*  xxx */
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
    description varchar(100),#描述
	created_date datetime,
	key idx_c (created_date)
);
`

func TestGenerate(t *testing.T) {
	s := parser.Parse(sql)
	file := Generate("service", "dept", s, true)
	t.Log(file)
}

func TestGeneratePanic(t *testing.T) {
	s := parser.Parse(sql)
	file := GeneratePanic("service", "dept", s, true)
	t.Log(file)
}

func TestGenerateGorm(t *testing.T) {
	s := parser.Parse(sql)
	file := GenerateGorm("service", "dept", s, true)
	t.Log(file)
}
