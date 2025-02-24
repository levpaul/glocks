package interpreter

import (
	"errors"
	"fmt"

	"github.com/levpaul/glocks/internal/parser"
)

const MAX_SCOPES = 255

type FunctionType int

const (
	FT_NONE FunctionType = iota
	FT_FUNCTION
)

type Scope map[string]bool

// Resolver is responsible for resolving variable names to their scope. It walks the entire AST
// before execution to resolve variable names to their scope.
type Resolver struct {
	i               *Interpreter
	scopes          []Scope
	currentFunction FunctionType
}

func (r *Resolver) VisitClassDeclaration(c *parser.ClassDeclaration) error {
	r.declare(c.Name)
	r.define(c.Name)
	return nil
}

func (r *Resolver) resolve(node parser.Node) error {
	return node.Accept(r)
}

func (r *Resolver) resolveNodes(nodes []parser.Node) error {
	for _, node := range nodes {
		if err := r.resolve(node); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) beginScope() error {
	if len(r.scopes) > MAX_SCOPES {
		return fmt.Errorf("maximum number of scopes (%d) exceeded", MAX_SCOPES)
	}
	r.scopes = append([]Scope{{}}, r.scopes...)
	return nil
}

func (r *Resolver) endScope() error {
	if len(r.scopes) < 1 {
		return errors.New("attempted to end a scope with no scopes to end")
	}
	r.scopes = r.scopes[1:]
	return nil
}

func (r *Resolver) declare(name string) {
	if len(r.scopes) == 0 {
		return
	}

	r.scopes[0][name] = false
}

func (r *Resolver) define(name string) {
	if len(r.scopes) == 0 {
		return
	}

	r.scopes[0][name] = true
}

// resolveLocal walks through the scopes stack, from narrowest to widest to find the 'distance' to resolution
func (r *Resolver) resolveLocal(node parser.Node, name string) {
	// This is different to the book as Java indexes Stacks with 0 being the bottom of the stack
	// whereas here I'm using the zero index as the top of the stack
	for i, scope := range r.scopes {
		if _, exists := scope[name]; exists {
			r.i.resolve(node, i)
			return
		}
	}
	return
}

func (r *Resolver) resolveFunction(f *parser.FunctionDeclaration, ft FunctionType) error {
	enclosingFunction := r.currentFunction
	r.currentFunction = ft
	defer func() { r.currentFunction = enclosingFunction }()

	if err := r.beginScope(); err != nil {
		return err
	}
	for _, p := range f.Params {
		r.declare(p)
		r.define(p)
	}
	if err := r.resolveNodes(f.Body); err != nil {
		return err
	}
	return r.endScope()
}

// ============================================================
//	VISITOR METHODS

func (r *Resolver) VisitIfStmt(i *parser.IfStmt) error {
	err := r.resolve(i.Expression)
	if err != nil {
		return err
	}
	r.resolve(i.Statement)
	if err != nil {
		return err
	}
	if i.ElseStatement != nil {
		return r.resolve(i.ElseStatement)
	}
	return nil
}

func (r *Resolver) VisitBlock(b *parser.Block) error {
	err := r.beginScope()
	if err != nil {
		return err
	}
	err = r.resolveNodes(b.Statements)
	if err != nil {
		return err
	}
	return r.endScope()
}

func (r *Resolver) VisitBinary(b *parser.Binary) error {
	err := r.resolve(b.Left)
	if err != nil {
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
	if len(r.scopes) > 0 {
		if defined, exists := r.scopes[0][v.TokenName]; exists && !defined {
			return fmt.Errorf("can't read local variable '%s' in its own initializer", v.TokenName)
		}
	}
	r.resolveLocal(v, v.TokenName)
	return nil
}

func (r *Resolver) VisitPrintStmt(p *parser.PrintStmt) error {
	return r.resolve(p.Arg)
}

func (r *Resolver) VisitVarStmt(v *parser.VarStmt) error {
	if len(r.scopes) > 0 {
		if _, exists := r.scopes[0][v.Name]; exists {
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
