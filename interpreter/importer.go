package interpreter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/parser"
	"github.com/pkg/errors"
)

const pkgdir = "ifql_pkgs"

// FileImporter implements Importer using the local filesystem.
// Package are expected to be in the `ifql_pkgs` directories.
type FileImporter struct {
	cache map[string]*SourcePackage
}

func NewFileImporter() *FileImporter {
	return &FileImporter{
		cache: make(map[string]*SourcePackage),
	}
}

func (fi *FileImporter) Import(p string, dir string) (Package, error) {
	fp := filepath.Join(dir, pkgdir, p)
	if pkg, ok := fi.cache[fp]; ok {
		return pkg, nil
	}

	if fi, err := os.Stat(fp); os.IsNotExist(err) || !fi.IsDir() {
		return nil, fmt.Errorf("could not find package %q at %q", p, fp)
	}

	srcFiles, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read package %q", p)
	}

	files := make(map[string]*ast.File)
	for _, fi := range srcFiles {
		if fi.IsDir() || filepath.Ext(fi.Name()) != ".ifql" {
			continue
		}
		f, err := os.Open(filepath.Join(fp, fi.Name()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open package file %q", fi.Name())
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read package file %q", fi.Name())
		}
		file, err := parser.NewAST(string(data))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse package file %q", fi.Name())
		}
		if file.Package == nil {
			return nil, fmt.Errorf("no package name declared in file %q", fi.Name())
		}
		files[fi.Name()] = file
	}

	pkgFile := ast.MergePackageFiles(files)

	pkg := &SourcePackage{
		path: p,
		file: pkgFile,
	}
	fi.cache[fp] = pkg
	return pkg, nil
}

type SourcePackage struct {
	path  string
	scope *Scope

	file *ast.File
}

func (p *SourcePackage) String() string {
	return fmt.Sprintf("package: %q path: %q scope: %p", p.name, p.path, p.scope)
}
func (p *SourcePackage) Name() string {
	return p.file.Package.Name
}

func (p *SourcePackage) Path() string {
	return p.path
}

func (p *SourcePackage) Complete() bool {
	return p.scope != nil
}

func (p *SourcePackage) Scope() *Scope {
	return p.scope
}

func (p *SourcePackage) SetScope(scope *Scope) {
	p.scope = scope
}

func (p *SourcePackage) File() *ast.File {
	return p.file
}

var DisabledImporter = disabledImporter{}

type disabledImporter struct{}

func (disabledImporter) Import(path string, dir string) (Package, error) {
	return nil, errors.New("imports disabled")
}
