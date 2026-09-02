package sni_test

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"

	"darvaza.org/x/sync/errors"
	"darvaza.org/x/tls/sni"
)

// dispatcherTimeout bounds every wait in these tests, so a Dispatcher
// that never answers fails the test instead of hanging the suite.
const dispatcherTimeout = 2 * time.Second

// addrLoopbackAny asks the OS for any free loopback port.
const addrLoopbackAny = "127.0.0.1:0"

// serveInBackground runs Serve on its own goroutine and returns the
// channel its result arrives on.
func serveInBackground(d *sni.Dispatcher, ln net.Listener) <-chan error {
	served := make(chan error, 1)
	go func() { served <- d.Serve(ln) }()
	return served
}

// acceptInBackground runs Accept on its own goroutine and returns the
// channel its result arrives on.
func acceptInBackground(d *sni.Dispatcher) <-chan acceptResult {
	accepted := make(chan acceptResult, 1)
	go func() {
		conn, err := d.Accept()
		accepted <- acceptResult{conn, err}
	}()
	return accepted
}

type acceptResult struct {
	conn net.Conn
	err  error
}

// awaitError fails unless ch delivers before the timeout, and returns
// what it delivered.
func awaitError(t *testing.T, ch <-chan error, name string) error {
	t.Helper()

	select {
	case err := <-ch:
		return err
	case <-time.After(dispatcherTimeout):
		t.Fatalf("%s: not observed in time", name)
		return nil
	}
}

// awaitAccept fails unless ch delivers before the timeout, and returns
// what it delivered.
func awaitAccept(t *testing.T, ch <-chan acceptResult) acceptResult {
	t.Helper()

	select {
	case r := <-ch:
		return r
	case <-time.After(dispatcherTimeout):
		t.Fatal("Accept: not observed in time")
		return acceptResult{}
	}
}

// shutdownWithin shuts the Dispatcher down under the test timeout.
func shutdownWithin(t *testing.T, d *sni.Dispatcher) error {
	t.Helper()

	ctx, cancel := core.WithTimeout(context.Background(), dispatcherTimeout)
	defer cancel()
	return d.Shutdown(ctx)
}

// stubListener answers Accept with each error in turn, then blocks
// until closed and answers net.ErrClosed, the way a real listener
// does once Close has been called on it. accepting closes on the
// first Accept, so a test can wait until Serve is in place.
type stubListener struct {
	errs      chan error
	closed    chan struct{}
	accepting chan struct{}
	once      sync.Once
}

func newStubListener(errs ...error) *stubListener {
	ln := &stubListener{
		errs:      make(chan error, len(errs)),
		closed:    make(chan struct{}),
		accepting: make(chan struct{}),
	}
	for _, err := range errs {
		ln.errs <- err
	}
	return ln
}

func (ln *stubListener) Accept() (net.Conn, error) {
	ln.once.Do(func() { close(ln.accepting) })

	select {
	case err := <-ln.errs:
		return nil, err
	case <-ln.closed:
		return nil, net.ErrClosed
	}
}

// awaitServing fails unless Serve has reached ln's Accept in time.
func awaitServing(t *testing.T, ln *stubListener) {
	t.Helper()

	select {
	case <-ln.accepting:
	case <-time.After(dispatcherTimeout):
		t.Fatal("Serve did not reach Accept in time")
	}
}

func (ln *stubListener) Close() error {
	close(ln.closed)
	return nil
}

func (*stubListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1}
}

// TestDispatcherServeArguments covers Serve's own refusals: a nil
// listener, and a second listener while the first is being served.
// Before Serve there is no listener for Addr to report.
func TestDispatcherServeArguments(t *testing.T) {
	d := &sni.Dispatcher{}
	core.AssertNil(t, d.Addr(), "Addr before Serve")
	core.AssertErrorIs(t, d.Serve(nil), core.ErrInvalid, "nil listener")

	ln := newStubListener()
	served := serveInBackground(d, ln)
	awaitServing(t, ln)

	core.AssertErrorIs(t, d.Serve(newStubListener()), core.ErrExists,
		"second Serve")

	core.AssertNoError(t, ln.Close(), "Close")
	core.AssertNoError(t, awaitError(t, served, "Serve"), "Serve")
}

