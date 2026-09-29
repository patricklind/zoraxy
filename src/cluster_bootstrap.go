package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"imuslab.com/zoraxy/mod/auth"
	"imuslab.com/zoraxy/mod/configstore"
	"imuslab.com/zoraxy/mod/tlscert"
)

const (
	configBackendLocal      = "local"
	configBackendPostgreSQL = "postgresql"
	clusterModeDisabled     = "disabled"
	clusterModeControlPlane = "control-plane"
	clusterModeDataPlane    = "data-plane"
)

type clusterConfig struct {
	backend        string
	mode           string
	dsn            string
	migrationMode  string
	pollInterval   time.Duration
	startupTimeout time.Duration
	certificateKey []byte
}

type environmentLookup func(string) string
type secretReader func(string) ([]byte, error)

func loadClusterConfig(getenv environmentLookup, readFile secretReader) (clusterConfig, error) {
	config := clusterConfig{
		backend:        strings.ToLower(strings.TrimSpace(getenv("ZORAXY_CONFIG_BACKEND"))),
		mode:           strings.ToLower(strings.TrimSpace(getenv("ZORAXY_CONFIGSTORE_MODE"))),
		migrationMode:  strings.ToLower(strings.TrimSpace(getenv("ZORAXY_CONFIGSTORE_MIGRATION_MODE"))),
		pollInterval:   time.Second,
		startupTimeout: 10 * time.Second,
	}
	if config.backend == "" {
		if config.mode == "" || config.mode == clusterModeDisabled {
			config.backend = configBackendLocal
		} else {
			// Preserve compatibility with the first configstore preview, where an
			// enabled role implied PostgreSQL.
			config.backend = configBackendPostgreSQL
		}
	}
	if config.backend != configBackendLocal && config.backend != configBackendPostgreSQL {
		return clusterConfig{}, fmt.Errorf("invalid ZORAXY_CONFIG_BACKEND %q", config.backend)
	}
	if config.backend == configBackendLocal {
		if config.mode != "" && config.mode != clusterModeDisabled {
			return clusterConfig{}, errors.New("local configuration backend cannot use a configstore cluster role")
		}
		config.mode = clusterModeDisabled
		return config, nil
	}
	if config.mode == "" || config.mode == clusterModeDisabled {
		return clusterConfig{}, errors.New("PostgreSQL configuration backend requires control-plane or data-plane mode")
	}
	if config.mode != clusterModeControlPlane && config.mode != clusterModeDataPlane {
		return clusterConfig{}, fmt.Errorf("invalid ZORAXY_CONFIGSTORE_MODE %q", config.mode)
	}
	if config.migrationMode != "verify" && config.migrationMode != "apply" {
		return clusterConfig{}, errors.New("ZORAXY_CONFIGSTORE_MIGRATION_MODE must be verify or apply")
	}

	directDSN := strings.TrimSpace(getenv("ZORAXY_CONFIGSTORE_DSN"))
	dsnFile := strings.TrimSpace(getenv("ZORAXY_CONFIGSTORE_DSN_FILE"))
	if directDSN != "" && dsnFile != "" {
		return clusterConfig{}, errors.New("set only one of ZORAXY_CONFIGSTORE_DSN and ZORAXY_CONFIGSTORE_DSN_FILE")
	}
	if dsnFile != "" {
		contents, err := readFile(dsnFile)
		if err != nil {
			return clusterConfig{}, fmt.Errorf("read configstore DSN file: %w", err)
		}
		config.dsn = strings.TrimSpace(string(contents))
	} else {
		config.dsn = directDSN
	}
	if config.dsn == "" {
		return clusterConfig{}, errors.New("PostgreSQL DSN is required when configstore is enabled")
	}
	certificateKeyFile := strings.TrimSpace(getenv("ZORAXY_CONFIGSTORE_CERTIFICATE_KEY_FILE"))
	if certificateKeyFile == "" {
		return clusterConfig{}, errors.New("ZORAXY_CONFIGSTORE_CERTIFICATE_KEY_FILE is required when configstore is enabled")
	}
	certificateKey, err := readFile(certificateKeyFile)
	if err != nil {
		return clusterConfig{}, fmt.Errorf("read configstore certificate key file: %w", err)
	}
	certificateKey = []byte(strings.TrimSpace(string(certificateKey)))
	if decoded, decodeErr := base64.StdEncoding.DecodeString(string(certificateKey)); decodeErr == nil && len(decoded) == 32 {
		certificateKey = decoded
	}
	if len(certificateKey) != 32 {
		return clusterConfig{}, errors.New("configstore certificate key must be 32 raw bytes or base64-encoded 32 bytes")
	}
	config.certificateKey = certificateKey

	if value := strings.TrimSpace(getenv("ZORAXY_CONFIGSTORE_POLL_INTERVAL")); value != "" {
		config.pollInterval, err = time.ParseDuration(value)
		if err != nil || config.pollInterval <= 0 {
			return clusterConfig{}, errors.New("ZORAXY_CONFIGSTORE_POLL_INTERVAL must be a positive duration")
		}
	}
	if value := strings.TrimSpace(getenv("ZORAXY_CONFIGSTORE_STARTUP_TIMEOUT")); value != "" {
		config.startupTimeout, err = time.ParseDuration(value)
		if err != nil || config.startupTimeout <= 0 {
			return clusterConfig{}, errors.New("ZORAXY_CONFIGSTORE_STARTUP_TIMEOUT must be a positive duration")
		}
	}
	return config, nil
}

