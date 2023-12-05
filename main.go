package main

import (
	"github.com/SmartBFT-Go/consensus/benchmark"
)

func main() {
	confFile := "conf/system.yml"
	benchmark.Benchmark(".", confFile)
}
