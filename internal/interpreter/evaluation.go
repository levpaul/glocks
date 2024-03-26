package interpreter

import (
	"errors"
	"fmt"
	"github.com/levpaul/glocks/internal/domain"
	"github.com/levpaul/glocks/internal/environment"
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/levpaul/glocks/internal/parser"
)

const NilStatementErrorMessage = "can not evaluate a nil expression"

type EarlyReturn struct {
	result domain.Value
}

func (e EarlyReturn) Error() string {
	return fmt.Sprintf("Returned early from a function with value '%v'", e.result)
}

func (i *Interpreter) VisitClassDeclaration(c *parser.ClassDeclaration) error {
	// declare(c.Name)
	// define(c.Name)
	return errors.New("class decl is unimplemented")
}

func (i *Interpreter) VisitReturnStmt(r *parser.ReturnStmt) error {
	var err error
	earlyReturn := EarlyReturn{}
	// return here to last func call
	if r.Expression != nil {
		earlyReturn.result, err = i.Evaluate(r.Expression)
		if err != nil {
			return err
		}
	}
	return earlyReturn
}

func (i *Interpreter) VisitFunctionDeclaration(f *parser.FunctionDeclaration) error {
	i.env.Define(f.Name, LoxFunction{
		declaration: f,
		closure:     i.env,
	})
	i.evalRes = nil
	return nil
}

func (i *Interpreter) VisitCallExpr(f *parser.CallExpr) error {
	callee, err := i.Evaluate(f.Callee)
	if err != nil {
		return err
	}

	var args []domain.Value
	for _, a := range f.Args {
		evaluatedArg, argErr := i.Evaluate(a)
		if argErr != nil {
			return argErr
		}
		args = append(args, evaluatedArg)
	}

	loxFunction, ok := callee.(parser.LoxCallable)
	if !ok {
		return fmt.Errorf("Expected %v to be of type Callable!", callee)
	}

	if len(args) != loxFunction.Arity() {
		return fmt.Errorf("Expected %d args to be passed to func, but only received %d.", loxFunction.Arity(), len(args))
	}

	i.evalRes, err = loxFunction.Call(i, args)
	return err
}

