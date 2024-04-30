package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit in main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	checkExpr := func(s *ast.SelectorExpr) {
		i, ok := s.X.(*ast.Ident)
		if !ok {
			return
		}
		if i.Name == "os" && s.Sel.Name == "Exit" {
			pass.Reportf(s.Pos(), "os.Exit not allowed in func main")
		}
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				return x.Name.Name == "main"
			case *ast.SelectorExpr:
				checkExpr(x)
			}
			return true
		})
	}

	return nil, nil
}
