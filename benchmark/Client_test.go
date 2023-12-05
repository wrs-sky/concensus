package benchmark

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var lock sync.Mutex

func TestAll(t *testing.T) {
	c := Client{
		stopChan: make(chan struct{}, 4),
	}
	//每个node启动监听
	for id := 1; id <= 4; id++ {
		go func(id int, deliverChan chan string) {
			for {
				select {
				case block := <-deliverChan:
					fmt.Println("block ", block)
				case <-c.stopChan:
					fmt.Println("stopChan b")
					return
				}
			}

		}(id, make(chan string, 10))
	}

	//统一处理block
	go func() {
		for {
			time.Sleep(10 * time.Second)
			select {
			case block := <-c.replyChan:
				c.HandleBlock(*block)
			case <-c.stopChan:
				fmt.Println("stopChan a")
				return
			}
		}
	}()

	for id := 1; id <= 4+1; id++ {
		c.stopChan <- struct{}{}
		fmt.Println("send stopChan ", id)
	}

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
