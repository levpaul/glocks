package interpreter

import (
	"testing"
)

func BenchmarkValidateThenCast(b *testing.B) {
	var v1, v2, res any
	v1 = 45.
	v2 = 46.
	e := Interpreter{}
	for i := 0; i < b.N; i++ {
		err := e.validateBothNumber(v1, v2)
		if err == nil {
			res = v1.(float64) - v2.(float64)
		}
	}
	nop(res)
}

func nop(a any) {}

func BenchmarkValidateNumber(b *testing.B) {
	v1 := 45
	v2 := "45"
	e := Interpreter{}
	for i := 0; i < b.N; i++ {
		_ = e.validateBothNumber(v1, v2)
	}
}
