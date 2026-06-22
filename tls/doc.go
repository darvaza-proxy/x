// Package tls provides utilities for working with TLS certificates: aliases for
// the standard library's certificate types, certificate verification ([Verify]),
// chain assembly ([Bundler], [Bundle]), configuration helpers ([WithStore],
// [NewConfig]) and the [Store] abstraction.
//
// Store is the central facility. Unlike the standard library's static
// configuration, a Store is a living entity: certificates are added, minted on
// demand, renewed, replicated and removed throughout its lifetime. The Store
// interfaces are built to be layered, so higher-level systems — distributed
// certificate serving, ACME issuance, auto-renewal, key replication — can
// compose behaviour over a simple local store while keeping it both flexible and
// consistent:
//
//   - Flexibility: the error sentinels and the callback hooks on
//     implementations are seams a decorator store builds on — a miss becomes
//     on-demand minting, a write becomes replication, a renewal becomes a swap
//     — rather than mere status reporting.
//   - Consistency: deduplicating implementations (see the certpool subpackage)
//     collapse identical certificates to a single reference and report a re-add
//     with an "already exists" sentinel, so layers compare by identity and stay
//     idempotent and convergent across nodes.
//
// These facilities build on the x509utils subpackage, which supplies the
// underlying x509 certificate and key primitives. Companion subpackages cover
// neighbouring concerns: sni screens incoming connections by server name, and
// store loads certificates into a Store from configuration and PEM sources.
package tls
