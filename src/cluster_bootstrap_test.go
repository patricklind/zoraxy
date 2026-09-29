package main

import (
	"errors"
	"testing"
	"time"
)

func clusterTestEnvironment(values map[string]string) environmentLookup {
	return func(key string) string { return values[key] }
}

func TestLoadClusterConfigDefaultsToDisabled(t *testing.T) {
	config, err := loadClusterConfig(clusterTestEnvironment(nil), func(string) ([]byte, error) {
		return nil, errors.New("must not read a secret")
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.mode != clusterModeDisabled {
		t.Fatalf("mode = %q", config.mode)
	}
	if config.backend != configBackendLocal {
		t.Fatalf("backend = %q", config.backend)
	}
}

func TestLoadClusterConfigReadsDSNSecret(t *testing.T) {
	config, err := loadClusterConfig(clusterTestEnvironment(map[string]string{
		"ZORAXY_CONFIG_BACKEND":              "postgresql",
		"ZORAXY_CONFIGSTORE_MODE":            "data-plane",
		"ZORAXY_CONFIGSTORE_MIGRATION_MODE":  "verify",
		"ZORAXY_CONFIGSTORE_DSN_FILE":        "/run/secrets/configstore-dsn",
		"ZORAXY_CONFIGSTORE_POLL_INTERVAL":   "250ms",
		"ZORAXY_CONFIGSTORE_STARTUP_TIMEOUT": "20s",
	}), func(path string) ([]byte, error) {
		if path != "/run/secrets/configstore-dsn" {
			t.Fatalf("secret path = %q", path)
		}
		return []byte(" postgres://cluster.example/zoraxy\n"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.dsn != "postgres://cluster.example/zoraxy" || config.pollInterval != 250*time.Millisecond || config.startupTimeout != 20*time.Second {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestLoadClusterConfigRejectsUnsafeAmbiguity(t *testing.T) {
	tests := []map[string]string{
		{
			"ZORAXY_CONFIG_BACKEND":   "local",
			"ZORAXY_CONFIGSTORE_MODE": "data-plane",
		},
		{
			"ZORAXY_CONFIG_BACKEND": "postgresql",
		},
		{
			"ZORAXY_CONFIG_BACKEND": "mysql",
		},
		{
			"ZORAXY_CONFIGSTORE_MODE":           "combined",
			"ZORAXY_CONFIGSTORE_MIGRATION_MODE": "verify",
			"ZORAXY_CONFIGSTORE_DSN":            "postgres://example/zoraxy",
		},
		{
			"ZORAXY_CONFIGSTORE_MODE":           "control-plane",
			"ZORAXY_CONFIGSTORE_MIGRATION_MODE": "automatic",
			"ZORAXY_CONFIGSTORE_DSN":            "postgres://example/zoraxy",
		},
		{
			"ZORAXY_CONFIGSTORE_MODE":           "data-plane",
			"ZORAXY_CONFIGSTORE_MIGRATION_MODE": "verify",
			"ZORAXY_CONFIGSTORE_DSN":            "postgres://example/zoraxy",
			"ZORAXY_CONFIGSTORE_DSN_FILE":       "/run/secrets/dsn",
		},
	}
	for index, values := range tests {
		if _, err := loadClusterConfig(clusterTestEnvironment(values), func(string) ([]byte, error) { return nil, nil }); err == nil {
			t.Fatalf("case %d was accepted", index)
		}
	}
}
