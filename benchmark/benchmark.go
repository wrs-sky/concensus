package benchmark

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"github.com/SmartBFT-Go/consensus/pkg/metrics/disabled"
	"github.com/SmartBFT-Go/consensus/pkg/wal"
	"github.com/golang/protobuf/proto"
	"time"
)

var testDir string
var suffix string
var logger smart.Logger

func Setup(configuration *Configuration) {
	blockCount := configuration.Block.Count
	suffix = configuration.Suffix
	logger = configuration.Logger
	testDir = configuration.Log.TestDir

	numNodes := configuration.Server.Num

	chains := setupNetwork(NetworkOptions{NumNodes: numNodes, BatchSize: 1, BatchTimeout: 10 * time.Second}, testDir)

	for blockSeq := 1; blockSeq < blockCount; blockSeq++ {
		err := chains[1].Order(Transaction{
			ClientID: "alice",
			ID:       fmt.Sprintf("tx%d", blockSeq),
		})
		if err != nil {
			panic(err)
		}
		for i := 1; i <= numNodes; i++ {
			chain := chains[i]
			block := chain.Listen()
			fmt.Sprintf("tx%d", block.Sequence)
			fmt.Sprintf("tx%d", blockSeq)
		}
	}

	for _, chain := range chains {
		chain.Stop()
	}
}

func setupNode(id int, opt NetworkOptions, network map[int]map[int]chan proto.Message, testDir string) *Chain {
	ingress := make(Ingress)
	for from := 1; from <= opt.NumNodes; from++ {
		ingress[from] = network[id][from]
	}

	egress := make(Egress)
	for to := 1; to <= opt.NumNodes; to++ {
		egress[to] = network[to][id]
	}

	met := &disabled.Provider{}
	walMet := wal.NewMetrics(met, "label1")
	bftMet := smart.NewMetrics(met, "label1")

	chain := NewChain(uint64(id), ingress, egress, logger, walMet, bftMet, opt, testDir)

	return chain
}

func setupNetwork(opt NetworkOptions, testDir string) map[int]*Chain {
	network := make(map[int]map[int]chan proto.Message)

	chains := make(map[int]*Chain)

	for id := 1; id <= opt.NumNodes; id++ {
		network[id] = make(map[int]chan proto.Message)
		for i := 1; i <= opt.NumNodes; i++ {
			network[id][i] = make(chan proto.Message, 128)
		}
	}

	for id := 1; id <= opt.NumNodes; id++ {
		chains[id] = setupNode(id, opt, network, testDir)
	}
	return chains
}
