package sni

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/sync/workgroup"
)

var (
	_ net.Listener = (*Dispatcher)(nil)
)

// A Handler is a function that will take responsibility over a given
// connection. The provided Context is used to indicate when a shut down
// has been initiated
type Handler func(context.Context, net.Conn) error

// The Dispatcher screens TCP connections and uses SNI to decide if
// they should be handled by a dedicated system or passed to
// the tls.Listener using it via Accept().
//
//	dispatcher := &sni.Dispatcher{
//		GetHandler: func(chi *tls.ClientHelloInfo) sni.Handler {
//			if chi.ServerName == "example.com" {
//				return exampleHandler
//			}
//			return nil // fall through to Accept
//		},
//	}
//
//	go dispatcher.Serve(rawListener)
//	tlsListener := tls.NewListener(dispatcher, cfg)
type Dispatcher struct {
	ch chan net.Conn

	ln  net.Listener
	log slog.Logger

	// Logger to report errors
	Logger slog.Logger
	// Context to be used as parent of the internal Canceller
	Context context.Context

	// GetHandler tells the Dispatcher if the connection associated with
	// a given ClientHelloInfo should be passed to a dedicated Handler
	// instead of passing it to the outer tls.Listener
	GetHandler func(*tls.ClientHelloInfo) Handler

	// OnAccept is optionally used to configure the inbound net.Conn
	OnAccept func(net.Conn) (net.Conn, error)

	// OnError is consulted on every error the Dispatcher meets, from the
	// underlying listener's Accept and from connection handlers alike,
	// once it has been logged, and decides whether it shuts the
	// Dispatcher down. When unset, an Accept error terminates and a
	// handler error is absorbed. An Accept error it waives is retried
	// at a bounded pace. The underlying listener being closed always
	// ends Serve cleanly and is not offered to it.
	OnError func(err error) bool

	// wg owns the Dispatcher's lifecycle: its context is the one handlers
	// receive, its cancellation is the shutdown, and its cause is the
	// first fatal error.
	wg workgroup.Group

	// mu guards ln; once gates init.
	mu   sync.Mutex
	once sync.Once
}

// init realises the accept channel, the logger and the workgroup's
// context, and runs at most once, before any use of the workgroup, so
// Context is honoured whichever method is called first.
func (d *Dispatcher) init() {
	d.once.Do(func() {
		// Accept()
		d.ch = make(chan net.Conn)

		// Cancel()
		d.wg.Parent = d.Context
		_ = d.wg.Context()

		// Logger
		d.log = d.Logger
	})
}

// Serve starts processing the underlying net.Listener. It returns
// nil once the listener has been closed or the Dispatcher shut down,
// and a fatal error from the listener's Accept otherwise. A handler's
// fatal error cancels the Dispatcher but leaves Serve waiting on the
// listener, so it is reported by [Dispatcher.Err] rather than here.
func (d *Dispatcher) Serve(ln net.Listener) error {
	if ln == nil {
		return core.ErrInvalid
	}

	d.init()

	d.mu.Lock()
	if d.ln != nil {
		d.mu.Unlock()
		return core.ErrExists
	}
	d.ln = ln
	d.mu.Unlock()

	return d.run()
}

func (d *Dispatcher) run() error {
	defer d.Cancel()

	var delay time.Duration
	for {
		conn, err := d.ln.Accept()
		if conn != nil {
			delay = 0
			d.spawnHandler(conn)
			continue
		}

		if stop, err := d.acceptError(err); stop {
			return err
		}

		// OnError waived the error: pace the retry so a listener
		// failing persistently cannot spin the loop.
		delay = nextAcceptDelay(delay)
		if !d.pause(delay) {
			return nil
		}
	}
}

// acceptError decides whether an error from the listener's Accept ends
// Serve, and with what. A closed listener or a shut-down Dispatcher is
// a clean end; anything else is offered to catch.
func (d *Dispatcher) acceptError(err error) (bool, error) {
	if d.wg.IsCancelled() || errors.Is(err, net.ErrClosed) {
		// bye
		return true, nil
	}

	// oops, unless OnError says otherwise
	err = d.catch(nil, err)
	return err != nil, err
}

// nextAcceptDelay paces the retries after an accept error OnError
// waived, doubling from 5 ms up to a second the way
// nextAcceptDelay calculates the next accept retry delay, doubling the current delay and limiting it to 5 milliseconds through 1 second.
func nextAcceptDelay(delay time.Duration) time.Duration {
	const first, limit = 5 * time.Millisecond, time.Second
	return min(max(2*delay, first), limit)
}

// pause waits delay before the next Accept, and reports false when the
// shutdown arrives first.
func (d *Dispatcher) pause(delay time.Duration) bool {
	select {
	case <-time.After(delay):
		return true
	case <-d.wg.Cancelled():
		return false
	}
}

// spawnHandler enrols the connection's handler in the workgroup. The
// catch routes the outcome through catch, which decides whether the
// Dispatcher terminates. A connection accepted after shutdown has
// nobody to run it and is dropped.
func (d *Dispatcher) spawnHandler(conn net.Conn) {
	err := d.wg.GoCatch(
		func(ctx context.Context) error {
			return d.handle(ctx, conn)
		},
		func(_ context.Context, err error) error {
			return d.catch(conn.RemoteAddr(), err)
		})
	if err != nil {
		d.drop(conn)
	}
}

// drop closes a connection the shutdown caught before anyone could take
// it, and says so at debug level.
func (d *Dispatcher) drop(conn net.Conn) {
	_ = conn.Close()
	d.logDebug(conn.RemoteAddr(), "dropped")
}

// logDebug records msg for peer at debug level, when enabled.
func (d *Dispatcher) logDebug(peer net.Addr, msg string) {
	if l, ok := d.debug(peer); ok {
		l.Print(msg)
	}
}

