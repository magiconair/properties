// Copyright 2016 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package assert provides a set of assert functions for testing.
package assert

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// skip defines the default call depth
const skip = 1

// Equal asserts that got and want are equal as defined by
// reflect.DeepEqual. The test fails with msg if they are not equal.
func Equal(t *testing.T, got, want interface{}, msg ...string) {
	_, file, line, _ := runtime.Caller(skip)
	if !reflect.DeepEqual(got, want) {
		fail(t, file, line, "got %v want %v %s", got, want, strings.Join(msg, " "))
	}
}

// Panic asserts that function fn() panics.
// It assumes that recover() either returns a string or
// an error and fails if the message does not match
// the regular expression in 'matches'.
func Panic(t *testing.T, fn func(), matches string) {
	_, file, line, _ := runtime.Caller(skip)
	defer func() {
		r := recover()
		if r == nil {
			fail(t, file, line, "did not panic")
		}
		switch r.(type) {
		case error:
			match(t, r.(error).Error(), matches, file, line)
		case string:
			match(t, r.(string), matches, file, line)
		}
	}()
	fn()
}

// Matches asserts that a value matches a given regular expression.
func Matches(t *testing.T, value, expr string) {
	_, file, line, _ := runtime.Caller(skip)
	match(t, value, expr, file, line)
}

func match(t *testing.T, value, expr string, file string, line int) {
	ok, err := regexp.MatchString(expr, value)
	if err != nil {
		fail(t, file, line, "invalid pattern %q. %s", expr, err)
	}
	if !ok {
		fail(t, file, line, "got %s which does not match %s", value, expr)
	}
}

func fail(t *testing.T, file string, line int, format string, args ...interface{}) {
	fmt.Printf("\t%s:%d: %s\n", filepath.Base(file), line, fmt.Sprintf(format, args...))
	t.Fail()
}
