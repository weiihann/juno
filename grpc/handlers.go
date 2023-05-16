package grpc

import (
	"github.com/NethermindEth/juno/db"
	"github.com/NethermindEth/juno/grpc/gen"
)

type handlers struct {
	db db.DB
}

func (h handlers) Tx(server gen.DB_TxServer) error {
	tx := h.db.NewTransaction(false)
	it, err := tx.NewIterator()
	if err != nil {
		return err
	}
	_ = it

	for {
		cursor, err := server.Recv()
		if err != nil {
			return err
		}
		_ = cursor

	}
}

func (h handlers) seek(server gen.DB_TxServer, key []byte) error {
	return nil
}
