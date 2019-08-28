package main

import (
	"flag"
	"sync"
	"context"
	"log"
	"time"

	"loadstress/client"
	"loadstress/messages"
)

var (
	driverOpts client.DriverOpts
	qps 	= flag.Int("q", 10, "The number of concurrent RPCs in seconds on each connection.")
	numConn   = flag.Int("c", 1, "The number of parallel connections.")
	duration  = flag.Int("d", 60, "Benchmark duration in seconds")
	driver_name = flag.String("driver_name", "", "Name of the driver for benchmark profiles.")
	wg       sync.WaitGroup
	mu    sync.Mutex
	testDriver client.Driver
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
	css := buildConnections(connectContext)
	for _, cs := range css {
		runWithConnection(deadlineContext, cs)
	}
}

func buildConnections(ctx context.Context) []client.ClientConnection {
	ccs := make([]client.ClientConnection, *numConn)
	var err error
	for i := range ccs {
		ccs[i], err = testDriver.CreateConnection(ctx, nil)
		if err != nil {
			log.Fatalf("create connection:%d failed:%v.", i, err)
		}
	}
	return ccs
}

func runWithConnection(ctx context.Context, conn client.ClientConnection) error {
	return nil
}

func runQps(ctx context.Context, conn client.ClientConnection) error {
	select {
		case <- ctx.Done():
			return nil
	}
	for i := 0; i < *qps; i++ {
		wg.Add(1)
		go func() {
			var req *loadstress_messages.SimpleRequest
			var resp *loadstress_messages.SimpleResponse
			defer wg.Done()
			req, _ = conn.BuildReq()
			resp, _ = conn.Call(ctx, req)
			conn.BuildResp(resp)
		}()
	}
}