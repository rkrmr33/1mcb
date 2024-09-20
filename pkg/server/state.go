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
	var newState byte
	if state {
		newState = 1
	}
	s.buf[i] = newState
	s.mux.Unlock()
}

func (s state) clone() []byte {
	s.mux.RLock()
	defer s.mux.RUnlock()
	copy := bytes.NewBuffer(nil)
	_, _ = bytes.NewReader(s.buf).WriteTo(copy)

	return copy.Bytes()
}
