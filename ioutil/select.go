package ioutil

import (
	"io"
	"time"
)

type selectMessage struct {
	s *selectReader
	n int
}

type Select struct {
	c chan selectMessage
}

// NewSelect returns a *Select which will select
// over the readers in r. Since select spawns
// goroutines to read from the readers in r,
// these readers should now be considered invalid.
// Instead, use the readers in the second return
// value, which will behave properly.
func NewSelect(r ...io.Reader) (*Select, []io.Reader) {
	c := make(chan selectMessage)
	ret := []io.Reader{}
	for i, v := range r {
		s := selectMessage{
			&selectReader{
				v,
				make([]byte, 1024),
				nil,
				nil,
				make(chan struct{}, 1),
			},
			i,
		}
		s.s.read = s.s.readBuf
		ret = append(ret, s.s)
		go read(s, c)
	}
	return &Select{c}, ret
}

func read(s selectMessage, c chan selectMessage) {
	n, err := s.s.r.Read(s.s.buf)
	s.s.buf, s.s.err = s.s.buf[:n], err
	s.s.c <- struct{}{}
	c <- s
}

// SelectTimeout waits until at least one reader
// is available to be read from, or the duration
// d has elapsed, whichever occurs first. If a
// reader is available, its index in the slice
// returned by NewSelect will be n, and as a convenience,
// the reader itself will be r. If the timeout occurs,
// r will be nil.
func (s *Select) SelectTimeout(d time.Duration) (n int, r io.Reader) {
	select {
	case sm := <-s.c:
		return sm.n, sm.s
	case <-time.After(d):
		return 0, nil
	}
}

// Select is equivalent to SelectTimeout, except that
// it will never time out, blocking forever until at least
// one reader is available to be read from.
func (s *Select) Select() (n int, r io.Reader) {
	sm := <-s.c
	return sm.n, sm.s
}

type selectReader struct {
	r   io.Reader
	buf []byte
	err error

	read func([]byte) (int, error)

	// Should be buffered
	c chan struct{}
}

func (s *selectReader) Read(p []byte) (int, error) {
	// This blocks until something is written
	// to the channel for the first time, and
	// then never blocks again since that thing
	// stays in the channel.
	s.c <- (<-s.c)
	return s.read(p)
}

func (s *selectReader) readBuf(p []byte) (int, error) {
	if len(s.buf) > 0 {
		n := copy(p, s.buf)
		s.buf = s.buf[n:]

		if len(s.buf) > 0 {
			return n, nil
		}

		s.read = s.readReader
		return n, s.err
	}

	s.read = s.readReader
	if s.err != nil {
		return 0, s.err
	}
	return s.r.Read(p)
}

func (s *selectReader) readReader(p []byte) (int, error) {
	return s.r.Read(p)
}
