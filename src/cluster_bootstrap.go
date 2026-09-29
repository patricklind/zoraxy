package main

import (
	"context"
	"database/sql"
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

	var err error
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
	config       clusterConfig
	db           *sql.DB
	store        configstore.Store
	controlPlane *configstore.ControlPlane
	current      configstore.Revision
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
	prepared := &preparedCluster{config: config, db: db, store: store, controlPlane: controlPlane}
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
	return nil
}
