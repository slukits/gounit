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

// falseErr default message for failed 'false'-assertion.
const falseErr = "expected given value to be false"

// False errors the test and returns false iff given value is not false;
// otherwise true is returned.
func (t *T) False(value bool) bool {
	t.t.Helper()
	if value {
		t.Errorf(assertErr, "false", falseErr)
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

// Neq errors and returns false if given values considered equal see [T.Eq];
// otherwise true is returned.
func (t *T) Neq(a, b interface{}, msg ...interface{}) bool {
	t.t.Helper()
	if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
		return true
	}

	if reflect.ValueOf(a).Kind() == reflect.Ptr {
		if a == b {
			t.Errorf(assertErr, "not-equal", fmt.Sprintf("%p == %p", a, b))
			return false
		}
		return true
	}

	err := ""
	switch a := a.(type) {
	case string:
		if a == b.(string) {
			err = "given strings are equal"
		}
	case fmt.Stringer:
		if a.String() == b.(fmt.Stringer).String() {
			err = "given Stringer return equal strings"
		}
	default:
		if fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b) {
			err = "given instances have same fmt-string representations"
		}
	}

	if err == "" {
		return true
	}
	t.Errorf(assertErr, "not-equal", err)
	return false
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
		t.Errorf(assertErr, "star-match", fmt.Sprintf(
			matchedErr, spaceRe.String(), str))
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
// stepper is elapsed without given condition returning true.  Use the
// returned channel to wait for either the fulfillment of the condition
// or the failing timeout.
func (t *T) Within(d *TimeStepper, cond func() bool) chan bool {
	done := make(chan bool)
	go func() {
		time.Sleep(d.Step())
		if cond() {
			done <- true
			return
		}
		for d.AddStep() {
			time.Sleep(d.Step())
			if cond() {
				done <- true
				return
			}
		}
		t.Errorf(assertErr, "within", withinErr)
		done <- false
	}()
	return done
}

// assertErr is the format-string for assertion errors.
const assertErr = "assert %s: %v"
