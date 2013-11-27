package peony

import (
	"reflect"
	"testing"
)

type M struct {
	*testing.T
}

func (m *M) BEFORE() {
	m.Log("BEFORE")
}

func (m *M) Do() {

}

func (m *M) AFTER() {
	m.Log("AFTER")
}

func TestIntecept(t *testing.T) {
	interceptors := NewInterceptors()
	interceptors.InterceptMethod((*M).BEFORE, BEFORE, 0)
	interceptors.InterceptMethod((*M).AFTER, AFTER, 0)
	c := &Controller{}
	m := &MethodAction{}
	m.Target = reflect.ValueOf(&M{t})
	m.TargetType = reflect.TypeOf((*M)(nil))
	c.action = m
	interceptors.Invoke(c, BEFORE)
	interceptors.Invoke(c, AFTER)
	interceptors.Invoke(c, FINALLY)
}
