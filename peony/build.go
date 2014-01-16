package main

import (
	"github.com/joinhack/peony"
	"github.com/joinhack/peony/mole"
)

var buildcmd = &Command{
	Name:    "build",
	Execute: buildCommand,
	Desc:    `build ImportPath(peony project)`,
}

func buildCommand(args []string) {
	if len(args) == 0 {
		usage(1)
	}
	importPath := args[0]
	srcRoot := peony.SearchSrcRoot(importPath)

	app := peony.NewApp(srcRoot, importPath)
	app.DevMode = true
	if err := mole.Build(app); err != nil {
		eprintf("build project error, %s\n", err.Error())
	}
}
