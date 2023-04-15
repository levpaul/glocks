package parser

import (
	"errors"
	"fmt"
	"github.com/levpaul/glocks/internal/lexer"
)

const NilExpressionErrorMessage = "can not evaluate a nil expression"

type Value any

type Evaluator struct {
	res any
}

func (e *Evaluator) validateBothNumber(left Value, right Value) error {
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

		// TODO: Add + < <= > >= != ==
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
func (e *Evaluator) Evaluate(expr Expr) (Value, error) {
	if expr == nil {
		return nil, errors.New(NilExpressionErrorMessage)
	}

	if err := expr.Accept(e); err != nil {
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
