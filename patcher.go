package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

type Patcher struct {
	fset *token.FileSet
	f    *ast.File
}

func NewPatcher(src []byte) (*Patcher, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	return &Patcher{fset, file}, err
}

type StructPatcher func(*ast.StructType) error

func (p *Patcher) PatchStruct(name string, patch StructPatcher) error {
	for i := range p.f.Decls {
		decl := p.f.Decls[i]
		typeDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if len(typeDecl.Specs) != 1 {
			continue
		}

		typeSpec, ok := typeDecl.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}

		if typeSpec.Name.Name != name {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return fmt.Errorf(name + " is not struct type")
		}

		if err := patch(structType); err != nil {
			return fmt.Errorf("patch struct: %w", err)
		}

		return nil
	}

	return fmt.Errorf("struct with name %s not found", name)
}

func (p *Patcher) Src() (b []byte, e error) {
	buf := bytes.NewBuffer(nil)
	e = printer.Fprint(buf, p.fset, p.f)
	b = buf.Bytes()
	return
}

func ChangeField(name, to string) StructPatcher {
	return func(s *ast.StructType) error {
		ex, err := parser.ParseExpr(to)
		if err != nil {
			return err
		}

		for i := range s.Fields.List {
			field := s.Fields.List[i]
			if len(field.Names) != 1 {
				return fmt.Errorf("bad field name")
			}
			fname := field.Names[0].Name
			if fname != name {
				continue
			}

			field.Type = ex
			return nil
		}
		return fmt.Errorf("field with name %s not found", name)
	}
}
