package interpreter

import (
	"errors"
	"fmt"
	"github.com/levpaul/glocks/internal/lexer"
	"github.com/levpaul/glocks/internal/parser"
)

const NilStatementErrorMessage = "can not evaluate a nil expression"

func (i *Interpreter) VisitLogicalConjunction(c parser.LogicalConjuction) error {
	left, err := i.Evaluate(c.Left)
	if err != nil {
		return err
	}

	if c.And { // AND case
		if isTruthy(left) {
			right, rErr := i.Evaluate(c.Right)
			if rErr != nil {
				return rErr
			}
			i.evalRes = isTruthy(right)
			return nil
		}
		i.evalRes = false
		return nil
	}

	// Case where OR is the conjunction
	if isTruthy(left) {
		i.evalRes = true
		return nil
	}
	right, rErr := i.Evaluate(c.Right)
	if rErr != nil {
		return rErr
	}
	i.evalRes = isTruthy(right)
	return nil
}

func (i *Interpreter) VisitIfStmt(ifStmt parser.IfStmt) error {
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

func (i *Interpreter) VisitBlock(b parser.Block) error {
	oldEnv := i.env
	i.env = &Environment{
		Enclosing: oldEnv,
		Values:    map[string]parser.Value{},
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

func (i *Interpreter) VisitAssignment(a parser.Assignment) error {
	v, err := i.Evaluate(a.Value)
	if err != nil {
		return err
	}
	if err = i.env.Set(a.TokenName, v); err != nil {
		return err
	}
	i.evalRes = v
	return nil
}

func (i *Interpreter) VisitVariable(v parser.Variable) (err error) {
	i.evalRes, err = i.env.Get(v.TokenName)
	return
}

func (i *Interpreter) VisitVarStmt(v parser.VarStmt) error {
	var err error
	var initializer parser.Value
	if v.Initializer != nil {
		initializer, err = i.Evaluate(v.Initializer)
		if err != nil {
			return err
		}
	}
	i.env.Define(v.Name, initializer)
	return nil
}

func (i *Interpreter) VisitExprStmt(s parser.ExprStmt) error {
	return s.E.Accept(i)
}

func (i *Interpreter) VisitPrintStmt(p parser.PrintStmt) error {
	err := p.Arg.Accept(i)
	if err != nil {
		return err
	}
	fmt.Println(i.evalRes)
	i.evalRes = nil
	return nil
}

func (i *Interpreter) VisitBinary(b parser.Binary) error {
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

func (i *Interpreter) VisitGrouping(g parser.Grouping) error {
	var err error
	i.evalRes, err = i.Evaluate(g.Expression)
	return err
}

func (i *Interpreter) VisitLiteral(l parser.Literal) error {
	i.evalRes = l.Value
	return nil
}

func (i *Interpreter) VisitUnary(u parser.Unary) error {
	var err error
	i.evalRes, err = i.Evaluate(u.Right)
	if err != nil {
		return err
	}

	switch u.Operator.Type {
	case lexer.BANG:
		val, ok := i.evalRes.(float64)
		if !ok {
			return fmt.Errorf("expected number with unary operator, had '%+v' instead", val)
		}
		i.evalRes = -val
	case lexer.MINUS:
		i.evalRes = isTruthy(i.evalRes)
	default:
		return fmt.Errorf("unexpected operator type in unary: %+v", u)
	}
	return nil
}

func (i *Interpreter) Evaluate(stmt parser.Node) (parser.Value, error) {
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

// isTruthy follows the ruby logic for truthiness
func isTruthy(v parser.Value) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

func isEqual(v1, v2 parser.Value) bool {
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil {
		return false
	}

	return v1 == v2
}

func (i *Interpreter) validateBothNumber(left parser.Value, right parser.Value) error {
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

func (i *Interpreter) validateBothString(left parser.Value, right parser.Value) error {
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
