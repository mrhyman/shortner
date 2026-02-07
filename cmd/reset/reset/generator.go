package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

type StructInfo struct {
	Name   string
	Fields []*ast.Field
}

func Generate(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}

		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		for _, pkg := range pkgs {
			structs := findResetStructs(pkg)
			if len(structs) > 0 {
				if err := generateFile(path, pkg.Name, structs); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func findResetStructs(pkg *ast.Package) []StructInfo {
	var result []StructInfo

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			if !hasResetComment(gen.Doc) {
				continue
			}

			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}

				result = append(result, StructInfo{
					Name:   ts.Name.Name,
					Fields: st.Fields.List,
				})
			}
		}
	}

	return result
}

func hasResetComment(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if c.Text == "// generate:reset" {
			return true
		}
	}
	return false
}

func generateResetMethod(buf *bytes.Buffer, s StructInfo) {
	fmt.Fprintf(buf, "func (x *%s) Reset() {\n", s.Name)
	buf.WriteString("\tif x == nil {\n\t\treturn\n\t}\n\n")

	for _, f := range s.Fields {
		for _, name := range f.Names {
			generateFieldReset(buf, name.Name, f.Type)
		}
	}

	buf.WriteString("}\n\n")
}

func generateFieldReset(buf *bytes.Buffer, name string, expr ast.Expr) {
	switch t := expr.(type) {

	case *ast.Ident:
		fmt.Fprintf(buf, "\tx.%s = %s\n", name, zeroValue(t.Name))

	case *ast.ArrayType: // slice
		fmt.Fprintf(buf, "\tx.%s = x.%s[:0]\n", name, name)

	case *ast.MapType:
		fmt.Fprintf(buf, "\tclear(x.%s)\n", name)

	case *ast.StarExpr:
		buf.WriteString(fmt.Sprintf("\tif x.%s != nil {\n", name))
		generatePointerReset(buf, name, t.X)
		buf.WriteString("\t}\n")

	default:
		// вложенная структура с Reset()
		buf.WriteString(fmt.Sprintf(
			"\tif r, ok := any(x.%s).(interface{ Reset() }); ok {\n\t\tr.Reset()\n\t}\n",
			name,
		))
	}
}

func generatePointerReset(buf *bytes.Buffer, name string, expr ast.Expr) {
	switch t := expr.(type) {
	case *ast.Ident:
		fmt.Fprintf(buf, "\t\t*x.%s = %s\n", name, zeroValue(t.Name))
	default:
		buf.WriteString(fmt.Sprintf(
			"\t\tif r, ok := any(x.%s).(interface{ Reset() }); ok {\n\t\t\tr.Reset()\n\t\t}\n",
			name,
		))
	}
}

func zeroValue(t string) string {
	switch t {
	case "int", "int64", "uint", "uint64", "float32", "float64":
		return "0"
	case "string":
		return `""`
	case "bool":
		return "false"
	case "error":
		return "nil"
	default:
		return "nil"
	}
}

func generateFile(dir, pkg string, structs []StructInfo) error {
	var buf bytes.Buffer

	buf.WriteString("// GENERATED CODE! DO NOT EDIT.\n\n")
	buf.WriteString("package " + pkg + "\n\n")

	for _, s := range structs {
		generateResetMethod(&buf, s)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	return os.WriteFile(
		filepath.Join(dir, "reset.gen.go"),
		src,
		0644,
	)
}
