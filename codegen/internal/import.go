package internal

import "go/ast"

type Import struct {
	SourceFile string
	InputAST   *ast.ImportSpec

	Alias string
	Path  string
}

func (i *Import) Init() {
	if i.InputAST.Name != nil {
		i.Alias = i.InputAST.Name.String()
	}
	i.Path = i.InputAST.Path.Value
}
