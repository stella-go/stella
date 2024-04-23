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
	return re.ReplaceAllStringFunc(s, toUpper)
}

func ToSnakeCase(s string) string {
	re := regexp.MustCompile(`[A-Z]`)
	snake := re.ReplaceAllStringFunc(s, toSnake)
	return strings.Trim(snake, "_")
}

func ToStrikeCase(s string) string {
	s = strings.ReplaceAll(s, "_", "-")
	re := regexp.MustCompile(`[A-Z]`)
	snake := re.ReplaceAllStringFunc(s, toStrike)
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
