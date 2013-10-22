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

func TestServer(t *testing.T) {
	svr := NewServer()
	loader, err := NewTemplateLoader(".")
	err = loader.load()
	if err != nil {
		panic(err)
	}
	svr.templateLoader = loader
	svr.Mapper("/", Index, &ActionMethod{Name: "xxeem"})
	svr.Mapper("/json", Json, &ActionMethod{Name: "xssxeem"})
	svr.Mapper("/template", Template, &ActionMethod{Name: "recover.go"})
	svr.Mapper("/xml", Xml, &ActionMethod{Name: "xml"})
	svr.Mapper("/<int:join>", text, &ActionMethod{Name: "xxeemw", Args: []*MethodArgType{&MethodArgType{Name: "join", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	svr.Run()
}
