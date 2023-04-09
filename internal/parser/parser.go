package parser

import (
	"errors"
	"fmt"
	"github.com/levpaul/glocks/internal/lexer"
)

type Parser struct {
	current int
	Tokens  []*lexer.Token
}

/* Expression Grammar
expression     → equality ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary | primary ;
primary        → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" ;
*/

// Start parsing with lowest precedence part of Expression and recursively decend to highest precedence Expr
// This is a Recursive Decent Parser
// expression → equality ;
func (p *Parser) expression() Expr {
	return p.equality()
}

// equality → comparison ( ( "!=" | "==" ) comparison )* ;
func (p *Parser) equality() Expr {
	res := p.comparison()
	for cur := p.Tokens[p.current]; p.match(lexer.BANG_EQUAL, lexer.EQUAL_EQUAL); {
		res = Binary{
			Left:     res,
			Right:    p.comparison(),
			Operator: cur,
		}
	}
	return res
}

// comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
func (p *Parser) comparison() Expr {
	res := p.term()
	for cur := p.Tokens[p.current]; p.match(lexer.LESS, lexer.LESS_EQUAL, lexer.GREATER, lexer.GREATER_EQUAL); {
		res = Binary{
			Left:     res,
			Right:    p.term(),
			Operator: cur,
		}
	}
	return res
}

// term → factor ( ( "-" | "+" ) factor )* ;
func (p *Parser) term() Expr {
	res := p.factor()
	for cur := p.Tokens[p.current]; p.match(lexer.MINUS, lexer.PLUS); {
		res = Binary{
			Left:     res,
			Right:    p.factor(),
			Operator: cur,
		}
	}
	return res
}

// factor → unary ( ( "/" | "*" ) unary )* ;
func (p *Parser) factor() Expr {
	res := p.unary()
	for cur := p.Tokens[p.current]; p.match(lexer.SLASH, lexer.STAR); {
		res = Binary{
			Left:     res,
			Right:    p.unary(),
			Operator: cur,
		}
	}
	return res
}

// unary → ( "!" | "-" ) unary | primary ;
func (p *Parser) unary() Expr {
	if cur := p.Tokens[p.current]; p.match(lexer.BANG, lexer.MINUS) {
		return Unary{
			Operator: cur,
			Right:    p.unary(),
		}
	}
	return p.primary()
}

// primary → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" ;
func (p *Parser) primary() Expr {
	cur := p.Tokens[p.current]

	// Deal with only token which expects further tokens, otherwise advance and switch
	if cur.Type == lexer.LEFT_PAREN {
		if p.advance() != nil {
			fmt.Println("BAD ERROR - cannot end program with '('")
			return nil
		}
		inner := p.expression()
		if !p.match(lexer.RIGHT_PAREN) {
			fmt.Println("Unexpected token, expected ')'")
			return nil
		}
		return Grouping{inner}
	}

	p.advance() // An error here doesn't matter?
	switch cur.Type {
	case lexer.NUMBER, lexer.STRING:
		return Literal{Value: cur.Literal}

	case lexer.TRUE:
		return Literal{Value: true}
	case lexer.FALSE:
		return Literal{Value: false}
	case lexer.NIL:
		return Literal{Value: nil}

	default:
		fmt.Println("WTFFF - unexpected place to end up - parsing primary with", cur)
		return nil
	}
}

func (p *Parser) advance() error {
	if p.current+1 >= len(p.Tokens) {
		fmt.Println("CANNOT ADVANCE, cur=", p.current)
		return errors.New("tried to advance parser, but already at the last token")
	}
	p.current++
	return nil
}

func (p *Parser) getCurrent() *lexer.Token {
	if p.current < 0 || p.current >= len(p.Tokens) {
		return nil
	}
	return p.Tokens[p.current]
}

func (p *Parser) match(t ...lexer.TokenType) bool {
	for _, tt := range t {
		cur := p.getCurrent()
		if cur == nil {
			fmt.Println("Tried to match where there is no further input")
			return false
		}
		if cur.Type == tt {
			fmt.Println("match found at ", cur, "with ", tt)
			p.advance()
			return true
		}
	}
	return false
}
