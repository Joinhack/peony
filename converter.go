package peony

import (
	"reflect"
	"strconv"
)

type Converter interface {
	Convert(value string) reflect.Value
}

type Converters struct {
	KindConverters map[reflect.Kind]Converter
}

type IntConverter struct {
	Converter
	Type reflect.Type //e.g. reflect.Type(int) reflect.Type(int8)
}

type UintConverter struct {
	Converter
	Type reflect.Type //e.g. reflect.Type(uint) reflect.Type(uint8)
}

type FloatConverter struct {
	Converter
	Type reflect.Type //e.g. reflect.Type(float) reflect.Type(float64)
}

type StringConverter struct {
	Converter
}

func (s *StringConverter) Convert(v string) reflect.Value {
	return reflect.ValueOf(v)
}

func (i *IntConverter) Convert(value string) reflect.Value {
	iValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		iValue = 0
	}
	val := reflect.New(i.Type)
	val.Elem().SetInt(iValue)
	return val.Elem()
}

func (i *UintConverter) Convert(value string) reflect.Value {
	iValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		iValue = 0
	}
	val := reflect.New(i.Type)
	val.Elem().SetUint(iValue)
	return val.Elem()
}

func (f *FloatConverter) Convert(value string) reflect.Value {
	fValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		fValue = 0
	}
	val := reflect.New(f.Type)
	val.Elem().SetFloat(fValue)
	return val.Elem()
}

func intConverter(intPtrType reflect.Type) Converter {
	return &IntConverter{Type: intPtrType.Elem()}
}

func uintConverter(uintPtrType reflect.Type) Converter {
	return &UintConverter{Type: uintPtrType.Elem()}
}

func floatConverter(floatPtrType reflect.Type) Converter {
	return &FloatConverter{Type: floatPtrType.Elem()}
}

func NewConverters() *Converters {
	c := &Converters{KindConverters: map[reflect.Kind]Converter{}}
	c.KindConverters[reflect.Int] = intConverter(reflect.TypeOf((*int)(nil)))
	c.KindConverters[reflect.Int8] = intConverter(reflect.TypeOf((*int8)(nil)))
	c.KindConverters[reflect.Int16] = intConverter(reflect.TypeOf((*int16)(nil)))
	c.KindConverters[reflect.Int32] = intConverter(reflect.TypeOf((*int32)(nil)))
	c.KindConverters[reflect.Int64] = intConverter(reflect.TypeOf((*int64)(nil)))

	c.KindConverters[reflect.Uint] = uintConverter(reflect.TypeOf((*uint)(nil)))
	c.KindConverters[reflect.Uint8] = uintConverter(reflect.TypeOf((*uint8)(nil)))
	c.KindConverters[reflect.Uint16] = uintConverter(reflect.TypeOf((*uint16)(nil)))
	c.KindConverters[reflect.Uint32] = uintConverter(reflect.TypeOf((*uint32)(nil)))
	c.KindConverters[reflect.Uint64] = uintConverter(reflect.TypeOf((*uint64)(nil)))

	c.KindConverters[reflect.Float32] = floatConverter(reflect.TypeOf((*float32)(nil)))
	c.KindConverters[reflect.Float64] = floatConverter(reflect.TypeOf((*float64)(nil)))

	c.KindConverters[reflect.String] = &StringConverter{}

	return c
}

func (c *Converters) Convert(value string, targetType reflect.Type) reflect.Value {
	converter := c.KindConverters[targetType.Kind()]
	if converter != nil {
		return converter.Convert(value)
	}
	return reflect.Zero(targetType)
}
