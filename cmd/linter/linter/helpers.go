package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func isPanic(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "panic"
}

func isFatalOrExit(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return (pkgIdent.Name == "log" && sel.Sel.Name == "Fatal") ||
		(pkgIdent.Name == "os" && sel.Sel.Name == "Exit")
}

func isInsideMain(pass *analysis.Pass, call *ast.CallExpr) bool {
	// пакет должен быть main
	if pass.Pkg.Name() != "main" {
		return false
	}

	// ищем enclosing функцию
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			if fn.Name.Name != "main" {
				continue
			}

			if call.Pos() > fn.Body.Pos() && call.End() < fn.Body.End() {
				return true
			}
		}
	}

	return false
}
