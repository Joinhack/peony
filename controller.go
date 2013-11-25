package peony

import (
	"fmt"
	"reflect"
)

type Controller struct {
	Resp           *Response
	Req            *Request
	actionName     string
	action         Action
	params         *Params
	render         Render
	templateLoader *TemplateLoader
	Session        *Session
	Server         *Server
}

type Action interface {
	GetName() string
	Args() []*MethodArgType
	Call([]reflect.Value) []reflect.Value
}

type MethodAction struct {
	Name       string
	value      reflect.Value
	MethodArgs []*MethodArgType
}

type TypeAction struct {
	Name       string
	RecvType   reflect.Type
	MethodName string
	MethodArgs []*MethodArgType
}

func (t *TypeAction) GetName() string {
	return t.Name
}

func (t *TypeAction) Args() []*MethodArgType {
	return t.MethodArgs
}

func (t *TypeAction) Call(in []reflect.Value) []reflect.Value {
	value := reflect.New(t.RecvType)
	method := value.MethodByName(t.MethodName)
	return method.Call(in)
}

func (m *MethodAction) GetName() string {
	return m.Name
}

func (m *MethodAction) Args() []*MethodArgType {
	return m.MethodArgs
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

func (a *ActionContainer) RegisterAction(action Action) {
	a.methodActions[action.GetName()] = action
}

func (c *Controller) NotFound(msg string, args ...interface{}) {
	c.Resp.WriteContentTypeCode(404, "text/html")
	text := msg
	if len(args) > 0 {
		text = fmt.Sprintf(msg, args)
	}
	c.Resp.Write([]byte(text))
}

func GetActionInvokeFilter(server *Server) Filter {
	return func(controller *Controller, _ []Filter) {
		converter := server.converter
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
}
