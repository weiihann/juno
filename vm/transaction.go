package vm

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/NethermindEth/juno/clients/feeder"
	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	"github.com/jinzhu/copier"
)

// 2^128
var queryBit = new(felt.Felt).Exp(new(felt.Felt).SetUint64(2), new(big.Int).SetUint64(128))

func marshalTxn(txn core.Transaction, zeroMaxFee bool) (json.RawMessage, error) {
	txnMap := make(map[string]any)

	var t feeder.Transaction
	if err := copier.Copy(&t, txn); err != nil {
		return nil, err
	}
	if zeroMaxFee {
		t.MaxFee = &felt.Zero
	}

	switch txn.(type) {
	case *core.InvokeTransaction:
		txnMap["Invoke"] = map[string]any{
			"V" + clearQueryBit(t.Version).Text(10): t,
		}
	case *core.DeployAccountTransaction:
		txnMap["DeployAccount"] = t
	case *core.DeclareTransaction:
		txnMap["Declare"] = map[string]any{
			"V" + clearQueryBit(t.Version).Text(10): t,
		}
	default:
		return nil, errors.New("unsupported txn type")
	}
	return json.Marshal(txnMap)
}

func clearQueryBit(v *felt.Felt) *felt.Felt {
	versionWithoutQueryBit := new(felt.Felt).Set(v)
	// if versionWithoutQueryBit >= queryBit
	if versionWithoutQueryBit.Cmp(queryBit) != -1 {
		versionWithoutQueryBit.Sub(versionWithoutQueryBit, queryBit)
	}
	return versionWithoutQueryBit
}
