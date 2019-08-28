package client

import (
	"loadstress/messages"
	"context"
)

type CreateOpts struct {
	Opts map[string]string
}

type ClientConnection interface{
	BuildReq() loadstress_messages.SimpleRequest
	Call(ctx context.Context, request *loadstress_messages.SimpleRequest) (loadstress_messages.SimpleResponse, error)
	BuildResp(response loadstress_messages.SimpleResponse)
}

type ConnectionDriver interface {
	CreateConnection(opts *CreateOpts) (*ClientConnection, error)
	CloseConnection(cliCon *ClientConnection) (error)
}

type Driver interface {
	Name() string
	ConnectionDriver
}

type InitFunc func(root string, options []string) (Driver, error)

var (
	// All registered drivers
	drivers map[string]InitFunc
)

func init() {
	drivers = make(map[string]InitFunc)
}
