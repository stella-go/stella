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

package generator

import (
	"regexp"
	"strings"
)

func FirstUpperCamelCase(s string) string {
	s = ToCamelCase(s)
	s = strings.ToUpper(s[0:1]) + s[1:]
	return s
}

func ToCamelCase(s string) string {
	re := regexp.MustCompile(`_(\w)`)
	return re.ReplaceAllStringFunc(s, ToUpper)
}

func ToSnakeCase(s string) string {
	re := regexp.MustCompile(`[A-Z]`)
	snake := re.ReplaceAllStringFunc(s, ToSnake)
	return strings.Trim(snake, "_")
}

func ToUpper(s string) string {
	return strings.ToUpper(s[1:])
}

func ToSnake(s string) string {
	return "_" + strings.ToLower(s[:1])
}
