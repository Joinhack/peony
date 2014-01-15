package peony

import (
	"go/build"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
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

type StaticInfo struct {
	UriPrefix, Path string
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
	Config     Config
	StaticInfo *StaticInfo
}

func SearchSrcRoot(imp string) string {
	pkg, err := build.Import(imp, "", build.FindOnly)
	if err != nil {
		ERROR.Fatalln("get  abslute import path  error:", err)
	}
	return pkg.SrcRoot
}

var (
	TRACE *log.Logger = log.New(os.Stdout, "TRACE ", log.Ldate|log.Ltime|log.Lshortfile)
	WARN  *log.Logger = log.New(os.Stdout, "WARN ", log.Ldate|log.Ltime|log.Lshortfile)
	INFO  *log.Logger = log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)
	ERROR *log.Logger = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
)

func init() {
	importPath := PEONY_IMPORTPATH
	PEONYPATH = filepath.Join(SearchSrcRoot(importPath), importPath)
}

func NewApp(sourcePath, importPath string) *App {
	app := &App{SourcePath: sourcePath, ImportPath: importPath}
	app.BasePath = filepath.Join(sourcePath, importPath)
	app.AppPath = filepath.Join(app.BasePath, "app")
	app.CodePaths = []string{app.AppPath}
	app.ViewPath = filepath.Join(app.AppPath, "views")
	app.LoadConfig()
	return app
}

func (a *App) getLogger(model, name, defaultout string) *log.Logger {
	prefix := a.Config.StringDefault(model, "log."+name+".prefix", strings.ToUpper(name)+" ")
	out := a.Config.StringDefault(model, "log."+name+".out", defaultout)
	var writer io.Writer
	switch out {
	case "stderr":
		writer = os.Stderr
	case "stdout":
		writer = os.Stdout

	default:
		if out == "nil" {
			out = os.DevNull
		}
		var err error
		writer, err = os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("open log file %s error, %s", out, err)
		}
	}
	return log.New(writer, prefix, log.Ldate|log.Ltime|log.Lshortfile)
}

func (a *App) LoadConfig() {
	var ok = false
	var section string
	a.Config = Config{}
	a.Config.ReadFile(filepath.Join(a.BasePath, "conf", "app.cnf"))
	if a.DevMode {
		section = "dev"
	} else {
		section = "prod"
	}

	TRACE = a.getLogger(section, "warn", "nil")
	WARN = a.getLogger(section, "warn", "stdout")
	INFO = a.getLogger(section, "info", "stdout")
	ERROR = a.getLogger(section, "error", "stderr")
	a.Security = a.Config.StringDefault(section, "app.secret", defaultSecKey)
	a.BindAddr = a.Config.StringDefault(section, "app.addr", defaultSecKey)
	staticInfo := &StaticInfo{}
	ok = false
	staticInfo.UriPrefix, ok = a.Config.String(section, "static.uri")
	if ok {
		staticInfo.Path, ok = a.Config.String(section, "static.path")
		if !path.IsAbs(staticInfo.Path) {
			staticInfo.Path = path.Join(a.BasePath, staticInfo.Path)
		}
	}
	if ok {
		a.StaticInfo = staticInfo
	}
}

func (a *App) NewServer() *Server {
	svr := NewServer(a)
	return svr
}
