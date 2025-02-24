package environment

import (
	"errors"
	"fmt"

	"github.com/levpaul/glocks/internal/domain"
)

// Environment is a recursive data structure that holds a map of variable names to values
// and a pointer to the enclosing environment. This allows for nested scopes and
// variable shadowing. The Environment is used to store variables and their values
// during execution.
type Environment struct {
	Enclosing *Environment
	Values    map[string]domain.Value
}

// NewEnvironment creates a new Environment with the given enclosing environment.
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

// Get retrieves the value of a variable from the environment. If the variable is not found
// in the current environment, it will search the enclosing environments recursively.
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

// Set sets the value of a variable in the environment. If the variable is not found
// in the current environment, it will search and set in the enclosing environments recursively.
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
