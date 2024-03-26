package interpreter

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"strings"
)

func New(log *zap.SugaredLogger) *Interpreter {
	return &Interpreter{log: log}
}

type Interpreter struct {
	log *zap.SugaredLogger
}

func (i *Interpreter) RunFile(file string) error {

	program, err := os.ReadFile(file)
	if err != nil {
		i.log.With("error", err).Errorf("Failed to read file '%s' from disk\n", file)
		return err
	}

	lineNumber := 1
	for _, line := range strings.Split(string(program), "\n") {
		if err = i.runLine(line); err != nil {
			i.log.With("error", err).
				Errorf("Failed to run line number %d from program '%s'; line:\n%s\n", lineNumber, file, line)
			return err
		}

		lineNumber++
	}

	i.log.Infof("Successfully ran %d lines of code from program '%s'\n", lineNumber, file)
	return nil
}

func (i *Interpreter) runLine(line string) error {
	fmt.Printf("Pretending to run '%s'... DONE!\n", line)
	return nil
}
