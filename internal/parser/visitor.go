package parser

type Visitor interface {
	VisitBinary(b Binary) error
	VisitGrouping(g Grouping) error
	VisitLiteral(l Literal) error
	VisitUnary(u Unary) error
	VisitExprStmt(s ExprStmt) error
	VisitPrintStmt(p PrintStmt) error
}
