package parser

type Value any

type Stmt interface {
	Accept(Visitor) error
}

type ExprStmt struct {
	E Expr
}

func (e ExprStmt) Accept(v Visitor) error {
	return v.VisitExprStmt(e)
}

type PrintStmt struct {
	Arg Expr
}

func (p PrintStmt) Accept(v Visitor) error {
	return v.VisitPrintStmt(p)
}

type VarStmt struct {
	Name        string
	Initializer Expr
}

func (v VarStmt) Accept(visitor Visitor) error {
	return visitor.VisitVarStmt(v)
}
