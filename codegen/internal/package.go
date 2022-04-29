package internal

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"github.com/jonoans/mongo-gen/utils"
	"golang.org/x/tools/go/packages"
)

type Package struct {
	InputUser           *packages.Package
	InputGenerated      *packages.Package
	InputGeneratedLines map[string][]string

	// Struct-related Values
	Structs       map[string]*Struct // Key: Struct name
	StructMethods map[string][]*Func // Key: Struct name, Value: map[FuncName]*Func

	// Others, key: filename
	Funcs      map[string][]*Func
	Imports    map[string][]*Import
	Interfaces map[string][]*Interface
	Values     map[string][]*Value
}

func (p *Package) Init() {
	p.Structs = make(map[string]*Struct)
	p.StructMethods = make(map[string][]*Func)
	p.Funcs = make(map[string][]*Func)
	p.Imports = make(map[string][]*Import)
	p.Interfaces = make(map[string][]*Interface)
	p.Values = make(map[string][]*Value)
	p.parseUserStructs()
	p.parseGeneratedNonStructs()
	p.prepareStructs()
	p.prepareResolvableFields()
}

func (p *Package) GeneratePackageFiles() map[string]*PackageFile {
	pkgFiles := map[string]*PackageFile{}

	for n, fc := range p.InputGeneratedLines {
		if n == "codegen_.go" {
			continue
		}

		if _, ok := pkgFiles[n]; !ok {
			pkgFiles[n] = &PackageFile{
				Filename: n,
			}
		}

		pkgFiles[n].FileContents = fc
	}

	for _, s := range p.Structs {
		if _, ok := pkgFiles[s.SourceFile]; !ok {
			pkgFiles[s.SourceFile] = &PackageFile{
				Filename: s.SourceFile,
			}
		}
		pkgFiles[s.SourceFile].Structs = append(pkgFiles[s.SourceFile].Structs, s)

		for _, m := range s.UserDefinedMethods {
			if _, ok := pkgFiles[m.SourceFile]; !ok {
				pkgFiles[m.SourceFile] = &PackageFile{
					Filename: m.SourceFile,
				}
			}
			pkgFiles[m.SourceFile].Functions = append(pkgFiles[m.SourceFile].Functions, m)
		}
	}

	for n, fs := range p.Funcs {
		if _, ok := pkgFiles[n]; !ok {
			pkgFiles[n] = &PackageFile{
				Filename: n,
			}
		}

		for _, f := range fs {
			pkgFiles[n].Functions = append(pkgFiles[n].Functions, f)
		}
	}

	for n, is := range p.Imports {
		if _, ok := pkgFiles[n]; !ok {
			pkgFiles[n] = &PackageFile{
				Filename: n,
			}
		}

		for _, i := range is {
			pkgFiles[n].Imports = append(pkgFiles[n].Imports, i)
		}
	}

	for n, is := range p.Interfaces {
		if _, ok := pkgFiles[n]; !ok {
			pkgFiles[n] = &PackageFile{
				Filename: n,
			}
		}

		for _, i := range is {
			pkgFiles[n].Interfaces = append(pkgFiles[n].Interfaces, i)
		}
	}

	for n, vs := range p.Values {
		if _, ok := pkgFiles[n]; !ok {
			pkgFiles[n] = &PackageFile{
				Filename: n,
			}
		}

		for _, v := range vs {
			switch v.InputTok {
			case token.CONST:
				pkgFiles[n].ConstValues = append(pkgFiles[n].ConstValues, v)
			case token.VAR:
				pkgFiles[n].VarValues = append(pkgFiles[n].VarValues, v)
			}
		}
	}

	return pkgFiles
}

func (p *Package) prepareResolvableFields() {
	for _, s := range p.Structs {
		s.InitResolverFieldsAndMethods()
	}
}

func (p *Package) prepareStructs() {
	for _, s := range p.Structs {
		structTypeObj := p.InputUser.Types.Scope().Lookup(s.Name).Type().Underlying().(*types.Struct)

		s.Parent = p
		s.InputType = structTypeObj
		s.ParsedMethods = p.StructMethods[s.Name]

		s.Init()
	}
	// No longer required
	p.StructMethods = nil
}

func (p *Package) parseUserStructs() {
	for _, file := range p.InputUser.Syntax {
		filename := p.InputUser.Fset.File(file.Pos()).Name()
		filename = utils.BaseFilename(filename)
		if filename == "codegen_.go" {
			continue
		}

		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if isFuncValueNameReserved(spec.Name.Name) {
							log.Printf("Skipping reserved name: %s", spec.Name.Name)
							continue
						}
						switch specType := spec.Type.(type) {
						case *ast.StructType:
							structName := spec.Name.Name
							p.Structs[structName] = &Struct{
								SourceFile: filename,
								InputAST:   specType,
								Name:       structName,
							}
						}
					}
				}
			}
		}
	}
}