func (i *Interpreter) VisitWhileStmt(w *parser.WhileStmt) error {
	for {
		exprRes, err := i.Evaluate(w.Expression)
		if err != nil {
			return err
		}
		if !isTruthy(exprRes) {
			break
		}

		i.evalRes, err = i.Evaluate(w.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) VisitLogicalConjunction(c *parser.LogicalConjuction) error {
	left, err := i.Evaluate(c.Left)
	if err != nil {
		return err
	}

	if c.And { // AND case
		if !isTruthy(left) { // short circuit
			i.evalRes = left
			return nil
		}
		right, rErr := i.Evaluate(c.Right)
		if rErr != nil {
			return rErr
		}
		i.evalRes = right
		return nil
	}

	// Case where OR is the conjunction
	if isTruthy(left) { // short-circuit
		i.evalRes = left
		return nil
	}
	right, rErr := i.Evaluate(c.Right)
	if rErr != nil {
		return rErr
	}
	i.evalRes = right
	return nil
}

func (i *Interpreter) VisitIfStmt(ifStmt *parser.IfStmt) error {
	val, err := i.Evaluate(ifStmt.Expression)
	if err != nil {
		return err
	}

	if isTruthy(val) {
		return ifStmt.Statement.Accept(i)
	}

	if ifStmt.ElseStatement == nil {
		return nil
	}

	return ifStmt.ElseStatement.Accept(i)
}

func (i *Interpreter) VisitBlock(b *parser.Block) error {
	// Create a new environment for execution of Block b
	oldEnv := i.env
	i.env = &environment.Environment{
		Enclosing: oldEnv,
		Values:    map[string]domain.Value{},
	}
	defer func() {
		i.env = oldEnv
	}()

	for _, stmt := range b.Statements {
		result, err := i.Evaluate(stmt)
		if err != nil {
			return err
		}
		if i.replMode && result != nil { // only print our statements which evaluate to a Value
			fmt.Println("evaluates to:", result)
		}
	}
	return nil
}

func (i *Interpreter) VisitAssignment(a *parser.Assignment) error {
	v, err := i.Evaluate(a.Value)
	if err != nil {
		return err
	}

	if dist, exists := i.locals[a]; exists {
		if err = i.env.SetAt(dist, a.TokenName, v); err != nil {
			return err
		}
	} else {
		if err = i.globals.Set(a.TokenName, v); err != nil {
			return err
		}
	}
	i.evalRes = v
	return nil
}

func (i *Interpreter) VisitVariable(v *parser.Variable) error {
	val, err := i.lookUpVariable(v.TokenName, v)
	if err != nil {
		return err
	}
	i.evalRes = val
	return nil
}

func (i *Interpreter) VisitVarStmt(v *parser.VarStmt) error {
	var err error
	var initializer domain.Value
	if v.Initializer != nil {
		initializer, err = i.Evaluate(v.Initializer)
		if err != nil {
			return err
		}
	}
	i.env.Define(v.Name, initializer)
	return nil
}

func (i *Interpreter) VisitPrintStmt(p *parser.PrintStmt) error {
	err := p.Arg.Accept(i)
	if err != nil {
		return err
	}
	fmt.Println(i.evalRes)
	i.evalRes = nil
	return nil
}

func (i *Interpreter) VisitBinary(b *parser.Binary) error {
	left, err := i.Evaluate(b.Left)
	if err != nil {
		return err
	}
	right, err := i.Evaluate(b.Right)
	if err != nil {
		return err
	}
	switch b.Operator.Type {
	case lexer.MINUS:
		if err = i.validateBothNumber(left, right); err != nil {
			return err
		}
		i.evalRes = left.(float64) - right.(float64)
	case lexer.SLASH:
		if err = i.validateBothNumber(left, right); err != nil {
			return err
		}
		i.evalRes = left.(float64) / right.(float64)
	case lexer.STAR:
		if err = i.validateBothNumber(left, right); err != nil {
			return err
		}
		i.evalRes = left.(float64) * right.(float64)
	case lexer.PLUS:
		if i.validateBothNumber(left, right) == nil {
			i.evalRes = left.(float64) + right.(float64)
		} else if i.validateBothString(left, right) == nil {
			i.evalRes = left.(string) + right.(string)
		} else {
			return fmt.Errorf("could not use + on values that are not both strings or numbers, values: '%v', '%v'", left, right)
		}
	case lexer.LESS:
		if err = i.validateBothNumber(left, right); err != nil {
			return err
		}
		i.evalRes = left.(float64) < right.(float64)
	case lexer.LESS_EQUAL:
		if err = i.validateBothNumber(left, right); err != nil {
			return err
		}
		i.evalRes = left.(float64) <= right.(float64)
	case lexer.GREATER:
		if err = i.validateBothNumber(left, right); err != nil {
			return err
		}
		i.evalRes = left.(float64) > right.(float64)
	case lexer.GREATER_EQUAL:
		if err = i.validateBothNumber(left, right); err != nil {
			return err
		}
		i.evalRes = left.(float64) <= right.(float64)
	case lexer.EQUAL_EQUAL:
		i.evalRes = isEqual(left, right)
	case lexer.BANG_EQUAL:
		i.evalRes = !isEqual(left, right)

	default:
		return fmt.Errorf("unexpected operator type in binary: %+v", b)
	}

	return nil
}

func (i *Interpreter) VisitGrouping(g *parser.Grouping) error {
	var err error
	i.evalRes, err = i.Evaluate(g.Expression)
	return err
}

func (i *Interpreter) VisitLiteral(l *parser.Literal) error {
	i.evalRes = l.Value
	return nil
}

func (i *Interpreter) VisitUnary(u *parser.Unary) error {
	var err error
	i.evalRes, err = i.Evaluate(u.Right)
	if err != nil {
		return err
	}

	switch u.Operator.Type {
	case lexer.MINUS:
		val, ok := i.evalRes.(float64)
		if !ok {
			return fmt.Errorf("expected number with unary operator, had '%+v' instead", val)
		}
		i.evalRes = -val
	case lexer.BANG:
		i.evalRes = isTruthy(i.evalRes)
	default:
		return fmt.Errorf("unexpected operator type in unary: %+v", u)
	}
	return nil
}

func (i *Interpreter) Evaluate(stmt parser.Node) (domain.Value, error) {
	if stmt == nil {
		return nil, errors.New(NilStatementErrorMessage)
	}

	if err := stmt.Accept(i); err != nil {
		return nil, err
	}
	retVal := i.evalRes
	i.evalRes = nil
	return retVal, nil
}

// isTruthy follows the ruby logic for truthiness - i.e. anything not-nil is truthy
func isTruthy(v domain.Value) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

func isEqual(v1, v2 domain.Value) bool {
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil {
		return false
	}

	return v1 == v2
}

func (i *Interpreter) validateBothNumber(left domain.Value, right domain.Value) error {
	// Benchmarks show that using a custom struct for values, where a member stores the specific underlying type
	// would increase the performance here by 30%, but it means trading off extra memory per value and still doesn't
	// help during evaluation anyway, as we need to type assert to run operations like addition etc, could be cool
	// to write richer benchmarks to evaluate overall memory vs speed costs after full interpreter impl
	// Note: those benchmarks only show performance difference when trying to type assert on passed struct types,
	// not when passing interface types.
	// Another note is that benchmarks show no difference in validating via type assertions and then re-asserting
	// for using the value - the compiler must be optimizing that for us in any case
	_, ok := left.(float64)
	if !ok {
		return fmt.Errorf("%v is not a number", left)
	}

	_, ok = right.(float64)
	if !ok {
		return fmt.Errorf("%v is not a number", right)
	}
	return nil
}

func (i *Interpreter) validateBothString(left domain.Value, right domain.Value) error {
	_, ok := left.(string)
	if !ok {
		return fmt.Errorf("%v is not a string", left)
	}

	_, ok = right.(string)
	if !ok {
		return fmt.Errorf("%v is not a string", right)
	}
	return nil
}
