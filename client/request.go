package client

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	"math/rand"
	"time"
)

// request 发送交易，超时重试
func (c *Client) request(blockSeq int) {

	//开启线程
	go func(blockSeq int) {
		//重试3次
		for i := 0; i < c.configuration.Client.RetryTimes; i++ {

			if err := c.obtainConfig(); err != nil {
				continue
			}
			if err := c.send(blockSeq); err != nil {
				continue
			}

			time.Sleep(time.Duration(c.configuration.Client.RetryTimeout) * time.Millisecond)

			//完成交易被down，则推出
			if _, ok := c.doneTX[blockSeq]; ok {
				break
			}
		}
	}(blockSeq)

}

// send 发送交易
func (c *Client) send(blockSeq int) error {
	//生成id随机数
	rand.Seed(time.Now().UnixNano())
	randID := rand.Intn(c.N) + 1
	err := c.chains[randID].Order(Transaction{
		ClientID: "alice",
		ID:       fmt.Sprintf("tx%d", blockSeq),
	})
	if err != nil {
		c.logger.Errorf("tx%d order failed and err: %v", blockSeq, err)
		return err
	}

	c.Infof(fmt.Sprintf("tx%d send to node%d", blockSeq, randID))
	return nil
}
