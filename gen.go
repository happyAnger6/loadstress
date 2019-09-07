package loadstress

import (
	"github.com/sirupsen/logrus"

	"time"
	"bytes"
	"fmt"
	"context"
	"sync/atomic"
	"errors"

	"loadstress/loadgen"
)

type GeneralLoadGen struct {
	caller loadgen.Caller
	timeoutInNS time.Duration	//请求超时时间
	qps uint64 //每秒的请求量
	durationInNS time.Duration //负载测试时间
	concurrency uint32 //并发量=qps*超时时间
	callPoll loadgen.CallPool
	lgCount uint64 //请求次数
	status uint32 //状态
	seqID	int64
	ctx context.Context
	cancelFunc context.CancelFunc
	resultCh chan *loadgen.CallResult
}

var logger = logrus.New()

func NewLoadGen(param GenParam) (loadgen.LoadGenerator, error) {
	if err := param.Check(); err != nil {
		return nil, err
	}

	lg := &GeneralLoadGen{
		caller: param.Caller,
		timeoutInNS: param.TimeoutInNS,
		qps: param.Qps,
		durationInNS: param.DurationInNS,
		status: loadgen.STATUS_INIT,
		seqID: 0,
		resultCh: param.ResultCh,
	}

	if err := lg.init(); err != nil {
		return nil, err
	}

	return lg, nil
}

func (lg *GeneralLoadGen) init() (error) {
	var buf bytes.Buffer
	buf.WriteString("GeneralLoadGen initialzer...")

	var concurrency uint64 = uint64(lg.timeoutInNS) / (1e9/lg.qps) + 1
	lg.concurrency = uint32(concurrency)
	lg.status = loadgen.STATUS_INIT

	pool, err := loadgen.NewCallPool(lg.concurrency)
	if err != nil {
		return err
	}

	lg.callPoll = pool
	buf.WriteString(fmt.Sprintf("Done. concurrency=%d.", concurrency))
	logger.Infoln(buf)
	return nil
}

func (lg *GeneralLoadGen) callOne(id int64) *loadgen.CallResult{
	atomic.AddUint64(&lg.lgCount, 1)
	if id <= 0{
		return &loadgen.CallResult{ID: -1, Err: errors.New("Invalid raw request.")}
	}
	start := time.Now().UnixNano()
	resp, err := lg.caller.Call(lg.timeoutInNS)
	end := time.Now().UnixNano()
	elapsedTime := time.Duration(end - start)
	var result loadgen.CallResult
	if err != nil {
		errMsg := fmt.Sprintf("Sync Call Error: %s.", err)
		result = loadgen.CallResult{
			ID:     id,
			Err:    errors.New(errMsg),
			Elapse: elapsedTime}
	} else {
		result = loadgen.CallResult{
			ID:     id,
			Resp:   resp,
			Elapse: elapsedTime}
	}
	return &result
}

// asyncSend 会异步地调用承受方接口。
func (lg *GeneralLoadGen) asyncCall() {
	lg.callPoll.P()
	reqID := atomic.AddInt64(&lg.seqID, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				err, ok := interface{}(p).(error)
				var errMsg string
				if ok {
					errMsg = fmt.Sprintf("Async Call Panic! (error: %s)", err)
				} else {
					errMsg = fmt.Sprintf("Async Call Panic! (clue: %#v)", p)
				}
				logger.Errorln(errMsg)
				result := &loadgen.CallResult{
					ID:   -1,
					Code: loadgen.RET_CALL_FAILED,
					Msg:  errMsg}
				lg.sendResult(result)
			}
			lg.callPoll.V()
		}()
		// 调用状态：0-未调用或调用中；1-调用完成；2-调用超时。
		var callStatus uint32
		timer := time.AfterFunc(lg.timeoutInNS, func() {
			if !atomic.CompareAndSwapUint32(&callStatus, 0, 2) {
				return
			}
			result := &loadgen.CallResult{
				ID:     reqID,
				Code:	loadgen.RET_TIMEOUT,
				Msg:    fmt.Sprintf("Timeout! (expected: < %v)", lg.timeoutInNS),
				Elapse: lg.timeoutInNS,
			}
			lg.sendResult(result)
		})
		result := lg.callOne(reqID)
		if !atomic.CompareAndSwapUint32(&callStatus, 0, 1) {
			return
		}
		timer.Stop()
		lg.sendResult(result)
	}()
}

