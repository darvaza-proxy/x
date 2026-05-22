// Package synctesting provides timing-aware assertion helpers for tests of
// channel- and synchronisation-primitive-driven code in darvaza.org/x/sync.
//
// All Assert* helpers accept core.T, so consumers pass *testing.T directly
// and this package's own self-tests drive failure paths through core.MockT.
//
// The Assert*/AssertMust* pairs follow the darvaza.org/core convention:
// the bare form returns a bool the caller can branch on; the Must form
// terminates the test on failure via t.FailNow.
//
// # Helpers
//
//   - WaitForCond: poll a predicate until true or timeout (no test
//     reporting; pure primitive).
//   - AssertEventually / AssertMustEventually: wait up to timeout for a
//     predicate to hold, polling at PollStep cadence.
//   - AssertClosed / AssertMustClosed: assert a channel becomes readable
//     (closed or sent to) within timeout.
//   - AssertOpen / AssertMustOpen: assert a channel stays open for a
//     timeout window.
//   - AssertReadersReady / AssertMustReadersReady: assert n values arrive
//     on a channel within a shared timeout; close-before-n counts as
//     failure.
//
// PollStep is the default polling cadence (1 ms) used by the Eventually
// helpers. Callers needing a different cadence call WaitForCond directly
// with an explicit step.
package synctesting
