package interpreter

import (
	"fmt"
	"github.com/levpaul/glocks/internal/builtins"
	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/environment"
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
	globals     *environment.Environment
	env         *environment.Environment
	evalRes     any
}

func newGlobalEnv() *environment.Environment {
	g := &environment.Environment{Values: map[string]domain.Value{}}

	g.Define("clock", &builtins.Clock{})

	return g
}

func (i *Interpreter) GetEnvironment() *environment.Environment {
	return i.env
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

func (i *Interpreter) ExecuteBlock(block parser.Block, env *environment.Environment) error {
	originalEnv := i.env
	defer func() { i.env = originalEnv }()
	i.env = env

	return i.VisitBlock(block)
}

type LoxFunction struct {
	declaration parser.FunctionDeclaration
}

func (l LoxFunction) Call(i parser.LoxInterpreter, args []domain.Value) (domain.Value, error) {
	env := environment.NewEnvironment(i.GetEnvironment())

	for i, p := range l.declaration.Params {
		env.Define(p, args[i])
	}

	block, ok := l.declaration.Body.(parser.Block)
	if !ok {
		return nil, fmt.Errorf("Expected function call to have associated block, but none found: '%s'", l.declaration.Body)
	}

	return nil, i.ExecuteBlock(block, env)
}

func (l LoxFunction) Arity() int {
	return len(l.declaration.Params)
}

//type LoxCallable interface {
//	Arity() int
//	Call(i LoxInterpreter, args []domain.Value) (domain.Value, error)
//}
