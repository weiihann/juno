package grpc

import (
	"fmt"
	"github.com/NethermindEth/juno/db"
	"github.com/NethermindEth/juno/grpc/gen"
	"github.com/NethermindEth/juno/utils"
)

type handlers struct {
	db db.DB
}

func (h handlers) Tx(server gen.KV_TxServer) error {
	dbTx := h.db.NewTransaction(false)
	tx := newTx(dbTx)

	for {
		cursor, err := server.Recv()
		if err != nil {
			return db.CloseAndWrapOnError(tx.cleanup, err)
		}

		err = h.handleTxCursor(cursor, tx, server)
		if err != nil {
			return db.CloseAndWrapOnError(tx.cleanup, err)
		}
	}
}

func (h handlers) handleTxCursor(
	cur *gen.Cursor,
	tx *tx,
	server gen.KV_TxServer,
) error {
	it, err := tx.iterator(cur.Cursor)
	if err != nil {
		return err
	}

	switch cur.Op {
	case gen.Op_SEEK:
		key := utils.Flatten(cur.BucketName, cur.K)
		it.Seek(key)
	case gen.Op_NEXT:
		it.Next()
	case gen.Op_CURRENT:
		var v []byte
		v, err = it.Value()
		if err != nil {
			return err
		}

		err = server.Send(&gen.Pair{
			K:        it.Key(),
			V:        v,
			CursorId: cur.Cursor,
		})
	default:
		err = fmt.Errorf("unknown operation %q", cur.Op)
	}

	return err
}
