package environment

import (
	"errors"
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

func (e *Environment) ancestor(distance int) (*Environment, error) {
	currEnv := e
	for i := 0; i < distance; i++ {
		if currEnv.Enclosing == nil {
			return nil, errors.New("could not find scope of variable, mismatch in declaration distances")
		}
		currEnv = currEnv.Enclosing
	}
	return currEnv, nil
}

func (e *Environment) GetAt(distance int, name string) (domain.Value, error) {
	targetEnv, err := e.ancestor(distance)
	if err != nil {
		return nil, err
	}
	return targetEnv.Get(name)
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

func (e *Environment) SetAt(distance int, name string, v domain.Value) error {
	targetEnv, err := e.ancestor(distance)
	if err != nil {
		return err
	}
	return targetEnv.Set(name, v)
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

func (e *Environment) Clone() Environment {
	newEnv := Environment{
		Enclosing: e.Enclosing,
		Values:    map[string]domain.Value{},
	}

	for k, v := range e.Values {
		newEnv.Values[k] = v
	}
	return newEnv
}
