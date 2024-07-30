package util

import (
	"reflect"
	"strings"
	"unicode/utf8"
)

// SumInts adds together the values of m.
func SumInts(m map[string]int64) int64 {
	var s int64
	for _, v := range m {
		s += v
	}
	return s
}
func FilterInvalidUTF8(input string) string {
	// Iterate over input string byte by byte
	result := make([]rune, 0, len(input))
	for len(input) > 0 {
		r, size := utf8.DecodeRuneInString(input)
		if r != utf8.RuneError || size > 1 {
			// Valid UTF-8 rune or multibyte sequence
			result = append(result, r)
		}
		input = input[size:]
	}

	filteredResult := strings.ReplaceAll(string(result), "\x00", "")
	return filteredResult
}

// SumFloats adds together the values of m.
func SumFloats(m map[string]float64) float64 {
	var s float64
	for _, v := range m {
		s += v
	}
	return s
}

func GetTagByField(entity any, fieldName string) string {
	field, ok := reflect.TypeOf(entity).Elem().FieldByName(fieldName)
	if !ok {
		return ""
	}

	return string(field.Tag.Get("trino"))
}

type Color int

var ColorEnum = struct {
	Red   Color
	Blue  Color
	Green Color
}{
	Red:   0,
	Blue:  1,
	Green: 2,
}

func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}
