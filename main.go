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

//go:generate statik -src=./templates -f -include=*.*
package main

import (
	"fmt"
	"os"

	"github.com/stella-go/stella/cmd"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	command := os.Args[1]
	switch command {
	case "generate":
		cmd.Generate()
	case "create":
		cmd.Create()
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `stella An efficient development tool

Usage: 

	sub-commands:
		generate	Generate template code.
		create		Create template project.

	stella <command> -h for more info.
`)
}
