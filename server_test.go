package peony

import (
	"reflect"
	"testing"
)

func Index() string {
	return "hi"
}

func Index2() Render {
	return NewJsonRender("hi")
}

type m struct {
}

func (_ *m) Index3() Render {
	return NewTemplateRender(nil)
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
	k := &m{}
	svr.Mapper("/", Index, &ActionMethod{Name: "xxeem"})
	svr.Mapper("/2", Index2, &ActionMethod{Name: "xssxeem"})
	svr.Mapper("/3", k.Index3, &ActionMethod{Name: "recover.go"})
	svr.Mapper("/<int:join>", text, &ActionMethod{Name: "xxeemw", Args: []*MethodArgType{&MethodArgType{Name: "join", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	svr.Run()
}
