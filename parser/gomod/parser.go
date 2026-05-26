package gomod

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"

	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/utils"
	"golang.org/x/mod/modfile"
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

// Parse parse go.mod info
func (p *parser) Parse() error {
	log.Info("Parsing GoMod Info ...")
	b, err := os.ReadFile(p.GoModFilePath)
	if err != nil {
		return err
	}
	file, err := modfile.Parse(p.GoModFilePath, b, nil)
	if err != nil {
		return err
	}
	for i := range file.Require {
		if err = p.parseGoModFilePackages(file.Require[i].Mod.Path, file.Require[i].Mod.Version); err != nil {
			return err
		}
	}
	if p.RunInDebugMode {
		for i := range p.KnownPkgs {
			p.Debugf(p.KnownPkgs[i].Name, "->", p.KnownPkgs[i].Path)
		}
	}
	return nil
}

func (p *parser) parseGoModFilePackages(pkgName string, version string) error {
	pathRunes := []rune{}
	for _, v := range pkgName {
		if !unicode.IsUpper(v) {
			pathRunes = append(pathRunes, v)
			continue
		}
		pathRunes = append(pathRunes, '!')
		pathRunes = append(pathRunes, unicode.ToLower(v))
	}
	pkgPath := filepath.Join(p.GoModCachePath, string(pathRunes)+"@"+version)
	pkgName = filepath.ToSlash(pkgName)
	p.KnownPkgs = append(p.KnownPkgs, model.Pkg{
		Name: pkgName,
		Path: pkgPath,
	})
	p.KnownNamePkg[pkgName] = &p.KnownPkgs[len(p.KnownPkgs)-1]
	p.KnownPathPkg[pkgPath] = &p.KnownPkgs[len(p.KnownPkgs)-1]

	return filepath.WalkDir(pkgPath, p.walkerFunc(pkgName, pkgPath))
}

func (p *parser) walkerFunc(pkgName string, pkgPath string) func(path string, entry os.DirEntry, err error) error {
	return func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if utils.ShouldSkipDir(pkgPath, path) {
			return filepath.SkipDir
		}
		if !utils.HasGoFiles(path) {
			return nil
		}
		name := filepath.Join(pkgName, strings.TrimPrefix(path, pkgPath))
		name = filepath.ToSlash(name)
		p.KnownPkgs = append(p.KnownPkgs, model.Pkg{
			Name: name,
			Path: path,
		})
		p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		return nil
	}
}
