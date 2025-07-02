package internal

import (
	"go/types"
	"reflect"
	"strings"
)

func isMap(t types.Type) bool {
	_, ok := t.(*types.Map)
	return ok
}

func isPointer(t types.Type) bool {
	_, ok := t.(*types.Pointer)
	return ok
}

func isSlice(t types.Type) bool {
	_, ok := t.(*types.Slice)
	return ok
}

func getElement(t types.Type) types.Type {
	switch x := t.(type) {
	case *types.Map:
		return x.Elem()
	case *types.Pointer:
		return x.Elem()
	case *types.Slice:
		return x.Elem()
	default:
		return t
	}
}

func isBuiltin(t types.Type) (types.Type, bool) {
	switch x := t.(type) {
	case *types.Basic:
		return t, true
	case *types.Map:
		return isBuiltin(x.Elem())
	case *types.Pointer:
		return isBuiltin(x.Elem())
	case *types.Slice:
		return isBuiltin(x.Elem())
	case *types.Struct:
		return t, false
	default:
		_, builtin := isBuiltin(t.Underlying())
		return t, builtin
	}
}

func structTagContainsMongogenFalse(f *Field) bool {
	tag := reflect.StructTag(f.StructTag).Get("mongogen")
	tags := strings.Split(tag, ",")
	for _, t := range tags {
		if strings.TrimSpace(t) == "false" {
			return true
		}
	}
	return false
}

func isBaseModel(t types.Type) bool {
	if typeStr := t.String(); typeStr == "github.com/jonoans/mongo-gen/codegen.BaseModel" {
		return true
	}
	return false
}

func isTime(t types.Type) bool {
	if typeStr := t.String(); typeStr == "time.Time" {
		return true
	}
	return false
}
