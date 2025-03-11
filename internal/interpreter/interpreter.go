package interpreter

import (
	"fmt"

	"github.com/levpaul/glocks/internal/builtins"
	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/environment"
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/levpaul/glocks/internal/parser"
	"github.com/levpaul/glocks/internal/resolver"
	"go.uber.org/zap"
)

// New creates a new Interpreter for Lox
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

// Interpreter is the main struct for the Lox interpreter, it is self-contained and
// can be used to run a Lox program.
type Interpreter struct {
	log        *zap.SugaredLogger
	s          *lexer.Scanner
	p          *parser.Parser
	r          *resolver.Resolver
	astPrinter parser.ExprPrinter
	replMode   bool
	globals    *environment.Environment
	env        *environment.Environment
	evalRes    any
}

func newGlobalEnv() *environment.Environment {
	g := &environment.Environment{Values: map[string]domain.Value{}}

	g.Define("clock", &builtins.Clock{})

	return g
}

// GetEnvironment returns the current environment of the interpreter
func (i *Interpreter) GetEnvironment() *environment.Environment {
	return i.env
}

// Run executes a Lox program.
func (i *Interpreter) Run(program string) error {
	var err error
	if err = i.run(program); err != nil {
		i.log.With("error", err).
			Errorf("Failed to run program:\n%s\n", program)
		return err
	}

	i.log.Info("Successfully ran program")
	return nil
}

// run executes Lox code. It splits the code into tokens, parses the tokens into an AST,
// resolves variable names to their scope, and then evaluates the AST.
func (i *Interpreter) run(code string) error {
	// Run a lexer on the line of code to tokenize it
	i.s = lexer.NewScanner(code, i.log)
	tokens := i.s.ScanTokens()

	// Run a parser on the tokens to parse them into an AST
	i.p = parser.NewParser(i.log, tokens)
	stmts, err := i.p.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse line, err='%w'", err)
	}

	// Invoke resolver on the AST to resolve variable names to their scope
	i.r = &resolver.Resolver{Scopes: []resolver.Scope{}}
	err = i.r.ResolveNodes(stmts)
	if err != nil {
		return fmt.Errorf("static analysis [resolver] FAILURE, err='%w'", err)
	}

	// Execute each statement in the program, via AST traversal of the parsed statements
	for _, stmt := range stmts {
		if i.replMode && i.log.Level() <= zap.DebugLevel {
			i.log.Debugf("AST repr of input: %s", i.astPrinter.Print(stmt))
		}

		result, err := i.Evaluate(stmt)
		if err != nil {
			if _, isEarlyRet := err.(EarlyReturn); isEarlyRet {
				return fmt.Errorf("unexpected 'return' expression found. Expected to be within a function")
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
func (i *Interpreter) ExecuteBlock(block *parser.Block, env *environment.Environment) error {
	oldEnv := i.env
	i.env = env
	defer func() { i.env = oldEnv }()
	return i.VisitBlock(block)
}

func (i *Interpreter) lookUpVariable(name string, node parser.Node) (domain.Value, error) {
	if distance, err := i.r.GetLocal(node); err != nil {
		return i.env.GetAt(distance, name)
	}

	return i.globals.Get(name)
}
