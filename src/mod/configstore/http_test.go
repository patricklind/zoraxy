package configstore

import (
	"errors"
	"net/http/httptest"
	"testing"
)

func TestExpectedRevision(t *testing.T) {
	for _, value := range []string{"42", `"42"`} {
		req := httptest.NewRequest("PUT", "/api/cluster/config", nil)
		req.Header.Set("If-Match", value)
		revision, err := ExpectedRevision(req)
		if err != nil || revision != 42 {
			t.Fatalf("If-Match %q returned %d, %v", value, revision, err)
		}
	}
}

func TestExpectedRevisionRejectsMissingAndWildcard(t *testing.T) {
	for _, value := range []string{"", "*", "not-a-number"} {
		req := httptest.NewRequest("PUT", "/api/cluster/config", nil)
		req.Header.Set("If-Match", value)
		_, err := ExpectedRevision(req)
		if err == nil {
			t.Fatalf("If-Match %q was accepted", value)
		}
		if value == "" && !errors.Is(err, ErrIfMatchRequired) {
			t.Fatalf("missing header error = %v", err)
		}
	}
}
