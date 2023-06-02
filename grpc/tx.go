package grpc

import (
	"github.com/NethermindEth/juno/db"
)

type tx struct {
	dbTx      db.Transaction
	iterators map[uint32]db.Iterator
}

func newTx(dbTx db.Transaction) *tx {
	return &tx{
		dbTx:      dbTx,
		iterators: make(map[uint32]db.Iterator),
	}
}

func (t *tx) iterator(id uint32) (db.Iterator, error) {
	it, exists := t.iterators[id]
	if !exists {
		var err error
		it, err = t.dbTx.NewIterator()
		if err != nil {
			return nil, err
		}

		t.iterators[id] = it
	}

	return it, nil
}

func (t *tx) cleanup() error {
	for _, it := range t.iterators {
		it.Close()
	}

	return nil
}
