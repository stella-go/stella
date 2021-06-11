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

package line

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	re = regexp.MustCompile("(.*?__LINE)(.*?)(__.*?)")
)

func Fill(root string) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
			lines := bytes.Split(content, []byte("\n"))
			p := strings.ReplaceAll(path, "\\", "/")
			for line, c := range lines {
				lines[line] = re.ReplaceAll(c, []byte(fmt.Sprintf("${1}:%s:%d${3}", p, line+1)))
			}
			content = bytes.Join(lines, []byte("\n"))
			return ioutil.WriteFile(path, content, info.Mode())
		}
		return nil
	})
}
