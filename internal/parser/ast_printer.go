package parser

import (
	"fmt"
	"strings"
)

type ExprPrinter struct{}

// Print walks through an expression and prints it in a Lisp like syntax
func (a *ExprPrinter) Print(expr Expr) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case Binary:
		return a.parenthesize(e.Operator.Lexeme, e.Left, e.Right)
	case Grouping:
		return a.parenthesize("group", e.Expression)
	case Literal:
		if e.Value == nil {
			return "nil"
		}
		return fmt.Sprintf("%+v", e.Value)
	case Unary:
		return a.parenthesize(e.Operator.Lexeme, e.Right)
	default:
		return fmt.Sprintf("<unsupported expr: raw=%+v>", expr)
	}
	return ""
}

func (a *ExprPrinter) parenthesize(name string, exprs ...Expr) string {
	builder := strings.Builder{}
	builder.WriteString("(")
	builder.WriteString(name)

	for _, e := range exprs {
		builder.WriteString(" ")
		builder.WriteString(a.Print(e))
	}

	builder.WriteString(")")
	return builder.String()
}
