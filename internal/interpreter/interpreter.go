package interpreter

import (
	"fmt"
	"go.uber.org/zap"
)

func New(log *zap.SugaredLogger) *Interpreter {
	return &Interpreter{log: log}
}

type Interpreter struct {
	log *zap.SugaredLogger
}

func (i *Interpreter) REPL() error {
	i.log.Error("REPL not implemented yet!")
	return fmt.Errorf("REPL not implemented")
}

func (i *Interpreter) RunFile(file string) error {
	i.log.Error("Run program not implemented yet!")
	return fmt.Errorf("run file not implemented")
}
