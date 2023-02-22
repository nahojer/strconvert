package strconvert

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Stringify converts v to a string.
//
// The following types are supported:
//   - Types registered using the WithStringifier option
//   - Types implementing [encoding.TextMmarshaler]
//   - Types implementing [encoding.BinaryMarshaler]
//   - ~string
//   - ~int, ~int8, ~int16, ~int32, ~int64
//   - [time.Duration]
//   - ~uint, ~uint8, ~uint16, ~uint32, ~uint64
//   - ~float32, ~float64
//   - ~complex64, ~complex128
//   - ~bool
//   - slices, arrays and maps of any of the aforementioned types
//
// Stringify errors for any unsupported type. More types may be be supported in
// the future.
//
// Float and complex types are formatted using the format byte 'f'
// and precision -1. See the documentation for [strconv.FormatFloat].
// Override this behaviour by registering custom stringifiers.
//
// Map values are sorted before formatted into the final string representation,
// ensuring consistent and predictable output.
func Stringify[V any](v V, optFns ...func(*Options)) (string, error) {
	opts := buildOptions(optFns)
	if opts.savedErr != nil {
		return "", opts.savedErr
	}
	return stringify(reflect.ValueOf(v), &opts)
}

func stringify(v reflect.Value, opts *Options) (string, error) {
	typ := v.Type()

	if fn, ok := opts.funcs[typ]; ok {
		out := fn.Call([]reflect.Value{v})
		err, _ := out[1].Interface().(error)
		if err != nil {
			return "", err
		}
		return out[0].String(), nil
	}

	if m := textMarshaler(v); m != nil {
		b, err := m.MarshalText()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	if m := binaryMarshaler(v); m != nil {
		b, err := m.MarshalBinary()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}

	switch typ.Kind() {
	case reflect.String:
		return v.String(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			return time.Duration(v.Int()).String(), nil
		}
		return strconv.FormatInt(v.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil

	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil

	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil

	case reflect.Complex64:
		return strconv.FormatComplex(v.Complex(), 'f', -1, 64), nil

	case reflect.Complex128:
		return strconv.FormatComplex(v.Complex(), 'f', -1, 128), nil

	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil

	case reflect.Slice, reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			return string(v.Bytes()), nil
		}

		strSlice := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			s, err := stringify(v.Index(i), opts)
			if err != nil {
				return "", fmt.Errorf("error stringifying slice item of index %d: %w", i, err)
			}
			strSlice[i] = s
		}
		return strings.Join(strSlice, string(opts.elemSep)), nil

	case reflect.Map:
		strSlice := make([]string, v.Len())
		for i, k := range v.MapKeys() {
			sk, err := stringify(k, opts)
			if err != nil {
				return "", fmt.Errorf("error stringifying key %v of map: %w", k, err)
			}

			v := v.MapIndex(k)
			sv, err := stringify(v, opts)
			if err != nil {
				return "", fmt.Errorf("error stringifying map value with key %s: %w", sk, err)
			}

			strSlice[i] = sk + string(opts.keySep) + sv
		}
		// Sort to get predictable output.
		sort.Strings(strSlice)

		return strings.Join(strSlice, string(opts.elemSep)), nil
	}

	return "", fmt.Errorf("unsupported field type %s", typ.Kind().String())
}