// TestDispatcherFallthrough verifies a connection with no dedicated
// handler reaches Accept intact, and that the listener being closed
// ends Serve cleanly. The logger's debug level is disabled, so the
// per-connection records are skipped rather than emitted.
func TestDispatcherFallthrough(t *testing.T) {
	ln, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")

	logger := mock.NewLoggerWithThreshold(slog.Error)
	d := &sni.Dispatcher{Logger: logger}
	served := serveInBackground(d, ln)

	client, err := net.Dial("tcp", ln.Addr().String())
	core.AssertMustNoError(t, err, "dial")
	defer func() { _ = client.Close() }()

	r := awaitAccept(t, acceptInBackground(d))
	core.AssertMustNoError(t, r.err, "Accept")
	defer func() { _ = r.conn.Close() }()

	// Serve has stored the listener by the time it hands out a connection.
	core.AssertEqual(t, ln.Addr().String(), d.Addr().String(), "Addr")

	// bytes travel through the accepted connection unchanged.
	_, err = client.Write([]byte("ping"))
	core.AssertMustNoError(t, err, "Write")
	buf := make([]byte, 4)
	_, err = io.ReadFull(r.conn, buf)
	core.AssertMustNoError(t, err, "ReadFull")
	core.AssertEqual(t, "ping", string(buf), "payload")

	// closing the listener underneath Serve is the clean way to end it.
	core.AssertNoError(t, ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, served, "Serve"), "Serve")
	core.AssertTrue(t, d.Cancelled(), "Cancelled")
	core.AssertNoError(t, d.Err(), "Err")

	// nothing was emitted below the logger's threshold.
	core.AssertEqual(t, 0, len(logger.GetMessages()), "records")
}

// TestDispatcherDispatch covers the SNI probe: GetHandler claiming a
// connection, declining it, and the probe failing on a stream that
// carries no ClientHello.
func TestDispatcherDispatch(t *testing.T) {
	t.Run("claimed", runTestDispatcherDispatchClaimed)
	t.Run("declined", runTestDispatcherDispatchDeclined)
	t.Run("malformed hello", runTestDispatcherDispatchMalformed)
}

// dispatchFixture is a Dispatcher serving a real loopback listener,
// with one client dialled in.
type dispatchFixture struct {
	d      *sni.Dispatcher
	ln     net.Listener
	client net.Conn
	served <-chan error
}

// newDispatchFixture starts d on a fresh listener and dials it once.
func newDispatchFixture(t *testing.T, d *sni.Dispatcher) *dispatchFixture {
	t.Helper()

	ln, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	t.Cleanup(func() { _ = ln.Close() })

	served := serveInBackground(d, ln)

	client, err := net.Dial("tcp", ln.Addr().String())
	core.AssertMustNoError(t, err, "dial")
	t.Cleanup(func() { _ = client.Close() })

	return &dispatchFixture{d: d, ln: ln, client: client, served: served}
}

// sendClientHello runs a TLS handshake from the client so a
// ClientHello for name reaches the Dispatcher. Meant to run in its
// own goroutine: the handshake never completes and its failure is
// not the subject.
func (f *dispatchFixture) sendClientHello(name string) {
	_ = tls.Client(f.client, &tls.Config{
		ServerName: name,
		MinVersion: tls.VersionTLS12,
	}).Handshake()
}

// finish shuts the Dispatcher down cleanly, closes the listener so
// Serve returns, and checks both report a clean stop.
func (f *dispatchFixture) finish(t *testing.T) {
	t.Helper()

	core.AssertNoError(t, shutdownWithin(t, f.d), "Shutdown")
	core.AssertNoError(t, f.ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, f.served, "Serve"), "Serve")
}

