package lru

import (
	"context"
	"crypto/tls"
	"io/fs"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

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

	cert, exp, ok := s.lru.Get(serverName)
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

	cert, exp, ok := s.lru.Get(suffix)
}

func (s *LRU) doGetUpstream(ctx context.Context, serverName string, chi *tls.ClientHelloInfo) (*tls.Certificate, error)

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
