package peony

import (
	"reflect"
	"testing"
)

func Index() string {
	return "hi"
}

func Index2() Render {
	return &JsonRender{Json: Index}
}

func text(join string) string {
	return join
}

func TestServer(t *testing.T) {
	svr := NewServer()
	svr.Mapper("/", Index, &ActionMethod{Name: "xxeem"})
	svr.Mapper("/2", Index2, &ActionMethod{Name: "xssxeem"})
	svr.Mapper("/<int:join>", text, &ActionMethod{Name: "xxeemw", Args: []*MethodArgType{&MethodArgType{Name: "join", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	//svr.Run()
	tmpl, _ := NewTemplateLoader(".")
	err := tmpl.load()
	t.Log(err)
}
