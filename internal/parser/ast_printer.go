package parser

import (
	"fmt"
	"strings"
)

type ExprPrinter struct {
	res string
}

func reflectPrint(i interface{}) error {
	fmt.Printf("%+v\n", i)
	return nil
}

// VisitGetExpr implements Visitor.
func (e *ExprPrinter) VisitGetExpr(g *GetExpr) error {
	return reflectPrint(g)
}

func (e *ExprPrinter) VisitClassDeclaration(c *ClassDeclaration) error {
	return reflectPrint(c)
}

func (e *ExprPrinter) VisitReturnStmt(r *ReturnStmt) error {
	return reflectPrint(r)
}

func (e *ExprPrinter) VisitFunctionDeclaration(f *FunctionDeclaration) error {
	return reflectPrint(f)
}

func (e *ExprPrinter) VisitCallExpr(f *CallExpr) error {
	return reflectPrint(f)
}

func (e *ExprPrinter) VisitWhileStmt(w *WhileStmt) error {
	return reflectPrint(w)
}

func (e *ExprPrinter) VisitLogicalConjunction(v *LogicalConjuction) error {
	return reflectPrint(v)
}

func (e *ExprPrinter) VisitIfStmt(i *IfStmt) error {
	return reflectPrint(i)
}

func (e *ExprPrinter) VisitBlock(b *Block) error {
	return reflectPrint(b)
}

func (e *ExprPrinter) VisitAssignment(v *Assignment) error {
	return reflectPrint(v)
}

func (e *ExprPrinter) VisitVariable(v *Variable) error {
	return reflectPrint(v)
}

func (e *ExprPrinter) VisitVarStmt(v *VarStmt) error {
	return reflectPrint(v)
}

func (e *ExprPrinter) VisitBinary(b *Binary) error {
	e.res = e.parenthesize(b.Operator.Lexeme, b.Left, b.Right)
	return nil
}

func (e *ExprPrinter) VisitGrouping(g *Grouping) error {
	e.res = e.parenthesize("group", g.Expression)
	return nil
}

func (e *ExprPrinter) VisitLiteral(l *Literal) error {
	if l.Value == nil {
		e.res = "nil"
	}
	e.res = fmt.Sprintf("%+v", l.Value)
	return nil
}

func (e *ExprPrinter) VisitUnary(u *Unary) error {
	e.res = e.parenthesize(u.Operator.Lexeme, u.Right)
	return nil
}

func (e *ExprPrinter) VisitPrintStmt(p *PrintStmt) error {
	e.res = e.parenthesize("print", p.Arg)
	return nil
}

// Print walks through an Expression and prints it in a Lisp like syntax
func (e *ExprPrinter) Print(stmt Node) string {
	if stmt == nil {
		return ""
	}

	stmt.Accept(e)
	return e.res
}

func (e *ExprPrinter) parenthesize(name string, exprs ...Node) string {
	builder := strings.Builder{}
	builder.WriteString("(")
	builder.WriteString(name)

	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(e.Print(expr))
	}

	builder.WriteString(")")
	return builder.String()
}
