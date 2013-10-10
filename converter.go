package peony

import (
	"reflect"
	"strconv"
)

type Convert func(string, reflect.Type) reflect.Value

type Converter struct {
	KindConverts map[reflect.Kind]Convert
}

func StringConvert(v string, typ reflect.Type) reflect.Value {
	return reflect.ValueOf(v)
}

func IntConvert(value string, typ reflect.Type) reflect.Value {
	iValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		WARN.Printf("can't convert \"%s\" to int\n", value)
		return reflect.Zero(typ)
	}
	val := reflect.New(typ)
	val.Elem().SetInt(iValue)
	return val.Elem()
}

func UintConvert(value string, typ reflect.Type) reflect.Value {
	iValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		WARN.Printf("can't convert \"%s\" to uint\n", value)
		return reflect.Zero(typ)
	}
	val := reflect.New(typ)
	val.Elem().SetUint(iValue)
	return val.Elem()
}

func FloatConvert(value string, typ reflect.Type) reflect.Value {
	fValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		WARN.Printf("can't convert \"%s\" to float\n", value)
		return reflect.Zero(typ)
	}
	val := reflect.New(typ)
	val.Elem().SetFloat(fValue)
	return val.Elem()
}

func NewConverter() *Converter {
	c := &Converter{KindConverts: map[reflect.Kind]Convert{}}
	c.KindConverts[reflect.Int] = IntConvert
	c.KindConverts[reflect.Int8] = IntConvert
	c.KindConverts[reflect.Int16] = IntConvert
	c.KindConverts[reflect.Int32] = IntConvert
	c.KindConverts[reflect.Int64] = IntConvert

	c.KindConverts[reflect.Uint] = UintConvert
	c.KindConverts[reflect.Uint8] = UintConvert
	c.KindConverts[reflect.Uint16] = UintConvert
	c.KindConverts[reflect.Uint32] = UintConvert
	c.KindConverts[reflect.Uint64] = UintConvert

	c.KindConverts[reflect.Float32] = FloatConvert
	c.KindConverts[reflect.Float64] = FloatConvert

	c.KindConverts[reflect.String] = StringConvert

	c.KindConverts[reflect.Ptr] = func(value string, typ reflect.Type) reflect.Value {
		return c.Convert(value, typ.Elem()).Addr()
	}
	return c
}

func (c *Converter) Convert(value string, typ reflect.Type) reflect.Value {
	converter := c.KindConverts[typ.Kind()]
	if converter != nil {
		return converter(value, typ)
	}
	return reflect.Zero(typ)
}

func ArgConvert(c *Converter, p *Params, argType *MethodArgType) reflect.Value {
	vals, ok := p.Values[argType.Name]
	if !ok || len(vals) == 0 {
		return reflect.Zero(argType.Type)
	}
	return c.Convert(vals[0], argType.Type)
}
