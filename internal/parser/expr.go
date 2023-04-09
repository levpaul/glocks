package parser

import "github.com/levpaul/glocks/internal/lexer"

type Expr interface {
}

type Binary struct {
	Left     Expr
	Right    Expr
	Operator *lexer.Token
}

type Grouping struct {
	Expression Expr
}

type Literal struct {
	Value any // Probably make a union type here
}

type Unary struct {
	Operator *lexer.Token
	Right    Expr
}
