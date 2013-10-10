package peony

import (
	"reflect"
	"strconv"
)

type Converter func(string, reflect.Type) reflect.Value

type Converters struct {
	KindConverters map[reflect.Kind]Converter
}

func StringConverter(v string, typ reflect.Type) reflect.Value {
	return reflect.ValueOf(v)
}

func IntConverter(value string, typ reflect.Type) reflect.Value {
	iValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	val := reflect.New(typ)
	val.Elem().SetInt(iValue)
	return val.Elem()
}

func UintConverter(value string, typ reflect.Type) reflect.Value {
	iValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	val := reflect.New(typ)
	val.Elem().SetUint(iValue)
	return val.Elem()
}

func FloatConverter(value string, typ reflect.Type) reflect.Value {
	fValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	val := reflect.New(typ)
	val.Elem().SetFloat(fValue)
	return val.Elem()
}

func NewConverters() *Converters {
	c := &Converters{KindConverters: map[reflect.Kind]Converter{}}
	c.KindConverters[reflect.Int] = IntConverter
	c.KindConverters[reflect.Int8] = IntConverter
	c.KindConverters[reflect.Int16] = IntConverter
	c.KindConverters[reflect.Int32] = IntConverter
	c.KindConverters[reflect.Int64] = IntConverter

	c.KindConverters[reflect.Uint] = UintConverter
	c.KindConverters[reflect.Uint8] = UintConverter
	c.KindConverters[reflect.Uint16] = UintConverter
	c.KindConverters[reflect.Uint32] = UintConverter
	c.KindConverters[reflect.Uint64] = UintConverter

	c.KindConverters[reflect.Float32] = FloatConverter
	c.KindConverters[reflect.Float64] = FloatConverter

	c.KindConverters[reflect.String] = StringConverter

	c.KindConverters[reflect.Ptr] = func(value string, typ reflect.Type) reflect.Value {
		return c.Convert(value, typ.Elem()).Addr()
	}

	return c
}

func (c *Converters) Convert(value string, typ reflect.Type) reflect.Value {
	converter := c.KindConverters[typ.Kind()]
	if converter != nil {
		return converter(value, typ)
	}
	return reflect.Zero(typ)
}

func ArgConvert(c *Converters, p *Params, argType *MethodArgType) reflect.Value {
	vals, ok := p.Values[argType.Name]
	if !ok || len(vals) == 0 {
		return reflect.Zero(argType.Type)
	}
	return c.Convert(vals[0], argType.Type)
}
