package configstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresStoreIntegration(t *testing.T) {
	dsn := os.Getenv("CONFIGSTORE_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("CONFIGSTORE_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := VerifySchema(ctx, db); err != nil {
		t.Fatalf("verify initialized schema: %v", err)
	}
	if err := ApplySchema(ctx, db); err != nil {
		t.Fatalf("idempotent schema migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE config_revisions, node_status`); err != nil {
		t.Fatal(err)
	}

	store := NewPostgresStore(db, 10*time.Millisecond)
	revision, err := store.Commit(ctx, 0, json.RawMessage(`{"routes":[]}`), "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	if revision.ID != 1 || len(revision.SHA256) != 64 {
		t.Fatalf("revision = %+v", revision)
	}
	if _, err := store.Commit(ctx, 0, json.RawMessage(`{"routes":[]}`), "stale-writer"); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale commit error = %v, want %v", err, ErrRevisionConflict)
	}

	watched, failures := store.Watch(ctx, 0)
	select {
	case got := <-watched:
		if got.ID != revision.ID {
			t.Fatalf("watched revision = %d, want %d", got.ID, revision.ID)
		}
	case err := <-failures:
		t.Fatalf("watch failed: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	status := NodeStatus{
		NodeID:          "8f196376-f209-4e21-ae03-fec14d77d0d7",
		NodeRole:        "data-plane",
		ConfigRevision:  1,
		AppliedRevision: 1,
	}
	if err := store.UpsertNodeStatus(ctx, status); err != nil {
		t.Fatal(err)
	}
	statuses, err := store.ListNodeStatuses(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].NodeID != status.NodeID || statuses[0].AppliedRevision != 1 {
		t.Fatalf("statuses = %+v", statuses)
	}

	if _, err := db.ExecContext(ctx, `TRUNCATE certificate_revisions, controller_leases`); err != nil {
		t.Fatal(err)
	}
	box, err := NewCertificateCipher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	certificatePEM, privateKeyPEM := testCertificatePair(t)
	certificateID := "38a4e4d3-4ef6-4b36-ac2a-d784d18d6977"
	certificate, err := store.CommitCertificate(ctx, 0, CertificateRevision{
		CertificateID:  certificateID,
		Metadata:       json.RawMessage(`{"name":"example.test"}`),
		CertificatePEM: certificatePEM,
		PrivateKeyPEM:  privateKeyPEM,
	}, box)
	if err != nil {
		t.Fatal(err)
	}
	if certificate.Revision != 1 || string(certificate.PrivateKeyPEM) != string(privateKeyPEM) {
		t.Fatalf("certificate revision = %+v", certificate)
	}
	var encryptedKey []byte
	if err := db.QueryRowContext(ctx, `SELECT encrypted_key FROM certificate_revisions WHERE certificate_id = $1 AND revision = 1`, certificateID).Scan(&encryptedKey); err != nil {
		t.Fatal(err)
	}
	if string(encryptedKey) == string(privateKeyPEM) {
		t.Fatal("private key was stored without encryption")
	}
	current, err := store.CurrentCertificate(ctx, certificateID, box)
	if err != nil || current.Revision != 1 {
		t.Fatalf("current certificate = %+v, %v", current, err)
	}
	certificates, err := store.ListCurrentCertificates(ctx, box)
	if err != nil || len(certificates) != 1 || certificates[0].CertificateID != certificateID {
		t.Fatalf("current certificates = %+v, %v", certificates, err)
	}
	if _, err := store.CommitCertificate(ctx, 0, certificate, box); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale certificate commit error = %v, want %v", err, ErrRevisionConflict)
	}
	certificateHandler, err := NewCertificateHTTPHandler(store, box)
	if err != nil {
		t.Fatal(err)
	}
	getRequest := httptest.NewRequest(http.MethodGet, CertificateAPIPath+"?id="+certificateID, nil)
	getResponse := httptest.NewRecorder()
	certificateHandler.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK || getResponse.Header().Get("ETag") != `"1"` || bytes.Contains(getResponse.Body.Bytes(), privateKeyPEM) {
		t.Fatalf("certificate GET status=%d etag=%q leaked-key=%v body=%s", getResponse.Code, getResponse.Header().Get("ETag"), bytes.Contains(getResponse.Body.Bytes(), privateKeyPEM), getResponse.Body.String())
	}
	putPayload, err := json.Marshal(certificateCommitRequest{
		CertificateID:  certificateID,
		Metadata:       json.RawMessage(`{"name":"example.test","fallback":true}`),
		CertificatePEM: certificatePEM,
		PrivateKeyPEM:  privateKeyPEM,
	})
	if err != nil {
		t.Fatal(err)
	}
	putRequest := httptest.NewRequest(http.MethodPut, CertificateAPIPath, bytes.NewReader(putPayload))
	putRequest.Header.Set("If-Match", `"1"`)
	putResponse := httptest.NewRecorder()
	certificateHandler.ServeHTTP(putResponse, putRequest)
	if putResponse.Code != http.StatusCreated || putResponse.Header().Get("ETag") != `"2"` || bytes.Contains(putResponse.Body.Bytes(), privateKeyPEM) {
		t.Fatalf("certificate PUT status=%d etag=%q leaked-key=%v body=%s", putResponse.Code, putResponse.Header().Get("ETag"), bytes.Contains(putResponse.Body.Bytes(), privateKeyPEM), putResponse.Body.String())
	}
	if _, err := store.CommitCertificate(ctx, 0, CertificateRevision{
		CertificateID:  "b7a92388-4adf-48f4-b5a4-509d596808fd",
		Metadata:       json.RawMessage(`{"name":"second.example.test","fallback":true}`),
		CertificatePEM: certificatePEM,
		PrivateKeyPEM:  privateKeyPEM,
	}, box); err == nil {
		t.Fatal("second fallback certificate was accepted")
	}
	if _, err := store.CommitCertificate(ctx, 0, CertificateRevision{
		CertificateID:  "5adb79f2-b040-4939-a01a-06c131d1d173",
		Metadata:       json.RawMessage(`{"name":"example.test"}`),
		CertificatePEM: certificatePEM,
		PrivateKeyPEM:  privateKeyPEM,
	}, box); err == nil {
		t.Fatal("duplicate certificate metadata name was accepted")
	}

	holderA := "f2e31343-f734-42f3-b5e8-72975ce354a7"
	holderB := "d61524c2-2265-4245-ae91-668cfdfda695"
	lease, acquired, err := store.AcquireCertificateControllerLease(ctx, holderA, 10*time.Second)
	if err != nil || !acquired || lease.HolderID != holderA {
		t.Fatalf("first lease = %+v, %v, %v", lease, acquired, err)
	}
	if _, acquired, err := store.AcquireCertificateControllerLease(ctx, holderB, 10*time.Second); err != nil || acquired {
		t.Fatalf("competing lease acquired = %v, %v", acquired, err)
	}
	if lease, acquired, err := store.AcquireCertificateControllerLease(ctx, holderA, 10*time.Second); err != nil || !acquired || lease.HolderID != holderA {
		t.Fatalf("lease renewal = %+v, %v, %v", lease, acquired, err)
	}
}
