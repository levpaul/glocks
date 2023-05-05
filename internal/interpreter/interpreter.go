package interpreter

import (
	"fmt"
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/levpaul/glocks/internal/parser"
	"go.uber.org/zap"
	"io"
	"strings"
)

func New(log *zap.SugaredLogger) *Interpreter {
	return &Interpreter{
		log:        log,
		s:          nil,
		p:          nil,
		astPrinter: parser.ExprPrinter{},
		replMode:   false,
		env:        &Environment{Values: map[string]parser.Value{}},
	}
}

type Interpreter struct {
	log         *zap.SugaredLogger
	s           *lexer.Scanner
	p           *parser.Parser
	astPrinter  parser.ExprPrinter
	replMode    bool
	printOutput io.Writer
	env         *Environment
	evalRes     any
}

func (i *Interpreter) VisitBlock(b parser.Block) error {
	oldEnv := i.env
	i.env = &Environment{
		Enclosing: oldEnv,
		Values:    map[string]parser.Value{},
	}
	defer func() {
		i.env = oldEnv
	}()

	for _, stmt := range b.Statements {
		result, err := i.Evaluate(stmt)
		if err != nil {
			return err
		}
		if i.replMode && result != nil { // only print our statements which evaluate to a Value
			fmt.Println("evaluates to:", result)
		}
	}
	return nil
}

func (i *Interpreter) Run(program string) error {
	var err error
	lineNumber := 1
	for _, line := range strings.Split(program, "\n") {
		if err = i.runLine(line); err != nil {
			i.log.With("error", err).
				Errorf("Failed to run line number %d; line:\n%s\n", lineNumber, line)
			return err
		}

		lineNumber++
	}

	i.log.Infof("Successfully ran %d lines of code'\n", lineNumber)
	return nil
}

func (i *Interpreter) runLine(line string) error {
	// TODO: add reset func to scanner/parser
	i.s = lexer.NewScanner(line, i.log)
	tokens := i.s.ScanTokens()
	i.p = parser.NewParser(i.log, tokens)
	stmts, err := i.p.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse line, err='%w'", err)
	}
	for _, stmt := range stmts {
		if i.replMode && i.log.Level() <= zap.DebugLevel {
			i.log.Debugf("AST repr of input: %s", i.astPrinter.Print(stmt))
		}

		result, err := i.Evaluate(stmt)
		if err != nil {
			return fmt.Errorf("failed to evaluate expression: '%w'", err)
		}
		if i.replMode && result != nil { // only print our statements which evaluate to a Value
			fmt.Println("evaluates to:", result)
		}
	}

	return nil
}
