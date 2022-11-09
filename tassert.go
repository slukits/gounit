// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
)

// trueErr default message for failed 'true'-assertion.
const trueErr = "expected given value to be true"

// True fails the test and returns false iff given value is not true;
// otherwise true is returned.
func (t T) True(value bool) bool {
	t.t.Helper()
	if !value {
		t.Errorf(assertErr, "true", trueErr)
		return false
	}
	return true
}

// TODO fails a test logging "not implemented yet".
func (t T) TODO() bool {
	t.t.Helper()
	t.Error("not implemented yet")
	return false
}

// notTrueErr default message for failed 'false'-assertion.
const notTrueErr = "expected given value be false"

// True passes if called [T.True] assertion with given argument fails;
// otherwise it fails.
func (n Not) True(value bool) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.True(value)
	n.t.errorer = err
	if passed {
		n.t.Errorf(assertErr, "not-true", notTrueErr)
		return false
	}
	return true
}

const eqTypeErr = "types mismatch %v != %v"

// Eq errors with an corresponding diff if possible and returns false if
// given values are not considered equal; otherwise true is returned.  a
// and b are considered equal if they are of the same type or one of
// them is string while the other one is a Stringer implementation and
//   - a == b in case of two pointers
//   - a == b in case of two strings
//   - a.String() == b.String() in case of Stringer implementations
//   - a == b.Stringer() or a.Stringer() == b in case of string and
//     Stringer implementation.
//   - fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b) in other cases
//
// if they are not of the same type or one of the above cases replacing
// "==" by "!=" is true then given values are considered not equal.
func (t T) Eq(a, b interface{}) bool {
	t.t.Helper()

	differentTypes := fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b)
	if differentTypes && !isStringers(a, b) {
		t.Errorf(assertErr, "equal: types", fmt.Sprintf(
			eqTypeErr, fmt.Sprintf("%T", a), fmt.Sprintf("%T", b)))
		return false
	}

	if reflect.ValueOf(a).Kind() == reflect.Ptr {
		if a != b {
			t.Errorf(assertErr, "equal: pointer", fmt.Sprintf("%p != %p", a, b))
			return false
		}
		return true
	}

	diff := diff(a, b, differentTypes)
	if diff != "" {
		t.Errorf(assertErr, "equal: string-representations", diff)
		return false
	}

	return true
}

func isStringers(a, b interface{}) bool {
	_, okA := a.(fmt.Stringer)
	_, okB := b.(fmt.Stringer)
	if !okA && !okB {
		return false
	}
	if okA && okB {
		return true
	}
	if okA {
		_, ok := b.(string)
		return ok
	}
	_, ok := a.(string)
	return ok
}

func diff(a, b interface{}, differentTypes bool) string {
	if differentTypes {
		if _a, ok := a.(fmt.Stringer); ok {
			a = _a.String()
		}
		if _b, ok := b.(fmt.Stringer); ok {
			b = _b.String()
		}
	}
	diff := ""
	switch a := a.(type) {
	case string:
		if a != b.(string) {
			diff = cmp.Diff(a, b.(string))
		}
	case fmt.Stringer:
		if a.String() != b.(fmt.Stringer).String() {
			diff = cmp.Diff(a.String(), b.(fmt.Stringer).String())
		}
	default:
		if fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b) {
			diff = cmp.Diff(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
		}
	}
	return diff
}

// Eq negation passes if called [T.Eq] assertion with given arguments
// fails; otherwise it fails.
func (n Not) Eq(a, b interface{}) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.Eq(a, b)
	n.t.errorer = err
	if passed {
		n.t.Errorf(assertErr, "not-equal", fmt.Sprintf("%p == %p", a, b))
		return false
	}
	return true
}

// containsErr default message for failed 'Contains'-assertion.
const containsErr = "%s doesn't contain %s"

// StringRepresentation documents what a string representation of any
// type is:
//   - the string if it is of type string,
//   - the return value of String if the Stringer interface is
//     implemented,
//   - fmt.Sprintf("%v", value) in all other cases.
type StringRepresentation interface{}

