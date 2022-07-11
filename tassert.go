// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TrueErr default message for failed 'true'-assertion.
const TrueErr = "expected given value to be true"

// True errors the test and returns false iff given value is not true;
// otherwise true is returned.  Given (formatted) message replaces the
// default error message, i.e. msg[0] must be a string if len(msg) == 1
// it must be a format-string iff len(msg) > 1.
func (t *T) True(value bool, msg ...interface{}) bool {
	t.t.Helper()
	if true != value {
		t.Error(assertErr("true", TrueErr, msg...))
		return false
	}
	return true
}

// FalseErr default message for failed 'false'-assertion.
const FalseErr = "expected given value to be false"

// False errors the test and returns false iff given value is not false;
// otherwise true is returned.  Given (formatted) message replaces the
// default error message, i.e. msg[0] must be a string if len(msg) == 1
// it must be a format-string iff len(msg) > 1.
func (t *T) False(value bool, msg ...interface{}) bool {
	t.t.Helper()
	if false != value {
		t.Error(assertErr("false", FalseErr, msg...))
		return false
	}
	return true
}

// Eq errors the test and returns false iff given values diff is not
// empty; otherwise true is returned.  Given (formatted) message replaces
// the default error message, i.e. msg[0] must be a string if len(msg)
// == 1 it must be a format-string iff len(msg) > 1.
func (t *T) Eq(a, b interface{}, msg ...interface{}) bool {
	t.t.Helper()
	diff := cmp.Diff(a, b)
	if diff == "" {
		return true
	}
	t.Error(assertErr("equal", diff, msg...))
	return false
}

// NeqErr default message for failed 'Neq'-assertion.
const NeqErr = "expected given values to differ"

// Neq errors the test and returns false iff given values diff is empty;
// otherwise true is returned.  Given (formatted) message replaces the
// default error message, i.e. msg[0] must be a string if len(msg) == 1
// it must be a format-string iff len(msg) > 1.
func (t *T) Neq(a, b interface{}, msg ...interface{}) bool {
	t.t.Helper()
	if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
		return true
	}
	diff := cmp.Diff(a, b)
	if diff != "" {
		return true
	}
	t.Error(assertErr("not-equal", NeqErr, msg...))
	return false
}

// ContainsErr default message for failed 'Contains'-assertion.
const ContainsErr = "%s doesn't contain %s"

// Contains errors the test and returns false iff given string doesn't
// contain given sub-string; otherwise true is returned.  Given
// (formatted) message replaces the default error message, i.e. msg[0]
// must be a string if len(msg) == 1 it must be a format-string iff
// len(msg) > 1.
func (t *T) Contains(str, sub string, msg ...interface{}) bool {
	t.t.Helper()
	if !strings.Contains(str, sub) {
		if !strings.HasSuffix(str, "\n") {
			str += "\n"
		}
		if !strings.HasPrefix(sub, "\n") {
			sub = "\n" + sub
		}
		t.Error(assertErr("contains", fmt.Sprintf(
			ContainsErr, str, sub), msg...))
		return false
	}
	return true
}

// MatchedErr default message for failed *'Matched'-assertion.
const MatchedErr = "Regexp '%s'\ndoesn't match '%s'"

// Matched errors the test and returns false iff given string isn't
// matched by given regex; otherwise true is returned.  Given
// (formatted) message replaces the default error message, i.e. msg[0]
// must be a string if len(msg) == 1 it must be a format-string iff
// len(msg) > 1.
func (t *T) Matched(str, regex string, msg ...interface{}) bool {
	t.t.Helper()
	re := regexp.MustCompile(regex)
	if !re.MatchString(str) {
		t.Error(assertErr("matched", fmt.Sprintf(
			MatchedErr, re.String(), str), msg...))
		return false
	}
	return true
}

// SpaceMatched escapes given slice-strings before it joins them with
// the `\s*`-separator and matches the result against given string:
//
//     <p>
//        some text
//     </p>
//
// would be matched by []string{"<p>", "some text", "</p>"}.
// SpaceMatched errors the test and returns false iff the matching
// fails; otherwise true is returned.   Given (formatted) message
// replaces the default error message, i.e. msg[0] must be a string if
// len(msg) == 1 it must be a format-string iff len(msg) > 1.
func (t *T) SpaceMatched(str string, ss []string,
	msg ...interface{}) bool {

	t.t.Helper()
	spaceRe := reGen(`\s*`, "", ss...)
	if !spaceRe.MatchString(str) {
		t.Error(assertErr("star-match", fmt.Sprintf(
			MatchedErr, spaceRe.String(), str), msg...))
		return false
	}
	return true
}

