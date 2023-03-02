package strconvert_test

import (
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

func testStringify[T any](t *testing.T, in T, want string, optFns ...func(*strconvert.Options)) {
	got, err := strconvert.Stringify(reflect.ValueOf(in), optFns...)
	if err != nil {
		t.Fatalf("strconvert.Stringify(%v) = \"\", %q; want %q, nil", in, err, want)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("strconvert.Stringify(%v) = %q, nil; want %q, nil", in, got, want)
	}
}

func testBadStringifier[T any](t *testing.T, stringifier func(v T) (string, error)) {
	var zero T
	got, err := strconvert.Stringify(reflect.ValueOf(zero), strconvert.WithStringifier(stringifier))
	if err == nil {
		t.Fatalf("strconvert.Stringify(%v, strconvert.WithStringifier[%s](...)) = %q, nil; want error", zero, reflect.TypeOf(zero), got)
	}
	if !strings.Contains(err.Error(), "not a valid stringifier argument type") {
		t.Errorf("unexpected error %q", err)
	}
}

func TestStringify(t *testing.T) {
	t.Run("builtin", func(t *testing.T) {
		testStringify(t, byte(0), "0")
		testStringify(t, uint(1), "1")
		testStringify(t, uint8(2), "2")
		testStringify(t, uint16(3), "3")
		testStringify(t, uint32(4), "4")
		testStringify(t, uint64(5), "5")
		testStringify(t, int(6), "6")
		testStringify(t, int8(7), "7")
		testStringify(t, int16(8), "8")
		testStringify(t, int32(9), "9")
		testStringify(t, int64(10), "10")
		testStringify(t, rune(11), "11")
		testStringify(t, float32(3.14159), "3.14159")
		testStringify(t, float64(2.71828), "2.71828")
		testStringify(t, complex64(3+2i), "(3+2i)")
		testStringify(t, complex128(5+20.3i), "(5+20.3i)")
		testStringify(t, false, "false")
		testStringify(t, true, "true")
		testStringify(t, "whatever", "whatever")
		testStringify(t, []byte("whatever"), "whatever")
		testStringify(t, time.Hour*5, "5h0m0s")
		testStringify(t, map[string]string{"key1": "value1", "key2": "value2"}, "key1:value1;key2:value2")
		testStringify(t, []string{"item1", "item2"}, "item1;item2")
		testStringify(t, [10]string{"item1", "item2"}, "item1;item2;;;;;;;;")
	})

	t.Run("text marshaler", func(t *testing.T) {
		testStringify(t, &TextStruct{"some text"}, "some text")
	})

	t.Run("binary marshaler", func(t *testing.T) {
		testStringify(t, &BinaryStruct{"some data"}, "some data")
	})

	t.Run("custom stringifier", func(t *testing.T) {
		type Custom struct {
			Value string
		}
		testStringify(t, Custom{"my custom value"}, "test: my custom value", strconvert.WithStringifier(func(c Custom) (string, error) {
			return "test: " + c.Value, nil
		}))

		testStringify(t, 3.14159, "3.142E+00", strconvert.WithStringifier(func(f float64) (string, error) {
			return strconv.FormatFloat(f, 'E', 3, 64), nil
		}))

		testStringify(t, 3.14159, "3.14159", strconvert.WithStringifier(func(f *float64) (string, error) {
			return "float64 is not the same type as float64", nil
		}))
	})

	t.Run("bad stringifier types", func(t *testing.T) {
		testBadStringifier(t, func(any) (string, error) { return "", nil })
		testBadStringifier(t, func(io.Writer) (string, error) { return "", nil })
		testBadStringifier(t, func(fmt.Scanner) (string, error) { return "", nil })
		testBadStringifier(t, func(chan int) (string, error) { return "", nil })
		testBadStringifier(t, func(chan string) (string, error) { return "", nil })
		testBadStringifier(t, func(map[string]int) (string, error) { return "", nil })
		testBadStringifier(t, func(map[float64]any) (string, error) { return "", nil })
		testBadStringifier(t, func(func()) (string, error) { return "", nil })
		testBadStringifier(t, func(uintptr) (string, error) { return "", nil })
		testBadStringifier(t, func(unsafe.Pointer) (string, error) { return "", nil })
	})
}
