package client

import (
	"loadstress/messages"
	"context"
	"fmt"
)

type CreateOpts struct {
	Opts map[string]interface{}
}

type ClientConnection interface{
	BuildReq() (*loadstress_messages.SimpleRequest, error)
	Call(ctx context.Context, request *loadstress_messages.SimpleRequest) (*loadstress_messages.SimpleResponse, error)
	BuildResp(response *loadstress_messages.SimpleResponse) (*loadstress_messages.CallResult, error)
}

type ConnectionDriver interface {
	CreateConnection(ctx context.Context, opts *CreateOpts) (ClientConnection, error)
	CloseConnection(cliCon ClientConnection) (error)
}

type Driver interface {
	Name() string
	GetID() (int64)
	ConnectionDriver
}

type DriverOpts struct {
	Host string
	Port int
}

type InitFunc func(root string, options *DriverOpts) (Driver, error)

var (
	// All registered drivers
	drivers map[string]InitFunc
)

func init() {
	drivers = make(map[string]InitFunc)
}

func Register(name string, initFunc InitFunc) error {
	if _, exists := drivers[name]; exists {
		return fmt.Errorf("Name already registerd:%s.", name)
	}

	drivers[name] = initFunc
	return nil
}

func GetDriver(name string, options *DriverOpts) (Driver, error) {
	if initFunc, exist := drivers[name]; exist {
		return initFunc(name, options)
	}

	return nil, fmt.Errorf("Driver: %s not exist.", name)
}