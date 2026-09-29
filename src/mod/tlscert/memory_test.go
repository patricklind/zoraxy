package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func memoryTestCertificate(t *testing.T, name string) ([]byte, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: name},
		DNSNames:     []string{name},
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

func TestReplaceMemoryCertificatesSwitchesAtomically(t *testing.T) {
	oldCertificate, oldKey := memoryTestCertificate(t, "old.example.test")
	newCertificate, newKey := memoryTestCertificate(t, "new.example.test")
	manager := &Manager{hostSpecificTlsBehavior: defaultHostSpecificTlsBehavior}
	if err := manager.ReplaceMemoryCertificates([]MemoryCertificate{{Name: "old", CertificatePEM: oldCertificate, PrivateKeyPEM: oldKey}}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.GetCert(&tls.ClientHelloInfo{ServerName: "old.example.test"}); err != nil {
		t.Fatal(err)
	}
	if err := manager.ReplaceMemoryCertificates([]MemoryCertificate{{Name: "new", CertificatePEM: newCertificate, PrivateKeyPEM: newKey}}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.GetCert(&tls.ClientHelloInfo{ServerName: "old.example.test"}); err == nil {
		t.Fatal("old certificate remained active")
	}
	if _, err := manager.GetCert(&tls.ClientHelloInfo{ServerName: "new.example.test"}); err != nil {
		t.Fatal(err)
	}
}

func TestReplaceMemoryCertificatesRejectsInvalidSnapshotWithoutChangingActive(t *testing.T) {
	certificate, key := memoryTestCertificate(t, "active.example.test")
	manager := &Manager{hostSpecificTlsBehavior: defaultHostSpecificTlsBehavior}
	if err := manager.ReplaceMemoryCertificates([]MemoryCertificate{{Name: "active", CertificatePEM: certificate, PrivateKeyPEM: key, Fallback: true}}); err != nil {
		t.Fatal(err)
	}
	if err := manager.ReplaceMemoryCertificates([]MemoryCertificate{{Name: "broken", CertificatePEM: certificate, PrivateKeyPEM: []byte("broken")}}); err == nil {
		t.Fatal("invalid snapshot accepted")
	}
	if _, err := manager.GetCert(&tls.ClientHelloInfo{ServerName: "unknown.example.test"}); err != nil {
		t.Fatalf("active fallback was lost: %v", err)
	}
}
