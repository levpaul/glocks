package parser

type Visitor interface {
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
}
