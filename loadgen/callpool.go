package loadgen


import (
	"errors"
	"fmt"
)

// CallPool 表示Goroutine池的接口。
type CallPool interface {
	P()
	V()
	Active() bool
	Total() uint32
	Spare() uint32
}

type myCallPool struct {
	total    uint32
	poolCh chan struct{}
	active   bool
}

func NewCallPool(total uint32) (CallPool, error) {
	cp := myCallPool{}
	if !cp.init(total) {
		errMsg :=
			fmt.Sprintf("The goroutine ticket pool can NOT be initialized! (total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	return &cp, nil
}

func (cp *myCallPool) init(total uint32) bool {
	if cp.active {
		return false
	}
	if total == 0 {
		return false
	}
	ch := make(chan struct{}, total)
	n := int(total)
	for i := 0; i < n; i++ {
		ch <- struct{}{}
	}
	cp.poolCh = ch
	cp.total = total
	cp.active = true
	return true
}

func (cp *myCallPool) P() {
	<-cp.poolCh
}

func (cp *myCallPool) V() {
	cp.poolCh <- struct{}{}
}

func (cp *myCallPool) Active() bool {
	return cp.active
}

func (cp *myCallPool) Total() uint32 {
	return cp.total
}

func (cp *myCallPool) Spare() uint32 {
	return uint32(len(cp.poolCh))
}
