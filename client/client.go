package client

import (
	"fmt"
	. "github.com/SmartBFT-Go/consensus/examples/naive_chain"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"sync"
)

type Client struct {
	N int //全部节点个数

	logger        smart.Logger
	configuration Configuration

	blockSeq       int                         //区块序号
	collector      map[int]map[uint64]struct{} //收集到的交易回复 collector[blockSeq][nodeID]
	versionMap     map[int]types.Version       //发送交易的版本号 versionMap[blockSeq]
	versionMapLock sync.Mutex

	doneTX map[int]struct{} //已经提交的交易	doneTX[blockSeq]
	chains map[int]*Chain   //区块链网络	chains[nodeID]

	deliverChanMap map[int]<-chan *Block //每个节点的接收通道	 deliverChanMap[nodeID]
	stopChanMap    map[int]chan struct{} //每个节点的停止通道	 stopChanMap[nodeID]
	MsgChan        chan Message          //消息通道
	closeChan      chan struct{}

	stopOnce sync.Once

	lock   sync.Mutex
	doneWG sync.WaitGroup
}

type Configuration struct {
	BlockCount   int
	RetryTimes   int
	RetryTimeout int
}

func NewClient(chains map[int]*Chain, configuration Configuration, logger smart.Logger) *Client {

	N := len(chains)
	deliverChanMap := make(map[int]<-chan *Block, N)
	stopChanMap := make(map[int]chan struct{}, N)

	for id := 1; id <= N; id++ {
		deliverChanMap[id] = chains[id].DeliverChan
		stopChanMap[id] = make(chan struct{}, 1)
	}

	//初始化client，建立全部节点的通道
	client := &Client{
		N:              N,
		chains:         chains,
		logger:         logger,
		configuration:  configuration,
		stopChanMap:    stopChanMap,
		deliverChanMap: deliverChanMap,
	}

	return client
}

func (c *Client) Start() {

	//区块配置
	count := c.configuration.BlockCount
	collector := make(map[int]map[uint64]struct{}, count)
	for id := 1; id <= count; id++ {
		collector[id] = make(map[uint64]struct{}, c.N)
	}
	c.blockSeq = 1
	c.collector = collector
	c.versionMap = make(map[int]types.Version, count)
	c.doneTX = make(map[int]struct{}, count)

	//channel配置
	c.MsgChan = make(chan Message)
	c.closeChan = make(chan struct{}, count)

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
