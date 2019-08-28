package client

type CreateOpts struct {
	Opts map[string]string
}

type ClientConnection interface{
	BuildReq()
	Call()
	BuildResp()
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
