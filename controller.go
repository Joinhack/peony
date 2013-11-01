package peony

import (
	"errors"
	"fmt"
	"reflect"
)

type Controller struct {
	resp           *Response
	req            *Request
	actionName     string
	action         Action
	params         *Params
	render         Render
	templateLoader *TemplateLoader
}

type Action interface {
	Args() []*MethodArgType
	Call([]reflect.Value) []reflect.Value
}

type MethodAction struct {
	Name       string
	MethodName string
	value      reflect.Value
	methodArgs []*MethodArgType
}

func (m *MethodAction) Args() []*MethodArgType {
	return m.methodArgs
}

func (m *MethodAction) Call(in []reflect.Value) []reflect.Value {
	return m.value.Call(in)
}

type MethodArgType struct {
	Name string // arg name
	Type reflect.Type
}

type ActionContainer struct {
	methodActions map[string]Action
}

func NewActionContainer() *ActionContainer {
	actions := &ActionContainer{methodActions: make(map[string]Action, 0)}
	return actions
}

func (a *ActionContainer) FindAction(name string) Action {
	return a.methodActions[name]
}

func (a *ActionContainer) RegisterMethodAction(action interface{}, method *MethodAction) error {
	if reflect.TypeOf(action).Kind() != reflect.Func {
		return errors.New("action should be method")
	}
	if a.methodActions[method.Name] != nil {
		return errors.New("Action already exist")
	}
	method.value = reflect.ValueOf(action)
	a.methodActions[method.Name] = method
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
	args := controller.action.Args()
	methodArgs := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		argValue := ArgConvert(converter, controller.params, arg)
		methodArgs = append(methodArgs, argValue)
	}
	rsSlice := controller.action.Call(methodArgs)
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
		controller.action = server.actions.FindAction(controller.actionName)
		if controller.action == nil {
			controller.NotFound("intenal error")
			ERROR.Println("can't find action method by name:", controller.actionName)
			return
		}
		ActionInvoke(server.converter, controller)
		controller.render.Apply(controller)
	}
}
