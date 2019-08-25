package loadstress

import (
	"loadstress/loadgen"
	"time"
	"bytes"
	"strings"
	"fmt"
	"errors"
)

//Generator params
type GenParam struct {
	Caller loadgen.Caller
	DurationInNS time.Duration
	TimeoutInNS time.Duration
	Qps uint64
	ResultCh chan *loadgen.CallResult
}


func (genParam *GenParam) Check() error {
	var errMsgs []string

	if genParam.Caller == nil {
		errMsgs = append(errMsgs, "Invalid caller!")
	}
	if genParam.TimeoutInNS == 0 {
		errMsgs = append(errMsgs, "Invalid timeoutInNS!")
	}
	if genParam.Qps == 0 {
		errMsgs = append(errMsgs, "Invalid qps(load per second)!")
	}
	if genParam.DurationInNS == 0 {
		errMsgs = append(errMsgs, "Invalid durationInNS!")
	}
	if genParam.ResultCh == nil {
		errMsgs = append(errMsgs, "Invalid result channel!")
	}
	var buf bytes.Buffer
	buf.WriteString("Checking the parameters...")
	if errMsgs != nil {
		errMsg := strings.Join(errMsgs, " ")
		buf.WriteString(fmt.Sprintf("NOT passed! (%s)", errMsg))
		logger.Infoln(buf.String())
		return errors.New(errMsg)
	}
	buf.WriteString(
		fmt.Sprintf("Passed. (timeoutInNS=%s, qps=%d, durationInNS=%s)",
			genParam.TimeoutInNS, genParam.Qps, genParam.DurationInNS))
	logger.Infoln(buf.String())
	return nil
}
