package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"time"

	"cxchain223/blockchain"
	"cxchain223/kvstore"
	"cxchain223/maker"
	"cxchain223/statdb"
	"cxchain223/statemachine"
	"cxchain223/txpool"
	"cxchain223/types"
	"cxchain223/utils/hash"
)

func main() {
	// 初始化区块链环境
	db := kvstore.NewLevelDB("./levedb")
	defer db.Close()

	genesisRoot := hash.Hash{}
	stateDB := statdb.NewMPTStatDB(db, genesisRoot)
	txPool := txpool.NewDefaultPool(stateDB)

	genesisHeader := blockchain.Header{
		Root:       genesisRoot,
		ParentHash: hash.Hash{},
		Height:     0,
		Coinbase:   types.Address{},
		Timestamp:  uint64(time.Now().Unix()),
		Nonce:      0,
	}

	chain := blockchain.NewBlockchain(genesisHeader, stateDB, txPool)

	config := maker.ChainConfig{
		Duration:   10 * time.Second,
		Coinbase:   types.Address{},
		Difficulty: 16,
	}

	exec := statemachine.StateMachine{}
	blockMaker := maker.NewBlockMaker(txPool, stateDB, exec, *chain, config)

	privateKey, err := generatePrivateKey()
	if err != nil {
		fmt.Println("Failed to generate private key:", err)
		return
	}

	addTransactions(txPool, privateKey)

	blockMaker.NewBlock()
	go blockMaker.Pack()

	time.Sleep(1 * time.Second)

	blockMaker.Interupt()

	header, body := blockMaker.Finalize()

	fmt.Printf("New block created with header: %+v and body: %+v\n", header, body)
}

// 生成并返回一个新的ECDSA私钥
func generatePrivateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// 模拟并添加一些交易到交易池中
func addTransactions(txPool *txpool.DefaultPool, privateKey *ecdsa.PrivateKey) {
	for i := 0; i < 5; i++ {
		tx := types.NewTransaction(types.Address{}, uint64(i+1), uint64(100), uint64(21000), uint64(i+1), []byte{})
		err := types.SignTx(&tx, privateKey)
		if err != nil {
			fmt.Println("Failed to sign transaction:", err)
			return
		}
		txPool.NewTx(&tx)
	}
}
