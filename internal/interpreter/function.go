package interpreter

import (
	"fmt"
	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/environment"
	"github.com/levpaul/glocks/internal/parser"
)

type LoxFunction struct {
	declaration parser.FunctionDeclaration
	env         environment.Environment
}

func (l LoxFunction) Call(i parser.LoxInterpreter, args []domain.Value) (domain.Value, error) {
	for idx, p := range l.declaration.Params {
		l.env.Define(p, args[idx])
	}

	block, ok := l.declaration.Body.(parser.Block)
	if !ok {
		return nil, fmt.Errorf("Expected function call to have associated block, but none found: '%s'", l.declaration.Body)
	}

	blockErr := i.ExecuteBlock(block, &l.env)
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
