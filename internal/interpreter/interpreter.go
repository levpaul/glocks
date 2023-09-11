package interpreter

import (
	"errors"
	"fmt"
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/levpaul/glocks/internal/parser"
	"go.uber.org/zap"
	"io"
	"strings"
)

func New(log *zap.SugaredLogger) *Interpreter {
	globals := newGlobalEnv()
	return &Interpreter{
		log:        log,
		s:          nil,
		p:          nil,
		astPrinter: parser.ExprPrinter{},
		replMode:   false,
		globals:    globals,
		env:        globals, // Set initial env to Global
	}
}

type Interpreter struct {
	log         *zap.SugaredLogger
	s           *lexer.Scanner
	p           *parser.Parser
	astPrinter  parser.ExprPrinter
	replMode    bool
	printOutput io.Writer
	globals     *Environment
	env         *Environment
	evalRes     any
}

func (i *Interpreter) VisitFunctionDeclaration(f parser.FunctionDeclaration) error {
	// TODO: impl - store func in current scope
	return errors.New("unimplemented thingamawhat")
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
