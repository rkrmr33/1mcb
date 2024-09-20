package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rkrmr33/1mcb/pkg/util"
)

type (
	eventHandler struct {
		mux  *sync.Mutex
		subs map[string]http.ResponseWriter
	}

	event struct {
		Type    eventType    `json:"type"`
		Payload eventPayload `json:"payload"`
	}

	eventPayload interface{}
	eventType    int

	keepaliveEvent struct{}

	helloEvent struct {
		TogglerID string `json:"togglerId"`
	}

	updateEvent struct {
		Index int  `json:"index"`
		State bool `json:"state"`
	}
)

const (
	KeepAliveEvent = iota
	HelloEvent
	UpdateEvent
)

var keepalivePayload = []byte(`{"type":0,"payload":{}}`)

func newEventHandler() *eventHandler {
	return &eventHandler{
		mux:  &sync.Mutex{},
		subs: map[string]http.ResponseWriter{},
	}
}

func (h *eventHandler) subscribe(w http.ResponseWriter, r *http.Request) {
	id := util.GetUID(r.RemoteAddr)
	h.mux.Lock()
	h.subs[id] = w
	h.mux.Unlock()

	// greet the user with their uid
	hello := event{Type: HelloEvent, Payload: helloEvent{TogglerID: id}}
	payload, _ := json.Marshal(hello)
	h.writePayload(id, w, payload)

	<-r.Context().Done()

	h.mux.Lock()
	delete(h.subs, id)
	h.mux.Unlock()
}

func (h *eventHandler) broadcastExcluding(e event, excludedID string) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	fmt.Printf("%s broadcasting event: %v\n", time.Now().Format(time.RFC3339), e)

	h.mux.Lock()
	wg.Add(len(h.subs))

	count := 0
	for id, w := range h.subs {
		if id == excludedID {
			wg.Done()
			continue
		}

		count++

		go func() {
			defer wg.Done()
			h.writePayload(id, w, payload)
		}()
	}
	h.mux.Unlock()

	wg.Wait()
	fmt.Printf("%s finished broadcasting event to: %d\n", time.Now().Format(time.RFC3339), count)

	return nil
}

func (h *eventHandler) writePayload(addr string, w http.ResponseWriter, payload []byte) {
	fmt.Printf("%s     --> event %s %s\n", time.Now().Format(time.RFC3339), string(payload), addr)

	_, err := fmt.Fprintf(w, "data: %s\n\n", payload)
	if err != nil {
		fmt.Printf("  failed writing to %s, removing subscription\n", addr)

		h.mux.Lock()
		delete(h.subs, addr)
		h.mux.Unlock()
	}

	w.(http.Flusher).Flush()
}
