package feeder

import (
	"context"
	"errors"
	"strconv"

	"github.com/NethermindEth/juno/adapters/feeder2core"
	"github.com/NethermindEth/juno/clients/feeder"
	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/juno/starknetdata"
)

var _ starknetdata.StarknetData = (*Feeder)(nil)

type Feeder struct {
	client *feeder.Client
}

func New(client *feeder.Client) *Feeder {
	return &Feeder{
		client: client,
	}
}

// BlockByNumber gets the block for a given block number from the feeder,
// then adapts it to the core.Block type.
func (f *Feeder) BlockByNumber(ctx context.Context, blockNumber uint64) (*core.Block, error) {
	return f.block(ctx, strconv.FormatUint(blockNumber, 10))
}

// BlockLatest gets the latest block from the feeder,
// then adapts it to the core.Block type.
func (f *Feeder) BlockLatest(ctx context.Context) (*core.Block, error) {
	return f.block(ctx, "latest")
}

func (f *Feeder) block(ctx context.Context, blockID string) (*core.Block, error) {
	response, err := f.client.Block(ctx, blockID)
	if err != nil {
		return nil, err
	}

	return feeder2core.AdaptBlock(response)
}

// Transaction gets the transaction for a given transaction hash from the feeder,
// then adapts it to the appropriate core.Transaction types.
func (f *Feeder) Transaction(ctx context.Context, transactionHash *felt.Felt) (core.Transaction, error) {
	response, err := f.client.Transaction(ctx, transactionHash)
	if err != nil {
		return nil, err
	}

	tx, err := feeder2core.AdaptTransaction(response.Transaction)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Class gets the class for a given class hash from the feeder,
// then adapts it to the core.Class type.
func (f *Feeder) Class(ctx context.Context, classHash *felt.Felt) (core.Class, error) {
	response, err := f.client.ClassDefinition(ctx, classHash)
	if err != nil {
		return nil, err
	}

	switch {
	case response.V1 != nil:
		compiledClass, cErr := f.client.CompiledClassDefinition(ctx, classHash)
		if cErr != nil {
			return nil, cErr
		}

		return feeder2core.AdaptCairo1Class(response.V1, compiledClass)
	case response.V0 != nil:
		return feeder2core.AdaptCairo0Class(response.V0)
	default:
		return nil, errors.New("empty class")
	}
}

// StateUpdate gets the state update for a given block number from the feeder,
// then adapts it to the core.StateUpdate type.
func (f *Feeder) StateUpdate(ctx context.Context, blockNumber uint64) (*core.StateUpdate, error) {
	response, err := f.client.StateUpdate(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	return feeder2core.AdaptStateUpdate(response)
}
