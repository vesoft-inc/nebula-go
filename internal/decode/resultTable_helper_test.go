package decode

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/vector"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

func TestDecodeDecimalValue_LongDecimalKeepsDecimalType(t *testing.T) {
	decimal := "1234567890123.456789"
	header := make([]byte, 16)
	order.PutUint32(header[:4], uint32(len(decimal)))
	order.PutUint32(header[8:12], 0)
	order.PutUint32(header[12:16], 0)

	nested := &vector.NestedVector{
		VectorData: header,
		NestedVectors: []*vector.NestedVector{
			{VectorData: []byte(decimal)},
		},
	}

	value := &NebulaValue{}
	err := defaultDecoder.(*vectorDecoder).decodeDecimalValue()(
		&decodeContext{},
		value,
		nested,
		0,
		&columnTypeSchemaBasic{typ: types.ColumnTypeDecimal},
	)
	require.NoError(t, err)

	require.Equal(t, types.ValueTypeDecimal, value.GetType())
	decoded, err := value.AsDecimal()
	require.NoError(t, err)
	require.Equal(t, decimal, decoded.String())
}
