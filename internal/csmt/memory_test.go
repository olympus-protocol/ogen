package csmt_test

import (
	"github.com/olympus-protocol/ogen/internal/csmt"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNodeSerializeDeserialize(t *testing.T) {
	nodes := []csmt.Node{
		csmt.NewNode(chainhash.Hash{1, 2, 3}, &chainhash.Hash{1, 3, 5}, &chainhash.Hash{2, 4, 6}, nil, nil, true),
		csmt.NewNode(chainhash.Hash{1, 2, 3}, nil, nil, &chainhash.Hash{2, 3, 4}, nil, false),
		csmt.NewNode(chainhash.Hash{1, 2, 3}, nil, nil, &chainhash.Hash{3, 4, 5}, &chainhash.Hash{4, 5, 6}, false),
		csmt.NewNode(chainhash.Hash{1, 2, 3}, nil, nil, nil, &chainhash.Hash{5, 6, 7}, false),
	}

	for _, node := range nodes {
		nodeSer := node.Marshal()

		nodeUnmarshal := new(csmt.Node)
		err := nodeUnmarshal.Unmarshal(nodeSer)
		assert.NoError(t, err)
		assert.Equal(t, &node, nodeUnmarshal)
	}
}
