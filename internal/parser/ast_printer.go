package parser

import (
	"fmt"
	"strings"
)

type ExprPrinter struct {
	res string
}

func (e *ExprPrinter) VisitVariable(v Variable) error {
	//TODO implement me
	panic("implement me")
}

func (e *ExprPrinter) VisitVarStmt(v VarStmt) error {
	//TODO implement me
	panic("implement me")
}

func (e *ExprPrinter) VisitBinary(b Binary) error {
	e.res = e.parenthesize(b.Operator.Lexeme, b.Left, b.Right)
	return nil
}

func (e *ExprPrinter) VisitGrouping(g Grouping) error {
	e.res = e.parenthesize("group", g.Expression)
	return nil
}

func (e *ExprPrinter) VisitLiteral(l Literal) error {
	if l.Value == nil {
		e.res = "nil"
	}
	e.res = fmt.Sprintf("%+v", l.Value)
	return nil
}

func (e *ExprPrinter) VisitUnary(u Unary) error {
	e.res = e.parenthesize(u.Operator.Lexeme, u.Right)
	return nil
}

func (e *ExprPrinter) VisitExprStmt(exprStmt ExprStmt) error {
	return exprStmt.E.Accept(e)
}

func (e *ExprPrinter) VisitPrintStmt(p PrintStmt) error {
	e.res = e.parenthesize("print", p.Arg)
	return nil
}

// Print walks through an expression and prints it in a Lisp like syntax
func (e *ExprPrinter) Print(stmt Stmt) string {
	if stmt == nil {
		return ""
	}

	stmt.Accept(e)
	return e.res
}

func (e *ExprPrinter) parenthesize(name string, exprs ...Expr) string {
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
