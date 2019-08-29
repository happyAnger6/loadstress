package grpc

import (
	"loadstress/client"
	"context"
	"loadstress/messages"
)

type Driver struct {

}

type GrpcConnection struct {

}

func Init() (client.Driver, error){
	d := &Driver {}

	return d, nil
}

func (d *Driver) Name() string {
	return "grpc"
}

func (d *Driver) CreateConnection(ctx context.Context, opts *client.CreateOpts) (client.ClientConnection, error){
	gc := GrpcConnection{}

	return &gc, nil
}

func (d *Driver) CloseConnection(connection client.ClientConnection) error{
	return nil
}

func (c *GrpcConnection) BuildReq() (*loadstress_messages.SimpleRequest, error) {
	req := loadstress_messages.SimpleRequest{

	}
	return &req, nil
}

func (c *GrpcConnection) Call(ctx context.Context, request *loadstress_messages.SimpleRequest) (*loadstress_messages.SimpleResponse, error){
	resp := loadstress_messages.SimpleResponse{

	}
	return &resp, nil
}


func (c *GrpcConnection) BuildResp(response *loadstress_messages.SimpleResponse) (*loadstress_messages.CallResult, error){
	result := loadstress_messages.CallResult{

	}

	return &result, nil
}

