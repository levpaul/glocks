package interpreter

import (
	"fmt"
	"github.com/levpaul/glocks/internal/parser"
)

type Environment struct {
	Enclosing *Environment
	Values    map[string]parser.Value
}

func (e *Environment) Define(name string, v parser.Value) {
	e.Values[name] = v
}

func (e *Environment) Get(name string) (parser.Value, error) {
	if val, found := e.Values[name]; found {
		return val, nil
	}

	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}

	return nil, fmt.Errorf("attempted to get variable '%s' but does not exist", name)
}

func (e *Environment) Set(name string, v parser.Value) error {
	if _, found := e.Values[name]; !found {
		if e.Enclosing == nil {
			return fmt.Errorf("attempted to set variable '%s' but does not exist", name)
		}
		return e.Enclosing.Set(name, v)
	}

	e.Values[name] = v
	return nil
}
