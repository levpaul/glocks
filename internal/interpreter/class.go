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
	klass LoxClass
}

func (l LoxInstance) String() string {
	return l.klass.Name + "instance"
}
