package swag

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"

	"github.com/go-openapi/spec"
)

// CheckSchemaType checks if typeName is not a name of primitive type
func CheckSchemaType(typeName string) error {
	if !IsPrimitiveType(typeName) {
		return fmt.Errorf("%s is not basic types", typeName)
	}
	return nil
}

// IsSimplePrimitiveType determine whether the type name is a simple primitive type
func IsSimplePrimitiveType(typeName string) bool {
	switch typeName {
	case stringTypeValue, numberTypeValue, integerTypeValue, booleanTypeValue:
		return true
	default:
		return false
	}
}

// IsPrimitiveType determine whether the type name is a primitive type
func IsPrimitiveType(typeName string) bool {
	switch typeName {
	case stringTypeValue, numberTypeValue, integerTypeValue, booleanTypeValue, arrayTypeValue, objectTypeValue, funcTypeValue:
		return true
	default:
		return false
	}
}

// IsNumericType determines whether the swagger type name is a numeric type
func IsNumericType(typeName string) bool {
	return typeName == integerTypeValue || typeName == numberTypeValue
}

// TransToValidSchemeType indicates type will transfer golang basic type to swagger supported type.
func TransToValidSchemeType(typeName string) string {
	switch typeName {
	case "uint", "int", "uint8", "int8", "uint16", "int16", "byte":
		return integerTypeValue
	case "uint32", "int32", "rune":
		return integerTypeValue
	case "uint64", "int64":
		return integerTypeValue
	case "float32", "float64":
		return numberTypeValue
	case "bool":
		return booleanTypeValue
	case "string":
		return stringTypeValue
	default:
		return typeName // to support user defined types
	}
}

// IsGolangPrimitiveType determine whether the type name is a golang primitive type
func IsGolangPrimitiveType(typeName string) bool {
	switch typeName {
	case "uint",
		"int",
		"uint8",
		"int8",
		"uint16",
		"int16",
		"byte",
		"uint32",
		"int32",
		"rune",
		"uint64",
		"int64",
		"float32",
		"float64",
		"bool",
		"string":
		return true
	default:
		return false
	}
}

// TransToValidCollectionFormat determine valid collection format
func TransToValidCollectionFormat(format string) string {
	switch format {
	case "csv", "multi", "pipes", "tsv", "ssv":
		return format
	default:
		return ""
	}
}

// TypeDocName get alias from comment '// @name ', otherwise the original type name to display in doc
func TypeDocName(pkgName string, typeSpec *ast.TypeSpec) string {
	if typeSpec != nil {
		if typeSpec.Comment != nil {
			for _, comment := range typeSpec.Comment.List {
				text := strings.TrimSpace(comment.Text)
				text = strings.TrimPrefix(text, "//")
				text = strings.TrimSpace(text)
				texts := strings.Split(text, " ")
				if len(texts) > 1 && strings.EqualFold(texts[0], "@name") {
					return texts[1]
				}
			}
		}
		if typeSpec.Name != nil {
			return fullTypeName(strings.Split(pkgName, ".")[0], typeSpec.Name.Name)
		}
	}

	return pkgName
}

// RefSchema build a reference schema
func RefSchema(refType string) *spec.Schema {
	return spec.RefSchema("#/definitions/" + refType)
}

// PrimitiveSchema build a primitive schema
func PrimitiveSchema(refType string) *spec.Schema {
	return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{refType}}}
}

// BuildCustomSchema build custom schema specified by tag swaggertype
func BuildCustomSchema(types []string) (*spec.Schema, error) {
	if len(types) == 0 {
		return nil, nil
	}

	switch types[0] {
	case primitiveTypeValue:
		if len(types) == 1 {
			return nil, errors.New("need primitive type after primitive")
		}
		return BuildCustomSchema(types[1:])
	case arrayTypeValue:
		if len(types) == 1 {
			return nil, errors.New("need array item type after array")
		}
		schema, err := BuildCustomSchema(types[1:])
		if err != nil {
			return nil, err
		}
		return spec.ArrayProperty(schema), nil
	default:
		err := CheckSchemaType(types[0])
		if err != nil {
			return nil, err
		}
		return PrimitiveSchema(types[0]), nil
	}
}
