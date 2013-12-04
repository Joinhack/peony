package peony

import (
	"net/url"
	"reflect"
	"testing"
)

type Sub struct {
	FLOAT float32
	INT   int64
}

type Info struct {
	Sub           Sub
	Name, Address string
	Id            int
}

func TestKindCoverter(t *testing.T) {
	c := NewConvertors()
	p := Params{Values: make(url.Values)}
	p.Values["10"] = []string{"10"}
	p.Values["10"] = []string{"10"}
	p.Values["1001a"] = []string{"1001a"}
	p.Values["1001.2"] = []string{"1001.2"}
	p.Values["j.Name"] = []string{"hhh"}
	p.Values["j.Id"] = []string{"1"}
	p.Values["j.Address"] = []string{"xxx"}
	p.Values["j.Sub.FLOAT"] = []string{"0.9"}
	p.Values["j.Sub.INT"] = []string{"2"}
	t.Log(c.Convert(&p, "10", reflect.TypeOf((*int)(nil)).Elem()).Int() == 10)
	t.Log(c.Convert(&p, "10", reflect.TypeOf((*int)(nil))).Elem().Int() == 10)
	t.Log(c.Convert(&p, "1001a", reflect.TypeOf((*int32)(nil)).Elem()).Int() == 0)

	t.Log(c.Convert(&p, "1001.2", reflect.TypeOf((*float32)(nil)).Elem()).Float())
	info := c.Convert(&p, "j", reflect.TypeOf((*Info)(nil))).Interface().(*Info)
	t.Log(info)
}
