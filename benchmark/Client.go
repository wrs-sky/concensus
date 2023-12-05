package benchmark

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"math/rand"
	"path/filepath"
	"sync"
	"time"
)

type Client struct {
	Quorum   int
	NumNodes int
	chains   map[int]*Chain
	logger   smart.Logger

	blockSeq      int
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

	for id := 1; id <= NumNodes; id++ {
		deliverChanMap[id] = chains[id].DeliverChan
	}

	client := &Client{
		Quorum:   NumNodes,
		NumNodes: NumNodes,
		chains:   chains,
		logger:   loggerBasic,

		blockSeq:      1,
		configuration: c,

		deliverChanMap: deliverChanMap,
		replyChan:      make(chan *Block, NumNodes),
		stopChan:       make(chan struct{}, NumNodes),
	}

	go client.Start()
	return client
}

func (c *Client) Send() {

	chains := c.chains
	for {
		blockSeq := c.blockSeq

		//生成id随机数
		rand.Seed(time.Now().UnixNano())
		randID := rand.Intn(c.NumNodes) + 1
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

	for id := 1; id <= c.NumNodes*2; id++ {
		c.stopChan <- struct{}{}
	}

	c.logger.Infof("Client closed channel")
	fmt.Printf("Client closed channel\n")

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

	go c.Send()
	//每个node启动监听
	for id := 1; id <= c.NumNodes; id++ {
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
	//todo:c.HandleBlock(*block)为空
	// blockSeq 跳跃
	// c.downWG.Wait()捕获不到

}

func (c *Client) HandleBlock(block Block) {
	c.lock.Lock()
	defer c.lock.Unlock()

	//todo: 业务逻辑
	c.logger.Infof("block gotten:%s", ObjToString(block))
	return
}
