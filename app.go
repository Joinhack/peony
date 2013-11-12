package peony

import (
	"log"
	"os"
	"path"
)

type App struct {
	SourcePath string
	ImportPath string
	CodePaths  []string
	ViewPath   string
	AppPath    string
	DevMode    bool
	BindAddr   string
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
	WARN = log.New(os.Stdout, "WARN ", log.Ldate|log.Ltime)
	INFO = log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	ERROR = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime)
}

func NewApp(sourcePath, importPath string) *App {
	app := &App{SourcePath: sourcePath, ImportPath: importPath}
	app.AppPath = path.Join(sourcePath, importPath, "app")
	app.CodePaths = []string{app.AppPath}
	app.ViewPath = path.Join(app.AppPath, "views")
	return app
}

func (a *App) NewServer() *Server {
	svr := NewServer(a)
	return svr
}
