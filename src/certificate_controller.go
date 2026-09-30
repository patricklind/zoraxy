package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"imuslab.com/zoraxy/mod/configstore"
)

func writeControllerSecret(path string, contents []byte, mode os.FileMode) error {
	if err := os.WriteFile(path, contents, mode); err != nil {
		return err
	}
	return os.Chmod(path, mode)
}

func (p *preparedCluster) runCertificateControllerWorker(ctx context.Context) error {
	store, ok := p.store.(*configstore.PostgresStore)
	if !ok {
		return errors.New("certificate controller requires PostgreSQL configstore")
	}
	startupDelay := time.NewTimer(2 * time.Second)
	defer startupDelay.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-startupDelay.C:
	}

	interval := time.Duration(acmeAutoRenewer.RenewTickInterval) * time.Second
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := p.renewAndPublishCertificates(ctx, store); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (p *preparedCluster) renewAndPublishCertificates(ctx context.Context, store *configstore.PostgresStore) error {
	revisions, err := store.ListCurrentCertificates(ctx, p.certificateCipher)
	if err != nil {
		return fmt.Errorf("load certificate revisions for renewal: %w", err)
	}
	selected := make([]string, 0, len(revisions))
	byName := make(map[string]configstore.CertificateRevision, len(revisions))
	for _, revision := range revisions {
		metadata, err := configstore.ParseCertificateMetadata(revision.Metadata)
		if err != nil {
			return fmt.Errorf("parse certificate %s metadata: %w", revision.CertificateID, err)
		}
		if !metadata.AutoRenew {
			continue
		}
		if err := writeControllerSecret(filepath.Join(CONF_CERT_STORE, metadata.Name+".pem"), revision.CertificatePEM, 0o644); err != nil {
			return fmt.Errorf("stage certificate %s: %w", metadata.Name, err)
		}
		if err := writeControllerSecret(filepath.Join(CONF_CERT_STORE, metadata.Name+".key"), revision.PrivateKeyPEM, 0o600); err != nil {
			return fmt.Errorf("stage certificate key %s: %w", metadata.Name, err)
		}
		if err := writeControllerSecret(filepath.Join(CONF_CERT_STORE, metadata.Name+".json"), metadata.ACME, 0o600); err != nil {
			return fmt.Errorf("stage ACME metadata %s: %w", metadata.Name, err)
		}
		selected = append(selected, metadata.Name)
		byName[metadata.Name] = revision
	}
	if len(selected) == 0 {
		return nil
	}
	acmeAutoRenewer.RenewerConfig.RenewAll = false
	acmeAutoRenewer.RenewerConfig.FilesToRenew = append([]string(nil), selected...)
	renewed, err := acmeAutoRenewer.CheckAndRenewCertificatesContext(ctx)
	if err != nil {
		return err
	}
	for _, renewedFile := range renewed {
		name := strings.TrimSuffix(filepath.Base(renewedFile), filepath.Ext(renewedFile))
		previous, exists := byName[name]
		if !exists {
			return fmt.Errorf("renewed certificate %q was not in the selected revision set", name)
		}
		certificatePEM, err := os.ReadFile(filepath.Join(CONF_CERT_STORE, name+".pem"))
		if err != nil {
			return err
		}
		privateKeyPEM, err := os.ReadFile(filepath.Join(CONF_CERT_STORE, name+".key"))
		if err != nil {
			return err
		}
		_, err = store.CommitCertificate(ctx, previous.Revision, configstore.CertificateRevision{
			CertificateID:  previous.CertificateID,
			Metadata:       previous.Metadata,
			CertificatePEM: certificatePEM,
			PrivateKeyPEM:  privateKeyPEM,
		}, p.certificateCipher)
		if err != nil {
			return fmt.Errorf("publish renewed certificate %s: %w", name, err)
		}
	}
	return nil
}