// logError records err at error level, naming the accept when it has
// no peer.
func (d *Dispatcher) logError(peer net.Addr, err error) {
	l, ok := d.error(peer, err)
	switch {
	case !ok:
	case peer == nil:
		l.Printf("accept: %s", err)
	default:
		l.Print(err)
	}
}

func (d *Dispatcher) handle(ctx context.Context, conn net.Conn) error {
	if d.OnAccept != nil {
		conn2, err := d.OnAccept(conn)
		if err != nil {
			_ = conn.Close()
			return err
		}
		conn = conn2
	}

	if d.GetHandler == nil {
		// no need to get the ClientHelloInfo here
		if l, ok := d.debug(conn.RemoteAddr()); ok {
			l.Print("connected")
		}
		return d.defaultHandler(ctx, conn)
	}

	return d.handleCHI(ctx, conn)
}

func (d *Dispatcher) handleCHI(ctx context.Context, conn net.Conn) error {
	// Get ClientHelloInfo
	chi, conn2, err := PeekClientHelloInfo(ctx, conn)
	if err != nil {
		_ = conn.Close()
		return err
	}

	if l, ok := d.debug(conn.RemoteAddr()); ok {
		l.WithField("sni", chi.ServerName).
			Print("connected")
	}

	// Get alternative handler
	h := d.GetHandler(chi)
	if h == nil {
		h = d.defaultHandler
	}

	return h(ctx, conn2)
}

// defaultHandler hands the connection to Accept, or drops it when the
// Dispatcher shuts down before a consumer collects it.
func (d *Dispatcher) defaultHandler(ctx context.Context, conn net.Conn) error {
	select {
	case d.ch <- conn:
	case <-ctx.Done():
		d.drop(conn)
	}
	return nil
}

// catch logs an error and decides whether it terminates the
// Dispatcher. A nil peer marks an error from the underlying listener's
// Accept, terminating unless OnError objects; a handler's error is
// absorbed unless OnError asks otherwise. A handler the shutdown itself
// interrupted has not failed: its cancellation is recorded as a drop
// and not offered to OnError. The returned error is the one the
// Dispatcher terminates on, and nil when it carries on.
func (d *Dispatcher) catch(peer net.Addr, err error) error {
	switch {
	case err == nil:
		d.logDebug(peer, "done")
		return nil
	case d.wg.IsCancelled() && errors.Is(err, context.Canceled):
		d.logDebug(peer, "dropped")
		return nil
	}

	d.logError(peer, err)

	if !d.terminate(err, peer == nil) {
		return nil
	}

	d.wg.Cancel(err)
	return err
}

// terminate asks OnError whether err shuts the Dispatcher down,
// falling back to the class's default when the hook is unset.
func (d *Dispatcher) terminate(err error, byDefault bool) bool {
	if d.OnError != nil {
		return d.OnError(err)
	}
	return byDefault
}

// Shutdown initiates a shutdown and waits until the workers are done
// or the given context times out.
func (d *Dispatcher) Shutdown(ctx context.Context) error {
	d.Cancel()

	select {
	case <-d.wg.Done():
		return d.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Accept returns a connection that wasn't dispatched through
// the Handler provided by GetHandler. Once the Dispatcher has been
// shut down it fails with the fatal error that stopped it, or with
// [net.ErrClosed] for a clean stop.
func (d *Dispatcher) Accept() (net.Conn, error) {
	d.init()

	select {
	case conn := <-d.ch:
		return conn, nil
	case <-d.wg.Cancelled():
		return nil, core.CoalesceError(d.Err(), net.ErrClosed)
	}
}

// Addr returns the address the underlying listener is using
func (d *Dispatcher) Addr() net.Addr {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.ln != nil {
		return d.ln.Addr()
	}
	return nil
}

// Close initiates a shut down but also returns
// the first fatal error if there was one
func (d *Dispatcher) Close() error {
	d.Cancel()
	return d.Err()
}

// Err tells the first fatal error, and nil after a clean shut down.
func (d *Dispatcher) Err() error {
	d.init()

	return filterCancelled(d.wg.Err())
}

// Wait waits until all workers are done, and reports as [Dispatcher.Err].
func (d *Dispatcher) Wait() error {
	d.init()

	return filterCancelled(d.wg.Wait())
}

// filterCancelled drops the cause a user-initiated shutdown leaves on
// filterCancelled removes context cancellation errors and preserves all other errors.
func filterCancelled(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

// Cancel initiates a shut down. It will prevent
// new dispatches and cancel existing workers, but
// the responsibility of closing the listener is on
// the tls.Listener
func (d *Dispatcher) Cancel() {
	d.init()

	d.wg.Cancel(nil)
}

// Cancelled tells if the Dispatcher has been shut down
func (d *Dispatcher) Cancelled() bool {
	d.init()

	return d.wg.IsCancelled()
}

func (d *Dispatcher) debug(peer net.Addr) (slog.Logger, bool) {
	return d.loggerWithFields(slog.Debug, peer, nil)
}

func (d *Dispatcher) error(peer net.Addr, err error) (slog.Logger, bool) {
	return d.loggerWithFields(slog.Error, peer, err)
}

func (d *Dispatcher) loggerWithFields(level slog.LogLevel, peer net.Addr, err error) (slog.Logger, bool) {
	l := d.log
	if l == nil {
		return nil, false
	}

	l, ok := l.WithLevel(level).WithEnabled()
	if !ok {
		return nil, false
	}

	l = l.WithField("dispatcher", d.ln.Addr().String())
	if peer != nil {
		l = l.WithField("peer", peer.String())
	}

	if err != nil {
		l = l.WithField(slog.ErrorFieldName, err)
	}

	return l, true
}
