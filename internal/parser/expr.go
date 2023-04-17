package parser

import (
	"github.com/levpaul/glocks/internal/lexer"
)

type Expr interface {
	Accept(v Visitor) error
}

type Binary struct {
	Left     Expr
	Right    Expr
	Operator *lexer.Token
}

func (b Binary) Accept(v Visitor) error {
	return v.VisitBinary(b)
}

type Grouping struct {
	Expression Expr
}

func (g Grouping) Accept(v Visitor) error {
	return v.VisitGrouping(g)
}

type Literal struct {
	Value any // Probably make a union type here
}

func (l Literal) Accept(v Visitor) error {
	return v.VisitLiteral(l)
}

type Unary struct {
	Operator *lexer.Token
	Right    Expr
}

func (u Unary) Accept(v Visitor) error {
	return v.VisitUnary(u)
}

type Variable struct {
	TokenName string
}

func (v Variable) Accept(visitor Visitor) error {
	return visitor.VisitVariable(v)
}
