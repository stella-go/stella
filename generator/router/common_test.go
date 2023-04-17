package router

import (
	"testing"

	"github.com/stella-go/stella/generator/parser"
)

func TestGenerateDoc(t *testing.T) {
	s := parser.Parse(sql)
	file := GenerateDoc(s, true)
	t.Log(file)
}
