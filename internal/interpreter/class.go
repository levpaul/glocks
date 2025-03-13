package interpreter

import (
	"fmt"

	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/parser"
)

type LoxClass struct {
	Name string
}

func (l LoxClass) Call(i parser.LoxInterpreter, args []domain.Value) (domain.Value, error) {
	instance := LoxInstance{
		klass: l,
	}
	return instance, nil
}

func (l LoxClass) String() string {
	return fmt.Sprintf("<class %s>", l.Name)
}

func (l LoxClass) Arity() int {
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
	return nil, fmt.Errorf("Undefined property '%s' on instance of class '%s'", name, l.klass.Name)
}
