package apis

import (
	"fmt"
	"go/ast"
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
	}
	p.resolveTypeAliases()
	return nil
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
