package peony

import (
	"reflect"
	"testing"
)

func A() string {
	return "A"
}

type S struct {
}

func (s *S) A() string {
	return "S.A"
}

func TestRegister(t *testing.T) {
	actions := NewActionContainer()
	actions.RegisterAction(&MethodAction{Name: "xxee", value: reflect.ValueOf(A)})
	actions.RegisterAction(&MethodAction{Name: "xxeem", value: reflect.ValueOf((&S{}).A)})
	var methodArgs []reflect.Value
	action := actions.FindAction("xxee")
	t.Log(action.Call(methodArgs))
	action = actions.FindAction("xxeem")
	t.Log(action.Call(methodArgs))
}
