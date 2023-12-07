package benchmark

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"os"
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

	fmt.Println("Configuration:", ObjToJson(c))
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

	client := NewClient(*c, chains)
	client.Start()

	go client.benchmark()

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
func (c *Client) benchmark() {

	for {
		if c.blockSeq == 10 {
			currentNodes := []uint64{1, 2, 3, 4, 5}
			c.MsgChan <- Message{
				Type: RECONFIG,
				Content: types.Reconfig{
					InLatestDecision: true,
					CurrentNodes:     currentNodes,
					CurrentConfig:    types.DefaultConfig,
				},
			}
		}

		c.MsgChan <- Message{
			Type: REQUSET,
		}

		time.Sleep(time.Duration(c.configuration.Block.IntervalTime) * time.Millisecond)

		if c.blockSeq > c.configuration.Block.Count {
			c.Infof("all txs order successfully")

			time.Sleep(5 * time.Second)
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
