package client

import (
	"fmt"
	"github.com/SmartBFT-Go/consensus/benchmark"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"math/rand"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	q      int
	f      int      //容错个数
	N      int      //节点个数
	quorum []uint64 //投票节点
	nodes  []uint64 //所有节点

	logger        smart.Logger
	configuration benchmark.Configuration

	blockSeq  int              //区块序号
	collector map[int]int      //收集到的交易
	doneTX    map[int]struct{} //已经提交的交易
	chains    map[int]*Chain   //区块链网络

	deliverChanMap map[int]<-chan *Block //每个节点的接收通道
	stopChanMap    map[int]chan struct{} //每个节点的停止通道
	MsgChan        chan Message          //消息通道
	closeChan      chan struct{}

	stopOnce sync.Once

	lock   sync.Mutex
	doneWG sync.WaitGroup
}

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

func NewClient(c benchmark.Configuration, chains map[int]*Chain) *Client {

	//初始化logger
	logFilePath := filepath.Join(c.Log.LogDir, "client.log")
	loggerBasic, err := benchmark.NewLogger(logFilePath)
	if err != nil {
		panic(err)
	}

	N := len(chains)
	deliverChanMap := make(map[int]<-chan *Block, N)
	stopChanMap := make(map[int]chan struct{}, N)

	for id := 1; id <= N; id++ {
		deliverChanMap[id] = chains[id].DeliverChan
		stopChanMap[id] = make(chan struct{}, 1)
	}

	client := &Client{
		chains:         chains,
		logger:         loggerBasic,
		configuration:  c,
		stopChanMap:    stopChanMap,
		deliverChanMap: deliverChanMap,
	}

	return client
}

func (c *Client) Start() {
	//节点配置
	c.obtainConfig()

	//区块配置
	collector := make(map[int]int, c.N)
	for id := 1; id <= c.N; id++ {
		collector[id] = 0
	}
	c.blockSeq = 1
	c.doneTX = make(map[int]struct{}, c.N+1)
	c.collector = collector

	//channel配置
	c.MsgChan = make(chan Message)
	c.closeChan = make(chan struct{}, c.N)

	c.stopOnce = sync.Once{}

	go c.run()
}

func (c *Client) Close() {
	c.stopOnce.Do(
		func() {
			c.Infof(fmt.Sprintf("Client is closing"))

			//关闭通道
			for id := 1; id <= c.N; id++ {
				c.stopChanMap[id] <- struct{}{}
			}

			c.MsgChan <- Message{
				Type: CLOSE,
			}

			c.closeChan <- struct{}{}
			c.Infof(fmt.Sprintf("Client closed channel"))
		},
	)
}

func (c *Client) Listen() {
	<-c.closeChan
}

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

func (c *Client) obtainConfig() error {
	//节点配置
	c.q, c.f, c.quorum, c.nodes = c.chains[1].ObtainConfig()
	c.N = len(c.nodes)

	c.Infof(fmt.Sprintf("c config: q=%d, f=%d, N=%d, quorum=%v, nodes=%v", c.q, c.f, c.N, c.quorum, c.nodes))
	return nil
}

func (c *Client) reconfig(reconfig types.Reconfig) {
	c.Infof(fmt.Sprintf("reconfig start,new nodes:%v", reconfig.CurrentNodes))

	for id := 1; id <= c.N; id++ {
		c.chains[id].Reconfig(reconfig)
	}

	time.Sleep(1 * time.Second)
	c.obtainConfig()
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

//Infof logger+fmt双向输出
func (c *Client) Infof(format string) {
	c.logger.Infof(format)
	fmt.Println(format)
}
