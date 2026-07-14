package storage

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
)

// Database adalah wrapper untuk BadgerDB
type Database struct {
	db *badger.DB
}

// NewDatabase membuka atau membuat database BadgerDB baru
func NewDatabase(path string) (*Database, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Disable default logger

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &Database{db: db}, nil
}

// Close menutup database
func (d *Database) Close() error {
	return d.db.Close()
}

// GetDB mengembalikan instance BadgerDB
func (d *Database) GetDB() *badger.DB {
	return d.db
}

// Backup membuat backup database
func (d *Database) Backup(path string) error {
	_, err := d.db.Backup(path, 0)
	return err
}

// RunGarbageCollection menjalankan garbage collection
func (d *Database) RunGarbageCollection() error {
	return d.db.RunValueLogGC(0.5)
}
