package schema

import (
	"go/ast"
	"testing"

	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArrayType_AttachesBasicItemSchema(t *testing.T) {
	p := newTestParser()

	schemaObject, err := p.ParseSchemaObject("/test/path", "pkg", "[]int32")

	require.NoError(t, err)
	require.NotNil(t, schemaObject)
	require.NotNil(t, schemaObject.Items)
	assert.Equal(t, "array", schemaObject.Type)
	assert.Equal(t, "integer", schemaObject.Items.Type)
	assert.Equal(t, "int64", schemaObject.Items.Format)
}

func TestParseArrayType_AttachesRefItemSchema(t *testing.T) {
	p := newTestParser()
	p.TypeSpecs["pkg"]["Item"] = &ast.TypeSpec{
		Name: ast.NewIdent("Item"),
		Type: &ast.StructType{Fields: &ast.FieldList{}},
	}

	schemaObject, err := p.ParseSchemaObject("/test/path", "pkg", "[]Item")

	require.NoError(t, err)
	require.NotNil(t, schemaObject)
	require.NotNil(t, schemaObject.Items)
	assert.Equal(t, "array", schemaObject.Type)
	assert.Equal(t, "#/components/schemas/pkg.Item", schemaObject.Items.Ref)
}

func TestParseCustomArrayAlias_AttachesRefItemSchema(t *testing.T) {
	p := newTestParser()
	p.TypeSpecs["pkg"]["Item"] = &ast.TypeSpec{
		Name: ast.NewIdent("Item"),
		Type: &ast.StructType{Fields: &ast.FieldList{}},
	}
	p.TypeSpecs["pkg"]["Items"] = &ast.TypeSpec{
		Name: ast.NewIdent("Items"),
		Type: &ast.ArrayType{Elt: ast.NewIdent("Item")},
	}

	schemaObject, err := p.ParseSchemaObject("/test/path", "pkg", "Items")

	require.NoError(t, err)
	require.NotNil(t, schemaObject)
	require.NotNil(t, schemaObject.Items)
	assert.Equal(t, "array", schemaObject.Type)
	assert.Equal(t, "#/components/schemas/pkg.Item", schemaObject.Items.Ref)
}

func newTestParser() *parser {
	openAPI := &OpenAPIObject{
		Components: ComponentsObject{
			Schemas: map[string]*SchemaObject{},
		},
	}
	return NewParser(model.Utils{
		PkgAndSpecs: &model.PkgAndSpecs{
			KnownNamePkg:            map[string]*model.Pkg{"pkg": &model.Pkg{Name: "pkg", Path: "/test/path"}},
			KnownPathPkg:            map[string]*model.Pkg{},
			KnownIDSchema:           map[string]*SchemaObject{},
			TypeSpecs:               map[string]map[string]*ast.TypeSpec{"pkg": {}},
			PkgPathAstPkgCache:      map[string]map[string]*ast.Package{},
			PkgNameImportedPkgAlias: map[string]map[string][]string{"pkg": {}},
		},
	}, openAPI).(*parser)
}
