package parser

import "github.com/levpaul/glocks/internal/lexer"

type Block struct {
	Statements []Node
}

func (b Block) Accept(v Visitor) error {
	return v.VisitBlock(b)
}

type Binary struct {
	Left     Node
	Right    Node
	Operator *lexer.Token
}

func (b Binary) Accept(v Visitor) error {
	return v.VisitBinary(b)
}

type Grouping struct {
	Expression Node
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
	Right    Node
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

type Assignment struct {
	TokenName string
	Value     Node
}

func (a Assignment) Accept(visitor Visitor) error {
	return visitor.VisitAssignment(a)
}

type Value any

type Node interface {
	Accept(Visitor) error
}

type ExprStmt struct {
	E Node
}

func (e ExprStmt) Accept(v Visitor) error {
	return v.VisitExprStmt(e)
}

type PrintStmt struct {
	Arg Node
}

func (p PrintStmt) Accept(v Visitor) error {
	return v.VisitPrintStmt(p)
}

type VarStmt struct {
	Name        string
	Initializer Node
}

func (v VarStmt) Accept(visitor Visitor) error {
	return visitor.VisitVarStmt(v)
}
