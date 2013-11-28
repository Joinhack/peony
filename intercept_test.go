package peony

import (
	"reflect"
	"testing"
)

var gt *testing.T

type M struct {
	*testing.T
}

func (m *M) BEFORE() {
	m.T = gt
	m.Log("BEFORE, init m.T")
}

func (m M) Do() {
	m.Log("Do")
}

func (m M) AFTER() {
	m.Log("AFTER")
}

func TestIntecept(t *testing.T) {
	gt = t
	interceptors := NewInterceptors()
	interceptors.InterceptMethod((*M).BEFORE, BEFORE, 0)
	interceptors.InterceptMethod((M).AFTER, AFTER, 0)
	c := &Controller{}
	ac := NewActionContainer()
	ac.RegisterMethodAction(M.Do, &Action{Name: "xx"})
	action := ac.FindAction("xx")
	c.action = action.Dup()
	interceptors.Invoke(c, BEFORE)
	c.action.Invoke([]reflect.Value{})
	interceptors.Invoke(c, AFTER)
	interceptors.Invoke(c, FINALLY)
}
