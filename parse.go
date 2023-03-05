package strconvert

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidParseArgument describes an in inalid argument parsed to Parse.
// (The argument to Parse must be addressable as defined by the reflect package.)
var ErrInvalidParseArgument = errors.New("")

// Parse parses the string s and stores the result in the underlying Go value
// pointed to by v. If v is addressable (as defined by the reflect package),
// Parse returns an ErrInvalidParseArgument.
//
// Parse is the inverse operation of calling Stringify with the same options and
// complementing custom parsers/stringifiers. The following types for the underlying
// Go value of v are supported:
//
//   - Types registered using the WithParser option
//   - Types implementing [encoding.TextUnmarshaler]
//   - Types implementing [encoding.BinaryUnmarshaler]
//   - ~string
//   - ~int, ~int8, ~int16, ~int32, ~int64
//   - [time.Duration]
//   - ~uint, ~uint8, ~uint16, ~uint32, ~uint64
//   - ~float32, ~float64
//   - ~complex64, ~complex128
//   - ~bool
//   - Any pointer to the above types
//   - slices, arrays and maps of any of the above types
//
// Parse errors for any unsupported type. More types may be be supported in
// the future.
func Parse(s string, v reflect.Value, optFns ...func(*Options)) error {
	opts := buildOptions(optFns)
	if opts.savedErr != nil {
		return opts.savedErr
	}
	if !v.CanAddr() {
		return ErrInvalidParseArgument
	}
	return parse(s, v, &opts)
}

func parse(s string, v reflect.Value, opts *Options) error {
	typ := v.Type()

	if fn, ok := opts.funcs[typ]; ok {
		out := fn.Call([]reflect.Value{reflect.ValueOf(s)})
		if err, _ := out[1].Interface().(error); err != nil {
			return err
		}
		v.Set(out[0])
		return nil
	}

	if t := textUnmarshaler(v); t != nil {
		return t.UnmarshalText([]byte(s))
	}

	if b := binaryUnmarshaler(v); b != nil {
		return b.UnmarshalBinary([]byte(s))
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if v.IsNil() {
			v.Set(reflect.New(typ))
		}
		v = v.Elem()
	}

	switch typ.Kind() {
	case reflect.String:
		v.SetString(s)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			d, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			v.SetInt(int64(d))
		} else {
			i, err := strconv.ParseInt(s, 0, typ.Bits())
			if err != nil {
				return err
			}
			v.SetInt(i)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 0, typ.Bits())
		if err != nil {
			return err
		}
		v.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, typ.Bits())
		if err != nil {
			return err
		}
		v.SetFloat(f)

	case reflect.Complex64:
		c, err := strconv.ParseComplex(s, 64)
		if err != nil {
			return err
		}
		v.SetComplex(c)

	case reflect.Complex128:
		c, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return err
		}
		v.SetComplex(c)

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)

	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			v.Set(reflect.ValueOf([]byte(s)))
		} else {
			elems := strings.Split(s, string(opts.elemSep))
			sl := reflect.MakeSlice(typ, len(elems), len(elems))
			for i, val := range elems {
				if err := parse(val, sl.Index(i), opts); err != nil {
					return err
				}
			}
			v.Set(sl)
		}

	case reflect.Array:
		elems := strings.Split(s, string(opts.elemSep))
		if len(elems) > v.Cap() {
			return fmt.Errorf("number of elements (%d) exceeds array capacity (%d)", len(elems), v.Cap())
		}
		for i := 0; i < v.Cap(); i++ {
			if i >= len(elems) {
				break
			}
			if err := parse(elems[i], v.Index(i), opts); err != nil {
				return err
			}
		}

	case reflect.Map:
		m := reflect.MakeMap(typ)
		if len(strings.TrimSpace(s)) != 0 {
			pairs := strings.Split(s, string(opts.elemSep))
			for _, pair := range pairs {
				kvpair := strings.Split(pair, string(opts.keySep))
				if len(kvpair) != 2 {
					return fmt.Errorf("invalid map item: %q", pair)
				}
				k := reflect.New(typ.Key()).Elem()
				if err := parse(kvpair[0], k, opts); err != nil {
					return err
				}
				v := reflect.New(typ.Elem()).Elem()
				if err := parse(kvpair[1], v, opts); err != nil {
					return err
				}
				m.SetMapIndex(k, v)
			}
		}
		v.Set(m)
	}

	return nil
}
