package workgroup

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"darvaza.org/x/sync/errors"
)

// TestRunner runs all tests for Runner
func TestRunner(t *testing.T) {
	t.Run("Creation", TestRunnerCreation)
	t.Run("Nil receiver", TestRunnerNilReceiver)

	// Run common WaitGroup tests
	t.Run("Common WaitGroup behaviour", func(t *testing.T) {
		testWaitGroupBehaviours(t, func() WaitGroup {
			return NewRunner()
		})
	})

	// Runner-specific tests
	t.Run("Wait with concurrent additions", runTestWaitConcurrentAdditions)
}

// TestRunnerCreation tests the creation and initialisation of a Runner
func TestRunnerCreation(t *testing.T) {
	t.Run("NewRunner", runTestNewRunner)
	t.Run("Manual initialisation", runTestManualInit)
}

func runTestNewRunner(t *testing.T) {
	runner := NewRunner()
	assert.NotNil(t, runner)
	assert.False(t, runner.IsNil())
	assert.False(t, runner.IsClosed())
	assert.Equal(t, 0, runner.Count())
}

func runTestManualInit(t *testing.T) {
	runner := new(Runner)
	assert.True(t, runner.IsNil())

	err := runner.Init()
	assert.NoError(t, err)
	assert.False(t, runner.IsNil())
	assert.False(t, runner.IsClosed())
	assert.Equal(t, 0, runner.Count())

	// Init should fail on already initialised runner
	err = runner.Init()
	assert.Equal(t, errors.ErrAlreadyInitialised, err)
}

// TestRunnerNilReceiver tests operations on nil Runner
func TestRunnerNilReceiver(t *testing.T) {
	var runner *Runner

	t.Run("IsNil", func(t *testing.T) {
		assert.True(t, runner.IsNil())
	})

	t.Run("IsClosed", func(t *testing.T) {
		assert.True(t, runner.IsClosed())
	})

	t.Run("Count", func(t *testing.T) {
		assert.Equal(t, 0, runner.Count())
	})

	t.Run("Init", func(t *testing.T) {
		err := runner.Init()
		assert.Equal(t, errors.ErrNilReceiver, err)
	})

	t.Run("Go", func(t *testing.T) {
		err := runner.Go(func() {})
		assert.Equal(t, errors.ErrNilReceiver, err)
	})

	t.Run("Wait", func(t *testing.T) {
		err := runner.Wait()
		assert.Equal(t, errors.ErrNilReceiver, err)
	})

	t.Run("Close", func(t *testing.T) {
		err := runner.Close()
		assert.Equal(t, errors.ErrNilReceiver, err)
	})
}

func runTestWaitConcurrentAdditions(t *testing.T) {
	runner := NewRunner()
	defer runner.Close()

	var counter int32
	numGoroutines := 5

	// Start a goroutine that will add more goroutines
	err := runner.Go(func() {
		time.Sleep(50 * time.Millisecond)
		for i := 0; i < numGoroutines; i++ {
			err := runner.Go(func() {
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt32(&counter, 1)
			})
			if err != nil {
				return
			}
		}
	})
	require.NoError(t, err)

	// Wait should block until all goroutines complete
	err = runner.Wait()
	assert.NoError(t, err)
	assert.Equal(t, int32(numGoroutines), counter)
}
