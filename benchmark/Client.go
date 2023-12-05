package benchmark

import (
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	"path/filepath"
)

type Client struct {
	in             Ingress
	out            Egress
	DeliverChanMap map[int]<-chan *Block
	Quorum         int
	NumNodes       int
	ReplyChan      chan *Block
}

func NewClient(c Configuration, chains map[int]*Chain) *Client {
	logFilePath := filepath.Join(c.Log.LogDir, "client.log")
	logger, err := NewLogger(logFilePath)
	if err != nil {
		panic(err)
	}

	logger.Infof("Starting client")

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
	}

	go client.listen()
	return client
}

func (c *Client) listen() error {
	return nil
}
