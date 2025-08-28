package fssyscall

import (
	"io/fs"
	"os"
	"syscall"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = handleTestCase{}
var _ core.TestCase = openTestCase{}

type handleTestCase struct {
	name    string
	handle  Handle
	wantErr bool
}

func (tc handleTestCase) Name() string {
	return tc.name
}

func (tc handleTestCase) Test(t *testing.T) {
	t.Helper()

	err := tc.handle.Close()

	if tc.wantErr {
		core.AssertError(t, err, "close error")
	} else {
		core.AssertNoError(t, err, "close")
	}
}

func newHandleTestCase(name string, handle Handle, wantErr bool) handleTestCase {
	return handleTestCase{
		name:    name,
		handle:  handle,
		wantErr: wantErr,
	}
}

type openTestCase struct {
	filename string
	mode     int
	perm     fs.FileMode
	name     string
	wantErr  bool
}

func (tc openTestCase) Name() string {
	return tc.name
}

func (tc openTestCase) Test(t *testing.T) {
	t.Helper()

	handle, err := Open(tc.filename, tc.mode, tc.perm)

	if tc.wantErr {
		core.AssertError(t, err, "open error")
		core.AssertEqual(t, ZeroHandle, handle, "zero handle on error")
	} else {
		core.AssertNoError(t, err, "open")
		core.AssertNotEqual(t, ZeroHandle, handle, "valid handle")

		// Clean up by closing the handle
		_ = handle.Close()
	}
}

func newOpenTestCase(name, filename string, mode int, perm fs.FileMode, wantErr bool) openTestCase {
	return openTestCase{
		filename: filename,
		mode:     mode,
		perm:     perm,
		name:     name,
		wantErr:  wantErr,
	}
}

func TestHandleClose(t *testing.T) {
	// Create a temporary file to get a valid handle
	tempFile, err := os.CreateTemp("", "syscall_test_*.tmp")
	core.AssertNoError(t, err, "create temp file")
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	validHandle := Handle(tempFile.Fd())

	testCases := []handleTestCase{
		newHandleTestCase("close valid handle", validHandle, false),
		newHandleTestCase("close zero handle", ZeroHandle, true),
	}

	core.RunTestCases(t, testCases)
}

func TestOpen(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "syscall_open_test_*.tmp")
	core.AssertNoError(t, err, "create temp file")
	tempFileName := tempFile.Name()
	_ = tempFile.Close()
	defer func() {
		_ = os.Remove(tempFileName)
	}()

	testCases := []openTestCase{
		newOpenTestCase("open existing file", tempFileName, os.O_RDONLY, 0644, false),
		newOpenTestCase("open non-existent file", "/non-existent/file", os.O_RDONLY, 0644, true),
	}

	core.RunTestCases(t, testCases)
}

func TestIsZero(t *testing.T) {
	t.Run("zero handle", runTestIsZeroTrue)
	t.Run("non-zero handle", runTestIsZeroFalse)
}

func runTestIsZeroTrue(t *testing.T) {
	t.Helper()

	result := ZeroHandle.IsZero()
	core.AssertTrue(t, result, "zero handle is zero")
}

func runTestIsZeroFalse(t *testing.T) {
	t.Helper()

	// Create a temporary file to get a valid handle
	tempFile, err := os.CreateTemp("", "syscall_is_zero_test_*.tmp")
	core.AssertNoError(t, err, "create temp file")
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	handle := Handle(tempFile.Fd())
	result := handle.IsZero()
	core.AssertFalse(t, result, "valid handle is not zero")
}

func TestHandleLocking(t *testing.T) {
	t.Run("lock unlock sequence", runTestHandleLockSequence)
	t.Run("try lock on unlocked handle", runTestHandleTryLock)
}

func runTestHandleLockSequence(t *testing.T) {
	t.Helper()

	// Create a temporary file to get a valid handle
	tempFile, err := os.CreateTemp("", "syscall_lock_test_*.tmp")
	core.AssertNoError(t, err, "create temp file")
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	handle := Handle(tempFile.Fd())

	// Lock the handle
	err = LockEx(handle)
	core.AssertNoError(t, err, "lock handle")

	// Unlock the handle
	err = UnlockEx(handle)
	core.AssertNoError(t, err, "unlock handle")
}

func runTestHandleTryLock(t *testing.T) {
	t.Helper()

	// Create a temporary file to get a valid handle
	tempFile, err := os.CreateTemp("", "syscall_trylock_test_*.tmp")
	core.AssertNoError(t, err, "create temp file")
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	handle := Handle(tempFile.Fd())

	// Try lock should succeed on unlocked handle
	err = TryLockEx(handle)
	core.AssertNoError(t, err, "try lock unlocked handle")

	// Clean up by unlocking
	err = UnlockEx(handle)
	core.AssertNoError(t, err, "unlock after try lock")
}

func TestConcurrentLocking(t *testing.T) {
	t.Run("try lock on already locked file", runTestConcurrentTryLock)
}

func runTestConcurrentTryLock(t *testing.T) {
	t.Helper()

	// Create a temporary file for shared access
	tempFile, err := os.CreateTemp("", "syscall_concurrent_test_*.tmp")
	core.AssertNoError(t, err, "create temp file")
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	tryLockErr := runConcurrentLockTest(tempFile.Name())

	core.AssertError(t, tryLockErr, "try lock on locked file")
	core.AssertErrorIs(t, tryLockErr, syscall.EBUSY, "error type")
}

func runConcurrentLockTest(filename string) error {
	lockReadyErr := make(chan error, 1)
	tryLockDone := make(chan error, 1)
	unlockDone := make(chan struct{})

	go holdLockAndWait(filename, lockReadyErr, unlockDone)
	go attemptTryLock(filename, lockReadyErr, tryLockDone)

	tryLockErr := <-tryLockDone
	close(unlockDone)

	return tryLockErr
}

func holdLockAndWait(filename string, lockReadyErr chan error, unlockDone chan struct{}) {
	file, err := os.Open(filename)
	if err != nil {
		lockReadyErr <- err
		return
	}
	defer func() {
		_ = file.Close()
	}()

	handle := Handle(file.Fd())
	err = LockEx(handle)
	if err != nil {
		lockReadyErr <- err
		return
	}

	// Signal successful lock acquisition
	lockReadyErr <- nil
	<-unlockDone
	_ = UnlockEx(handle)
}

func attemptTryLock(filename string, lockReadyErr chan error, tryLockDone chan error) {
	// Wait for lock holder to signal success or failure
	lockErr := <-lockReadyErr
	if lockErr != nil {
		// Lock holder failed, propagate the error
		tryLockDone <- lockErr
		return
	}

	// Lock was successfully acquired by holder, now attempt TryLock
	file, err := os.Open(filename)
	if err != nil {
		tryLockDone <- err
		return
	}
	defer func() {
		_ = file.Close()
	}()

	handle := Handle(file.Fd())
	err = TryLockEx(handle)
	tryLockDone <- err
}
