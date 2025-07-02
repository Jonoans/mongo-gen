package internal

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"regexp"
	"strings"
)

var (
	mapKeyRegex    = regexp.MustCompile(`^map\[(.*?)\]`)
	sliceTypeRegex = regexp.MustCompile(`^\[.*?\](.*)`)
	letters        = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type ResolverFieldReferences struct {
	Level  int
	InLoop bool
	// Where to assign results
	AssignmentVar string
	// Where to obtain ObjectID reference
	IDReferenceVar string

	// Standard References
	RootField     string
	ErrorField    string
	InitBoolField string
	ResolvedField string
}

type Field struct {
	Parent        *Struct
	InputAST      *ast.Field
	InputTypesVar *types.Var

	ParentField *Field
	ChildField  *Field

	Name       string
	Type       string
	StructTag  string
	MapKeyType string

	OwnType      types.Type
	ResolvedType types.Type

	IsBaseModelDerivative bool
	IsBuiltIn             bool
	IsEmbedded            bool
	IsReference           bool

	IsMap     bool
	IsPointer bool
	IsSlice   bool

	IsResolvable bool
	References   *ResolverFieldReferences
}

func getAlphabetLetter(i int) string {
	if i > len(letters) {
		log.Fatal("No. of levels > length of letters, your models are too deep")
	}
	return string(letters[i])
}

func (f *Field) Init() {
	f.Name = astIdentSliceToString(f.InputAST.Names)
	f.Type = astObjectToString(f.InputAST.Type)
	f.StructTag = astObjectToString(f.InputAST.Tag)

	f.OwnType = f.InputTypesVar.Type().Underlying()
	f.ResolvedType, f.IsBuiltIn = isBuiltin(f.OwnType)
	f.IsEmbedded = f.InputTypesVar.Embedded()

	if !f.IsBuiltIn {
		f.IsBaseModelDerivative = !structTagContainsMongogenFalse(f) && isBaseModel(f.OwnType)
		if f.IsBaseModelDerivative && !f.IsEmbedded {
			log.Fatal("BaseModel must be embedded")
		}

		if !f.IsBaseModelDerivative && !f.IsEmbedded && !isTime(f.ResolvedType) {
			f.IsReference = true
		}

		f.IsMap = isMap(f.OwnType)
		f.IsPointer = isPointer(f.OwnType)
		f.IsSlice = isSlice(f.OwnType)

		if f.IsMap {
			match := mapKeyRegex.FindStringSubmatchIndex(f.Type)
			f.MapKeyType = f.Type[match[2]:match[3]]
		}
	}

	if f.IsEmbedded && !structTagContainsMongogenFalse(f) && !f.IsBaseModelDerivative {
		f.IsBaseModelDerivative = f.checkEmbeddedIsBaseModelDerivative(f.OwnType)
	}
}

func (f *Field) checkEmbeddedIsBaseModelDerivative(currType types.Type) bool {
	underlying := currType.Underlying().(*types.Struct)
	for i := 0; i < underlying.NumFields(); i++ {
		field := underlying.Field(i)
		ownType := field.Type()

		if isBaseModel(ownType) {
			return true
		}

		if field.Embedded() {
			if f.checkEmbeddedIsBaseModelDerivative(ownType) {
				return true
			}
		}
	}

	return false
}

func (f *Field) getParent() *Field {
	switch f.ParentField {
	case nil:
		return f
	default:
		return f.ParentField.getParent()
	}
}

// ********** SECTION Child Field ********** //
func (f *Field) CreateChildField() {
	if f.OwnType != f.ResolvedType {
		f.initOwnChildField()
		f.ChildField.CreateChildField()
	}
}

func (f *Field) initOwnChildField() {
	elemType := getElement(f.OwnType)
	f.ChildField = &Field{}

	// Set parent fields
	f.ChildField.Parent = f.Parent
	f.ChildField.ParentField = f

	// Fill type information
	f.ChildField.OwnType = elemType
	f.ChildField.ResolvedType = f.ResolvedType

	// Fill flags
	f.ChildField.IsMap = isMap(elemType)
	f.ChildField.IsPointer = isPointer(elemType)
	f.ChildField.IsSlice = isSlice(elemType)

	// Fill type strings
	if f.IsMap {
		match := mapKeyRegex.FindStringSubmatchIndex(f.Type)
		f.MapKeyType = f.Type[match[2]:match[3]]
		f.ChildField.Type = f.Type[match[1]:len(f.Type)]
		return
	}

	if f.IsPointer {
		f.ChildField.Type = f.Type[1:]
		return
	}

	if f.IsSlice {
		match := sliceTypeRegex.FindStringSubmatchIndex(f.Type)
		f.ChildField.Type = f.Type[match[2]:len(f.Type)]
		return
	}
}

// ********** SECTION Resolvable Field ********** //
func (f *Field) CreateStubResolvableFields() []*Field {
	if !f.IsResolvable || f.References == nil {
		return nil
	}

	errorField, initBoolField, resolvedField := "err"+f.Name, "init"+f.Name, "resolved"+f.Name
	fields := []*Field{
		{
			Name: errorField,
			Type: "error",
		},
		{
			Name: initBoolField,
			Type: "bool",
		},
		{
			Name: resolvedField,
		},
	}

	currField := f
	for {
		if currField.IsMap {
			fields[2].Type += "map[" + currField.MapKeyType + "]"
		} else if currField.IsPointer {
			fields[2].Type += "*"
		} else if currField.IsSlice {
			fields[2].Type += "[]"
		} else {
			fields[2].Type += currField.Type
		}

		if currField.ChildField != nil {
			currField = currField.ChildField
			continue
		}
		break
	}

	return fields
}

func (f *Field) MutateResolvableFieldType() {
	newType := ""
	currField := f
	for {
		if currField.IsMap {
			newType += "map[" + currField.MapKeyType + "]"
		} else if currField.IsPointer {
			newType += "*"
		} else if currField.IsSlice {
			newType += "[]"
		}

		if currField.ChildField == nil {
			newType += "bson.ObjectID"
			break
		}

		if currField.IsBuiltIn {
			newType += currField.Type
		}

		currField = currField.ChildField
		continue
	}

	f.Type = newType
}

// ********** SECTION Resolver Function ********** //
func (f *Field) createResolverFieldReferences() {
	rootFieldName := f.getParent().Name
	f.References = &ResolverFieldReferences{}

	if f.ParentField == nil {
		f.References.Level = 0
	} else {
		f.References.Level = f.ParentField.References.Level + 1
	}

	f.References.RootField = "m." + rootFieldName
	f.References.ErrorField = "m.err" + rootFieldName
	f.References.InitBoolField = "m.init" + rootFieldName
	f.References.ResolvedField = "m.resolved" + rootFieldName
}

func (f *Field) BuildResolverMethod() *Func {
	f.createResolverFieldReferences()
	f.References.AssignmentVar = f.References.ResolvedField
	f.References.IDReferenceVar = f.References.RootField

	method := &Func{}
	method.Parent = f.Parent
	method.SourceFile = f.Parent.SourceFile
	method.Name = "GetResolved_" + f.Name

	// Create resolver method signature
	resolverMethod := &ast.FuncDecl{}
	resolverMethod.Recv = &ast.FieldList{}
	resolverMethod.Recv.List = []*ast.Field{{
		Names: []*ast.Ident{ast.NewIdent("m")},
		Type:  ast.NewIdent("*" + f.Parent.Name),
	}}
	resolverMethod.Name = ast.NewIdent(method.Name)
	resolverMethod.Type = &ast.FuncType{}
	resolverMethod.Type.Results = &ast.FieldList{}
	resolverMethod.Type.Results.List = []*ast.Field{
		{
			Type: ast.NewIdent(f.Type),
		},
		{
			Type: ast.NewIdent("error"),
		},
	}

	// Create resolver method body
	resolverMethod.Body = &ast.BlockStmt{}
	resolverMethod.Body.List = []ast.Stmt{
		&ast.IfStmt{
			Cond: ast.NewIdent(f.References.InitBoolField),
			Body: &ast.BlockStmt{List: []ast.Stmt{f.returnResolvedFieldAndError()}},
		},
	}
	resolverMethod.Body.List = append(resolverMethod.Body.List, f.buildResolverBody()...)
	resolverMethod.Body.List = append(resolverMethod.Body.List,
		f.assignInitBoolTrue(),
		f.returnResolvedFieldAndError(),
	)

	method.InputAST = resolverMethod
	return method
}

func (f *Field) buildResolverBody() []ast.Stmt {
	body := []ast.Stmt{}

	if f.IsMap || f.IsPointer || f.IsSlice {
		do := []ast.Stmt{}
		if !f.References.InLoop {
			do = append(do,
				f.assignInitBoolTrue(),
				f.returnResolvedFieldAndError(),
			)
		} else {
			if f.ParentField != nil && f.ParentField.IsMap {
				do = append(do, f.newAssignmentNil())
			}
			do = append(do, f.continueStmt())
		}

		body = append(body,
			f.checkIdReferenceNil(
				do...,
			),
		)

		if f.IsPointer && f.ChildField.ChildField == nil {
			findMethod := f.findByObjectID
			if f.ChildField.IsSlice {
				findMethod = f.findByObjectIDs
			}

			body = append(body, f.newAssignmentVar())
			body = append(body, findMethod()...)
			return body
		} else if f.IsSlice && f.ChildField.ChildField == nil {
			if f.ParentField == nil || !f.ParentField.IsPointer {
				body = append(body, f.newAssignmentVar())
			}
			body = append(body, f.findByObjectIDs()...)
			return body
		}

		if f.IsMap {
			alpha := getAlphabetLetter(f.References.Level)
			f.ChildField.createResolverFieldReferences()
			keyVar, valVar := "k"+alpha, "v"+alpha
			f.ChildField.References.AssignmentVar = f.References.AssignmentVar + "[" + keyVar + "]"
			f.ChildField.References.IDReferenceVar = valVar
			f.ChildField.References.InLoop = true
			loopBodyFuncs := f.ChildField.buildResolverBody()
			body = append(
				body,
				// f.newAssignmentVar(),
				f.forLoopMapSlice(
					keyVar, valVar,
					loopBodyFuncs...,
				),
			)
		} else if f.IsSlice {
			alpha := getAlphabetLetter(f.References.Level)
			f.ChildField.createResolverFieldReferences()
			keyVar, valVar := "k"+alpha, "v"+alpha
			f.ChildField.References.AssignmentVar = f.References.AssignmentVar + "[" + keyVar + "]"
			f.ChildField.References.IDReferenceVar = valVar
			f.ChildField.References.InLoop = true
			loopBodyFuncs := f.ChildField.buildResolverBody()
			body = append(
				body,
				f.newAssignmentVar(),
				f.forLoopMapSlice(
					keyVar, valVar,
					loopBodyFuncs...,
				),
			)
			return body
		} else if f.IsPointer {
			body = append(body, f.newAssignmentVar())
			alpha := getAlphabetLetter(f.References.Level)
			f.ChildField.createResolverFieldReferences()
			f.ChildField.References.AssignmentVar = alpha + "Assign"
			f.ChildField.References.IDReferenceVar = alpha + "ID"
			body = append(body,
				f.dereferenceAssignmentVarPtr(f.ChildField.References.AssignmentVar),
				f.dereferenceIDReferenceVarPtr(f.ChildField.References.IDReferenceVar),
			)
			body = append(body, f.ChildField.buildResolverBody()...)
			return body
		}
	} else {
		body = append(body, f.findByObjectID()...)
	}

	return body
}

func (*Field) continueStmt() *ast.BranchStmt {
	return &ast.BranchStmt{Tok: token.CONTINUE}
}

func (f *Field) forLoopMapSlice(keyVar, valVar string, do ...ast.Stmt) *ast.RangeStmt {
	loop := &ast.RangeStmt{}
	loop.Key = ast.NewIdent(keyVar)
	loop.Value = ast.NewIdent(valVar)
	loop.Tok = token.DEFINE
	loop.X = ast.NewIdent(f.References.IDReferenceVar)
	loop.Body = &ast.BlockStmt{
		List: do,
	}
	return loop
}

func (f *Field) dereferenceAssignmentVarPtr(varName string) *ast.AssignStmt {
	if !f.IsPointer {
		log.Fatalf("ID reference variable is not a pointer: %s", f.References.IDReferenceVar)
	}

	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(varName)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.StarExpr{
			X: ast.NewIdent(f.References.AssignmentVar),
		}},
	}
}

