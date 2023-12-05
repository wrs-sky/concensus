package benchmark

import "testing"

func TestBenchmark(t *testing.T) {
	confFile := "conf/system.yml"
	Benchmark("./..", confFile)
}
