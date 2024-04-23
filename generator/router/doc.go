// Copyright 2010-2023 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/stella-go/stella/generator"
	"github.com/stella-go/stella/generator/parser"
	"github.com/stella-go/stella/version"
)

type RequestBean struct {
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type ResultBean struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type LinkedMap struct {
	m     map[string]interface{}
	names []string
}

func NewLinkedMap() *LinkedMap {
	return &LinkedMap{m: make(map[string]interface{}), names: make([]string, 0)}
}

func (p *LinkedMap) Put(key string, value interface{}) {
	p.m[key] = value
	p.names = append(p.names, key)
}

func (p *LinkedMap) MarshalJSON() ([]byte, error) {
	s := "{"
	for _, n := range p.names {
		v := p.m[n]
		bts, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		value := string(bts)
		s += fmt.Sprintf("\"%s\": %s,", n, value)
	}
	s = strings.Trim(s, ",") + "}"
	return []byte(s), nil
}

var typeSample = map[string]interface{}{
	"TINYINT":   1,
	"INT":       1,
	"BIGINT":    10000,
	"FLOAT":     3.14,
	"CHAR":      "c",
	"VARCHAR":   "s",
	"TEXT":      "abc",
	"DATE":      "2023-04-05",
	"DATETIME":  "2023-04-05 06:07:08",
	"TIMESTAMP": 1681466601123,
	"default":   struct{}{},
}

func GenerateDoc(statements []*parser.Statement, banner bool) string {
	header := `HOST=127.0.0.1
PORT=8080
`

	paragraphs := make([]string, 0)

	for _, statement := range statements {
		name := ""
		if statement.Comment != nil {
			name = statement.Comment.Comment
		} else {
			name = statement.TableName.Name
		}
		paragraphs = append(paragraphs, "## ==================== "+name+" ====================\n")

		paragraph := fields_doc(statement)
		paragraphs = append(paragraphs, paragraph)

		paragraph = c_doc(statement)
		paragraphs = append(paragraphs, paragraph)

		paragraph = u_doc(statement)
		paragraphs = append(paragraphs, paragraph)

		paragraph = r_doc(statement)
		paragraphs = append(paragraphs, paragraph)

		paragraph = d_doc(statement)
		paragraphs = append(paragraphs, paragraph)
	}

	bannerS := ""
	if banner {
		bannerS = fmt.Sprintf("\n/**\n * Auto Generate by github.com/stella-go/stella %s on %s.\n */\n", version.VERSION, time.Now().Format("2006/01/02"))
	}
	return fmt.Sprintf("# Application Document\n%s\n%s\n%s", bannerS, header, strings.Join(paragraphs, "\n"))
}

func fields_doc(statement *parser.Statement) string {
	name := ""
	if statement.Comment != nil {
		name = statement.Comment.Comment
	} else {
		name = statement.TableName.Name
	}
	maxLenth := 0
	for _, column := range statement.Columns {
		length := len(column.ColumnName.Name)
		if length > maxLenth {
			maxLenth = length
		}
	}
	doc := fmt.Sprintf("### %s Fields\n", name)
	for _, column := range statement.Columns {
		name := ""
		if statement.Comment != nil {
			name = column.Comment.Comment
		} else {
			name = column.ColumnName.Name
		}
		doc += fmt.Sprintf(fmt.Sprintf("%%-%ds : %%s\n", maxLenth), generator.ToSnakeCase(column.ColumnName.Name), name)
	}
	return doc
}

func c_doc(statement *parser.Statement) string {
	name := ""
	if statement.Comment != nil {
		name = statement.Comment.Comment
	} else {
		name = statement.TableName.Name
	}
	data := NewLinkedMap()
	for _, column := range statement.Columns {
		if column.AutoIncrement || column.OnUpdate || column.CurrentTimestamp {
			continue
		}
		data.Put(generator.ToSnakeCase(column.ColumnName.Name), typeSample[column.Type])
	}
	bts, err := json.Marshal(&RequestBean{Timestamp: 1681466601123, Data: data})
	if err != nil {
		panic(err)
	}
	content := string(bts)

	bts, err = json.Marshal(&ResultBean{Code: 200, Message: "success"})
	if err != nil {
		panic(err)
	}
	result := string(bts)

	return fmt.Sprintf(`### Create %s
- Ruquest
POST /api/%s

Content-Type: application/json;

%s

- Response
%s

- example
`+"```bash"+`
curl -XPOST -H "Content-Type: application/json" "http://${HOST}:${PORT}/api/%s" -d '%s'
%s
`+"```"+`
`, name, generator.ToStrikeCase(statement.TableName.Name), content, result, generator.ToStrikeCase(statement.TableName.Name), content, result)
}

func u_doc(statement *parser.Statement) string {
	name := ""
	if statement.Comment != nil {
		name = statement.Comment.Comment
	} else {
		name = statement.TableName.Name
	}
	data := NewLinkedMap()
	for _, column := range statement.Columns {
		if column.OnUpdate || column.CurrentTimestamp {
			continue
		}
		data.Put(generator.ToSnakeCase(column.ColumnName.Name), typeSample[column.Type])
	}
	bts, err := json.Marshal(&RequestBean{Timestamp: 1681466601123, Data: data})
	if err != nil {
		panic(err)
	}
	content := string(bts)

	bts, err = json.Marshal(&ResultBean{Code: 200, Message: "success"})
	if err != nil {
		panic(err)
	}
	result := string(bts)

	return fmt.Sprintf(`### Update %s
- Ruquest
PUT /api/%s

Content-Type: application/json;

%s

- Response
%s

- example
`+"```bash"+`
curl -XPUT -H "Content-Type: application/json" "http://${HOST}:${PORT}/api/%s" -d '%s'
%s
`+"```"+`
`, name, generator.ToStrikeCase(statement.TableName.Name), content, result, generator.ToStrikeCase(statement.TableName.Name), content, result)
}

func r_doc(statement *parser.Statement) string {
	paragraph := ""
	name := ""
	if statement.Comment != nil {
		name = statement.Comment.Comment
	} else {
		name = statement.TableName.Name
	}
	data := NewLinkedMap()
	for _, column := range statement.Columns {
		if column.AutoIncrement || column.OnUpdate || column.CurrentTimestamp {
			continue
		}
		data.Put(generator.ToSnakeCase(column.ColumnName.Name), typeSample[column.Type])
	}
	data.Put("page", 1)
	data.Put("size", 10)
	bts, err := json.Marshal(&RequestBean{Timestamp: 1681466601123, Data: data})
	if err != nil {
		panic(err)
	}
	content := string(bts)

	data = NewLinkedMap()
	for _, column := range statement.Columns {
		data.Put(generator.ToSnakeCase(column.ColumnName.Name), typeSample[column.Type])
	}
	bts, err = json.Marshal(&ResultBean{Code: 200, Message: "success", Data: map[string]interface{}{
		"count": 1,
		"list":  []*LinkedMap{data},
	}})
	if err != nil {
		panic(err)
	}
	result := string(bts)

	paragraph += fmt.Sprintf(`### Query All %s
- Ruquest
POST /api/%s/all

Content-Type: application/json;

%s

- Response
%s

- example
`+"```bash"+`
curl -XPOST -H "Content-Type: application/json" "http://${HOST}:${PORT}/api/%s/many" -d '%s'
%s
`+"```"+`

`, name, generator.ToStrikeCase(statement.TableName.Name), content, result, generator.ToStrikeCase(statement.TableName.Name), content, result)

	data = NewLinkedMap()
	primaryKeys := getPrimaryKeyPairs(statement)
	if len(primaryKeys) > 0 {
		keys := primaryKeys[0]
		for _, column := range keys {
			data.Put(generator.ToSnakeCase(column.ColumnName.Name), typeSample[column.Type])
		}
	}
	bts, err = json.Marshal(&RequestBean{Timestamp: 1681466601123, Data: data})
	if err != nil {
		panic(err)
	}
	content = string(bts)

	data = NewLinkedMap()
	for _, column := range statement.Columns {
		data.Put(generator.ToSnakeCase(column.ColumnName.Name), typeSample[column.Type])
	}
	bts, err = json.Marshal(&ResultBean{Code: 200, Message: "success", Data: data})
	if err != nil {
		panic(err)
	}
	result = string(bts)

	paragraph += fmt.Sprintf(`### Query One %s
- Request
POST /api/%s/one

Content-Type: application/json;

%s

- Response
%s

- example
`+"```bash"+`
curl -XPOST -H "Content-Type: application/json" "http://${HOST}:${PORT}/api/%s/one" -d '%s'
%s
`+"```"+`
`, name, generator.ToStrikeCase(statement.TableName.Name), content, result, generator.ToStrikeCase(statement.TableName.Name), content, result)
	return paragraph
}

func d_doc(statement *parser.Statement) string {
	name := ""
	if statement.Comment != nil {
		name = statement.Comment.Comment
	} else {
		name = statement.TableName.Name
	}
	data := NewLinkedMap()
	primaryKeys := getPrimaryKeyPairs(statement)
	if len(primaryKeys) > 0 {
		keys := primaryKeys[0]
		for _, column := range keys {
			data.Put(generator.ToSnakeCase(column.ColumnName.Name), typeSample[column.Type])
		}
	}
	bts, err := json.Marshal(&RequestBean{Timestamp: 1681466601123, Data: data})
	if err != nil {
		panic(err)
	}
	content := string(bts)

	bts, err = json.Marshal(&ResultBean{Code: 200, Message: "success"})
	if err != nil {
		panic(err)
	}
	result := string(bts)

	return fmt.Sprintf(`### Delete %s
- Request
DELETE /api/%s

Content-Type: application/json;

%s

- Response
%s

- example
`+"```bash"+`
curl -XDELETE -H "Content-Type: application/json" "http://${HOST}:${PORT}/api/%s" -d '%s'
%s
`+"```"+`
`, name, generator.ToStrikeCase(statement.TableName.Name), content, result, generator.ToStrikeCase(statement.TableName.Name), content, result)
}

func getPrimaryKeyPairs(statement *parser.Statement) [][]*parser.ColumnDefinition {
	keyPairs := make([][]*parser.ColumnDefinition, 0)
	for _, col := range statement.Columns {
		if col.PrimaryKey {
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
	return keyPairs
}
