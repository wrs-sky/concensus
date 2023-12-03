package main

import (
	"fmt"
	"github.com/SmartBFT-Go/consensus/benchmark"
)

func main() {
	Setup()
}

func Setup() {

	//加载配置
	confFile := "conf/system.yml"

	//加载配置，并初始化
	configuration, err := benchmark.InitConfig(".", confFile)
	if err != nil {
		fmt.Errorf("benchmark InitConfig(%s) error:%v", confFile, err)
		panic(err.Error())
	}

	fmt.Println("main init here")

	benchmark.Setup(configuration)

}
