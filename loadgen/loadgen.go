package loadgen

type LoadGenerator interface {
	//启动
	Start() bool
	//停止
	Stop() bool
	//获取当前状态
	Status() uint32
	//产生的负载个数
	GenCount() uint64
}
