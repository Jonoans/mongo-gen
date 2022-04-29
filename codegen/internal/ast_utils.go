package internal

import (
	"fmt"
	"go/ast"
	"log"
	"reflect"
	"strings"
)

func astIdentSliceToString(exprs []*ast.Ident) string {
	exprsStr := make([]string, len(exprs))
	for i, expr := range exprs {
		exprsStr[i] = astObjectToString(expr)
	}
	return strings.Join(exprsStr, "")
}

func astObjectToString(expr ast.Expr) string {
	switch eval := expr.(type) {
	case *ast.Ident:
		return eval.Name
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", astObjectToString(eval.X))
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", astObjectToString(eval.X))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", eval.X, eval.Sel)
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", astObjectToString(eval.Len), astObjectToString(eval.Elt))
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", astObjectToString(eval.X), astObjectToString(eval.Index))
	case *ast.BasicLit:
		if eval == nil {
			return ""
		}
		return eval.Value
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", astObjectToString(eval.Key), astObjectToString(eval.Value))
	default:
		if eval != nil {
			log.Fatalf("need to implement support for: %s", reflect.TypeOf(eval).String())
		}
	}
	return ""
}
