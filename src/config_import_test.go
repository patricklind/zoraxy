package main

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func configArchiveEntries(t *testing.T, entries map[string]string) []*zip.File {
	t.Helper()
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for name, contents := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		t.Fatal(err)
	}
	return reader.File
}

func TestValidateConfigArchive(t *testing.T) {
	valid := configArchiveEntries(t, map[string]string{
		"conf/proxy/root.config": `{}`,
		"sys.db":                 "database",
	})
	if err := validateConfigArchive(valid); err != nil {
		t.Fatalf("valid archive rejected: %v", err)
	}
}

func TestValidateConfigArchiveRejectsUnsafeEntries(t *testing.T) {
	tests := map[string]map[string]string{
		"parent traversal": {"../sys.db": "database"},
		"absolute path":    {"/tmp/sys.db": "database"},
		"foreign root":     {"plugins/evil": "payload"},
		"duplicate db":     {"sys.db": "one", "conf/../sys.db": "two"},
	}
	for name, entries := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateConfigArchive(configArchiveEntries(t, entries)); err == nil {
				t.Fatal("unsafe archive was accepted")
			}
		})
	}
}

func TestValidateConfigArchiveRejectsExpandedSizeLimit(t *testing.T) {
	files := configArchiveEntries(t, map[string]string{"conf/large": strings.Repeat("x", 1024)})
	files[0].UncompressedSize64 = maxConfigArchiveExpandedSize + 1
	if err := validateConfigArchive(files); err == nil {
		t.Fatal("oversized archive was accepted")
	}
}
