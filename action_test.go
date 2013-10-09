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
	actions := NewActionMethods()
	actions.RegisterAction(A, &ActionMethod{Name: "xxee"})
	actions.RegisterAction((&S{}).A, &ActionMethod{Name: "xxeem"})
	var methodArgs []reflect.Value
	action := actions.FindAction("xxee")
	t.Log(action.value.Call(methodArgs))
	action = actions.FindAction("xxeem")
	t.Log(action.value.Call(methodArgs))
}
