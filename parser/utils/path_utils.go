package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func HasGoFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}
	return false
}

func ShouldSkipDir(root, path string) bool {
	if root == path {
		return false
	}
	name := filepath.Base(path)
	if name == "vendor" || name == "testdata" || name == "node_modules" {
		return true
	}
	return strings.HasPrefix(name, ".")
}
