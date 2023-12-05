package benchmark

import (
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"path/filepath"
	"sync"
)

type Client struct {
	Quorum   int
	NumNodes int
	chains   map[int]*Chain
	logger   smart.Logger

	deliverChanMap map[int]<-chan *Block
	replyChan      chan *Block
	closeChan      chan struct{}
	lock           sync.Mutex
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
		NumNodes: NumNodes,
		Quorum:   NumNodes,
		chains:   chains,
		logger:   loggerBasic,

		deliverChanMap: deliverChanMap,
		replyChan:      make(chan *Block, NumNodes),
		closeChan:      make(chan struct{}),
	}

	go client.Listen()
	return client
}

func (c *Client) Send() {

}

func (c *Client) Close() {
	for id := 1; id <= c.NumNodes+1; id++ {
		c.closeChan <- struct{}{}
	}

	close(c.replyChan)
	for block := range c.replyChan {
		c.HandleBlock(*block)
	}
}

func (c *Client) Listen() {

	//每个node启动监听
	for id := 1; id <= c.NumNodes; id++ {
		go func(id int, deliverChan <-chan *Block) {
			for {
				select {
				case block := <-deliverChan:
					c.replyChan <- block
				case <-c.closeChan:
					return
				}
			}

		}(id, c.deliverChanMap[id])

		c.logger.Infof("Client  start listening on node %d", id)
	}

	//统一处理block
	go func() {
		for {
			select {
			case block := <-c.replyChan:
				c.HandleBlock(*block)
			case <-c.closeChan:
				return
			}
		}
	}()

}

func (c *Client) HandleBlock(block Block) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return
}
