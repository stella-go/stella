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

package cmd

import (
	"bufio"
	"flag"
	"fmt"
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
	flagSet := flag.NewFlagSet("stella generate", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, `
stella An efficient development tool. %s

Usage: 
	stella generate -i init.sql -o model 

`, version.VERSION)
		flagSet.PrintDefaults()
	}
	i := flagSet.String("i", "", "input sql file")
	sub := flagSet.String("sub", "", "sql subset")

	std := flagSet.Bool("std", false, "stdout print")
	o := flagSet.String("o", "", "output dictionary")
	f := flagSet.String("f", "", "output file name")
	p := flagSet.String("p", "", "package name")

	m := flagSet.Bool("m", true, "generate models")
	gorm := flagSet.Bool("gorm", false, "models with gorm tags")

	c := flagSet.Bool("curd", false, "generate curd")
	asc := flagSet.String("asc", "", "order by")
	desc := flagSet.String("desc", "", "reverse order by")
	logic := flagSet.String("logic", "", "logic delete")
	round := flagSet.String("round", "s", "round time [s/ms/μs]")

	generateRouter := flagSet.Bool("router", false, "generate router")
	generateService := flagSet.Bool("service", false, "generate service")

	panicStyle := flagSet.Bool("panic", false, "panic style")

	banner := flagSet.Bool("banner", true, "output banner")

	h := flagSet.Bool("h", false, "print help info")
	help := flagSet.Bool("help", false, "print help info")
	flagSet.Parse(os.Args[2:])
	if *h || *help {
		flagSet.Usage()
		return
	}
	generate(*p, *i, *sub, *o, *std, *f, *banner, *m, *gorm, *c, *logic, *asc, *desc, *round, *generateRouter, *generateService, *panicStyle)
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
		sqlBytes, err := os.ReadFile(input)
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

func generate(pkg string, input string, sub string, output string, std bool, file string, banner bool, m bool, gorm bool, c bool, logic string, asc string, desc string, round string, generateRouter bool, generateService bool, panicStyle bool) {
	sql := readFileWithStdin(input, sub)
	statements := parser.Parse(sql)
	if len(statements) == 0 {
		return
	}

	if generateRouter {
		{
			p, f, o := fill(pkg, output, file, "router")
			filename := f + "_auto.go"
			if f != "router" {
				filename = f + "_router_auto.go"
			}
			content := func() string {
				if panicStyle {
					return router.GeneratePanic(p, f, statements, banner)
				} else {
					return router.Generate(p, f, statements, banner)
				}
			}()
			writeFileTryFormat(std, o, filename, content)
		}
		{
			_, f, o := fill(pkg, output, file, "doc")
			filename := f + "_auto.md"
			if f != "doc" {
				filename = f + "_doc_auto.md"
			}
			content := router.GenerateDoc(statements, banner)
			writeFileTryFormat(std, o, filename, content)
		}
	}

	if generateService {
		p, f, o := fill(pkg, output, file, "service")
		filename := f + "_auto.go"
		if f != "service" {
			filename = f + "_service_auto.go"
		}
		content := func() string {
			if gorm {
				return service.GenerateGorm(p, f, statements, banner)
			} else {
				if panicStyle {
					return service.GeneratePanic(p, f, statements, banner)
				} else {
					return service.Generate(p, f, statements, banner)
				}
			}
		}()
		writeFileTryFormat(std, o, filename, content)
	}

	if m {
		p, f, o := fill(pkg, output, file, "model")
		filename := f + "_auto.go"
		if f != "model" {
			filename = f + "_model_auto.go"
		}
		content := model.Generate(p, statements, banner, gorm)
		writeFileTryFormat(std, o, filename, content)
	}

	if c {
		p, f, o := fill(pkg, output, file, "model")
		filename := f + "_curd_auto.go"
		if f != "model" {
			filename = f + "_model_curd_auto.go"
		}
		content := func() string {
			if panicStyle {
				return curd.GeneratePanic(p, statements, banner, logic, asc, desc, round)
			} else {
				return curd.Generate(p, statements, banner, logic, asc, desc, round)
			}
		}()
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
		exist, err = isExists(fullPath)
		if err != nil {
			printError("read outputh file error", err)
			fmt.Println(content)
			return
		}
		if exist {
			fmt.Printf("output file \"%s\" exists, do you want to remove it? [y/N] ", fullPath)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(answer)
			if answer == "" || answer == "N" || answer == "n" {
				fmt.Println(content)
				return
			}
			if answer == "Y" || answer == "y" {
				err := os.RemoveAll(fullPath)
				if err != nil {
					printError("remove outputh path error", err)
					fmt.Println(content)
					return
				}
			}
		}

		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			printError("write file error", err)
			fmt.Println(content)
			return
		}
		bts, err := gofmt.Run(fullPath, false)
		if err == nil && bts != nil {
			os.WriteFile(fullPath, bts, 0644)
		}
	}
}

func Create() {
	flagSet := flag.NewFlagSet("stella create", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, `
stella An efficient development tool. %s
Usage: 
	stella create -n my-project

`, version.VERSION)
		flagSet.PrintDefaults()
	}
	l := flagSet.String("l", "go", "projcet language")
	t := flagSet.String("t", "server", "project type [server/sdk]")
	n := flagSet.String("n", "demo", "project name")
	o := flagSet.String("o", ".", "output dictionary")

	h := flagSet.Bool("h", false, "print help info")
	help := flagSet.Bool("help", false, "print help info")
	flagSet.Parse(os.Args[2:])
	if *h || *help {
		flagSet.Usage()
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
	if exist {
		fmt.Printf("project path \"%s\" exists, do you want to remove it? [y/N] ", projDir)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer == "" || answer == "N" || answer == "n" {
			return
		}
		if answer == "Y" || answer == "y" {
			err := os.RemoveAll(projDir)
			if err != nil {
				printError("remove outputh path error", err)
				return
			}
		}
	} else {
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
	flagSet := flag.NewFlagSet("stella line", flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, `
stella An efficient development tool. %s
Usage: 
	stella line [path/to [path/to ...]]

	stella line, By default it is equivalent to "stella line ."
`, version.VERSION)
		flagSet.PrintDefaults()
	}

	h := flagSet.Bool("h", false, "print help info")
	help := flagSet.Bool("help", false, "print help info")
	ignore := flagSet.String("ignore", "", "ignore file patterns")
	include := flagSet.String("include", "*.*", "include file patterns")
	s := flagSet.Bool("s", false, "use file short name")

	flagSet.Parse(os.Args[2:])
	if *h || *help {
		flagSet.Usage()
		return
	}
	ignores := strings.Split(*ignore, ",")
	includes := strings.Split(*include, ",")
	args := flagSet.Args()
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
