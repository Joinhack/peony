package peony

import (
	"reflect"
	"testing"
	"unsafe"
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

type uncommonType struct {
	name    *string  // name of type
	pkgPath *string  // import path; nil for built-in types like int, string
	methods []method // methods associated with type
}

type iword unsafe.Pointer

type method struct {
	name    *string        // name of method
	pkgPath *string        // nil for exported Names; otherwise import path
	mtyp    *rtype         // method type (without receiver)
	typ     *rtype         // .(*FuncType) underneath (with receiver)
	ifn     unsafe.Pointer // fn used in interface call (one-word receiver)
	tfn     unsafe.Pointer // fn used for normal method call
}

type rtype struct {
	size          uintptr        // size in bytes
	hash          uint32         // hash of type; avoids computation in hash tables
	_             uint8          // unused/padding
	align         uint8          // alignment of variable with this type
	fieldAlign    uint8          // alignment of struct field with this type
	kind          uint8          // enumeration for C
	alg           *uintptr       // algorithm table (../runtime/runtime.h:/Alg)
	gc            unsafe.Pointer // garbage collection data
	string        *string        // string form; unnecessary but undeniably useful
	*uncommonType                // (relatively) uncommon fields
	ptrToThis     *rtype         // type for pointer to this type, if used in binary or has methods
}

type emptyInterface struct {
	typ  *rtype
	word iword
}

func ts(i interface{}) {
	eface := *(*emptyInterface)(unsafe.Pointer(&i))
	println(unsafe.Sizeof(eface))
}
func TestServer(t *testing.T) {
	svr := NewServer(":8080")
	loader, err := NewTemplateLoader(".")
	err = loader.load()
	if err != nil {
		panic(err)
	}
	svr.templateLoader = loader
	err = svr.Mapper("/", (*AS)(nil), &TypeAction{Name: "AS.T", MethodName: "T"})
	if err != nil {
		panic(err)
	}
	ts((*AS)(nil))
	ts((*AS)(nil))
	svr.Mapper("/json", Json, &MethodAction{Name: "xssxeem"})
	svr.Mapper("/template", Template, &MethodAction{Name: "recover.go"})
	svr.Mapper("/xml", Xml, &MethodAction{Name: "xml"})
	svr.Mapper("/<int:join>", text, &MethodAction{Name: "xxeemw", methodArgs: []*MethodArgType{&MethodArgType{Name: "join", Type: reflect.TypeOf((*string)(nil)).Elem()}}})
	svr.Run()
}
