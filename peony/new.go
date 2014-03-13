package main

import (
	"fmt"
	"github.com/joinhack/peony"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

var newShortDesc = `new importpath(peony project), create a peony project`
var newcmd = &Command{
	Name:    "new",
	Execute: newapp,
	Desc:    newShortDesc,
}

func genConfig(appPath string, data map[string]string) error {
	var src, dest *os.File
	var fileContent []byte
	var err error
	var tmpl *template.Template
	name := filepath.Join(appPath, "conf", "app.cnf.template")

	if src, err = os.Open(name); err != nil {
		return err
	}
	tmpfile := name
	defer func() { src.Close(); os.Remove(tmpfile) }()
	name = filepath.Join(appPath, "conf", "app.cnf")
	if dest, err = os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
		return nil
	}
	defer dest.Close()
	if fileContent, err = ioutil.ReadAll(src); err != nil {
		return err
	}
	if tmpl, err = template.New("").Parse(string(fileContent)); err != nil {
		return err
	}
	return tmpl.Execute(dest, data)
}

func newapp(args []string) {
	if len(args) == 0 {
		eprintf(newShortDesc)
	}
	filepath.Join(peony.GetPeonyPath(), "templates")
	importPath := args[0]

	gopath := build.Default.GOPATH

	if gopath == "" {
		eprintf("please set the GOPATH\n", importPath)
	}

	if filepath.IsAbs(importPath) {
		eprintf("importpath[%s] looks like the file path.\n", importPath)
	}

	if _, err := build.Import(importPath, "", build.FindOnly); err == nil {
		eprintf("importpath[%s] already exist.\n", importPath)
	}

	if _, err := build.Import(peony.PEONY_IMPORTPATH, "", build.FindOnly); err != nil {
		eprintf("peony source is required.\n")
	}

	tmplatesPath := filepath.Join(peony.GetPeonyPath(), "templates")
	errorsPath := filepath.Join(peony.GetPeonyPath(), "views", "errors")

	srcPath := filepath.Join(filepath.SplitList(gopath)[0], "src")
	appPath := filepath.Join(srcPath, filepath.FromSlash(importPath))
	if err := os.Mkdir(appPath, 0777); err != nil {
		eprintf("mdir app dir error, %s\n", err.Error())
	}
	if err := copyDir(tmplatesPath, appPath); err != nil {
		eprintf("copy dir error, %s\n", err.Error())
	}
	if err := copyDir(errorsPath, filepath.Join(appPath, "app", "views", "errors")); err != nil {
		eprintf("copy dir error, %s\n", err.Error())
	}
	appName := filepath.Base(filepath.FromSlash(importPath))
	param := map[string]string{
		"AppName":    appName,
		"ImportPath": importPath,
		"SecKey":     peony.GenSecKey(),
	}
	if err := genConfig(appPath, param); err != nil {
		eprintf("generator configure error, %s\n", err.Error())
	}
	fmt.Println("app already is ready, please execute command: peony run", importPath)
}
