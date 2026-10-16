package parser

import (
	"fmt"
	"github.com/aveyuan/go-swagger3/parser/utils"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func (p *parser) verifyAndSetPaths() error {
	// check modulePath is exist
	if err := p.verifyModulePath(); err != nil {
		return err
	}

	// check go.mod file is exist
	if err := p.setGoModFilePath(); err != nil {
		return err
	}

	// check mainFilePath is exist
	if err := p.verifyMainFilePath(); err != nil {
		return err
	}

	// get module name from go.mod file
	if err := p.getModuleNameFromGoModFile(); err != nil {
		return err
	}

	// check go module cache path is exist ($GOPATH/pkg/mod)
	if err := p.setGoModCachePath(); err != nil {
		return err
	}

	// check handlerPath is exist
	if err := p.setHandlerPath(); err != nil {
		return err
	}
	return nil
}

func (p *parser) setHandlerPath() error {
	var err error
	if p.HandlerPath != "" {
		p.HandlerPath, err = filepath.Abs(p.HandlerPath)
		if err != nil {
			return err
		}
		_, err := os.Stat(p.HandlerPath)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			return fmt.Errorf("cannot get information of %s: %s", p.HandlerPath, err)
		}
	}
	p.Debugf("handler path: %s", p.HandlerPath)
	return nil
}

func (p *parser) setGoModCachePath() error {
	candidates := goModCacheCandidates()
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		p.GoModCachePath = path
		p.Debugf("go module cache path: %s", p.GoModCachePath)
		return nil
	}
	return fmt.Errorf("cannot find go module cache path, tried: %s. please make sure go modules are downloaded (go mod download), or set the GOMODCACHE environment variable to specify the path manually", strings.Join(candidates, ", "))
}

// goModCacheCandidates returns possible go module cache paths in priority order:
//  1. output of `go env GOMODCACHE` (the actual value configured by the go toolchain,
//     which respects GOMODCACHE env and go env -w overrides)
//  2. GOMODCACHE environment variable
//  3. $GOPATH/pkg/mod
//  4. $HOME/go/pkg/mod
func goModCacheCandidates() []string {
	var candidates []string
	seen := make(map[string]bool)
	add := func(path string) {
		if path != "" && !seen[path] {
			seen[path] = true
			candidates = append(candidates, path)
		}
	}

	if out, err := exec.Command("go", "env", "GOMODCACHE").Output(); err == nil {
		add(strings.TrimSpace(string(out)))
	}
	add(os.Getenv("GOMODCACHE"))

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		if current, err := user.Current(); err == nil {
			goPath = filepath.Join(current.HomeDir, "go")
		}
	}
	if goPath != "" {
		add(filepath.Join(goPath, "pkg", "mod"))
	}
	return candidates
}

func (p *parser) getModuleNameFromGoModFile() error {
	moduleName := utils.GetModuleNameFromGoMod(p.GoModFilePath)
	if moduleName == "" {
		return fmt.Errorf("cannot get module name from %s", p.GoModFilePath)
	}
	p.ModuleName = moduleName
	p.Debugf("module name: %s", p.ModuleName)
	return nil
}

func (p *parser) verifyMainFilePath() error {
	if p.MainFilePath == "" {
		fns, err := filepath.Glob(filepath.Join(p.ModulePath, "*.go"))
		if err != nil {
			return err
		}
		for _, fn := range fns {
			if utils.IsMainFile(fn) {
				p.MainFilePath = fn
				break
			}
		}
	} else {
		mainFileInfo, err := os.Stat(p.MainFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			return fmt.Errorf("cannot get information of %s: %s", p.MainFilePath, err)
		}
		if mainFileInfo.IsDir() {
			return fmt.Errorf("mainFilePath should not be a directory")
		}
	}
	p.Debugf("main file path: %s", p.MainFilePath)
	return nil
}

func (p *parser) setGoModFilePath() error {
	goModFilePath := filepath.Join(p.ModulePath, "go.mod")
	goModFileInfo, err := os.Stat(goModFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("cannot get information of %s: %s", goModFilePath, err)
	}
	if goModFileInfo.IsDir() {
		return fmt.Errorf("%s should be a file", goModFilePath)
	}
	p.GoModFilePath = goModFilePath
	p.Debugf("go.mod file path: %s", p.GoModFilePath)
	return err
}

func (p *parser) verifyModulePath() error {
	var err error
	p.ModulePath, err = filepath.Abs(p.ModulePath)
	if err != nil {
		return err
	}
	moduleInfo, err := os.Stat(p.ModulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("cannot get information of %s: %s", p.ModulePath, err)
	}
	if !moduleInfo.IsDir() {
		return fmt.Errorf("modulePath should be a directory")
	}
	p.Debugf("module path: %s", p.ModulePath)
	return err
}
