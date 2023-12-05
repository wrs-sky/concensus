# Byzantine Fault-Tolerant Replicated State Machine Library



This is a Byzantine fault-tolerant (BFT) state machine replication (SMR) library. 
It is an open source library written in Go.
The implementation is inspired by the [BFT-SMaRt project](https://github.com/bft-smart/library). 
For more information on this library see our [wiki page](https://github.com/SmartBFT-Go/consensus/wiki).


## License

The source code files are made available under the Apache License, Version 2.0 (Apache-2.0), located in the [LICENSE](LICENSE) file.

## Quick start
To run any demonstration you first need to configure BFT-SMaRt to define the protocol behavior and the location of each replica.

Configuration file :/conf/system.yml
```shell
sh build.sh
./output/consensus
```


## Todo
1. 编写Client输出日志
2. 编写测试用例
3. 视图更改阅读
