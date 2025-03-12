package resolver

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

// Scope is a map of variable names to whether they have been defined or not
type Scope map[string]bool

// Resolver is responsible for resolving variable names to their scope. It walks the entire AST
// before execution to resolve variable names to their scope.
type Resolver struct {
	// Scopes is a stack of scopes, with the current scope being the top of the stack
	Scopes []Scope
	// currentFunction is the type of function that is currently being resolved
	currentFunction FunctionType
	// locals is a map of nodes to their depth in the scope chain
	locals map[parser.Node]int
}

func NewResolver() *Resolver {
	return &Resolver{
		Scopes:          []Scope{{}},
		currentFunction: FT_NONE,
		locals:          make(map[parser.Node]int),
	}
}

// resolve resolves a single node by calling Accept on the node, which in turn calls the appropriate Visit method
// of the passed Node
func (r *Resolver) resolve(node parser.Node) error {
	return node.Accept(r)
}

// resolveNodes resolves a slice of nodes by calling resolve on each node
func (r *Resolver) ResolveNodes(nodes []parser.Node) error {
	for _, node := range nodes {
		if err := r.resolve(node); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) beginScope() error {
	if len(r.Scopes) > MAX_SCOPES {
		return fmt.Errorf("maximum number of scopes (%d) exceeded", MAX_SCOPES)
	}
	r.Scopes = append([]Scope{{}}, r.Scopes...)
	return nil
}

func (r *Resolver) endScope() error {
	if len(r.Scopes) < 1 {
		return errors.New("attempted to end a scope with no scopes to end")
	}
	r.Scopes = r.Scopes[1:]
	return nil
}

// declare declares a variable in the current scope, but does not define it
func (r *Resolver) declare(name string) {
	if len(r.Scopes) == 0 {
		return
	}

	r.Scopes[0][name] = false
}

// define defines a variable in the current scope marking it as true in the scope map
func (r *Resolver) define(name string) {
	if len(r.Scopes) == 0 {
		return
	}

	r.Scopes[0][name] = true
}

// resolveLocal walks through the scopes stack, from narrowest to widest to find the 'distance' to resolution
func (r *Resolver) resolveLocal(node parser.Node, name string) {
	// This is different to the book as Java indexes Stacks with 0 being the bottom of the stack
	// whereas here I'm using the zero index as the top of the stack
	for i, scope := range r.Scopes {
		if _, exists := scope[name]; exists {
			r.SetDepth(node, i)
			return
		}
	}
}

// resolveFunction resolves a function declaration, including its parameters and body
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
	if err := r.ResolveNodes(f.Body); err != nil {
		return err
	}
	return r.endScope()
}

func (r *Resolver) SetDepth(node parser.Node, depth int) {
	r.locals[node] = depth
}

func (r *Resolver) GetLocal(node parser.Node) (int, error) {
	if depth, exists := r.locals[node]; exists {
		return depth, nil
	}
	return 0, errors.New("could not find local variable")
}