func (f *Field) dereferenceIDReferenceVarPtr(varName string) *ast.AssignStmt {
	if !f.IsPointer {
		log.Fatalf("ID reference variable is not a pointer: %s", f.References.IDReferenceVar)
	}

	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(varName)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.StarExpr{
			X: ast.NewIdent(f.References.IDReferenceVar),
		}},
	}
}

func (f *Field) getAssignmentToken() token.Token {
	if strings.HasPrefix(f.References.AssignmentVar, "m.") {
		return token.ASSIGN
	}
	return token.DEFINE
}

func (f *Field) newTypeOfSelf() ast.Expr {
	if f.IsMap {
		return &ast.CallExpr{
			Fun:  ast.NewIdent("make"),
			Args: []ast.Expr{ast.NewIdent(f.Type)},
		}
	}

	if f.IsPointer {
		return &ast.CallExpr{
			Fun:  ast.NewIdent("new"),
			Args: []ast.Expr{ast.NewIdent(f.Type[1:])},
		}
	}

	if f.IsSlice {
		var lenArg ast.Expr
		lenArg = ast.NewIdent("0")
		if f.ChildField.ChildField != nil {
			lenArg = &ast.CallExpr{
				Fun:  ast.NewIdent("len"),
				Args: []ast.Expr{ast.NewIdent(f.References.IDReferenceVar)},
			}
		}

		return &ast.CallExpr{
			Fun: ast.NewIdent("make"),
			Args: []ast.Expr{
				ast.NewIdent(f.Type),
				lenArg,
			},
		}
	}

	return ast.NewIdent(f.Type + "{}")
}

