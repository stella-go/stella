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

package generator

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/stella-go/stella/version"
)

var (
	reCamel = regexp.MustCompile(`_(\w)`)
	reUpper = regexp.MustCompile(`[A-Z]`)
)

// ImportsSet is a set of import paths, replacing map[string]common.Void.
type ImportsSet map[string]struct{}

func NewImportsSet(imports ...string) ImportsSet {
	m := make(ImportsSet)
	for _, i := range imports {
		m[i] = struct{}{}
	}
	return m
}

func (m ImportsSet) Add(imports ...string) {
	for _, i := range imports {
		m[i] = struct{}{}
	}
}

func (m ImportsSet) Lines() []string {
	lines := make([]string, 0, len(m))
	for i := range m {
		if i == "" {
			continue
		}
		lines = append(lines, "\t\""+i+"\"")
	}
	return lines
}

// Banner returns an auto-generation banner comment or empty string.
func Banner(banner bool) string {
	if !banner {
		return ""
	}
	return fmt.Sprintf("\n/**\n * Auto Generate by github.com/stella-go/stella %s on %s.\n */\n", version.VERSION, time.Now().Format("2006/01/02"))
}

func FirstUpperCamelCase(s string) string {
	s = ToCamelCase(s)
	s = strings.ToUpper(s[0:1]) + s[1:]
	return s
}

func ToCamelCase(s string) string {
	return reCamel.ReplaceAllStringFunc(s, toUpper)
}

func ToSnakeCase(s string) string {
	snake := reUpper.ReplaceAllStringFunc(s, toSnake)
	return strings.Trim(snake, "_")
}

func ToStrikeCase(s string) string {
	s = strings.ReplaceAll(s, "_", "-")
	snake := reUpper.ReplaceAllStringFunc(s, toStrike)
	return strings.Trim(snake, "-")
}

func toUpper(s string) string {
	return strings.ToUpper(s[1:])
}

func toSnake(s string) string {
	return "_" + strings.ToLower(s[:1])
}

func toStrike(s string) string {
	return "-" + strings.ToLower(s[:1])
}
