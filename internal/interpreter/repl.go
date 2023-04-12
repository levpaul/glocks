package interpreter

import (
	"github.com/chzyer/readline"
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/levpaul/glocks/internal/parser"
	"io"
)

func (i *Interpreter) REPL() error {
	var s *lexer.Scanner
	var p *parser.Parser
	var astPrinter parser.ExprPrinter
	var line string

	rl, err := readline.New("> ")
	if err != nil {
		i.log.With("error", err).Error("Failed to initialize readline library")
		return err
	}
	defer rl.Close()
	rl.CaptureExitSignal()

	for {
		line, err = rl.Readline()
		if err != nil {
			switch err {
			case readline.ErrInterrupt:
				continue
			case io.EOF:
				i.log.Info("EOF detected, exiting...")
				return nil
			}
			i.log.With("error", err).Error("Unexpected error occurred, exiting")
			return err
		}

		switch line {
		case "exit":
			i.log.Info("Exiting glocks repl")
			return nil
		default: // REPL process line
			s = lexer.NewScanner(line, i.log)
			tokens := s.ScanTokens()
			p = parser.NewParser(i.log, tokens)
			expr, err := p.Parse()
			if err != nil {
				i.log.With("error", err).Error("failed to parse line")
				continue
			}
			i.log.Infof("AST repr of input: %s", astPrinter.Print(expr))
		}
	}

	return nil
}
