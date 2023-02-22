package strconvert

import (
	"errors"
	"fmt"
	"reflect"
)

// Options for modifying and/or extending the behaviour of [Stringify] and
// [Parse].
type Options struct {
	elemSep, keySep rune
	funcs           map[reflect.Type]reflect.Value
	savedErr        error
}

// WithParser registers a custom parser function for a concrete type.
//
// [Parse] errors if return type T is an interface, channel, function,
// map, uintptr, or unsafe.Pointer. More types may be be supported in
// the future.
func WithParser[T any](fn func(s string) (T, error)) func(*Options) {
	return func(o *Options) {
		if o.funcs == nil {
			o.funcs = make(map[reflect.Type]reflect.Value)
		}
		v := reflect.ValueOf(fn)
		retType := v.Type().Out(0)
		switch retType.Kind() {
		default:
			o.funcs[retType] = v
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Uintptr, reflect.UnsafePointer, reflect.Map:
			o.savedErr = errors.Join(o.savedErr, fmt.Errorf("%s is not a valid parser return type", retType.Kind()))
		}
	}
}

// WithStringifier registers a custom stringifier function for a concrete type.
//
// [Stringify] errors if argument type T is an interface, channel, function,
// map, uintptr, or unsafe.Pointer. More types may be be supported in the
// future.
func WithStringifier[T any](fn func(v T) (string, error)) func(*Options) {
	return func(o *Options) {
		if o.funcs == nil {
			o.funcs = make(map[reflect.Type]reflect.Value)
		}
		v := reflect.ValueOf(fn)
		argType := v.Type().In(0)
		switch argType.Kind() {
		default:
			o.funcs[argType] = v
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Uintptr, reflect.UnsafePointer, reflect.Map:
			o.savedErr = errors.Join(o.savedErr, fmt.Errorf("%s is not a valid stringifier argument type", argType.Kind()))
		}
	}
}

// WithElementSeparator overrides the default element separator used for
// parsing/stringifying slices, arrays and maps.
func WithElementSeparator(r rune) func(*Options) {
	return func(o *Options) {
		o.elemSep = r
	}
}

// WithKeySeparator override the default key separator used for
// parsing/stringifying key, value pairs in maps.
func WithKeySeparator(r rune) func(*Options) {
	return func(o *Options) {
		o.keySep = r
	}
}

func buildOptions(optFns []func(*Options)) Options {
	opts := Options{
		elemSep: ';',
		keySep:  ':',
	}
	for _, fn := range optFns {
		fn(&opts)
	}
	return opts
}