// runTestDispatcherDispatchClaimed: GetHandler claims the connection by
// its SNI, receiving the stream with the ClientHello still unread, and
// the handler's clean return is logged.
func runTestDispatcherDispatchClaimed(t *testing.T) {
	t.Helper()

	logger := mock.NewLogger()
	names := make(chan string, 1)

	d := &sni.Dispatcher{
		Logger: logger,
		GetHandler: func(chi *tls.ClientHelloInfo) sni.Handler {
			return func(_ context.Context, conn net.Conn) error {
				defer func() { _ = conn.Close() }()
				names <- chi.ServerName
				return nil
			}
		},
	}
	f := newDispatchFixture(t, d)
	go f.sendClientHello("example.com")

	select {
	case name := <-names:
		core.AssertEqual(t, "example.com", name, "ServerName")
	case <-time.After(dispatcherTimeout):
		t.Fatal("handler was not reached in time")
	}

	f.finish(t)
	assertLogged(t, logger, slog.Debug, "done")
}

// runTestDispatcherDispatchDeclined: GetHandler returns nil, so the
// connection falls through to Accept with the ClientHello replayed
// ahead of the rest of the stream.
func runTestDispatcherDispatchDeclined(t *testing.T) {
	t.Helper()

	d := &sni.Dispatcher{
		GetHandler: func(*tls.ClientHelloInfo) sni.Handler { return nil },
	}
	f := newDispatchFixture(t, d)
	go f.sendClientHello("example.com")

	r := awaitAccept(t, acceptInBackground(d))
	core.AssertMustNoError(t, r.err, "Accept")
	defer func() { _ = r.conn.Close() }()

	// the first byte is the TLS handshake record type the probe consumed.
	buf := make([]byte, 1)
	_, err := r.conn.Read(buf)
	core.AssertMustNoError(t, err, "Read")
	core.AssertEqual(t, byte(0x16), buf[0], "record type")

	f.finish(t)
}

// runTestDispatcherDispatchMalformed: the probe fails on a stream that
// is not TLS, the connection is closed, and the failure is logged and
// absorbed like any other handler error.
func runTestDispatcherDispatchMalformed(t *testing.T) {
	t.Helper()

	logger := mock.NewLogger()
	d := &sni.Dispatcher{
		Logger:     logger,
		GetHandler: func(*tls.ClientHelloInfo) sni.Handler { return nil },
	}
	f := newDispatchFixture(t, d)

	_, err := f.client.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
	core.AssertMustNoError(t, err, "Write")

	// the Dispatcher closes the connection once the probe fails.
	core.AssertNoError(t, f.client.SetReadDeadline(time.Now().Add(dispatcherTimeout)),
		"deadline")
	_, err = f.client.Read(make([]byte, 1))
	core.AssertError(t, err, "peer Read")
	core.AssertNotErrorIs(t, err, context.DeadlineExceeded, "peer Read")

	f.finish(t)
	core.AssertNoError(t, d.Err(), "Err")
	assertLogged(t, logger, slog.Error, "")
}

// TestDispatcherShutdownContextDone verifies Shutdown gives up with the
// context's error while a handler is still running.
func TestDispatcherShutdownContextDone(t *testing.T) {
	release := make(chan struct{})
	running := make(chan struct{})

	d := &sni.Dispatcher{
		GetHandler: func(*tls.ClientHelloInfo) sni.Handler {
			return func(_ context.Context, conn net.Conn) error {
				defer func() { _ = conn.Close() }()
				close(running)
				<-release
				return nil
			}
		},
	}
	f := newDispatchFixture(t, d)
	go f.sendClientHello("example.com")

	select {
	case <-running:
	case <-time.After(dispatcherTimeout):
		t.Fatal("handler was not reached in time")
	}

	expired, cancel := context.WithCancel(context.Background())
	cancel()
	core.AssertErrorIs(t, d.Shutdown(expired), context.Canceled, "Shutdown")

	close(release)
	core.AssertNoError(t, f.ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, f.served, "Serve"), "Serve")
	core.AssertNoError(t, d.Wait(), "Wait")
}

