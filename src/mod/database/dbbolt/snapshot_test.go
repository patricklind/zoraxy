package dbbolt

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestSnapshotProducesReadableConsistentDatabase(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "source.db")
	database, err := NewBoltDatabase(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := database.NewTable("settings"); err != nil {
		t.Fatal(err)
	}
	if err := database.Write("settings", "revision", uint64(42)); err != nil {
		t.Fatal(err)
	}

	var snapshot bytes.Buffer
	if err := database.Snapshot(&snapshot); err != nil {
		t.Fatal(err)
	}

	restoredPath := filepath.Join(t.TempDir(), "restored.db")
	if err := os.WriteFile(restoredPath, snapshot.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	restored, err := bolt.Open(restoredPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()

	err = restored.View(func(tx *bolt.Tx) error {
		if got := tx.Bucket([]byte("settings")).Get([]byte("revision")); len(got) == 0 {
			t.Fatal("snapshot does not contain committed revision")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
