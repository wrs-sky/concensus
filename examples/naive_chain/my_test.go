package naive

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestChainV2(t *testing.T) {

	testDir, err := os.MkdirTemp("", "naive_chain")
	assert.NoErrorf(t, err, "generate temporary test dir")
	defer os.RemoveAll(testDir)

	numNodes := 4
	chains := setupNetwork(t, NetworkOptions{NumNodes: numNodes, BatchSize: 1, BatchTimeout: 10 * time.Second}, testDir)

	blockSeq := 1
	err = chains[1].Order(Transaction{
		ClientID: "alice",
		ID:       fmt.Sprintf("tx%d", blockSeq),
	})
	assert.NoError(t, err)
	for i := 1; i <= numNodes; i++ {
		chain := chains[i]
		block := chain.Listen()
		assert.Equal(t, uint64(blockSeq), block.Sequence)
		assert.Equal(t, []Transaction{{ID: fmt.Sprintf("tx%d", blockSeq), ClientID: "alice"}}, block.Transactions)
	}

	for _, chain := range chains {
		chain.node.Stop()
	}
}