func validateRoutingRevision(payload json.RawMessage) error {
	candidate, err := (zoraxyRoutingBuilder{router: dynamicProxyRouter}).Build(context.Background(), configstore.Revision{Payload: payload})
	if err != nil {
		return err
	}
	return candidate.Close()
}

type preparedCluster struct {
	config            clusterConfig
	db                *sql.DB
	store             configstore.Store
	controlPlane      *configstore.ControlPlane
	current           configstore.Revision
	certificateCipher *configstore.CertificateCipher
	certificates      []configstore.CertificateRevision
}

var configStoreDataPlaneMode bool

// prepareCluster completes every external preflight before a proxy listener is
// allowed to start. Runtime activation waits until the router exists.
func prepareCluster() (*preparedCluster, error) {
	config, err := loadClusterConfig(os.Getenv, os.ReadFile)
	if err != nil {
		return nil, err
	}
	if config.mode == clusterModeDisabled {
		return &preparedCluster{config: config}, nil
	}
	if config.mode == clusterModeControlPlane && !requireAuth {
		return nil, errors.New("control-plane mode requires management authentication")
	}

	db, err := sql.Open("pgx", config.dsn)
	if err != nil {
		return nil, fmt.Errorf("open configstore postgres: %w", err)
	}
	startupContext, cancel := context.WithTimeout(context.Background(), config.startupTimeout)
	defer cancel()
	if config.migrationMode == "apply" {
		err = configstore.ApplySchema(startupContext, db)
	} else {
		err = configstore.VerifySchema(startupContext, db)
	}
	if err != nil {
		db.Close()
		return nil, err
	}

	store := configstore.NewPostgresStore(db, config.pollInterval)
	certificateCipher, err := configstore.NewCertificateCipher(config.certificateKey)
	if err != nil {
		db.Close()
		return nil, err
	}
	controlPlane, err := configstore.NewControlPlane(store, store, validateRoutingRevision, func(r *http.Request) string {
		username, lookupErr := authAgent.GetUserName(nil, r)
		if lookupErr != nil {
			return "unknown"
		}
		return username
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	prepared := &preparedCluster{config: config, db: db, store: store, controlPlane: controlPlane, certificateCipher: certificateCipher}
	if config.mode == clusterModeDataPlane {
		configStoreDataPlaneMode = true
		prepared.current, err = store.Current(startupContext)
		if err != nil {
			db.Close()
			if errors.Is(err, configstore.ErrNoRevision) {
				return nil, errors.New("data-plane mode requires an existing configuration revision")
			}
			return nil, fmt.Errorf("load initial configuration revision: %w", err)
		}
		prepared.certificates, err = store.ListCurrentCertificates(startupContext, certificateCipher)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("load initial certificate revisions: %w", err)
		}
	}
	return prepared, nil
}

func (p *preparedCluster) Start(authRouter *auth.RouterDef) error {
	if p == nil {
		return errors.New("cluster preflight is required")
	}
	if p.config.mode == clusterModeDisabled {
		return nil
	}
	if authRouter == nil || dynamicProxyRouter == nil {
		return errors.New("management and proxy routers must be initialized before configstore startup")
	}
	if p.config.mode == clusterModeControlPlane {
		if err := p.controlPlane.RegisterManagementAPI(authRouter); err != nil {
			p.db.Close()
			return err
		}
		postgresStore, ok := p.store.(*configstore.PostgresStore)
		if !ok {
			p.db.Close()
			return errors.New("certificate API requires PostgreSQL configstore")
		}
		certificateHandler, err := configstore.NewCertificateHTTPHandler(postgresStore, p.certificateCipher)
		if err != nil {
			p.db.Close()
			return err
		}
		if err := authRouter.HandleFunc(configstore.CertificateAPIPath, certificateHandler.ServeHTTP); err != nil {
			p.db.Close()
			return err
		}
		migrationHandler := routingMigrationHandler{router: dynamicProxyRouter, store: p.store}
		if err := authRouter.HandleFunc(routingExportAPIPath, migrationHandler.export); err != nil {
			p.db.Close()
			return err
		}
		if err := authRouter.HandleFunc(routingShadowAPIPath, migrationHandler.shadow); err != nil {
			p.db.Close()
			return err
		}
		configStoreReady.Store(true)
		return nil
	}

	activator, err := newZoraxyRoutingActivator(dynamicProxyRouter)
	if err != nil {
		p.db.Close()
		return err
	}
	startupContext, cancel := context.WithTimeout(context.Background(), p.config.startupTimeout)
	defer cancel()
	if err := activateCertificateRevisions(tlsCertManager, p.certificates); err != nil {
		p.db.Close()
		return fmt.Errorf("activate initial certificate revisions: %w", err)
	}
	desiredConfigRevision.Store(p.current.ID)
	if err := activator.Activate(startupContext, p.current); err != nil {
		p.db.Close()
		return fmt.Errorf("activate initial configuration revision %d: %w", p.current.ID, err)
	}
	appliedConfigRevision.Store(p.current.ID)
	configStoreReady.Store(true)
	dynamicProxyRouterConfigured <- struct{}{}

	node := configstore.NodeStatus{
		NodeID:          nodeUUID,
		NodeRole:        clusterModeDataPlane,
		ConfigRevision:  p.current.ID,
		AppliedRevision: p.current.ID,
	}
	observedActivator := configstore.ActivatorFunc(func(ctx context.Context, revision configstore.Revision) error {
		desiredConfigRevision.Store(revision.ID)
		if err := activator.Activate(ctx, revision); err != nil {
			return err
		}
		appliedConfigRevision.Store(revision.ID)
		return nil
	})
	go func() {
		if err := p.controlPlane.RunDataNode(context.Background(), observedActivator, node); err != nil {
			configStoreReady.Store(false)
			SystemWideLogger.PrintAndLog("configstore", "Configuration follower stopped", err)
		}
	}()
	go p.runCertificateFollower(context.Background())
	return nil
}

func activateCertificateRevisions(manager *tlscert.Manager, revisions []configstore.CertificateRevision) error {
	candidates := make([]tlscert.MemoryCertificate, 0, len(revisions))
	for _, revision := range revisions {
		metadata, err := configstore.ParseCertificateMetadata(revision.Metadata)
		if err != nil {
			return fmt.Errorf("decode certificate %s metadata: %w", revision.CertificateID, err)
		}
		candidates = append(candidates, tlscert.MemoryCertificate{
			Name:           metadata.Name,
			CertificatePEM: revision.CertificatePEM,
			PrivateKeyPEM:  revision.PrivateKeyPEM,
			Fallback:       metadata.Fallback,
		})
	}
	return manager.ReplaceMemoryCertificates(candidates)
}

func certificateRevisionSignature(revisions []configstore.CertificateRevision) string {
	var signature strings.Builder
	for _, revision := range revisions {
		fmt.Fprintf(&signature, "%s:%d;", revision.CertificateID, revision.Revision)
	}
	return signature.String()
}

func (p *preparedCluster) runCertificateFollower(ctx context.Context) {
	store, ok := p.store.(*configstore.PostgresStore)
	if !ok {
		return
	}
	activeSignature := certificateRevisionSignature(p.certificates)
	ticker := time.NewTicker(p.config.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			revisions, err := store.ListCurrentCertificates(ctx, p.certificateCipher)
			if err != nil {
				SystemWideLogger.PrintAndLog("configstore", "Certificate follower failed to read revisions", err)
				continue
			}
			signature := certificateRevisionSignature(revisions)
			if signature == activeSignature {
				continue
			}
			if err := activateCertificateRevisions(tlsCertManager, revisions); err != nil {
				SystemWideLogger.PrintAndLog("configstore", "Certificate revision activation failed", err)
				continue
			}
			activeSignature = signature
		}
	}
}
