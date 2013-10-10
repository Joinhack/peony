package peony

import (
	"reflect"
	"testing"
)

func TestKindCoverter(t *testing.T) {
	c := NewConverter()
	t.Log(c.Convert("10", reflect.TypeOf((*int)(nil)).Elem()).Int() == 10)
	t.Log(c.Convert("10", reflect.TypeOf((*int)(nil))).Elem().Int() == 10)
	t.Log(c.Convert("1001a", reflect.TypeOf((*int32)(nil)).Elem()).Int() == 0)
	t.Log(c.Convert("1001.2", reflect.TypeOf((*float32)(nil)).Elem()).Float())
}
