//go:generate stella line -include="*.go" .
package main

import (
	"{{ project-name }}/router"

	"github.com/stella-go/siu"
)

func main() {
	siu.Route(&router.HelloRouter{})
	siu.Run()
}
