package parser

import "github.com/levpaul/glocks/internal/lexer"

type Expr interface {
	Accept(v Visitor)
}

type Visitor interface {
	VisitBinary(b Binary)
	VisitGrouping(g Grouping)
	VisitLiteral(l Literal)
	VisitUnary(u Unary)
}

type Binary struct {
	Left     Expr
	Right    Expr
	Operator *lexer.Token
}

func (b Binary) Accept(v Visitor) {
	v.VisitBinary(b)
}

type Grouping struct {
	Expression Expr
}

func (g Grouping) Accept(v Visitor) {
	v.VisitGrouping(g)
}

type Literal struct {
	Value any // Probably make a union type here
}

func (l Literal) Accept(v Visitor) {
	v.VisitLiteral(l)
}

type Unary struct {
	Operator *lexer.Token
	Right    Expr
}

func (u Unary) Accept(v Visitor) {
	v.VisitUnary(u)
}