// TestDispatcherShutdownUnblocksPeek verifies Shutdown releases a
// handler parked in the SNI probe by a client that connected but never
// sent its ClientHello, and that the interrupted probe is recorded as
// a drop rather than an error.
func TestDispatcherShutdownUnblocksPeek(t *testing.T) {
	logger := mock.NewLogger()
	d := &sni.Dispatcher{
		Logger:     logger,
		GetHandler: func(*tls.ClientHelloInfo) sni.Handler { return nil },
	}
	f := newDispatchFixture(t, d)

	// Shutdown waits for the handler, so it returning in time proves
	// the parked probe was released.
	core.AssertNoError(t, shutdownWithin(t, d), "Shutdown")
	core.AssertNoError(t, d.Err(), "Err")

	// and the peer observes the connection closed.
	core.AssertNoError(t, f.client.SetReadDeadline(time.Now().Add(dispatcherTimeout)),
		"deadline")
	_, err := f.client.Read(make([]byte, 1))
	core.AssertError(t, err, "peer Read")
	core.AssertNotErrorIs(t, err, context.DeadlineExceeded, "peer Read")

	core.AssertNoError(t, f.ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, f.served, "Serve"), "Serve")

	assertLogged(t, logger, slog.Debug, "dropped")
	for _, msg := range logger.GetMessages() {
		core.AssertNotEqual(t, slog.Error, msg.Level, "record %q level", msg.Message)
	}
}

// TestDispatcherConnAfterCancel verifies a connection the listener
// yields after the Dispatcher was cancelled is dropped rather than
// handled, since the workgroup refuses to enrol its handler, and the
// drop is recorded.
func TestDispatcherConnAfterCancel(t *testing.T) {
	logger := mock.NewLogger()
	d := &sni.Dispatcher{Logger: logger}
	d.Cancel()

	f := newDispatchFixture(t, d)

	core.AssertNoError(t, f.client.SetReadDeadline(time.Now().Add(dispatcherTimeout)),
		"deadline")
	_, err := f.client.Read(make([]byte, 1))
	core.AssertError(t, err, "peer Read")
	core.AssertNotErrorIs(t, err, context.DeadlineExceeded, "peer Read")

	core.AssertNoError(t, f.ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, f.served, "Serve"), "Serve")
	assertLogged(t, logger, slog.Debug, "dropped")
}

// assertLogged fails unless logger recorded a message at level whose
// text contains want; an empty want accepts any message at that level,
// for records whose text belongs to another package.
func assertLogged(t *testing.T, logger *mock.Logger, level slog.LogLevel, want string) {
	t.Helper()

	for _, msg := range logger.GetMessages() {
		if msg.Level == level && strings.Contains(msg.Message, want) {
			return
		}
	}
	t.Errorf("no %v record containing %q", level, want)
}

// TestDispatcherAcceptAfterShutdown verifies a blocked Accept returns
// net.ErrClosed once the Dispatcher shuts down cleanly, and keeps
// answering so afterwards.
func TestDispatcherAcceptAfterShutdown(t *testing.T) {
	d := &sni.Dispatcher{}
	accepted := acceptInBackground(d)

	core.AssertNoError(t, shutdownWithin(t, d), "Shutdown")

	r := awaitAccept(t, accepted)
	core.AssertNil(t, r.conn, "conn")
	core.AssertErrorIs(t, r.err, net.ErrClosed, "Accept")

	_, err := d.Accept()
	core.AssertErrorIs(t, err, net.ErrClosed, "Accept again")
	core.AssertNoError(t, d.Close(), "Close")
	core.AssertNoError(t, d.Wait(), "Wait")
}

