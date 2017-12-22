package receiver

import (
	"github.com/anchnet/gateway/receiver/rpc"
	"github.com/anchnet/gateway/receiver/socket"
)

func Start() {
	go rpc.StartRpc()
	go socket.StartSocket()
}
