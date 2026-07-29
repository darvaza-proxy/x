// Package macos provides cgo-less access to the CoreFoundation and
// Security.framework APIs the darwin system certificate-pool loader needs,
// bound at run time through the ebitengine/purego FFI.
//
// The bindings are darwin-only; on every other platform this package is
// empty. Callers obtain a [Bindings] with [New] and drive the framework
// through it: entry points such as [Bindings.CopyTrustSettingsCertificates]
// hand back opaque handles ([Array], [Certificate], [Policy], [Trust]) whose
// own methods wrap the remaining calls. The package imposes no policy and
// keeps no state beyond the resolved symbols; the caller owns the
// trust-evaluation flow.
package macos

// cspell:words ebitengine purego
