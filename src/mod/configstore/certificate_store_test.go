package configstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func testCertificatePair(t *testing.T) ([]byte, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.test"},
		DNSNames:     []string{"example.test"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
}

func TestCertificateCipherBindsIdentityAndRevision(t *testing.T) {
	box, err := NewCertificateCipher(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := box.seal("11111111-1111-1111-1111-111111111111", 1, []byte("private-key"))
	if err != nil {
		t.Fatal(err)
	}
	opened, err := box.open("11111111-1111-1111-1111-111111111111", 1, sealed)
	if err != nil || string(opened) != "private-key" {
		t.Fatalf("open = %q, %v", opened, err)
	}
	if _, err := box.open("22222222-2222-2222-2222-222222222222", 1, sealed); err == nil {
		t.Fatal("ciphertext accepted for another certificate")
	}
	if _, err := box.open("11111111-1111-1111-1111-111111111111", 2, sealed); err == nil {
		t.Fatal("ciphertext accepted for another revision")
	}
}

func TestCertificateCipherRequiresAES256Key(t *testing.T) {
	if _, err := NewCertificateCipher(make([]byte, 16)); err == nil {
		t.Fatal("short encryption key accepted")
	}
}

func TestCommitCertificateRejectsMismatchedKeyBeforeDatabase(t *testing.T) {
	certificatePEM, _ := testCertificatePair(t)
	_, anotherPrivateKey := testCertificatePair(t)
	box, err := NewCertificateCipher(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	store := &PostgresStore{}
	_, err = store.CommitCertificate(t.Context(), 0, CertificateRevision{
		CertificateID:  "11111111-1111-1111-1111-111111111111",
		Metadata:       json.RawMessage(`{}`),
		CertificatePEM: certificatePEM,
		PrivateKeyPEM:  anotherPrivateKey,
	}, box)
	if err == nil {
		t.Fatal("mismatched private key accepted")
	}
}

func TestParseCertificateMetadataRejectsUnknownOrMissingName(t *testing.T) {
	for _, payload := range []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"name":"example.test","unexpected":true}`),
		json.RawMessage(`{"name":"example.test"}{}`),
	} {
		if _, err := ParseCertificateMetadata(payload); err == nil {
			t.Fatalf("metadata %s was accepted", payload)
		}
	}
	metadata, err := ParseCertificateMetadata(json.RawMessage(`{"name":" example.test ","fallback":true}`))
	if err != nil || metadata.Name != "example.test" || !metadata.Fallback {
		t.Fatalf("metadata = %+v, %v", metadata, err)
	}
}