func (p *Package) parseGeneratedNonStructs() {
	for _, file := range p.InputGenerated.Syntax {
		filename := p.InputGenerated.Fset.File(file.Pos()).Name()
		filename = utils.BaseFilename(filename)
		if filename == "codegen_.go" {
			continue
		}

		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if isFuncValueNameReserved(spec.Name.Name) {
							continue
						}
						switch spec.Type.(type) {
						case *ast.InterfaceType:
							interfaceName := spec.Name.Name
							p.Interfaces[filename] = append(p.Interfaces[filename],
								&Interface{
									SourceFile: filename,
									InputAST:   decl,
									Name:       interfaceName,
								},
							)
						}
					case *ast.ValueSpec:
						varName := astIdentSliceToString(spec.Names)
						if isFuncValueNameReserved(varName) {
							log.Printf("Skipping reserved name: %s", varName)
							continue
						}

						p.Values[filename] = append(p.Values[filename],
							&Value{
								SourceFile: filename,
								InputTok:   decl.Tok,
								InputAST:   spec,
								Name:       varName,
							},
						)
					case *ast.ImportSpec:
						p.Imports[filename] = append(p.Imports[filename],
							&Import{
								SourceFile: filename,
								InputAST:   spec,
							},
						)
					}
				}
			case *ast.FuncDecl:
				funcName := astObjectToString(decl.Name)
				if decl.Recv != nil {
					if isMethodNameResolver(funcName) {
						continue
					}

					rcvType := decl.Recv.List[0].Type

					var structTypeName string
					switch rcvType := rcvType.(type) {
					case *ast.StarExpr:
						structTypeName = astObjectToString(rcvType.X)
					case *ast.Ident:
						structTypeName = astObjectToString(rcvType)
					}

					p.StructMethods[structTypeName] = append(p.StructMethods[structTypeName],
						&Func{
							SourceFile: filename,
							InputAST:   decl,
							Name:       funcName,
							FileSet:    p.InputGenerated.Fset,
							Generated:  true,
						},
					)
				} else {
					found := false
					for _, f := range p.Funcs[filename] {
						if f.Name == funcName {
							found = true
						}
					}

					if found {
						continue
					}

					if isFuncValueNameReserved(funcName) {
						log.Printf("Skipping reserved name: %s", funcName)
						continue
					}

					p.Funcs[filename] = append(p.Funcs[filename],
						&Func{
							SourceFile: filename,
							InputAST:   decl,
							Name:       funcName,
							FileSet:    p.InputGenerated.Fset,
							Generated:  true,
						},
					)
				}
			}
		}
	}
}

func (p *Package) parseUserStructsMigrate() {
	for _, file := range p.InputUser.Syntax {
		filename := p.InputUser.Fset.File(file.Pos()).Name()
		filename = utils.BaseFilename(filename)
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if isFuncValueNameReserved(spec.Name.Name) {
							log.Printf("Skipping reserved name: %s", spec.Name.Name)
							continue
						}
						switch specType := spec.Type.(type) {
						case *ast.StructType:
							structName := spec.Name.Name
							p.Structs[structName] = &Struct{
								SourceFile: filename,
								InputAST:   specType,
								Name:       structName,
							}
						case *ast.InterfaceType:
							interfaceName := spec.Name.Name
							p.Interfaces[filename] = append(p.Interfaces[filename],
								&Interface{
									SourceFile: filename,
									InputAST:   decl,
									Name:       interfaceName,
								},
							)
						default:
							log.Printf("Unsupported type found, contact the developer if you require support: %T", specType)
						}
					case *ast.ValueSpec:
						varName := astIdentSliceToString(spec.Names)
						p.Values[filename] = append(p.Values[filename],
							&Value{
								SourceFile: filename,
								InputTok:   decl.Tok,
								InputAST:   spec,
								Name:       varName,
							},
						)
					case *ast.ImportSpec:
						p.Imports[filename] = append(p.Imports[filename],
							&Import{
								SourceFile: filename,
								InputAST:   spec,
							},
						)
					}
				}
			case *ast.FuncDecl:
				funcName := astObjectToString(decl.Name)
				if decl.Recv != nil {
					rcvType := decl.Recv.List[0].Type

					var structTypeName string
					switch rcvType := rcvType.(type) {
					case *ast.StarExpr:
						structTypeName = astObjectToString(rcvType.X)
					case *ast.Ident:
						structTypeName = astObjectToString(rcvType)
					}

					p.StructMethods[structTypeName] = append(p.StructMethods[structTypeName],
						&Func{
							SourceFile: filename,
							InputAST:   decl,
							Name:       funcName,
							FileSet:    p.InputUser.Fset,
							Generated:  false,
						},
					)
				} else {
					if isFuncValueNameReserved(funcName) {
						log.Printf("Skipping reserved name: %s", funcName)
						continue
					}

					p.Funcs[filename] = append(p.Funcs[filename],
						&Func{
							SourceFile: filename,
							InputAST:   decl,
							Name:       funcName,
							FileSet:    p.InputUser.Fset,
							Generated:  false,
						},
					)
				}
			}
		}
	}
}
