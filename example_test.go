package strconvert_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nahojer/strconvert"
)

func ExampleWithStringifier() {
	// Override how float64 values is formatted.
	strconvert.WithStringifier(func(f float64) (string, error) {
		return strconv.FormatFloat(f, 'E', 6, 64), nil
	})

	// Do the same for pointer float64 values.
	strconvert.WithStringifier(func(f *float64) (string, error) {
		if f == nil {
			return strconv.FormatFloat(0, 'E', 6, 64), nil
		}
		return strconv.FormatFloat(*f, 'E', 6, 64), nil
	})

	// Extend Stringify's type support by registering a stringifier for our
	// custom type.
	type PrefixedValue struct {
		Value  string
		Prefix string
	}
	strconvert.WithStringifier(func(pv PrefixedValue) (string, error) {
		return pv.Prefix + pv.Value, nil
	})
}

func ExampleWithParser() {
	// Extend Parse's type support by registering a parser for our custom type.
	type PrefixedValue struct {
		Value  string
		Prefix string
	}
	strconvert.WithParser(func(s string) (PrefixedValue, error) {
		prefix := "--"
		value, found := strings.CutPrefix(s, prefix)
		if !found {
			return PrefixedValue{}, fmt.Errorf("failed to parse PrefixedValue from %q", s)
		}
		return PrefixedValue{Value: value, Prefix: prefix}, nil
	})
}
