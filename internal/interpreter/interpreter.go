package interpreter

import (
	"bufio"
	"fmt"
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/levpaul/glocks/internal/parser"
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

func (i *Interpreter) REPL() error {
	reader := bufio.NewScanner(os.Stdin)
	var s *lexer.Scanner
	var p *parser.Parser
	var astPrinter parser.ExprPrinter

	// TODO: Handle Ctrl+D (^D) input, and arrow keys...
	for {
		fmt.Print("> ")
		stopped := reader.Scan()
		if !stopped {
			i.log.With("error", reader.Err()).Error("Failed to read input")
			continue
		}

		s = lexer.NewScanner(reader.Text(), i.log)
		tokens := s.ScanTokens()
		p = parser.NewParser(i.log, tokens)
		expr, err := p.Parse()
		if err != nil {
			i.log.With("error", err).Error("failed to parse line")
			continue
		}

		i.log.Infof("AST repr of input: %s", astPrinter.Print(expr))
	}

	return nil
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