// TestDispatcherUncollectedConnClosed verifies a connection nobody
// collects through Accept is dropped on shutdown rather than holding
// the handler, and the Dispatcher, open; and that the drop is recorded.
func TestDispatcherUncollectedConnClosed(t *testing.T) {
	ln, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = ln.Close() }()

	logger := mock.NewLogger()
	handled := make(chan struct{})
	d := &sni.Dispatcher{
		Logger: logger,
		OnAccept: func(conn net.Conn) (net.Conn, error) {
			close(handled)
			return conn, nil
		},
	}
	served := serveInBackground(d, ln)

	client, err := net.Dial("tcp", ln.Addr().String())
	core.AssertMustNoError(t, err, "dial")
	defer func() { _ = client.Close() }()

	select {
	case <-handled:
	case <-time.After(dispatcherTimeout):
		t.Fatal("handler was not reached in time")
	}

	// Shutdown waits for the handler, so it returning in time proves
	// the parked hand-off was released.
	core.AssertNoError(t, shutdownWithin(t, d), "Shutdown")

	// and the peer observes the connection closed.
	core.AssertNoError(t, client.SetReadDeadline(time.Now().Add(dispatcherTimeout)),
		"deadline")
	_, err = client.Read(make([]byte, 1))
	core.AssertError(t, err, "peer Read")
	core.AssertNotErrorIs(t, err, context.DeadlineExceeded, "peer Read")

	core.AssertNoError(t, ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, served, "Serve"), "Serve")
	assertLogged(t, logger, slog.Debug, "dropped")
}

// TestDispatcherAcceptError covers an error from the underlying
// listener's Accept in each OnError setting.
func TestDispatcherAcceptError(t *testing.T) {
	t.Run("terminates by default", runTestDispatcherAcceptErrorTerminates)
	t.Run("continues when OnError objects", runTestDispatcherAcceptErrorContinues)
	t.Run("shutdown interrupts the retry pause", runTestDispatcherAcceptErrorPaused)
}

// runTestDispatcherAcceptErrorTerminates: with OnError unset, the
// accept error is fatal. Serve, Err, Close and a blocked Accept all
// report it, and the record names the accept.
func runTestDispatcherAcceptErrorTerminates(t *testing.T) {
	t.Helper()

	errBoom := errors.New("accept boom")
	logger := mock.NewLogger()
	ln := newStubListener(errBoom)

	d := &sni.Dispatcher{Logger: logger}
	accepted := acceptInBackground(d)
	served := serveInBackground(d, ln)

	core.AssertErrorIs(t, awaitError(t, served, "Serve"), errBoom, "Serve")
	core.AssertErrorIs(t, d.Err(), errBoom, "Err")
	core.AssertErrorIs(t, d.Close(), errBoom, "Close")

	r := awaitAccept(t, accepted)
	core.AssertErrorIs(t, r.err, errBoom, "Accept")

	assertLogged(t, logger, slog.Error, "accept")
}

// runTestDispatcherAcceptErrorContinues: OnError sees the accept error
// and declines to terminate, so Serve keeps accepting, at a bounded
// pace, until the listener closes, and nothing fatal is recorded.
func runTestDispatcherAcceptErrorContinues(t *testing.T) {
	t.Helper()

	errBoom := errors.New("accept boom")
	seen := make(chan error, 1)
	ln := newStubListener(errBoom)

	d := &sni.Dispatcher{
		OnError: func(err error) bool {
			seen <- err
			return false
		},
	}
	served := serveInBackground(d, ln)

	core.AssertErrorIs(t, awaitError(t, seen, "OnError"), errBoom, "OnError")
	core.AssertFalse(t, d.Cancelled(), "Cancelled")

	core.AssertNoError(t, ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, served, "Serve"), "Serve")
	core.AssertNoError(t, d.Err(), "Err")
}

