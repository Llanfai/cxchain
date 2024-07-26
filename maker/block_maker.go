package maker

import (
	"cxchain223/blockchain"
	"cxchain223/statdb"
	"cxchain223/statemachine"
	"cxchain223/txpool"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"cxchain223/utils/xtime"
	"time"
)

type ChainConfig struct {
	Duration   time.Duration
	Coinbase   types.Address
	Difficulty uint64
}

type BlockMaker struct {
	txpool txpool.TxPool
	state  statdb.StatDB
	exec   statemachine.IMachine

	config ChainConfig
	chain  blockchain.Blockchain

	nextHeader *blockchain.Header
	nextBody   *blockchain.Body

	interupt chan bool
}

func NewBlockMaker(txpool txpool.TxPool, state statdb.StatDB, exec statemachine.IMachine, chain blockchain.Blockchain, config ChainConfig) *BlockMaker {
	return &BlockMaker{
		txpool:   txpool,
		state:    state,
		exec:     exec,
		config:   config,
		chain:    chain,
		interupt: make(chan bool),
	}
}

func (maker *BlockMaker) NewBlock() {
	maker.nextBody = &blockchain.Body{
		Transactions: []types.Transaction{},
		Receiptions:  []types.Receiption{},
	}
	maker.nextHeader = &blockchain.Header{
		ParentHash: maker.chain.CurrentHeader.Hash(),
		Coinbase:   maker.config.Coinbase,
	}
}

func (maker *BlockMaker) Pack() {
	end := time.After(maker.config.Duration)
	for {
		select {
		case <-maker.interupt:
			return
		case <-end:
			return
		default:
			maker.pack()
		}
	}
}

func (maker *BlockMaker) pack() {
	tx := maker.txpool.Pop()
	if tx == nil {
		return
	}
	receiption := maker.exec.Execute1(maker.state, *tx)
	maker.nextBody.Transactions = append(maker.nextBody.Transactions, *tx)
	maker.nextBody.Receiptions = append(maker.nextBody.Receiptions, *receiption)
}

func (maker *BlockMaker) Interupt() {
	maker.interupt <- true
}

func (maker *BlockMaker) Finalize() (*blockchain.Header, *blockchain.Body) {
	maker.nextHeader.Timestamp = xtime.Now()
	maker.nextHeader.Nonce = 0

	for n := 0; ; n++ {
		maker.nextHeader.Nonce = uint64(n)
		if validHash(maker.nextHeader.Hash(), maker.config.Difficulty) {
			break
		}
	}

	return maker.nextHeader, maker.nextBody
}

func validHash(hash hash.Hash, difficulty uint64) bool {
	leadingZeros := int(difficulty / 8)
	remainingBits := int(difficulty % 8)

	for i := 0; i < leadingZeros; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	if remainingBits > 0 {
		mask := byte(0xff << (8 - remainingBits))
		if hash[leadingZeros]&mask != 0 {
			return false
		}
	}

	return true
}
