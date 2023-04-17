package parser

type Visitor interface {
	VisitBinary(b Binary) error
	VisitGrouping(g Grouping) error
	VisitLiteral(l Literal) error
	VisitUnary(u Unary) error
	VisitVariable(v Variable) error
	VisitExprStmt(s ExprStmt) error
	VisitPrintStmt(p PrintStmt) error
	VisitVarStmt(v VarStmt) error
}
