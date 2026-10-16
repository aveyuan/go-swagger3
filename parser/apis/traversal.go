package apis

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"
)

func (p *parser) parseImportsAndTypeSpecs() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		if _, ok := p.TypeSpecs[pkgName]; !ok {
			p.TypeSpecs[pkgName] = map[string]*ast.TypeSpec{}
		}
		p.PkgNameImportedPkgAlias[pkgName] = map[string][]string{}

		astPkgs, err := p.schemaParser.GetPkgAst(pkgPath)
		if err != nil {
			if p.RunInStrictMode {
				return fmt.Errorf("parseImportsAndTypeSpecs: parse of %s package cause error: %s", pkgPath, err)
			}
			p.Debugf("parseImportsAndTypeSpecs: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		for _, astPackage := range astPkgs {
			for _, astFile := range astPackage.Files {
				p.parseImportStatementsFromFile(astFile, pkgName)
				p.parseTypeSpecsFromFile(astFile, pkgName)
			}
		}

		// 依赖包（位于 go module cache 中）的完整 AST 在提取完 import 和类型声明后
		// 不再需要（函数体占 AST 内存的绝大部分），及时释放缓存，
		// 避免依赖多的项目（如 gin/gorm/grpc 全家桶）内存暴涨导致 OOM。
		// 类型声明已提取到 p.TypeSpecs，后续解析只依赖它，不受释放影响。
		if !p.isOwnModulePkg(pkgPath) {
			delete(p.PkgPathAstPkgCache, pkgPath)
		}
	}
	p.resolveTypeAliases()
	return nil
}

// isOwnModulePkg 判断包是否属于被解析的项目自身（而非 go module cache 中的依赖包）
func (p *parser) isOwnModulePkg(pkgPath string) bool {
	return pkgPath == p.ModulePath || strings.HasPrefix(pkgPath, p.ModulePath+string(filepath.Separator))
}

func (p *parser) resolveTypeAliases() {
	for pkgName, aliases := range p.TypeAliases {
		for alias, original := range aliases {
			if originalTypeSpec, ok := p.TypeSpecs[pkgName][original]; ok {
				p.TypeSpecs[pkgName][alias] = originalTypeSpec
			}
		}
	}
}

func (p *parser) parseParametersAndPaths() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		// 依赖包的 AST 已在第一阶段提取后释放，不再重新解析。
		// 依赖包不含项目的接口注解（@Route 等），跳过不影响结果。
		if _, ok := p.PkgPathAstPkgCache[pkgPath]; !ok {
			continue
		}
		astPkgs, err := p.schemaParser.GetPkgAst(pkgPath)
		if err != nil {
			if p.RunInStrictMode {
				return fmt.Errorf("parseParametersAndPaths: parse of %s package cause error: %s", pkgPath, err)
			}
			p.Debugf("parseParametersAndPaths: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		for _, astPackage := range astPkgs {
			for _, astFile := range astPackage.Files {
				if err := p.parseParametersAndPathsFromFile(astFile, pkgPath, pkgName); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *parser) parseParametersAndPathsFromFile(astFile *ast.File, pkgPath string, pkgName string) error {
	for _, astDeclaration := range astFile.Decls {
		if err := p.parseFuncDeclaration(astDeclaration, pkgPath, pkgName); err != nil {
			return err
		}
		if err := p.parsePathFromFuncDeclaration(astDeclaration, pkgPath, pkgName); err != nil {
			return err
		}
	}
	return nil
}