// StarMatched escapes given slice-strings before it joins them with
// the `.*?`-separator and matches the result against given string:
//
//     <p>
//        some text
//     </p>
//
// would be matched by []string{"p", "me", "x", "/p"}.
// SpaceMatched errors the test and returns false iff the matching
// fails; otherwise true is returned.   Given (formatted) message
// replaces the default error message, i.e. msg[0] must be a string if
// len(msg) == 1 it must be a format-string iff len(msg) > 1.
func (t *T) StarMatched(
	str string, ss []string, msg ...interface{},
) bool {
	t.t.Helper()
	startRe := reGen(`.*?`, `(?s)`, ss...)
	if !startRe.MatchString(str) {
		t.Error(assertErr("star-match", fmt.Sprintf(
			MatchedErr, startRe.String(), str), msg...))
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

// ErrErr default message for failed "Err"-assertion
const ErrErr = "given value doesn't implement 'error'"

// Err errors the test and returns false iff given values doesn't
// implement the error-interface; otherwise true is returned.  Given
// (formatted) message replaces the default error message, i.e. msg[0]
// must be a string if len(msg) == 1 it must be a format-string iff
// len(msg) > 1.
func (t *T) Err(err interface{}, msg ...interface{}) bool {

	t.t.Helper()

	_, ok := err.(error)
	if !ok {
		t.Error(assertErr("error", ErrErr, msg...))
		return false
	}
	return true
}

// ErrIsErr default message for failed "ErrIs"-assertion
const ErrIsErr = "given error doesn't wrap target-error"

// ErrIs errors the test and returns false iff given err doesn't
// implement the error-interface or doesn't wrap given target; otherwise
// true is returned.  Given (formatted) message replaces the default
// error message, i.e. msg[0] must be a string if len(msg) == 1 it must
// be a format-string iff len(msg) > 1.
func (t *T) ErrIs(
	err interface{}, target error, msg ...interface{},
) bool {
	t.t.Helper()

	e, ok := err.(error)
	if !ok {
		t.Error(assertErr("error is", ErrIsErr, msg...))
		return false
	}
	if errors.Is(e, target) {
		return true
	}
	t.Error(assertErr("error is", fmt.Sprintf(
		"%s: %+v\n%+v", ErrIsErr, e, target), msg...))
	return false
}

// ErrMatchedErr default message for failed "ErrMatched"-assertion
const ErrMatchedErr = "given regexp '%s' doesn't match '%s'"

// ErrMatched errors the test and returns false iff given err doesn't
// implement the error-interface or its message isn't matched by given
// regex; otherwise true is returned.  Given (formatted) message replaces
// the default error message, i.e. msg[0] must be a string if len(msg)
// == 1 it must be a format-string iff len(msg) > 1.
func (t *T) ErrMatched(err interface{}, re string,
	msg ...interface{}) bool {

	t.t.Helper()

	e, ok := err.(error)
	if !ok {
		t.Error(assertErr("error matched", ErrMatchedErr, msg...))
		return false
	}

	re = strings.ReplaceAll(re, "%s", ".*?")
	regexp := regexp.MustCompile(re)
	if !regexp.MatchString(e.Error()) {
		t.Error(assertErr("error matched", fmt.Sprintf(
			ErrMatchedErr, re, e.Error()), msg...))
		return false
	}

	return true
}

// PanicsErr default message for failed "Panics"-assertion
const PanicsErr = "given function doesn't panic"

// Panics errors the test and returns false iff given function doesn't
// panic; otherwise true is returned.  Given (formatted) message
// replaces the default error message, i.e. msg[0] must be a string if
// len(msg) == 1 it must be a format-string iff len(msg) > 1.
func (t *T) Panics(f func(), msg ...interface{}) (hasPanicked bool) {
	t.t.Helper()
	defer func() {
		if recover() == nil {
			t.Error(assertErr("panics", PanicsErr, msg...))
			hasPanicked = false
			return
		}
		hasPanicked = true
	}()
	f()
	return true
}

// WithinErr default message for failed "Within"-assertion
const WithinErr = "timeout while condition unfulfilled"

// Within tries after each step of given time-stepper if given condition
// returns true and fails the test iff the whole duration of given time
// stepper is elapsed without given condition returning true.  Use the
// returned channel to wait for either the fulfillment of the condition
// or the failing timeout.
func (t *T) Within(
	d *TimeStepper, cond func() bool, mm ...interface{},
) chan bool {
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
		t.Error(assertErr("within", WithinErr, mm...))
		done <- false
	}()
	return done
}

// Assert is the format-string for assertion errors.
const Assert = "assert %s: %v"

// FormatMsgErr is the error message in case an assertion was called
// with more then one optional message argument whereas the first
// argument is not a (format-)string.
const FormatMsgErr = "expected first message-argument to be string; got %T"

func assertErr(label, msg string, args ...interface{}) string {
	if len(args) == 0 {
		return fmt.Sprintf(Assert, label, msg)
	}
	if len(args) == 1 {
		return fmt.Sprintf(Assert, label, args[0])
	}
	ext, ok := args[0].(string)
	if !ok {
		return fmt.Sprintf("%s: %s", fmt.Sprintf(Assert, label, msg),
			fmt.Sprintf(FormatMsgErr, args[0]))
	}
	return fmt.Sprintf(Assert, label, fmt.Sprintf(ext, args[1:]...))
}
