package parser

import (
	"errors"
	"fmt"
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

// Parse starts parsing with lowest precedence part of Expression and recursively descend to highest precedence Expr
// This is a Recursive Decent Parser
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

func (p *Parser) Parse() ([]Stmt, error) {
	stmts := []Stmt{}
	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			p.synchronize()
			return nil, err // REPL only?
		}

		stmts = append(stmts, stmt)
	}
	return stmts, nil
}

func (p *Parser) declaration() (s Stmt, err error) {
	if p.match(lexer.VAR) {
		return p.varDeclaration()
	}
	return p.statement()
}

func (p *Parser) varDeclaration() (s Stmt, err error) {
	name, err := p.consume(lexer.IDENTIFIER)
	if err != nil {
		return nil, fmt.Errorf("expected an identifier after 'var'; err='%w'", err)
	}

	var initializer Expr
	if p.match(lexer.EQUAL) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(lexer.SEMICOLON)
	if err != nil {
		return nil, errors.New("expected semi-colon after var declaration")
	}

	return &VarStmt{
		Name:        name.Lexeme,
		Initializer: initializer,
	}, nil
}

func (p *Parser) statement() (s Stmt, err error) {
	cur := p.tokens[p.current]
	if p.match(lexer.PRINT) {
		var arg Expr
		arg, err = p.expression()
		if err != nil {
			return
		}
		s = PrintStmt{Arg: arg}
	} else if s, err = p.expression(); err != nil { // Expression Statement
		return nil, err
	}

	if !p.match(lexer.SEMICOLON) {
		return nil, cur.GenerateTokenError("Expected ; after statement")
	}
	return
}

func (p *Parser) expression() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}
	return expr, nil
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

// primary → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" | IDENTIFIER ;
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

	case lexer.IDENTIFIER:
		return Variable{TokenName: cur.Lexeme}, nil

	default:
		return nil, cur.GenerateTokenError("Could not parse expression, expected a primary expression")
	}
}

func (p *Parser) advance() error {
	p.current++
	if p.current >= len(p.tokens) {
		return errors.New("parser already at end of input when advance() was called")
	}
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

func (p *Parser) consume(t lexer.TokenType) (*lexer.Token, error) {
	cur := p.getCurrent()
	if cur == nil {
		return nil, fmt.Errorf("tried to consume token but could not get current")
	}

	if cur.Type == t {
		p.advance()
		return cur, nil
	}

	return nil, fmt.Errorf("tried to consume token of type %v but current is %v", t, cur)
}

func (p *Parser) isAtEnd() bool {
	if p.current >= len(p.tokens) {
		return true
	}
	return p.tokens[p.current].Type == lexer.EOF
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
