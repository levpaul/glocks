package builtins

import (
	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/parser"
	"time"
)

type Clock struct{}

func (c *Clock) Arity() int {
	return 0
}
func (c *Clock) Call(i parser.LoxInterpreter, args []domain.Value) (domain.Value, error) {
	return float64(time.Now().Unix()), nil
}
