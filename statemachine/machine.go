package statemachine

import (
	"cxchain223/statdb"
	"cxchain223/trie"
	"cxchain223/types"
	"cxchain223/utils/rlp"
)

type IMachine interface {
	Execute(state trie.ITrie, tx types.Transaction)
	Execute1(state statdb.StatDB, tx types.Transaction) *types.Receiption
}

type StateMachine struct {
}

func (m StateMachine) Execute(state trie.ITrie, tx types.Transaction) {
	from := tx.From()
	to := tx.To
	value := tx.Value
	gasUsed := tx.Gas
	if tx.Gas < 21000 {
		return
	} else {
		gasUsed = 21000
	}
	gasUsed = gasUsed * tx.GasPrice
	cost := value + gasUsed

	data, err := state.Load(from[:])
	if err != nil {
		return
	}
	var account types.Account
	_ = rlp.DecodeBytes(data, &account)

	if account.Amount < cost {
		return
	}

	account.Amount = account.Amount - cost
	data, err = rlp.EncodeToBytes(account)
	if err != nil {
		return
	}

	state.Store(from[:], data)

	data, err = state.Load(to[:])
	var toAccount types.Account
	if err != nil {
		toAccount = types.Account{}
	} else {
		rlp.DecodeBytes(data, &toAccount)
	}
	toAccount.Amount = toAccount.Amount + value
	data, err = rlp.EncodeToBytes(toAccount)
	if err != nil {
		return
	}

	state.Store(to[:], data)
}

func (m StateMachine) Execute1(state statdb.StatDB, tx types.Transaction) *types.Receiption {
	from := tx.From()
	to := tx.To
	value := tx.Value
	gasUsed := tx.Gas
	receipt := &types.Receiption{
		Status: 0, // Assume failure initially
		TxHash: tx.Hash(),
	}

	// Check if gas is sufficient
	if tx.Gas < 21000 {
		return receipt
	} else {
		gasUsed = 21000
	}
	gasUsed = gasUsed * tx.GasPrice
	cost := value + gasUsed

	// Load sender account
	fromAccount := state.Load(from)
	if fromAccount == nil {
		return receipt
	}

	// Check if sender has enough balance
	if fromAccount.Amount < cost {
		return receipt
	}

	// Deduct cost from sender account
	fromAccount.Amount -= cost
	state.Store(from, *fromAccount)

	// Load receiver account
	toAccount := state.Load(to)
	if toAccount == nil {
		toAccount = &types.Account{}
	}

	// Add value to receiver account
	toAccount.Amount += value
	state.Store(to, *toAccount)

	// Set receipt status to success
	receipt.Status = 1
	return receipt
}
