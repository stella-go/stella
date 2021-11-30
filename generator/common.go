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
