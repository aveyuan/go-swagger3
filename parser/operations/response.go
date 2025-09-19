package operations

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/iancoleman/orderedmap"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
)

func (p *parser) parseResponseComment(pkgPath, pkgName string, operation *oas.OperationObject, comment string) error {
	// {status}  {jsonType}  {goType}     {description}
	// 201       object      models.User  "User Model"
	// for cases of empty return payload
	// {status} {description}
	// 204 "User Model"
	// for cases of simple types
	// 200 {string} string "..."
	// re := regexp.MustCompile(`(?P<status>[\d]+)[\s]*(?P<jsonType>[\w\{\}]+)?[\s]+(?P<goType>[\w\-\.\/\[\]]+)?[^"]*(?P<description>.*)?`)
	re := regexp.MustCompile(`(?P<status>[\d]+)[\s]*(?P<jsonType>[\w\{\}]+)?[\s]+(?P<goType>[\w\-\.\/\[\]\{\}=]+)?[^"]*(?P<description>.*)?`)

	matches := re.FindStringSubmatch(comment)
	if len(matches) <= 2 {
		return fmt.Errorf("parseResponseComment can not parse response comment \"%s\"", comment)
	}

	status := matches[1]
	statusInt, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("parseResponseComment: http status must be int, but got %s", status)
	}
	if !utils.IsValidHTTPStatusCode(statusInt) {
		return fmt.Errorf("parseResponseComment: Invalid http status code %s", status)
	}

	responseObject := &oas.ResponseObject{
		Content: map[string]*oas.MediaTypeObject{},
	}
	responseObject.Description = strings.Trim(matches[4], "\"")

	switch matches[2] {

	case "object", "array", "{object}", "{array}":
		err = p.complexResponseObject(pkgPath, pkgName, matches[3], responseObject)
	case "{string}", "{integer}", "{boolean}", "string", "integer", "boolean":
		err = p.simpleResponseObject(matches[2], responseObject)
	case "":

	default:
		return fmt.Errorf("parseResponseComment: invalid jsonType %s", matches[2])
	}

	if err != nil {
		return err
	}

	operation.Responses[status] = responseObject
	return nil
}

var genericPattern = regexp.MustCompile(`\{([^}]+)\}`)

// function to parse cases of jsonType in case "object", "array", "{object}", "{array}":
func (p *parser) complexResponseObject(pkgPath, pkgName, typ string, responseObject *oas.ResponseObject) error {
	var containerType string
	// 解析泛型格式：提取外层容器类型和内部字段映射
	matches := genericPattern.FindStringSubmatch(typ)
	fieldMappings := make(map[string]string) // 存储字段名到类型的映射（如 "data" => "feedback.GetQuestionnaireResp"）

	if len(matches) > 1 {
		// 提取外层容器类型（如 "yhttp.DataRes"）
		containerType = strings.TrimSpace(strings.Replace(typ, matches[0], "", 1))
		// 解析内部字段映射（如 "data=feedback.GetQuestionnaireResp, total=int"）
		fieldsPart := matches[1]
		for _, field := range strings.Split(fieldsPart, ",") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			kv := strings.SplitN(field, "=", 2)
			if len(kv) != 2 {
				continue // 跳过格式错误的字段（如未包含=）
			}
			fieldName := strings.TrimSpace(kv[0])
			fieldType := strings.TrimSpace(kv[1])
			fieldMappings[fieldName] = fieldType
		}
	} else {
		// 非泛型结构，直接使用传入的类型
		containerType = typ
	}

	// 处理字段映射（替换容器类型中的指定字段）
	if len(fieldMappings) > 0 {
		// 1. 解析外层容器类型的原始 Schema（包含其自身所有字段）
		containerSchema, err := p.ParseSchemaObject(pkgPath, pkgName, containerType)
		if err != nil {
			return fmt.Errorf("解析容器类型 %s 失败: %v", containerType, err)
		}
		if containerSchema == nil {
			containerSchema = &oas.SchemaObject{Type: "object", Properties: orderedmap.New()}
		}
		if containerSchema.Properties == nil {
			containerSchema.Properties = orderedmap.New()
		}

		// 2. 遍历所有需要替换的字段，逐个更新
		for fieldName, fieldType := range fieldMappings {
			// 解析字段对应的嵌套类型 Schema
			fieldSchema, err := p.ParseSchemaObject(pkgPath, pkgName, fieldType)
			if err != nil {
				return fmt.Errorf("解析字段 %s 的类型 %s 失败: %v", fieldName, fieldType, err)
			}
			if fieldSchema == nil {
				return fmt.Errorf("字段 %s 的类型 %s 解析结果为空", fieldName, fieldType)
			}

			// 3. 用嵌套类型的 Schema 覆盖容器中对应字段
			containerSchema.Properties.Set(fieldName, fieldSchema)
		}

		// 4. 将处理后的容器 Schema 关联到响应
		responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
			Schema: *containerSchema,
		}
		return nil
	}

	// 处理数组、映射的原有逻辑（保持不变）
	re := regexp.MustCompile(`\[\w*\]`)
	goType := re.ReplaceAllString(containerType, "[]")
	if strings.HasPrefix(goType, "map[]") {
		schema, err := p.ParseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.Debug("parseResponseComment cannot parse goType", goType)
		}
		responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
			Schema: *schema,
		}
		return nil
	} else if strings.HasPrefix(goType, "[]") {
		goType = strings.Replace(goType, "[]", "", -1)
		typeName, err := p.RegisterType(pkgPath, pkgName, goType)
		if err != nil {
			return err
		}

		var s oas.SchemaObject
		if utils.IsBasicGoType(typeName) {
			s = oas.SchemaObject{Type: "string"}
		} else {
			s = oas.SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(typeName)}
		}

		responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
			Schema: oas.SchemaObject{
				Type:  "array",
				Items: &s,
			},
		}
		return nil
	}

	// 非泛型类型：直接解析原类型的所有字段（保持不变）
	typeName, err := p.RegisterType(pkgPath, pkgName, containerType)
	if err != nil {
		return err
	}
	if utils.IsBasicGoType(typeName) {
		responseObject.Content[oas.ContentTypeText] = &oas.MediaTypeObject{
			Schema: oas.SchemaObject{Type: "string"},
		}
	} else if utils.IsInterfaceType(typeName) {
		responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
			Schema: oas.SchemaObject{Type: "object"},
		}
	} else {
		responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
			Schema: oas.SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(typeName)},
		}
	}
	return nil
}

func (p *parser) simpleResponseObject(jsonType string, responseObject *oas.ResponseObject) error {
	formattedType := jsonType
	if strings.HasPrefix(jsonType, "{") && strings.HasSuffix(jsonType, "}") {
		formattedType = jsonType[1 : len(jsonType)-1]
	}

	responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{Schema: oas.SchemaObject{Type: formattedType}}
	return nil
}
