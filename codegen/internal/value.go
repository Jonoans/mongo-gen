package internal

import (
	"go/ast"
	"go/token"
)

type Interface struct {
	SourceFile string
	InputAST   *ast.GenDecl
	Name       string
}

type Value struct {
	SourceFile string
	InputTok   token.Token
	InputAST   *ast.ValueSpec
	Name       string
}

type Type struct {
	SourceFile string
}
