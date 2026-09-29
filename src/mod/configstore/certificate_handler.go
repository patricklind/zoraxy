package configstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	CertificateAPIPath    = "/api/cluster/certificates"
	maxCertificatePayload = 8 << 20
)

type certificateCommitRequest struct {
	CertificateID  string          `json:"certificate_id"`
	Metadata       json.RawMessage `json:"metadata"`
	CertificatePEM []byte          `json:"certificate_pem"`
	PrivateKeyPEM  []byte          `json:"private_key_pem"`
}

type certificateResponse struct {
	CertificateID  string          `json:"certificate_id"`
	Revision       uint64          `json:"revision"`
	Metadata       json.RawMessage `json:"metadata"`
	CertificatePEM []byte          `json:"certificate_pem"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CertificateHTTPHandler struct {
	store *PostgresStore
	box   *CertificateCipher
}

func NewCertificateHTTPHandler(store *PostgresStore, box *CertificateCipher) (*CertificateHTTPHandler, error) {
	if store == nil || box == nil {
		return nil, errors.New("certificate store and cipher are required")
	}
	return &CertificateHTTPHandler{store: store, box: box}, nil
}

func (h *CertificateHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPut:
		h.commit(w, r)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func certificatePublicResponse(revision CertificateRevision) certificateResponse {
	return certificateResponse{
		CertificateID:  revision.CertificateID,
		Revision:       revision.Revision,
		Metadata:       revision.Metadata,
		CertificatePEM: revision.CertificatePEM,
		CreatedAt:      revision.CreatedAt,
	}
}

func (h *CertificateHTTPHandler) get(w http.ResponseWriter, r *http.Request) {
	certificateID := strings.TrimSpace(r.URL.Query().Get("id"))
	if certificateID == "" {
		http.Error(w, "certificate id is required", http.StatusBadRequest)
		return
	}
	revision, err := h.store.CurrentCertificate(r.Context(), certificateID, h.box)
	if errors.Is(err, ErrNoCertificateRevision) {
		http.Error(w, "certificate does not exist", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "certificate store unavailable", http.StatusServiceUnavailable)
		return
	}
	SetRevisionHeaders(w, revision.Revision)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(certificatePublicResponse(revision))
}

func (h *CertificateHTTPHandler) commit(w http.ResponseWriter, r *http.Request) {
	expected, err := ExpectedRevision(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionRequired)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxCertificatePayload))
	if err != nil {
		http.Error(w, "certificate payload exceeds 8 MiB", http.StatusRequestEntityTooLarge)
		return
	}
	var request certificateCommitRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid certificate payload", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid certificate payload", http.StatusBadRequest)
		return
	}
	revision, err := h.store.CommitCertificate(r.Context(), expected, CertificateRevision{
		CertificateID:  strings.TrimSpace(request.CertificateID),
		Metadata:       request.Metadata,
		CertificatePEM: request.CertificatePEM,
		PrivateKeyPEM:  request.PrivateKeyPEM,
	}, h.box)
	if errors.Is(err, ErrRevisionConflict) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	SetRevisionHeaders(w, revision.Revision)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(certificatePublicResponse(revision))
}
