package peony

import (
	"go/build"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

var (
	PEONYPATH        string
	PEONY_IMPORTPATH = "github.com/joinhack/peony"
	// from base64 encodeUrl
	defaultSecKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
)

func GenSecKey() string {
	val := make([]byte, 64)
	copy(val, defaultSecKey)
	for i := 0; i < 64; i++ {
		r := rand.Intn(64)
		tmp := val[r]
		val[r] = val[i]
		val[i] = tmp
	}
	return string(val)
}

type App struct {
	SourcePath string
	ImportPath string
	CodePaths  []string
	ViewPath   string
	AppPath    string
	BasePath   string
	BindAddr   string
	Security   string
	DevMode    bool
}

func SearchSrcRoot(imp string) string {
	pkg, err := build.Import(imp, "", build.FindOnly)
	if err != nil {
		ERROR.Fatalln("get  abslute import path  error:", err)
	}
	return pkg.SrcRoot
}

var (
	WARN  *log.Logger
	INFO  *log.Logger
	ERROR *log.Logger
)

func getLogger(name string) *log.Logger {
	return nil
}

func init() {
	WARN = log.New(os.Stdout, "WARN ", log.Ldate|log.Ltime|log.Lshortfile)
	INFO = log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)
	ERROR = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
	importPath := PEONY_IMPORTPATH
	PEONYPATH = filepath.Join(SearchSrcRoot(importPath), importPath)
}

func NewApp(sourcePath, importPath string) *App {
	app := &App{SourcePath: sourcePath, ImportPath: importPath}
	app.BasePath = filepath.Join(sourcePath, importPath)
	app.AppPath = filepath.Join(app.BasePath, "app")
	app.CodePaths = []string{app.AppPath}
	app.ViewPath = filepath.Join(app.AppPath, "views")
	return app
}

func (a *App) NewServer() *Server {
	svr := NewServer(a)
	return svr
}
