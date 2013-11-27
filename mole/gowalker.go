package mole

import (
	"fmt"
	"github.com/joinhack/peony"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//Generate the source code
//All code generator should start with "@"
type CodeGen interface {
	Generate(appName, serverName string, alias map[string]string) string
	BuildAlias(alias map[string][]string)
}

type CodeGenCreater func(comment string, spec CodeGenSpec) (bool, CodeGen)

type MapperCommentCodeGen struct {
	UrlExpr    string
	ActionInfo *ActionInfo
}

func (m *MapperCommentCodeGen) Generate(appName, serverName string, alias map[string]string) string {
	pkgName := alias[m.ActionInfo.ImportPath]
	var code string
	info := m.ActionInfo
	argsList := []string{}
	for _, arg := range m.ActionInfo.Args {
		argsList = append(argsList, fmt.Sprintf("&peony.ArgType{Name:\"%s\", Type:%s}", arg.Name, arg.TypeExpr(alias)))
	}
	var argCode string
	if len(argsList) > 0 {
		argCode = fmt.Sprintf("Args:[]*peony.ArgType{%s}", strings.Join(argsList, ",\n\t\t"))
	}
	if info.RecvName == "" {
		code = fmt.Sprintf("\t%s.FuncMapper(\"%s\", %s.%s, &peony.Action{Name:\"%s\", %s})\n", serverName, m.UrlExpr, pkgName, info.MethodSpec.Name, info.ActionName, argCode)
	} else {
		code = fmt.Sprintf("\t%s.MethodMapper(\"%s\", (*%s.%s).%s, &peony.Action{Name: \"%s\", %s})\n", serverName, m.UrlExpr, pkgName, m.ActionInfo.RecvName, info.Name, info.ActionName, argCode)
	}
	return code
}

func (m *MapperCommentCodeGen) BuildAlias(alias map[string][]string) {
	m.ActionInfo.BuildAlias(alias)
}

var (
	CodeGenCreaters = []CodeGenCreater{}
	MapperRegexp    = regexp.MustCompile("@Mapper\\(\"(.*)\"\\)")
)

func RegisterCodeGenCreater(builder CodeGenCreater) {
	CodeGenCreaters = append(CodeGenCreaters, builder)
}

func init() {
	RegisterCodeGenCreater(MapperCommentCodeGenCreater)
}

func MapperCommentCodeGenCreater(comment string, spec CodeGenSpec) (bool, CodeGen) {
	if actionInfo, ok := spec.(*ActionInfo); ok {
		expr := MapperRegexp.FindStringSubmatch(comment)
		if expr == nil {
			return false, nil
		}
		return true, &MapperCommentCodeGen{expr[1], actionInfo}
	}
	return false, nil
}

type SourceInfo struct {
	Pkgs []*PkgInfo
}

type PkgInfo struct {
	Name       string
	ImportPath string
	CodeGens   []CodeGen
}

type CodeGenSpec interface {
	BuildAlias(map[string][]string)
}

type ActionInfo struct {
	CodeGenSpec
	MethodSpec
	ImportPath string
	ActionName string
}

func (a *ActionInfo) BuildAlias(alias map[string][]string) {
	for _, arg := range a.Args {
		if arg.ImportPath != "" && !contains(alias[arg.Expr.PkgName], arg.ImportPath) {
			alias[arg.Expr.PkgName] = append(alias[arg.Expr.PkgName], arg.ImportPath)
		}
	}
}

type MethodSpec struct {
	Name     string
	RecvName string
	Args     []*MethodArg
}

type MethodArg struct {
	Name       string
	ImportPath string
	Expr       TypeExpr
}

type TypeExpr struct {
	Expr    string
	PkgName string
	PkgIdx  int
	Valid   bool
}

func (m *MethodArg) TypeExpr(alias map[string]string) string {
	isPtr := strings.HasPrefix(m.Expr.Expr, "*")
	if pkgName, ok := alias[m.ImportPath]; ok {
		if isPtr {
			return fmt.Sprintf("reflect.TypeOf((%s%s.%s)(nil))", m.Expr.Expr[:m.Expr.PkgIdx], pkgName, m.Expr.Expr[m.Expr.PkgIdx:])
		} else {
			return fmt.Sprintf("reflect.TypeOf((*%s%s.%s)(nil)).Elem()", m.Expr.Expr[:m.Expr.PkgIdx], pkgName, m.Expr.Expr[m.Expr.PkgIdx:])
		}
	} else {
		//this is for builtin types
		if isPtr {
			return fmt.Sprintf("reflect.TypeOf((%s)(nil))", m.Expr.Expr)
		} else {
			return fmt.Sprintf("reflect.TypeOf((*%s)(nil)).Elem()", m.Expr.Expr)
		}
	}
}

//return recv name and action name
func actionName(funcDecl *ast.FuncDecl) (string, string) {
	var recvName string
	var prefix string
	methodName := funcDecl.Name.Name
	if funcDecl.Recv != nil {
		typ := funcDecl.Recv.List[0].Type
		if starExpr, ok := typ.(*ast.StarExpr); ok {
			prefix = starExpr.X.(*ast.Ident).Name

		} else {
			prefix = typ.(*ast.Ident).Name
		}
		recvName = prefix
		prefix += "."
	}
	return recvName, prefix + methodName
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

func NewTypeExpr(pkgName, importPath string, imports map[string]string, expr ast.Expr) (string, TypeExpr) {
	switch t := expr.(type) {
	case *ast.Ident:
		if IsBultinType(t.Name) {
			pkgName = ""
			importPath = ""
		}
		return importPath, TypeExpr{t.Name, pkgName, 0, true}
	case *ast.SelectorExpr:
		_, e := NewTypeExpr(pkgName, importPath, imports, t.X)
		return imports[e.Expr], TypeExpr{t.Sel.Name, e.Expr, e.PkgIdx, e.Valid}
	case *ast.StarExpr:
		i, e := NewTypeExpr(pkgName, importPath, imports, t.X)
		return i, TypeExpr{"*" + e.Expr, e.PkgName, e.PkgIdx + 1, e.Valid}
	case *ast.ArrayType:
		i, e := NewTypeExpr(pkgName, importPath, imports, t.Elt)
		return i, TypeExpr{"[]" + e.Expr, e.PkgName, e.PkgIdx + 2, e.Valid}
	case *ast.Ellipsis:
		i, e := NewTypeExpr(pkgName, importPath, imports, t.Elt)
		return i, TypeExpr{"[]" + e.Expr, e.PkgName, e.PkgIdx + 3, e.Valid}
	default:
		log.Println("Failed to generate name for field.")
		ast.Print(nil, expr)
	}
	return "", TypeExpr{Valid: false}
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

func processAction(pkgInfo *PkgInfo, imports map[string]string, funcDecl *ast.FuncDecl) {

	if !funcDecl.Name.IsExported() {
		return
	}
	importPath := pkgInfo.ImportPath
	actionInfo := &ActionInfo{ImportPath: importPath}
	actionInfo.Name = funcDecl.Name.Name
	n := len(funcDecl.Type.Params.List)
	if n > 0 {
		actionInfo.Args = make([]*MethodArg, 0, n)
	}
	for _, param := range funcDecl.Type.Params.List {
		importPath, typeExpr := NewTypeExpr(pkgInfo.Name, pkgInfo.ImportPath, imports, param.Type)
		//ignore this action
		if !typeExpr.Valid {
			return
		}
		for _, name := range param.Names {
			actionInfo.Args = append(actionInfo.Args, &MethodArg{
				name.Name,
				importPath,
				typeExpr,
			})
		}
	}
	if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
		for _, comment := range funcDecl.Doc.List {
			if comment.Text[:2] == "//" {
				content := strings.TrimSpace(comment.Text[2:])
				if len(content) == 0 || content[0] != '@' {
					continue
				}
				for _, codeGenCreater := range CodeGenCreaters {
					if ok, codeGen := codeGenCreater(content, actionInfo); ok {
						pkgInfo.CodeGens = append(pkgInfo.CodeGens, codeGen)
					}
				}
			}
		}
	}
	actionInfo.RecvName, actionInfo.ActionName = actionName(funcDecl)

	//actionInfo.Methods = append(actionInfo.Methods, methodSpec)
}

func processFile(file *ast.File, pkgInfo *PkgInfo) {
	imports := map[string]string{}
	processImports(imports, file.Imports)
	for _, decl := range file.Decls {
		switch decl.(type) {
		case *ast.FuncDecl:
			processAction(pkgInfo, imports, decl.(*ast.FuncDecl))
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
			}, parser.ParseComments)

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
							SourceLines: source,
						})
					}
					return rerrList
				}
				ast.Print(nil, err)
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
