package peony

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	NotFunc   = errors.New("action should be a func")
	NotMethod = errors.New("action should be a method")
)

type Controller struct {
	Resp           *Response
	Req            *Request
	Session        *Session
	Server         *Server
	actionName     string
	action         *Action
	params         *Params
	render         Render
	templateLoader *TemplateLoader
}

var ControllerPtrType = reflect.TypeOf((*Controller)(nil))

type Action struct {
	Name       string
	call       reflect.Value
	function   interface{}   //if Action is function, not nil
	method     interface{}   //if Action is method, not nil
	targetType reflect.Type  //if is method action, targetType is recv type. e.g. (*X).DO *X is  the targetType
	targetPtr  reflect.Value // the ptr value for targetType, always is a ptr value
	Args       []*ArgType
}

func (a *Action) Invoke(args []reflect.Value) []reflect.Value {
	var callArgs []reflect.Value
	if a.method != nil {
		target := a.targetPtr
		if a.targetType.Kind() != reflect.Ptr {
			target = target.Elem()
		}
		callArgs = append(append(callArgs, target), args...)
	} else {
		callArgs = append(callArgs, args...)
	}
	return a.call.Call(callArgs)
}

func (a *Action) Dup() *Action {
	newAction := new(Action)
	*newAction = *a
	if a.method != nil {
		var ptr reflect.Value
		var targetType = a.targetType
		//when targetType is ptr like (*Struct).Call, the first arguments should be ptr's elem
		if targetType.Kind() == reflect.Ptr {
			ptr = reflect.New(targetType.Elem())
		} else {
			ptr = reflect.New(targetType)
		}
		newAction.targetPtr = ptr
	}
	return newAction
}

type ArgType struct {
	Name string // arg name
	Type reflect.Type
}

type ActionContainer struct {
	Actions map[string]*Action
}

func NewActionContainer() *ActionContainer {
	actions := &ActionContainer{Actions: make(map[string]*Action)}
	return actions
}

func (a *ActionContainer) FindAction(name string) *Action {
	return a.Actions[name]
}

func (a *ActionContainer) RegisterFuncAction(function interface{}, action *Action) error {
	funcVal := reflect.ValueOf(function)
	funcType := funcVal.Type()
	if funcType.Kind() != reflect.Func {
		ERROR.Println("registor func action error:", NotFunc)
		return NotFunc
	}
	action.function = function
	action.call = funcVal
	a.Actions[action.Name] = action
	return nil
}

func (a *ActionContainer) RegisterMethodAction(method interface{}, action *Action) error {
	methodVal := reflect.ValueOf(method)
	methodType := methodVal.Type()
	numIn := methodType.NumIn()
	if numIn < 1 {
		ERROR.Println("register method action error:", NotMethod)
		return NotMethod
	}
	targetType := methodType.In(0)
	action.method = method
	action.call = methodVal
	action.targetType = targetType
	a.Actions[action.Name] = action
	return nil
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
		args := controller.action.Args
		callArgs := make([]reflect.Value, 0, len(args))
		for _, arg := range args {
			argValue := ArgConvert(converter, controller.params, arg)
			callArgs = append(callArgs, argValue)
		}
		rsSlice := controller.action.Invoke(callArgs)
		if len(rsSlice) > 0 {
			rsVal := rsSlice[0]
			if rsVal.Type().Kind() == reflect.String {
				controller.render = &TextRender{Text: rsVal.String()}
				return
			} else if rsVal.Type().Implements(renderType) {
				rs := rsVal.Interface()
				if rs != nil {
					controller.render = rs.(Render)
				}
			}
		}
	}
}
