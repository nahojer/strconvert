package strconvert_test

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/nahojer/strconvert"
)

func testParse[T any](t *testing.T, in string, want T, optFns ...func(*strconvert.Options)) {
	var got T
	kind := reflect.TypeOf(got).Kind()
	if err := strconvert.Parse(in, reflect.Indirect(reflect.ValueOf(&got)), optFns...); err != nil {
		t.Fatalf("Parse(%q, <%s>) = 0, %q; want %v, nil", in, kind, err, want)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("Parse(%q, <%s>) = %v, nil; want %v, nil", in, kind, got, want)
	}
}

func testBadParser[T any](t *testing.T, parser func(string) (T, error)) {
	var got T
	err := strconvert.Parse("", reflect.Indirect(reflect.ValueOf(&got)), strconvert.WithParser(parser))
	if err == nil {
		t.Fatalf("Parse(\"\", %v, WithParser[%s](...)) = nil; want error", got, reflect.TypeOf(got))
	}
	if !strings.Contains(err.Error(), "not a valid parser return type") {
		t.Errorf("unexpected error %q", err)
	}
}

func TestParse(t *testing.T) {
	t.Run("builtin", func(t *testing.T) {
		testParse(t, "0", byte(0))
		testParse(t, "1", uint(1))
		testParse(t, "2", uint8(2))
		testParse(t, "3", uint16(3))
		testParse(t, "4", uint32(4))
		testParse(t, "5", uint64(5))
		testParse(t, "6", int(6))
		testParse(t, "7", int8(7))
		testParse(t, "8", int16(8))
		testParse(t, "9", int32(9))
		testParse(t, "10", int64(10))
		testParse(t, "11", rune(11))
		testParse(t, "3.14159", float32(3.14159))
		testParse(t, "2.71828", float64(2.71828))
		testParse(t, "(3+2i)", complex64(3+2i))
		testParse(t, "5+20.3i", complex128(5+20.3i))
		testParse(t, "false", false)
		testParse(t, "true", true)
		testParse(t, "whatever", "whatever")
		testParse(t, "whatever", []byte("whatever"))
		testParse(t, "5h0m0s", time.Hour*5)
		testParse(t, "key1:value1;key2:value2", map[string]string{"key1": "value1", "key2": "value2"})
		testParse(t, "item1;item2", []string{"item1", "item2"})
		testParse(t, "item1;item2", [10]string{"item1", "item2"})
	})

	t.Run("text unmarshaler", func(t *testing.T) {
		testParse(t, "some text", TextStruct{"some text"})
	})

	t.Run("binary unmarshaler", func(t *testing.T) {
		testParse(t, "some data", TextStruct{"some data"})
	})

	t.Run("custom parser", func(t *testing.T) {
		type Custom struct {
			Value string
		}
		testParse(t, "my custom value", Custom{"test: my custom value"}, strconvert.WithParser(func(s string) (Custom, error) {
			return Custom{"test: " + s}, nil
		}))

		testParse(t, "3,14159", 3.14159, strconvert.WithParser(func(s string) (float64, error) {
			return strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64)
		}))
	})

	t.Run("addressable struct field", func(t *testing.T) {
		strct := &struct{ Got float64 }{}
		kind := reflect.TypeOf(strct.Got).Kind()
		in := "42.37"
		want := 42.37
		if err := strconvert.Parse(in, reflect.Indirect(reflect.ValueOf(strct)).FieldByName("Got")); err != nil {
			t.Fatalf("Parse(%q, <%s>) = 0, %q; want %v, nil", in, kind, err, want)
		}
		if !cmp.Equal(strct.Got, want) {
			t.Errorf("Parse(%q, <%s>) = %v, nil; want %v, nil", in, kind, strct.Got, want)
		}
	})

	t.Run("bad parser types", func(t *testing.T) {
		testBadParser(t, func(string) (any, error) { return "", nil })
		testBadParser(t, func(string) (fmt.Stringer, error) { var s fmt.Stringer; return s, nil })
		testBadParser(t, func(string) (io.Reader, error) { var r io.Reader; return r, nil })
		testBadParser(t, func(string) (chan float32, error) { return make(chan float32), nil })
		testBadParser(t, func(string) (chan string, error) { return make(chan string), nil })
		testBadParser(t, func(string) (map[string]int, error) { return make(map[string]int), nil })
		testBadParser(t, func(string) (map[float64]any, error) { return make(map[float64]any), nil })
		testBadParser(t, func(string) (func(), error) { return func() {}, nil })
		testBadParser(t, func(string) (uintptr, error) { return 0, nil })
		testBadParser(t, func(string) (unsafe.Pointer, error) { var p unsafe.Pointer; return p, nil })
	})

	t.Run("invalid parse error", func(t *testing.T) {
		// Not addressable.
		err := strconvert.Parse("123", reflect.ValueOf(123))
		if err == nil {
			t.Fatalf("Parse(123, <int>) = 0, nil; want error")
		}
		if !errors.Is(err, strconvert.ErrInvalidParseArgument) {
			t.Errorf("unexpected error %q", err)
		}

		// Nil.
		err = strconvert.Parse("123", reflect.ValueOf(nil))
		if err == nil {
			t.Fatalf("Parse(123, nil) = 0, nil; want error")
		}
		if !errors.Is(err, strconvert.ErrInvalidParseArgument) {
			t.Errorf("unexpected error %q", err)
		}
	})
}
