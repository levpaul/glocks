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
	if err = i.runLine(program); err != nil {
		i.log.With("error", err).
			Errorf("Failed to run program:\n%s\n", program)
		return err
	}

	i.log.Info("Successfully ran program")
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
			if _, isEarlyRet := err.(EarlyReturn); isEarlyRet {
				return fmt.Errorf("Unexpected 'return' expression found. Expected to be within a function")
			}
			return fmt.Errorf("failed to evaluate expression: '%w'", err)
		}
		if i.replMode && result != nil { // only print our statements which evaluate to a Value
			fmt.Println("evaluates to:", result)
		}
	}

	return nil
}

// ExecuteBlock takes a Block and an Environment, executing Block with specific Environment env
func (i *Interpreter) ExecuteBlock(block parser.Block, env *environment.Environment) error {
	originalEnv := i.env
	defer func() { i.env = originalEnv }()
	i.env = env

	return i.VisitBlock(block)
}
