package internal

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

var (
	structHookMethodNames = []string{
		"Queried",
		"Creating", "Created",
		"Saving", "Saved",
		"Updating", "Updated",
		"Deleting", "Deleted",
	}
	structDatabaseMethodNames = []string{
		"Create", "Update", "Delete",
		"CreateWithCtx", "UpdateWithCtx", "DeleteWithCtx",
	}
	reservedMethodNames    = []string{"CollectionName"}
	reservedFuncValueNames = []string{}
)

func InitReservedValues(reservedFuncValues []string) {
	// Initialise reserved function and value names
	reservedFuncValueNames = append(reservedFuncValueNames, reservedFuncValues...)
	reservedMethodNames = append(reservedMethodNames, structHookMethodNames...)

	// Initialiase reserved method names
	reservedMethodNames = append(reservedMethodNames, structDatabaseMethodNames...)
}

func isCollectionNameMethod(f *ast.FuncDecl) bool {
	if astObjectToString(f.Name) != "CollectionName" {
		return false
	}

	if len(f.Type.Params.List) != 0 {
		return false
	}

	if f.Type.Results == nil || len(f.Type.Results.List) != 1 {
		return false
	}

	if astObjectToString(f.Type.Results.List[0].Type) != "string" {
		return false
	}

	return true
}

func isHookMethod(f *ast.FuncDecl) bool {
	found, funcName := false, astObjectToString(f.Name)
	for _, hookMethodName := range structHookMethodNames {
		if funcName == hookMethodName {
			found = true
			break
		}
	}

	if !found {
		return false
	}

	if len(f.Type.Params.List) != 0 {
		return false
	}

	if f.Type.Results == nil || len(f.Type.Results.List) != 1 {
		return false
	}

	if astObjectToString(f.Type.Results.List[0].Type) != "error" {
		return false
	}

	return true
}

// Unused
func isDatabaseMethod(f *ast.FuncDecl) bool {
	found, funcName := false, astObjectToString(f.Name)
	for _, databaseMethodName := range structDatabaseMethodNames {
		if funcName == databaseMethodName {
			found = true
			break
		}
	}

	if !found {
		return false
	}

	if len(f.Type.Params.List) != 0 {
		return false
	}

	if f.Type.Results == nil || len(f.Type.Results.List) != 1 {
		return false
	}

	if astObjectToString(f.Type.Results.List[0].Type) != "error" {
		return false
	}

	return true
}

func isFuncValueNameReserved(funcValueName string) bool {
	for _, reserved := range reservedFuncValueNames {
		if funcValueName == reserved {
			return true
		}
	}
	return false
}

func isMethodNameReserved(funcName string) bool {
	for _, reserved := range reservedMethodNames {
		if funcName == reserved {
			return true
		}
	}
	return false
}

func isMethodNameResolver(funcName string) bool {
	return strings.HasPrefix(funcName, "GetResolved_")
}

func buildCollectionNameMethod(s *Struct) *Func {
	f := &Func{SourceFile: s.SourceFile, Name: "CollectionName"}
	f.Parent = s

	// Function signature
	f.InputAST = &ast.FuncDecl{}
	f.InputAST.Recv = &ast.FieldList{}
	f.InputAST.Recv.List = []*ast.Field{{Type: &ast.StarExpr{X: &ast.Ident{Name: s.Name}}}}
	f.InputAST.Name = ast.NewIdent(f.Name)
	f.InputAST.Type = &ast.FuncType{}
	f.InputAST.Type.Results = &ast.FieldList{}
	f.InputAST.Type.Results.List = []*ast.Field{{Type: ast.NewIdent("string")}}

	// Function Body
	collectionName := strcase.ToLowerCamel(s.Name)
	f.InputAST.Body = &ast.BlockStmt{}
	f.InputAST.Body.List = []ast.Stmt{
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(collectionName),
				},
			},
		},
	}

	return f
}

func buildDatabaseMethod(s *Struct, dbMethodInfo *structDbMethod) *Func {
	funcParams := []*ast.Field{}
	for _, param := range dbMethodInfo.params {
		if param.paramName == "" {
			continue
		}

		funcParams = append(funcParams, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(param.paramName)},
			Type:  ast.NewIdent(param.paramType),
		})
	}

	funcArgs := []ast.Expr{}
	for _, param := range dbMethodInfo.params {
		funcArgs = append(funcArgs, ast.NewIdent(param.argUsage))
	}

	f := &Func{SourceFile: s.SourceFile, Name: dbMethodInfo.name}
	f.Parent = s

	// Function signature
	f.InputAST = &ast.FuncDecl{}
	f.InputAST.Recv = &ast.FieldList{}
	f.InputAST.Recv.List = []*ast.Field{{Names: []*ast.Ident{ast.NewIdent("m")}, Type: &ast.StarExpr{X: &ast.Ident{Name: s.Name}}}}
	f.InputAST.Name = ast.NewIdent(f.Name)
	f.InputAST.Type = &ast.FuncType{}
	f.InputAST.Type.Params = &ast.FieldList{}
	f.InputAST.Type.Params.List = funcParams
	f.InputAST.Type.Results = &ast.FieldList{}
	f.InputAST.Type.Results.List = []*ast.Field{{Type: ast.NewIdent("error")}}

	// Function Body
	f.InputAST.Body = &ast.BlockStmt{}
	f.InputAST.Body.List = []ast.Stmt{
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.CallExpr{
					Fun:  ast.NewIdent(dbMethodInfo.globalFuncName),
					Args: funcArgs,
				},
			},
		},
	}

	return f
}

func buildHookMethod(s *Struct, methodName string) *Func {
	f := &Func{SourceFile: s.SourceFile, Name: methodName}
	f.Parent = s

	// Function signature
	f.InputAST = &ast.FuncDecl{}
	f.InputAST.Recv = &ast.FieldList{}
	f.InputAST.Recv.List = []*ast.Field{{Names: []*ast.Ident{ast.NewIdent("m")}, Type: &ast.StarExpr{X: &ast.Ident{Name: s.Name}}}}
	f.InputAST.Name = ast.NewIdent(f.Name)
	f.InputAST.Type = &ast.FuncType{}
	f.InputAST.Type.Results = &ast.FieldList{}
	f.InputAST.Type.Results.List = []*ast.Field{{Type: ast.NewIdent("error")}}

	// Function Body
	f.InputAST.Body = &ast.BlockStmt{}
	f.InputAST.Body.List = []ast.Stmt{
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent("nil"),
			},
		},
	}

	return f
}
