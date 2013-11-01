package peony

import (
	"testing"
	"unsafe"
)

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

func typeaddr(i interface{}) *rtype {
	eface := *(*emptyInterface)(unsafe.Pointer(&i))
	return eface.typ
}

type TestType struct {
	elem  string
	elem2 int
}

type ptrType struct {
	rtype `reflect:"ptr"`
	elem  *rtype // pointer element (pointed at) type
}

func TestAddr(t *testing.T) {
	//Test type have the same type addr
	t.Logf("TestType Ptr addr:%p", typeaddr((*TestType)(nil)))
	t.Logf("TestType Ptr addr:%p", typeaddr((*TestType)(nil)))
	t.Logf("TestType Ptr addr:%p", typeaddr(&TestType{}))

	t.Logf("TestType addr:%p", (*ptrType)(unsafe.Pointer(typeaddr(&TestType{}))).elem)
	t.Logf("TestType addr:%p", (*ptrType)(unsafe.Pointer(typeaddr((*TestType)(nil)))).elem)
	t.Logf("TestType addr:%p", typeaddr(TestType{}))
}
