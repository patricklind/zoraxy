package configstore

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var ErrNoCertificateRevision = errors.New("no certificate revision exists")

type CertificateRevision struct {
	CertificateID  string          `json:"certificate_id"`
	Revision       uint64          `json:"revision"`
	Metadata       json.RawMessage `json:"metadata"`
	PrivateKeyPEM  []byte          `json:"-"`
	CertificatePEM []byte          `json:"certificate_pem"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CertificateMetadata struct {
	Name     string `json:"name"`
	Fallback bool   `json:"fallback"`
}

func ParseCertificateMetadata(payload json.RawMessage) (CertificateMetadata, error) {
	var metadata CertificateMetadata
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		return CertificateMetadata{}, errors.New("certificate metadata is not valid")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return CertificateMetadata{}, errors.New("certificate metadata is not valid")
	}
	metadata.Name = strings.TrimSpace(metadata.Name)
	if metadata.Name == "" {
		return CertificateMetadata{}, errors.New("certificate metadata name is required")
	}
	return metadata, nil
}

type CertificateCipher struct {
	aead cipher.AEAD
}

func NewCertificateCipher(key []byte) (*CertificateCipher, error) {
	if len(key) != 32 {
		return nil, errors.New("certificate encryption key must contain exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &CertificateCipher{aead: aead}, nil
}

func certificateAAD(certificateID string, revision uint64) []byte {
	aad := make([]byte, len(certificateID)+1+8)
	copy(aad, certificateID)
	aad[len(certificateID)] = 0
	binary.BigEndian.PutUint64(aad[len(certificateID)+1:], revision)
	return aad
}

func (c *CertificateCipher) seal(certificateID string, revision uint64, plaintext []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate certificate key nonce: %w", err)
	}
	return c.aead.Seal(nonce, nonce, plaintext, certificateAAD(certificateID, revision)), nil
}

func (c *CertificateCipher) open(certificateID string, revision uint64, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < c.aead.NonceSize() {
		return nil, errors.New("encrypted certificate key is truncated")
	}
	nonce := ciphertext[:c.aead.NonceSize()]
	plaintext, err := c.aead.Open(nil, nonce, ciphertext[c.aead.NonceSize():], certificateAAD(certificateID, revision))
	if err != nil {
		return nil, errors.New("decrypt certificate key")
	}
	return plaintext, nil
}

func scanCertificateRevision(row interface{ Scan(...any) error }, box *CertificateCipher) (CertificateRevision, error) {
	var revision CertificateRevision
	var metadata, encryptedKey []byte
	err := row.Scan(&revision.CertificateID, &revision.Revision, &metadata, &encryptedKey, &revision.CertificatePEM, &revision.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return CertificateRevision{}, ErrNoCertificateRevision
	}
	if err != nil {
		return CertificateRevision{}, err
	}
	revision.Metadata = json.RawMessage(metadata)
	revision.PrivateKeyPEM, err = box.open(revision.CertificateID, revision.Revision, encryptedKey)
	if err != nil {
		return CertificateRevision{}, err
	}
	if _, err := tls.X509KeyPair(revision.CertificatePEM, revision.PrivateKeyPEM); err != nil {
		return CertificateRevision{}, fmt.Errorf("stored certificate revision is invalid: %w", err)
	}
	return revision, nil
}

func (s *PostgresStore) CurrentCertificate(ctx context.Context, certificateID string, box *CertificateCipher) (CertificateRevision, error) {
	if certificateID == "" || box == nil {
		return CertificateRevision{}, errors.New("certificate id and cipher are required")
	}
	return scanCertificateRevision(s.db.QueryRowContext(ctx, `
		SELECT certificate_id::text, revision, metadata, encrypted_key, certificate_pem, created_at
		FROM certificate_revisions WHERE certificate_id = $1
		ORDER BY revision DESC LIMIT 1`, certificateID), box)
}

// ListCurrentCertificates returns one fully validated, decrypted revision per
// certificate. Callers must keep the private key in memory and must never log
// or return it through an API.
func (s *PostgresStore) ListCurrentCertificates(ctx context.Context, box *CertificateCipher) ([]CertificateRevision, error) {
	if box == nil {
		return nil, errors.New("certificate cipher is required")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT ON (certificate_id)
			certificate_id::text, revision, metadata, encrypted_key, certificate_pem, created_at
		FROM certificate_revisions
		ORDER BY certificate_id, revision DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revisions := make([]CertificateRevision, 0)
	for rows.Next() {
		revision, err := scanCertificateRevision(rows, box)
		if err != nil {
			return nil, err
		}
		revisions = append(revisions, revision)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return revisions, nil
}

func (s *PostgresStore) CommitCertificate(ctx context.Context, expectedRevision uint64, certificate CertificateRevision, box *CertificateCipher) (CertificateRevision, error) {
	if certificate.CertificateID == "" || box == nil {
		return CertificateRevision{}, errors.New("certificate id and cipher are required")
	}
	metadata, err := ParseCertificateMetadata(certificate.Metadata)
	if err != nil {
		return CertificateRevision{}, err
	}
	if _, err := tls.X509KeyPair(certificate.CertificatePEM, certificate.PrivateKeyPEM); err != nil {
		return CertificateRevision{}, fmt.Errorf("certificate and private key do not match: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return CertificateRevision{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, $2))`, certificate.CertificateID, revisionLockID); err != nil {
		return CertificateRevision{}, err
	}
	var current uint64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(revision), 0) FROM certificate_revisions WHERE certificate_id = $1`, certificate.CertificateID).Scan(&current); err != nil {
		return CertificateRevision{}, err
	}
	if current != expectedRevision {
		return CertificateRevision{}, ErrRevisionConflict
	}
	if metadata.Fallback {
		var anotherFallback bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM (
					SELECT DISTINCT ON (certificate_id) certificate_id, metadata
					FROM certificate_revisions
					WHERE certificate_id <> $1
					ORDER BY certificate_id, revision DESC
				) latest
				WHERE COALESCE((metadata->>'fallback')::boolean, false)
			)`, certificate.CertificateID).Scan(&anotherFallback); err != nil {
			return CertificateRevision{}, err
		}
		if anotherFallback {
			return CertificateRevision{}, errors.New("another certificate is already configured as fallback")
		}
	}
	next := current + 1
	encryptedKey, err := box.seal(certificate.CertificateID, next, certificate.PrivateKeyPEM)
	if err != nil {
		return CertificateRevision{}, err
	}
	result, err := scanCertificateRevision(tx.QueryRowContext(ctx, `
		INSERT INTO certificate_revisions (certificate_id, revision, metadata, encrypted_key, certificate_pem)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING certificate_id::text, revision, metadata, encrypted_key, certificate_pem, created_at`,
		certificate.CertificateID, next, []byte(certificate.Metadata), encryptedKey, certificate.CertificatePEM), box)
	if err != nil {
		return CertificateRevision{}, err
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_notify('zoraxy_certificate_revision', $1)`, certificate.CertificateID); err != nil {
		return CertificateRevision{}, err
	}
	if err := tx.Commit(); err != nil {
		return CertificateRevision{}, err
	}
	return result, nil
}

type ControllerLease struct {
	HolderID   string    `json:"holder_id"`
	ValidUntil time.Time `json:"valid_until"`
}

func (s *PostgresStore) AcquireCertificateControllerLease(ctx context.Context, holderID string, ttl time.Duration) (ControllerLease, bool, error) {
	if holderID == "" {
		return ControllerLease{}, false, errors.New("controller holder id is required")
	}
	if ttl < 5*time.Second || ttl > 5*time.Minute {
		return ControllerLease{}, false, errors.New("controller lease ttl must be between 5 seconds and 5 minutes")
	}
	var lease ControllerLease
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO controller_leases (lease_name, holder_id, valid_until)
		VALUES ('certificate-controller', $1, clock_timestamp() + make_interval(secs => $2))
		ON CONFLICT (lease_name) DO UPDATE SET
			holder_id = EXCLUDED.holder_id,
			valid_until = EXCLUDED.valid_until
		WHERE controller_leases.valid_until <= clock_timestamp()
		   OR controller_leases.holder_id = EXCLUDED.holder_id
		RETURNING holder_id::text, valid_until`, holderID, int64(ttl/time.Second)).Scan(&lease.HolderID, &lease.ValidUntil)
	if errors.Is(err, sql.ErrNoRows) {
		return ControllerLease{}, false, nil
	}
	if err != nil {
		return ControllerLease{}, false, err
	}
	return lease, true, nil
}
