// Copyright 2026 vesoft inc. All Rights Reserved.
package nebula_ng

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
	"google.golang.org/protobuf/proto"
)

func TestUnaryDeliveryBytesAcrossServers(t *testing.T) {
	baseline := os.Getenv("NEBULA_UNARY_BASELINE_ADDRESS")
	candidate := os.Getenv("NEBULA_UNARY_CANDIDATE_ADDRESS")
	if baseline == "" || candidate == "" {
		t.Skip("two explicit unary server addresses required")
	}
	clients := make([]*connection, 0, 2)
	for _, address := range []string{baseline, candidate} {
		client, err := NewNebulaClient(address, "root", "NebulaGraph01")
		require.NoError(t, err)
		t.Cleanup(func() { _ = client.Close() })
		clients = append(clients, client.(*driverConn).conn.(*connection))
	}
	for _, query := range []string{
		"RETURN 1 AS x", "FOR x IN [] RETURN x", "FINISH",
		"RETURN null AS n, true AS b, 'abc' AS s, [1,2,3] AS xs",
		"FOR i IN range(1,1000000) RETURN i AS x ORDER BY i",
		"FOR i IN range(0,127) RETURN i AS i ORDER BY i NEXT RETURN repeat(CAST(i % 10 AS STRING),131072) AS x",
	} {
		t.Run(query, func(t *testing.T) {
			responses := make([]*graph.ExecuteResponse, 0, 2)
			for _, client := range clients {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				response, err := client.graphClient.Execute(ctx, &graph.ExecuteRequest{SessionId: client.sessionId, Stmt: []byte(query)})
				cancel()
				require.NoError(t, err)
				require.Equal(t, "00000", string(response.GetStatus().GetCode()), "%s", response.Status.String())
				responses = append(responses, response)
			}
			require.True(t, proto.Equal(responses[0].Status, responses[1].Status))
			require.Equal(t, responses[0].Cursor, responses[1].Cursor)
			require.Equal(t, responses[0].Result != nil, responses[1].Result != nil)
			for i, response := range responses {
				require.NotNil(t, response.Summary, "server %d", i)
			}
			before, err := proto.MarshalOptions{Deterministic: true}.Marshal(responses[0].Result)
			require.NoError(t, err)
			after, err := proto.MarshalOptions{Deterministic: true}.Marshal(responses[1].Result)
			require.NoError(t, err)
			require.Equal(t, before, after)
		})
	}
}
