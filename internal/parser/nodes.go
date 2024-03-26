package parser

import (
	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/environment"
	"github.com/levpaul/glocks/internal/lexer"
)

type FunctionDeclaration struct {
	Name   string
	Params []string
	Body   Node
}

func (f FunctionDeclaration) Accept(v Visitor) error {
	return v.VisitFunctionDeclaration(f)
}

type CallExpr struct {
	LoxCallable
	Callee Node
	Paren  *lexer.Token // for debugging + reporting
	Args   []Node
}

func (f CallExpr) Call(i LoxInterpreter, args []domain.Value) domain.Value {
	//env := environment.Environment{}
	panic("implement me")
}

func (f CallExpr) Accept(v Visitor) error {
	return v.VisitCallExpr(f)
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
	VisitCallExpr(f CallExpr) error
	VisitFunctionDeclaration(f FunctionDeclaration) error
}

type LoxInterpreter interface {
	Evaluate(Node) (domain.Value, error)
	ExecuteBlock(Block, *environment.Environment) error
	GetEnvironment() *environment.Environment
}

type LoxCallable interface {
	Arity() int
	Call(i LoxInterpreter, args []domain.Value) (domain.Value, error)
}
