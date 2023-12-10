package client

import (
	"encoding/json"
	"fmt"
	"github.com/SmartBFT-Go/consensus/pkg/types"
	"time"
)

func (c *Client) obtainVersion() (types.Version, error) {

	versionCount := make(map[string]int)
	//获取版本
	for id := 1; id <= c.N; id++ {
		version := c.chains[id].ObtainVersion()
		versionCount[ObjToJson(version)]++
	}
	// 找到出现次数最多的版本和次数
	var maxVersion types.Version
	maxCount := 0
	for version, count := range versionCount {
		if count > maxCount {
			maxCount = count
			_ = json.Unmarshal([]byte(version), &maxVersion)
		}
	}

	c.Infof(fmt.Sprintf("Version: q=%d, f=%d, N=%d, quorum=%v, nodes=%v",
		maxVersion.Q, maxVersion.F, maxVersion.N, maxVersion.Quorum, maxVersion.Nodes))

	if maxCount >= maxVersion.N {
		return maxVersion, nil
	}

	return maxVersion, fmt.Errorf("no maxVersion")
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
