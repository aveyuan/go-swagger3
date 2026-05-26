package module

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/utils"
	log "github.com/sirupsen/logrus"
)

type Parser interface {
	Parse() error
}

type parser struct {
	model.Utils
}

func NewParser(utils model.Utils) Parser {
	return &parser{
		Utils: utils,
	}
}

// Parse parse sub-package
func (p *parser) Parse() error {
	log.Info("Parsing Modules ...")
	walker := func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if utils.ShouldSkipDir(p.ModulePath, path) {
			return filepath.SkipDir
		}
		if !utils.HasGoFiles(path) {
			return nil
		}
		name := filepath.Join(p.ModuleName, strings.TrimPrefix(path, p.ModulePath))
		name = filepath.ToSlash(name)
		p.KnownPkgs = append(p.KnownPkgs, model.Pkg{
			Name: name,
			Path: path,
		})
		p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		return nil
	}
	return filepath.WalkDir(p.ModulePath, walker)
}
