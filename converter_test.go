package peony

import (
	"reflect"
	"testing"
)

func TestKindCoverter(t *testing.T) {
	t.Log(NewConverters().Convert("10", reflect.TypeOf((*int)(nil)).Elem()).Int() == 10)

	t.Log(NewConverters().Convert("1001a", reflect.TypeOf((*int32)(nil)).Elem()).Int() == 0)

	t.Log(NewConverters().Convert("1001.2", reflect.TypeOf((*float32)(nil)).Elem()).Float())
}
