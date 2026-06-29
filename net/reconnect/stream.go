package reconnect

import (
	"bufio"
	"bytes"
	"context"
	"io"

	"darvaza.org/core"
	"darvaza.org/x/fs"
	"darvaza.org/x/sync/workgroup"
)

var (
	_ io.Closer = (*StreamSession[any, any])(nil)
	_ WorkGroup = (*StreamSession[any, any])(nil)
)

// StreamSession provides an asynchronous stream session
// using message types for receiving and sending.
// Exported fields are configured before calling
// [StreamSession.Spawn] and must not be modified afterwards.
// The session must be spawned before using any other method.
type StreamSession[Input, Output any] struct {
	in  chan Input
	out chan Output

	// Conn specifies the underlying connection
	Conn io.ReadWriteCloser
	// Context is an optional [context.Context] to allow cascading cancellations.
	Context context.Context

	// Split identifies the next encoded [Input] type in the inbound stream.
	// If not set, [bufio.ScanLines] will be used.
	Split bufio.SplitFunc
	// Marshal is used, if MarshalTo isn't set, to encode an [Output] type.
	// If neither is set, [StreamSession.Spawn] will fail.
	Marshal func(Output) ([]byte, error)
	// MarshalTo, if set, is used to write the encoded representation of
	// an [Output] type.
	MarshalTo func(Output, io.Writer) error
	// Unmarshal is used to decode an [Input] type previously identified
	// by [StreamSession.Split].
	// If not set, [StreamSession.Spawn] will fail.
	Unmarshal func([]byte) (Input, error)

	// SetReadDeadline is an optional hook called before reading a message
	SetReadDeadline func() error
	// SetWriteDeadline is an optional hook called before writing a message
	SetWriteDeadline func() error
	// UnsetReadDeadline is an optional hook called after having read a message
	UnsetReadDeadline func() error
	// UnsetWriteDeadline is an optional hook called after having written a message
	UnsetWriteDeadline func() error

	// OnError is optionally called when an error occurs
	OnError func(error)

	wg workgroup.Group

	// QueueSize specifies how many [Output] type entries can be buffered
	// for delivery before [StreamSession.Send] blocks.
	QueueSize uint
}

func (s *StreamSession[Input, Output]) init() error {
	switch {
	case s.in != nil:
		return core.QuietWrap(fs.ErrExist, "session already started")
	case s.Conn == nil:
		return core.QuietWrap(fs.ErrInvalid, "missing Conn")
	case s.Unmarshal == nil:
		return core.QuietWrap(fs.ErrInvalid, "missing Unmarshal")
	case s.Marshal == nil && s.MarshalTo == nil:
		return core.QuietWrap(fs.ErrInvalid, "missing Marshal/MarshalTo")
	}

	s.setDefaults()

	if fn := s.OnError; fn != nil {
		// the former core.ErrGroup delivered the cancellation cause
		// to OnError; OnCancel is its workgroup.Group equivalent.
		s.wg.OnCancel = func(_ context.Context, cause error) {
			fn(cause)
		}
	}

	s.in = make(chan Input)
	s.out = make(chan Output, s.QueueSize)
	return nil
}

// noopHook is the default for the optional read/write deadline hooks.
var noopHook = func() error { return nil }

func (s *StreamSession[_, _]) setDefaults() {
	s.Context = core.Coalesce(s.Context, context.Background())
	if s.wg.Parent == nil {
		s.wg.Parent = s.Context
	}
	_ = s.wg.Context() // ensure the context is initialised

	s.Split = core.Coalesce(s.Split, bufio.ScanLines)
	if s.MarshalTo == nil {
		// keep the branch: newMarshalTo panics on a nil Marshal, so it
		// must not be evaluated when MarshalTo is already set.
		s.MarshalTo = newMarshalTo(s.Marshal)
	}

	s.SetReadDeadline = core.Coalesce(s.SetReadDeadline, noopHook)
	s.SetWriteDeadline = core.Coalesce(s.SetWriteDeadline, noopHook)
	s.UnsetReadDeadline = core.Coalesce(s.UnsetReadDeadline, noopHook)
	s.UnsetWriteDeadline = core.Coalesce(s.UnsetWriteDeadline, noopHook)
}

func newMarshalTo[T any](fn func(T) ([]byte, error)) func(T, io.Writer) error {
	if fn == nil {
		panic("unreachable")
	}

	return func(v T, w io.Writer) error {
		return doMarshalTo[T](v, w, fn)
	}
}

func doMarshalTo[T any](v T, w io.Writer, fn func(T) ([]byte, error)) error {
	b, err := fn(v)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	for buf.Len() > 0 {
		if _, err = buf.WriteTo(w); err != nil {
			return err
		}
	}
	return nil
}

// Spawn starts the [StreamSession]'s workers. It fails if the
// session has already been started, or if Conn, Unmarshal, or
// a marshalling function is missing.
func (s *StreamSession[_, _]) Spawn() error {
	if err := s.init(); err != nil {
		return err
	}

	s.goWithKill(s.runReader, s.killReader)
	s.goWithKill(s.runWriter, s.killWriter)
	return nil
}

