package operations

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/schema"
)

type Parser interface {
	Parse(pkgPath, pkgName string, astComments []*ast.Comment) error
}

type parser struct {
	OpenAPI *openApi3Schema.OpenAPIObject

	model.Utils
	schema.Parser
	usedOperationIds map[string]struct{} // Track used operation IDs
}

// pendingRoute holds route info that will be added after tag checking
type pendingRoute struct {
	path      string
	method    string
	operation *openApi3Schema.OperationObject
}

func NewParser(utils model.Utils, api *openApi3Schema.OpenAPIObject, schemaParser schema.Parser) Parser {
	return &parser{
		Utils:            utils,
		OpenAPI:          api,
		Parser:           schemaParser,
		usedOperationIds: make(map[string]struct{}),
	}
}

func (p *parser) Parse(pkgPath, pkgName string, astComments []*ast.Comment) error {
	operation := &openApi3Schema.OperationObject{Responses: map[string]*openApi3Schema.ResponseObject{}}
	var pendingRoute *pendingRoute

	if !strings.HasPrefix(pkgPath, p.ModulePath) || (p.HandlerPath != "" && !strings.HasPrefix(pkgPath, p.HandlerPath)) {
		return nil
	}

	// First pass: collect ONLY tags and route info (no schema parsing yet)
	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if len(comment) == 0 {
			continue
		}

		attribute := strings.Fields(comment)[0]
		switch strings.ToLower(attribute) {
		case "@route", "@router":
			route, err := p.parseRouteInfo(comment)
			if err != nil {
				return err
			}
			pendingRoute = route
			pendingRoute.operation = operation
		case "@resource", "@tag":
			// Only parse tags in first pass
			p.parseResourceAndTag(comment, attribute, operation)
		}
	}

	// If no route was found, nothing to add
	if pendingRoute == nil {
		return nil
	}

	// If filterTag is specified, check if operation has matching tag
	if p.FilterTag != "" && !p.operationMatchesFilter(operation) {
		p.Debugf("Skipping operation due to tag filter: tags=%v, filterTag=%s", operation.Tags, p.FilterTag)
		return nil
	}

	// Second pass: now parse everything including schemas (tag check passed)
	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if len(comment) == 0 {
			continue
		}

		if err := p.parseOperationFromComment(pkgPath, pkgName, comment, operation); err != nil {
			return err
		}
	}

	// Now add the route to OpenAPI
	return p.addRouteToOpenAPI(pendingRoute)
}

// operationMatchesFilter checks if operation matches the filter tag
func (p *parser) operationMatchesFilter(operation *openApi3Schema.OperationObject) bool {
	if p.FilterTag == "" {
		return true
	}
	for _, tag := range operation.Tags {
		if tag == p.FilterTag {
			return true
		}
	}
	return false
}

// validateOperationID checks if an operation ID is unique and registers it if it is.
// Returns an error if the operation ID is already used.
func (p *parser) validateOperationID(operationID string) error {
	if operationID == "" {
		return nil
	}
	if _, exists := p.usedOperationIds[operationID]; exists {
		return fmt.Errorf("operation ID '%s' is not unique", operationID)
	}
	p.usedOperationIds[operationID] = struct{}{}
	return nil
}

func (p *parser) parseOperationFromComment(pkgPath string, pkgName string, comment string, operation *openApi3Schema.OperationObject) error {
	attribute := strings.Fields(comment)[0]
	switch strings.ToLower(attribute) {
	case "@title":
		operation.Summary = strings.TrimSpace(comment[len(attribute):])
	case "@description":
		operation.Description = strings.Join([]string{operation.Description, strings.TrimSpace(comment[len(attribute):])}, " ")
	case "@param":
		return p.parseParamComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@header":
		return p.parseHeaders(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@success", "@failure":
		return p.parseResponseComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@resource", "@tag":
		p.parseResourceAndTag(comment, attribute, operation)
	case "@route", "@router":
		return p.parseRouteComment(operation, comment)
	case "@operationid":
		operationID := strings.TrimSpace(comment[len(attribute):])
		if err := p.validateOperationID(operationID); err != nil {
			return err
		}
		operation.OperationID = operationID
	}
	return nil
}
