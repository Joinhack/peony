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

var ControllerPtrType = reflect.TypeOf((*Controller)(nil))

type Action interface {
	GetName() string
	Dup() Action
	Args() []*ArgType
	Call([]reflect.Value) []reflect.Value
}

type FuncAction struct {
	Name       string
	value      reflect.Value
	MethodArgs []*ArgType
}

type MethodAction struct {
	Name       string
	TargetType reflect.Type
	Target     reflect.Value
	MethodName string
	MethodArgs []*ArgType
}

func (m *MethodAction) GetName() string {
	return m.Name
}

func (m *MethodAction) Dup() Action {
	methodAction := new(MethodAction)
	*methodAction = *m
	methodAction.Target = reflect.New(methodAction.TargetType)
	return methodAction
}
func (m *MethodAction) Args() []*ArgType {
	return m.MethodArgs
}

func (m *MethodAction) Call(in []reflect.Value) []reflect.Value {
	value := m.Target
	method := value.MethodByName(m.MethodName)
	return method.Call(in)
}

func (m *FuncAction) GetName() string {
	return m.Name
}

func (f *FuncAction) Dup() Action {
	FuncAction := new(FuncAction)
	*FuncAction = *f
	return FuncAction
}

func (f *FuncAction) Args() []*ArgType {
	return f.MethodArgs
}

func (f *FuncAction) Call(in []reflect.Value) []reflect.Value {
	return f.value.Call(in)
}

type ArgType struct {
	Name string // arg name
	Type reflect.Type
}

type ActionContainer struct {
	Actions map[string]Action
}

func NewActionContainer() *ActionContainer {
	actions := &ActionContainer{Actions: make(map[string]Action, 0)}
	return actions
}

func (a *ActionContainer) FindAction(name string) Action {
	return a.Actions[name]
}

func (a *ActionContainer) RegisterAction(action Action) {
	a.Actions[action.GetName()] = action
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
