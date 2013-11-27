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

func text(join string) string {
	return join
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
	svr := app.NewServer()
	svr.Init()
	err = svr.MethodMapper("/", (*AS).T, &Action{Name: "AS.T"})
	if err != nil {
		panic(err)
	}
	svr.FuncMapper("/json", Json, &Action{Name: "xssxeem"})
	svr.FuncMapper("/template", Template, &Action{Name: "recover.go"})
	svr.FuncMapper("/xml", Xml, &Action{Name: "xml"})
	svr.FuncMapper("/<int:join>", text, &Action{Name: "xxeemw", Args: []*ArgType{&ArgType{Name: "join", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	err = svr.Run()
	if err != nil {
		t.Fatal(err)
	}
}
