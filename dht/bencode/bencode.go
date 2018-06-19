package bencode

import (
	"errors"
	"fmt"
	"strconv"
)

type Kind int

const (
	String Kind = 0
	Number      = 1
	Array       = 2
	Object      = 3
)

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'

}

func toBencodeString(in string) string {
	return fmt.Sprintf("%v:%v", len(in), in)
}

type Value struct {
	Kind Kind

	String_ string

	Number int

	Array []*Value

	Object map[string]*Value
}

func (value *Value) GetString() string {
	return value.String_
}

func (value *Value) GetNumber() int {
	return value.Number
}

func (value *Value) GetArray() []*Value {
	return value.Array
}

func (value *Value) GetObject() map[string]*Value {
	return value.Object
}

// convert to json like string
// use https://jsonlint.com/ to prettify more!
func (value *Value) Prettify() string {
	if value.Kind == String {
		return `"` + value.GetString() + `"`
	} else if value.Kind == Number {
		return fmt.Sprintf("%v", value.GetNumber())
	} else if value.Kind == Array {
		prettify := "["
		a := value.GetArray()
		for i := 0; i < len(a); i++ {
			prettify += a[i].Prettify() + ","
		}
		if len(prettify) > 0 && prettify[len(prettify)-1] == ',' {
			prettify = prettify[:len(prettify)-1]
		}
		return prettify + "]"
	} else if value.Kind == Object {
		prettify := "{"
		o := value.GetObject()
		for k, v := range o {
			prettify += fmt.Sprintf(`"%v": %v,`, k, v.Prettify())
		}
		if len(prettify) > 0 && prettify[len(prettify)-1] == ',' {
			prettify = prettify[:len(prettify)-1]
		}
		return prettify + "}"
	} else {
		panic("impossible")
	}
}

func (value *Value) Encode() string {
	if value.Kind == String {
		return toBencodeString(value.GetString())
	} else if value.Kind == Number {
		return fmt.Sprintf("i%ve", value.GetNumber())
	} else if value.Kind == Array {
		prettify := "l"
		a := value.GetArray()
		for i := 0; i < len(a); i++ {
			prettify += a[i].Encode()
		}
		return prettify + "e"
	} else if value.Kind == Object {
		prettify := "d"
		o := value.GetObject()
		for k, v := range o {
			prettify += fmt.Sprintf("%v%v", toBencodeString(k), v.Encode())
		}
		return prettify + "e"
	} else {
		panic("impossible")
	}
}

func Decode(b string) (*Value, error) {
	ctx := Context{b: b}
	return ctx.ParseValue()
}

func NewString(in string) *Value {
	return &Value{Kind: String, String_: in}
}

func NewNumber(in int) *Value {
	return &Value{Kind: Number, Number: in}
}

func NewArray(in interface{}) (*Value, error) {
	newValue := make([]*Value, 0)
	for _, v := range in.([]interface{}) {
		switch v.(type) {
		case int:
			newValue = append(newValue, NewNumber(v.(int)))
		case string:
			newValue = append(newValue, NewString(v.(string)))
		case []interface{}:
			if newArray, err := NewArray(v.([]interface{})); err != nil {
				return nil, err
			} else {
				newValue = append(newValue, newArray)
			}
		case map[string]interface{}:
			if newObject, err := NewObject(v.(map[string]interface{})); err != nil {
				return nil, err
			} else {
				newValue = append(newValue, newObject)
			}
		default:
			return nil, errors.New("only support int, string, slice and map")
		}
	}

	return &Value{Kind: Array, Array: newValue}, nil
}

func NewObject(in interface{}) (*Value, error) {
	newValue := make(map[string]*Value)
	for k, v := range in.(map[string]interface{}) {
		switch v.(type) {
		case int:
			newValue[k] = NewNumber(v.(int))
		case string:
			newValue[k] = NewString(v.(string))
		case []interface{}:
			if newArray, err := NewArray(v.([]interface{})); err != nil {
				return nil, err
			} else {
				newValue[k] = newArray
			}
		case map[string]interface{}:
			if newObject, err := NewObject(v.(map[string]interface{})); err != nil {
				return nil, err
			} else {
				newValue[k] = newObject
			}
		default:
			return nil, errors.New("only support int, string, slice and map")
		}
	}
	return &Value{Kind: Object, Object: newValue}, nil
}

