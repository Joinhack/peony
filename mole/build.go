package mole

import (
	"fmt"
	"github.com/joinhack/peony"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
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
		if len(pkg.CodeGens) > 0 && !contains(alias[pkg.Name], pkg.ImportPath) {
			alias[pkg.Name] = append(alias[pkg.Name], pkg.ImportPath)
		}
		for _, codeGen := range pkg.CodeGens {
			codeGen.BuildAlias(alias)
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

//find go execute file path
func findGO() (p string, err error) {
	var fstat os.FileInfo
	gobin := "go"
	if runtime.GOOS == "windows" {
		gobin = "go.exe"
	}
	p = path.Join(build.Default.GOROOT, "bin", gobin)
	fstat, err = os.Stat(p)
	if err == nil {
		if m := fstat.Mode(); !fstat.IsDir() && m&0x0111 != 0 {
			return
		}
	}
	p, err = exec.LookPath("go")
	return
}

var ERRORRegexp = regexp.MustCompile(`(?m)^([^:#]+):(\d+):(\d+)?:? (.*)$`)

func newBuildError(out string) error {
	matchs := ERRORRegexp.FindAllStringSubmatch(out, -1)
	if matchs == nil {
		return &peony.Error{
			Title:       "Complie error",
			Description: string(out),
		}
	}
	fileSources := map[string][]string{}
	for _, match := range matchs {

		filep, _ := filepath.Abs(match[1])
		line, _ := strconv.Atoi(match[2])
		filename := filep[len(filepath.Dir(filep)):]

		column := 0
		if match[3] != "" {
			column, _ = strconv.Atoi(match[3])
		}
		desc := match[4]
		var source []string
		var hasSource = false
		if source, hasSource = fileSources[filep]; !hasSource {
			source = peony.MustReadLines(filep)
			fileSources[filep] = source
		}
		//just get the first error
		return &peony.Error{
			Title:       "Complie error",
			FileName:    filename,
			Path:        filep,
			SourceLines: source,
			Line:        line,
			Column:      column,
			Description: desc,
		}
	}
	return nil
}

func GetBinPath(app *peony.App) (string, error) {
	pkg, err := build.Default.Import(app.ImportPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	binPath := pkg.BinDir
	_, pn := filepath.Split(app.ImportPath)
	if runtime.GOOS == "windows" {
		pn += ".exe"
	}
	return filepath.Join(binPath, "peony."+pn), nil
}

//build app
//Generate main.go by code generator. detail code generated by CodeGen(Code Generator)
//build app execute file by command "go build" to $GOPATH/bin
func Build(app *peony.App) error {
	si, err := ProcessSources(app.CodePaths)
	if err != nil {
		return err
	}
	codeGens := []CodeGen{}
	for _, pkg := range si.Pkgs {
		codeGens = append(codeGens, pkg.CodeGens...)
	}
	alias := getAlais(si)
	//rename peony
	alias["github.com/joinhack/peony"] = "peony"

	args := map[string]interface{}{
		"importPaths": alias,
		"codeGens":    codeGens,
	}
	genSource(path.Join(app.AppPath, "tmp"), "main.go", MAIN, args)
	var gopath string
	if gopath, err = findGO(); err != nil {
		return err
	}
	var binPath string
	binPath, err = GetBinPath(app)
	if err != nil {
		return err
	}
	cmd := exec.Command(gopath, "build", "-o", binPath, path.Join(app.ImportPath, "app", "tmp"))
	if output, err := cmd.CombinedOutput(); err != nil {
		peony.ERROR.Println(string(output))
		return newBuildError(string(output))
	}
	return nil
}

func genSource(dir, filename, tpl string, args map[string]interface{}) {
	code := peony.ExecuteTemplate(template.Must(template.New("").Parse(tpl)), args)
	finfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0777)
			if err != nil {
				log.Fatalln("create dir error:", dir)
				return
			}
		} else {
			log.Fatalln("stat error:", err)
		}
	} else if !finfo.IsDir() {
		log.Fatalln("Not dir, shoul be a dir.")
	}
	filepath := path.Join(dir, filename)
	os.Remove(filepath)
	var file *os.File
	file, err = os.Create(filepath)

	if err != nil {
		log.Fatalln("Open file error:", err)
	}
	defer file.Close()
	_, err = file.WriteString(code)
	if err != nil {
		log.Fatalln("Write source eror:", err)
	}

}

var MAIN = `
package main
import (
	"reflect"
	"time"
	"fmt"
	"flag"{{range $path, $alias := $.importPaths }}
	{{$alias}} "{{$path}}"{{end}}
)

var (
	_ = reflect.Ptr
	bindAddr   *string = flag.String("bindAddr", "", "By default, read from app.conf")
	importPath *string = flag.String("importPath", ".", "Go ImportPath for the app.")
	srcPath    *string = flag.String("srcPath", ".", "Path to the source root.")
	devMode    *bool   = flag.Bool("devMode", false, "Run mode")
)

func main() {
	flag.Parse()
	app := peony.NewApp(*srcPath, *importPath)
	if devMode != nil {
		app.DevMode = *devMode
	}
	app.LoadConfig()
	if *bindAddr != "" {
		app.BindAddr = *bindAddr
	}
	svr := app.NewServer()
	svr.Init()
{{range $idx, $codeGen := $.codeGens }}{{$codeGen.Generate "app" "svr" $.importPaths}}{{end}}

	svr.Router.Refresh()

	go func(){
		time.Sleep(1 * time.Second)
		fmt.Println("Server is running, listening on", app.BindAddr)
	}()
	if err := <- svr.Run(); err != nil {
		panic(err)
	}
}
`
