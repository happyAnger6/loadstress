package client

type CreateOpts struct {
	Opts map[string]string
}

type ClientConnection struct {

}

type ConnectionDriver interface {
	CreateConnection(opts *CreateOpts) (*ClientConnection, error)
	CloseConnection(cliCon *ClientConnection) (error)
}

type CallDriver interface {
	BuildReq()
	Call()
	GetResp()
}

type Driver interface {
	Name() string
	ConnectionDriver
	CallDriver
}

type InitFunc func(root string, options []string) (Driver, error)

var (
	// All registered drivers
	drivers map[string]InitFunc
)

func init() {
	drivers = make(map[string]InitFunc)
}
