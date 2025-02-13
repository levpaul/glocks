package parser

import (
	"fmt"
	"testing"

	"github.com/levpaul/glocks/internal/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRawTokensSuccess(t *testing.T) {
	type testData struct {
		inputTokens    []*lexer.Token
		expectedOutput string
	}

	td := []testData{
		{
			inputTokens: []*lexer.Token{
				{
					Type:    lexer.NUMBER,
					Lexeme:  "1",
					Literal: 1,
				},
				{
					Type:   lexer.BANG_EQUAL,
					Lexeme: "!=",
				},
				{
					Type:    lexer.NUMBER,
					Lexeme:  "2",
					Literal: 2,
				},
			},
			expectedOutput: "(!= 1 2)",
		},
		{
			inputTokens: []*lexer.Token{
				{
					Type:   lexer.LEFT_PAREN,
					Lexeme: "(",
				},
				{
					Type:    lexer.NUMBER,
					Lexeme:  "1",
					Literal: 1,
				},
				{
					Type:   lexer.BANG_EQUAL,
					Lexeme: "!=",
				},
				{
					Type:    lexer.NUMBER,
					Lexeme:  "2",
					Literal: 2,
				},
				{
					Type:   lexer.LESS_EQUAL,
					Lexeme: "<=",
				},
				{
					Type:    lexer.NUMBER,
					Lexeme:  "3",
					Literal: 3,
				},
				{
					Type:   lexer.LESS,
					Lexeme: "<",
				},
				{
					Type:    lexer.NUMBER,
					Lexeme:  "4",
					Literal: 4,
				},
				{
					Type:   lexer.RIGHT_PAREN,
					Lexeme: ")",
				},
				{
					Type:   lexer.PLUS,
					Lexeme: "+",
				},
				{
					Type:    lexer.NUMBER,
					Lexeme:  "43",
					Literal: 43,
				},
				{
					Type:   lexer.MINUS,
					Lexeme: "-",
				},
				{
					Type:    lexer.STRING,
					Lexeme:  "hehehe",
					Literal: "hehehe",
				},
				{
					Type:   lexer.STAR,
					Lexeme: "*",
				},
				{
					Type:   lexer.TRUE,
					Lexeme: "true",
				},
			},
			expectedOutput: "(+ (+ (group (!= 1 (<= (<= 2 3) 4))) 43) (* hehehe true))",
		},
	}
	printer := ExprPrinter{}
	for _, test := range td {
		p := NewParser(zap.S(), test.inputTokens)
		fmt.Println("parser made", test.inputTokens)
		res, err := p.expressionStmt()
		fmt.Println("res err", res, err)
		assert.Nil(t, err, "Unexpected parser error")
		assert.Equal(t, test.expectedOutput, printer.Print(res), "Failed test with input tokens %v", test.inputTokens)
	}
}

func TestScannedTokensSuccess(t *testing.T) {
	type testData struct {
		inputExpression string
		expectedOutput  string
	}

	td := []testData{
		{
			inputExpression: "1 != 2",
			expectedOutput:  "(!= 1 2)",
		},
		{
			inputExpression: `"hehehe"`,
			expectedOutput:  `hehehe`,
		},
		{
			inputExpression: `"hehehe" + 42`,
			expectedOutput:  `(+ hehehe 42)`,
		},
		{
			inputExpression: `4 / 5`,
			expectedOutput:  `(/ 4 5)`,
		},
		{
			inputExpression: `(1 != (3 < 4))`,
			expectedOutput:  `(group (!= 1 (group (< 3 4))))`,
		},
		{
			inputExpression: `(1 +1) * 43 - "hehehe" * true`,
			expectedOutput:  `(- (* (group (+ 1 1)) 43) (* hehehe true))`,
		},
	}

	printer := ExprPrinter{}
	for _, test := range td {
		scanner := lexer.NewScanner(test.inputExpression, zap.S())
		scannedTokens := scanner.ScanTokens()
		require.NotEmpty(t, scannedTokens, "Unexpectedly found not tokens after scanning input Expression: '%s'", test.inputExpression)
		// Remove last token, as scanner adds EOF token to end
		p := NewParser(zap.S(), scannedTokens[:len(scannedTokens)-1])
		res, err := p.expressionStmt()
		assert.Nil(t, err, "Expected no errors from parser")
		assert.Equal(t, test.expectedOutput, printer.Print(res), "Failed test with input Expression %s", test.inputExpression)
	}
}
