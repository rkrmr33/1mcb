package server

import (
	"bytes"
	"sync"
)

type (
	state struct {
		mux *sync.RWMutex
		buf []byte
	}
)

func newState(buf []byte) state {
	return state{
		mux: &sync.RWMutex{},
		buf: buf,
	}
}

func (s state) update(i int, state bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	bucketIdx := i / 8
	bitIdx := i % 8
	if state {
		s.buf[bucketIdx] |= 1 << bitIdx
	} else {
		s.buf[bucketIdx] &= ^(1 << bitIdx)
	}
}

func (s state) clone() []byte {
	s.mux.RLock()
	defer s.mux.RUnlock()

	copy := bytes.NewBuffer(nil)
	_, _ = bytes.NewReader(s.buf).WriteTo(copy)

	return copy.Bytes()
}
