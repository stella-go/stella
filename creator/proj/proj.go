// Copyright 2010-2021 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proj

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rakyll/statik/fs"

	_ "github.com/stella-go/stella/statik"
)

const (
	projectName      = "{{ project-name }}"
	projectNameCamel = "{{ ProjectName }}"
	projectNameSnake = "{{ project_name }}"
	projectNameUpper = "{{ PROJECT-NAME }}"
)

func Create(language string, stype string, name string, output string) error {
	statikFS, err := fs.New()
	if err != nil {
		return err
	}
	root := fmt.Sprintf("/%s/%s-project-template", language, stype)

	newRoot := output + "/" + name

	err = fs.Walk(statikFS, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		newPath := strings.ReplaceAll(path, root, newRoot)
		if info.IsDir() {
			err := os.MkdirAll(newPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			file, err := statikFS.Open(path)
			if err != nil {
				return err
			}
			raw, err := ioutil.ReadAll(file)
			if err != nil {
				return err
			}
			content := string(raw)
			camelName := nameToCamel(name)
			snakeName := nameToSnake(name)
			upperName := nameToUpper(name)
			content = strings.ReplaceAll(content, projectName, name)
			content = strings.ReplaceAll(content, projectNameCamel, camelName)
			content = strings.ReplaceAll(content, projectNameSnake, snakeName)
			content = strings.ReplaceAll(content, projectNameUpper, upperName)
			err = ioutil.WriteFile(newPath, []byte(content), info.Mode())
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func nameToCamel(name string) string {
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, ".", " ")
	tokens := strings.Split(name, " ")
	newName := ""
	for _, t := range tokens {
		if len(t) < 2 {
			newName += strings.ToUpper(t)
		} else {
			nt := strings.ToUpper(t[:1]) + strings.ToLower(t[1:])
			newName += nt
		}
	}
	return newName
}

func nameToSnake(name string) string {
	//todo maybe origin project name is a Camel
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return strings.ToLower(name)
}

func nameToUpper(name string) string {
	return strings.ToUpper(name)
}
