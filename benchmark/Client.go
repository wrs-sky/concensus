package benchmark

import (
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"path/filepath"
)

type Client struct {
	in             Ingress
	out            Egress
	DeliverChanMap map[int]<-chan *Block
	Quorum         int
	NumNodes       int
	ReplyChan      chan *Block
	logger         smart.Logger
}

func NewClient(c Configuration, chains map[int]*Chain) *Client {
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
		DeliverChanMap: deliverChanMap,
		NumNodes:       NumNodes,
		Quorum:         NumNodes,
		ReplyChan:      make(chan *Block),
		logger:         loggerBasic,
	}

	go client.Listen()
	return client
}

func (c *Client) Listen() error {

	for id := 1; id <= c.NumNodes; id++ {
		go func(id int, deliverChan <-chan *Block) {
			for {
				select {
				case block := <-deliverChan:
					c.logger.Infof("block received %v", block)
					c.ReplyChan <- block
				}
			}

		}(id, c.DeliverChanMap[id])
	}

	return nil
}
