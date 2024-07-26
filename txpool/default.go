package txpool

import (
	"cxchain223/statdb"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"sort"
)

type SortedTxs interface {
	GasPrice() uint64
	Push(tx *types.Transaction)
	Replace(tx *types.Transaction)
	Pop() *types.Transaction
	Nonce() uint64
}

type DefaultSortedTxs []*types.Transaction

func (sorted DefaultSortedTxs) GasPrice() uint64 {
	if len(sorted) == 0 {
		return 0
	}
	return sorted[0].GasPrice
}

func (sorted *DefaultSortedTxs) Push(tx *types.Transaction) {
	*sorted = append(*sorted, tx)
	sort.SliceStable(*sorted, func(i, j int) bool {
		return (*sorted)[i].GasPrice > (*sorted)[j].GasPrice
	})
}

func (sorted *DefaultSortedTxs) Replace(tx *types.Transaction) {
	for i, t := range *sorted {
		if t.Nonce == tx.Nonce {
			(*sorted)[i] = tx
			sort.SliceStable(*sorted, func(i, j int) bool {
				return (*sorted)[i].GasPrice > (*sorted)[j].GasPrice
			})
			return
		}
	}
}

func (sorted *DefaultSortedTxs) Pop() *types.Transaction {
	if len(*sorted) == 0 {
		return nil
	}
	first := (*sorted)[0]
	*sorted = (*sorted)[1:]
	if len(*sorted) == 0 {
		*sorted = nil // 如果切片为空，将其设置为 nil
	}
	return first
}

func (sorted DefaultSortedTxs) Nonce() uint64 {
	if len(sorted) == 0 {
		return 0
	}
	return sorted[len(sorted)-1].Nonce
}

type pendingTxs []SortedTxs

func (p pendingTxs) Len() int {
	return len(p)
}

func (p pendingTxs) Less(i, j int) bool {
	return p[i].GasPrice() < p[j].GasPrice()
}

func (p pendingTxs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type DefaultPool struct {
	Stat     statdb.StatDB
	all      map[hash.Hash]bool
	txs      pendingTxs
	pendings map[types.Address][]SortedTxs
	queue    map[types.Address][]*types.Transaction
}

func NewDefaultPool(stat statdb.StatDB) *DefaultPool {
	pool := &DefaultPool{
		Stat:     stat,
		all:      make(map[hash.Hash]bool),
		txs:      pendingTxs{},
		pendings: make(map[types.Address][]SortedTxs),
		queue:    make(map[types.Address][]*types.Transaction),
	}
	return pool
}

func (pool *DefaultPool) SetStatRoot(root hash.Hash) {
	pool.Stat.SetStatRoot(root)
}

func (pool *DefaultPool) NewTx(tx *types.Transaction) {
	account := pool.Stat.Load(tx.From())
	if account == nil || account.Nonce >= tx.Nonce {
		return
	}

	nonce := account.Nonce
	blks := pool.pendings[tx.From()]
	if len(blks) > 0 {
		last := blks[len(blks)-1]
		nonce = last.Nonce()
	}
	if tx.Nonce > nonce+1 {
		pool.addQueueTx(tx)
	} else if tx.Nonce == nonce+1 {
		pool.pushPendingTx(blks, tx)
	} else {
		pool.replacePendingTx(blks, tx)
	}
	pool.all[tx.Hash()] = true
	sort.Sort(pool.txs)
}

func (pool *DefaultPool) replacePendingTx(blks []SortedTxs, tx *types.Transaction) {
	for _, blk := range blks {
		if blk.Nonce() >= tx.Nonce {
			if blk.GasPrice() <= tx.GasPrice {
				blk.Replace(tx)
			}
			break
		}
	}
}

func (pool *DefaultPool) pushPendingTx(blks []SortedTxs, tx *types.Transaction) {
	if len(blks) == 0 {
		blk := DefaultSortedTxs{tx}
		blks = append(blks, &blk)
		pool.pendings[tx.From()] = blks
		pool.txs = append(pool.txs, &blk)
		sort.Sort(pool.txs)
	} else {
		last := blks[len(blks)-1]
		if last.GasPrice() <= tx.GasPrice {
			last.Push(tx)
		} else {
			blk := DefaultSortedTxs{tx}
			blks = append(blks, &blk)
			pool.pendings[tx.From()] = blks
			pool.txs = append(pool.txs, &blk)
			sort.Sort(pool.txs)
		}
	}
}

func (pool *DefaultPool) addQueueTx(tx *types.Transaction) {
	list := pool.queue[tx.From()]
	list = append(list, tx)
	pool.queue[tx.From()] = list
	// sort transactions in the queue based on Nonce
	sort.SliceStable(pool.queue[tx.From()], func(i, j int) bool {
		return pool.queue[tx.From()][i].Nonce < pool.queue[tx.From()][j].Nonce
	})
}

func (pool *DefaultPool) Pop() *types.Transaction {
	for _, sortedTxs := range pool.txs {
		if tx := sortedTxs.Pop(); tx != nil {
			return tx
		}
	}
	return nil
}

func (pool *DefaultPool) NotifyTxEvent(txs []*types.Transaction) {
	// Implement transaction event notification logic
}
