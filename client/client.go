package client

import (
	"fmt"
	"github.com/SmartBFT-Go/consensus/benchmark"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"path/filepath"
	"sync"
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

//Infof logger+fmt双向输出
func (c *Client) Infof(format string) {
	c.logger.Infof(format)
	fmt.Println(format)
}
