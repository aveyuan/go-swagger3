package operations

import (
	"fmt"
	"strings"

	"github.com/iancoleman/orderedmap"
	oas "github.com/aveyuan/go-swagger3/openApi3Schema"
	"github.com/aveyuan/go-swagger3/parser/utils"
)

func (p *parser) parseParamComment(pkgPath, pkgName string, operation *oas.OperationObject, comment string) error {
	// {name}  {in}  {goType}  {required}  {description}		{example (optional)}
	// user    body  User      true        "Info of a user."
	// f       file  ignored   true        "Upload a file." 	"/home/arlet/go-swagger3/main.go"
	fields, err := parseParamFields(comment)
	if err != nil {
		return err
	}

	parameterObject := oas.ParameterObject{}
	appendName(&parameterObject, fields[0])
	appendIn(&parameterObject, fields[1])
	appendRequired(&parameterObject, fields[3])
	appendDescription(&parameterObject, fields[4])
	appendExample(&parameterObject, fields[5])

	goType := normalizeGoType(fields[2])

	switch parameterObject.In {
	// file, form
	case "form", "file":
		appendRequestBody(operation, parameterObject, goType)
		return nil
	// body
	case "body":
		return p.parseRequestBody(pkgPath, pkgName, operation, parameterObject, goType)

	// path, query, header, cookie
	default:
		return p.appendQueryParam(pkgPath, pkgName, operation, parameterObject, goType)
	}
}

func (p *parser) parseRequestBody(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) error {
	if operation.RequestBody == nil {
		operation.RequestBody = &oas.RequestBodyObject{
			Content:  map[string]*oas.MediaTypeObject{},
			Required: parameterObject.Required,
		}
	}
	if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[]") || goType == "time.Time" {
		return p.parseArrayMapOrTimeType(pkgPath, pkgName, operation, goType)
	}
	return p.parseGoBasicTypeOrStructType(pkgPath, pkgName, operation, goType)
}

func (p *parser) parseGoBasicTypeOrStructType(pkgPath string, pkgName string, operation *oas.OperationObject, goType string) error {
	typeName, err := p.RegisterType(pkgPath, pkgName, goType)
	if err != nil {
		return err
	}
	if utils.IsBasicGoType(typeName) {
		operation.RequestBody.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{Schema: oas.SchemaObject{Type: "string"}}
		return nil
	}
	operation.RequestBody.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{Schema: oas.SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(typeName)}}
	return nil
}

func (p *parser) parseArrayMapOrTimeType(pkgPath string, pkgName string, operation *oas.OperationObject, goType string) error {
	parsedSchemaObject, err := p.ParseSchemaObject(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parseResponseComment cannot parse goType", goType)
		return err
	}
	if parsedSchemaObject != nil {
		operation.RequestBody.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{Schema: *parsedSchemaObject}
	}
	return nil
}

func (p *parser) appendQueryParam(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) error {
	if parameterObject.In == "path" {
		parameterObject.Required = true
	}
	if goType == "time.Time" {
		return p.appendTimeParam(pkgPath, pkgName, operation, parameterObject, goType)
	}
	if utils.IsGoTypeOASType(goType) {
		p.appendGoTypeParams(parameterObject, goType, operation)
	}
	if utils.IsEnumType(goType) {
		p.appendEnumParamRef(goType, parameterObject, operation)
		return nil
	}
	if parameterObject.Name == "." {
		schema, err := p.ParseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			return err
		}
		if schema.Properties == nil {
			return fmt.Errorf("NilSchemaProperties : parseHeaders can not parse Header schema %s", "")
		}
		for _, key := range schema.Properties.Keys() {

			prop, ok := schema.Properties.Get(key)
			if !ok {
				continue
			}
			propObject, ok := prop.(*oas.SchemaObject)
			if !ok {
				continue
			}

			requiredFunc := func() bool {

				return len(propObject.Required) > 0
			}

			temp := &oas.ParameterObject{
				Name:        key,
				In:          parameterObject.In,
				Description: propObject.Description,
				Required:    requiredFunc(),
				Example:     propObject.Example,
				Schema:      propObject,
				Ref:         propObject.Ref,
			}

			p.appendGoTypeParams(*temp, propObject.Type, operation)
		}
		return nil
	}

	return nil
}

func (p *parser) appendTimeParam(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) (err error) {
	parameterObject.Schema, err = p.ParseSchemaObject(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parseResponseComment cannot parse goType", goType)
	}
	operation.Parameters = append(operation.Parameters, parameterObject)
	return err
}

func (p *parser) appendGoTypeParams(parameterObject oas.ParameterObject, goType string, operation *oas.OperationObject) {
	parameterObject.Schema = &oas.SchemaObject{
		Type:        utils.GoTypesOASTypes[goType],
		Format:      utils.GoTypesOASFormats[goType],
		Description: parameterObject.Description,
	}
	operation.Parameters = append(operation.Parameters, parameterObject)
}

func (p *parser) appendModelSchemaRef(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) error {
	typeName, err := p.RegisterType(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parse param model type failed", goType)
		return err
	}
	parameterObject.Schema = &oas.SchemaObject{
		Ref:  utils.AddSchemaRefLinkPrefix(typeName),
		Type: typeName,
	}
	operation.Parameters = append(operation.Parameters, parameterObject)
	return nil
}

func (p *parser) appendEnumParamRef(goType string, parameterObject oas.ParameterObject, operation *oas.OperationObject) {
	if strings.Contains(goType, "model.") {
		goType = strings.Replace(goType, "model.", "", -1)
	}
	parameterObject.Schema = &oas.SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(goType)}
	operation.Parameters = append(operation.Parameters, parameterObject)
}

func appendRequestBody(operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) {
	if !(parameterObject.In == "file" || parameterObject.In == "form") {
		return
	}
	if operation.RequestBody == nil {
		operation.RequestBody = &oas.RequestBodyObject{
			Content: map[string]*oas.MediaTypeObject{
				oas.ContentTypeForm: {Schema: oas.SchemaObject{Type: "object", Properties: orderedmap.New()}},
			},
			Required: parameterObject.Required,
		}
	}
	if parameterObject.In == "file" {
		operation.RequestBody.Content[oas.ContentTypeForm].Schema.Properties.Set(parameterObject.Name, &oas.SchemaObject{
			Type:        "string",
			Format:      "binary",
			Description: parameterObject.Description,
		})
	}
	if utils.IsGoTypeOASType(goType) {
		operation.RequestBody.Content[oas.ContentTypeForm].Schema.Properties.Set(parameterObject.Name, &oas.SchemaObject{
			Type:        utils.GoTypesOASTypes[goType],
			Format:      utils.GoTypesOASFormats[goType],
			Description: parameterObject.Description,
		})
	}
}

func appendRequired(paramObject *oas.ParameterObject, isRequired string) {
	switch strings.ToLower(isRequired) {
	case "true", "required":
		paramObject.Required = true
	}
}

func appendDescription(parameterObject *oas.ParameterObject, description string) {
	parameterObject.Description = description
}

func appendIn(parameterObject *oas.ParameterObject, in string) {
	parameterObject.In = in
}

func appendName(parameterObject *oas.ParameterObject, name string) {
	parameterObject.Name = name
}

func appendExample(parameterObject *oas.ParameterObject, example string) {
	if example == "" {
		parameterObject.Example = nil
	} else {
		parameterObject.Example = example
	}
}
