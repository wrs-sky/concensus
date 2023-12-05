package benchmark

import (
	"fmt"
	"testing"
)

func TestBenchmark(t *testing.T) {
	confFile := "conf/system.yml"
	Benchmark("./..", confFile)
}

func TestSetupWithClient(t *testing.T) {

	workDir := "./.."
	confFile := "conf/system.yml"

	//加载配置，并初始化
	c, err := InitConfig(workDir, confFile)
	if err != nil {
		fmt.Errorf("benchmark InitConfig(%s) error:%v", confFile, err)
		panic(err.Error())
	}

	fmt.Println("main init here")
	SetupWithClient(c)
}
