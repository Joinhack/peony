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
	p.Values["m[0]"] = []string{"1"}
	p.Values["m[1]"] = []string{"2"}
	p.Values["m[2]"] = []string{"3"}
	p.Values["m[5]"] = []string{"3"}
	p.Values["m[100]"] = []string{"100"}

	t.Log(c.Convert(&p, "10", reflect.TypeOf((*int)(nil)).Elem()).Int() == 10)
	t.Log(c.Convert(&p, "10", reflect.TypeOf((*int)(nil))).Elem().Int() == 10)
	t.Log(c.Convert(&p, "1001a", reflect.TypeOf((*int32)(nil)).Elem()).Int() == 0)

	t.Log(c.Convert(&p, "1001.2", reflect.TypeOf((*float32)(nil)).Elem()).Float())
	info := c.Convert(&p, "j", reflect.TypeOf((*Info)(nil))).Interface().(*Info)
	t.Log(info)
	t.Log(c.Convert(&p, "m", reflect.TypeOf((*[]int)(nil)).Elem()).Index(1).Int())
	t.Log(c.Convert(&p, "m", reflect.TypeOf((*[]int)(nil)).Elem()).Index(100).Int())
	t.Log(c.Convert(&p, "m", reflect.TypeOf((*[]int)(nil)).Elem()).Index(100).Int())

}

func TestKindRevertCoverter(t *testing.T) {
	c := NewConvertors()
	p := map[string]string{}
	c.ReverseConvert(p, "i", 10)
	c.ReverseConvert(p, "i32", int32(10))
	t.Log(p)
	c.ReverseConvert(p, "i64", int64(10))
	t.Log(p)

	c.ReverseConvert(p, "ui64", uint64(10))
	t.Log(p)

	c.ReverseConvert(p, "ui", uint(10))
	t.Log(p)

	p = map[string]string{}
	c.ReverseConvert(p, "i", 10.2)
	t.Log(p)

	p = map[string]string{}
	c.ReverseConvert(p, "i", float64(10.2))
	t.Log(p)

	p = map[string]string{}
	c.ReverseConvert(p, "i", "ss")
	t.Log(p)

	p = map[string]string{}
	c.ReverseConvert(p, "i", []string{"1", "2", "3"})
	t.Log(p)

	p = map[string]string{}
	c.ReverseConvert(p, "i", []int{1, 2, 3})
	t.Log(p)

	p = map[string]string{}
	c.ReverseConvert(p, "i", []float32{1.1, 2.2, 3.3})
	t.Log(p)
}
