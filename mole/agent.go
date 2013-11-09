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
	app          *peony.App
	appCmd       *AppCmd
	AppAddr      string
	notifier     *peony.Notifier
	forceRefresh bool
	proxy        *httputil.ReverseProxy
}

func (a *Agent) Path() string {
	return a.app.AppPath
}

func (a *Agent) ForceRefresh() bool {
	return a.forceRefresh
}

func (a *Agent) Refresh() error {
	a.forceRefresh = false
	a.appCmd.Kill()
	if err := Build(a.app); err != nil {
		return err
	}
	return a.appCmd.Start()
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
	agent.forceRefresh = true
	var binPath string
	binPath, err = GetBinPath(app)
	if err != nil {
		return
	}
	agent.appCmd = NewAppCmd(app, binPath, appAddr)
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
