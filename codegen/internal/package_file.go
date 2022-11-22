package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jonoans/mongo-gen/config"
	"golang.org/x/tools/imports"
)

type PackageFile struct {
	Filename    string
	PackageName string

	FileContents []string
	Structs      []*Struct
	Functions    []*Func
	ConstValues  []*Value
	Imports      []*Import
	Interfaces   []*Interface
	VarValues    []*Value
}

func (p *PackageFile) WriteToFile(cfg *config.OutputConfig) {
	p.PackageName = cfg.PackageName
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(fmt.Sprintf("package %s\n\n", p.PackageName))

	p.writeImports(buffer)
	p.writeConsts(buffer)
	p.writeVars(buffer)
	p.writeInterfaces(buffer)
	p.writeStructs(buffer)
	p.writeStructCollectionNameMethods(buffer)
	p.writeFuncs(buffer)
	p.writeStructHookMethods(buffer)
	p.writeStructResolverMethods(buffer)
	p.writeStructDatabaseMethods(buffer)
	p.writeBufferToFile(cfg.PackagePath, buffer)
}

func (p *PackageFile) writeImports(buffer *bytes.Buffer) {
	if len(p.Imports) == 0 {
		return
	}

	decl := &ast.GenDecl{Tok: token.IMPORT}
	for _, i := range p.Imports {
		decl.Specs = append(decl.Specs, i.InputAST)
	}

	p.writeDeclsToBuffer(buffer, decl)
}

func (p *PackageFile) writeConsts(buffer *bytes.Buffer) {
	if len(p.ConstValues) == 0 {
		return
	}

	decl := &ast.GenDecl{Tok: token.CONST}
	for _, v := range p.ConstValues {
		decl.Specs = append(decl.Specs, v.InputAST)
	}

	p.writeDeclsToBuffer(buffer, decl)
}

func (p *PackageFile) writeVars(buffer *bytes.Buffer) {
	if len(p.VarValues) == 0 {
		return
	}

	decl := &ast.GenDecl{Tok: token.VAR}
	for _, v := range p.VarValues {
		decl.Specs = append(decl.Specs, v.InputAST)
	}

	p.writeDeclsToBuffer(buffer, decl)
}

func (p *PackageFile) writeInterfaces(buffer *bytes.Buffer) {
	if len(p.Interfaces) == 0 {
		return
	}

	for _, i := range p.Interfaces {
		p.writeDeclsToBuffer(buffer, i.InputAST)
	}
}

func (p *PackageFile) writeFuncs(buffer *bytes.Buffer) {
	for _, f := range p.Functions {
		p.writeFuncToBuffer(buffer, f)
	}
}

func (p *PackageFile) writeStructs(mainBuffer *bytes.Buffer) {
	buffer := bytes.NewBuffer(nil)
	for _, s := range p.Structs {
		s.Fields = append(s.EmbeddedFields, s.Fields...)
	}

	template := GetTemplate("struct")
	err := template.Execute(buffer, p)
	if err != nil {
		log.Fatalf("Could not write to output file: %s", err)
	}

	mainBuffer.Write(buffer.Bytes())
}

func (p *PackageFile) writeStructCollectionNameMethods(buffer *bytes.Buffer) {
	for _, s := range p.Structs {
		if s.IsCollection {
			p.writeFuncToBuffer(buffer, s.CollectionNameMethod)
		}
	}
}

func (p *PackageFile) writeStructHookMethods(buffer *bytes.Buffer) {
	for _, s := range p.Structs {
		if s.IsCollection {
			for _, m := range s.HookMethods {
				p.writeFuncToBuffer(buffer, m)
			}
		}
	}
}

func (p *PackageFile) writeStructResolverMethods(buffer *bytes.Buffer) {
	for _, s := range p.Structs {
		for _, m := range s.ResolverMethods {
			p.writeFuncToBuffer(buffer, m)
		}
	}
}

func (p *PackageFile) writeStructDatabaseMethods(buffer *bytes.Buffer) {
	for _, s := range p.Structs {
		if s.IsCollection {
			for _, m := range s.DatabaseMethods {
				p.writeFuncToBuffer(buffer, m)
			}
		}
	}
}

func (p *PackageFile) writeFuncToBuffer(buffer io.Writer, f *Func) {
	if f.FileSet != nil && f.Generated {
		startLine, endLine := f.FileSet.Position(f.InputAST.Pos()).Line, f.FileSet.Position(f.InputAST.End()).Line
		funcBytes := []byte(strings.Join(p.FileContents[startLine-1:endLine], "\n"))
		_, err := buffer.Write(funcBytes)
		if err != nil {
			log.Fatalf("Could not write to output file: %s", err)
		}

		buffer.Write([]byte("\n\n")) // Shouldn't be an issue
		return
	}
	p.writeDeclsToBuffer(buffer, f.InputAST)
}

func (*PackageFile) writeDeclsToBuffer(buffer io.Writer, decl ...ast.Decl) {
	for _, d := range decl {
		err := printer.Fprint(buffer, token.NewFileSet(), d)
		if err != nil {
			log.Fatalf("Could not write to buffer: %s", err)
		}

		buffer.Write([]byte("\n\n")) // Shouldn't be an issue
	}
}

func (p *PackageFile) writeBufferToFile(packagePath string, buffer *bytes.Buffer) {
	outBytes, err := imports.Process(packagePath, buffer.Bytes(), &imports.Options{})
	if err != nil {
		log.Fatalf("Error formatting package: %s", err)
	}

	outputFilepath := filepath.Join(packagePath, p.Filename)
	fh, err := os.OpenFile(outputFilepath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	defer fh.Close()
	if err != nil {
		log.Fatalf("Could not open output file: %s", err)
	}

	_, err = fh.Write(outBytes)
	if err != nil {
		log.Fatalf("Could not write to output file: %s", err)
	}
}

func (p *PackageFile) Init() {
	for _, i := range p.Imports {
		i.Init()
	}
}

func (p *PackageFile) Sort() {
	p.sortStructs()
	p.sortFunctions()
	p.sortConsts()
	p.sortInterfaces()
	p.sortValues()
}

func (p *PackageFile) sortStructs() {
	sort.Slice(p.Structs, func(i, j int) bool {
		return p.Structs[i].Name < p.Structs[j].Name
	})
}

func (p *PackageFile) sortFunctions() {
	sort.Slice(p.Functions, func(i, j int) bool {
		if p.Functions[i].Name == p.Functions[j].Name {
			if p.Functions[i].Parent == nil || p.Functions[j].Parent == nil {
				return p.Functions[i].Parent == nil
			}
			return p.Functions[i].Parent.Name < p.Functions[j].Parent.Name
		}
		return p.Functions[i].Name < p.Functions[j].Name
	})
}

func (p *PackageFile) sortConsts() {
	sort.Slice(p.ConstValues, func(i, j int) bool {
		return p.ConstValues[i].Name < p.ConstValues[j].Name
	})
}

func (p *PackageFile) sortInterfaces() {
	sort.Slice(p.Interfaces, func(i, j int) bool {
		return p.Interfaces[i].Name < p.Interfaces[j].Name
	})
}

func (p *PackageFile) sortValues() {
	sort.Slice(p.VarValues, func(i, j int) bool {
		return p.VarValues[i].Name < p.VarValues[j].Name
	})
}
