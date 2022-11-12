package test

import (
	"strings"
	"testing"
)

// methods to help with assertions
// a basic version of testify!

func EqualStr(t *testing.T, expected, actual, message string) {
	if expected != actual {
		t.Errorf(`
Not equal:
expected: %s
actual  : %s
message : %s`, expected, actual, message)
	}
}

func Contains(t *testing.T, expectedSubStr, actual, message string) {
	if !strings.Contains(actual, expectedSubStr) {
		t.Errorf(`
Is not contained within string:
expected substring: %s
actual            : %s
message           : %s`, expectedSubStr, actual, message)
	}
}

func EqualInt(t *testing.T, expected, actual int, message string) {
	if expected != actual {
		t.Errorf(`
Not equal:
expected: %d
actual  : %d
message : %s`, expected, actual, message)
	}
}

func IsTrue(t *testing.T, actual bool, message string) {
	if !actual {
		t.Errorf(`
Not true:
expected to be true, but was false
message : %s`, message)
	}
}

func IsFalse(t *testing.T, actual bool, message string) {
	if actual {
		t.Errorf(`
Not false:
expected to be false, but was true
message : %s`, message)
	}
}

func NoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(`
Error:
expected to be nil, but was an error
error   : %s
message : %s`, err, message)
	}
}

func HasError(t *testing.T, err error, message string) {
	if err == nil {
		t.Errorf(`
Error was expected:
expected to not be nil, but was nil
message : %s`, message)
	}
}

func NotNil(t *testing.T, obj interface{}, message string) {
	if obj == nil {
		t.Errorf(`
Error:
expected to not be nil, but was nil
message : %s`, message)
	}
}
