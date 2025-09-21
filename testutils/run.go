package testutils

import (
	"testing"

	"darvaza.org/core"
)

// Run executes a sub-test with the given name using the appropriate Run method
// based on the type of the testing interface. This function provides a unified
// way to create sub-tests that work with both *testing.T and core.T implementations.
//
// The function uses type assertions to determine which Run method signature is
// available and calls it accordingly. For simple core.T implementations that don't
// support sub-tests, the function uses RunTest to execute the test function via sandboxed testing.RunTests
// and reports results back to the original testing interface.
//
// Returns true if the sub-test passed, false if it failed. This allows callers
// to respond to sub-test failures programmatically.
func Run(t core.T, name string, fn func(core.T)) bool {
	switch tt := t.(type) {
	case interface {
		Run(string, func(*testing.T)) bool
	}:
		return tt.Run(name, func(subT *testing.T) { fn(subT) })
	case interface {
		Run(string, func(core.T)) bool
	}:
		return tt.Run(name, fn)
	default:
		// Use testing.RunTests via RunTest.
		ok := RunTest(name, func(subT *testing.T) { fn(subT) })
		if !ok {
			t.Errorf("Test case %q failed", name)
		} else {
			t.Logf("Test case %q passed", name)
		}
		return ok
	}
}

// RunSuccessCases executes a slice of test cases that are expected to pass,
// using RunTest for sandboxed execution to verify their behaviour. This function is primarily used
// for testing the TestCase implementations themselves, ensuring they correctly
// report success when given valid inputs.
//
// Each test case is converted to testing.InternalTest and executed via RunTest using sandboxed testing.RunTests.
// If a test case fails when it should have passed, an error is reported to the
// provided testing interface. The function works directly with core.TestCase interface.
func RunSuccessCases(t core.T, cases []core.TestCase) {
	t.Helper()

	doRunTests(t, cases, func(subT core.T, name string, ok bool) {
		subT.Helper()
		if ok {
			subT.Logf("TestCase %q passed", name)
		} else {
			subT.Errorf("TestCase %q should have passed but failed", name)
		}
	})
}

// RunFailureCases executes a slice of test cases that are expected to fail,
// using RunTest for sandboxed execution to verify their behaviour. This function is primarily used
// for testing the TestCase implementations themselves, ensuring they correctly
// detect and report failures when given invalid inputs.
//
// Each test case is converted to testing.InternalTest and executed via RunTest using sandboxed testing.RunTests.
// If a test case passes when it should have failed, an error is reported to the
// provided testing interface. The function works directly with core.TestCase interface.
func RunFailureCases(t core.T, cases []core.TestCase) {
	t.Helper()

	doRunTests(t, cases, func(subT core.T, name string, ok bool) {
		subT.Helper()
		if !ok {
			subT.Logf("TestCase %q failed as expected", name)
		} else {
			subT.Errorf("TestCase %q should have failed but passed", name)
		}
	})
}

// doRunTests executes core.TestCase cases via RunTest (sandboxed testing.RunTests) and reports
// results using the provided callback. Tests each case individually to enable
// per-test reporting.
func doRunTests(t core.T, cases []core.TestCase, reportResults func(t core.T, name string, ok bool)) {
	t.Helper()

	for _, tc := range cases {
		testName := tc.Name()
		ok := RunTest(testName, tc.Test)
		reportResults(t, testName, ok)
	}
}
