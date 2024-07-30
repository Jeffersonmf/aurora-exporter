package util

import (
	"context"
	"fmt"
)

// T can be any type
func Print[T any](s []T) {
	for _, v := range s {
		fmt.Print(v)
	}
}

func SplitAnySlice[T any](s []T) ([]T, []T) {
	mid := len(s) / 2
	return s[:mid], s[mid:]
}

func Flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

func GetLast[T any](s []T) T {
	var myZero T
	if len(s) == 0 {
		return myZero
	}
	return s[len(s)-1]
}

type Stringer interface {
	String() string
}

func Concat[T Stringer](vals []T) string {
	result := ""
	for _, val := range vals {
		// this is where the .String() method
		// is used. That's why we need a more specific
		// constraint instead of the any constraint
		result += val.String()
	}

	return result
}

func AddValueInContext(ctx context.Context, key any, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

func ReadValuefromContext(ctx context.Context, key any) any {
	value := ctx.Value(key)

	switch value {
	case nil:
		value = ""
		fallthrough

	default:
		return value
	}
}

type Predicate[A any] func(A) bool

func Filter[A any](input []A, pred Predicate[A]) []A {
	output := []A{}
	for _, element := range input {
		if pred(element) {
			output = append(output, element)
		}
	}
	return output
}
