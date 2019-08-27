package grpc

import "loadstress/client"

type Driver struct {

}

func Init() (*client.Driver, error){
	d := &Driver {}

	return d, nil
}
