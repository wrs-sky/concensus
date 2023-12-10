package client

import (
	"fmt"
	"github.com/SmartBFT-Go/consensus/pkg/types"
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
					c.Infof(fmt.Sprintf("Client stop listening on node%d", id))
					return
				case block := <-c.deliverChanMap[id]:
					c.HandleBlock(*block, id)
				}
			}
		}(id, c)

		c.logger.Infof("Client start listening on node%d", id)
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
		c.obtainVersion()
	case RECONFIG:
		c.reVersion(msg.Content.(types.Reconfig))
	default:
		c.logger.Errorf("msgType error")
	}
}
