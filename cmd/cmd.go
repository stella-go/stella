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

package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/stella-go/stella/creator/proj"
	"github.com/stella-go/stella/generator/curd"
	"github.com/stella-go/stella/generator/model"
	"github.com/stella-go/stella/generator/parser"
	"github.com/stella-go/stella/gofmt"
	"github.com/stella-go/stella/line"
	"github.com/stella-go/stella/version"
)

func Generate() {
	flageSet := flag.NewFlagSet("stella generate", flag.ExitOnError)
	flageSet.Usage = func() {
		fmt.Fprintf(os.Stderr, `
stella An efficient development tool. %s

Usage: 
	stella generate -p model -i init.sql -o model -f model

`, version.VERSION)
		flageSet.PrintDefaults()
	}
	p := flageSet.String("p", "model", "package name")
	i := flageSet.String("i", "", "input sql file")
	o := flageSet.String("o", "", "output dictionary")
	f := flageSet.String("f", "", "output file name")
	banner := flageSet.Bool("banner", true, "output banner")
	m := flageSet.Bool("m", false, "only generate models")
	logic := flageSet.String("logic", "", "logic delete")
	desc := flageSet.String("desc", "", "reverse order by")
	h := flageSet.Bool("h", false, "print help info")
	help := flageSet.Bool("help", false, "print help info")
	flageSet.Parse(os.Args[2:])
	if *h || *help {
		flageSet.Usage()
		return
	}
	generate(*p, *i, *o, *f, *banner, *m, *logic, *desc)
}

func readFileWithStdin(input string) string {
	sql := ""
	if input == "" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text == "EOF" {
				break
			}
			sql += text
		}
	} else {
		sqlBytes, err := ioutil.ReadFile(input)
		if err != nil {
			printError("read sql file error", err)
			return ""
		}
		sql = string(sqlBytes)
	}
	return sql
}

func generate(pkg string, input string, output string, file string, banner bool, onlymodel bool, logic string, desc string) {
	if file == "" {
		file = pkg
	}
	sql := readFileWithStdin(input)
	statements := parser.Parse(sql)

	filename := file + "_auto.go"
	content := model.Generate(pkg, statements, banner)
	writeFileTryFormat(output, filename, content)

	if !onlymodel {
		filename = pkg + "_auto_curd.go"
		content = curd.Generate(pkg, statements, banner, logic, desc)
		writeFileTryFormat(output, filename, content)
	}
}

// func generateModel(pkg string, input string, output string) {
// 	filename := pkg + ".go"
// 	sql := readFileWithStdin(input)
// 	statements := parser.Parse(sql)
// 	content := model.Generate(pkg, statements)
// 	writeFileTryFormat(output, filename, content)
// }

// func generateCURD(pkg string, input string, output string) {
// 	filename := pkg + "_curd.go"
// 	sql := readFileWithStdin(input)
// 	statements := parser.Parse(sql)
// 	content := curd.Generate(pkg, statements)
// 	writeFileTryFormat(output, filename, content)
// }

func isExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func writeFileTryFormat(output string, filename string, content string) {
	if output == "" || filename == "" {
		fmt.Println(content)
	} else {
		exist, err := isExists(output)
		if err != nil {
			printError("read outputh path error", err)
			fmt.Println(content)
			return
		}
		if !exist {
			err := os.MkdirAll(output, 0755)
			if err != nil {
				printError("create outputh path error", err)
				fmt.Println(content)
				return
			}
		}
		fullPath := path.Join(output, filename)
		err = ioutil.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			printError("write file error", err)
			fmt.Println(content)
			return
		}
		bts, err := gofmt.Run(fullPath, false)
		if err == nil && bts != nil {
			ioutil.WriteFile(fullPath, bts, 0644)
		}
	}
}

func Create() {
	flageSet := flag.NewFlagSet("stella create", flag.ExitOnError)
	flageSet.Usage = func() {
		fmt.Fprintf(os.Stderr, `
stella An efficient development tool. %s
Usage: 
	stella create -n my-project

`, version.VERSION)
		flageSet.PrintDefaults()
	}
	l := flageSet.String("l", "go", "projcet language [go/java/node]")
	t := flageSet.String("t", "server", "project type [server/sdk]")
	n := flageSet.String("n", "demo", "project name")
	o := flageSet.String("o", ".", "output dictionary")

	h := flageSet.Bool("h", false, "print help info")
	help := flageSet.Bool("help", false, "print help info")
	flageSet.Parse(os.Args[2:])
	if *h || *help {
		flageSet.Usage()
		return
	}

	createProj(*l, *t, *n, *o)
}

func createProj(language string, stype string, name string, output string) {
	switch language {
	case "go":
		stype = "server"
	case "node":
		stype = "sdk"
	}
	projDir := fmt.Sprintf("%s/%s", output, name)
	exist, err := isExists(projDir)
	if err != nil {
		printError("read project path error", err)
		return
	}
	if !exist {
		err := os.MkdirAll(projDir, 0755)
		if err != nil {
			printError("create project path error", err)
			return
		}
	}
	err = proj.Create(language, stype, name, output)
	if err != nil {
		printError("create project error", err)
	}
}

func Line() {
	flageSet := flag.NewFlagSet("stella line", flag.ContinueOnError)
	flageSet.Usage = func() {
		fmt.Fprintf(os.Stderr, `
stella An efficient development tool. %s
Usage: 
	stella line [path/to [path/to ...]]

	stella line, By default it is equivalent to "stella line ."
`, version.VERSION)
		flageSet.PrintDefaults()
	}

	h := flageSet.Bool("h", false, "print help info")
	help := flageSet.Bool("help", false, "print help info")
	ignore := flageSet.String("ignore", "", "ignore file patterns")
	include := flageSet.String("include", "*.*", "include file patterns")
	s := flageSet.Bool("s", false, "use file short name")

	flageSet.Parse(os.Args[2:])
	if *h || *help {
		flageSet.Usage()
		return
	}
	ignores := strings.Split(*ignore, ",")
	includes := strings.Split(*include, ",")
	args := flageSet.Args()
	roots := make([]string, 0)
	if len(args) < 1 {
		roots = append(roots, ".")
	} else {
		roots = append(roots, args...)
	}
	fillLine(roots, includes, ignores, *s)
}

func fillLine(roots []string, includes []string, ignores []string, shortName bool) {
	for _, root := range roots {
		err := line.Fill(root, includes, ignores, shortName)
		if err != nil {
			printError("fill line error", err)
		}
	}
}

func printError(message string, err error) {
	fmt.Fprintf(os.Stderr, message+": %v", err)
}