func (f *Field) newAssignmentNil() *ast.AssignStmt {
	assignmentToken := f.getAssignmentToken()
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(f.References.AssignmentVar)},
		Tok: assignmentToken,
		Rhs: []ast.Expr{ast.NewIdent("nil")},
	}
}

func (f *Field) newAssignmentVar() *ast.AssignStmt {
	assignmentToken := f.getAssignmentToken()
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(f.References.AssignmentVar)},
		Tok: assignmentToken,
		Rhs: []ast.Expr{f.newTypeOfSelf()},
	}
}

func (f *Field) beforeFindActions() (string, *ast.AssignStmt) {
	var oldAssignmentVar string
	var assignmentStmt *ast.AssignStmt

	if f.ParentField != nil {
		if f.ParentField.IsMap {
			oldAssignmentVar = f.References.AssignmentVar
			if oldAssignmentVar[0] == '&' {
				oldAssignmentVar = oldAssignmentVar[1:]
			}

			newAssignmentVar := getAlphabetLetter(f.References.Level) + "Assign"
			f.References.AssignmentVar = newAssignmentVar

			assignmentStmt = &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent(f.References.AssignmentVar)},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{f.newTypeOfSelf()},
			}
		}
	}

	if !f.IsPointer {
		if f.References.AssignmentVar[0] != '&' {
			f.References.AssignmentVar = "&" + f.References.AssignmentVar
		}
	}

	return oldAssignmentVar, assignmentStmt
}

