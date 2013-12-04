package peony

import (
	"fmt"
	"reflect"
	"strconv"
)

type Convert func(*Params, string, reflect.Type) reflect.Value

type ReverseConvert func(map[string]string, string, interface{})

type Convertor struct {
	Convert        Convert
	ReverseConvert ReverseConvert
}

type Convertors struct {
	KindConvertors map[reflect.Kind]*Convertor
	TypeConvertors map[reflect.Type]*Convertor
}

func ValueConvertor(convert func(v string, typ reflect.Type) reflect.Value, reverseConvert ReverseConvert) *Convertor {
	return &Convertor{func(p *Params, name string, typ reflect.Type) reflect.Value {
		vals, ok := p.Values[name]
		if !ok || len(vals) == 0 {
			return reflect.Zero(typ)
		}
		return convert(vals[0], typ)
	}, reverseConvert}
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

// func StructConvert(v string, typ reflect.Type) reflect.Value {
// 	return
// }

func IntReverseConvert(p map[string]string, name string, v interface{}) {
	p[name] = fmt.Sprintf("%d", v)
}

func NewConvertors() *Convertors {
	c := &Convertors{
		KindConvertors: map[reflect.Kind]*Convertor{},
		TypeConvertors: map[reflect.Type]*Convertor{},
	}
	c.KindConvertors[reflect.Int] = ValueConvertor(IntConvert, IntReverseConvert)
	c.KindConvertors[reflect.Int8] = ValueConvertor(IntConvert, IntReverseConvert)
	c.KindConvertors[reflect.Int16] = ValueConvertor(IntConvert, IntReverseConvert)
	c.KindConvertors[reflect.Int32] = ValueConvertor(IntConvert, IntReverseConvert)
	c.KindConvertors[reflect.Int64] = ValueConvertor(IntConvert, IntReverseConvert)

	c.KindConvertors[reflect.Uint] = ValueConvertor(UintConvert, IntReverseConvert)
	c.KindConvertors[reflect.Uint8] = ValueConvertor(UintConvert, IntReverseConvert)
	c.KindConvertors[reflect.Uint16] = ValueConvertor(UintConvert, IntReverseConvert)
	c.KindConvertors[reflect.Uint32] = ValueConvertor(UintConvert, IntReverseConvert)
	c.KindConvertors[reflect.Uint64] = ValueConvertor(UintConvert, IntReverseConvert)

	c.KindConvertors[reflect.Float32] = ValueConvertor(FloatConvert, IntReverseConvert)
	c.KindConvertors[reflect.Float64] = ValueConvertor(FloatConvert, IntReverseConvert)

	c.KindConvertors[reflect.String] = ValueConvertor(StringConvert, IntReverseConvert)

	//c.KindConverts[reflect.Struct] = StructConvert

	c.KindConvertors[reflect.Ptr] = &Convertor{
		func(p *Params, name string, typ reflect.Type) reflect.Value {
			return c.Convert(p, name, typ.Elem()).Addr()
		},
		func(p map[string]string, name string, v interface{}) {
			c.ReverseConvert(p, name, reflect.ValueOf(v).Elem().Interface())
		},
	}
	return c
}

func (c *Convertors) Convert(p *Params, name string, typ reflect.Type) reflect.Value {
	converter := c.TypeConvertors[typ]
	if converter == nil {
		converter = c.KindConvertors[typ.Kind()]
	}
	if converter != nil {
		return converter.Convert(p, name, typ)
	}
	return reflect.Zero(typ)
}

func (c *Convertors) ReverseConvert(p map[string]string, name string, val interface{}) {
	typ := reflect.TypeOf(val)
	converter := c.TypeConvertors[typ]
	if converter == nil {
		converter = c.KindConvertors[typ.Kind()]
	}
	if converter != nil {
		converter.ReverseConvert(p, name, typ)
	}
}
