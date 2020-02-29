package proxy

import (
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/server/proxy/transfer"
	"strings"
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
	var cmd string = strings.TrimSpace(strings.ToLower(command))
	if commander, ok := commandsMap[cmd]; ok {
		if commander == nil {
			return nil, errors.New(fmt.Sprintf("Unable to collect command: %s", cmd))
		}
		return commander, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Command unavailable: %s", cmd))
	}
}
