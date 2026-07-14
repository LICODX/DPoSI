package types

// GenesisConfig berisi konfigurasi genesis blockchain
type GenesisConfig struct {
	// Konsensus parameters
	MaxNodesPerCycle    int32  `toml:"max_nodes_per_cycle"`   // 100
	BlocksPerNode       int32  `toml:"blocks_per_node"`       // 10
	BlockTimeMs         int64  `toml:"block_time_ms"`         // 3000ms
	SpeedTestInterval   int32  `toml:"speed_test_interval"`   // Setiap 10 blok
	SlashThreshold      float64 `toml:"slash_threshold"`       // Score < threshold -> slash
	OfflineThreshold    int32  `toml:"offline_threshold"`     // Berapa kali offline sebelum slash

	// Network parameters
	MinimumStake        uint64 `toml:"minimum_stake"`         // Minimum stake untuk join
	MaximumSlashing     uint64 `toml:"maximum_slashing"`      // Max slash amount
	UndelegationPeriod  int64  `toml:"undelegation_period"`   // Blok sampai unlock

	// Speed test parameters
	SpeedTestTimeout    int32  `toml:"speed_test_timeout"`    // Timeout tes (detik)
	SpeedTestDataSize   int64  `toml:"speed_test_data_size"`  // Ukuran data test (bytes)
	VerifierCount       int32  `toml:"verifier_count"`        // Berapa validator verify
	VerifyThreshold     float64 `toml:"verify_threshold"`      // % persetujuan needed

	// Storage parameters
	DBPath              string `toml:"db_path"`
	PruneInterval       int64  `toml:"prune_interval"`        // Blok sampai prune

	// Initial state
	TotalSupply         uint64 `toml:"total_supply"`
	InitialBalance      map[string]uint64 `toml:"-"`  // Address -> balance
}

// NodeConfig berisi konfigurasi lokal node
type NodeConfig struct {
	// Node identity
	NodeID     string `toml:"node_id"`
	Address    string `toml:"address"`      // IP:Port
	PublicKey  string `toml:"public_key"`   // Base58 encoded
	PrivateKey string `toml:"private_key"`  // Encrypted

	// Network
	P2PPort    uint16   `toml:"p2p_port"`
	BootNodes  []string `toml:"boot_nodes"`
	MaxPeers   uint16   `toml:"max_peers"`

	// Storage
	DBPath     string `toml:"db_path"`

	// Stake & delegation
	StakeAmount uint64 `toml:"stake_amount"`

	// Geo location untuk handicap
	Region     string `toml:"region"`  // "AS", "EU", "NA", etc
}
