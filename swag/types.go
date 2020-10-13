package swag

import (
	"go/ast"

	"github.com/go-openapi/spec"
)

// Schema parsed schema
type Schema struct {
	PkgPath string
	Name    string
	*spec.Schema
}

// TypeSpecDef the whole information of a typeSpec
type TypeSpecDef struct {
	PkgPath  string
	File     *ast.File
	TypeSpec *ast.TypeSpec
}

// Name name of the typeSpec
func (t *TypeSpecDef) Name() string {
	return t.TypeSpec.Name.Name
}

// FullName full name of the typeSpec
func (t *TypeSpecDef) FullName() string {
	return fullTypeName(t.File.Name.Name, t.TypeSpec.Name.Name)
}

// AstFileInfo information of a ast.File
type AstFileInfo struct {
	File        *ast.File
	Path        string
	PackagePath string
}

// PackageDefinitions files and definition in a package
type PackageDefinitions struct {
	Name            string
	Files           map[string]*ast.File
	TypeDefinitions map[string]*TypeSpecDef
}
