package internal

import "go/ast"

type CustomType struct {
	SourceFile string
	Name       string
	InputAST   *ast.TypeSpec
}
