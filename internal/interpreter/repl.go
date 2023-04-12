package interpreter

import (
	"errors"
	"fmt"
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/levpaul/glocks/internal/parser"
	"golang.org/x/term"
	"os"
)

const maxHistory = 20
const errorExitVal = "== errorOccurredExitNow =="

var unexpectedErrorReason string

func (i *Interpreter) REPL() error {
	var s *lexer.Scanner
	var p *parser.Parser
	var astPrinter parser.ExprPrinter

	history := make([]string, maxHistory)

	for {
		fmt.Print("> ")

		line, err := i.processLine()
		if err != nil {
			i.log.With("error", err).Error("Unexpected error occurred, exiting")
			return err
		}
		history = append([]string{line}, history[:maxHistory-1]...)

		switch line {
		case errorExitVal:
			return errors.New("unexpected error")
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

func (i *Interpreter) processLine() (string, error) {
	oldStdinState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		i.log.With("error", err).Error("Failed to make stdin raw")
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldStdinState)
	oldStdoutState, err := term.MakeRaw(int(os.Stdout.Fd()))
	if err != nil {
		i.log.With("error", err).Error("Failed to make stdout raw")
		return "", err
	}
	defer term.Restore(int(os.Stdout.Fd()), oldStdoutState)

	b := make([]byte, 1)
	var res string
	for {
		_, err = os.Stdin.Read(b)
		if err != nil {
			return "", err
		}
		switch c := b[0]; c {
		case '\r', '\n':
			os.Stdout.WriteString("\n")
			return res, nil
		case 0x03: // Ctrl+C
			return "", nil
		case 0x04: // Ctrl+D
			return "exit", nil
		default:
			//fmt.Printf("\nthe char %q was hit\n", string(b[0]))
			fmt.Print(string(c))
			res += string(c)
		}
	}
	return res, nil
}
