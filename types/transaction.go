package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"cxchain223/crypto/secp256k1"
	"cxchain223/crypto/sha3"
	"cxchain223/utils/hash"
	"cxchain223/utils/rlp"

	"math/big"
)

type Receiption struct {
	TxHash hash.Hash
	Status int
	// GasUsed int
	// Logs
}

type Transaction struct {
	txdata
	Signature
	// 添加 From 字段
}

type txdata struct {
	To       Address
	Nonce    uint64
	Value    uint64
	Gas      uint64
	GasPrice uint64
	Input    []byte
}

type Signature struct {
	R, S *big.Int
	V    uint8
}

func (tx Transaction) Hash() hash.Hash {
	return hash.BytesToHash([]byte{})
}

// NewTransaction creates a new unsigned transaction
func NewTransaction(to Address, nonce, value, gas, gasPrice uint64, input []byte) Transaction {
	tx := Transaction{
		txdata: txdata{
			To:       to,
			Nonce:    nonce,
			Value:    value,
			Gas:      gas,
			GasPrice: gasPrice,
			Input:    input,
		},
	}
	return tx
}

// SignTx signs a transaction with the given private key
func SignTx(tx *Transaction, privateKey *ecdsa.PrivateKey) error {
	// Encode the transaction data to RLP
	toSign, err := rlp.EncodeToBytes(tx.txdata)
	if err != nil {
		return err
	}

	// Hash the encoded data
	msg := sha3.Keccak256(toSign)

	// Sign the hash with the private key
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, msg[:])
	if err != nil {
		return err
	}

	// Calculate V value by trying to recover the public key
	var v uint8
	for i := 0; i < 2; i++ {
		sig := append(r.Bytes(), s.Bytes()...)
		sig = append(sig, byte(i))

		// Try to recover the public key with the signature
		pubKey, err := secp256k1.RecoverPubkey(msg[:], sig)
		if err == nil && pubKey != nil {
			v = uint8(i + 27)
			break
		}
	}

	// Set the signature in the transaction
	tx.Signature = Signature{
		R: r,
		S: s,
		V: v,
	}
	return nil
}

// From recovers the sender's address from the transaction data and signature
func (tx Transaction) From() Address {
	txdata := tx.txdata

	// Encode the transaction data to RLP
	toSign, err := rlp.EncodeToBytes(txdata)
	if err != nil {
		// Return an empty address if encoding fails
		return Address{}
	}

	// Hash the transaction data
	msg := sha3.Keccak256(toSign)

	// Combine the signature
	sig := make([]byte, 65)
	copy(sig[:32], tx.Signature.R.Bytes())
	copy(sig[32:64], tx.Signature.S.Bytes())
	sig[64] = tx.Signature.V - 27 // V value should be 27 or 28, adjust to 0 or 1

	// Recover the public key using secp256k1
	pubKey, err := secp256k1.RecoverPubkey(msg[:], sig)
	if err != nil || pubKey == nil {
		// Return an empty address if public key recovery fails
		return Address{}
	}

	// Convert the public key to an address and return it
	return PubKeyToAddress(pubKey)
}
