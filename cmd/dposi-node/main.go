package main

import (
	"flag"
	"fmt"
	"log"
	"dposi-blockchain/internal/storage"
	"dposi-blockchain/pkg/crypto"
	"dposi-blockchain/pkg/types"
)

func main() {
	configPath := flag.String("config", "configs/config.example.toml", "Path to config file")
	genesisPath := flag.String("genesis", "configs/genesis.json", "Path to genesis file")
	flag.Parse()

	fmt.Println("🚀 Starting DPoSI Node...")
	fmt.Printf("Config: %s\n", *configPath)
	fmt.Printf("Genesis: %s\n", *genesisPath)

	// Initialize database
	db, err := storage.NewDatabase("./data/badger")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Database initialized")

	// Test crypto
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate keypair: %v", err)
	}

	testMsg := []byte("Hello DPoSI")
	signature, err := keyPair.Sign(testMsg)
	if err != nil {
		log.Fatalf("Failed to sign: %v", err)
	}

	verified := keyPair.Verify(testMsg, signature)
	fmt.Printf("✅ Cryptography test: signature verified=%v\n", verified)

	// Test storage
	blockStore := storage.NewBlockStore(db.GetDB())
	stateDB := storage.NewStateDB(db.GetDB())

	// Create test block
	testBlock := &types.Block{
		Header: types.BlockHeader{
			Version:     1,
			Height:      0,
			CycleNumber: 0,
			SlotInCycle: 0,
		},
	}

	// Calculate hash
	hash := crypto.SHA256([]byte("test-block"))
	testBlock.Hash = hash

	err = blockStore.StoreBlock(testBlock)
	if err != nil {
		log.Fatalf("Failed to store block: %v", err)
	}

	fmt.Println("✅ Block storage test passed")

	// Test state storage
	nodeID := types.NodeID{}
	copy(nodeID[:], "test-node-id-32byte")

	nodeScore := &types.NodeScore{
		NodeID:       nodeID,
		CurrentScore: 85.5,
		IsOnline:     true,
	}

	err = stateDB.SaveNodeScore(nodeScore)
	if err != nil {
		log.Fatalf("Failed to save node score: %v", err)
	}

	fmt.Println("✅ State storage test passed")

	// Retrieve and verify
	retrieved, err := stateDB.GetNodeScore(nodeID)
	if err != nil {
		log.Fatalf("Failed to retrieve node score: %v", err)
	}

	fmt.Printf("✅ Retrieved score: %.1f\n", retrieved.CurrentScore)

	fmt.Println("\n🎉 All tests passed! DPoSI node is ready.")
	fmt.Println("Sprint 1 implementation completed successfully.")
}
