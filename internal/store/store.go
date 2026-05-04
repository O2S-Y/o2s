// Package store wraps an embedded BoltDB used for todos, snippets & metadata.
//
// We deliberately picked bbolt instead of SQLite to keep the binary CGO-free
// and trivially cross-compilable to Windows / macOS / Linux.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/O2S-Y/o2s/internal/config"
	bolt "go.etcd.io/bbolt"
)

const (
	BucketTodos     = "todos"
	BucketNotes     = "notes"
	BucketMeta      = "meta"
	BucketSnippets  = "snippets"
)

type Store struct{ DB *bolt.DB }

// Open opens (or creates) the o2s.db file inside the data directory.
func Open() (*Store, error) {
	dir, err := config.DataDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "o2s.db")
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range []string{BucketTodos, BucketNotes, BucketMeta, BucketSnippets} {
			if _, err := tx.CreateBucketIfNotExists([]byte(b)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{DB: db}, nil
}

func (s *Store) Close() error { return s.DB.Close() }

// Put marshals `v` as JSON and stores it under `bucket/key`.
func (s *Store) Put(bucket, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q missing", bucket)
		}
		return b.Put([]byte(key), data)
	})
}

// Get unmarshals JSON from `bucket/key` into `out`. Returns false if missing.
func (s *Store) Get(bucket, key string, out any) (bool, error) {
	found := false
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		raw := b.Get([]byte(key))
		if raw == nil {
			return nil
		}
		found = true
		return json.Unmarshal(raw, out)
	})
	return found, err
}

// Delete removes a key.
func (s *Store) Delete(bucket, key string) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}

// Iter walks all keys in a bucket; cb may unmarshal raw bytes.
func (s *Store) Iter(bucket string, cb func(k string, raw []byte) error) error {
	return s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if err := cb(string(k), v); err != nil {
				return err
			}
		}
		return nil
	})
}
