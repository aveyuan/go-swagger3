package schema

import (
	"github.com/iancoleman/orderedmap"
	. "github.com/aveyuan/go-swagger3/openApi3Schema"
	"github.com/aveyuan/go-swagger3/parser/utils"
	"strings"
)

func (p *parser) parseBasicTypeSchemaObject(pkgPath string, pkgName string, typeName string) (*SchemaObject, error, bool) {
	var schemaObject SchemaObject
	var err error
	// handler basic and some specific typeName
	if strings.HasPrefix(typeName, "[]") {
		return p.parseArrayType(pkgPath, pkgName, typeName, schemaObject, err)
	} else if IsMapType(typeName) {
		return p.parseMapType(pkgPath, pkgName, typeName, schemaObject)
	} else if typeName == "time.Time" {
		return p.parseTimeType(schemaObject)
	} else if strings.HasPrefix(typeName, "interface{}") {
		return p.parseInterfaceType()
	} else if utils.IsGoTypeOASType(typeName) {
		return p.parseBasicGoType(schemaObject, typeName)
	}
	return nil, nil, false
}

func IsMapType(typeName string) bool {
	return strings.HasPrefix(typeName, "map[]") || strings.HasPrefix(typeName, "map[")
}

func (p *parser) parseBasicGoType(schemaObject SchemaObject, typeName string) (*SchemaObject, error, bool) {
	schemaObject.Type = utils.GoTypesOASTypes[typeName]
	schemaObject.Format = utils.GoTypesOASFormats[typeName]
	return &schemaObject, nil, true
}

func (p *parser) parseInterfaceType() (*SchemaObject, error, bool) {
	return &SchemaObject{Type: "object"}, nil, true
}

func (p *parser) parseTimeType(schemaObject SchemaObject) (*SchemaObject, error, bool) {
	schemaObject.Type = "string"
	schemaObject.Format = "date-time"
	return &schemaObject, nil, true
}

func (p *parser) parseArrayType(pkgPath string, pkgName string, typeName string, schemaObject SchemaObject, err error) (*SchemaObject, error, bool) {
	schemaObject.Type = "array"
	itemTypeName := typeName[2:]
	schemaObject.Items, err = p.parseArrayItemSchema(pkgPath, pkgName, itemTypeName)
	if err != nil {
		return nil, err, true
	}
	return &schemaObject, nil, true
}

func (p *parser) parseArrayItemSchema(pkgPath string, pkgName string, itemTypeName string) (*SchemaObject, error) {
	if itemTypeName == "time.Time" {
		return &SchemaObject{Type: "string", Format: "date-time"}, nil
	}
	if utils.IsInterfaceType(itemTypeName) {
		return &SchemaObject{Type: "object"}, nil
	}
	if utils.IsGoTypeOASType(itemTypeName) {
		return &SchemaObject{
			Type:   utils.GoTypesOASTypes[itemTypeName],
			Format: utils.GoTypesOASFormats[itemTypeName],
		}, nil
	}

	schema, ok := p.KnownIDSchema[utils.GenSchemaObjectID(pkgName, itemTypeName, p.SchemaWithoutPkg)]
	if ok && schema.ID != "" {
		return &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(schema.ID)}, nil
	}

	typeName, err := p.RegisterType(pkgPath, pkgName, itemTypeName)
	if err != nil {
		return nil, err
	}
	if utils.IsGoTypeOASType(typeName) {
		return &SchemaObject{
			Type:   utils.GoTypesOASTypes[typeName],
			Format: utils.GoTypesOASFormats[typeName],
		}, nil
	}
	if utils.IsInterfaceType(typeName) {
		return &SchemaObject{Type: "object"}, nil
	}
	if typeName == "" {
		return p.ParseSchemaObject(pkgPath, pkgName, itemTypeName)
	}
	return &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(typeName)}, nil
}

func (p *parser) parseMapType(pkgPath string, pkgName string, typeName string, schemaObject SchemaObject) (*SchemaObject, error, bool) {
	schemaObject.Type = "object"
	itemTypeName := mapValueType(typeName)
	schema, ok := p.KnownIDSchema[utils.GenSchemaObjectID(pkgName, itemTypeName, p.SchemaWithoutPkg)]
	if ok {
		schemaObject.Items = &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(schema.ID)}
		return &schemaObject, nil, true
	}
	schemaProperty, err := p.ParseSchemaObject(pkgPath, pkgName, itemTypeName)
	if err != nil {
		return nil, err, true
	}
	schemaObject.Properties = orderedmap.New()
	schemaObject.Properties.Set("key", schemaProperty)
	return &schemaObject, nil, true
}

func mapValueType(typeName string) string {
	if strings.HasPrefix(typeName, "map[]") {
		return typeName[5:]
	}
	if strings.HasPrefix(typeName, "map[") {
		end := strings.Index(typeName, "]")
		if end > -1 && end+1 < len(typeName) {
			return typeName[end+1:]
		}
	}
	return typeName
}
