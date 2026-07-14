package types

import "time"

// BlockHeader adalah header dari setiap blok
type BlockHeader struct {
	Version       uint32    // Versi blockchain
	Height        int64     // Nomor blok (height)
	CycleNumber   int64     // Siklus ke berapa (1-1000)
	SlotInCycle   int64     // Posisi dalam siklus (0-999)
	Timestamp     time.Time // Waktu produksi blok
	ProducerID    NodeID    // Node yang memproduksi blok ini
	ParentHash    [32]byte  // Hash dari blok sebelumnya
	StateRoot     [32]byte  // Root hash dari merkle tree state
	TransactionRoot [32]byte // Root hash dari merkle tree tx
	Nonce         uint64    // Random nonce untuk security
}

// BlockBody adalah body dari setiap blok
type BlockBody struct {
	Transactions  []Transaction  // Daftar transaksi dalam blok
	Timestamp     int64          // Timestamp produksi
}

// Block merepresentasikan satu blok lengkap
type Block struct {
	Header    BlockHeader
	Body      BlockBody
	Signature []byte    // Ed25519 signature dari producer
	Hash      [32]byte  // SHA-256 hash dari block header
}

// Transaction merepresentasikan satu transaksi
type Transaction struct {
	TxID          [32]byte  // Unique transaction ID
	Sender        NodeID    // Pengirim
	Receiver      NodeID    // Penerima
	Amount        uint64    // Jumlah transfer
	Fee           uint64    // Gas fee
	Nonce         uint64    // Nonce untuk mencegah replay attack
	Timestamp     int64     // Waktu transaksi
	Signature     []byte    // Ed25519 signature dari sender
	Data          []byte    // Optional data (smart contract)
}

// BlockStatus melacak status finality blok
type BlockStatus struct {
	Height          int64
	Hash            [32]byte
	ProducerID      NodeID
	Status          string    // "pending", "committed", "finalized"
	ValidatorVotes  int       // Berapa validator yang approve
	FinalityTime    int64     // Waktu sampai finalized (ms)
	Slashed         bool      // Apakah producer di-slash
}

// CommitVote merepresentasikan vote dari validator terhadap blok
type CommitVote struct {
	ValidatorID   NodeID
	BlockHash     [32]byte
	BlockHeight   int64
	Timestamp     int64
	Signature     []byte
	Approved      bool      // true = approve, false = reject
}
