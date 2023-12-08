package client

import (
	"fmt"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"time"
)

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
