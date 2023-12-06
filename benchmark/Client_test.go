package benchmark

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var lock sync.Mutex

func TestAll(t *testing.T) {
	c := Client{
		stopChan: make(chan struct{}, 4),
		NumNodes: 4,
	}

	go func() {
		for {
			blockSeq := c.blockSeq

			//生成id随机数
			rand.Seed(time.Now().UnixNano())
			randID := rand.Intn(c.NumNodes) + 1

			fmt.Printf("tx%d send to node%d\n", blockSeq, randID)

			c.blockSeq = blockSeq + 1
			time.Sleep(1 * time.Second)

			if c.blockSeq > 10 {
				return
			}
		}
	}() //每个node启动监听
	time.Sleep(5 * time.Second)

	c.Quorum++
	c.Quorum++
	c.Quorum++
	time.Sleep(10 * time.Second)

}

func TestClient_Chan(t *testing.T) {
	deliverChan := make(chan string, 10)
	closeChan := make(chan struct{})

	for id := 1; id <= 4; id++ {
		go func() {
			for {
				select {
				case block := <-deliverChan:
					chanTest(block)
				case <-closeChan:
					return
				default:
					time.Sleep(1 * time.Second)
					fmt.Println("default")
				}
			}
		}()
	}

	for id := 1; id <= 3; id++ {
		closeChan <- struct{}{}
	}

	time.Sleep(20 * time.Second)
}

func chanTest(block string) {
	lock.Lock()
	defer lock.Unlock()
	fmt.Println("chanTest ", block)
	time.Sleep(3 * time.Second)
}