// Contains fails the test and returns false iff given value's string
// representation doesn't contain given sub-string; otherwise true is
// returned.
func (t T) Contains(value StringRepresentation, sub string) bool {
	t.t.Helper()
	str := toString(value)
	if !strings.Contains(str, sub) {
		if !strings.HasSuffix(str, "\n") {
			str += "\n"
		}
		if !strings.HasPrefix(sub, "\n") {
			sub = "\n" + sub
		}
		t.Errorf(assertErr, "contains", fmt.Sprintf(containsErr, str, sub))
		return false
	}
	return true
}

func toString(value interface{}) string {
	switch value := value.(type) {
	case string:
		return value
	case fmt.Stringer:
		return value.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

// Not implements negations of [T]-assertions, e.g. [Not.True].  Negated
// assertions can be accessed through [T]'s Not field.
type Not struct{ t *T }

// notContainsErr default message for failed Not-'Contains'-assertion.
const notContainsErr = "\n'%s'\ndoes contain\n'%s'"

// Contains negation passes if called [T.Contains] assertion with given
// arguments fails; otherwise it fails.
func (n Not) Contains(value StringRepresentation, sub string) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.Contains(value, sub)
	n.t.errorer = err
	if passed {
		n.t.Errorf(assertErr, "does't contain",
			fmt.Sprintf(notContainsErr, toString(value), sub))
		return false
	}
	return true
}

// matchedErr default message for failed *'Matched'-assertion.
const matchedErr = "Regexp\n'%s'\ndoesn't match\n'%s'"

// Matched fails the test and returns false iff given values string
// interpretation isn't matched by given regex; otherwise true is
// returned.
func (t T) Matched(value StringRepresentation, regex string) bool {
	t.t.Helper()
	str := toString(value)
	re := regexp.MustCompile(regex)
	if !re.MatchString(str) {
		t.Errorf(assertErr, "matched",
			fmt.Sprintf(matchedErr, re.String(), str))
		return false
	}
	return true
}

// notMatchedErr default message for failed *'Matched'-assertion.
const notMatchedErr = "Regexp '%s'\n matches '%s'"

// Matched negation passes if called [T.Matched] assertion with given
// arguments fails; otherwise it fails.
func (n Not) Matched(value StringRepresentation, regex string) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.Matched(value, regex)
	n.t.errorer = err
	if passed {
		n.t.Errorf(assertErr, "don't-match",
			fmt.Sprintf(matchedErr, regex, toString(value)))
		return false
	}
	return true
}

// SpaceMatched escapes given variadic strings before it joins them with
// the `\s*`-separator and matches the result against given value's
// string representation, e.g.:
//
//	<p>
//	   some text
//	</p>
//
// would be matched by
//
//	t.SpaceMatched(value, "<p>", "some text", "</p>").
//
// SpaceMatched fails the test and returns false iff the matching
// fails; otherwise true is returned.
func (t T) SpaceMatched(value StringRepresentation, ss ...string) bool {
	t.t.Helper()
	spaceRe, str := reGen(`\s*`, "", ss...), toString(value)
	if !spaceRe.MatchString(str) {
		t.Errorf(assertErr, "space-match", fmt.Sprintf(
			matchedErr, spaceRe.String(), str))
		return false
	}
	return true
}

// SpaceMatched negation passes if called [T.SpaceMatched] assertion with given
// arguments fails; otherwise it fails.
func (n Not) SpaceMatched(value StringRepresentation, ss ...string) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.SpaceMatched(value, ss...)
	n.t.errorer = err
	if passed {
		spaceRe := reGen(`\s*`, "", ss...)
		n.t.Errorf(assertErr, "not: space-match", fmt.Sprintf(
			notMatchedErr, spaceRe.String(), toString(value)))
		return false
	}
	return true
}

// StarMatched escapes given variadic-strings before it joins them with
// the `.*?`-separator and matches the result against given value's
// string representation, e.g.:
//
//	<p>
//	   some text
//	</p>
//
// would be matched by
//
//	t.StarMatch(str, "p", "me", "x", "/p").
//
// StarMatched fails the test and returns false iff the matching
// fails; otherwise true is returned.
func (t T) StarMatched(value StringRepresentation, ss ...string) bool {
	t.t.Helper()
	startRe, str := reGen(`.*?`, `(?s)`, ss...), toString(value)
	if !startRe.MatchString(str) {
		t.Errorf(assertErr, "star-match", fmt.Sprintf(
			matchedErr, startRe.String(), str))
		return false
	}
	return true
}

