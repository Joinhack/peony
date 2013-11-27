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
	actions.RegisterFuncAction(A, &Action{Name: "1"})
	actions.RegisterMethodAction((*S).A, &Action{Name: "2"})
	var methodArgs []reflect.Value
	action := actions.FindAction("1")
	t.Log(action.Invoke(methodArgs))
	action = actions.FindAction("2")
	t.Log(action.Dup().Invoke(methodArgs))
}
