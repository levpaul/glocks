package parser

type Stmt interface {
	Accept(Visitor) error
}

type ExprStmt struct {
	expr Expr
}

func (e ExprStmt) Accept(v Visitor) error {
	return v.VisitExprStmt(e)
}

type PrintStmt struct {
	arg Expr
}

func (p PrintStmt) Accept(v Visitor) error {
	return v.VisitPrintStmt(p)
}

type VarStmt struct {
	name string
	val  Value
}

func (v VarStmt) Accept(visitor Visitor) error {
	return visitor.VisitVarStmt(v)
}