package interpreter

import (
	"fmt"

	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/environment"
	"github.com/levpaul/glocks/internal/parser"
)

type LoxFunction struct {
	declaration *parser.FunctionDeclaration
	closure     *environment.Environment
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
		return nil, nil
	}

	if earlyReturn, isEarlyReturn := blockErr.(EarlyReturn); isEarlyReturn {
		return earlyReturn.result, nil
	}

	return nil, blockErr
}

func (l LoxFunction) Arity() int {
	return len(l.declaration.Params)
}

func (l LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", l.declaration.Name)
}
