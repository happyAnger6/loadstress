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
	"google.golang.org/grpc/status"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"github.com/square/go-jose/json"
	"google.golang.org/grpc/codes"
)

type Driver struct {
	id int64
}

type GrpcConnection struct {
	conn *grpc.ClientConn
	driver *Driver
	callTimeout time.Duration
}

type GrpcResp struct {
	RespMsg string
	Status *status.Status
}

func GrpcRespToBytes(r *GrpcResp) []byte {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return data
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

	respMsg := GrpcResp{
		RespMsg: nil,
		Status: nil,
	}

	r, err := cc.SayHello(ctx, &pb.HelloRequest{Name: defaultName})
	if err != nil{
		s, ok := status.FromError(err)
		if ok {
			respMsg.Status = s
		}
	}

	respMsg.RespMsg = r.GetMessage()
	data, err := json.Marshal(respMsg)
	if err != nil {
		resp.Payload.Body = data
	}
	return &resp, nil
}

func grpcCode2RetStatus(code codes.Code) loadstress_messages.RetStatus {
	switch code {
	case codes.OK:
		return loadstress_messages.RetStatus_SUCCESS
	case codes.DeadlineExceeded:
		return loadstress_messages.RetStatus_CALL_TIMEOUT
	default:
		return loadstress_messages.RetStatus_CALL_FAILED
	}
}

func (c *GrpcConnection) BuildResp(response *loadstress_messages.SimpleResponse, elapse time.Duration) (*loadstress_messages.CallResult, error){
	var respMsg GrpcResp
	json.Unmarshal(response.Payload.Body, &respMsg)

	result := loadstress_messages.CallResult{
		Errmsg: respMsg.Status.Err().Error(),
		Status: grpcCode2RetStatus(respMsg.Status.Code()),
		Elapsed: uint64(elapse),
	}
	return &result, nil
}

