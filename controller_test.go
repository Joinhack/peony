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

func (s S) B() string {
	return "S.A"
}

func TestRegister(t *testing.T) {
	actions := NewActionContainer()
	actions.RegisterFuncAction(A, &Action{Name: "1"})
	actions.RegisterMethodAction((*S).A, &Action{Name: "2"})
	actions.RegisterMethodAction(S.B, &Action{Name: "3"})
	var callArgs []reflect.Value
	action := actions.FindAction("1")
	t.Log(action.Invoke(callArgs))
	action = actions.FindAction("2")
	t.Log(action.Dup().Invoke(callArgs))
	action = actions.FindAction("3")
	t.Log(action.Dup().Invoke(callArgs))
}