// runTestDispatcherAcceptErrorPaused: OnError waives the accept error
// and shuts the Dispatcher down, so Serve returns from the retry pause
// without the listener ever being closed. The stub's Accept blocks
// once its error is spent, so a pause the shutdown did not interrupt
// would leave Serve parked there.
func runTestDispatcherAcceptErrorPaused(t *testing.T) {
	t.Helper()

	errBoom := errors.New("accept boom")
	ln := newStubListener(errBoom)

	var d sni.Dispatcher
	d.OnError = func(error) bool {
		d.Cancel()
		return false
	}
	served := serveInBackground(&d, ln)

	core.AssertNoError(t, awaitError(t, served, "Serve"), "Serve")
	core.AssertTrue(t, d.Cancelled(), "Cancelled")
	core.AssertNoError(t, d.Err(), "Err")

	core.AssertNoError(t, ln.Close(), "Close listener")
}

// TestDispatcherHandlerError covers a handler failing on a connection,
// here OnAccept rejecting it, in each OnError setting.
func TestDispatcherHandlerError(t *testing.T) {
	t.Run("absorbed by default", runTestDispatcherHandlerErrorAbsorbed)
	t.Run("terminates when OnError asks", runTestDispatcherHandlerErrorTerminates)
}

// dialTwice dials the listener twice: OnAccept rejects whichever is
// handled first, and the other shows whether the Dispatcher is still
// serving. Which is which does not matter to the caller.
func dialTwice(t *testing.T, addr string) (first, second net.Conn) {
	t.Helper()

	first, err := net.Dial("tcp", addr)
	core.AssertMustNoError(t, err, "first dial")
	second, err = net.Dial("tcp", addr)
	core.AssertMustNoError(t, err, "second dial")
	return first, second
}

// runTestDispatcherHandlerErrorAbsorbed: with OnError unset the
// rejection is logged with its peer and absorbed; the next connection
// still falls through to Accept and nothing fatal is recorded.
func runTestDispatcherHandlerErrorAbsorbed(t *testing.T) {
	t.Helper()

	ln, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = ln.Close() }()

	errReject := errors.New("rejected")
	logger := mock.NewLogger()

	d := &sni.Dispatcher{
		Logger:   logger,
		OnAccept: newRejectFirst(errReject),
	}
	served := serveInBackground(d, ln)

	first, second := dialTwice(t, ln.Addr().String())
	defer func() { _ = first.Close() }()
	defer func() { _ = second.Close() }()

	r := awaitAccept(t, acceptInBackground(d))
	core.AssertMustNoError(t, r.err, "Accept")
	_ = r.conn.Close()

	core.AssertNoError(t, shutdownWithin(t, d), "Shutdown")
	core.AssertNoError(t, d.Err(), "Err")
	core.AssertNoError(t, ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, served, "Serve"), "Serve")

	assertLogged(t, logger, slog.Error, errReject.Error())
}

// runTestDispatcherHandlerErrorTerminates: OnError asks to terminate on
// the rejection, so it becomes the fatal error Accept and Err report.
func runTestDispatcherHandlerErrorTerminates(t *testing.T) {
	t.Helper()

	ln, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = ln.Close() }()

	errReject := errors.New("rejected")

	d := &sni.Dispatcher{
		OnAccept: newRejectFirst(errReject),
		OnError:  func(error) bool { return true },
	}
	served := serveInBackground(d, ln)

	client, err := net.Dial("tcp", ln.Addr().String())
	core.AssertMustNoError(t, err, "dial")
	defer func() { _ = client.Close() }()

	r := awaitAccept(t, acceptInBackground(d))
	core.AssertErrorIs(t, r.err, errReject, "Accept")
	core.AssertErrorIs(t, d.Wait(), errReject, "Wait")

	core.AssertNoError(t, ln.Close(), "Close listener")
	core.AssertNoError(t, awaitError(t, served, "Serve"), "Serve")
}

// newRejectFirst returns an OnAccept hook that rejects the first
// connection with err and passes the rest through. Handlers run
// concurrently, so the first is decided by an atomic swap.
func newRejectFirst(err error) func(net.Conn) (net.Conn, error) {
	var rejected atomic.Bool
	return func(conn net.Conn) (net.Conn, error) {
		if rejected.CompareAndSwap(false, true) {
			return nil, err
		}
		return conn, nil
	}
}
