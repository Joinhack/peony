package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/joinhack/peony"
	"github.com/joinhack/peony/mole"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var packcmd = &Command{
	Name:    "pack",
	Execute: pack,
	Desc:    `pack importpath(peony project), the output dir is .`,
}

func panicOnError(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s %s", msg, err.Error()))
	}
}

func targzWrite(tarWriter *tar.Writer, h *tar.Header, r io.Reader) {
	err := tarWriter.WriteHeader(h)
	panicOnError(err, "Failed to write tar entry header")

	_, err = io.Copy(tarWriter, r)
	panicOnError(err, "Failed to copy")
}

var shell string = `#! /usr/bin/env bash
SCRIPTPATH=$(dirname "$0")
"$SCRIPTPATH/{{.BinName}}" -srcPath ".."
`

var cmd string = `@echo off
{{.BinName}} -srcPath ".."
`

func tarapp(srcDir, appName string, tarWriter *tar.Writer) {
	filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		srcFile, err := os.Open(srcPath)
		panicOnError(err, "Failed to read source file")
		defer srcFile.Close()

		header := &tar.Header{
			Name:    filepath.Join(appName, strings.TrimLeft(srcPath[len(srcDir):], string(os.PathSeparator))),
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}
		targzWrite(tarWriter, header, srcFile)

		return nil
	})
}

func tarbin(binPath, binName, appName string, tarWriter *tar.Writer) {
	var info os.FileInfo
	var file *os.File
	var tarh *tar.Header
	var err error
	file, err = os.Open(binPath)
	panicOnError(err, "open binary file error")
	defer file.Close()
	info, err = file.Stat()
	panicOnError(err, "stat error")

	tarh = &tar.Header{
		Name:    filepath.Join(appName, "bin", binName),
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}
	targzWrite(tarWriter, tarh, file)
}

func tryTarErrors(srcDir, appName string, tarWriter *tar.Writer) {
	errorsDir := filepath.Join(srcDir, "app", "views", "errors")
	_, err := os.Stat(errorsDir)
	if err == nil {
		return
	}
	if err != nil && !os.IsNotExist(err) {
		return
	}
	path := peony.GetPeonyPath()
	errsPath := filepath.Join(path, "views", "errors")
	filepath.Walk(errsPath, func(srcPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		srcFile, err := os.Open(srcPath)
		panicOnError(err, "Failed to read source file")
		defer srcFile.Close()

		header := &tar.Header{
			Name:    filepath.Join(appName, "app", strings.TrimLeft(srcPath[len(path):], string(os.PathSeparator))),
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}
		targzWrite(tarWriter, header, srcFile)

		return nil
	})

}

func tarshell(binName, appName string, tarWriter *tar.Writer) {
	now := time.Now()
	args := map[string]string{"BinName": binName}
	var targzWriteScript = func(s, n string) {
		val := peony.ExecuteTemplate(template.Must(template.New("").Parse(s)), args)
		tarh := &tar.Header{
			Name:    filepath.Join(appName, "bin", n),
			Size:    int64(len(val)),
			Mode:    int64(0766),
			ModTime: now,
		}
		targzWrite(tarWriter, tarh, bytes.NewReader([]byte(val)))
	}
	targzWriteScript(shell, "run.sh")
	targzWriteScript(cmd, "run.bat")
}

func targz(app *peony.App) {

	n := filepath.Base(app.ImportPath)
	n = app.GetStringConfig("app.name", n)
	destFilename := filepath.Join(".", fmt.Sprintf("%s.tar.gz", n))
	srcDir := app.BasePath
	os.Remove(destFilename)

	zipFile, err := os.Create(destFilename)
	panicOnError(err, "Failed to create archive")
	defer zipFile.Close()

	gzipWriter := gzip.NewWriter(zipFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	tarapp(srcDir, app.AppName, tarWriter)

	var binPath string
	binPath, err = mole.GetBinPath(app)
	panicOnError(err, "get binary path error")
	_, binName := filepath.Split(binPath)
	tarbin(binPath, binName, app.AppName, tarWriter)

	tarshell(binName, app.AppName, tarWriter)

	tryTarErrors(srcDir, app.AppName, tarWriter)
	return
}

func pack(args []string) {
	if len(args) == 0 {
		usage(1)
	}
	importPath := args[0]
	srcRoot := peony.SearchSrcRoot(importPath)

	app := peony.NewApp(srcRoot, importPath)
	app.DevMode = true

	if err := mole.Build(app); err != nil {
		eprintf("build project error, %s\n", err.Error())
		return
	}
	targz(app)
}
