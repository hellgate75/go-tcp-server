package proxy

import (
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-server/client/proxy/transfer"
	"github.com/hellgate75/go-tcp-server/common"
)

var sendersMap map[string]common.Sender = make(map[string]common.Sender)

var filled bool = false

func initMap() {
	sendersMap["transfer-file"] = transfer.New()
	filled = true
}

func GetSender(command string) (common.Sender, error) {
	if !filled {
		initMap()
	}
	if sender, ok := sendersMap[command]; ok {
		return sender, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Sender unavailable: %s", command))
	}
}

func Help() []string {
	var list []string = make([]string, 0)
	for _, sender := range sendersMap {
		list = append(list, sender.Helper())
	}
	return list
}
