package interpreter

import (
	"fmt"

	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/environment"
	"github.com/levpaul/glocks/internal/parser"
)

type LoxFunction struct {
	declaration   *parser.FunctionDeclaration
	closure       *environment.Environment
	isInitializer bool
}

// Call executes a Lox function with the given interpreter and arguments.
// It creates a new environment with the function's closure as the enclosing scope,
// binds the function parameters to the provided argument values,
// and executes the function body.
func (l LoxFunction) Call(i parser.LoxInterpreter, args []domain.Value) (domain.Value, error) {
	env := environment.NewEnvironment(l.closure)
	for idx, p := range l.declaration.Params {
		env.Define(p, args[idx])
	}

	blockErr := i.ExecuteBlock(&parser.Block{Statements: l.declaration.Body}, env)
	if blockErr == nil {
		if l.isInitializer {
			return l.closure.GetAt(0, "this")
		}
		return nil, nil
	}

	if earlyReturn, isEarlyReturn := blockErr.(EarlyReturn); isEarlyReturn {
		if l.isInitializer {
			return l.closure.GetAt(0, "this")
		}
		return earlyReturn.result, nil
	}

	return nil, blockErr
}

// Arity returns the number of parameters a function has.
func (l LoxFunction) Arity() int {
	return len(l.declaration.Params)
}

// String returns a string representation of the function.
func (l LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", l.declaration.Name)
}

func (l LoxFunction) Bind(instance LoxInstance) LoxFunction {
	env := environment.NewEnvironment(l.closure)
	env.Define("this", instance)
	return LoxFunction{
		declaration:   l.declaration,
		closure:       env,
		isInitializer: l.isInitializer,
	}
}
