package operations

import (
	"fmt"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"net/http"
	"regexp"
	"strings"
)

// parseRouteInfo extracts route information without adding it to OpenAPI
func (p *parser) parseRouteInfo(comment string) (*pendingRoute, error) {
	sourceString := strings.TrimSpace(comment[len("@Router"):])

	// /path [method]
	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	matches := re.FindStringSubmatch(sourceString)
	if len(matches) != 3 {
		return nil, fmt.Errorf("Can not parse router comment \"%s\", skipped", comment)
	}

	return &pendingRoute{
		path:   matches[1],
		method: strings.ToUpper(matches[2]),
	}, nil
}

// addRouteToOpenAPI adds a pending route to the OpenAPI paths
func (p *parser) addRouteToOpenAPI(route *pendingRoute) error {
	if route == nil {
		return nil
	}

	_, ok := p.OpenAPI.Paths[route.path]
	if !ok {
		p.OpenAPI.Paths[route.path] = &oas.PathItemObject{}
	}

	switch route.method {
	case http.MethodGet:
		p.OpenAPI.Paths[route.path].Get = route.operation
	case http.MethodPost:
		p.OpenAPI.Paths[route.path].Post = route.operation
	case http.MethodPatch:
		p.OpenAPI.Paths[route.path].Patch = route.operation
	case http.MethodPut:
		p.OpenAPI.Paths[route.path].Put = route.operation
	case http.MethodDelete:
		p.OpenAPI.Paths[route.path].Delete = route.operation
	case http.MethodOptions:
		p.OpenAPI.Paths[route.path].Options = route.operation
	case http.MethodHead:
		p.OpenAPI.Paths[route.path].Head = route.operation
	case http.MethodTrace:
		p.OpenAPI.Paths[route.path].Trace = route.operation
	}

	return nil
}

func (p *parser) parseRouteComment(operation *oas.OperationObject, comment string) error {
	sourceString := strings.TrimSpace(comment[len("@Router"):])

	// /path [method]
	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	matches := re.FindStringSubmatch(sourceString)
	if len(matches) != 3 {
		return fmt.Errorf("Can not parse router comment \"%s\", skipped", comment)
	}

	_, ok := p.OpenAPI.Paths[matches[1]]
	if !ok {
		p.OpenAPI.Paths[matches[1]] = &oas.PathItemObject{}
	}

	switch strings.ToUpper(matches[2]) {
	case http.MethodGet:
		p.OpenAPI.Paths[matches[1]].Get = operation
	case http.MethodPost:
		p.OpenAPI.Paths[matches[1]].Post = operation
	case http.MethodPatch:
		p.OpenAPI.Paths[matches[1]].Patch = operation
	case http.MethodPut:
		p.OpenAPI.Paths[matches[1]].Put = operation
	case http.MethodDelete:
		p.OpenAPI.Paths[matches[1]].Delete = operation
	case http.MethodOptions:
		p.OpenAPI.Paths[matches[1]].Options = operation
	case http.MethodHead:
		p.OpenAPI.Paths[matches[1]].Head = operation
	case http.MethodTrace:
		p.OpenAPI.Paths[matches[1]].Trace = operation
	}

	return nil
}
