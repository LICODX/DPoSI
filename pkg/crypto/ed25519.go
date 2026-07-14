package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// KeyPair menyimpan pasangan public dan private key
type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateKeyPair menghasilkan keypair Ed25519 baru
func GenerateKeyPair() (*KeyPair, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	return &KeyPair{
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

// Sign menandatangani message dengan private key
func (kp *KeyPair) Sign(message []byte) ([]byte, error) {
	if kp.PrivateKey == nil {
		return nil, fmt.Errorf("private key not available")
	}

	signature := ed25519.Sign(kp.PrivateKey, message)
	return signature, nil
}

// Verify memverifikasi signature dengan public key
func (kp *KeyPair) Verify(message []byte, signature []byte) bool {
	if kp.PublicKey == nil {
		return false
	}

	return ed25519.Verify(kp.PublicKey, message, signature)
}

// VerifyWithPublicKey memverifikasi signature menggunakan public key saja
func VerifyWithPublicKey(message []byte, signature []byte, publicKey ed25519.PublicKey) bool {
	if publicKey == nil || len(signature) != ed25519.SignatureSize {
		return false
	}

	return ed25519.Verify(publicKey, message, signature)
}

// PublicKeyToHex mengkonversi public key ke hex string
func PublicKeyToHex(pubKey ed25519.PublicKey) string {
	return hex.EncodeToString(pubKey)
}

// HexToPublicKey mengkonversi hex string ke public key
func HexToPublicKey(hexStr string) (ed25519.PublicKey, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}

	if len(bytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(bytes))
	}

	return ed25519.PublicKey(bytes), nil
}

// PrivateKeyToHex mengkonversi private key ke hex string (SECURE: gunakan dengan hati-hati)
func PrivateKeyToHex(privKey ed25519.PrivateKey) string {
	return hex.EncodeToString(privKey)
}

// HexToPrivateKey mengkonversi hex string ke private key
func HexToPrivateKey(hexStr string) (ed25519.PrivateKey, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}

	if len(bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(bytes))
	}

	return ed25519.PrivateKey(bytes), nil
}
