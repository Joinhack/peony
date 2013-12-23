package mole

import (
	"errors"
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
	"strings"
	"text/template"
)

//Generate the source code
//All code generator should start with "@"
type CodeGen interface {
	Generate(appName, serverName string, alias map[string]string) string
	BuildAlias(alias map[string][]string)
}

type CodeGenCreater func(comment string, spec CodeGenSpec) (CodeGen, error)

type MapperCommentCodeGen struct {
	Ignore bool
	*ActionInfo
	UrlExpr     string
	HttpMethods []string
}

type InterceptCommentCodeGen struct {
	*ActionInfo
	When     int
	Priority int
}

var mapperTemplate = template.Must(template.New("").Parse(`
	{{.serverName}}.{{.funcOrMethod}}("{{.url}}", []string{"{{.httpMethods}}"}, 
		{{.action}}, &peony.Action{
			Name: "{{.info.ActionName}}",
			{{if .info.Args}}
			Args: []*peony.ArgType{ 
				{{range .info.Args}}
				&peony.ArgType{
					Name: "{{.Name}}", 
					Type: {{.TypeExpr $.alias}},
				},
			{{end}}}{{end}}},
	)
`))

//generate the code for mapper tag
func (m *MapperCommentCodeGen) Generate(appName, serverName string, alias map[string]string) string {
	pkgName := alias[m.ImportPath]
	var code string
	info := m.ActionInfo
	httpMethods := strings.Join(m.HttpMethods, `", "`)
	url := m.UrlExpr
	params := map[string]interface{}{
		"info":        info,
		"httpMethods": httpMethods,
		"serverName":  serverName,
		"alias":       alias,
		"url":         url,
	}
	if info.RecvName == "" {
		params["funcOrMethod"] = "FuncMapper"
		params["action"] = pkgName + "." + info.ActionName
		code = peony.ExecuteTemplate(mapperTemplate, params)
	} else {
		params["funcOrMethod"] = "MethodMapper"
		params["action"] = fmt.Sprintf("(*%s.%s).%s", pkgName, info.RecvName, info.Name)
		code = peony.ExecuteTemplate(mapperTemplate, params)
	}
	return code
}

//generate the code for intercept tag
func (i *InterceptCommentCodeGen) Generate(appName, serverName string, alias map[string]string) string {
	pkgName := alias[i.ImportPath]
	var code string
	info := i.ActionInfo
	code = fmt.Sprintf("\t%s.InterceptMethod((*%s.%s).%s, %d, %d)\n", serverName, pkgName, info.RecvName, info.Name, i.When, i.Priority)
	return code
}

type CodeGenCreaters map[string][]CodeGenCreater

func NewErrorFromPosition(position token.Position, err error) *peony.Error {
	var path = position.Filename
	var lines []string
	var rerr error
	if lines, rerr = peony.ReadLines(path); rerr != nil {
		peony.ERROR.Println("read file error:", rerr)
		path = ""
	}
	return &peony.Error{
		Title:       "Compile error",
		Path:        path,
		FileName:    path,
		Line:        position.Line,
		Column:      position.Column,
		Description: err.Error(),
		SourceLines: lines,
	}
}

var MapperStructCodeGen = make(map[string]*MapperCommentCodeGen)

func (c *CodeGenCreaters) processComments(fileSet *token.FileSet, commentGroup *ast.CommentGroup, typ string, sepc CodeGenSpec) ([]CodeGen, error) {

	codeGens := []CodeGen{}
	if commentGroup == nil || len(commentGroup.List) <= 0 {
		return codeGens, nil
	}
	for _, comment := range commentGroup.List {
		if comment.Text[:2] == "//" {
			content := strings.TrimSpace(comment.Text[2:])
			if len(content) == 0 || content[0] != '@' {
				continue
			}
			for _, codeGenCreater := range (*c)[typ] {
				if codeGen, err := codeGenCreater(content, sepc); err == nil {
					codeGens = append(codeGens, codeGen)
				} else {
					if err == NotMatch {
						continue
					}
					position := fileSet.Position(comment.Pos())
					perr := NewErrorFromPosition(position, err)
					return nil, perr
				}
			}
		}
	}
	return codeGens, nil
}

