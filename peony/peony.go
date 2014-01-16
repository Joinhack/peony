package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

var tmpl = `Usage: peony command [arguments]
The commands are:
	{{range .}}{{.Desc }}
	{{end}}
`

type Command struct {
	Execute func([]string)
	Desc    string
	Name    string
}

var Commands = map[string]*Command{}

func registCommand(c *Command) {
	Commands[c.Name] = c
}

func init() {
	registCommand(runcmd)
	registCommand(newcmd)
	registCommand(buildcmd)
}

func eprintf(f string, opts ...interface{}) {
	fmt.Fprintf(os.Stderr, f, opts...)
	os.Exit(1)
}

func usage(c int) {
	var tpl = template.Must(template.New("").Parse(tmpl))
	tpl.Execute(os.Stderr, Commands)
	os.Exit(c)
}

func copyDir(src, dest string) error {

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		suffix := path[len(src):]
		destPath := filepath.Join(dest, suffix)
		if info.IsDir() {
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				if err := os.Mkdir(destPath, info.Mode()); err != nil {
					return err
				}

			}
			return nil
		}
		return copyFile(path, destPath)
	})
}

func copyFile(src, dest string) (err error) {
	var destFile, srcFile *os.File
	destFile, err = os.Create(dest)
	if err != nil {
		return
	}
	defer destFile.Close()
	srcFile, err = os.Open(src)
	if err != nil {
		return
	}
	_, err = io.Copy(destFile, srcFile)
	return
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
