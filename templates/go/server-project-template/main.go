//go:generate stella line
package main

import (
	"{{ project-name }}/rotater"

	"github.com/stella-go/siu"
)

func main() {
	siu.Rotate(&rotater.HelloRotate{})
	siu.Run()
}
