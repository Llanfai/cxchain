package types

import "cxchain223/crypto/sha3"

type Address [20]byte

func PubKeyToAddress(pub []byte) Address {
	// h := sha3.Keccak256(pub)
	var addr Address
	// TODO hash得到addr
	hashedPubKey := sha3.Keccak256(pub)
	copy(addr[:], hashedPubKey[12:32])
	return addr
}
