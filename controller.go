package peony

import (
	"errors"
	"fmt"
	"reflect"
)

type Controller struct {
	resp           *Response
	req            *Request
	method         *ActionMethod
	action         string
	params         *Params
	render         Render
	templateLoader *TemplateLoader
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
	a.resp.Write([]byte(text))
}

func ActionInvoke(converter *Converter, controller *Controller) {
	methodArgs := make([]reflect.Value, 0, len(controller.method.Args))
	for _, arg := range controller.method.Args {
		argValue := ArgConvert(converter, controller.params, arg)
		methodArgs = append(methodArgs, argValue)
	}
	rsSlice := controller.method.value.Call(methodArgs)
	if len(rsSlice) > 0 {
		rs := rsSlice[0]
		if rs.Type().Kind() == reflect.String {
			controller.render = &TextRender{Text: rs.String()}
		} else if rs.Type().Implements(renderType) {
			controller.render = rs.Interface().(Render)
		}
	}
}

func GetActionFilter(server *Server) Filter {
	return func(controller *Controller, _ []Filter) {
		// bind actionMethod to controller
		controller.method = server.actions.FindAction(controller.action)
		if controller.method == nil {
			controller.NotFound("intenal error")
			ERROR.Println("can't find action method by name:", controller.action)
			return
		}
		ActionInvoke(server.converter, controller)
		controller.render.Apply(controller)
	}
}
