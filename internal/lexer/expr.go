package lexer

import (
	"github.com/levpaul/glocks/internal/token"
)

type Expr interface {
}

type Binary struct {
	Left     Expr
	Right    Expr
	Operator token.Token
}

type Grouping struct {
	Expression Expr
}

type Literal struct {
	Value any // Probably make a union type here
}

type Unary struct {
	Operator token.Token
	Right    Expr
}
