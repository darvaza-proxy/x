package reconnect

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/fs"

	"darvaza.org/core"
)

var (
	_ io.Closer = (*StreamSession[any, any])(nil)
)

// StreamSession provides an asynchronous stream session
// using message types for receiving and sending.
type StreamSession[Input, Output any] struct {
	wg  core.ErrGroup
	in  chan Input
	out chan Output

	// QueueSize specifies how many [Output] type entries can be buffered
	// for delivery before [StreamSession.Send] blocks.
	QueueSize uint
	// Conn specifies the underlying connection
	Conn io.ReadWriteCloser
	// Context is an optional [context.Context] to allow cascading cancellations.
	Context context.Context

	// Split identifies the next encoded [Input] type in the inbound stream.
	// If not set, [bufio.SplitLine] will be used.
	Split bufio.SplitFunc
	// Marshal is used, if MarshalTo isn't set, to encode an [Output] type.
	// If neither is set, [StreamSession.Go] will fail.
	Marshal func(Output) ([]byte, error)
	// MarshalTo, if set, is used to write the encoded representation of
	// and [Output] type.
	MarshalTo func(Output, io.Writer) error
	// Unmarshal is used to decode an [Input] type previously identified
	// by [StreamSession.Split].
	// If not net, [StreamSession.Go] will fail.
	Unmarshal func([]byte) (Input, error)

	// SetReadDeadline is an optional hook called before reading the a message
	SetReadDeadline func() error
	// SetWriteDeadline is an optional hook called before writing a message
	SetWriteDeadline func() error
	// UnsetReadDeadline is an optional hook called after having read a message
	UnsetReadDeadline func() error
	// UnsetWriteDeadline is an optional hook called after having wrote a message
	UnsetWriteDeadline func() error

	// OnError is optionally called when an error occurs
	OnError func(error)
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

	if err := s.setDefaults(); err != nil {
		return err
	}

	s.wg = core.ErrGroup{
		Parent: s.Context,
	}

	s.wg.OnError(s.OnError)
	s.wg.SetDefaults()

	s.in = make(chan Input)
	s.out = make(chan Output, s.QueueSize)
	return nil
}

// revive:disable:cognitive-complexity
func (s *StreamSession[_, _]) setDefaults() error {
	// revive:enable:cognitive-complexity
	if s.Context == nil {
		s.Context = context.Background()
	}

	if s.Split == nil {
		s.Split = bufio.ScanLines
	}

	if s.MarshalTo == nil {
		s.MarshalTo = newMarshalTo(s.Marshal)
	}

	if s.SetReadDeadline == nil {
		s.SetReadDeadline = func() error { return nil }
	}

	if s.SetWriteDeadline == nil {
		s.SetWriteDeadline = func() error { return nil }
	}

	if s.UnsetReadDeadline == nil {
		s.UnsetReadDeadline = func() error { return nil }
	}

	if s.UnsetWriteDeadline == nil {
		s.UnsetWriteDeadline = func() error { return nil }
	}

	return nil
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

// Spawn starts the [StreamSession].
func (s *StreamSession[_, _]) Spawn() error {
	if err := s.init(); err != nil {
		return err
	}

	barrier := make(chan struct{})

	// reader
	s.wg.Go(func(ctx context.Context) error {
		close(barrier)
		return s.runReader(ctx)
	}, func() error {
		return s.Conn.Close()
	})
	// writer
	s.wg.Go(s.runWriter, s.killWriter)

	<-barrier
	return nil
}

func (s *StreamSession[_, _]) runReader(_ context.Context) error {
	r := bufio.NewScanner(s.Conn)
	r.Split(s.Split)

	defer close(s.in)

	if err := s.SetReadDeadline(); err != nil {
		return err
	}

	for r.Scan() {
		if err := s.readerStep(r.Bytes()); err != nil {
			return err
		}
	}

	return r.Err()
}

func (s *StreamSession[_, _]) readerStep(raw []byte) error {
	if err := s.UnsetReadDeadline(); err != nil {
		return err
	}

	msg, err := s.Unmarshal(raw)
	if err != nil {
		return err
	}

	s.in <- msg

	return s.SetReadDeadline()
}

func (s *StreamSession[_, _]) runWriter(_ context.Context) error {
	for req := range s.out {
		if err := s.SetWriteDeadline(); err != nil {
			return err
		}

		if err := s.MarshalTo(req, s.Conn); err != nil {
			return err
		}

		if err := s.UnsetWriteDeadline(); err != nil {
			return err
		}
	}

	return nil
}

func (s *StreamSession[_, _]) killWriter() error {
	close(s.out)
	return nil
}

// Go spawns a goroutine within the session's context.
func (s *StreamSession[_, _]) Go(fn func(context.Context) error) {
	mustStarted(s)

	s.wg.Go(fn, nil)
}

// Close initiates a shutdown of the session.
func (s *StreamSession[_, _]) Close() error {
	mustStarted(s)

	s.wg.Cancel(nil)
	return nil
}

// Wait blocks until all workers are done.
func (s *StreamSession[_, _]) Wait() error {
	mustStarted(s)

	return s.wg.Wait()
}

// Done returns a channel that will be closed with all workers are done
func (s *StreamSession[_, _]) Done() <-chan struct{} {
	mustStarted(s)

	return s.wg.Done()
}

// Send sends a message asynchronously, unless the queue is full.
func (s *StreamSession[_, Output]) Send(m Output) error {
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

// Recv returns a channel where inbound messages can be received.
func (s *StreamSession[Input, _]) Recv() <-chan Input {
	mustStarted(s)

	return s.in
}

// Next blocks until a new message is received
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
