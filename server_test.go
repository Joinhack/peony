package peony

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func Index() string {
	return "hi"
}

func Json() Renderer {
	return RenderJson("hi")
}

func Template() Renderer {
	return RenderTemplate(nil)
}

func Xml() Renderer {
	return RenderXml("hi")
}

func Text(join string) string {
	return join
}

func File(path string) Renderer {
	return RenderFile(path)
}

type AS int

func (a *AS) T() Renderer {
	*a = 10
	return RenderText("s")
}

func TestServer(t *testing.T) {
	var err error
	app := NewApp(".", ".")
	app.ViewPath, err = filepath.Abs(".")
	app.BindAddr = GetRandomListenAddr()
	app.DevMode = true
	svr := app.NewServer()
	svr.Init()
	err = svr.MethodMapper("/", HttpMethods, (*AS).T, &Action{Name: "AS.T"})
	if err != nil {
		panic(err)
	}
	svr.FuncMapper("/json", HttpMethods, Json, &Action{Name: "xssxeem"})
	svr.FuncMapper("/template", HttpMethods, Template, &Action{Name: "recover.go"})
	svr.FuncMapper("/xml", HttpMethods, Xml, &Action{Name: "xml"})
	svr.FuncMapper("/<int:join>", HttpMethods, Text, &Action{Name: "xxeemw", Args: []*ArgType{&ArgType{Name: "join", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	svr.FuncMapper("/static/<string:path>", HttpMethods, File, &Action{Name: "file", Args: []*ArgType{&ArgType{Name: "path", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	err = svr.Listen()
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan int, 1)
	go func() {
		select {
		case <-svr.Run():
		case <-done:
		case <-time.After(30 * time.Second):
		}
		svr.CloseListener()
	}()
	res, _ := http.Get("http://127.0.0.1" + app.BindAddr + "/json")
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	t.Log(string(bs))
	done <- 1
	return
}
