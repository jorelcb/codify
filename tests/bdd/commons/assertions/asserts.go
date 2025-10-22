package assertions

import (
	"fmt"
	"reflect"

	"github.com/stretchr/testify/assert"
)

// asserter implements assert.TestingT to capture assertion errors
type asserter struct {
	err error
}

func (a *asserter) Errorf(format string, args ...interface{}) {
	a.err = fmt.Errorf(format, args...)
}

// expectedAndActualAssertion is a function type for assertions that compare expected vs actual
type expectedAndActualAssertion func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool

// AssertExpectedAndActual wraps testify assertions to return errors instead of failing tests
func AssertExpectedAndActual(a expectedAndActualAssertion, expected, actual interface{}, msgAndArgs ...interface{}) error {
	var t asserter
	a(&t, expected, actual, msgAndArgs...)
	return t.err
}

// actualAssertion is a function type for assertions that check a single value
type actualAssertion func(t assert.TestingT, actual interface{}, msgAndArgs ...interface{}) bool

// AssertActual wraps testify assertions for single-value checks
func AssertActual(a actualAssertion, actual interface{}, msgAndArgs ...interface{}) error {
	var t asserter
	a(&t, actual, msgAndArgs...)
	return t.err
}

// AssertType checks if two values have the same type
func AssertType(a expectedAndActualAssertion, expected, actual interface{}, msgAndArgs ...interface{}) error {
	expectedType := reflect.TypeOf(expected).String()
	actualType := reflect.TypeOf(actual).String()
	return AssertExpectedAndActual(a, expectedType, actualType, msgAndArgs...)
}

// boolAssertion is a function type for boolean assertions
type boolAssertion func(t assert.TestingT, value bool, msgAndArgs ...interface{}) bool

// AssertBool wraps boolean testify assertions
func AssertBool(a boolAssertion, value bool, msgAndArgs ...interface{}) error {
	var t asserter
	a(&t, value, msgAndArgs...)
	return t.err
}