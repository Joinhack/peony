package main

import (
	"fmt"
	"github.com/joinhack/peony"
	"go/build"
	"os"
	"path/filepath"
)

var newShortDesc = `new ImportPath(peony project)`
var newcmd = &Command{
	Name:    "new",
	Execute: newapp,
	Desc:    newShortDesc,
}

func newapp(args []string) {
	if len(args) == 0 {
		eprintf(newShortDesc)
	}
	filepath.Join(peony.PEONYPATH, "templates")
	importPath := args[0]

	gopath := build.Default.GOPATH

	if gopath == "" {
		eprintf("please set the GOPATH\n", importPath)
	}

	if filepath.IsAbs(importPath) {
		eprintf("importPath[%s] looks like the file path.\n", importPath)
	}

	if _, err := build.Import(importPath, "", build.FindOnly); err == nil {
		eprintf("importPath[%s] already exist.\n", importPath)
	}

	if _, err := build.Import(peony.PEONY_IMPORTPATH, "", build.FindOnly); err != nil {
		eprintf("peony source is required.\n")
	}

	tmplatesPath := filepath.Join(peony.PEONYPATH, "templates")

	srcPath := filepath.Join(filepath.SplitList(gopath)[0], "src")
	appPath := filepath.Join(srcPath, filepath.FromSlash(importPath))
	if err := os.Mkdir(appPath, 0777); err != nil {
		eprintf("mdir app dir error, %s\n", err.Error())
	}
	if err := copyDir(tmplatesPath, appPath); err != nil {
		eprintf("copy dir error, %s\n", err.Error())
	}
	fmt.Println("app already is ready, please execute command: peony run", importPath)
}
