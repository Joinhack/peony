package mole

import (
	"fmt"
	"github.com/joinhack/peony"
	"log"
	"os"
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
		for _, method := range pkg.Actions.Methods {
			for _, arg := range method.Args {
				if !contains(alias[arg.Expr.PkgName], arg.ImportPath) {
					alias[arg.Expr.PkgName] = append(alias[arg.Expr.PkgName], arg.ImportPath)
				}
			}
		}
	}

	for aliasName, importPaths := range alias {
		for idx, importPath := range importPaths {
			name := aliasName
			if idx > 0 {
				name = fmt.Sprintf("%s%d", aliasName, idx)
			}
			rs[importPath] = name
		}
	}

	return rs
}

func Build(app *peony.App) error {
	// si, err := ProcessSources(app.Sources)
	// if err != nil {
	// 	return err
	// }
	// //alias := getAlais(si)

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

	var file *os.File
	file, err = os.Open(dir + string(os.PathSeparator) + filename)
	if err != nil {
		log.Fatalln("Open file error:", err)
		return
	}
	defer file.Close()
	_, err = file.WriteString(code)
	if err != nil {
		log.Fatalln("write source eror:", err)
		return
	}

}

var MAIN = `
package main
import (
	"github.com/joinhack/peony"

)

func main() {
}
`