func (c *CodeGenCreaters) ProcessStructComments(fileSet *token.FileSet, commentGroup *ast.CommentGroup, structInfo *StructInfo) error {
	codeGens, err := c.processComments(fileSet, commentGroup, "struct", structInfo)
	if err != nil {
		return err
	}
	for _, codeGen := range codeGens {
		if codeGen, ok := codeGen.(*MapperCommentCodeGen); ok {
			MapperStructCodeGen[fmt.Sprintf("%s.%s", structInfo.ImportPath, structInfo.Name)] = codeGen
		}
	}
	return nil
}

func hasMapperCodeGen(codeGens []CodeGen) bool {
	for _, codeGen := range codeGens {
		if _, ok := codeGen.(*MapperCommentCodeGen); ok {
			return true
		}
	}
	return false
}

func (c *CodeGenCreaters) ProcessActionComments(fileSet *token.FileSet, commentGroup *ast.CommentGroup, actionInfo *ActionInfo, codeGens *[]CodeGen) error {
	commentCodeGens, err := c.processComments(fileSet, commentGroup, "action", actionInfo)
	if err != nil {
		return err
	}
	//support for mapper in struct
	if len(actionInfo.RecvName) != 0 && (!hasMapperCodeGen(commentCodeGens) || len(commentCodeGens) == 0) {
		if mapperCodeGen, ok := MapperStructCodeGen[fmt.Sprintf("%s.%s", actionInfo.ImportPath, actionInfo.RecvName)]; ok {
			codeGen := &MapperCommentCodeGen{}
			*codeGen = *mapperCodeGen
			codeGen.ActionInfo = actionInfo
			if !strings.HasSuffix(codeGen.UrlExpr, "/") {
				codeGen.UrlExpr += "/"
			}
			codeGen.UrlExpr += strings.ToLower(actionInfo.Name)
			commentCodeGens = append(commentCodeGens, codeGen)
		}
	}
	*codeGens = append(*codeGens, commentCodeGens...)
	return nil
}

var (
	codeGenCreaters = CodeGenCreaters{}
)

func (c *CodeGenCreaters) RegisterCodeGenCreater(name string, builder CodeGenCreater) {
	(*c)[name] = append((*c)[name], builder)
}

func init() {
	codeGenCreaters.RegisterCodeGenCreater("action", MapperCommentCodeGenCreater)
	codeGenCreaters.RegisterCodeGenCreater("struct", MapperCommentCodeGenCreater)
	codeGenCreaters.RegisterCodeGenCreater("action", InterceptCommentCodeGenCreater)
}

var (
	NotMatch               = errors.New("the comment not match")
	UrlArgumentRequired    = errors.New("url argument is required")
	ArgMustbeString        = errors.New("Arg must be string")
	ArgMustbeBool          = errors.New("Arg must be bool")
	WhenArgMustbeString    = errors.New(`arg 'when' must be string, e.g. "BEFORE", "AFTER", "FINALLY","PANIC"`)
	MethodsArgMustbeString = errors.New("methods arg must be string array")

	PriorityArgMustbeInt = errors.New("prioprity arg must be int")

	ArgMustbeArray        = errors.New("Arg must be array")
	UnknownArgument       = errors.New("unknown argument")
	UnknownMethodArgument = errors.New("unknown method, method should be POST, GET, PUT, DELETE")
	NotSupportFunc        = errors.New("Intecept must used for method, not support func. method e.g. func (*Struct) Method{...}")
	UnkownArguemnt        = errors.New("unknown argument")
)

