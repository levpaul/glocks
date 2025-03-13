package resolver

import (
	"errors"
	"fmt"

	"github.com/levpaul/glocks/internal/parser"
)

func (r *Resolver) VisitIfStmt(i *parser.IfStmt) error {
	if err := r.resolve(i.Expression); err != nil {
		return err
	}
	if err := r.resolve(i.Statement); err != nil {
		return err
	}
	if i.ElseStatement != nil {
		return r.resolve(i.ElseStatement)
	}
	return nil
}

func (r *Resolver) VisitBlock(b *parser.Block) error {
	if err := r.beginScope(); err != nil {
		return err
	}
	if err := r.ResolveNodes(b.Statements); err != nil {
		return err
	}
	return r.endScope()
}

func (r *Resolver) VisitBinary(b *parser.Binary) error {
	if err := r.resolve(b.Left); err != nil {
		return err
	}
	return r.resolve(b.Right)
}

func (r *Resolver) VisitGrouping(g *parser.Grouping) error {
	return r.resolve(g.Expression)
}

func (r *Resolver) VisitLiteral(l *parser.Literal) error {
	return nil
}

func (r *Resolver) VisitUnary(u *parser.Unary) error {
	return r.resolve(u.Right)
}

func (r *Resolver) VisitVariable(v *parser.Variable) error {
	if len(r.Scopes) > 0 {
		if defined, declared := r.Scopes[0][v.TokenName]; declared && !defined {
			return fmt.Errorf("can't read local variable '%s' in its own initializer", v.TokenName)
		}
	}
	r.resolveLocal(v, v.TokenName)
	return nil
}

func (r *Resolver) VisitPrintStmt(p *parser.PrintStmt) error {
	return r.resolve(p.Arg)
}

// VisitVarStmt declares a variable in the current scope, and optionally initializes it
// with an expression. The resolver will check that the variable is not already declared in
// the current scope.
func (r *Resolver) VisitVarStmt(v *parser.VarStmt) error {
	if len(r.Scopes) > 0 {
		if _, exists := r.Scopes[0][v.Name]; exists {
			return fmt.Errorf("already exists a variable with name='%s' in scope", v.Name)
		}
	}

	r.declare(v.Name)
	if v.Initializer != nil {
		err := r.resolve(v.Initializer)
		if err != nil {
			return err
		}
	}
	r.define(v.Name)
	return nil
}

func (r *Resolver) VisitAssignment(v *parser.Assignment) error {
	err := r.resolve(v.Value)
	if err != nil {
		return err
	}
	r.resolveLocal(v, v.TokenName)
	return nil
}

func (r *Resolver) VisitLogicalConjunction(v *parser.LogicalConjuction) error {
	err := r.resolve(v.Left)
	if err != nil {
		return err
	}
	return r.resolve(v.Right)
}

func (r *Resolver) VisitWhileStmt(w *parser.WhileStmt) error {
	err := r.resolve(w.Expression)
	if err != nil {
		return err
	}
	return r.resolve(w.Body)
}

func (r *Resolver) VisitCallExpr(f *parser.CallExpr) error {
	err := r.resolve(f.Callee)
	if err != nil {
		return err
	}
	for _, arg := range f.Args {
		err = r.resolve(arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) VisitFunctionDeclaration(f *parser.FunctionDeclaration) error {
	r.declare(f.Name)
	r.define(f.Name)
	return r.resolveFunction(f, FT_FUNCTION)
}

func (r *Resolver) VisitReturnStmt(rs *parser.ReturnStmt) error {
	if r.currentFunction == FT_NONE {
		return errors.New("detected return statement from global scope - not allowed")
	}
	return r.resolve(rs.Expression)
}

// VisitGetExpr implements parser.Visitor.
func (r *Resolver) VisitGetExpr(g *parser.GetExpr) error {
	return r.resolve(g.Instance)
}

// VisitClassDeclaration declares and defines a class from a ClassDeclaration node
func (r *Resolver) VisitClassDeclaration(c *parser.ClassDeclaration) error {
	r.declare(c.Name)
	r.define(c.Name)
	return nil
}
