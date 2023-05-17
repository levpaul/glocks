package parser

import (
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrintExpression(t *testing.T) {
	// Create -123 * (45.67) as an Expression (tree) and print it
	expr := Binary{
		Left: Unary{
			Operator: &lexer.Token{
				Type:    lexer.MINUS,
				Lexeme:  "-",
				Literal: nil,
				Line:    1,
			},
			Right: Literal{Value: 123},
		},
		Operator: &lexer.Token{
			Type:    lexer.STAR,
			Lexeme:  "*",
			Literal: nil,
			Line:    1,
		},
		Right: Grouping{Expression: Literal{Value: 45.67}},
	}

	walker := ExprPrinter{}
	assert.Equal(t, "(* (- 123) (group 45.67))", walker.Print(expr))
}