//create the mapper for comment generator.
//e.g. @Mapper @Mapper("/index") @Mapper(url="/index")
//@Mapper(url="/index", methods=["*"]) @Mapper(url="/index", methods=["POST","GET", "PUT"])
func MapperCommentCodeGenCreater(comment string, spec CodeGenSpec) (CodeGen, error) {
	if !strings.HasPrefix(comment, "@Mapper") {
		return nil, NotMatch
	}
	lexer := &CommentLexer{}
	cfun, err := lexer.Parse(comment)
	if err != nil {
		return nil, err
	}
	url := ""
	methods := peony.HttpMethods
	ignore := false
	hasUrlArg := false
	if len(cfun.Args) > 0 {
		for idx, arg := range cfun.Args {
			switch {
			case (idx == 0 && arg.Name == "") || arg.Name == "url":
				//set url argument
				if arg.Value.ValueType() != CommentStringType {
					return nil, ArgMustbeString
				}
				url = string(*arg.Value.(*CommentStringValue))
				hasUrlArg = true
			case arg.Name == "methods":
				//set methods argument
				if arg.Value.ValueType() != CommentArrayType {
					return nil, ArgMustbeArray
				}
				array := *arg.Value.(*CommentArrayValue)
				methods = []string{}
				for _, value := range array {
					if value.ValueType() != CommentStringType {
						return nil, MethodsArgMustbeString
					}
					meth := string(*value.(*CommentStringValue))
					if meth == "*" {
						methods = peony.HttpMethods
						break
					}
					//is the httpmethods suppport.
					if !peony.StringSliceContain(peony.ExtendHttpMethods, meth) {
						return nil, UnknownMethodArgument
					}
					//append method
					if !peony.StringSliceContain(methods, meth) {
						methods = append(methods, meth)
					}
				}
			case arg.Name == "ignore":
				if arg.Value.ValueType() != CommentBoolType {
					return nil, ArgMustbeBool
				}
				ignore = bool(*arg.Value.(*CommentBoolValue))
			default:
				return nil, UnknownArgument
			}
		}
	}
	if structInfo, ok := spec.(*StructInfo); ok {
		//for struct code generate. use recive name for the prefix
		if !hasUrlArg {
			url = "/" + strings.ToLower(structInfo.Name)
		}
		return &MapperCommentCodeGen{ignore, nil, url, methods}, nil
	}
	if actionInfo, ok := spec.(*ActionInfo); ok {
		if !hasUrlArg {
			//use default rule
			url = "/" + peony.ParseAction(actionInfo.ActionName)
		}
		if !ignore && url == "" {
			return nil, UrlArgumentRequired
		}
		return &MapperCommentCodeGen{ignore, actionInfo, url, methods}, nil
	}
	return nil, NotMatch
}

//create the intercept for comment generator.
//Intercept code generator.
//e.g. @Intercept("BEFORE") @Intercept(when="BEFORE")
//@Intercept("BEFORE", priority=1) @Intercept(when="BEFORE", priority=1)
func InterceptCommentCodeGenCreater(comment string, spec CodeGenSpec) (CodeGen, error) {
	if actionInfo, ok := spec.(*ActionInfo); ok {

		//The synax analyze is complex, so I do string compare first.
		if !strings.HasPrefix(comment, "@Intercept") {
			return nil, NotMatch
		}

		if actionInfo.RecvName == "" {
			//it's func, now we don't support
			return nil, NotSupportFunc
		}
		lexer := &CommentLexer{}
		cfun, err := lexer.Parse(comment)
		priority := 0
		if err != nil {
			return nil, err
		}

		when := 0
		for idx, arg := range cfun.Args {
			if (idx == 0 && arg.Name == "") || arg.Name == "when" {
				if arg.Value.ValueType() != CommentStringType {
					return nil, WhenArgMustbeString
				}
				whenString := string(*arg.Value.(*CommentStringValue))
				switch whenString {
				case "BEFORE":
					when = peony.BEFORE
				case "AFTER":
					when = peony.AFTER
				case "FINALLY":
					when = peony.FINALLY
				case "PANIC":
					when = peony.PANIC
				default:
					return nil, UnkownArguemnt
				}
			} else if arg.Name == "priority" {
				if arg.Value.ValueType() != CommentIntType {
					return nil, WhenArgMustbeString
				}
				priority = int(*arg.Value.(*CommentIntValue))
			}
		}
		return &InterceptCommentCodeGen{actionInfo, when, priority}, nil
	}
	return nil, NotMatch
}

