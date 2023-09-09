package parser

import (
	"github.com/levpaul/glocks/internal/lexer"
)

type FunctionCallStmt struct {
	Callee Node
	Args   []Node
}

func (f FunctionCallStmt) Call(i LoxInterpreter, args []Value) Value {
	// TODO: impl... this code needs to generate AST nodes, drop in "values", evaluate nodes and return value? dynamically?
	return nil
}

func (f FunctionCallStmt) Accept(v Visitor) error {
	return v.VisitFunctionCallStmt(f)
}

type WhileStmt struct {
	Expression Node
	Body       Node
}

func (w WhileStmt) Accept(v Visitor) error {
	return v.VisitWhileStmt(w)
}

type LogicalConjuction struct {
	Left  Node
	And   bool
	Right Node
}

func (c LogicalConjuction) Accept(v Visitor) error {
	return v.VisitLogicalConjunction(c)
}

type IfStmt struct {
	Expression    Node
	Statement     Node
	ElseStatement Node
}

func (i IfStmt) Accept(v Visitor) error {
	return v.VisitIfStmt(i)
}

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

type Value any

type Visitor interface {
	VisitIfStmt(i IfStmt) error
	VisitBlock(b Block) error
	VisitBinary(b Binary) error
	VisitGrouping(g Grouping) error
	VisitLiteral(l Literal) error
	VisitUnary(u Unary) error
	VisitVariable(v Variable) error
	VisitExprStmt(s ExprStmt) error
	VisitPrintStmt(p PrintStmt) error
	VisitVarStmt(v VarStmt) error
	VisitAssignment(v Assignment) error
	VisitLogicalConjunction(v LogicalConjuction) error
	VisitWhileStmt(w WhileStmt) error
	VisitFunctionCallStmt(f FunctionCallStmt) error
}

type LoxInterpreter interface {
	Evaluate(Node) (Value, error)
}

type LoxCallable interface {
	Arity() int
	Call(i LoxInterpreter, args []Value) (Value, error)
}
