package peony

import (
	"log"
	"os"
)

type App struct {
	Sources []string
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

func NewApp() *App {
	app := &App{}
	return app
}

func (app *App) Bind(url string) {

}
