package peony

import (
	"errors"
	"reflect"
)

type Action struct {
	resp   *Response
	req    *Request
	method *ActionMethod
	params *Params
}

func NewAction(w *Response, r *Request) *Action {
	c := &Action{resp: w, req: r}
	return c
}

type ActionMethod struct {
	Name  string        //action name
	value reflect.Value //method
	Args  []*MethodArgType
}

type MethodArgType struct {
	Name string // arg name
	Type reflect.Type
}

type ActionMethods struct {
	actions map[string]*ActionMethod
}

func NewActionMethods() *ActionMethods {
	actions := &ActionMethods{actions: make(map[string]*ActionMethod, 0)}
	return actions
}

func (a *ActionMethods) FindAction(name string) *ActionMethod {
	return a.actions[name]
}

func (a *ActionMethods) RegisterAction(action interface{}, method *ActionMethod) error {
	if reflect.TypeOf(action).Kind() != reflect.Func {
		return errors.New("action should be method")
	}
	if a.actions[method.Name] != nil {
		return errors.New("Action already exist")
	}
	method.value = reflect.ValueOf(action)
	a.actions[method.Name] = method
	return nil
}

func ActionInvoke(converters *Converters, action *Action) {
	methodArgs := make([]reflect.Value, 0, len(action.method.Args))
	for _, arg := range action.method.Args {
		argValue := ArgConvert(converters, action.params, arg)
		methodArgs = append(methodArgs, argValue)
	}
	action.method.value.Call(methodArgs)
}
