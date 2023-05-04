package grpc

import (
	"github.com/NethermindEth/juno/db"
	"github.com/NethermindEth/juno/grpc/gen"
)

type handlers struct {
	db db.DB
}

func (h handlers) Tx(cursor *gen.Cursor, server gen.DB_TxServer) error {
	tx := h.db.NewTransaction(false)
	it, err := tx.NewIterator()
	if err != nil {
		return err
	}

	for it.Seek(nil); it.Valid(); it.Next() {
		value, err := it.Value()
		if err != nil {
			return err
		}

		err = server.Send(&gen.Pair{
			K: it.Key(),
			V: value,
		})
		if err != nil {
			return nil
		}
	}

	return nil
}