// sendResult 用于发送调用结果。
func (lg *GeneralLoadGen) sendResult(result *loadgen.CallResult) bool {
	if atomic.LoadUint32(&lg.status) != loadgen.STATUS_RUNNING{
		lg.printIgnoredResult(result, "stopped load generator")
		return false
	}
	select {
	case lg.resultCh <- result:
		return true
	default:
		lg.printIgnoredResult(result, "full result channel")
		return false
	}
}

// printIgnoredResult 打印被忽略的结果。
func (lg *GeneralLoadGen) printIgnoredResult(result *loadgen.CallResult, cause string) {
	resultMsg := fmt.Sprintf(
		"ID=%d, Code=%d, Msg=%s, Elapse=%v",
		result.ID, result.Code, result.Msg, result.Elapse)
	logger.Warnf("Ignored result: %s. (cause: %s)\n", resultMsg, cause)
}

// prepareStop 用于为停止载荷发生器做准备。
func (lg *GeneralLoadGen) prepareToStop(ctxError error) {
	logger.Infof("Prepare to stop load generator (cause: %s)...", ctxError)
	atomic.CompareAndSwapUint32(
		&lg.status, loadgen.STATUS_RUNNING, loadgen.STATUS_STOPPING)
	logger.Infof("Closing result channel...")
	close(lg.resultCh)
	atomic.StoreUint32(&lg.status, loadgen.STATUS_STOPPED)
}

func (lg *GeneralLoadGen) loadLoop(emit <-chan time.Time) {
	for {
		select {
		case <-lg.ctx.Done():
			lg.prepareToStop(lg.ctx.Err())
			return
		default:
		}
		lg.asyncCall()
		if lg.qps > 0 {
			select {
			case <-emit:
			case <-lg.ctx.Done():
				lg.prepareToStop(lg.ctx.Err())
				return
			}
		}
	}
}

func (lg *GeneralLoadGen) Start() bool {
	logger.Infoln("GeneralLoadGen starting...")

	if !atomic.CompareAndSwapUint32(
		&lg.status, loadgen.STATUS_INIT, loadgen.STATUS_STARTING) {
			if !atomic.CompareAndSwapUint32(
				&lg.status, loadgen.STATUS_STOPPED, loadgen.STATUS_STARTING)	{
					return false
			}
	}

	var emitter <- chan time.Time
	if lg.qps > 0 {
		interval := time.Duration(1e9/lg.qps)
		emitter = time.Tick(interval)
		logger.Infoln("emit interval:%d", interval)
	}

	lg.ctx, lg.cancelFunc = context.WithTimeout(
		context.Background(), lg.timeoutInNS)

	lg.lgCount = 0
	atomic.StoreUint32(&lg.status, loadgen.STATUS_RUNNING)

	go func() {
		logger.Infoln("starting load generate loads.")
		lg.loadLoop(emitter)
		logger.Infoln("lgearte loads finished. total counts(%d)", lg.lgCount)
	}()

	return true
}

func (lg *GeneralLoadGen) Stop() bool {
	if !atomic.CompareAndSwapUint32(
		&lg.status, loadgen.STATUS_RUNNING, loadgen.STATUS_STOPPING) {
		return false
	}
	lg.cancelFunc()
	for {
		if atomic.LoadUint32(&lg.status) == loadgen.STATUS_STOPPED {
			break
		}
		time.Sleep(time.Microsecond)
	}
	return true
}

func (lg *GeneralLoadGen) Status() uint32 {
	return lg.status
}

func (lg *GeneralLoadGen) GenCount() uint64 {
	return lg.lgCount
}