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
	err = svr.Mapper("/", (*AS)(nil), &MethodAction{Name: "AS.T", MethodName: "T"})
	if err != nil {
		panic(err)
	}
	svr.Mapper("/json", Json, &FuncAction{Name: "xssxeem"})
	svr.Mapper("/template", Template, &FuncAction{Name: "recover.go"})
	svr.Mapper("/xml", Xml, &FuncAction{Name: "xml"})
	svr.Mapper("/<int:join>", text, &FuncAction{Name: "xxeemw", Args: []*ArgType{&ArgType{Name: "join", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	// err = svr.Run()
	// if err != nil {
	// 	t.Fatal(err)
	// }
}
