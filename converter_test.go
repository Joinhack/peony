package peony

import (
	"net/url"
	"reflect"
	"testing"
)

func TestKindCoverter(t *testing.T) {
	c := NewConvertors()
	p := Params{Values: make(url.Values)}
	p.Values["10"] = []string{"10"}
	p.Values["10"] = []string{"10"}
	p.Values["1001a"] = []string{"1001a"}
	p.Values["1001.2"] = []string{"1001.2"}
	t.Log(c.Convert(&p, "10", reflect.TypeOf((*int)(nil)).Elem()).Int() == 10)
	t.Log(c.Convert(&p, "10", reflect.TypeOf((*int)(nil))).Elem().Int() == 10)
	t.Log(c.Convert(&p, "1001a", reflect.TypeOf((*int32)(nil)).Elem()).Int() == 0)
	t.Log(c.Convert(&p, "1001.2", reflect.TypeOf((*float32)(nil)).Elem()).Float())
}
