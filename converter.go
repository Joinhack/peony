package peony

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"reflect"
	"strconv"
	"strings"
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

func nextKey(key string) string {
	fieldLen := strings.IndexAny(key, ".[")
	if fieldLen == -1 {
		return key
	}
	return key[:fieldLen]
}

func getMultipartFile(p *Params, n string) multipart.File {
	for _, fileH := range p.Files[n] {
		file, err := fileH.Open()
		if err != nil {
			WARN.Println("open file header error,", err)
			return nil
		}
		return file
	}
	return nil
}

func ReaderConvert(p *Params, n string, typ reflect.Type) reflect.Value {
	file := getMultipartFile(p, n)
	if file == nil {
		return reflect.Zero(typ)
	}
	return reflect.ValueOf(file)
}

func ByteSliceConvert(p *Params, n string, typ reflect.Type) reflect.Value {
	file := getMultipartFile(p, n)
	if file == nil {
		return reflect.Zero(typ)
	}
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		WARN.Println("read all multipart file  error,", err)
		return reflect.Zero(typ)
	}
	return reflect.ValueOf(bs)
}

func FileConvert(p *Params, n string, typ reflect.Type) reflect.Value {
	file := getMultipartFile(p, n)
	if file == nil {
		return reflect.Zero(typ)
	}
	if osFile, ok := file.(*os.File); ok {
		return reflect.ValueOf(osFile)
	}
	//store temp file
	osFile, err := ioutil.TempFile("", "peony-upload-file")
	if err != nil {
		WARN.Println("create temp file error,", err)
		return reflect.Zero(typ)
	}
	p.tmpFiles = append(p.tmpFiles, osFile)
	_, err = io.Copy(osFile, file)
	if err != nil {
		WARN.Println("save data to temp file error,", err)
		return reflect.Zero(typ)
	}
	_, err = osFile.Seek(0, 0)
	if err != nil {
		WARN.Println("seek to begin of temp file error,", err)
		return reflect.Zero(typ)
	}
	return reflect.ValueOf(osFile)
}

//covert struct
func GetStructConvert(c *Convertors) Convert {
	return func(p *Params, n string, typ reflect.Type) reflect.Value {
		result := reflect.New(typ).Elem()
		fieldValues := make(map[string]reflect.Value)
		for key, _ := range p.Values {
			if !strings.HasPrefix(key, n+".") {
				continue
			}

			suffix := key[len(n)+1:]
			fieldName := nextKey(suffix)
			fieldLen := len(fieldName)
			//convert the field
			if _, ok := fieldValues[fieldName]; !ok {
				fieldValue := result.FieldByName(fieldName)
				if !fieldValue.IsValid() {
					WARN.Println("W: bindStruct: Field not found:", fieldName)
					continue
				}
				if !fieldValue.CanSet() {
					WARN.Println("W: bindStruct: Field not settable:", fieldName)
					continue
				}
				convertVal := c.Convert(p, key[:len(n)+1+fieldLen], fieldValue.Type())
				fieldValue.Set(convertVal)
				fieldValues[fieldName] = convertVal
			}
		}
		return result
	}
}

func GetStructReverseConvert(c *Convertors) ReverseConvert {
	return func(p map[string]string, name string, v interface{}) {
		val := reflect.ValueOf(v)
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			structField := typ.Field(i)
			fieldValue := val.Field(i)
			if structField.PkgPath == "" {
				c.ReverseConvert(p, fmt.Sprintf("%s.%s", name, structField.Name), fieldValue.Interface())
			}
		}
	}
}

func IntReverseConvert(p map[string]string, name string, v interface{}) {
	p[name] = fmt.Sprintf("%d", v)
}

func StringReverseConvert(p map[string]string, name string, v interface{}) {
	p[name] = fmt.Sprintf("%s", v)
}

func FloatReverseConvert(p map[string]string, name string, v interface{}) {
	p[name] = fmt.Sprintf("%f", v)
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

	c.KindConvertors[reflect.Float32] = ValueConvertor(FloatConvert, FloatReverseConvert)
	c.KindConvertors[reflect.Float64] = ValueConvertor(FloatConvert, FloatReverseConvert)

	c.KindConvertors[reflect.String] = ValueConvertor(StringConvert, StringReverseConvert)

	c.KindConvertors[reflect.Struct] = &Convertor{
		GetStructConvert(c),
		GetStructReverseConvert(c),
	}

	c.KindConvertors[reflect.Ptr] = &Convertor{
		func(p *Params, name string, typ reflect.Type) reflect.Value {
			return c.Convert(p, name, typ.Elem()).Addr()
		},
		func(p map[string]string, name string, v interface{}) {
			c.ReverseConvert(p, name, reflect.ValueOf(v).Elem().Interface())
		},
	}

	c.TypeConvertors[reflect.TypeOf((io.Reader)(nil))] = &Convertor{ReaderConvert, nil}
	c.TypeConvertors[reflect.TypeOf((io.ReadWriter)(nil))] = &Convertor{ReaderConvert, nil}
	c.TypeConvertors[reflect.TypeOf((*os.File)(nil))] = &Convertor{FileConvert, nil}
	c.TypeConvertors[reflect.TypeOf([]byte)] = &Convertor{ByteSliceConvert, nil}
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
