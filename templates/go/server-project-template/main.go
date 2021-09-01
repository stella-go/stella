//go:generate stella line
package main

import (
	"{{ project-name }}/router"

	"github.com/stella-go/siu"
)

func main() {
	siu.Route(&router.HelloRouter{})
	siu.Run()
}
