package analyzer

import (
	"fmt"
	"pascc/internal/semantic"
)

type ErrIncorrectType struct {
	Expected semantic.Type
	Got      semantic.Type
}

func (e ErrIncorrectType) Error() string {
	return fmt.Sprintf("Incorrect type, expected %s, got: %s", e.Expected, e.Got)
}

type ErrUnOperationNotAllowed struct {
	Operation semantic.OperatorType
	Type      semantic.Type
}

func (e ErrUnOperationNotAllowed) Error() string {
	return fmt.Sprintf("Unary operation %s is not allowed on type %s", e.Operation, e.Type)
}

type ErrBiOperationNotAllowed struct {
	Operation semantic.OperatorType
	LHS       semantic.Type
	RHS       semantic.Type
}

func (e ErrBiOperationNotAllowed) Error() string {
	return fmt.Sprintf("Binary operation %s is not allowed on types %s & %s", e.Operation, e.LHS, e.RHS)
}

type ErrVariableDoesntExist struct {
	Name string
}

func (e ErrVariableDoesntExist) Error() string {
	return fmt.Sprintf("Variable '%s' doesn't exist", e.Name)
}

type ErrSymbolAlreadyExists struct {
	Name string
}

func (e ErrSymbolAlreadyExists) Error() string {
	return fmt.Sprintf("'%s' already exists!", e.Name)
}

type ErrUnsupportedType struct {
	Type semantic.Type
}

func (e ErrUnsupportedType) Error() string {
	return fmt.Sprintf("Unsupported Type: %s", e.Type)
}

type ErrExpectedType struct {
	Type semantic.Type
}

func (e ErrExpectedType) Error() string {
	return fmt.Sprintf("Expected Type: %s", e.Type)
}

type AnalyzerError string

func (pe AnalyzerError) Error() string {
	return string(pe)
}

const (
	ErrUnexpectedVoid = "Expression doesn't return anything"
	ErrLValueError    = "This expression is not assignable"
	ErrImmutable      = "Expression is immutable"
)
