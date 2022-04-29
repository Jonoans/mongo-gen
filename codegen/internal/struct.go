package internal

import (
	"go/ast"
	"go/types"
	"sort"
	"strings"
)

var structDbMethods = []*structDbMethod{
	{"AggregateFirst", "AggregateFirst", []*structDbMethodParam{
		{"", "ModelInterface", "m"},
		{"pipeline", "interface{}", "pipeline"},
		{"opts", "...*options.AggregateOptions", "opts..."},
	}, []string{"bool", "error"}},
	{"Find", "FindOne", []*structDbMethodParam{
		{"", "ModelInterface", "m"},
		{"query", "interface{}", "query"},
		{"opts", "...*options.FindOneOptions", "opts..."},
	}, []string{"error"}},
	{"FindWithCtx", "FindOneWithCtx", []*structDbMethodParam{
		{"ctx", "context.Context", "ctx"},
		{"", "ModelInterface", "m"},
		{"query", "interface{}", "query"},
		{"opts", "...*options.FindOneOptions", "opts..."},
	}, []string{"error"}},
	{"FindByObjectID", "FindByObjectID", []*structDbMethodParam{
		{"", "ModelInterface", "m"},
		{"id", "interface{}", "id"},
		{"opts", "...*options.FindOneOptions", "opts..."},
	}, []string{"error"}},
	{"FindByObjectIDWithCtx", "FindByObjectIDWithCtx", []*structDbMethodParam{
		{"ctx", "context.Context", "ctx"},
		{"", "ModelInterface", "m"},
		{"id", "interface{}", "id"},
		{"opts", "...*options.FindOneOptions", "opts..."},
	}, []string{"error"}},
	{"Create", "InsertOne", []*structDbMethodParam{
		{"", "ModelInterface", "m"},
		{"opts", "...*options.InsertOneOptions", "opts..."},
	}, []string{"error"}},
	{"CreateWithCtx", "InsertOneWithCtx", []*structDbMethodParam{
		{"ctx", "context.Context", "ctx"},
		{"", "ModelInterface", "m"},
		{"opts", "...*options.InsertOneOptions", "opts..."},
	}, []string{"error"}},
	{"Update", "Update", []*structDbMethodParam{
		{"", "ModelInterface", "m"},
		{"opts", "...*options.UpdateOptions", "opts..."},
	}, []string{"error"}},
	{"UpdateWithCtx", "UpdateWithCtx", []*structDbMethodParam{
		{"ctx", "context.Context", "ctx"},
		{"", "ModelInterface", "m"},
		{"opts", "...*options.UpdateOptions", "opts..."},
	}, []string{"error"}},
	{"Delete", "Delete", []*structDbMethodParam{
		{"", "ModelInterface", "m"},
		{"opts", "...*options.DeleteOptions", "opts..."},
	}, []string{"error"}},
	{"DeleteWithCtx", "DeleteWithCtx", []*structDbMethodParam{
		{"ctx", "context.Context", "ctx"},
		{"", "ModelInterface", "m"},
		{"opts", "...*options.DeleteOptions", "opts..."},
	}, []string{"error"}},
}

type structDbMethodParam struct {
	paramName string
	paramType string
	argUsage  string
}

type structDbMethod struct {
	name           string
	globalFuncName string
	params         []*structDbMethodParam
	retTypes       []string
}

type Struct struct {
	Parent     *Package
	SourceFile string
	InputAST   *ast.StructType
	InputType  *types.Struct

	ParsedMethods []*Func

	// Sorted collection of methods
	CollectionNameMethod *Func
	HookMethods          []*Func // Source file will be set to struct's source file
	DatabaseMethods      []*Func // Source file will be set to struct's source file
	UserDefinedMethods   []*Func // Retain source file for user defined methods
	ResolverMethods      []*Func // Source file will be set to struct's source file

	Name           string
	IsCollection   bool
	EmbeddedFields []*Field
	Fields         []*Field
	ResolverFields []*Field
}

func (s *Struct) Init() {
	s.initFields()
	s.classifyDefinedMethods()
	s.initMethods()
}

// ******* SECTION Fields ******* //
func (s *Struct) initFields() {
	for i, field := range s.InputAST.Fields.List {
		t := s.InputType.Field(i)
		if !t.Exported() {
			continue
		}

		f := &Field{}
		f.Parent = s
		f.InputAST = field
		f.InputTypesVar = t
		f.Init()

		if f.IsBaseModelDerivative {
			s.IsCollection = true
		}

		if f.IsEmbedded {
			s.EmbeddedFields = append(s.EmbeddedFields, f)
		} else {
			s.Fields = append(s.Fields, f)
		}
	}
}

func (s *Struct) InitResolverFieldsAndMethods() {
	packagePath := s.Parent.InputUser.PkgPath
	packageTrimPrefix := packagePath + "."
	for _, field := range s.Fields {
		if field.IsReference {
			fieldResolvedType := field.ResolvedType.String()
			if strings.HasPrefix(fieldResolvedType, packageTrimPrefix) {
				referencedType := strings.Replace(fieldResolvedType, packageTrimPrefix, "", 1)
				referencedStruct := s.Parent.Structs[referencedType]
				if referencedStruct.IsCollection {
					field.IsResolvable = true
					if field.IsMap || field.IsPointer || field.IsSlice {
						field.CreateChildField()
					}
					s.ResolverMethods = append(s.ResolverMethods, field.BuildResolverMethod())
					s.ResolverFields = append(s.ResolverFields, field.CreateStubResolvableFields()...)
					field.MutateResolvableFieldType()
				}
			}
		}
	}
}

// ********** SECTION Methods ********** //
func (s *Struct) classifyDefinedMethods() {
	for _, f := range s.ParsedMethods {
		// Sort
		astNode := f.InputAST
		if isCollectionNameMethod(astNode) {
			s.CollectionNameMethod = f
		} else if isHookMethod(astNode) {
			f.SourceFile = s.SourceFile
			s.HookMethods = append(s.HookMethods, f)
		} else {
			if isMethodNameReserved(f.Name) {
				continue
			}
			f.Parent = s
			s.UserDefinedMethods = append(s.UserDefinedMethods, f)
		}
	}

	// No longer required
	s.ParsedMethods = nil
}

func (s *Struct) initMethods() {
	if s.IsCollection {
		if s.CollectionNameMethod == nil {
			s.CollectionNameMethod = buildCollectionNameMethod(s)
			s.CollectionNameMethod.Parent = s
		}

		for _, m := range structDbMethods {
			s.DatabaseMethods = append(s.DatabaseMethods, buildDatabaseMethod(s, m))
		}

		existingHookMethods := make(map[string]bool)
		for _, hookMethod := range s.HookMethods {
			existingHookMethods[hookMethod.Name] = true
			hookMethod.Parent = s
		}

		for _, hookMethod := range structHookMethodNames {
			if !existingHookMethods[hookMethod] {
				s.HookMethods = append(s.HookMethods, buildHookMethod(s, hookMethod))
			}
		}

		s.sortHookMethods()
	}
}

func (s *Struct) sortHookMethods() {
	nameIndex := make(map[string]int)
	for i, n := range structHookMethodNames {
		nameIndex[n] = i
	}

	sort.Slice(s.HookMethods, func(i, j int) bool {
		return nameIndex[s.HookMethods[i].Name] < nameIndex[s.HookMethods[j].Name]
	})
}
