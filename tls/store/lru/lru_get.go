package lru

import (
	"context"

	"darvaza.org/core"
	"darvaza.org/x/container/list"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

type doGetResult struct {
	Cert *tls.Certificate
	Err  error
}

func (r doGetResult) Export() (*tls.Certificate, error) {
	switch {
	case r.Cert != nil:
		return r.Cert, nil
	case r.Err != nil:
		return nil, r.Err
	default:
		return nil, core.NewPanicWrap(1, core.ErrUnreachable, "doGet returned uninitialized results")
	}
}

// GetCertificate retrieves a TLS certificate for the given [tls.ClientHelloInfo].
// It is thread-safe.
func (s *LRU) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	ctx, serverName, err := tls.SplitClientHelloInfo(chi)
	if err != nil {
		return nil, err
	} else if err := s.checkInit(); err != nil {
		return nil, err
	} else if err := ctx.Err(); err != nil {
		return nil, err
	}

	select {
	case result := <-s.doGetChan(ctx, serverName, chi):
		return result.Export()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Get retrieves a TLS certificate for the given server name.
// It is thread-safe.
func (s *LRU) Get(ctx context.Context, serverName string) (*tls.Certificate, error) {
	if err := s.checkInit(); err != nil {
		return nil, err
	} else if err := ctx.Err(); err != nil {
		return nil, err
	}

	select {
	case result := <-s.doGetChan(ctx, serverName, nil):
		return result.Export()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// doGetChan initiates an asynchronous retrieval of a TLS certificate for the given server name.
// It returns a receive-only channel that will contain the result of the certificate lookup.
// The method sanitizes the server name and handles potential errors during retrieval.
// If an error occurs during name sanitization or certificate retrieval, it will be sent through the channel.
// It is thread-safe.
func (s *LRU) doGetChan(ctx context.Context, serverName string, chi *tls.ClientHelloInfo) <-chan doGetResult {
	result := make(chan doGetResult, 1)

	serverName, err := sanitizeName(serverName)
	if err != nil {
		// invalid
		defer close(result)
		result <- doGetResult{Err: err}
		return result
	}

	go func() {
		defer close(result)

		defer func() {
			if err := core.AsRecovered(recover()); err != nil {
				// doGet panicked
				result <- doGetResult{Err: err}
			}
		}()

		cert, err := s.doGet(ctx, serverName, chi)
		result <- doGetResult{Cert: cert, Err: err}
	}()

	return result
}

// doGet retrieves a TLS certificate for a given server name using a thread-safe, cache-first approach.
// It first checks the LRU cache for an existing certificate, and if not found, initiates
// an asynchronous retrieval process. The method handles locking and potential context cancellation.
func (s *LRU) doGet(ctx context.Context, serverName string, chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// automatic conditional unlock
	locked := true
	s.mu.Lock()
	defer func() {
		if locked {
			s.mu.Unlock()
		}
	}()

	// check cache
	if cert, ok := s.lruGet(serverName); ok {
		// hit
		return cert, nil
	}

	// initiate asynchronous retrieval
	result := make(chan doGetResult, 1)
	defer close(result)

	if err := s.spawnSingleFlightGet(ctx, serverName, chi, result); err != nil {
		// failed to spawn goroutine
		return nil, err
	}

	// release lock and wait for result
	s.mu.Unlock()
	locked = false

	select {
	case res := <-result:
		return res.Export()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// this is not thread-safe.
func (s *LRU) lruGet(serverName string) (*tls.Certificate, bool) {
	// exact match
	cert, ok := s.lruGetFirstNameHit(s.names, serverName)
	if ok {
		return cert, true
	}

	// suffix match
	suffix, ok := x509utils.NameAsSuffix(serverName)
	if ok {
		return s.lruGetFirstNameHit(s.suffixes, suffix)
	}

	return nil, false
}

// lruGetFirstNameHit searches for a certificate in the given map using the provided name.
// It returns the first valid certificate found and marks it to postpone eviction.
// If no valid certificate is found, it returns nil and false.
// It is not thread-safe.
func (s *LRU) lruGetFirstNameHit(m map[string]*list.List[*lruEntry], name string) (*tls.Certificate, bool) {
	if l, ok := m[name]; ok {
		if e, ok := s.lruGetFirstEntry(l); ok {
			s.postponeEviction(e)
			return e.cert, true
		}
	}
	return nil, false
}

// lruGetFirstEntry finds the first valid entry in a list of lruEntries.
// It returns the first valid entry and true if found, or nil and false otherwise.
// Invalid entries are automatically evicted during the search.
// It is not thread-safe.
func (s *LRU) lruGetFirstEntry(l *list.List[*lruEntry]) (*lruEntry, bool) {
	var out *lruEntry

	if s == nil || l == nil {
		return nil, false
	}

	l.ForEach(func(e *lruEntry) bool {
		if e.Valid() {
			out = e
		} else {
			s.evict(e)
		}

		return out != nil // stop once set
	})

	return out, out != nil
}
