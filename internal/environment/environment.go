package environment

import (
	"fmt"
	"github.com/levpaul/glocks/internal/domain"
)

type Environment struct {
	Enclosing *Environment
	Values    map[string]domain.Value
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		Enclosing: enclosing,
		Values:    map[string]domain.Value{},
	}
}

func (e *Environment) Define(name string, v domain.Value) {
	e.Values[name] = v
}

func (e *Environment) Get(name string) (domain.Value, error) {
	if val, found := e.Values[name]; found {
		return val, nil
	}

	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}

	return nil, fmt.Errorf("attempted to get variable '%s' but does not exist", name)
}

func (e *Environment) Set(name string, v domain.Value) error {
	if _, found := e.Values[name]; !found {
		if e.Enclosing == nil {
			return fmt.Errorf("attempted to set variable '%s' but does not exist", name)
		}
		return e.Enclosing.Set(name, v)
	}

	e.Values[name] = v
	return nil
}
