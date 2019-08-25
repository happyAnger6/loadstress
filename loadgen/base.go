package loadgen

import "time"

type RetCode int

const (
	RET_SUCCESS RetCode = 0
	RET_TIMEOUT = 1001 //响应超时
	RET_CANCEL = 2001 //应用取消
	RET_CALL_FAILED = 3001 //调用失败
	RET_SERVER_FAILED = 4001 //服务返回失败
)

type CallResp []byte

type CallResult struct {
	ID int64
	Resp CallResp
	Err error
	Code RetCode
	Msg string
	Elapse time.Duration
}

const (
	STATUS_INIT uint32 = 0
	STATUS_STARTING
	STATUS_RUNNING
	STATUS_STOPPING
	STATUS_STOPPED
)
