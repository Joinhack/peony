package mole

import (
	"github.com/joinhack/peony"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type Agent struct {
	app      *peony.App
	AppAddr  string
	notifier *peony.Notifier
	proxy    *httputil.ReverseProxy
}

func (a *Agent) Path() string {
	return a.app.AppPath
}

func (a *Agent) Apply() {
	println(111)
}

func (a *Agent) IgnoreDir(info os.FileInfo) bool {
	switch info.Name() {
	case "tmp", "view":
		return true
	default:
		return false
	}
}

func (a *Agent) IgnoreFile(info os.FileInfo) bool {
	if strings.HasSuffix(info.Name(), ".go") {
		return false
	}
	return true
}

func NewAgent(app *peony.App, appAddr string) (agent *Agent, err error) {
	agent = &Agent{app: app}
	targetSvrUrl := &url.URL{Scheme: "http", Host: appAddr}
	agent.proxy = httputil.NewSingleHostReverseProxy(targetSvrUrl)
	agent.notifier = peony.NewNotifier()
	agent.notifier.Watch(agent)
	return
}

func processError(err error, w http.ResponseWriter) {
	w.Write([]byte(err.Error()))
}

func (a *Agent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := a.notifier.Notify()
	if err != nil {
		processError(err, w)
		return
	}
	a.proxy.ServeHTTP(w, r)
}

func (a *Agent) Run(addr string) {
	http.ListenAndServe(addr, a)
}
