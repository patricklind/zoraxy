package configstore

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxConfigPayload = 16 << 20

type Validator func(json.RawMessage) error
type ActorResolver func(*http.Request) string

type HTTPHandler struct {
	store    Store
	validate Validator
	actor    ActorResolver
}

func NewHTTPHandler(store Store, validate Validator, actor ActorResolver) *HTTPHandler {
	return &HTTPHandler{store: store, validate: validate, actor: actor}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getCurrent(w, r)
	case http.MethodPut:
		h.commit(w, r)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) getCurrent(w http.ResponseWriter, r *http.Request) {
	revision, err := h.store.Current(r.Context())
	if errors.Is(err, ErrNoRevision) {
		http.Error(w, "no configuration revision exists", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "configuration store unavailable", http.StatusServiceUnavailable)
		return
	}
	SetRevisionHeaders(w, revision.ID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(revision)
}

func (h *HTTPHandler) commit(w http.ResponseWriter, r *http.Request) {
	expected, err := ExpectedRevision(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionRequired)
		return
	}
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxConfigPayload))
	if err != nil {
		http.Error(w, "configuration payload exceeds 16 MiB", http.StatusRequestEntityTooLarge)
		return
	}
	if !json.Valid(payload) {
		http.Error(w, "configuration payload is not valid JSON", http.StatusBadRequest)
		return
	}
	if h.validate != nil {
		if err := h.validate(payload); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}
	createdBy := "unknown"
	if h.actor != nil {
		createdBy = h.actor(r)
	}
	if createdBy == "" {
		createdBy = "unknown"
	}
	revision, err := h.store.Commit(r.Context(), expected, payload, createdBy)
	if errors.Is(err, ErrRevisionConflict) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "configuration store unavailable", http.StatusServiceUnavailable)
		return
	}
	SetRevisionHeaders(w, revision.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(revision)
}
