package peony

import (
	. "github.com/joinhack/goconf"
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
	peonyPath        = ""
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
	AppName       string
	BindAddr   string
	Security   string
	DevMode    bool
	Trunk      bool
	Config     Config
	section    string
	StaticInfo *StaticInfo
}

func GetPeonyPath() string {
	if peonyPath == "" {
		importPath := PEONY_IMPORTPATH
		peonyPath = filepath.Join(SearchSrcRoot(importPath), importPath)
	}
	return peonyPath
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

func NewApp(sourcePath, importPath string) *App {
	app := &App{SourcePath: sourcePath, ImportPath: importPath}
	app.BasePath = filepath.Join(sourcePath, importPath)
	_, app.AppName = filepath.Split(app.BasePath)
	app.AppPath = filepath.Join(app.BasePath, "app")
	app.CodePaths = []string{app.AppPath}
	app.ViewPath = filepath.Join(app.AppPath, "views")
	return app
}

func (a *App) getLogger(name, defaultout string) *log.Logger {
	prefix := a.GetStringConfig("log."+name+".prefix", strings.ToUpper(name)+" ")
	out := a.GetStringConfig("log."+name+".out", defaultout)
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

func (a *App) GetStringConfig(key, s string) string {
	return a.Config.StringDefault(a.section, key, s)
}

func (a *App) GetFloatConfig(key string, f float64) float64 {
	return a.Config.FloatDefault(a.section, key, f)
}

func (a *App) GetBoolConfig(key string, b bool) bool {
	return a.Config.BoolDefault(a.section, key, b)
}

func (a *App) GetIntConfig(key string, f int64) int64 {
	return a.Config.IntDefault(a.section, key, f)
}

func (a *App) LoadConfig() {
	var ok = false
	if a.DevMode {
		a.section = "dev"
	} else {
		a.section = "prod"
	}
	a.Config = Config{}
	a.Config.ReadFile(filepath.Join(a.BasePath, "conf", "app.cnf"))
	a.AppName = a.GetStringConfig("app.name", a.AppName)

	TRACE = a.getLogger("trace", a.GetStringConfig("log.trace.output", "nil"))
	WARN = a.getLogger("warn", a.GetStringConfig("log.warn.output", "stdout"))
	INFO = a.getLogger("info", a.GetStringConfig("log.info.output", "stdout"))
	ERROR = a.getLogger("error", a.GetStringConfig("log.error.output", "stderr"))
	a.Security = a.GetStringConfig("app.secret", defaultSecKey)
	a.BindAddr = a.GetStringConfig("app.addr", ":8000")
	a.Trunk = a.GetBoolConfig("http.trunk", true)
	staticInfo := &StaticInfo{}
	ok = false
	staticInfo.UriPrefix, ok = a.Config.String(a.section, "static.uri")
	if ok {
		staticInfo.Path, ok = a.Config.String(a.section, "static.path")
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
