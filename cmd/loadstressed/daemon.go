package main

import (
	"flag"
	"sync"
	"context"
	"log"
	"time"
	"fmt"
	"sync/atomic"

	"loadstress/client"
	"loadstress/messages"
	_ "loadstress/client/grpc"

	"github.com/sirupsen/logrus"
)

const (
	STATUS_INT int32 = 0
	STATUS_STARTING
	STATUS_RUNNING
	STATUS_STOPPING
	STATUS_STOPPED
)

var (
	driverOpts client.DriverOpts
	qps 	= flag.Int("q", 10, "The number of concurrent RPCs in seconds on each connection.")
	numConn   = flag.Int("c", 1, "The number of parallel connections.")
	duration  = flag.Int("d", 60, "Benchmark duration in seconds")
	driver_name = flag.String("driver_name", "", "Name of the driver for benchmark profiles.")
	wg	sync.WaitGroup
	mu    sync.Mutex
	testDriver client.Driver
	restulCh = make(chan *loadstress_messages.CallResult, 100)
	status = STATUS_INT
	logger = logrus.New()
)

func init() {
	flag.StringVar(&driverOpts.Host, "server", "127.0.0.1", "Server to connect to.")
	flag.IntVar(&driverOpts.Port, "port", 50051, "Port to connect to.")
}
func main() {
	flag.Parse()

	testDriver, _ = client.GetDriver(*driver_name, &driverOpts)
	connectContext, connectCancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer connectCancel()

	deadline := time.Duration(*duration) * time.Second
	deadlineContext, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(deadline))
	defer deadlineCancel()

	status = STATUS_RUNNING
	css := buildConnections(connectContext)
	for _, cs := range css {
		wg.Add(1)
		go runWithConnection(deadlineContext, cs)
	}

	for {
		select {
		case <-deadlineContext.Done():
			stop(deadlineContext.Err())
			wg.Wait()
			close(restulCh)
			logger.Infof("finished loadstess.")
			return
		case r := <-restulCh:
			fmt.Printf("callId:%d result:%v elapsed:%d ns\n", r.Resp.RespId, r.Status, r.Elapsed)
		}
	}

}

func stop(err error) error{
	logger.Infof("stop loadstress because of:%s\n", err)
	if(!atomic.CompareAndSwapInt32(&status, STATUS_RUNNING, STATUS_STOPPING)){
		return nil
	}
	atomic.StoreInt32(&status, STATUS_STOPPED)
	return nil
}

func buildConnections(ctx context.Context) []client.ClientConnection {
	ccs := make([]client.ClientConnection, *numConn)
	var err error

	var optMap = map[string]interface{}{}
	optMap["timeout"] = int64(5)
	opts := client.CreateOpts{
		optMap,
	}

	for i := range ccs {
		ccs[i], err = testDriver.CreateConnection(ctx, &opts)
		if err != nil {
			log.Fatalf("create connection:%d failed:%v.", i, err)
		}
	}
	return ccs
}

func runWithConnection(ctx context.Context, conn client.ClientConnection) error {
	defer wg.Done()

	throttle := time.Tick(time.Second)
	var qwg sync.WaitGroup
	for {
		select {
		case <- ctx.Done():
			return stop(ctx.Err())
		case <- throttle:
			qwg.Add(1)
			runQps(ctx, conn, &wg)
		}
	}

	qwg.Wait()
	return nil
}

func handleCallError(req *loadstress_messages.SimpleRequest, err error) {

}

func sendResult(r *loadstress_messages.CallResult) bool {
	if(atomic.LoadInt32(&status) != STATUS_RUNNING) {
		return false
	}

	select {
		case restulCh <- r:
			return true
		default:
			logger.Warn("result channel full")
			return false
	}
}

func asynCall(ctx context.Context, conn client.ClientConnection, _wg *sync.WaitGroup) error {
	defer _wg.Done()

	req, err :=  conn.BuildReq()
	if err != nil {
		handleCallError(req, err)
		return err
	}

	resp, _ := conn.Call(ctx, req)

	callResult, _ := conn.BuildResp(resp)
	sendResult(callResult)

	return nil
}

func runQps(ctx context.Context, conn client.ClientConnection, _wg *sync.WaitGroup) error {
	defer _wg.Done()

	var qwg sync.WaitGroup
	for i := 0; i < *qps; i++ {
		select {
		case <-ctx.Done():
			return stop(ctx.Err())
		default:
			qwg.Add(1)
			go asynCall(ctx, conn, &qwg)
		}
	}

	qwg.Wait()
	return nil
}
