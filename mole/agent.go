package mole

import (
	"fmt"
	"github.com/joinhack/peony"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"
)

type Agent struct {
	app            *peony.App
	appCmd         *AppCmd
	appBinPath     string
	AppAddr        string
	templateLoader *peony.TemplateLoader
	notifier       *peony.Notifier
	forceRefresh   bool
	proxy          *httputil.ReverseProxy
}

func (a *Agent) Path() []string {
	return []string{a.app.AppPath}
}

func (a *Agent) ForceRefresh() bool {
	return a.forceRefresh
}

func (a *Agent) Refresh() error {
	if a.appCmd != nil {
		a.appCmd.Kill()
	}
	a.appCmd = NewAppCmd(a.app, a.appBinPath, a.AppAddr)
	if err := Build(a.app); err != nil {
		return err
	}
	err := a.appCmd.Start()
	if err == nil {
		a.forceRefresh = false
	}
	return err
}

func (a *Agent) IgnoreDir(info os.FileInfo) bool {
	switch info.Name() {
	case "tmp", "views":
		return true
	default:
		return false
	}
}

func (a *Agent) IgnoreFile(f string) bool {
	if strings.HasSuffix(f, ".go") {
		return false
	}
	return true
}

func getListenAddr() string {
	ipcon, err := net.Listen("tcp", ":0")
	if err != nil {
		peony.ERROR.Fatalln("getListenPort error:", err)
	}
	ipcon.Close()
	p := ipcon.Addr().(*net.TCPAddr).Port
	return fmt.Sprintf(":%d", p)
}

func NewAgent(app *peony.App) (agent *Agent, err error) {
	agent = &Agent{app: app}
	appAddr := getListenAddr()
	targetSvrUrl := &url.URL{Scheme: "http", Host: appAddr}
	agent.proxy = httputil.NewSingleHostReverseProxy(targetSvrUrl)
	agent.notifier = peony.NewNotifier()
	agent.forceRefresh = true
	var binPath string
	binPath, err = GetBinPath(app)
	if err != nil {
		return
	}
	agent.AppAddr = appAddr
	agent.appBinPath = binPath
	//watch template. template should first watched by notifier
	agent.templateLoader = peony.NewTemplateLoader([]string{
		path.Join(path.Join(peony.PEONYPATH, "views")),
	})
	agent.templateLoader.FuncMap["IsDevMode"] = func() bool { return true }
	if err := agent.templateLoader.Refresh(); err != nil {
		panic(err)
	}
	agent.notifier.Watch(agent)
	return
}

func (a *Agent) processError(err error, w http.ResponseWriter, r *http.Request) {
	c := peony.NewController(w, r, a.templateLoader)
	peony.RenderError(err).Apply(c)
}

func (a *Agent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := a.notifier.Notify()
	if err != nil {
		a.processError(err, w, r)
		return
	}
	switch r.Header.Get("Upgrade") {
	case "websocket", "Websocket":
		proxyWebsocket(w, r, a.AppAddr)
	default:
		a.proxy.ServeHTTP(w, r)
	}
}

func (a *Agent) Run(addr string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Kill)
	go func() {
		<-sigChan
		if a.appCmd != nil {
			a.appCmd.Kill()
		}
	}()
	go func() {
		time.Sleep(1 * time.Second)
		peony.INFO.Println("peony agent is listening on", addr)
	}()
	err := http.ListenAndServe(addr, a)
	if err != nil {
		peony.ERROR.Fatalln(err.Error())
	}
}

func proxyWebsocket(w http.ResponseWriter, r *http.Request, host string) {
	d, err := net.Dial("tcp", host)
	if err != nil {
		http.Error(w, "Error contacting backend server.", 500)
		peony.ERROR.Printf("Error dialing websocket backend %s: %v", host, err)
		return
	}
	defer d.Close()
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Not a hijacker?", 500)
		return
	}
	nc, _, err := hj.Hijack()
	if err != nil {
		peony.ERROR.Printf("Hijack error: %v", err)
		return
	}
	defer nc.Close()

	err = r.Write(d)
	if err != nil {
		peony.ERROR.Printf("Error copying request to target: %v", err)
		return
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}
	go cp(d, nc)
	go cp(nc, d)
	<-errc
}
