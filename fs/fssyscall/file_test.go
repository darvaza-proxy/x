package fssyscall_test

import (
	"os"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/fs/fssyscall"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = fileLockTestCase{}

type fileLockTestCase struct {
	expectedErr error
	fn          func(*os.File) error
	file        *os.File
	name        string
}

func (tc fileLockTestCase) Name() string {
	return tc.name
}

func (tc fileLockTestCase) Test(t *testing.T) {
	t.Helper()

	err := tc.fn(tc.file)

	if tc.expectedErr != nil {
		core.AssertError(t, err, "operation error")
		core.AssertErrorIs(t, err, tc.expectedErr, "error type")
	} else {
		core.AssertNoError(t, err, "operation")
	}
}

func newFileLockTestCase(name string, fn func(*os.File) error, file *os.File, expectedErr error) fileLockTestCase {
	return fileLockTestCase{
		fn:          fn,
		name:        name,
		file:        file,
		expectedErr: expectedErr,
	}
}

func fileLockTestCases() []fileLockTestCase {
	// Create a temporary file for valid file tests
	tempFile, err := os.CreateTemp("", "fssyscall_test_*.tmp")
	if err != nil {
		panic("failed to create temp file for tests: " + err.Error())
	}

	return []fileLockTestCase{
		// Nil file tests - should return core.ErrInvalid
		newFileLockTestCase("FLockEx with nil file", fssyscall.FLockEx, nil, core.ErrInvalid),
		newFileLockTestCase("FUnlockEx with nil file", fssyscall.FUnlockEx, nil, core.ErrInvalid),
		newFileLockTestCase("FTryLockEx with nil file", fssyscall.FTryLockEx, nil, core.ErrInvalid),

		// Valid file tests - behaviour depends on platform
		newFileLockTestCase("FLockEx with valid file", fssyscall.FLockEx, tempFile, nil),
		newFileLockTestCase("FUnlockEx with valid file", fssyscall.FUnlockEx, tempFile, nil),
		newFileLockTestCase("FTryLockEx with valid file", fssyscall.FTryLockEx, tempFile, nil),
	}
}

func TestFileLocking(t *testing.T) {
	testCases := fileLockTestCases()

	// Clean up temp file after tests
	defer func() {
		for _, tc := range testCases {
			if tc.file != nil {
				_ = tc.file.Close()
				_ = os.Remove(tc.file.Name())
			}
		}
	}()

	core.RunTestCases(t, testCases)
}

// Test lock/unlock sequence
func TestLockUnlockSequence(t *testing.T) {
	t.Run("lock then unlock", runTestLockUnlockSequence)
}

func runTestLockUnlockSequence(t *testing.T) {
	t.Helper()

	tempFile, err := os.CreateTemp("", "fssyscall_lock_test_*.tmp")
	core.AssertNoError(t, err, "create temp file")
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	// Lock the file
	err = fssyscall.FLockEx(tempFile)
	core.AssertNoError(t, err, "lock file")

	// Unlock the file
	err = fssyscall.FUnlockEx(tempFile)
	core.AssertNoError(t, err, "unlock file")
}

// Test try lock behaviour
func TestTryLockBehaviour(t *testing.T) {
	t.Run("try lock on unlocked file", runTestTryLockUnlocked)
}

func runTestTryLockUnlocked(t *testing.T) {
	t.Helper()

	tempFile, err := os.CreateTemp("", "fssyscall_trylock_test_*.tmp")
	core.AssertNoError(t, err, "create temp file")
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	// Try lock should succeed on unlocked file
	err = fssyscall.FTryLockEx(tempFile)
	core.AssertNoError(t, err, "try lock unlocked file")

	// Clean up by unlocking
	err = fssyscall.FUnlockEx(tempFile)
	core.AssertNoError(t, err, "unlock after try lock")
}

// Test error handling scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("nil file validation", runTestNilFileValidation)
}

func runTestNilFileValidation(t *testing.T) {
	t.Helper()

	// Test all functions with nil file
	err := fssyscall.FLockEx(nil)
	core.AssertError(t, err, "FLockEx with nil")
	core.AssertErrorIs(t, err, core.ErrInvalid, "FLockEx error type")

	err = fssyscall.FUnlockEx(nil)
	core.AssertError(t, err, "FUnlockEx with nil")
	core.AssertErrorIs(t, err, core.ErrInvalid, "FUnlockEx error type")

	err = fssyscall.FTryLockEx(nil)
	core.AssertError(t, err, "FTryLockEx with nil")
	core.AssertErrorIs(t, err, core.ErrInvalid, "FTryLockEx error type")
}
