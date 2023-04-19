package interpreter

import (
	"fmt"
	"github.com/levpaul/glocks/internal/parser"
)

type Environment struct {
	Values map[string]parser.Value
}

func (e *Environment) Define(name string, v parser.Value) {
	e.Values[name] = v
}

func (e *Environment) Get(name string) (parser.Value, error) {
	if val, found := e.Values[name]; found {
		return val, nil
	}
	return nil, fmt.Errorf("attempted to get variable '%s' but does not exist", name)
}
