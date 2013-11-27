package peony

import (
	"fmt"
	"reflect"
)

const (
	BEFORE int = iota
	AFTER
	FINALLY
	PANIC
)

type Interceptors map[string][]*Interceptor

func NewInterceptors() Interceptors {
	return make(map[string][]*Interceptor)
}

type Interceptor struct {
	When     int
	Priority int
	Call     reflect.Value
	NumIn    int
	Function interface{}
	Method   interface{}
	Target   reflect.Type
}

func GetInterceptorFilter(svr *Server) Filter {
	return func(c *Controller, filter []Filter) {
		// just now can't support interceptor for func action
		if _, ok := c.action.(*MethodAction); !ok {
			filter[0](c, filter[1:])
			return
		}
		interceptors := svr.interceptors
		defer interceptors.Invoke(c, FINALLY)
		interceptors.Invoke(c, BEFORE)
		//already get the result
		if c.render != nil {
			return
		}
		filter[0](c, filter[1:])
		interceptors.Invoke(c, AFTER)
	}
}

func (i *Interceptors) Invoke(c *Controller, when int) {
	var methodAction *MethodAction
	methodAction = c.action.(*MethodAction)
	interceptors := i.GetInterceptor(methodAction.TargetType, when)
	for _, interceptor := range interceptors {
		rsSlice := interceptor.Invoke(c)
		if len(rsSlice) > 0 && rsSlice[0].Type().Implements(renderType) {
			c.render = rsSlice[0].Interface().(Render)
			return
		}
	}
}

func (i *Interceptor) Invoke(c *Controller) []reflect.Value {
	var args = []reflect.Value{}
	if i.Function != nil {
		if i.NumIn == 1 {
			//if func have arg controller
			args = append(args, reflect.ValueOf(c))
		}
	} else {
		var methodAction *MethodAction
		var ok bool
		if methodAction, ok = c.action.(*MethodAction); !ok {
			ERROR.Println("the action are not MethodAction")
			return []reflect.Value{}
		}
		args = append(args, methodAction.Target)
		if i.NumIn == 2 {
			//if func have arg controller
			args = append(args, reflect.ValueOf(c))
		}
	}
	return i.Call.Call(args)
}

func (i *Interceptors) genKey(target reflect.Type, when int) string {
	return fmt.Sprintf("%p:%i", target, when)
}

func (i *Interceptors) AddInterceptor(interceptor *Interceptor) {
	key := i.genKey(interceptor.Target, interceptor.When)
	interceptors := (*i)[key]
	interceptors = append(interceptors, interceptor)
	(*i)[key] = interceptors
}

func (i *Interceptors) GetInterceptor(typ reflect.Type, when int) []*Interceptor {
	key := i.genKey(typ, when)
	return (*i)[key]
}

//intercept function
func (i *Interceptors) InterceptFunc(call interface{}, target interface{}, when int, priority int) {
	callVal := reflect.ValueOf(call)
	callType := callVal.Type()
	numIn := callType.NumIn()

	var interceptor *Interceptor
	switch {
	case callType.Kind() != reflect.Func:
		goto FAIL
	case numIn > 1:
		goto FAIL
	case numIn == 1:
		callType.In(0)
		if callType.In(0) != ControllerPtrType {
			goto FAIL
		}
	}
	interceptor = &Interceptor{When: when,
		Call:     callVal,
		Function: call,
		NumIn:    numIn,
		Target:   reflect.TypeOf(target),
		Priority: priority,
	}
	i.AddInterceptor(interceptor)
	return
FAIL:
	ERROR.Fatalln("call must be like func([*Controller]) [Render]")
}

//intercept method
func (i *Interceptors) InterceptMethod(call interface{}, when, priority int) {
	callVal := reflect.ValueOf(call)
	callType := callVal.Type()
	numIn := callType.NumIn()
	var interceptor *Interceptor
	switch {
	case callType.Kind() != reflect.Func:
		goto FAIL
	case numIn > 2 || numIn < 1:
		goto FAIL
	case numIn == 2:
		callType.In(1)
		if callType.In(1) != ControllerPtrType {
			goto FAIL
		}
	}

	interceptor = &Interceptor{When: when,
		Call:     callVal,
		Method:   call,
		NumIn:    numIn,
		Target:   callType.In(0),
		Priority: priority,
	}
	i.AddInterceptor(interceptor)
	return
FAIL:
	ERROR.Fatalln("call must be like (*Struct).Method([*Controller])")
}
