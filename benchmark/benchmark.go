package benchmark

import (
	"fmt"
	"github.com/SmartBFT-Go/consensus/client"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"os"
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

	fmt.Println("Configuration:", client.ObjToJson(c))
	fmt.Println("---------------------")

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

	//初始化logger
	logFilePath := filepath.Join(c.Log.LogDir, "client.log")
	loggerBasic, err := NewLogger(logFilePath)
	if err != nil {
		panic(err)
	}

	client := client.NewClient(chains, client.Configuration{
		BlockCount:   c.Block.Count,
		RetryTimes:   c.Client.RetryTimes,
		RetryTimeout: c.Client.RetryTimeout,
	}, loggerBasic)
	client.Start()

	go run(client)

	go func() {
		client.Listen()
		fmt.Println("client listen done")
		exit(chains)
	}()

	duration := c.System.Timeout
	for i := 0; i < duration; i++ {
		time.Sleep(1 * time.Second)
		if i > duration/2 {
			fmt.Printf("%ds count down\n", duration-i)
		}
	}
	client.Close()
	exit(chains)
}

//测试程序
func run(c *client.Client) {
	blockSeq := 1
	for {
		if blockSeq == 10 {
			currentNodes := []uint64{1, 2, 3, 4, 5}
			c.MsgChan <- client.Message{
				Type: client.RECONFIG,
				Content: types.Reconfig{
					InLatestDecision: true,
					CurrentNodes:     currentNodes,
					CurrentConfig:    types.DefaultConfig,
				},
			}
		}

		c.MsgChan <- client.Message{
			Type:    client.REQUSET,
			Content: blockSeq,
		}

		time.Sleep(time.Duration(configuration.Block.IntervalTime) * time.Millisecond)
		blockSeq++

		if blockSeq > configuration.Block.Count {
			c.Infof("all txs order successfully")

			time.Sleep(10 * time.Second)
			c.Close()
			return
		}
	}

}

func exit(chains map[int]*Chain) {
	for _, chain := range chains {
		chain.Stop()
	}
	fmt.Println("Done")
	os.Exit(0)
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
