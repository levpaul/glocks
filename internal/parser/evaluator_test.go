package parser

import (
	"fmt"
	"reflect"
	"testing"
)

func BenchmarkValidateThenCast(b *testing.B) {
	var v1, v2, res any
	v1 = 45.
	v2 = 46.
	e := Evaluator{}
	for i := 0; i < b.N; i++ {
		err := e.validateBothNumber(v1, v2)
		if err == nil {
			res = v1.(float64) - v2.(float64)
		}
	}
	nop(res)
}

func nop(a any) {
	return
}

func BenchmarkValidateNumber(b *testing.B) {
	v1 := 45
	v2 := "45"
	e := Evaluator{}
	for i := 0; i < b.N; i++ {
		e.validateBothNumber(v1, v2)
	}
}

func BenchmarkValidateNumberReflect(b *testing.B) {
	v1 := 45
	v2 := "45"
	e := Evaluator{}
	for i := 0; i < b.N; i++ {
		e.validateBothNumberReflect(v1, v2)
	}
}

func (e *Evaluator) validateBothNumberReflect(left Value, right Value) error {
	if reflect.TypeOf(left).Kind() != reflect.Float64 {
		return fmt.Errorf("%v is not a number", left)
	}

	if reflect.TypeOf(right).Kind() != reflect.Float64 {
		return fmt.Errorf("%v is not a number", right)
	}
	return nil
}
