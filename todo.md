```go
Consensus.Start()
1. 确认通信配置，节点和相关配置网络验证

// 共识提交要完全排序的请求，并通过调用 Deliver（） 来交付给应用程序提案。这些建议包含由汇编程序组合在一起的批量请求。
type Consensus struct{a
    collector	控制共识的整个流程
    view		消息接受改变视图
}

流程
1. 客户端发布消息
2. 放入RequestPool
3. leader不断读取RequestPool,知道NextBatch
4. 获取Metadata然后boardcast
chain.node.consensus.SubmitRequest(txn.ToBytes())
c.RequestPool.Submit(request)

rp.submittedChan <- struct{}{}:

Controller.propose()
BatchBuilder.NextBatch()

leader
view.Propose()
将提案发送给自己，以便预先做好准备并将其记录在 WAL 中，然后再将其发送到其他节点。
case v.incMsgs <- msg:


```