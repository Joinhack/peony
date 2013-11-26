package peony

import (
	"reflect"
)

const (
	BEFORE int = iota
	AFTER
	FINALLY
	PANIC
)

type Interceptors struct {
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

func (i *Interceptors) AddInterceptor(interceptor *Interceptor) {
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

//intercept function
func (i *Interceptors) InterceptMethod(call interface{}, when, priority int) {
	callVal := reflect.ValueOf(call)
	callType := callVal.Type()
	numIn := callType.NumIn()
	var interceptor *Interceptor
	switch {
	case callType.Kind() != reflect.Func:
		goto FAIL
	case numIn > 2:
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
