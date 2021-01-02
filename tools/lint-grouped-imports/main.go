package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const modPackageName = "github.com/joesonw/lte"

func main() {
	var files []string
	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return err
	})
	if err != nil {
		log.Fatalf("Failed to walk files: %s", err)
	}
	for _, filePath := range files {
		if filePath == "--" {
			continue
		}

		if strings.HasSuffix(filePath, ".pb.go") {
			continue
		}

		v := visitor{fset: token.NewFileSet()}
		f, err := parser.ParseFile(v.fset, filePath, nil, 0)
		if err != nil {
			log.Fatalf("Failed to parse file %s: %s", filePath, err)
		}

		ast.Walk(&v, f)
		if len(v.imports) == 0 {
			continue
		}

		throw := func() {
			fmt.Println(filePath)
			fmt.Println(v.String())
			fmt.Println("")
			fmt.Println("imports are grouped correctly. should be 3 groups, builtins, third-party and within module")
			os.Exit(1)
		}

		lastImport := v.imports[0]
		for _, il := range v.imports[1:] {
			lastStage := getStage(lastImport.path)
			lastLine := lastImport.line
			stage := getStage(il.path)
			line := il.line
			if stage < lastStage {
				throw()
			}
			if stage == lastStage && line != lastLine+1 {
				throw()
			}
			if stage > lastStage && line != lastLine+2 {
				throw()
			}
			lastImport = il
		}
	}
}

func getStage(name string) int {
	if strings.HasPrefix(name, modPackageName) {
		return 3
	}

	if strings.Contains(name, ".") {
		return 2
	}

	return 1
}

type importLine struct {
	name string
	path string
	line int
	node *ast.ImportSpec
}

type visitor struct {
	fset    *token.FileSet
	imports []importLine
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	importSpec, ok := node.(*ast.ImportSpec)
	if ok {
		name := ""
		if importSpec.Name != nil {
			name = importSpec.Name.Name
		}
		path, _ := strconv.Unquote(importSpec.Path.Value)
		v.imports = append(v.imports, importLine{
			name: name,
			path: path,
			line: v.fset.File(importSpec.Path.Pos()).Line(importSpec.Pos()),
			node: importSpec,
		})
	}

	return v
}

func (v *visitor) String() string {
	if len(v.imports) == 0 {
		return ""
	}
	lines := []string{"import ("}
	lastLine := v.imports[0].line
	for _, il := range v.imports {
		for il.line > (lastLine + 1) {
			lastLine++
			lines = append(lines, "")
		}
		lastLine = il.line
		line := "    "
		if il.name != "" {
			line += il.name + " "
		}
		line += strconv.Quote(il.path)
		lines = append(lines, line)
	}
	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}
