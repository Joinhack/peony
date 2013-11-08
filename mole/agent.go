package mole

import (
	"github.com/joinhack/peony"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

type Agent struct {
	app     *peony.App
	AppAddr string
	proxy   *httputil.ReverseProxy
}

func NewAgent(app *peony.App, appAddr string) (agent *Agent, err error) {
	agent = &Agent{app: app}
	targetSvrUrl := &url.URL{Scheme: "http", Host: appAddr}
	agent.proxy = httputil.NewSingleHostReverseProxy(targetSvrUrl)
	return
}

func (a *Agent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.proxy.ServeHTTP(w, r)
}

func (a *Agent) Run(addr string) {
	http.ListenAndServe(addr, a)
}
