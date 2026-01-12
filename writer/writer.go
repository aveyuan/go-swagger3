package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	log "github.com/sirupsen/logrus"
)

type Writer interface {
	Write(openApiObject oas.OpenAPIObject, path string, generateYAML bool, schemaWithoutPkg bool, filterTag string) error
}

type fileWriter struct{}

func NewFileWriter() *fileWriter {
	return &fileWriter{}
}

func filterSchemaWithoutPkg(openApiObject oas.OpenAPIObject) {

	for key := range openApiObject.Components.Schemas {
		key_sep := strings.Split(key, ".")
		key_sep_last := key_sep[len(key_sep)-1]

		_, ok := openApiObject.Components.Schemas[key_sep_last]
		if len(key_sep) > 1 && ok {
			delete(openApiObject.Components.Schemas, key_sep_last)
		}
	}

}

func filterPathsByTag(openApiObject *oas.OpenAPIObject, filterTag string) {
	if filterTag == "" {
		return
	}

	filteredPaths := make(oas.PathsObject)
	for path, pathItem := range openApiObject.Paths {
		filteredPathItem := &oas.PathItemObject{}
		hasMatchingOperation := false

		// Check each HTTP method
		if pathItem.Get != nil && containsTag(pathItem.Get.Tags, filterTag) {
			filteredPathItem.Get = pathItem.Get
			hasMatchingOperation = true
		}
		if pathItem.Post != nil && containsTag(pathItem.Post.Tags, filterTag) {
			filteredPathItem.Post = pathItem.Post
			hasMatchingOperation = true
		}
		if pathItem.Put != nil && containsTag(pathItem.Put.Tags, filterTag) {
			filteredPathItem.Put = pathItem.Put
			hasMatchingOperation = true
		}
		if pathItem.Patch != nil && containsTag(pathItem.Patch.Tags, filterTag) {
			filteredPathItem.Patch = pathItem.Patch
			hasMatchingOperation = true
		}
		if pathItem.Delete != nil && containsTag(pathItem.Delete.Tags, filterTag) {
			filteredPathItem.Delete = pathItem.Delete
			hasMatchingOperation = true
		}
		if pathItem.Options != nil && containsTag(pathItem.Options.Tags, filterTag) {
			filteredPathItem.Options = pathItem.Options
			hasMatchingOperation = true
		}
		if pathItem.Head != nil && containsTag(pathItem.Head.Tags, filterTag) {
			filteredPathItem.Head = pathItem.Head
			hasMatchingOperation = true
		}
		if pathItem.Trace != nil && containsTag(pathItem.Trace.Tags, filterTag) {
			filteredPathItem.Trace = pathItem.Trace
			hasMatchingOperation = true
		}

		// Copy path-level metadata
		filteredPathItem.Ref = pathItem.Ref
		filteredPathItem.Summary = pathItem.Summary
		filteredPathItem.Description = pathItem.Description

		// Only include the path if it has at least one matching operation
		if hasMatchingOperation {
			filteredPaths[path] = filteredPathItem
		}
	}

	openApiObject.Paths = filteredPaths
	log.Infof("Filtered paths by tag '%s', %d paths remaining", filterTag, len(filteredPaths))
}

func containsTag(tags []string, targetTag string) bool {
	for _, tag := range tags {
		if tag == targetTag {
			return true
		}
	}
	return false
}

func (w *fileWriter) Write(openApiObject oas.OpenAPIObject, path string, generateYAML bool, schemaWithoutPkg bool, filterTag string) error {
	if !schemaWithoutPkg {
		filterSchemaWithoutPkg(openApiObject)
	}
	filterPathsByTag(&openApiObject, filterTag)
	log.Info("Writing to open api object file ...")
	fd, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Can not create the file %s: %v", path, err)
	}
	defer fd.Close()

	output, err := json.MarshalIndent(openApiObject, "", "  ")
	if err != nil {
		return err
	}
	if generateYAML {
		output, err = yaml.JSONToYAML(output)
		if err != nil {
			return err
		}
	}
	_, err = fd.WriteString(string(output))
	return err
}
