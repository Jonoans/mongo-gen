package internal

import (
	"go/ast"
	"go/token"
)

type Func struct {
	Parent     *Struct
	SourceFile string
	InputAST   *ast.FuncDecl
	FileSet    *token.FileSet

	Name      string
	Generated bool
}
