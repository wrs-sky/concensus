package benchmark

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"

	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"github.com/SmartBFT-Go/consensus/pkg/metrics/disabled"
	"github.com/SmartBFT-Go/consensus/pkg/wal"
	"github.com/golang/protobuf/proto"
	"path/filepath"
)

func setupNode(id int, opt NetworkOptions, network map[int]map[int]chan proto.Message, testDir string, logger smart.Logger) *Chain {
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
		logFilePath := filepath.Join(configuration.Log.LogDir,
			fmt.Sprintf("node%d.log", id))
		logger, err := NewLogger(logFilePath)
		if err != nil {
			panic(err)
		}
		chains[id] = setupNode(id, opt, network, testDir, logger)
	}
	return chains
}
