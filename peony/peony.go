package main

import (
	"flag"
	"os"
	"text/template"
)

var tmpl = `Example Usage:
peony run [ImportPath]
peony build [ImportPath]
`

type Command struct {
	Execute func([]string)
	Name    string
}

var Commands = map[string]*Command{}

func registCommand(c *Command) {
	Commands[c.Name] = c
}

func init() {
	registCommand(runcmd)
}

func usage(c int) {
	var tpl = template.Must(template.New("").Parse(tmpl))
	tpl.Execute(os.Stderr, nil)
	os.Exit(c)
}

func main() {
	flag.Parse()
	flag.Usage = func() {
		usage(1)
	}
	args := flag.Args()
	if len(args) < 1 {

		usage(1)
	}
	if command, ok := Commands[args[0]]; ok {
		command.Execute(args[1:])
	} else {
		usage(1)
	}

}
