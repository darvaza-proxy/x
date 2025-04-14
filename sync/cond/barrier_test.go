package cond

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBarrierIsNil(t *testing.T) {
	t.Run("nil barrier", func(t *testing.T) {
		var b *Barrier
		assert.True(t, b.IsNil(), "nil barrier should return true for IsNil()")
	})

	t.Run("uninitialized barrier", func(t *testing.T) {
		b := &Barrier{}
		assert.True(t, b.IsNil(), "uninitialized barrier should return true for IsNil()")
	})

	t.Run("initialized barrier", func(t *testing.T) {
		b := &Barrier{}
		err := b.Init()
		assert.NoError(t, err, "Init() should not return an error")
		assert.False(t, b.IsNil(), "initialized barrier should return false for IsNil()")
	})
}

func TestBarrierInit(t *testing.T) {
	b := &Barrier{}
	err := b.Init()
	assert.NoError(t, err, "Init() should not return an error")

	// After initialization, we should be able to acquire a token
	select {
	case token := <-b.Acquire():
		b.Release(token)
		assert.True(t, true, "token was acquired successfully")
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "failed to acquire token after initialization")
	}
}

func TestBarrierAcquireRelease(t *testing.T) {
	b := &Barrier{}
	defer b.Close()

	err := b.Init()
	assert.NoError(t, err, "Init() should not return an error")

	// First acquisition should succeed immediately
	var token Token
	select {
	case token = <-b.Acquire():
		assert.NotNil(t, token, "acquired token should not be nil")
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "failed to acquire token")
		return
	}

	// Second acquisition should block (we'll test with a goroutine)
	var wg sync.WaitGroup
	wg.Add(1)

	secondAcquired := false
	go func() {
		defer wg.Done()
		select {
		case <-b.Acquire():
			secondAcquired = true
		case <-time.After(50 * time.Millisecond):
			// This should timeout as we haven't released yet
		}
	}()

	// Wait for the goroutine to finish its attempt
	wg.Wait()
	assert.False(t, secondAcquired, "second acquire should not succeed while token is held")

	// Now release the token
	b.Release(token)

	// Now we should be able to acquire again
	select {
	case token = <-b.Acquire():
		b.Release(token)
		assert.NotNil(t, token, "token should be acquirable after release")
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "failed to acquire token after release")
	}
}

func TestBarrierToken(t *testing.T) {
	b := &Barrier{}
	err := b.Init()
	assert.NoError(t, err, "Init() should not return an error")

	// Get should not remove the token from the barrier
	token := b.Token()
	require.NotNil(t, token, "token should not be nil")

	// We should still be able to acquire after Get
	select {
	case acquired := <-b.Acquire():
		b.Release(acquired)
		assert.NotNil(t, acquired, "token should be acquirable after Get")
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "failed to acquire token after Get")
	}
}

func TestBarrierReset(t *testing.T) {
	b := &Barrier{}
	_ = b.Init()

	// Get the initial token
	initialToken := b.Token()
	require.NotNil(t, initialToken, "initial token should not be nil")

	// Set up a goroutine that waits on the token
	waitComplete := false
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		initialToken.Wait() // Should block until Reset closes the token
		waitComplete = true
	}()

	// Broadcast the barrier to release the goroutine
	b.Broadcast()

	// Wait for the goroutine to complete
	wg.Wait()
	assert.True(t, waitComplete, "Wait() should have completed after Reset()")

	// Get the new token and verify it's different
	newToken := b.Token()
	require.NotNil(t, newToken, "new token should not be nil")
	assert.NotEqual(t, initialToken, newToken, "new token should be different from initial token")

	// Verify the behaviour of the new token
	var newWg sync.WaitGroup
	newWaitComplete := false
	newWg.Add(1)

	go func() {
		defer newWg.Done()
		select {
		case <-newToken:
			newWaitComplete = true
		case <-time.After(50 * time.Millisecond):
			// Should timeout as we haven't reset
		}
	}()

	newWg.Wait()
	assert.False(t, newWaitComplete, "Wait on new token should not complete without Reset")
}

func TestTokenWait(t *testing.T) {
	// Create a token
	token := make(Token)

	// Test waiting with timeout using a channel
	waitComplete := false
	timeoutCh := make(chan bool, 1)

	go func() {
		token.Wait()
		waitComplete = true
		timeoutCh <- true
	}()

	// The wait should block until we close the token
	select {
	case <-timeoutCh:
		assert.Fail(t, "Wait() should block until token is closed")
	case <-time.After(50 * time.Millisecond):
		// Expected timeout
	}

	// Close the token and check that Wait() completes
	close(token)

	select {
	case <-timeoutCh:
		assert.True(t, waitComplete, "Wait() should have completed after token was closed")
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Wait() did not complete after token was closed")
	}
}

func TestConcurrentBarrierOperations(t *testing.T) {
	b := &Barrier{}
	_ = b.Init()

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()

			// Each goroutine will acquire and release the token
			token := <-b.Acquire()

			// Hold the token briefly
			time.Sleep(10 * time.Millisecond)

			// Release it
			b.Release(token)
		}()
	}

	wg.Wait()

	// After all goroutines are done, we should still be able to acquire
	select {
	case token := <-b.Acquire():
		assert.NotNil(t, token, "token should be acquirable after concurrent operations")
		b.Release(token)
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "failed to acquire token after concurrent operations")
	}
}
