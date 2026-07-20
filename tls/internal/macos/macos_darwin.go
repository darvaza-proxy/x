//go:build darwin

package macos

// cspell:words purego ebitengine CoreFoundation dlopen Dlsym OSStatus RTLD fptr

import (
	"fmt"
	"unsafe"

	"darvaza.org/core"

	"github.com/ebitengine/purego"
)

// System frameworks, loaded by full path so dlopen resolves them without a
// search-path dance. They are resident for the life of the process, so the
// handles are never freed.
const (
	frameworkCoreFoundation = "/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation"
	frameworkSecurity       = "/System/Library/Frameworks/Security.framework/Security"

	// errSecSuccess is the OSStatus returned on success.
	errSecSuccess = 0
	// errSecNoTrustSettings is the OSStatus SecTrustSettingsCopyCertificates
	// returns for a domain that holds no trust settings at all. It is a clean
	// "nothing here", not a failure.
	errSecNoTrustSettings = -25263
)

// Domain selects one of the three trust-settings domains Security.framework
// keeps, matching the SecTrustSettingsDomain enumeration.
type Domain uint32

// The trust-settings domains, in the order the system resolves them: a
// setting in a more specific domain overrides a less specific one.
const (
	DomainUser   Domain = 0
	DomainAdmin  Domain = 1
	DomainSystem Domain = 2
)

// OSStatus is a non-success result code returned by a Security.framework
// call, tagged with the name of the call that produced it.
type OSStatus struct {
	call   string
	status int32
}

// Error implements the error interface.
func (e OSStatus) Error() string {
	return fmt.Sprintf("%s: OSStatus %d", e.call, e.status)
}

// Bindings holds the CoreFoundation and Security.framework entry points
// resolved through purego. Construct it with [New]. It owns no policy and
// no algorithm: it only bridges the framework, handing back opaque handles
// ([Array], [Certificate], [Policy], [Trust]) and, at the edges, plain Go
// values. The caller drives the trust-evaluation flow.
type Bindings struct {
	// Security.framework
	secTrustSettingsCopyCertificates func(domain uint32, out *uintptr) int32
	secPolicyCreateBasicX509         func() uintptr
	secTrustCreateWithCertificates   func(certs, policies uintptr, trust *uintptr) int32
	secTrustEvaluateWithError        func(trust uintptr, cfErr *uintptr) bool
	secCertificateCopyData           func(cert uintptr) uintptr

	// CoreFoundation
	cfArrayGetCount        func(arr uintptr) int
	cfArrayGetValueAtIndex func(arr uintptr, i int) uintptr
	cfDataGetLength        func(data uintptr) int
	cfDataGetBytePtr       func(data uintptr) unsafe.Pointer
	cfRelease              func(ref uintptr)
}

// New loads CoreFoundation and Security.framework and resolves every symbol
// up front, so a missing entry point is reported once rather than surfacing
// mid-enumeration.
func New() (*Bindings, error) {
	cf, err := purego.Dlopen(frameworkCoreFoundation, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, core.Wrap(err, "CoreFoundation")
	}
	sec, err := purego.Dlopen(frameworkSecurity, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, core.Wrap(err, "Security")
	}

	var b Bindings
	var errs core.CompoundError
	bind := func(fptr any, h uintptr, name string) {
		sym, e := purego.Dlsym(h, name)
		if e != nil {
			_ = errs.AppendError(core.Wrap(e, name))
			return
		}
		purego.RegisterFunc(fptr, sym)
	}

	bind(&b.secTrustSettingsCopyCertificates, sec, "SecTrustSettingsCopyCertificates")
	bind(&b.secPolicyCreateBasicX509, sec, "SecPolicyCreateBasicX509")
	bind(&b.secTrustCreateWithCertificates, sec, "SecTrustCreateWithCertificates")
	bind(&b.secTrustEvaluateWithError, sec, "SecTrustEvaluateWithError")
	bind(&b.secCertificateCopyData, sec, "SecCertificateCopyData")
	bind(&b.cfArrayGetCount, cf, "CFArrayGetCount")
	bind(&b.cfArrayGetValueAtIndex, cf, "CFArrayGetValueAtIndex")
	bind(&b.cfDataGetLength, cf, "CFDataGetLength")
	bind(&b.cfDataGetBytePtr, cf, "CFDataGetBytePtr")
	bind(&b.cfRelease, cf, "CFRelease")

	if err := errs.AsError(); err != nil {
		return nil, err
	}
	return &b, nil
}

// Array is a CoreFoundation array of certificates returned by the framework.
// Its elements are borrowed and must not be released individually; releasing
// the array frees them.
type Array struct {
	b   *Bindings
	ref uintptr
}

