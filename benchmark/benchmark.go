package benchmark

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"github.com/SmartBFT-Go/consensus/pkg/metrics/disabled"
	"github.com/SmartBFT-Go/consensus/pkg/wal"
	"github.com/golang/protobuf/proto"
	"path/filepath"
	"time"
)

var configuration *Configuration

func Benchmark(workDir string, confFile string) {
	//加载配置，并初始化
	c, err := InitConfig(workDir, confFile)
	if err != nil {
		fmt.Errorf("benchmark InitConfig(%s) error:%v", confFile, err)
		panic(err.Error())
	}

	fmt.Println("main init here")

	SetupWithClient(c)
}

func SetupWithClient(c *Configuration) {

	configuration = c
	numNodes := configuration.Server.Num

	chains := setupNetwork(NetworkOptions{
		NumNodes:     numNodes,
		BatchSize:    uint64(configuration.Server.BatchSize),
		BatchTimeout: 10 * time.Second,
	}, configuration.Log.TestDir)

	client := NewClient(*c, chains)

	client.Start()

	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		if i > 10 {
			fmt.Printf("%ds count down\n", 30-i)
		}
	}
	client.Close()

	for _, chain := range chains {
		chain.Stop()
	}
	fmt.Println("Done")
}

func Setup(c *Configuration) {

	blockCount := configuration.Block.Count
	numNodes := configuration.Server.Num

	chains := setupNetwork(NetworkOptions{
		NumNodes:     numNodes,
		BatchSize:    uint64(c.Server.BatchSize),
		BatchTimeout: 10 * time.Second,
	}, c.Log.TestDir)

	for blockSeq := 1; blockSeq < blockCount; blockSeq++ {
		err := chains[1].Order(Transaction{
			ClientID: "alice",
			ID:       fmt.Sprintf("tx%d", blockSeq),
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Sprintf("tx%d order successfully", blockSeq))

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
	fmt.Println("Done")
}

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
