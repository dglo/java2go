package parser

import "go/ast"

func makeField(name string, texpr ast.Expr) *ast.Field {
	fname := make([]*ast.Ident, 0)
	if name != "" {
		fname = append(fname, ast.NewIdent(name))
	}

	return &ast.Field{Names: fname, Type: texpr}
}
