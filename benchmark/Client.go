package benchmark

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"math/rand"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	q      int
	f      int
	N      int
	quorum []uint64
	nodes  []uint64

	chains map[int]*Chain
	logger smart.Logger

	blockSeq      int
	collector     map[int]int
	configuration Configuration

	deliverChanMap map[int]<-chan *Block
	replyChan      chan *Block
	stopChan       chan struct{}
	closeChan      chan struct{}

	lock   sync.Mutex
	doneWG sync.WaitGroup
}

func NewClient(c Configuration, chains map[int]*Chain) *Client {

	//初始化logger
	logFilePath := filepath.Join(c.Log.LogDir, "client.log")
	loggerBasic, err := NewLogger(logFilePath)
	if err != nil {
		panic(err)
	}

	NumNodes := len(chains)
	deliverChanMap := make(map[int]<-chan *Block, NumNodes)
	collector := make(map[int]int, NumNodes+1)

	for id := 1; id <= NumNodes; id++ {
		collector[id] = 0
		deliverChanMap[id] = chains[id].DeliverChan
	}

	client := &Client{
		chains: chains,
		logger: loggerBasic,

		blockSeq:      1,
		collector:     collector,
		configuration: c,

		deliverChanMap: deliverChanMap,
		replyChan:      make(chan *Block, NumNodes),
		stopChan:       make(chan struct{}, NumNodes),
		closeChan:      make(chan struct{}, 1),
	}

	client.q, client.f, client.quorum, client.nodes = chains[1].ObtainConfig()
	client.N = len(client.nodes)

	client.logger.Infof("client config: q=%d, f=%d, N=%d, quorum=%v, nodes=%v", client.q, client.f, client.N, client.quorum, client.nodes)
	fmt.Printf("client config: q=%d, f=%d, N=%d, quorum=%v, nodes=%v\n", client.q, client.f, client.N, client.quorum, client.nodes)
	return client
}

func (c *Client) Send() {

	chains := c.chains
	for {
		blockSeq := c.blockSeq

		//生成id随机数
		rand.Seed(time.Now().UnixNano())
		randID := rand.Intn(c.N) + 1
		err := chains[randID].Order(Transaction{
			ClientID: "alice",
			ID:       fmt.Sprintf("tx%d", blockSeq),
		})
		if err != nil {
			c.logger.Errorf("tx%d order failed and err: %v", blockSeq, err)
			continue
		}

		c.logger.Infof("tx%d send to node%d", blockSeq, randID)
		fmt.Printf("tx%d send to node%d\n", blockSeq, randID)

		c.blockSeq = blockSeq + 1
		time.Sleep(1 * time.Second)

		if c.blockSeq > c.configuration.Block.Count {
			c.logger.Infof("all txs order successfully")
			return
		}
	}

}

func (c *Client) Close() {
	fmt.Printf("Client is closing\n")
	c.logger.Infof("Client is closing")

	for id := 1; id <= c.N+1; id++ {
		c.stopChan <- struct{}{}
	}

	c.logger.Infof("Client close channel")
	fmt.Printf("Client close channel\n")

	close(c.replyChan)
	for block := range c.replyChan {
		c.HandleBlock(*block)
	}
	c.logger.Infof("Client closed")
	fmt.Printf("Client closed\n")

	c.closeChan <- struct{}{}
}

func (c *Client) Listen() {
	<-c.closeChan
}

func (c *Client) Start() {

	//每个node启动监听
	for id := 1; id <= c.N; id++ {
		go func(id int) {
			for {
				select {
				case <-c.stopChan:
					return
				case block := <-c.deliverChanMap[id]:
					c.replyChan <- block
				}
			}
		}(id)

		c.logger.Infof("Client start listening on node %d", id)
	}

	//统一处理block
	go func() {

		for {
			select {
			case <-c.stopChan:
				return
			case block := <-c.replyChan:
				c.HandleBlock(*block)
			}
		}
	}()

	//启动发送
	go c.Send()

}

func (c *Client) HandleBlock(block Block) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.logger.Infof("block gotten:%s", ObjToString(block))

	//处理区块
	for _, transaction := range block.Transactions {
		txID, err := strconv.Atoi(transaction.ID[2:])
		if err != nil {
			c.logger.Errorf("txID convert failed and err: %v", err)
			continue
		}
		c.collector[txID] = c.collector[txID] + 1
		if c.collector[txID] == c.q {
			c.logger.Infof("tx%d committed successfully", txID)
			fmt.Printf("tx%d committed successfully\n", txID)
		}
	}
	return
}
