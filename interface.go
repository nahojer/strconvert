package strconvert

import (
	"encoding"
	"reflect"
)

func textUnmarshaler(field reflect.Value) (t encoding.TextUnmarshaler) {
	interfaceFrom(field, func(v any, ok *bool) {
		t, *ok = v.(encoding.TextUnmarshaler)
	})
	return t
}

func textMarshaler(field reflect.Value) (t encoding.TextMarshaler) {
	interfaceFrom(field, func(v any, ok *bool) {
		t, *ok = v.(encoding.TextMarshaler)
	})
	return t
}

func binaryUnmarshaler(field reflect.Value) (b encoding.BinaryUnmarshaler) {
	interfaceFrom(field, func(v any, ok *bool) {
		b, *ok = v.(encoding.BinaryUnmarshaler)
	})
	return b
}

func binaryMarshaler(field reflect.Value) (t encoding.BinaryMarshaler) {
	interfaceFrom(field, func(v any, ok *bool) {
		t, *ok = v.(encoding.BinaryMarshaler)
	})
	return t
}

func interfaceFrom(field reflect.Value, fn func(any, *bool)) {
	if !field.CanInterface() {
		return
	}

	var ok bool
	fn(field.Interface(), &ok)
	if !ok && field.CanAddr() {
		fn(field.Addr().Interface(), &ok)
	}
}
