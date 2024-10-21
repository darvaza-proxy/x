package lru

import (
	"context"
	"crypto/tls"
	"io/fs"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// GetCertificate implements the tls.GetCertificate callback for retrieving a certificate
// based on the ClientHelloInfo. It extracts the server name, initializes the LRU store,
// and attempts to retrieve a matching certificate through the doGet method.
func (s *LRU) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	serverName, err := getName(chi)
	if err == nil {
		err = s.init()
	}
	if err != nil {
		return nil, err
	}

	return s.doGet(chi.Context(), serverName, chi)
}

// Get retrieves a TLS certificate for the given server name from the LRU cache.
// It initializes the cache, validates the server name, and attempts to find a matching certificate.
// Returns the certificate if found, or an error if initialization fails, the context is canceled,
// or the server name is invalid.
func (s *LRU) Get(ctx context.Context, serverName string) (*tls.Certificate, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	name, ok := x509utils.SanitizeName(serverName)
	if !ok {
		return nil, core.QuietWrap(core.ErrInvalid, "invalid name: %q", serverName)
	}

	chi, _ := ClientHelloInfo(ctx)
	return s.doGet(ctx, name, chi)
}

func (s *LRU) doGet(ctx context.Context, serverName string, chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := s.doGetByName(ctx, serverName)
	switch {
	case cert != nil:
		return cert, nil
	case err != nil && err != ErrNotFound:
		return nil, err
	}

	cert, err = s.doGetBySuffix(ctx, serverName)
	switch {
	case cert != nil:
		return cert, nil
	case err != nil && err != fs.ErrNotExist && err != fs.ErrInvalid:
		return nil, err
	}

	return s.doGetUpstream(ctx, serverName, chi)
}

func (s *LRU) doGetByName(ctx context.Context, serverName string) (*tls.Certificate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exp, ok := s.unsafeGet(serverName)
	if !ok || time.Now().After(exp) {
		return nil, fs.ErrNotExist
	}

	return cert, nil
}

func (s *LRU) doGetBySuffix(ctx context.Context, serverName string) (*tls.Certificate, error) {
	suffix, ok := x509utils.NameAsSuffix(serverName)
	if !ok {
		return nil, fs.ErrInvalid
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cert, exp, ok := s.unsafeGet(suffix)
	if !ok || time.Now().After(exp) {
		return nil, fs.ErrNotExist
	}
	return cert, nil
}

func (*LRU) unsafeGet(string) (*tls.Certificate, time.Time, bool) {
	// 	core.Panic(core.ErrTODO)
	return nil, time.Time{}, false
}

func (*LRU) doGetUpstream(context.Context, string, *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return nil, core.ErrTODO
}

func getName(chi *tls.ClientHelloInfo) (string, error) {
	if chi == nil {
		return "", core.Wrap(core.ErrInvalid, "chi not provided")
	}

	name := chi.ServerName
	if name == "" {
		name = chi.Conn.LocalAddr().String()
	}

	s, ok := x509utils.SanitizeName(name)
	if !ok {
		return "", core.Wrapf(core.ErrInvalid, "invalid serverName: %q", name)
	}
	return s, nil
}
