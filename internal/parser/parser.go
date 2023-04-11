package parser

import (
	"errors"
	"github.com/levpaul/glocks/internal/lexer"
	"go.uber.org/zap"
)

/*
	Expression Grammar

expression     → equality ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary | primary ;
primary        → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" ;
*/
type Parser struct {
	current int
	tokens  []*lexer.Token
	log     *zap.SugaredLogger
}

func NewParser(log *zap.SugaredLogger, tokens []*lexer.Token) *Parser {
	if tokens == nil {
		tokens = []*lexer.Token{}
	}
	return &Parser{
		current: 0,
		tokens:  tokens,
		log:     log,
	}
}

func (p *Parser) Parse() (Expr, error) {
	expr, err := p.expression()
	if err != nil {
		p.synchronize()
		return nil, err
	}

	return expr, nil
}

// Start parsing with lowest precedence part of Expression and recursively descend to highest precedence Expr
// This is a Recursive Decent Parser
// expression → equality ;
func (p *Parser) expression() (Expr, error) {
	return p.equality()
}

// equality → comparison ( ( "!=" | "==" ) comparison )* ;
func (p *Parser) equality() (Expr, error) {
	res, err := p.comparison()
	if err != nil {
		return nil, err
	}
	for cur := p.tokens[p.current]; p.match(lexer.BANG_EQUAL, lexer.EQUAL_EQUAL); {
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		res = Binary{
			Left:     res,
			Right:    right,
			Operator: cur,
		}
	}

	return res, nil
}

// comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
func (p *Parser) comparison() (Expr, error) {
	res, err := p.term()
	if err != nil {
		return nil, err
	}
	for cur := p.tokens[p.current]; p.match(lexer.LESS, lexer.LESS_EQUAL, lexer.GREATER, lexer.GREATER_EQUAL); {
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		res = Binary{
			Left:     res,
			Right:    right,
			Operator: cur,
		}
	}
	return res, nil
}

// term → factor ( ( "-" | "+" ) factor )* ;
func (p *Parser) term() (Expr, error) {
	res, err := p.factor()
	if err != nil {
		return nil, err
	}
	for cur := p.tokens[p.current]; p.match(lexer.MINUS, lexer.PLUS); {
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		res = Binary{
			Left:     res,
			Right:    right,
			Operator: cur,
		}
	}
	return res, nil
}

// factor → unary ( ( "/" | "*" ) unary )* ;
func (p *Parser) factor() (Expr, error) {
	res, err := p.unary()
	if err != nil {
		return nil, err
	}
	for cur := p.tokens[p.current]; p.match(lexer.SLASH, lexer.STAR); {
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		res = Binary{
			Left:     res,
			Right:    right,
			Operator: cur,
		}
	}
	return res, nil
}

// unary → ( "!" | "-" ) unary | primary ;
func (p *Parser) unary() (Expr, error) {
	if cur := p.tokens[p.current]; p.match(lexer.BANG, lexer.MINUS) {
		right, err := p.primary()
		if err != nil {
			return nil, err
		}
		return Unary{
			Operator: cur,
			Right:    right,
		}, nil
	}
	return p.primary()
}

// primary → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" ;
func (p *Parser) primary() (Expr, error) {
	cur := p.tokens[p.current]

	// Deal with only token which expects further tokens, otherwise advance and switch
	if cur.Type == lexer.LEFT_PAREN {
		if p.advance() != nil {
			return nil, cur.GenerateTokenError("BAD ERROR - cannot end program with '('")
		}
		inner, err := p.expression()
		if err != nil {
			return nil, err
		}
		if !p.match(lexer.RIGHT_PAREN) {
			return nil, cur.GenerateTokenError("unexpected token, expected ')'")
		}
		return Grouping{inner}, nil
	}

	_ = p.advance()
	switch cur.Type {
	case lexer.NUMBER, lexer.STRING:
		return Literal{Value: cur.Literal}, nil

	case lexer.TRUE:
		return Literal{Value: true}, nil
	case lexer.FALSE:
		return Literal{Value: false}, nil
	case lexer.NIL:
		return Literal{Value: nil}, nil

	default:
		return nil, cur.GenerateTokenError("Could not parse expression, expected a primary expression")
	}
}

func (p *Parser) advance() error {
	if p.current+1 >= len(p.tokens) {
		return errors.New("tried to advance parser, but already at the last token")
	}
	p.current++
	return nil
}

func (p *Parser) getCurrent() *lexer.Token {
	if p.current < 0 || p.current >= len(p.tokens) {
		return nil
	}
	return p.tokens[p.current]
}

func (p *Parser) getPrevious() *lexer.Token {
	if p.current <= 0 || p.current > len(p.tokens) {
		return nil
	}
	return p.tokens[p.current-1]
}

func (p *Parser) match(t ...lexer.TokenType) bool {
	for _, tt := range t {
		cur := p.getCurrent()
		if cur == nil {
			p.log.Error("Tried to match where there is no further input")
			return false
		}
		if cur.Type == tt {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) synchronize() {
	for p.advance() == nil {
		if p.getPrevious().Type == lexer.SEMICOLON {
			return
		}

		switch p.getCurrent().Type {
		case lexer.CLASS, lexer.FUN, lexer.VAR, lexer.FOR, lexer.IF, lexer.WHILE, lexer.PRINT, lexer.RETURN:
			return
		}
	}

	p.log.Debug("failed to synchronize, reached end of tokens")
}
