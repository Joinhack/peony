package peony

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
)

type Set map[interface{}]interface{}

func (s *Set) Add(a interface{}) {
	(*s)[a] = nil
}

func (s *Set) Has(a interface{}) bool {
	if _, ok := (*s)[a]; ok == true {
		return true
	} else {
		return false
	}
}

func NewSet() Set {
	return make(Set, 4)
}

// Reads the lines of the given file.  Panics in the case of error.
func MustReadLines(filename string) []string {
	r, err := ReadLines(filename)
	if err != nil {
		panic("read file error:" + filename)
	}
	return r
}

// Reads the lines of the given file.  Panics in the case of error.
func ReadLines(filename string) ([]string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bytes), "\n"), nil
}

type template interface {
	Execute(io.Writer, interface{}) error
}

func ExecuteTemplate(t template, args interface{}) string {
	b := bytes.Buffer{}
	err := t.Execute(&b, args)
	if err != nil {
		log.Println(err)
	}
	return b.String()
}

func StringSliceContain(vals []string, val string) bool {
	for _, v := range vals {
		if v == val {
			return true
		}
	}
	return false
}

func FindMethod(recvType reflect.Type, funcVal reflect.Value) *reflect.Method {
	for i := 0; i < recvType.NumMethod(); i++ {
		method := recvType.Method(i)
		if method.Func.Pointer() == funcVal.Pointer() {
			return &method
		}
	}
	return nil
}
