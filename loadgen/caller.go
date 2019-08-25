package loadgen

import "time"

type Caller interface {
	Call(timeoutInNs time.Duration) (CallResp, error)
}
