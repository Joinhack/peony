package peony

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

var (
	WebSocketConnType = reflect.TypeOf((*websocket.Conn)(nil))
	RequestType       = reflect.TypeOf((*Request)(nil))
	ResponseType      = reflect.TypeOf((*Response)(nil))
	SessionType       = reflect.TypeOf((*Session)(nil))
	ControllerType    = reflect.TypeOf((*Controller)(nil))
	AppType           = reflect.TypeOf((*App)(nil))
	FlashType           = reflect.TypeOf((*Flash)(nil))
	NotFunc           = errors.New("action should be a func")
	NotMethod         = errors.New("action should be a method")
	ValueMustbePtr    = errors.New("value must be should be ptr")
)

type Controller struct {
	Resp           *Response
	Req            *Request
	Session        *Session
	Server         *Server
	actionName     string
	action         *Action
	Params         *Params
	render         Renderer
	flash          Flash
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

func (c *Controller) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
}

//Get Param Value
func (c *Controller) GetParam(name string, val interface{}) error {
	valType := reflect.TypeOf(val)
	if valType.Kind() != reflect.Ptr {
		return ValueMustbePtr
	}
	value := reflect.ValueOf(val)
	paramValue := c.Server.convertors.Convert(c.Params, name, valType)
	value.Elem().Set(paramValue.Elem())
	return nil
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
	Actions     map[string]*Action //e.g. key is Controller.Call or Function
	FuncActions map[string]*Action //e.g. key is the type
}

func NewActionContainer() *ActionContainer {
	actions := &ActionContainer{
		Actions:     make(map[string]*Action),
		FuncActions: make(map[string]*Action),
	}
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
	a.FuncActions[fmt.Sprintf("%p", function)] = action
	return nil
}

func (a *ActionContainer) FindActionByFunc(i interface{}) *Action {
	funcType := reflect.TypeOf(i)
	if funcType.Kind() != reflect.Func {
		ERROR.Println("registor func action error:", NotFunc)
		panic(NotFunc)
	}
	return a.FuncActions[fmt.Sprintf("%p", i)]
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

	c.render = NotFound(msg, args...)
}

func GetActionInvokeFilter(server *Server) Filter {
	return func(controller *Controller, _ []Filter) {
		convertors := server.convertors
		args := controller.action.Args
		callArgs := make([]reflect.Value, 0, len(args))
		for _, arg := range args {
			var argValue reflect.Value
			switch arg.Type {
			case WebSocketConnType:
				argValue = reflect.ValueOf(controller.Req.WSConn)
			case RequestType:
				argValue = reflect.ValueOf(controller.Req)
			case ResponseType:
				argValue = reflect.ValueOf(controller.Resp)
			case SessionType:
				argValue = reflect.ValueOf(controller.Session)
			case AppType:
				argValue = reflect.ValueOf(controller.Server.App)
			case FlashType:
				argValue = reflect.ValueOf(&controller.flash)
			default:
				argValue = convertors.Convert(controller.Params, arg.Name, arg.Type)
			}
			callArgs = append(callArgs, argValue)
		}
		rsSlice := controller.action.Invoke(callArgs)
		if len(rsSlice) > 0 {
			rsVal := rsSlice[0]
			if rsVal.Type().Kind() == reflect.String {
				controller.render = RenderText(rsVal.String())
				return
			} else if rsVal.Type().Implements(RendererType) {
				rs := rsVal.Interface()
				if rs != nil {
					controller.render = rs.(Renderer)
				}
			}
		}
	}
}
