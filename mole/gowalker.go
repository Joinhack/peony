package mole

import (
	"github.com/joinhack/peony"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type SourceInfo struct {
	Pkgs []*PkgInfo
}

type PkgInfo struct {
	Name       string
	ImportPath string
	Actions    ActionInfo
}

type ActionInfo struct {
	Methods []*MethodExpr
}

type MethodExpr struct {
	Name       string
	ImportPath string
	ActionName string
	Args       []*MethodArg
}

type MethodArg struct {
	Name       string
	ImportPath string
	Expr       TypeExpr
}

type TypeExpr struct {
	Expr    string
	PkgName string
	Valid   bool
}

func actionName(funcDecl *ast.FuncDecl) string {
	prefix := ""
	methodName := funcDecl.Name.Name
	if funcDecl.Recv != nil {
		typ := funcDecl.Recv.List[0].Type
		if starExpr, ok := typ.(*ast.StarExpr); ok {
			prefix = starExpr.X.(*ast.Ident).Name

		} else {
			prefix = typ.(*ast.Ident).Name
		}
		prefix += "."
	}
	return prefix + methodName
}

var _BUILTIN_TYPES = map[string]struct{}{
	"bool":       struct{}{},
	"byte":       struct{}{},
	"complex128": struct{}{},
	"complex64":  struct{}{},
	"error":      struct{}{},
	"float32":    struct{}{},
	"float64":    struct{}{},
	"int":        struct{}{},
	"int16":      struct{}{},
	"int32":      struct{}{},
	"int64":      struct{}{},
	"int8":       struct{}{},
	"rune":       struct{}{},
	"string":     struct{}{},
	"uint":       struct{}{},
	"uint16":     struct{}{},
	"uint32":     struct{}{},
	"uint64":     struct{}{},
	"uint8":      struct{}{},
	"uintptr":    struct{}{},
}

func IsBultinType(name string) bool {
	_, ok := _BUILTIN_TYPES[name]
	return ok
}

func NewTypeExpr(pkgName string, expr ast.Expr) TypeExpr {
	switch t := expr.(type) {
	case *ast.Ident:
		if IsBultinType(t.Name) {
			pkgName = ""
		}
		return TypeExpr{t.Name, pkgName, true}
	case *ast.SelectorExpr:
		e := NewTypeExpr(pkgName, t.X)
		return TypeExpr{t.Sel.Name, e.Expr, e.Valid}
	case *ast.StarExpr:
		e := NewTypeExpr(pkgName, t.X)
		return TypeExpr{"*" + e.Expr, e.PkgName, e.Valid}
	case *ast.ArrayType:
		e := NewTypeExpr(pkgName, t.Elt)
		return TypeExpr{"[]" + e.Expr, e.PkgName, e.Valid}
	case *ast.Ellipsis:
		e := NewTypeExpr(pkgName, t.Elt)
		return TypeExpr{"[]" + e.Expr, e.PkgName, e.Valid}
	default:
		log.Println("Failed to generate name for field.")
		ast.Print(nil, expr)
	}
	return TypeExpr{Valid: false}
}

func processImports(imports map[string]string, importSpecs []*ast.ImportSpec) {
	for _, importSpec := range importSpecs {
		alias := ""
		if importSpec.Name != nil {
			alias = importSpec.Name.Name
			if alias == "_" {
				continue
			}
		}
		value := importSpec.Path.Value
		path := value[1 : len(value)-1]
		if alias == "" {
			pkg, err := build.Import(path, "", 0)
			if err != nil {
				log.Println(err)
				continue
			}
			alias = pkg.Name
		}
		imports[alias] = path
	}
}

func processAction(actionInfo *ActionInfo, initImportPath, pkgName string, imports map[string]string, funcDecl *ast.FuncDecl) {

	if !funcDecl.Name.IsExported() {
		return
	}
	importPath := initImportPath
	MethodExpr := &MethodExpr{Name: funcDecl.Name.Name, ImportPath: initImportPath}
	MethodExpr.ImportPath = importPath
	n := len(funcDecl.Type.Params.List)
	if n > 0 {
		MethodExpr.Args = make([]*MethodArg, 0, n)
	}
	for _, param := range funcDecl.Type.Params.List {
		typeExpr := NewTypeExpr(pkgName, param.Type)
		//ignore this action
		if !typeExpr.Valid {
			return
		}
		if typeExpr.PkgName != "" {
			var ok bool
			importPath, ok = imports[typeExpr.PkgName]
			if !ok && typeExpr.PkgName != pkgName {
				log.Println("unknown package path:", typeExpr.PkgName)
			}
			// if importPath is not exits, I guess it's should be default importPath
			if importPath == "" {
				importPath = initImportPath
			}
		}
		for _, name := range param.Names {
			MethodExpr.Args = append(MethodExpr.Args, &MethodArg{
				name.Name,
				importPath,
				typeExpr,
			})
		}
	}
	MethodExpr.ActionName = actionName(funcDecl)
	actionInfo.Methods = append(actionInfo.Methods, MethodExpr)
}

func processFile(file *ast.File, pkgInfo *PkgInfo) {
	actions := ActionInfo{}
	imports := map[string]string{}
	processImports(imports, file.Imports)
	for _, decl := range file.Decls {
		switch decl.(type) {
		case *ast.FuncDecl:
			processAction(&actions, pkgInfo.ImportPath, pkgInfo.Name, imports, decl.(*ast.FuncDecl))
		case *ast.GenDecl:
			genDecl := decl.(*ast.GenDecl)
			if genDecl.Tok != token.TYPE || len(genDecl.Specs) != 1 {
				continue
			}
			//spec := genDecl.Specs[0]
			//var typeSpec *ast.TypeSpec
			//typeSpec = spec.(*ast.TypeSpec)
		}

	}
	pkgInfo.Actions = actions
}

func processPackage(si *SourceInfo, importPath string, pkg *ast.Package) {
	pkgInfo := &PkgInfo{ImportPath: importPath, Name: pkg.Name}
	if pkg.Name == "controllers" {
		for _, file := range pkg.Files {
			processFile(file, pkgInfo)
		}
	}
	si.Pkgs = append(si.Pkgs, pkgInfo)
}

func NewSourceInfo() *SourceInfo {
	s := &SourceInfo{}
	return s
}

//analzy the inmport path
func importPathFromPath(src string) string {
	for _, path := range filepath.SplitList(build.Default.GOPATH) {
		path = filepath.Join(path, "src")
		if strings.HasPrefix(src, path) {
			return filepath.ToSlash(src[len(path)+1:])
		}
	}
	goroot := filepath.Join(build.Default.GOROOT, "src", "pkg")
	if strings.HasPrefix(src, goroot) {
		peony.WARN.Println("Source should in GOPATH, but found it in GOROOT")
		return filepath.ToSlash(src[len(goroot)+1:])
	}
	peony.ERROR.Println("Unexpected! Code path is not in GOPATH:", src)
	return ""
}

func ProcessSources(roots []string) (*SourceInfo, error) {
	si := NewSourceInfo()
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Println("error scan app source:", err)
				return nil
			}
			//if is normal file or name is temp skip
			//directory is needed
			if !info.IsDir() || info.Name() == "tmp" {
				return nil
			}

			fileSet := token.NewFileSet()
			astPkgs, err := parser.ParseDir(fileSet, path, func(info os.FileInfo) bool {
				name := info.Name()
				return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
			}, 0)

			if err != nil {
				//err is ErrorList
				if errList, ok := err.(scanner.ErrorList); ok {
					fileSources := map[string][]string{}
					rerrList := make(peony.ErrorList, 0, len(errList))
					for _, err := range errList {
						var hasSource = false
						var source []string
						if source, hasSource = fileSources[err.Pos.Filename]; !hasSource {
							source = peony.MustReadLines(err.Pos.Filename)
							fileSources[err.Pos.Filename] = source
						}

						rerrList = append(rerrList, &peony.Error{
							Title:       "Compile error",
							FileName:    err.Pos.Filename,
							Path:        err.Pos.Filename,
							Description: err.Msg,
							Line:        err.Pos.Line,
							Column:      err.Pos.Column,
							SouceLines:  source,
						})
					}
					return rerrList
				}
			}

			//parse the importPath
			importPath := importPathFromPath(root)
			if path != root {
				importPath = importPathFromPath(path)
			}

			//ignore the main package
			delete(astPkgs, "main")
			//ignore the empty package
			if len(astPkgs) == 0 {
				return nil
			}

			for _, pkg := range astPkgs {
				processPackage(si, importPath, pkg)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

	}
	return si, nil
}
