package benchmark

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var lock sync.Mutex

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
