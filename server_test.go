package peony

import (
	"path/filepath"
	"reflect"
	"testing"
)

func Index() string {
	return "hi"
}

func Json() Render {
	return NewJsonRender("hi")
}

func Template() Render {
	return NewTemplateRender(nil)
}

func Xml() Render {
	return NewXmlRender("hi")
}

func Text(join string) string {
	return join
}

func File(path string) Render {
	return NewFileRender(path)
}

type AS int

func (a *AS) T() Render {
	*a = 10
	return NewTextRender("s")
}

func TestServer(t *testing.T) {
	var err error
	app := NewApp(".", ".")
	app.ViewPath, err = filepath.Abs(".")
	app.BindAddr = ":8080"
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
	err = svr.Run()
	if err != nil {
		t.Fatal(err)
	}
}
