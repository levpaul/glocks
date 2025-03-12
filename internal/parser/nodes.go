package parser

import (
	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/environment"
	"github.com/levpaul/glocks/internal/lexer"
)

// GetExpr is a node that represents a get expression - that is a dot expression
// that gets a property from an instance of a class.
type GetExpr struct {
	Instance Node
	Name     *lexer.Token
}

func (g *GetExpr) Accept(v Visitor) error {
	return v.VisitGetExpr(g)
}

// ClassDeclaration is a node that represents a class declaration.
type ClassDeclaration struct {
	Name    string
	Methods []Node
}

func (c *ClassDeclaration) Accept(v Visitor) error {
	return v.VisitClassDeclaration(c)
}

type ReturnStmt struct {
	Expression Node
}

func (r *ReturnStmt) Accept(v Visitor) error {
	return v.VisitReturnStmt(r)
}

type FunctionDeclaration struct {
	Name   string
	Params []string
	Body   []Node
}

func (f *FunctionDeclaration) Accept(v Visitor) error {
	return v.VisitFunctionDeclaration(f)
}

type CallExpr struct {
	Callee Node
	Paren  *lexer.Token // for debugging + reporting
	Args   []Node
}

func (f *CallExpr) Accept(v Visitor) error {
	return v.VisitCallExpr(f)
}

type WhileStmt struct {
	Expression Node
	Body       Node
}

func (w *WhileStmt) Accept(v Visitor) error {
	return v.VisitWhileStmt(w)
}

type LogicalConjuction struct {
	Left  Node
	And   bool
	Right Node
}

func (c *LogicalConjuction) Accept(v Visitor) error {
	return v.VisitLogicalConjunction(c)
}

type IfStmt struct {
	Expression    Node
	Statement     Node
	ElseStatement Node
}

func (i *IfStmt) Accept(v Visitor) error {
	return v.VisitIfStmt(i)
}

type Block struct {
	Statements []Node
}

func (b *Block) Accept(v Visitor) error {
	return v.VisitBlock(b)
}

type Binary struct {
	Left     Node
	Right    Node
	Operator *lexer.Token
}

func (b *Binary) Accept(v Visitor) error {
	return v.VisitBinary(b)
}

type Grouping struct {
	Expression Node
}

func (g *Grouping) Accept(v Visitor) error {
	return v.VisitGrouping(g)
}

type Literal struct {
	Value any // Probably make a union type here
}

func (l *Literal) Accept(v Visitor) error {
	return v.VisitLiteral(l)
}

type Unary struct {
	Operator *lexer.Token
	Right    Node
}

func (u *Unary) Accept(v Visitor) error {
	return v.VisitUnary(u)
}

type Variable struct {
	TokenName string
}

func (v *Variable) Accept(visitor Visitor) error {
	return visitor.VisitVariable(v)
}

type Assignment struct {
	TokenName string
	Value     Node
}

func (a *Assignment) Accept(visitor Visitor) error {
	return visitor.VisitAssignment(a)
}

// Node represents a node in the AST. All nodes must implement the Accept method
// which allows the node to be visited by a Visitor.
type Node interface {
	Accept(Visitor) error
}

type PrintStmt struct {
	Arg Node
}

func (p *PrintStmt) Accept(v Visitor) error {
	return v.VisitPrintStmt(p)
}

type VarStmt struct {
	Name        string
	Initializer Node
}

func (v *VarStmt) Accept(visitor Visitor) error {
	return visitor.VisitVarStmt(v)
}

// Visitor is an interface that must be implemented by any object that wishes to
// be applied to the AST.
type Visitor interface {
	VisitIfStmt(i *IfStmt) error
	VisitBlock(b *Block) error
	VisitBinary(b *Binary) error
	VisitGrouping(g *Grouping) error
	VisitLiteral(l *Literal) error
	VisitUnary(u *Unary) error
	VisitVariable(v *Variable) error
	VisitPrintStmt(p *PrintStmt) error
	VisitVarStmt(v *VarStmt) error
	VisitAssignment(v *Assignment) error
	VisitLogicalConjunction(v *LogicalConjuction) error
	VisitWhileStmt(w *WhileStmt) error
	VisitCallExpr(f *CallExpr) error
	VisitFunctionDeclaration(f *FunctionDeclaration) error
	VisitReturnStmt(r *ReturnStmt) error
	VisitClassDeclaration(c *ClassDeclaration) error
	VisitGetExpr(g *GetExpr) error
}

type LoxInterpreter interface {
	Evaluate(Node) (domain.Value, error)
	ExecuteBlock(*Block, *environment.Environment) error
	GetEnvironment() *environment.Environment
}

type LoxCallable interface {
	Arity() int
	Call(i LoxInterpreter, args []domain.Value) (domain.Value, error)
}