// StarMatched passes if called T.StarMatched assertion with given
// arguments fails; otherwise it fails.
func (n Not) StarMatched(value StringRepresentation, ss ...string) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.StarMatched(value, ss...)
	n.t.errorer = err
	if passed {
		startRe := reGen(`.*?`, `(?s)`, ss...)
		n.t.Errorf(assertErr, "not: star-match", fmt.Sprintf(
			notMatchedErr, startRe.String(), toString(value)))
		return false
	}
	return true
}

func reGen(sep string, flags string, ss ...string) *regexp.Regexp {
	quoted := []string{}
	for _, s := range ss {
		if strings.Contains(s, "\n") {
			for _, line := range strings.Split(s, "\n") {
				quoted = append(
					quoted, regexp.QuoteMeta(strings.TrimSpace(line)))
			}
			continue
		}
		quoted = append(quoted, regexp.QuoteMeta(s))
	}
	return regexp.MustCompile(flags + strings.Join(quoted, sep))
}

// errErr default message for failed "Err"-assertion
const errErr = "given value doesn't implement 'error'"

// Err fails the test and returns false iff given value doesn't
// implement the error-interface; otherwise true is returned.
func (t T) Err(err interface{}) bool {

	t.t.Helper()

	_, ok := err.(error)
	if !ok {
		t.Errorf(assertErr, "error", errErr)
		return false
	}
	return true
}

// errIsErr default message for failed "ErrIs"-assertion
const errIsErr = "given error doesn't wrap target-error"

// ErrIs fails the test and returns false iff given err doesn't
// implement the error-interface or doesn't wrap given target; otherwise
// true is returned.
func (t T) ErrIs(err interface{}, target error) bool {
	t.t.Helper()

	e, ok := err.(error)
	if !ok {
		t.Errorf(assertErr, "error is", errIsErr)
		return false
	}
	if errors.Is(e, target) {
		return true
	}
	t.Errorf(assertErr, "error is",
		fmt.Sprintf("%s: %+v\n%+v", errIsErr, e, target))
	return false
}

// errMatchedErr default message for failed "ErrMatched"-assertion
const errMatchedErr = "given regexp '%s' doesn't match '%s'"

// ErrMatched fails the test and returns false iff given err doesn't
// implement the error-interface or its message isn't matched by given
// regex; otherwise true is returned.
func (t T) ErrMatched(err interface{}, re string) bool {

	t.t.Helper()

	e, ok := err.(error)
	if !ok {
		t.Error(assertErr, "error matched", errMatchedErr)
		return false
	}

	re = strings.ReplaceAll(re, "%s", ".*?")
	regexp := regexp.MustCompile(re)
	if !regexp.MatchString(e.Error()) {
		t.Errorf(assertErr, "error matched", fmt.Sprintf(
			errMatchedErr, re, e.Error()))
		return false
	}

	return true
}

// panicsErr default message for failed "Panics"-assertion
const panicsErr = "given function doesn't panic"

// Panics fails the test and returns false iff given function doesn't
// panic; otherwise true is returned.
func (t T) Panics(f func()) (hasPanicked bool) {
	t.t.Helper()
	defer func() {
		t.t.Helper()
		if r := recover(); r == nil {
			t.Errorf(assertErr, "panics", panicsErr)
			hasPanicked = false
			return
		}
		hasPanicked = true
	}()
	f()
	return true
}

// withinErr default message for failed "Within"-assertion
const withinErr = "timeout while condition unfulfilled"

// Within tries after each step of given time-stepper if given condition
// returns true and fails the test iff the whole duration of given time
// stepper is elapsed without given condition returning true.
func (t T) Within(d *TimeStepper, cond func() bool) (fulfilled bool) {
	done := make(chan bool)
	go func(c chan bool) {
		time.Sleep(d.Step())
		if cond() {
			c <- true
			return
		}
		for d.AddStep() {
			time.Sleep(d.Step())
			if cond() {
				c <- true
				return
			}
		}
		c <- false
	}(done)
	t.t.Helper()
	if success := <-done; !success {
		t.Errorf(assertErr, "within", withinErr)
		return false
	}
	return true
}

// assertErr is the format-string for assertion errors.
const assertErr = "assert %s:\n%v"
