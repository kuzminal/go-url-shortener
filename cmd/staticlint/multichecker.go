// staticlint для запуска проверок кода.
//
// Для того чтобы запустить проверку, сначала нужно собрать multichecker командой :
//
//	go build cmd/staticlint/multichecker.go
//
// Затем можно запускать собранный файл с указанием проверяемого пакета:
//
//	./multichecker ./././cmd/shortener/
package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

// OsExitCheckAnalyzer проверка на наличие os.Exit в функции main в пакете main.
var OsExitCheckAnalyzer = &analysis.Analyzer{
	Name: "errcheck",
	Doc:  "check for unchecked errors",
	Run:  run,
}

// run основная функция, которая вызывается при запуске линтера.
func run(pass *analysis.Pass) (interface{}, error) {
	expr := func(x *ast.FuncDecl) {
		// проверяем, что выражение представляет собой вызов функции,
		// у которой имя равно main
		if x.Name.Name == "main" {
			ast.Inspect(x.Body, func(node ast.Node) bool {
				switch ex := node.(type) {
				case *ast.SelectorExpr:
					if ex.Sel.Name == "Exit" {
						if ex.X.(*ast.Ident).Name == "os" {
							pass.Reportf(ex.Pos(), "os.Exit in main.main() is not allowed")
						}
					}
				}
				return true
			})

		}
	}

	for _, file := range pass.Files {
		// проверяю, что мы в файле main
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				switch ex := node.(type) {
				case *ast.FuncDecl:
					expr(ex)
				}
				return true
			})
		}
	}
	return nil, nil
}

func main() {
	mychecks := []*analysis.Analyzer{
		OsExitCheckAnalyzer,
		assign.Analyzer,
		shift.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	multichecker.Main(
		mychecks...,
	)
}
