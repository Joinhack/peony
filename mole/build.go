package mole

import (
	"fmt"
	"github.com/joinhack/peony"
	"go/format"
	"log"
	"os"
	"path"
	"text/template"
)

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

//get importPath alias key is importPath, value is alias
func getAlais(si *SourceInfo) map[string]string {
	alias := map[string][]string{}
	rs := map[string]string{}
	for _, pkg := range si.Pkgs {
		if !contains(alias[pkg.Name], pkg.ImportPath) {
			alias[pkg.Name] = append(alias[pkg.Name], pkg.ImportPath)
		}
		for _, action := range pkg.Actions {
			for _, arg := range action.Args {
				if !contains(alias[arg.Expr.PkgName], arg.ImportPath) {
					alias[arg.Expr.PkgName] = append(alias[arg.Expr.PkgName], arg.ImportPath)
				}
			}
		}
	}

	for aliasName, importPaths := range alias {
		for idx, importPath := range importPaths {
			name := fmt.Sprintf("%s%d", aliasName, idx)
			rs[importPath] = name
		}
	}

	return rs
}

func Build(app *peony.App) error {
	si, err := ProcessSources(app.CodePaths)
	if err != nil {
		return err
	}
	actions := []*ActionInfo{}

	for _, pkg := range si.Pkgs {
		actions = append(actions, pkg.Actions...)
	}

	args := map[string]interface{}{
		"importPaths": getAlais(si),
		"actions":     actions,
	}
	genSource(path.Join(app.AppPath, "tmp"), "main.go", MAIN, args)
	return nil
}

func genSource(dir, filename, tpl string, args map[string]interface{}) {
	code := peony.ExecuteTemplate(template.Must(template.New("").Parse(tpl)), args)
	finfo, err := os.Stat(dir)
	if err != nil && !os.IsExist(err) {
		err = os.Mkdir(dir, 0777)
		if err != nil {
			log.Fatalln("create dir error:", dir)
			return
		}
	}
	if !finfo.IsDir() {
		log.Fatalln("Not dir, shoul be a dir.")
		return
	}
	filepath := path.Join(dir, filename)
	os.Remove(filepath)
	var file *os.File
	file, err = os.Create(filepath)

	if err != nil {
		log.Fatalln("Open file error:", err)
		return
	}
	defer file.Close()
	var codeBytes []byte
	codeBytes, err = format.Source([]byte(code))
	if err != nil {
		log.Fatalln("Source format eror:", err, code)
		return
	}
	code = string(codeBytes)
	_, err = file.WriteString(code)
	if err != nil {
		log.Fatalln("Write source eror:", err)
		return
	}

}

var MAIN = `
package main
import (
	"github.com/joinhack/peony"
	"flag"{{range $path, $alias := $.importPaths }}
	{{$alias}} "{{$path}}"{{end}}
)

var (
	runMode    *string = flag.String("runMode", "", "Run mode.")
	bindAddr   *string = flag.String("port", ":8080", "By default, read from app.conf")
	importPath *string = flag.String("importPath", "", "Go Import Path for the app.")
	srcPath    *string = flag.String("srcPath", "", "Path to the source root.")
)

func main() {
	flag.Parse()
	app := peony.NewApp(*srcPath, *importPath)
	app.BindAddr = *bindAddr
	svr := app.NewServer()
	{{range $idx, $action := $.actions }}
	svr.Mapper("/", {{index $.importPaths .ImportPath}}{{if .RecvName}}.{{.RecvName}}{{end}}.{{.Name}})
	{{end}}
	svr.Run()
}
`
