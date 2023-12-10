package client

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"strconv"
	"time"
)

// request 发送交易，超时重试
func (c *Client) request(blockSeq int) {

	//开启线程
	go func(blockSeq int) {
		var err = fmt.Errorf("times out")
		//重试3次
		for i := 0; i < c.configuration.RetryTimes; i++ {

			err = c.send(blockSeq)
			time.Sleep(time.Duration(c.configuration.RetryTimeout) * time.Millisecond)

			if err != nil {
				continue
			}
			//完成交易被down，则推出
			if _, ok := c.doneTX[blockSeq]; ok {
				return
			}
		}
		c.logger.Errorf("tx%d failed,err: %v", blockSeq, err)
	}(blockSeq)

}

// send 发送交易
func (c *Client) send(blockSeq int) error {
	version, err := c.obtainVersion()
	if err != nil {
		return err
	}

	c.versionMapLock.Lock()
	c.versionMap[blockSeq] = version
	c.versionMapLock.Unlock()

	leaderID := version.LeaderID
	err = c.chains[int(leaderID)].Order(Transaction{
		ClientID: "alice",
		ID:       fmt.Sprintf("tx%d", blockSeq),
	})
	if err != nil {
		c.logger.Errorf("tx%d order failed and err: %v", blockSeq, err)
		return err
	}

	c.Infof(fmt.Sprintf("tx%d send to node%d", blockSeq, leaderID))
	return nil
}

func (c *Client) HandleBlock(block Block, id int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.logger.Infof("block gotten from node%d:%s ", id, ObjToJson(block))

	//处理区块
	for _, transaction := range block.Transactions {
		txID, err := strconv.Atoi(transaction.ID[2:])
		if err != nil {
			c.logger.Errorf("txID convert failed and err: %v", err)
			continue
		}

		if _, ok := c.doneTX[txID]; ok {
			continue
		}

		c.collector[txID][uint64(id)] = struct{}{}
		version := c.versionMap[txID]

		if canBlockBeSubmitted(c.collector[txID], version) {
			c.doneTX[txID] = struct{}{}
			c.Infof(fmt.Sprintf("tx%d committed successfully", txID))
		}
	}

	if len(c.doneTX) == c.configuration.BlockCount {
		c.Infof(fmt.Sprintf("all txs committed successfully"))
		c.Close()
	}
	return
}

func canBlockBeSubmitted(collector map[uint64]struct{}, version types.Version) bool {
	if version.Type == types.OPTIMISTIC {
		//todo 乐观协议
		return false
	}
	return len(collector) >= version.Q
}