// Certificate is a SecCertificateRef. When obtained from [Array.At] it is
// borrowed from that array and must not be released on its own.
type Certificate struct {
	b   *Bindings
	ref uintptr
}

// Policy is an owned SecPolicyRef. Release it with [Policy.Release].
type Policy struct {
	b   *Bindings
	ref uintptr
}

// Trust is an owned SecTrustRef. Release it with [Trust.Release].
type Trust struct {
	b   *Bindings
	ref uintptr
}

// CopyTrustSettingsCertificates returns the certificates that carry an
// explicit trust setting in domain d, via SecTrustSettingsCopyCertificates.
// A domain with no trust settings yields a zero [Array] and a nil error, so
// the caller can walk every domain uniformly.
func (b *Bindings) CopyTrustSettingsCertificates(d Domain) (Array, error) {
	var out uintptr
	switch status := b.secTrustSettingsCopyCertificates(uint32(d), &out); status {
	case errSecSuccess:
		return Array{b: b, ref: out}, nil
	case errSecNoTrustSettings:
		return Array{b: b}, nil
	default:
		return Array{}, OSStatus{call: "SecTrustSettingsCopyCertificates", status: status}
	}
}

// Count returns the number of certificates in the array. A zero Array — an
// empty or absent domain — reports 0.
func (a Array) Count() int {
	if a.ref == 0 {
		return 0
	}
	return a.b.cfArrayGetCount(a.ref)
}

// At returns the certificate at index i. The reference is borrowed from the
// array and stays valid until the array is released; do not release it.
func (a Array) At(i int) Certificate {
	return Certificate{b: a.b, ref: a.b.cfArrayGetValueAtIndex(a.ref, i)}
}

// Release frees the array. A zero Array is a no-op.
func (a Array) Release() {
	if a.ref != 0 {
		a.b.cfRelease(a.ref)
	}
}

// CopyData copies the DER encoding out of the certificate, via
// SecCertificateCopyData. The copy is independent of the framework-owned
// storage it is read from.
func (c Certificate) CopyData() ([]byte, error) {
	if c.ref == 0 {
		return nil, core.Wrap(core.ErrInvalid, "certificate: null")
	}
	dataRef := c.b.secCertificateCopyData(c.ref)
	if dataRef == 0 {
		return nil, core.Wrap(core.ErrInvalid, "SecCertificateCopyData: null")
	}
	defer c.b.cfRelease(dataRef)

	length := c.b.cfDataGetLength(dataRef)
	ptr := c.b.cfDataGetBytePtr(dataRef)
	if ptr == nil || length <= 0 {
		return nil, core.Wrap(core.ErrInvalid, "CFData: empty")
	}

	// Copy before the owning CFData is released.
	der := make([]byte, length)
	copy(der, unsafe.Slice((*byte)(ptr), length))
	return der, nil
}

// NewBasicX509Policy creates a basic X.509 policy via
// SecPolicyCreateBasicX509. It checks certificate validity — dates,
// signatures, chaining — without any protocol's extended-key-usage
// requirement, so it asks the neutral "is this a trusted anchor" question a
// bare CA root can answer.
func (b *Bindings) NewBasicX509Policy() Policy {
	return Policy{b: b, ref: b.secPolicyCreateBasicX509()}
}

// Release frees the policy. A zero Policy is a no-op.
func (p Policy) Release() {
	if p.ref != 0 {
		p.b.cfRelease(p.ref)
	}
}

// NewTrust builds a trust object that evaluates certificate c against policy
// p, via SecTrustCreateWithCertificates. Both references are borrowed for the
// call; their owners may release them once the returned [Trust] is released.
func (b *Bindings) NewTrust(c Certificate, p Policy) (Trust, error) {
	var trust uintptr
	if status := b.secTrustCreateWithCertificates(c.ref, p.ref, &trust); status != errSecSuccess {
		return Trust{}, OSStatus{call: "SecTrustCreateWithCertificates", status: status}
	}
	return Trust{b: b, ref: trust}, nil
}

// EvaluateWithError reports whether the certificate is trusted under the
// policy, via SecTrustEvaluateWithError. The system applies its own trust
// settings and dynamic distrust internally, so a true result means "trusted
// as configured right now". The CFError explaining a refusal is released
// unread; the caller acts on the verdict alone.
func (t Trust) EvaluateWithError() bool {
	var cfErr uintptr
	if t.b.secTrustEvaluateWithError(t.ref, &cfErr) {
		return true
	}
	if cfErr != 0 {
		t.b.cfRelease(cfErr)
	}
	return false
}

// Release frees the trust object. A zero Trust is a no-op.
func (t Trust) Release() {
	if t.ref != 0 {
		t.b.cfRelease(t.ref)
	}
}
