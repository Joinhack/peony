package peony

import (
	"errors"
	"fmt"
	"reflect"
)

type Controller struct {
	resp   *Response
	req    *Request
	method *ActionMethod
	action string
	params *Params
}

func NewController(w *Response, r *Request) *Controller {
	c := &Controller{resp: w, req: r}
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

func (a *Controller) NotFound(msg string, args ...interface{}) {
	a.resp.WriteHeader(404, "text/html")
	text := msg
	if len(args) > 0 {
		text = fmt.Sprintf(msg, args)
	}
	a.resp.Output.Write([]byte(text))
}

func ActionInvoke(converter *Converter, controller *Controller) {
	methodArgs := make([]reflect.Value, 0, len(controller.method.Args))
	for _, arg := range controller.method.Args {
		argValue := ArgConvert(converter, controller.params, arg)
		methodArgs = append(methodArgs, argValue)
	}
	controller.method.value.Call(methodArgs)
}
