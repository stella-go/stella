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
	"strconv"
	"strings"

	"github.com/stella-go/stella/creator/proj"
	"github.com/stella-go/stella/generator/curd"
	"github.com/stella-go/stella/generator/model"
	"github.com/stella-go/stella/generator/parser"
	"github.com/stella-go/stella/generator/router"
	"github.com/stella-go/stella/generator/service"
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
	stella generate -i init.sql -o model 

`, version.VERSION)
		flageSet.PrintDefaults()
	}
	i := flageSet.String("i", "", "input sql file")
	sub := flageSet.String("sub", "", "sql subset")

	std := flageSet.Bool("std", false, "stdout print")
	o := flageSet.String("o", "", "output dictionary")
	f := flageSet.String("f", "", "output file name")
	p := flageSet.String("p", "", "package name")

	m := flageSet.Bool("m", true, "generate models")

	c := flageSet.Bool("c", true, "generate curd")
	asc := flageSet.String("asc", "", "order by")
	desc := flageSet.String("desc", "", "reverse order by")
	logic := flageSet.String("logic", "", "logic delete")
	round := flageSet.String("round", "s", "round time [s/ms/μs]")

	generateRouter := flageSet.Bool("router", false, "generate router")
	generateService := flageSet.Bool("service", false, "generate service")

	banner := flageSet.Bool("banner", true, "output banner")

	h := flageSet.Bool("h", false, "print help info")
	help := flageSet.Bool("help", false, "print help info")
	flageSet.Parse(os.Args[2:])
	if *h || *help {
		flageSet.Usage()
		return
	}
	generate(*p, *i, *sub, *o, *std, *f, *banner, *m, *c, *logic, *asc, *desc, *round, *generateRouter, *generateService)
}

func readFileWithStdin(input string, sub string) string {
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
	if sub != "" {
		split := strings.Split(strings.TrimSpace(sub), ",")
		start, end := 0, 0x7FFFFFFF
		if len(split) == 1 {
			n, err := strconv.Atoi(split[0])
			if err != nil {
				printError("parse subset error", err)
				return sql
			}
			start, end = n, n
		} else {
			if split[0] == "" {
				n, err := strconv.Atoi(split[1])
				if err != nil {
					printError("parse subset error", err)
					return sql
				}
				end = n
			} else if split[1] == "" {
				m, err := strconv.Atoi(split[0])
				if err != nil {
					printError("parse subset error", err)
					return sql
				}
				start = m
			} else {
				m, err := strconv.Atoi(split[0])
				if err != nil {
					printError("parse subset error", err)
					return sql
				}
				start = m
				n, err := strconv.Atoi(split[1])
				if err != nil {
					printError("parse subset error", err)
					return sql
				}
				end = n
			}

		}
		lines := strings.Split(sql, "\n")
		if start < 1 {
			start = 1
		}
		if end > len(lines)+1 {
			end = len(lines) + 1
		}
		lines = lines[start-1 : end-1]
		for i, line := range lines {
			fmt.Printf("%-4d %s\n", start+i, line)
		}
		sql = strings.Join(lines, "\n")
	}
	return sql
}

func generate(pkg string, input string, sub string, output string, std bool, file string, banner bool, m bool, c bool, logic string, asc string, desc string, round string, generateRouter bool, generateService bool) {
	sql := readFileWithStdin(input, sub)
	statements := parser.Parse(sql)
	if len(statements) == 0 {
		return
	}

	if generateRouter {
		p, f, o := fill(pkg, output, file, "router")
		filename := f + "_auto.go"
		content := router.Generate(p, statements, banner)
		writeFileTryFormat(std, o, filename, content)
	}

	if generateService {
		p, f, o := fill(pkg, output, file, "service")
		filename := f + "_auto.go"
		content := service.Generate(p, statements, banner)
		writeFileTryFormat(std, o, filename, content)
	}

	if m {
		p, f, o := fill(pkg, output, file, "model")
		filename := f + "_auto.go"
		content := model.Generate(p, statements, banner)
		writeFileTryFormat(std, o, filename, content)
	}

	if c {
		p, f, o := fill(pkg, output, file, "model")
		filename := f + "_curd_auto.go"
		content := curd.Generate(p, statements, banner, logic, asc, desc, round)
		writeFileTryFormat(std, o, filename, content)
	}
}

func fill(pkg string, output string, file string, defaultValue string) (string, string, string) {
	if output == "" {
		output = defaultValue
	}
	if file == "" {
		file = defaultValue
	}
	base := path.Base(output)
	if pkg == "" {
		pkg = base
	}
	return pkg, file, output
}

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

func writeFileTryFormat(std bool, output string, filename string, content string) {
	if std || output == "" || filename == "" {
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
	l := flageSet.String("l", "go", "projcet language")
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
	fmt.Fprintf(os.Stderr, message+": %v\n", err)
}
