package storage

import (
	"encoding/json"
	"fmt"
	"dposi-blockchain/pkg/types"
	"github.com/dgraph-io/badger/v4"
)

// StateDB menyimpan state blockchain (skor, delegasi, cycle info)
type StateDB struct {
	db *badger.DB
}

// NewStateDB membuat instance StateDB baru
func NewStateDB(db *badger.DB) *StateDB {
	return &StateDB{db: db}
}

// SaveNodeScore menyimpan skor node
func (s *StateDB) SaveNodeScore(nodeScore *types.NodeScore) error {
	key := fmt.Sprintf("score:%x", nodeScore.NodeID)

	data, err := json.Marshal(nodeScore)
	if err != nil {
		return fmt.Errorf("failed to marshal node score: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// GetNodeScore mengambil skor node
func (s *StateDB) GetNodeScore(nodeID types.NodeID) (*types.NodeScore, error) {
	var nodeScore *types.NodeScore
	key := fmt.Sprintf("score:%x", nodeID)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &nodeScore)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get node score: %w", err)
	}

	return nodeScore, nil
}

// SaveDelegation menyimpan delegasi
func (s *StateDB) SaveDelegation(delegation *types.Delegation) error {
	key := fmt.Sprintf("delegation:%x:%x", delegation.DelegatorID, delegation.ValidatorID)

	data, err := json.Marshal(delegation)
	if err != nil {
		return fmt.Errorf("failed to marshal delegation: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// GetDelegation mengambil delegasi
func (s *StateDB) GetDelegation(delegatorID, validatorID types.NodeID) (*types.Delegation, error) {
	var delegation *types.Delegation
	key := fmt.Sprintf("delegation:%x:%x", delegatorID, validatorID)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &delegation)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get delegation: %w", err)
	}

	return delegation, nil
}

// SaveCycle menyimpan informasi siklus
func (s *StateDB) SaveCycle(cycle *types.Cycle) error {
	key := fmt.Sprintf("cycle:%d", cycle.CycleNumber)

	data, err := json.Marshal(cycle)
	if err != nil {
		return fmt.Errorf("failed to marshal cycle: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// GetCycle mengambil informasi siklus
func (s *StateDB) GetCycle(cycleNumber int64) (*types.Cycle, error) {
	var cycle *types.Cycle
	key := fmt.Sprintf("cycle:%d", cycleNumber)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &cycle)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get cycle: %w", err)
	}

	return cycle, nil
}

// GetCurrentCycle mengambil siklus saat ini
func (s *StateDB) GetCurrentCycle() (*types.Cycle, error) {
	var currentCycle *types.Cycle

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("cycle:")
		opts.Reverse = true

		it := txn.NewIterator(opts)
		defer it.Close()

		if it.Valid() {
			item := it.Item()
			data, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			err = json.Unmarshal(data, &currentCycle)
			return err
		}

		return nil
	})

	return currentCycle, err
}
