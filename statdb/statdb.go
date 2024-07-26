package statdb

import (
	"cxchain223/kvstore"
	"cxchain223/trie"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"cxchain223/utils/rlp"
)

type StatDB interface {
	SetStatRoot(root hash.Hash)
	Load(addr types.Address) *types.Account
	Store(addr types.Address, account types.Account)
	Root() hash.Hash
}

type MPTStatDB struct {
	state *trie.State
}

func NewMPTStatDB(db kvstore.KVDatabase, root hash.Hash) *MPTStatDB {
	return &MPTStatDB{
		state: trie.NewState(db, root),
	}
}

func (db *MPTStatDB) SetStatRoot(root hash.Hash) {
	db.state = trie.NewState(db.state.DB(), root)
}

func (db *MPTStatDB) Load(addr types.Address) *types.Account {
	data, err := db.state.Load(addr[:])
	if err != nil {
		return nil
	}
	var account types.Account
	err = rlp.DecodeBytes(data, &account)
	if err != nil {
		return nil
	}
	return &account
}

func (db *MPTStatDB) Store(addr types.Address, account types.Account) {
	data, err := rlp.EncodeToBytes(account)
	if err != nil {
		panic(err)
	}
	err = db.state.Store(addr[:], data)
	if err != nil {
		panic(err)
	}
}

func (db *MPTStatDB) Root() hash.Hash {
	return db.state.Root()
}