// goWithKill supervises run and fires kill once the group's context is
// cancelled, replacing the shutdown argument of the former
// core.ErrGroup.Go so a worker blocked on I/O can unwind.
//
// The kill watcher is enrolled before run so an early cancellation —
// the reader reaching EOF and winding the group down before the writer
// is wired up — cannot drop it. If the group is already cancelled the
// watcher cannot be enrolled, so run is never started and kill closes
// the resource directly.
func (s *StreamSession[_, _]) goWithKill(run WorkerFunc, kill func() error) {
	if err := s.wg.Go(func(ctx context.Context) {
		<-ctx.Done()
		_ = kill()
	}); err != nil {
		_ = kill()
		return
	}

	_ = s.wg.GoCatch(run, nil)
}

func (s *StreamSession[_, _]) runReader(ctx context.Context) error {
	r := bufio.NewScanner(s.Conn)
	r.Split(s.Split)

	defer close(s.in)

	if err := s.SetReadDeadline(); err != nil {
		return err
	}

	for r.Scan() {
		if err := s.readerStep(ctx, r.Bytes()); err != nil {
			return err
		}
	}

	if err := r.Err(); err != nil {
		return err
	}

	// A clean EOF ends the inbound stream but does not cancel the group
	// on its own; do it here so the writer and the kill watchers unwind
	// instead of parking forever and leaving Wait to block.
	s.wg.Cancel(nil)
	return nil
}

func (s *StreamSession[_, _]) readerStep(ctx context.Context, raw []byte) error {
	if err := s.UnsetReadDeadline(); err != nil {
		return err
	}

	msg, err := s.Unmarshal(raw)
	if err != nil {
		return err
	}

	select {
	case s.in <- msg:
	case <-ctx.Done():
		// A shutdown raced the delivery. Stop reading rather than
		// parking on the unbuffered channel, which closing the
		// connection would not unblock.
		return context.Cause(ctx)
	}

	return s.SetReadDeadline()
}

func (s *StreamSession[_, _]) killReader() error {
	return s.Conn.Close()
}

func (s *StreamSession[_, _]) runWriter(_ context.Context) error {
	for req := range s.out {
		if err := s.writeOne(req); err != nil {
			return err
		}
	}
	return nil
}

func (s *StreamSession[_, Output]) writeOne(req Output) error {
	if err := s.SetWriteDeadline(); err != nil {
		return err
	}

	if err := s.MarshalTo(req, s.Conn); err != nil {
		return err
	}

	if f, ok := s.Conn.(fs.Flusher); ok {
		if err := f.Flush(); err != nil {
			return err
		}
	}

	return s.UnsetWriteDeadline()
}

func (s *StreamSession[_, _]) killWriter() error {
	close(s.out)
	return nil
}

// Go spawns a goroutine within the session's context.
func (s *StreamSession[_, _]) Go(funcs ...WorkerFunc) {
	mustStarted(s)

	for _, fn := range funcs {
		if fn != nil {
			_ = s.wg.GoCatch(fn, nil)
		}
	}
}

// GoCatch spawns a goroutine within the session's context,
// and allows a catcher function to filter returned errors.
func (s *StreamSession[_, _]) GoCatch(run WorkerFunc, catch CatcherFunc) {
	mustStarted(s)

	if run != nil {
		_ = s.wg.GoCatch(run, catch)
	}
}

// Close initiates a shutdown of the session.
func (s *StreamSession[_, _]) Close() error {
	mustStarted(s)

	s.wg.Cancel(nil)
	return nil
}

// Shutdown initiates a shutdown and waits until it's
// done or the given context has expired.
func (s *StreamSession[_, _]) Shutdown(ctx context.Context) error {
	mustStarted(s)

	s.wg.Cancel(nil)
	select {
	case <-s.Done():
		return s.wg.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Wait blocks until all workers are done.
func (s *StreamSession[_, _]) Wait() error {
	mustStarted(s)

	return s.wg.Wait()
}

// Done returns a channel that will be closed when all workers are done.
func (s *StreamSession[_, _]) Done() <-chan struct{} {
	mustStarted(s)

	return s.wg.Done()
}

// Err returns the error that initiated a shutdown.
func (s *StreamSession[Input, Output]) Err() error {
	mustStarted(s)

	return s.wg.Err()
}

// Send queues a message for asynchronous delivery, blocking
// while the queue is full. It fails with [fs.ErrClosed] once
// the session has been shut down.
func (s *StreamSession[_, Output]) Send(m Output) error {
	mustStarted(s)

	// TODO: implement TrySend() non-blocking variant via counter.
	var err error
	s.trySend(m, &err)
	return err
}

func (s *StreamSession[_, Output]) trySend(m Output, err *error) {
	defer func() {
		if e := recover(); e != nil {
			*err = fs.ErrClosed
		}
	}()

	s.out <- m
}

// Recv returns the channel where inbound messages are delivered.
// The channel is closed when the inbound stream ends.
func (s *StreamSession[Input, _]) Recv() <-chan Input {
	mustStarted(s)

	return s.in
}

// Next blocks until a new message is received, returning false
// once the inbound stream has ended.
func (s *StreamSession[Input, _]) Next() (Input, bool) {
	mustStarted(s)

	m, ok := <-s.in
	return m, ok
}

func mustStarted[Input, Output any](s *StreamSession[Input, Output]) {
	if s.in == nil {
		err := core.QuietWrap(fs.ErrInvalid, "method called before starting the session")
		core.Panic(err)
	}
}
