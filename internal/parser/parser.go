package parser

import (
	"errors"
	"fmt"

	"github.com/levpaul/glocks/internal/lexer"
	"go.uber.org/zap"
)

// Parse starts parsing with lowest precedence part of Expression and recursively descend to highest precedence Node
// This is a Recursive Decent Parser - such that precedence is handled by the order of function calls, structuring the
// Abstract Syntax Tree (AST) in a way that is easy to evaluate by the interpreter
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

func (p *Parser) Parse() ([]Node, error) {
	var stmts []Node
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

// declaration    → funDecl
// | varDecl
// | statement
// | classDecl ;
func (p *Parser) declaration() (s Node, err error) {
	if p.match(lexer.CLASS) {
		return p.classDeclaration()
	}
	if p.match(lexer.FUN) {
		return p.funcDeclaration("function")
	}
	if p.match(lexer.VAR) {
		return p.varDeclaration()
	}
	return p.statement()
}

// classDecl → "class" IDENTIFIER "{" function* "}" ;
func (p *Parser) classDeclaration() (Node, error) {
	name, err := p.consume(lexer.IDENTIFIER)
	if err != nil {
		return nil, fmt.Errorf("expected class name")
	}

	_, err = p.consume(lexer.LEFT_BRACE)
	if err != nil {
		return nil, fmt.Errorf("expected a '{' after class name; err=%w", err)
	}

	var methods []Node
	for !p.peekMatch(lexer.RIGHT_BRACE) && !p.isAtEnd() {
		m, err := p.funcDeclaration("method")
		if err != nil {
			return nil, fmt.Errorf("error with function declaration within class; err=%w", err)
		}
		methods = append(methods, m)
	}

	if p.isAtEnd() {
		return nil, fmt.Errorf("expected a '}' after class declaration")
	}

	_, err = p.consume(lexer.RIGHT_BRACE)
	if err != nil {
		return nil, fmt.Errorf("expected a '}' after a class body")
	}
	return &ClassDeclaration{
		Name:    name.Lexeme,
		Methods: methods,
	}, nil
}

// function       → IDENTIFIER "(" parameters? ")" block ;
func (p *Parser) funcDeclaration(kind string) (s Node, err error) {
	name, err := p.consume(lexer.IDENTIFIER)
	if err != nil {
		return nil, fmt.Errorf("expected %s name", kind)
	}

	_, err = p.consume(lexer.LEFT_PAREN)
	if err != nil {
		return nil, fmt.Errorf("expected a '(' after function identifier; err=%w", err)
	}

	var params []string
	if !p.match(lexer.RIGHT_PAREN) {
		for {
			if len(params) >= 255 {
				return nil, errors.New("can't have more than 255 parameters")
			}

			param, paramErr := p.consume(lexer.IDENTIFIER)
			if paramErr != nil {
				return nil, paramErr
			}
			params = append(params, param.Lexeme)
			if p.match(lexer.COMMA) {
				continue
			}
			if p.match(lexer.RIGHT_PAREN) {
				break
			} else {
				return nil, errors.New("Expected closing ')' after parameter list")
			}
		}
	}

	_, err = p.consume(lexer.LEFT_BRACE)
	if err != nil {
		return nil, fmt.Errorf("expected a '{' before function body; err=%w", err)
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	bodyInf, ok := body.(*Block)
	if !ok {
		return nil, errors.New("expected a block following function, did not get one")
	}

	return &FunctionDeclaration{
		Name:   name.Lexeme,
		Params: params,
		Body:   bodyInf.Statements,
	}, nil
}

func (p *Parser) varDeclaration() (s Node, err error) {
	name, err := p.consume(lexer.IDENTIFIER)
	if err != nil {
		return nil, fmt.Errorf("expected an identifier after 'var'; err='%w'", err)
	}

	var initializer Node
	if p.match(lexer.EQUAL) {
		initializer, err = p.expressionStmt()
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

// statement → exprStmt
// | forStmt
// | ifStmt
// | printStmt
// | returnStmt
// | whileStmt
// | block
func (p *Parser) statement() (s Node, err error) {
	startToken := p.tokens[p.current]
	// Take care to fall through if case requires a semi-colon
	switch startToken.Type {
	case lexer.LEFT_BRACE:
		_ = p.advance()
		return p.block()
	case lexer.PRINT:
		_ = p.advance()
		var arg Node
		arg, err = p.expressionStmt()
		if err != nil {
			return
		}
		s = &PrintStmt{Arg: arg}
	case lexer.WHILE:
		_ = p.advance()
		return p.whileStatement()
	case lexer.RETURN:
		_ = p.advance()
		s, err = p.returnStatement()
	case lexer.FOR:
		_ = p.advance()
		return p.forStatement()
	case lexer.IF:
		_ = p.advance()
		return p.ifStatement()
	default:
		// Default case is an Expression Statement
		if s, err = p.expressionStmt(); err != nil {
			return nil, err
		}
	}

	if err == nil && !p.match(lexer.SEMICOLON) { // exprStmt, print + return expect semi-colons
		return nil, startToken.GenerateTokenError("Expected ; after Statement")
	}
	return
}

// forStmt        → "for" "(" ( varDecl | exprStmt | ";" )
// expression? ";"
// expression? ")" statement ;
func (p *Parser) forStatement() (Node, error) {
	var err error
	if _, err = p.consume(lexer.LEFT_PAREN); err != nil {
		return nil, err
	}

	var initializer Node
	if !p.match(lexer.SEMICOLON) {
		if p.match(lexer.VAR) {
			initializer, err = p.varDeclaration()
		} else {
			initializer, err = p.expressionStmt()
		}
		if err != nil {
			return nil, err
		}
	}

	var condition Node
	if !p.match(lexer.SEMICOLON) {
		condition, err = p.expressionStmt()
		if err != nil {
			return nil, err
		}

		if !p.match(lexer.SEMICOLON) {
			return nil, errors.New("Expect ';' after loop condition.")
		}
	}

	var increment Node
	if !p.match(lexer.RIGHT_PAREN) {
		increment, err = p.expressionStmt()
		if err != nil {
			return nil, err
		}
		if !p.match(lexer.RIGHT_PAREN) {
			return nil, errors.New("Expect ')' after loop increment.")
		}
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		body = &Block{Statements: []Node{
			body,
			increment,
		}}
	}

	if condition == nil {
		condition = &Literal{Value: true}
	}
	loop := &WhileStmt{
		Body:       body,
		Expression: condition,
	}

	if initializer == nil {
		return loop, nil
	}

	return &Block{Statements: []Node{initializer, loop}}, nil
}

// whileStmt → "while" "(" expression ")" statement ;
func (p *Parser) whileStatement() (Node, error) {
	var err error
	if _, err = p.consume(lexer.LEFT_PAREN); err != nil {
		return nil, err
	}

	expr, err := p.expressionStmt()
	if err != nil {
		return nil, err
	}

	if _, err = p.consume(lexer.RIGHT_PAREN); err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return &WhileStmt{
		Expression: expr,
		Body:       body,
	}, nil
}

// returnStmt → "return" expression? ";"
func (p *Parser) returnStatement() (Node, error) {
	if p.peekMatch(lexer.SEMICOLON) {
		return &ReturnStmt{}, nil
	}
	expr, err := p.expressionStmt()
	if err != nil {
		return nil, err
	}

	return &ReturnStmt{Expression: expr}, nil
}

// ifStmt → "if" "(" expressionStmt ")" statement ( "else" statement )? ;
func (p *Parser) ifStatement() (Node, error) {
	var err error
	ifStmt := &IfStmt{}
	if !p.match(lexer.LEFT_PAREN) {
		return nil, p.getPrevious().GenerateTokenError("Expected open paren after 'if' Statement")
	}
	ifStmt.Expression, err = p.expressionStmt()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.RIGHT_PAREN) {
		return nil, p.getPrevious().GenerateTokenError("Expected closed paren after 'if' Statement Expression")
	}

	ifStmt.Statement, err = p.statement()
	if err != nil {
		return nil, err
	}

	if p.match(lexer.ELSE) {
		ifStmt.ElseStatement, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return ifStmt, nil
}

// block → "{" declaration* "}" ;
func (p *Parser) block() (Node, error) {
	var nodes []Node
	open := p.getPrevious()
	for !p.match(lexer.RIGHT_BRACE) {
		if p.isAtEnd() {
			return nil, open.GenerateTokenError("Reached end of file, expected closing brace")
		}
		n, err := p.declaration()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	return &Block{Statements: nodes}, nil
}

// expressionStmt -> assignment
func (p *Parser) expressionStmt() (Node, error) {
	return p.assignment()
}

// assignment -> IDENTIFIER = assignment
// _____________ | equality
func (p *Parser) assignment() (Node, error) {
	expr, err := p.logicalConjunction()
	if err != nil {
		return nil, err
	}

	exprToken := p.getPrevious()

	if !p.match(lexer.EQUAL) {
		return expr, nil
	}

	v, ok := expr.(*Variable)
	if !ok {
		return nil, exprToken.GenerateTokenError("Expected variable for assignment but did not find")
	}

	rhs, err := p.assignment()
	if err != nil {
		return nil, err
	}

	return &Assignment{
		TokenName: v.TokenName,
		Value:     rhs,
	}, nil
}

// logicalConjunction parses out "and" or "or" operators, using the same precedence for each - this
// is opposed to C like precedence where "and" has a higher precedence than "or"
func (p *Parser) logicalConjunction() (Node, error) {
	left, err := p.equality()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.AND, lexer.OR) {
		return left, nil
	}

	conj := &LogicalConjuction{
		Left: left,
		And:  p.getPrevious().Type == lexer.AND,
	}

	right, err := p.logicalConjunction()
	if err != nil {
		return nil, err
	}

	conj.Right = right
	return conj, nil
}

// equality → comparison ( ( "!=" | "==" ) comparison )* ;
func (p *Parser) equality() (Node, error) {
	res, err := p.comparison()
	if err != nil {
		return nil, err
	}
	for cur := p.tokens[p.current]; p.match(lexer.BANG_EQUAL, lexer.EQUAL_EQUAL); {
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		res = &Binary{
			Left:     res,
			Right:    right,
			Operator: cur,
		}
	}

	return res, nil
}

// comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
func (p *Parser) comparison() (Node, error) {
	res, err := p.term()
	if err != nil {
		return nil, err
	}
	for cur := p.tokens[p.current]; p.match(lexer.LESS, lexer.LESS_EQUAL, lexer.GREATER, lexer.GREATER_EQUAL); {
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		res = &Binary{
			Left:     res,
			Right:    right,
			Operator: cur,
		}
	}
	return res, nil
}

// term → factor ( ( "-" | "+" ) factor )* ;
func (p *Parser) term() (Node, error) {
	res, err := p.factor()
	if err != nil {
		return nil, err
	}
	for cur := p.tokens[p.current]; p.match(lexer.MINUS, lexer.PLUS); {
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		res = &Binary{
			Left:     res,
			Right:    right,
			Operator: cur,
		}
	}
	return res, nil
}

// factor → unary ( ( "/" | "*" ) unary )* ;
func (p *Parser) factor() (Node, error) {
	res, err := p.unary()
	if err != nil {
		return nil, err
	}
	for cur := p.tokens[p.current]; p.match(lexer.SLASH, lexer.STAR); {
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		res = &Binary{
			Left:     res,
			Right:    right,
			Operator: cur,
		}
	}
	return res, nil
}

// unary → ( "!" | "-" ) unary | call;
func (p *Parser) unary() (Node, error) {
	if cur := p.tokens[p.current]; p.match(lexer.BANG, lexer.MINUS) {
		right, err := p.primary()
		if err != nil {
			return nil, err
		}
		return &Unary{
			Operator: cur,
			Right:    right,
		}, nil
	}
	return p.call()
}

// call → primary ( "(" arguments? ")" | "." IDENTIFIER )* ;
func (p *Parser) call() (Node, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(lexer.LEFT_PAREN) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(lexer.DOT) {
			name, err := p.consume(lexer.IDENTIFIER)
			if err != nil {
				return nil, fmt.Errorf("expected identifier after '.' in call expression; err=%w", err)
			}
			expr = &GetExpr{Instance: expr, Name: name}
		} else {
			return expr, nil
		}
	}
}

func (p *Parser) finishCall(callee Node) (Node, error) {
	var args []Node
	if !p.match(lexer.RIGHT_PAREN) {
		for {
			if len(args) >= 255 {
				return nil, p.tokens[p.current].GenerateTokenError("Can't have more than 255 arguments.")
			}
			arg, err := p.expressionStmt()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if !p.match(lexer.COMMA) {
				if !p.match(lexer.RIGHT_PAREN) {
					return nil, p.tokens[p.current].GenerateTokenError("Expected closing parenthesis after arg list in function call")
				}
				break
			}
		}
	}

	return &CallExpr{
		Callee: callee,
		Args:   args,
		Paren:  p.getCurrent(),
	}, nil
}

// primary → NUMBER | STRING | "true" | "false" | "nil" | "(" expressionStmt ")" | IDENTIFIER ;
func (p *Parser) primary() (Node, error) {
	cur := p.tokens[p.current]

	// Deal with only token which expects further tokens, otherwise advance and switch
	if cur.Type == lexer.LEFT_PAREN {
		if p.advance() != nil {
			return nil, cur.GenerateTokenError("BAD ERROR - cannot end program with '('")
		}
		inner, err := p.expressionStmt()
		if err != nil {
			return nil, err
		}
		if !p.match(lexer.RIGHT_PAREN) {
			return nil, cur.GenerateTokenError("unexpected token, expected ')'")
		}
		return &Grouping{inner}, nil
	}

	_ = p.advance()
	switch cur.Type {
	case lexer.NUMBER, lexer.STRING:
		return &Literal{Value: cur.Literal}, nil

	case lexer.TRUE:
		return &Literal{Value: true}, nil
	case lexer.FALSE:
		return &Literal{Value: false}, nil
	case lexer.NIL:
		return &Literal{Value: nil}, nil

	case lexer.IDENTIFIER:
		return &Variable{TokenName: cur.Lexeme}, nil

	default:
		return nil, cur.GenerateTokenError("Could not parse Expression, expected a primary Expression")
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

// peekMatch will attempt to match one of many token types and if it does, will NOT advance the parser head
// and return true, else returns false
func (p *Parser) peekMatch(t ...lexer.TokenType) bool {
	for _, tt := range t {
		cur := p.getCurrent()
		if cur == nil {
			p.log.Error("Tried to match where there is no further input")
			return false
		}
		if cur.Type == tt {
			return true
		}
	}
	return false
}

// match will attempt to match one of many token types and if it does, will advance the parser head
// and return true, else returns false
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

// consume looks at the next token and tries to match it to type t. If a match is successful then the parser is advanced
// otherwise an error is returned
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

// synchronize will attempt to skip tokens until it gets past a statement terminator (semicolon) or a new statement
// it's mostly used for REPL recovery
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
