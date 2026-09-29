package configstore

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

var ErrIfMatchRequired = errors.New("If-Match revision is required")

// ExpectedRevision parses an HTTP If-Match value in either `42` or `"42"`
// form. Wildcards are rejected because config commits must name an exact base.
func ExpectedRevision(r *http.Request) (uint64, error) {
	value := strings.TrimSpace(r.Header.Get("If-Match"))
	if value == "" || value == "*" {
		return 0, ErrIfMatchRequired
	}
	value = strings.Trim(value, `"`)
	revision, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, errors.New("If-Match must contain a numeric configuration revision")
	}
	return revision, nil
}

func SetRevisionHeaders(w http.ResponseWriter, revision uint64) {
	value := strconv.FormatUint(revision, 10)
	w.Header().Set("ETag", `"`+value+`"`)
	w.Header().Set("X-Zoraxy-Config-Revision", value)
}
