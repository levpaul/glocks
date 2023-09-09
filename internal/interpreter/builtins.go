package interpreter

import (
	"github.com/levpaul/glocks/internal/parser"
	"time"
)

type clock struct{}

func (c *clock) Arity() int {
	return 0
}
func (c *clock) Call(i parser.LoxInterpreter, args []parser.Value) (parser.Value, error) {
	return float64(time.Now().Unix()), nil
}
