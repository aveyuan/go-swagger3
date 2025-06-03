package operations

import (
	"go/ast"

	"github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/stretchr/testify/mock"
)

type MockSchemaParser struct {
	mock.Mock
}

func (m *MockSchemaParser) GetPkgAst(pkgPath string) (map[string]*ast.Package, error) {
	args := m.Called(pkgPath)
	return args.Get(0).(map[string]*ast.Package), args.Error(1)
}

func (m *MockSchemaParser) RegisterType(pkgPath, pkgName, typeName string) (string, error) {
	args := m.Called(pkgPath, pkgName, typeName)
	return args.String(0), args.Error(1)
}

func (m *MockSchemaParser) ParseSchemaObject(pkgPath, pkgName, typeName string) (*openApi3Schema.SchemaObject, error) {
	args := m.Called(pkgPath, pkgName, typeName)
	if obj := args.Get(0); obj != nil {
		return obj.(*openApi3Schema.SchemaObject), args.Error(1)
	}
	return nil, args.Error(1)
}
