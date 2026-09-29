package configstore

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type handlerStore struct {
	current Revision
	err     error
}

func (s *handlerStore) Current(context.Context) (Revision, error) { return s.current, s.err }
func (s *handlerStore) Commit(_ context.Context, expected uint64, payload json.RawMessage, actor string) (Revision, error) {
	if s.err != nil {
		return Revision{}, s.err
	}
	if expected != s.current.ID {
		return Revision{}, ErrRevisionConflict
	}
	s.current = Revision{ID: expected + 1, Payload: payload, CreatedBy: actor}
	return s.current, nil
}
func (s *handlerStore) Watch(context.Context, uint64) (<-chan Revision, <-chan error) {
	revisions := make(chan Revision)
	failures := make(chan error)
	close(revisions)
	close(failures)
	return revisions, failures
}

func TestHTTPHandlerRequiresAndChecksRevision(t *testing.T) {
	store := &handlerStore{current: Revision{ID: 4}}
	handler := NewHTTPHandler(store, nil, func(*http.Request) string { return "admin" })

	missing := httptest.NewRecorder()
	handler.ServeHTTP(missing, httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"routes":[]}`)))
	if missing.Code != http.StatusPreconditionRequired {
		t.Fatalf("missing If-Match status = %d", missing.Code)
	}

	conflictRequest := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"routes":[]}`))
	conflictRequest.Header.Set("If-Match", "3")
	conflict := httptest.NewRecorder()
	handler.ServeHTTP(conflict, conflictRequest)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status = %d", conflict.Code)
	}

	commitRequest := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"routes":[]}`))
	commitRequest.Header.Set("If-Match", `"4"`)
	committed := httptest.NewRecorder()
	handler.ServeHTTP(committed, commitRequest)
	if committed.Code != http.StatusCreated {
		t.Fatalf("commit status = %d, body = %s", committed.Code, committed.Body.String())
	}
	if committed.Header().Get("ETag") != `"5"` {
		t.Fatalf("ETag = %q, want %q", committed.Header().Get("ETag"), `"5"`)
	}
}

func TestHTTPHandlerDoesNotCommitInvalidCandidate(t *testing.T) {
	store := &handlerStore{current: Revision{ID: 2}}
	handler := NewHTTPHandler(store, func(json.RawMessage) error {
		return errors.New("candidate listener conflicts with an active port")
	}, nil)
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"routes":[]}`))
	req.Header.Set("If-Match", "2")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	if store.current.ID != 2 {
		t.Fatalf("store advanced to revision %d after validation failure", store.current.ID)
	}
}
