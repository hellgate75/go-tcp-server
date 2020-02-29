package proxy

import (
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/server/proxy/transfer"
)

var commandsMap map[string]common.Commander = make(map[string]common.Commander)

var filled bool = false

func initMap() {
	commandsMap["transfer-file"] = transfer.New()
	filled = true
}

func GetCommander(command string) (common.Commander, error) {
	if !filled {
		initMap()
	}
	if commander, ok := commandsMap[command]; ok {
		if commander == nil {
			return nil, errors.New(fmt.Sprintf("Unable to collect command: %s", command))
		}
		return commander, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Command unavailable: %s", command))
	}
}
