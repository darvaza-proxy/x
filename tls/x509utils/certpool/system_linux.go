//go:build linux

package certpool

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// NewSystemCertPool returns a [CertPool] populated
// with all system valid certificates and an aggregation of
// errors.
func NewSystemCertPool() (*CertPool, error) {
	var errs core.CompoundError
	var pool CertPool

	// Possible certificate files; stop after finding one.
	loadCertsFirstFile(&pool, &errs, certFiles...)

	// Possible directories with certificate files; all will be read.
	for _, dir := range certDirectories {
		if err := loadCertsDir(&pool, dir); err != nil {
			errs.AppendError(err)
		}
	}

	switch {
	case pool.Count() > 0:
		// success
		return &pool, errs.AsError()
	case errs.Ok():
		// no certs and no errors... don't bother again.
		return nil, ErrNoCertificatesFound
	default:
		// no cert, but we got errors to report.
		return nil, &errs
	}
}

func loadCertsFirstFile(pool *CertPool, errs *core.CompoundError, files ...string) {
	initial := pool.Count()
	addFn := newCertAdder(pool, SystemCAOnly, errs)
	for _, fileName := range files {
		err := loadCertsFile(addFn, fileName)
		switch {
		case pool.Count() > initial:
			return
		case err != nil:
			errs.AppendError(err)
		}
	}
}

func loadCertsFile(addFn x509utils.DecodePEMBlockFunc, fileName string) error {
	b, err := os.ReadFile(fileName)
	switch {
	case err != nil:
		return err
	case len(b) == 0:
		// empty file:
		return &fs.PathError{
			Op:   "read",
			Path: fileName,
			Err:  errors.New("empty certificates file"),
		}
	default:
		if err := x509utils.ReadPEM(b, addFn); err != nil {
			// bad content
			return &fs.PathError{
				Op:   "pem.Decode",
				Path: fileName,
				Err:  err,
			}
		}
		return nil
	}
}

func loadCertsDir(pool *CertPool, dir string) error {
	var errs core.CompoundError

	initial := pool.Count()
	addFn := newCertAdder(pool, SystemCAOnly, &errs)
	err := x509utils.ReadDirPEM(os.DirFS(dir), ".", addFn)
	switch {
	case pool.Count() > initial:
		// certs loaded. ignore errors
		return nil
	case err != nil:
		errs.AppendError(err)
	}

	// amend Path
	for i, err := range errs.Errs {
		if e, ok := err.(*fs.PathError); ok {
			errs.Errs[i] = amendPathError(e, dir)
		}
	}

	return errs.AsError()
}

func amendPathError(err *fs.PathError, baseDir string) *fs.PathError {
	path := err.Path

	// join
	switch {
	case baseDir == "":
		//
	case path == "", path == ".":
		path = baseDir
	default:
		path = filepath.Join(baseDir, path)
	}

	if path != err.Path {
		// Copy-On-Write
		err = &fs.PathError{
			Op:   err.Op,
			Path: path,
			Err:  err.Err,
		}
	}
	return err
}

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Possible certificate files; stop after finding one.
var certFiles = []string{
	"/etc/ssl/certs/ca-certificates.crt",                // Debian/Ubuntu/Gentoo etc.
	"/etc/pki/tls/certs/ca-bundle.crt",                  // Fedora/RHEL 6
	"/etc/ssl/ca-bundle.pem",                            // OpenSUSE
	"/etc/pki/tls/cacert.pem",                           // OpenELEC
	"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem", // CentOS/RHEL 7
	"/etc/ssl/cert.pem",                                 // Alpine Linux
	"/etc/certs/ca-certificates.crt",                    // used by some embedded systems[citation-needed]
}

// Possible directories with certificate files; all will be read.
var certDirectories = []string{
	"/etc/ssl/certs",                   // SLES10/SLES11, https://golang.org/issue/12139
	"/etc/pki/tls/certs",               // Fedora/RHEL
	"/system/etc/security/cacerts",     // Android
	"/usr/local/share/ca-certificates", // locally added certificates
}
