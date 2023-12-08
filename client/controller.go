package client

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"strconv"
)

type Message struct {
	Type    MsgType
	Content interface{}
}

type MsgType int

const (
	RECONFIG = iota
	REQUSET
	ObtainConfig
	CLOSE
)

func (c *Client) run() {
	//每个node启动监听
	for id := 1; id <= c.N; id++ {
		go func(id int, c *Client) {
			for {
				select {
				case <-c.stopChanMap[id]:
					//todo:无法全部关闭
					c.Infof(fmt.Sprintf("Client stop listening on node %d", id))
					return
				case block := <-c.deliverChanMap[id]:
					c.HandleBlock(*block)
				}
			}
		}(id, c)

		c.logger.Infof("Client start listening on node %d", id)
	}

	//消息监听
	go func() {
		for {
			select {
			case message := <-c.MsgChan:
				if message.Type == CLOSE {
					c.Infof(fmt.Sprintf("Client stop message listening"))
					c.Close()
					return
				}
				c.HandleMessage(message)
			}
		}
	}()
}

func (c *Client) HandleMessage(msg Message) {
	switch msg.Type {
	case REQUSET:
		c.request(msg.Content.(int))
	case ObtainConfig:
		c.obtainConfig()
	case RECONFIG:
		c.reconfig(msg.Content.(types.Reconfig))
	default:
		c.logger.Errorf("msgType error")
	}
}

func (c *Client) HandleBlock(block Block) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.logger.Infof("block gotten:%s", ObjToJson(block))

	//处理区块
	for _, transaction := range block.Transactions {
		txID, err := strconv.Atoi(transaction.ID[2:])
		if err != nil {
			c.logger.Errorf("txID convert failed and err: %v", err)
			continue
		}

		c.collector[txID] = c.collector[txID] + 1
		if c.collector[txID] >= c.q {
			c.doneTX[txID] = struct{}{}
			c.Infof(fmt.Sprintf("tx%d committed successfully", txID))
		}
	}

	if len(c.doneTX) == c.configuration.Block.Count {
		c.Infof(fmt.Sprintf("all txs committed successfully"))
		c.Close()
	}
	return
}
