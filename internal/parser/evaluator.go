package parser

import (
	"errors"
	"fmt"
	"github.com/levpaul/glocks/internal/lexer"
)

const NilStatementErrorMessage = "can not evaluate a nil expression"

type Value any

type Evaluator struct {
	res any
}

func (e *Evaluator) VisitVariable(v Variable) error {
	return errors.New("variable expression impl has not been made yet")
}

func (e *Evaluator) VisitVarStmt(v VarStmt) error {
	return errors.New("variable statement impl has not been made yet")
}

func (e *Evaluator) VisitExprStmt(s ExprStmt) error {
	return s.expr.Accept(e)
}

func (e *Evaluator) VisitPrintStmt(p PrintStmt) error {
	err := p.arg.Accept(e)
	if err != nil {
		return err
	}
	fmt.Println(e.res)
	e.res = nil
	return nil
}

func (e *Evaluator) VisitBinary(b Binary) error {
	left, err := e.Evaluate(b.Left)
	if err != nil {
		return err
	}
	right, err := e.Evaluate(b.Right)
	if err != nil {
		return err
	}
	switch b.Operator.Type {
	case lexer.MINUS:
		if err = e.validateBothNumber(left, right); err != nil {
			return err
		}
		e.res = left.(float64) - right.(float64)
	case lexer.SLASH:
		if err = e.validateBothNumber(left, right); err != nil {
			return err
		}
		e.res = left.(float64) / right.(float64)
	case lexer.STAR:
		if err = e.validateBothNumber(left, right); err != nil {
			return err
		}
		e.res = left.(float64) * right.(float64)
	case lexer.PLUS:
		if e.validateBothNumber(left, right) == nil {
			e.res = left.(float64) + right.(float64)
		} else if e.validateBothString(left, right) == nil {
			e.res = left.(string) + right.(string)
		} else {
			return fmt.Errorf("could not use + on values that are not both strings or numbers, values: '%v', '%v'", left, right)
		}
	case lexer.LESS:
		if err = e.validateBothNumber(left, right); err != nil {
			return err
		}
		e.res = left.(float64) < right.(float64)
	case lexer.LESS_EQUAL:
		if err = e.validateBothNumber(left, right); err != nil {
			return err
		}
		e.res = left.(float64) <= right.(float64)
	case lexer.GREATER:
		if err = e.validateBothNumber(left, right); err != nil {
			return err
		}
		e.res = left.(float64) > right.(float64)
	case lexer.GREATER_EQUAL:
		if err = e.validateBothNumber(left, right); err != nil {
			return err
		}
		e.res = left.(float64) <= right.(float64)
	case lexer.EQUAL_EQUAL:
		e.res = isEqual(left, right)
	case lexer.BANG_EQUAL:
		e.res = !isEqual(left, right)

	default:
		return fmt.Errorf("unexpected operator type in binary: %+v", b)
	}

	return nil
}

func (e *Evaluator) VisitGrouping(g Grouping) error {
	var err error
	e.res, err = e.Evaluate(g.Expression)
	return err
}

func (e *Evaluator) VisitLiteral(l Literal) error {
	e.res = l.Value
	return nil
}

func (e *Evaluator) VisitUnary(u Unary) error {
	var err error
	e.res, err = e.Evaluate(u.Right)
	if err != nil {
		return err
	}

	switch u.Operator.Type {
	case lexer.BANG:
		val, ok := e.res.(float64)
		if !ok {
			return fmt.Errorf("expected number with unary operator, had '%+v' instead", val)
		}
		e.res = -val
	case lexer.MINUS:
		e.res = isTruthy(e.res)
	default:
		return fmt.Errorf("unexpected operator type in unary: %+v", u)
	}
	return nil
}

// Print walks through an expression and prints it in a Lisp like syntax
func (e *Evaluator) Evaluate(stmt Stmt) (Value, error) {
	if stmt == nil {
		return nil, errors.New(NilStatementErrorMessage)
	}

	if err := stmt.Accept(e); err != nil {
		return nil, err
	}
	return e.res, nil
}

// isTruthy follows the ruby logic for truthiness
func isTruthy(v Value) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

func isEqual(v1, v2 Value) bool {
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil {
		return false
	}

	return v1 == v2
}

func (e *Evaluator) validateBothNumber(left Value, right Value) error {
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

func (e *Evaluator) validateBothString(left Value, right Value) error {
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