type SourceInfo struct {
	Pkgs []*PkgInfo
}

type PkgInfo struct {
	Name        string
	ImportPath  string
	MapperTypes []string
	CodeGens    []CodeGen
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

type StructInfo struct {
	CodeGenSpec
	Name       string
	PkgName    string
	ImportPath string
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

func processAction(fileSet *token.FileSet, pkgInfo *PkgInfo, imports map[string]string, funcDecl *ast.FuncDecl) error {

	if !funcDecl.Name.IsExported() {
		return nil
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
			return nil
		}
		for _, name := range param.Names {
			actionInfo.Args = append(actionInfo.Args, &MethodArg{
				name.Name,
				importPath,
				typeExpr,
			})
		}
	}
	actionInfo.RecvName, actionInfo.ActionName = actionName(funcDecl)
	if err := codeGenCreaters.ProcessActionComments(fileSet, funcDecl.Doc, actionInfo, &pkgInfo.CodeGens); err != nil {
		return err
	}
	return nil
}

func processTypes(fileSet *token.FileSet, file *ast.File, pkgInfo *PkgInfo) error {
	for _, decl := range file.Decls {
		switch specDecl := decl.(type) {
		case *ast.GenDecl:
			if specDecl.Tok == token.TYPE && len(specDecl.Specs) == 1 {
				typeSpec := specDecl.Specs[0].(*ast.TypeSpec)
				if _, ok := typeSpec.Type.(*ast.StructType); ok {
					structInfo := &StructInfo{Name: typeSpec.Name.Name, PkgName: pkgInfo.Name, ImportPath: pkgInfo.ImportPath}
					err := codeGenCreaters.ProcessStructComments(fileSet, specDecl.Doc, structInfo)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func processFile(fileSet *token.FileSet, file *ast.File, pkgInfo *PkgInfo) error {
	imports := map[string]string{}
	processImports(imports, file.Imports)
	if err := processTypes(fileSet, file, pkgInfo); err != nil {
		return err
	}
	for _, decl := range file.Decls {
		switch specDecl := decl.(type) {
		case *ast.FuncDecl:
			if err := processAction(fileSet, pkgInfo, imports, specDecl); err != nil {
				return err
			}
		}
	}
	return nil
}

func processPackage(si *SourceInfo, importPath string, pkg *ast.Package, fileSet *token.FileSet) error {
	pkgInfo := &PkgInfo{ImportPath: importPath, Name: pkg.Name}
	if pkg.Name == "controllers" {
		for _, file := range pkg.Files {
			if err := processFile(fileSet, file, pkgInfo); err != nil {
				return err
			}
		}
	}
	si.Pkgs = append(si.Pkgs, pkgInfo)
	return nil
}

func NewSourceInfo() *SourceInfo {
	s := &SourceInfo{}
	return s
}

//analzye the import path
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
					for _, err := range errList {
						var hasSource = false
						var source []string
						if source, hasSource = fileSources[err.Pos.Filename]; !hasSource {
							source = peony.MustReadLines(err.Pos.Filename)
							fileSources[err.Pos.Filename] = source
						}

						return &peony.Error{
							Title:       "Compile error",
							FileName:    err.Pos.Filename,
							Path:        err.Pos.Filename,
							Description: err.Msg,
							Line:        err.Pos.Line,
							Column:      err.Pos.Column,
							SourceLines: source,
						}
					}
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
				if err := processPackage(si, importPath, pkg, fileSet); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

	}
	return si, nil
}
