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

// True errors the test and returns false iff given value is not true;
// otherwise true is returned.
func (t *T) True(value bool) bool {
	t.t.Helper()
	if !value {
		t.Errorf(assertErr, "true", trueErr)
		return false
	}
	return true
}

func (t *T) TODO() bool {
	t.t.Helper()
	t.Error("not implemented yet")
	return false
}

// notTrueErr default message for failed 'false'-assertion.
const notTrueErr = "expected given value be not true"

// True passes if called true assertion with given argument fails;
// otherwise it errors.
func (n *not) True(value bool) bool {
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
// and b are considered equal if they are of the same type and
//   - a == b in case of two pointers
//   - a == b in case of two strings
//   - a.String() == b.String() in case of Stringer implementations
//   - fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b) in other cases
//
// if they are not of the same type or one of the above cases replacing
// "==" by "!=" is true then given values are considered not equal.
func (t *T) Eq(a, b interface{}) bool {
	t.t.Helper()

	if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
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

	diff := t.diff(a, b)
	if diff != "" {
		t.Errorf(assertErr, "equal: string-representations", diff)
		return false
	}

	return true
}

func (t *T) diff(a, b interface{}) string {
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

// Eq passes if called equals assertion with given arguments fails;
// otherwise it errors.
func (n *not) Eq(a, b interface{}) bool {
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

// Contains errors the test and returns false iff given string doesn't
// contain given sub-string; otherwise true is returned.
func (t *T) Contains(str, sub string) bool {
	t.t.Helper()
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

// notContainsErr default message for failed Not-'Contains'-assertion.
const notContainsErr = "%s does contain %s"

// Contains passes if called contains assertion with given arguments fails;
// otherwise it errors.
func (n *not) Contains(str, sub string) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.Contains(str, sub)
	n.t.errorer = err
	if passed {
		n.t.Errorf(assertErr, "does't contain",
			fmt.Sprintf(notContainsErr, str, sub))
		return false
	}
	return true
}

// matchedErr default message for failed *'Matched'-assertion.
const matchedErr = "Regexp '%s'\ndoesn't match '%s'"

// Matched errors the test and returns false iff given string isn't
// matched by given regex; otherwise true is returned.
func (t *T) Matched(str, regex string) bool {
	t.t.Helper()
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

// Matched passes if called match assertion with given arguments fails;
// otherwise it errors.
func (n *not) Matched(str string, regex string) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.Matched(str, regex)
	n.t.errorer = err
	if passed {
		n.t.Errorf(assertErr, "don't-match",
			fmt.Sprintf(matchedErr, regex, str))
		return false
	}
	return true
}

// SpaceMatched escapes given variadic strings before it joins them with
// the `\s*`-separator and matches the result against given string str:
//
//	<p>
//	   some text
//	</p>
//
// would be matched by t.SpaceMatched(str, "<p>", "some text", "</p>").
// SpaceMatched errors the test and returns false iff the matching
// fails; otherwise true is returned.
func (t *T) SpaceMatched(str string, ss ...string) bool {

	t.t.Helper()
	spaceRe := reGen(`\s*`, "", ss...)
	if !spaceRe.MatchString(str) {
		t.Errorf(assertErr, "space-match", fmt.Sprintf(
			matchedErr, spaceRe.String(), str))
		return false
	}
	return true
}

// SpaceMatch passes if called space match assertion with given
// arguments fails; otherwise it errors.
func (n *not) SpaceMatched(str string, ss ...string) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.SpaceMatched(str, ss...)
	n.t.errorer = err
	if passed {
		spaceRe := reGen(`\s*`, "", ss...)
		n.t.Errorf(assertErr, "not: space-match",
			fmt.Sprintf(notMatchedErr, spaceRe.String(), str))
		return false
	}
	return true
}

// StarMatched escapes given variadic-strings before it joins them with
// the `.*?`-separator and matches the result against given string str:
//
//	<p>
//	   some text
//	</p>
//
// would be matched by t.StarMatch(str, "p", "me", "x", "/p").
// SpaceMatched errors the test and returns false iff the matching
// fails; otherwise true is returned.
func (t *T) StarMatched(str string, ss ...string) bool {
	t.t.Helper()
	startRe := reGen(`.*?`, `(?s)`, ss...)
	if !startRe.MatchString(str) {
		t.Errorf(assertErr, "star-match", fmt.Sprintf(
			matchedErr, startRe.String(), str))
		return false
	}
	return true
}

// StarMatched passes if called star match assertion with given
// arguments fails; otherwise it errors.
func (n *not) StarMatched(str string, ss ...string) bool {
	n.t.t.Helper()
	err := n.t.errorer
	n.t.errorer = func(i ...interface{}) {}
	passed := n.t.StarMatched(str, ss...)
	n.t.errorer = err
	if passed {
		startRe := reGen(`.*?`, `(?s)`, ss...)
		n.t.Errorf(assertErr, "not: star-match", fmt.Sprintf(
			notMatchedErr, startRe.String(), str))
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

// Err errors the test and returns false iff given values doesn't
// implement the error-interface; otherwise true is returned.
func (t *T) Err(err interface{}) bool {

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

// ErrIs errors the test and returns false iff given err doesn't
// implement the error-interface or doesn't wrap given target; otherwise
// true is returned.
func (t *T) ErrIs(err interface{}, target error) bool {
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

// ErrMatched errors the test and returns false iff given err doesn't
// implement the error-interface or its message isn't matched by given
// regex; otherwise true is returned.
func (t *T) ErrMatched(err interface{}, re string) bool {

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

// Panics errors the test and returns false iff given function doesn't
// panic; otherwise true is returned.  Given (formatted) message
// replaces the default error message, i.e. msg[0] must be a string if
// len(msg) == 1 it must be a format-string iff len(msg) > 1.
func (t *T) Panics(f func(), msg ...interface{}) (hasPanicked bool) {
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
func (t *T) Within(d *TimeStepper, cond func() bool) (fulfilled bool) {
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
	// select {
	// case success := <-done:
	// 	if success {
	// 		return
	// 	}
	// 	t.Errorf(assertErr, "within", withinErr)
	// 	return false
	// case <-t.Timeout(2 * d.Duration()):
	// 	panic("gounit: within: should have concluded by now")
	// }
	return true
}

// assertErr is the format-string for assertion errors.
const assertErr = "assert %s: %v"
