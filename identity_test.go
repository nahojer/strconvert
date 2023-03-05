package strconvert_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nahojer/strconvert"
)

// BinaryStruct implements encoding.TextUnmarshaler and
// encoding.TextMarshaler.
type TextStruct struct {
	Value string
}

func (t *TextStruct) MarshalText() (text []byte, err error) {
	return []byte(t.Value), nil
}

func (t *TextStruct) UnmarshalText(text []byte) error {
	t.Value = string(text)
	return nil
}

// BinaryStruct implements encoding.BinaryUnmarshaler and
// encoding.BinaryMarshaler.
type BinaryStruct struct {
	Value string
}

func (b *BinaryStruct) MarshalBinary() (data []byte, err error) {
	return []byte(b.Value), nil
}

func (b *BinaryStruct) UnmarshalText(data []byte) error {
	b.Value = string(data)
	return nil
}

func testIdentity[T any](t *testing.T, orig T) {
	s, err := strconvert.Stringify(reflect.ValueOf(&orig))
	if err != nil {
		t.Fatalf("Stringify(%v) = \"\", %q;", orig, err)
	}
	var parsed T
	if err := strconvert.Parse(s, reflect.Indirect(reflect.ValueOf(&parsed))); err != nil {
		typ := reflect.TypeOf(parsed)
		t.Fatalf("Parse(%q, <%s>) = %v, %q", s, typ.Kind(), reflect.Zero(typ), err)
	}
	if !cmp.Equal(parsed, orig) {
		t.Errorf("parsed %v, orig %v", parsed, orig)
	}
}

// Fuzz test as many types as possible. Fuzz funcs cannot be a substrings of each other.
// Run fuzzall.sh to execute all fuzz tests.
func Fuzz1_Uint(f *testing.F) {
	f.Add(uint(42))
	f.Fuzz(func(t *testing.T, orig uint) { testIdentity(t, orig) })
}
func Fuzz2_Uint8(f *testing.F) {
	f.Add(uint8(50))
	f.Fuzz(func(t *testing.T, orig uint8) { testIdentity(t, orig) })
}
func Fuzz3_Uint16(f *testing.F) {
	f.Add(uint16(62))
	f.Fuzz(func(t *testing.T, orig uint16) { testIdentity(t, orig) })
}
func Fuzz4_Uint32(f *testing.F) {
	f.Add(uint32(73))
	f.Fuzz(func(t *testing.T, orig uint32) { testIdentity(t, orig) })
}
func Fuzz5_Uint64(f *testing.F) {
	f.Add(uint64(81))
	f.Fuzz(func(t *testing.T, orig uint64) { testIdentity(t, orig) })
}
func Fuzz6_Int(f *testing.F) {
	f.Add(int(19))
	f.Fuzz(func(t *testing.T, orig int) { testIdentity(t, orig) })
}
func Fuzz7_Int8(f *testing.F) {
	f.Add(int8(-20))
	f.Fuzz(func(t *testing.T, orig int8) { testIdentity(t, orig) })
}
func Fuzz8_Int16(f *testing.F) {
	f.Add(int16(7))
	f.Fuzz(func(t *testing.T, orig int16) { testIdentity(t, orig) })
}
func Fuzz9_Int32(f *testing.F) {
	f.Add(int32(2024))
	f.Fuzz(func(t *testing.T, orig int32) { testIdentity(t, orig) })
}
func Fuzz10_Int64(f *testing.F) {
	f.Add(int64(30000))
	f.Fuzz(func(t *testing.T, orig int64) { testIdentity(t, orig) })
}
func Fuzz11_Rune(f *testing.F) {
	f.Add(rune(3))
	f.Fuzz(func(t *testing.T, orig rune) { testIdentity(t, orig) })
}
func Fuzz12_Float32(f *testing.F) {
	f.Add(float32(13124.123123))
	f.Fuzz(func(t *testing.T, orig float32) { testIdentity(t, orig) })
}
func Fuzz13_Float64(f *testing.F) {
	f.Add(float64(00.123))
	f.Fuzz(func(t *testing.T, orig float64) { testIdentity(t, orig) })
}
func Fuzz14_Bool(f *testing.F) {
	f.Add(true)
	f.Fuzz(func(t *testing.T, orig bool) { testIdentity(t, orig) })
}
func Fuzz15_String(f *testing.F) {
	f.Add("whatever banana apple")
	f.Fuzz(func(t *testing.T, orig string) { testIdentity(t, orig) })
}
func Fuzz16_Byte(f *testing.F) {
	f.Add(byte(1))
	f.Fuzz(func(t *testing.T, orig byte) { testIdentity(t, orig) })
}
func Fuzz17_Bytes(f *testing.F) {
	f.Add([]byte("some bytes"))
	f.Fuzz(func(t *testing.T, orig []byte) { testIdentity(t, orig) })
}

func TestComplexIdentity(t *testing.T) {
	testIdentity(t, complex64(49+23i))
	testIdentity(t, complex128(-1000-50i))
}

func TestDurationIdentity(t *testing.T) {
	testIdentity(t, time.Second*36)
	testIdentity(t, time.Minute*22)
	testIdentity(t, time.Microsecond*123123)
}

func TestMapIdentity(t *testing.T) {
	testIdentity(t, map[int]float64{0: 1.2, 1: 3.4, 2: 5.6})
	testIdentity(t, map[string]int{"meaning of life": 42})
}

func TestSliceIdentity(t *testing.T) {
	testIdentity(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	testIdentity(t, []float64{5.6, 122.33, 724.123, -123.11})
}

func TestArrayIdentity(t *testing.T) {
	testIdentity(t, [10]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	testIdentity(t, [30]float64{})
}

func TestTextStructIdentity(t *testing.T) {
	testIdentity(t, TextStruct{Value: "text unmarshaler and marshaler"})
}

func TestBinaryStructIdentity(t *testing.T) {
	testIdentity(t, BinaryStruct{Value: "binary unmarshaler and marshaler"})
}