func (f *Field) afterFindActions(varName string) *ast.AssignStmt {
	var assignmentStmt *ast.AssignStmt
	if !f.IsPointer {
		if f.References.AssignmentVar[0] == '&' {
			f.References.AssignmentVar = f.References.AssignmentVar[1:]
		}
	}

	if varName != "" {
		assignmentStmt = &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(varName)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(f.References.AssignmentVar)},
		}
		f.References.AssignmentVar = varName
	}

	return assignmentStmt
}

func (f *Field) findByObjectID() []ast.Stmt {
	retStmt := []ast.Stmt{}
	oldAssignVar, findBeforeStmt := f.beforeFindActions()
	if findBeforeStmt != nil {
		retStmt = append(retStmt, findBeforeStmt)
	}

	retStmt = append(retStmt,
		&ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{ast.NewIdent(f.References.ErrorField)},
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent("FindByObjectID"),
					Args: []ast.Expr{
						ast.NewIdent(f.References.AssignmentVar),
						ast.NewIdent(f.References.IDReferenceVar),
					},
				},
			},
		},
	)

	if findAfterStmt := f.afterFindActions(oldAssignVar); findAfterStmt != nil {
		retStmt = append(retStmt, findAfterStmt)
	}

	if !f.IsResolvable {
		retStmt = append(retStmt, f.checkErrorFieldNil(
			f.assignInitBoolTrue(),
			f.returnResolvedFieldAndError(),
		))
	}

	return retStmt
}

func (f *Field) findByObjectIDs() []ast.Stmt {
	retStmt := []ast.Stmt{}
	oldAssignVar, assignBefore := f.beforeFindActions()
	if assignBefore != nil {
		retStmt = append(retStmt, assignBefore)
	}

	if !f.IsPointer {
		if f.References.AssignmentVar[0] != '&' {
			f.References.AssignmentVar = "&" + f.References.AssignmentVar
		}
	}

	retStmt = append(retStmt,
		&ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{ast.NewIdent(f.References.ErrorField)},
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent("FindByObjectIDs"),
					Args: []ast.Expr{
						ast.NewIdent(f.References.AssignmentVar),
						ast.NewIdent(f.References.IDReferenceVar),
					},
				},
			},
		},
	)

	if findAfterStmt := f.afterFindActions(oldAssignVar); findAfterStmt != nil {
		retStmt = append(retStmt, findAfterStmt)
	}

	if f.ChildField.ChildField != nil {
		retStmt = append(retStmt, f.checkErrorFieldNil(
			f.assignInitBoolTrue(),
			f.returnResolvedFieldAndError(),
		))
	}

	return retStmt
}

func (f *Field) checkErrorFieldNil(do ...ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent(f.References.ErrorField),
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{List: do},
	}
}

func (f *Field) checkIdReferenceNil(do ...ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent(f.References.IDReferenceVar),
			Op: token.EQL,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{List: do},
	}
}

func (f *Field) assignInitBoolTrue() *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(f.References.InitBoolField)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ast.NewIdent("true")},
	}
}

func (f *Field) returnResolvedFieldAndError() *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent(f.References.ResolvedField),
			ast.NewIdent(f.References.ErrorField),
		},
	}
}
