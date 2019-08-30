package grpc

import (
	"context"
	"fmt"
	"loadstress/client"
	"loadstress/messages"
	"log"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

type Driver struct {
	id int64
}

type GrpcConnection struct {
	conn *grpc.ClientConn
	driver *Driver
	callTimeout time.Duration
}

func Init() (client.Driver, error){
	d := &Driver {}

	return d, nil
}

func (d *Driver) Name() string {
	return "grpc"
}

func (d* Driver) GenerateID() int64 {
	return atomic.AddInt64(&d.id, 1)
}

func (d *Driver) CreateConnection(ctx context.Context, opts *client.CreateOpts) (client.ClientConnection, error){
	host := opts.Opts["host"]
	port := opts.Opts["port"]
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("NewClientConn(%q) failed to create a ClientConn %v", addr, err)
		return nil, err
	}

	timeout := opts.Opts["timeout"].(int64)
	gc := GrpcConnection{
		conn: conn,
		driver: d,
		callTimeout: time.Duration(timeout),
	}
	return &gc, nil
}

func (d *Driver) CloseConnection(connection client.ClientConnection) error{
	if c, ok := connection.(*GrpcConnection); ok {
		err := c.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *GrpcConnection) BuildReq() (*loadstress_messages.SimpleRequest, error) {
	id := c.driver.GenerateID()
	req := loadstress_messages.SimpleRequest{
		ReqId: id,
	}
	return &req, nil
}

const defaultName string = "anger6"
func (c *GrpcConnection) Call(ctx context.Context, request *loadstress_messages.SimpleRequest) (*loadstress_messages.SimpleResponse, error){
	resp := loadstress_messages.SimpleResponse{
		RespId: request.ReqId,
	}
	cc := pb.NewGreeterClient(c.conn)

	cctx, cancel := context.WithTimeout(ctx, c.callTimeout*time.Second)
	defer cancel()

	r, err := cc.SayHello(cctx, &pb.HelloRequest{Name: defaultName})
	if err != nil{
		resp.Payload.Body = r.XXX_NoUnkeyedLiteral
	}
	return &resp, nil
}


func (c *GrpcConnection) BuildResp(response *loadstress_messages.SimpleResponse) (*loadstress_messages.CallResult, error){
	result := loadstress_messages.CallResult{

	}

	return &result, nil
}

