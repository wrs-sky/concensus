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
	f      int      //容错个数
	N      int      //节点个数
	quorum []uint64 //投票节点
	nodes  []uint64 //所有节点

	logger        smart.Logger
	configuration Configuration

	blockSeq  int              //区块序号
	collector map[int]int      //收集到的交易
	doneTX    map[int]struct{} //已经提交的交易
	chains    map[int]*Chain   //区块链网络

	deliverChanMap map[int]<-chan *Block //每个节点的接收通道
	stopChan       chan struct{}
	closeChan      chan struct{}

	stopOnce sync.Once

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

	N := len(chains)
	deliverChanMap := make(map[int]<-chan *Block, N)
	for id := 1; id <= N; id++ {
		deliverChanMap[id] = chains[id].DeliverChan
	}

	client := &Client{
		chains:        chains,
		logger:        loggerBasic,
		configuration: c,

		deliverChanMap: deliverChanMap,
	}

	return client
}

func (c *Client) Start() {

	//节点配置
	c.q, c.f, c.quorum, c.nodes = c.chains[1].ObtainConfig()
	c.N = len(c.nodes)

	c.Infof(fmt.Sprintf("c config: q=%d, f=%d, N=%d, quorum=%v, nodes=%v", c.q, c.f, c.N, c.quorum, c.nodes))

	//区块配置
	collector := make(map[int]int, c.N)
	for id := 1; id <= c.N; id++ {
		collector[id] = 0
	}
	c.blockSeq = 1
	c.doneTX = make(map[int]struct{}, c.N+1)
	c.collector = collector

	//channel配置
	c.stopOnce = sync.Once{}
	c.stopChan = make(chan struct{}, c.N)
	c.closeChan = make(chan struct{}, c.N)

	go c.run()
}

func (c *Client) run() {
	//每个node启动监听
	for id := 1; id <= c.N; id++ {
		go func(id int, c *Client) {
			for {
				select {
				case <-c.stopChan:
					//todo:无法全部关闭
					c.Infof(fmt.Sprintf("Client stop listening on node %d", id))
					break
				case block := <-c.deliverChanMap[id]:
					c.HandleBlock(*block)
				}
			}
		}(id, c)

		c.logger.Infof("Client start listening on node %d", id)
	}

	//启动发送
	go c.Send()
}

func (c *Client) Send() {
	c.doneWG.Add(1)
	defer c.doneWG.Done()

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

		c.Infof(fmt.Sprintf("tx%d send to node%d", blockSeq, randID))

		c.blockSeq = blockSeq + 1
		time.Sleep(time.Duration(c.configuration.Block.IntervalTime) * time.Millisecond)

		if c.blockSeq > c.configuration.Block.Count {
			c.logger.Infof("all txs order successfully")
			return
		}
	}

}

func (c *Client) Close() {
	c.stopOnce.Do(
		func() {
			c.Infof(fmt.Sprintf("Client is closing"))

			for id := 1; id <= c.N; id++ {
				c.stopChan <- struct{}{}
			}
			c.closeChan <- struct{}{}

			c.Infof(fmt.Sprintf("Client closed channel"))
		},
	)
}

func (c *Client) Listen() {
	<-c.closeChan
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
