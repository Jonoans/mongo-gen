package codegen

import (
	"go/ast"
	"log"

	"golang.org/x/tools/go/packages"
)

const definitionsPkgPath = "github.com/jonoans/mongo-gen/codegen/internal/definitions"

var (
	mode = packages.NeedName |
		packages.NeedFiles |
		packages.NeedImports |
		packages.NeedTypes |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedModule |
		packages.NeedDeps
	cachedPkgs  = map[string]*packages.Package{}
	loadDefMode = packages.NeedFiles | packages.NeedSyntax
)

func loadPackage(patterns ...string) ([]*packages.Package, error) {
	toLoad := []string{}
	for _, pattern := range patterns {
		if _, ok := cachedPkgs[pattern]; !ok {
			toLoad = append(toLoad, pattern)
		}
	}

	pkgLoadCfg := packages.Config{Mode: mode}
	loadedPkgs, err := packages.Load(&pkgLoadCfg, toLoad...)
	if err != nil {
		return nil, err
	}

	for _, pkg := range loadedPkgs {
		cachedPkgs[pkg.PkgPath] = pkg
	}

	retVal := make([]*packages.Package, len(patterns))
	for i, pattern := range patterns {
		retVal[i] = cachedPkgs[pattern]
	}

	return retVal, nil
}

// getInternalDefinitons return declarations from definitions package
// and a list of identifiers found in the definitions package
func getInternalDefinitions() ([]ast.Decl, []string) {
	loadCfg := &packages.Config{Mode: loadDefMode}
	pkgs, err := packages.Load(loadCfg, definitionsPkgPath)
	if err != nil {
		log.Fatalf("Error loading internal package: %s", err)
	}

	decls := []ast.Decl{}
	for _, f := range pkgs[0].Syntax {
		f.Comments = nil
		decls = append(decls, f.Decls...)
	}

	idents := []string{}
	for _, d := range decls {
		switch d := d.(type) {
		case *ast.GenDecl:
			for _, d := range d.Specs {
				switch s := d.(type) {
				case *ast.TypeSpec:
					idents = append(idents, s.Name.Name)
				}
			}
		case *ast.FuncDecl:
			idents = append(idents, d.Name.Name)
		}
	}

	return decls, idents
}
