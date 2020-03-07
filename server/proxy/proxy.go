package proxy

import (
	"github.com/hellgate75/go-tcp-modules/server/proxy"
	"github.com/hellgate75/go-tcp-server/common"

)

func GetCommander(command string) (common.Commander, error) {
	return proxy.GetCommander(command)
}
