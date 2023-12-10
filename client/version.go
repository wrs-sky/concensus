package client

import (
	"fmt"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"time"
)

func (c *Client) obtainVersion() (types.Version, error) {
	//节点配置
	version := c.chains[1].ObtainVersion()

	return version, nil
}

func (c *Client) reVersion(reconfig types.Reconfig) {
	c.Infof(fmt.Sprintf("reconfig start,new nodes:%v", reconfig.CurrentNodes))

	for id := 1; id <= c.N; id++ {
		c.chains[id].ReVersion(reconfig)
	}

	time.Sleep(2 * time.Second)
	version := c.chains[1].ObtainVersion()
	c.Infof(fmt.Sprintf("ReVersion: q=%d, f=%d, N=%d, quorum=%v, nodes=%v",
		version.Q, version.F, version.N, version.Quorum, version.Nodes))
}
