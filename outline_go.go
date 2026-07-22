package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// parseGoOutline parses Go source code into structured symbols using go/parser & go/ast
func parseGoOutline(filename string, src []byte) ([]SymbolItem, error) {
	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var symbols []SymbolItem

	for _, decl := range fileAST.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			line := fset.Position(d.Pos()).Line
			kind := KindFunction
			name := d.Name.Name

			var recvStr string
			if d.Recv != nil && len(d.Recv.List) > 0 {
				kind = KindMethod
				recvType := d.Recv.List[0].Type
				recvStr = fmt.Sprintf("(%s) ", exprToString(recvType))
			}

			symbols = append(symbols, SymbolItem{
				Name:   name,
				Kind:   kind,
				Line:   line,
				Indent: 0,
				Text:   fmt.Sprintf("[%s] %s%s", kind, recvStr, name),
			})

		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					line := fset.Position(typeSpec.Pos()).Line
					typeName := typeSpec.Name.Name

					switch t := typeSpec.Type.(type) {
					case *ast.StructType:
						symbols = append(symbols, SymbolItem{
							Name:   typeName,
							Kind:   KindStruct,
							Line:   line,
							Indent: 0,
							Text:   fmt.Sprintf("[struct] %s", typeName),
						})
						// Extract struct fields
						if t.Fields != nil {
							for _, field := range t.Fields.List {
								fieldLine := fset.Position(field.Pos()).Line
								for _, fieldName := range field.Names {
									fieldType := exprToString(field.Type)
									symbols = append(symbols, SymbolItem{
										Name:   fieldName.Name,
										Kind:   KindField,
										Line:   fieldLine,
										Indent: 1,
										Text:   fmt.Sprintf("   ├── [field] %s: %s", fieldName.Name, fieldType),
									})
								}
							}
						}

					case *ast.InterfaceType:
						symbols = append(symbols, SymbolItem{
							Name:   typeName,
							Kind:   KindInterface,
							Line:   line,
							Indent: 0,
							Text:   fmt.Sprintf("[interface] %s", typeName),
						})
					}
				}
			}
		}
	}

	return symbols, nil
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", expr))
	}
}
