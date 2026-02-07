package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "paniclint",
	Doc:  "forbids panic usage and log.Fatal/os.Exit outside main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// panic()
			if isPanic(call) {
				pass.Reportf(call.Pos(), "panic() usage is forbidden")
				return true
			}

			// log.Fatal / os.Exit
			if isFatalOrExit(call) {
				if !isInsideMain(pass, call) {
					pass.Reportf(call.Pos(), "log.Fatal/os.Exit is allowed only in main.main")
				}
			}

			return true
		})
	}

	return nil, nil
}
