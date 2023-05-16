//go:build integration
// +build integration

package grpc

import (
	"context"
	"github.com/NethermindEth/juno/grpc/gen"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
)

func TestClient(t *testing.T) {
	conn, err := grpc.Dial(":8888", grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := gen.NewDBClient(conn)
	stream, err := client.Tx(context.Background())
	require.NoError(t, err)

	for {
		err = stream.Send(&gen.Cursor{})
		require.NoError(t, err)

		pair, err := stream.Recv()
		if err != nil {
			spew.Dump("error", err)
			break
		}

		spew.Dump(pair)
	}
}
