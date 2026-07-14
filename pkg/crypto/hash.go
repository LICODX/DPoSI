package crypto

import (
	"crypto/sha256"
	"fmt"
)

// Hash32 merepresentasikan hash 32 bytes (SHA-256)
type Hash32 [32]byte

// SHA256 menghitung SHA-256 hash dari data
func SHA256(data []byte) Hash32 {
	hash := sha256.Sum256(data)
	return hash
}

// SHA256Hex menghitung SHA-256 hash dan mengembalikan hex string
func SHA256Hex(data []byte) string {
	return fmt.Sprintf("%x", SHA256(data))
}

// MerkleRoot menghitung merkle root dari daftar hash
func MerkleRoot(hashes []Hash32) Hash32 {
	if len(hashes) == 0 {
		return Hash32{}
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// Jika jumlah hash ganjil, duplikasi yang terakhir
	if len(hashes)%2 != 0 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}

	// Hitung parent nodes
	var parentNodes []Hash32
	for i := 0; i < len(hashes); i += 2 {
		combined := append(hashes[i][:], hashes[i+1][:]...)
		parentNodes = append(parentNodes, SHA256(combined))
	}

	// Rekursi sampai satu root
	return MerkleRoot(parentNodes)
}

// DoubleHash menghitung hash dari hash (untuk kestabilan)
func DoubleHash(data []byte) Hash32 {
	first := SHA256(data)
	return SHA256(first[:])
}