type Context struct {
	b string
}

func (ctx *Context) RemoveACharacter(c byte) error {
	if len(ctx.b) < 1 || ctx.b[0] != c {
		return errors.New("syntax error")
	}
	ctx.b = ctx.b[1:]
	return nil
}

func (ctx *Context) PeekACharacter() (byte, error) {
	if len(ctx.b) < 1 {
		return '0', errors.New("syntax error")
	}
	return ctx.b[0], nil
}

func (ctx *Context) GetString() (string, error) {
	p := 0

	for p < len(ctx.b) && isDigit(ctx.b[p]) {
		p++
	}

	Len, err := strconv.ParseInt(ctx.b[:p], 10, 64)
	if err != nil {
		return "", errors.New("syntax error")
	}

	ctx.b = ctx.b[p:]

	if err = ctx.RemoveACharacter(':'); err != nil {
		return "", errors.New("syntax error")
	}

	if len(ctx.b) < int(Len) {
		return "", errors.New("syntax error")
	}

	str := ctx.b[:Len]
	ctx.b = ctx.b[Len:]
	return str, nil
}

func (ctx *Context) ParseString() (*Value, error) {
	if string_, err := ctx.GetString(); err != nil {
		return nil, err
	} else {
		return &Value{Kind: String, String_: string_}, nil
	}
}

func (ctx *Context) ParseNumber() (*Value, error) {
	if err := ctx.RemoveACharacter('i'); err != nil {
		return nil, errors.New("syntax error")
	}

	p := 0

	for p < len(ctx.b) && isDigit(ctx.b[p]) {
		p++
	}

	number, err := strconv.ParseInt(ctx.b[:p], 10, 64)
	if err != nil {
		return nil, errors.New("syntax error")
	}

	ctx.b = ctx.b[p:]

	if err = ctx.RemoveACharacter('e'); err != nil {
		return nil, errors.New("syntax error")

	}

	return &Value{Kind: Number, Number: int(number)}, nil
}

func (ctx *Context) ParseArray() (*Value, error) {
	if err := ctx.RemoveACharacter('l'); err != nil {
		return nil, errors.New("syntax error")
	}

	value := &Value{Kind: Array}
	for {
		// dispatch
		ele, err := ctx.ParseValue()
		if err != nil {
			return nil, err
		}

		value.Array = append(value.Array, ele)

		// read 'e' represent end of array
		if err := ctx.RemoveACharacter('e'); err == nil {
			break
		}
	}
	return value, nil
}

func (ctx *Context) ParseObject() (*Value, error) {
	if err := ctx.RemoveACharacter('d'); err != nil {
		return nil, errors.New("syntax error")
	}

	value := &Value{Kind: Object}

	for {
		// read key
		key, err := ctx.GetString()
		if err != nil {
			return nil, err
		}

		// dispatch
		attribute, err := ctx.ParseValue()
		if err != nil {
			return nil, err
		}

		// save to map
		if value.Object == nil {
			m := make(map[string]*Value)
			value.Object = m
		}

		value.Object[key] = attribute

		// read 'e' represent end of object
		if err := ctx.RemoveACharacter('e'); err == nil {
			break
		}
	}

	return value, nil
}

func (ctx *Context) ParseValue() (*Value, error) {
	c, err := ctx.PeekACharacter()
	if err != nil {
		return nil, err
	}

	if c >= '0' && c <= '9' {
		return ctx.ParseString()
	}

	switch c {
	case 'i':
		return ctx.ParseNumber()
	case 'l':
		return ctx.ParseArray()
	case 'd':
		return ctx.ParseObject()
	default:
		return nil, errors.New(fmt.Sprintf("error character: %v", c))
	}
}
