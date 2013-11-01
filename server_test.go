package peony

import (
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

type AS struct {
}

func (a *AS) T() Render {
	return NewTextRender("s")
}

func TestServer(t *testing.T) {
	svr := NewServer(":8080")
	loader, err := NewTemplateLoader(".")
	err = loader.load()
	if err != nil {
		panic(err)
	}
	svr.templateLoader = loader
	svr.MethodMapper("/", (*AS).T, &MethodAction{Name: "AS.T"})
	svr.MethodMapper("/json", Json, &MethodAction{Name: "xssxeem"})
	svr.MethodMapper("/template", Template, &MethodAction{Name: "recover.go"})
	svr.MethodMapper("/xml", Xml, &MethodAction{Name: "xml"})
	svr.MethodMapper("/<int:join>", text, &MethodAction{Name: "xxeemw", methodArgs: []*MethodArgType{&MethodArgType{Name: "join", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	svr.Run()
}
