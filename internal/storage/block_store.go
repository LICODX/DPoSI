package storage

import (
	"encoding/json"
	"fmt"
	"dposi-blockchain/pkg/types"
	"github.com/dgraph-io/badger/v4"
)

// BlockStore menyimpan dan mengambil blok dari database
type BlockStore struct {
	db *badger.DB
}

// NewBlockStore membuat instance BlockStore baru
func NewBlockStore(db *badger.DB) *BlockStore {
	return &BlockStore{db: db}
}

// StoreBlock menyimpan blok ke database
func (bs *BlockStore) StoreBlock(block *types.Block) error {
	key := fmt.Sprintf("block:%d", block.Header.Height)

	data, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	err = bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return fmt.Errorf("failed to store block: %w", err)
	}

	// Store hash -> height mapping untuk quick lookup
	hashKey := fmt.Sprintf("blockhash:%x", block.Hash)
	heightData := fmt.Sprintf("%d", block.Header.Height)

	err = bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(hashKey), []byte(heightData))
	})

	return err
}

// GetBlock mengambil blok berdasarkan height
func (bs *BlockStore) GetBlock(height int64) (*types.Block, error) {
	var block *types.Block
	key := fmt.Sprintf("block:%d", height)

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &block)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	return block, nil
}

// GetBlockByHash mengambil blok berdasarkan hash
func (bs *BlockStore) GetBlockByHash(hash [32]byte) (*types.Block, error) {
	hashKey := fmt.Sprintf("blockhash:%x", hash)

	var height int64
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hashKey))
		if err != nil {
			return err
		}

		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		_, err = fmt.Sscanf(string(data), "%d", &height)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find block by hash: %w", err)
	}

	return bs.GetBlock(height)
}

// DeleteBlock menghapus blok (untuk pruning)
func (bs *BlockStore) DeleteBlock(height int64) error {
	key := fmt.Sprintf("block:%d", height)

	err := bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	return err
}

// LastBlockHeight mengambil height blok terakhir
func (bs *BlockStore) LastBlockHeight() (int64, error) {
	var lastHeight int64 = -1

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("block:")
		opts.Reverse = true

		it := txn.NewIterator(opts)
		defer it.Close()

		if it.Valid() {
			item := it.Item()
			key := string(item.Key())
			_, err := fmt.Sscanf(key, "block:%d", &lastHeight)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return lastHeight, err
}

// BlockExists mengecek apakah blok ada
func (bs *BlockStore) BlockExists(height int64) (bool, error) {
	key := fmt.Sprintf("block:%d", height)

	exists := false
	err := bs.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		exists = true
		return nil
	})

	return exists, err
}
