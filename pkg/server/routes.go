package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type (
	toggleRequest struct {
		Index     int
		State     bool
		TogglerID string
	}
)

func (s *server) rootHandler(w http.ResponseWriter, r *http.Request) {
	tpl := s.templates.Lookup("index.html")
	if tpl == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title":      "1 Million Checkboxes",
		"Checkboxes": s.state.clone(),
	}

	if err := tpl.Execute(w, data); err != nil {
		fmt.Println("Failed to execute template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *server) toggleHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if errors.Is(err, &http.MaxBytesError{}) {
		http.Error(w, (&http.MaxBytesError{}).Error(), http.StatusBadRequest)
		return
	}

	req := toggleRequest{}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "failed to parse request", http.StatusBadRequest)
		fmt.Printf("failed to parse request body: %s\n", err)
		return
	}

	// update state
	s.state.update(req.Index, req.State)

	// broadcast
	e := event{Type: UpdateEvent, Payload: updateEvent{Index: req.Index, State: req.State}}
	go func(e event, id string) {
		if err := s.eventHandler.broadcastExcluding(e, id); err != nil {
			fmt.Printf("failed to broadcast: %s", err)
		}
	}(e, req.TogglerID)

	w.WriteHeader(http.StatusCreated)
}

func (s *server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(200)

	s.eventHandler.subscribe(w, r)
}

func (s *server) getStateHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Write(s.state.clone())
}
