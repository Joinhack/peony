package main

import (
	"github.com/joinhack/peony"
	"github.com/joinhack/peony/mole"
)

var runcmd = &Command{
	Name:    "run",
	Execute: run,
	Desc:    `run importpath(peony project) [addr](default :8000)`,
}

func run(args []string) {
	if len(args) == 0 {
		usage(1)
	}
	importPath := args[0]
	srcRoot := peony.SearchSrcRoot(importPath)

	app := peony.NewApp(srcRoot, importPath)
	app.DevMode = true
	ag, _ := mole.NewAgent(app)
	var addr = ":8000"
	if len(args) > 1 {
		addr = args[1]
	}
	ag.Run(addr)
}
