package interpreter

import (
	"fmt"

	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/parser"
)

type LoxClass struct {
	Name       string
	Methods    map[string]LoxFunction
	SuperClass domain.Value
}

func (l LoxClass) Call(i parser.LoxInterpreter, args []domain.Value) (domain.Value, error) {
	instance := LoxInstance{
		klass:  l,
		fields: map[string]domain.Value{},
	}
	initializer, exists := l.Methods["init"]
	if exists {
		if _, err := initializer.Bind(instance).Call(i, args); err != nil {
			return nil, err
		}
	}

	return instance, nil
}

func (l LoxClass) String() string {
	return fmt.Sprintf("<class %s>", l.Name)
}

func (l LoxClass) Arity() int {
	initializer, exists := l.Methods["init"]
	if exists {
		return initializer.Arity()
	}
	return 0
}

type LoxInstance struct {
	// functions []LoxFunction
	klass  LoxClass
	fields map[string]domain.Value
}

func (l LoxInstance) String() string {
	return l.klass.Name + " instance"
}

func (l LoxInstance) Get(name string) (domain.Value, error) {
	if val, exists := l.fields[name]; exists {
		return val, nil
	}

	if method, exists := l.klass.Methods[name]; exists {
		return method.Bind(l), nil
	}

	return nil, fmt.Errorf("Undefined property '%s' on instance of class '%s'", name, l.klass.Name)
}

func (l LoxInstance) Set(name string, value domain.Value) {
	l.fields[name] = value
}
