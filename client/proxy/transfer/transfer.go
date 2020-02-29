package transfer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-server/common"
	"io/ioutil"
	"os"
	"time"
)

type tranfer struct{}

var serverCommand string = "transfer-file"

func (tranfer *tranfer) SendMessage(conn *tls.Conn, params ...interface{}) error {
	var paramsLen int = len(params)
	if paramsLen < 2 {
		return errors.New(fmt.Sprintf("Insufficient number of parameters, expected 2 but give %v", paramsLen))
	}
	var origin string = fmt.Sprintf("%v", params[0])
	var destination string = fmt.Sprintf("%v", params[1])
	var perm = "0664"
	if len(params) > 2 {
		perm = fmt.Sprintf("%v", params[2])
	}
	_, err1 := os.Stat(origin)
	if err1 != nil {
		return err1
	}
	file, err2 := os.Open(origin)
	if err2 != nil {
		return err2
	}
	data, err3 := ioutil.ReadAll(file)
	if err3 != nil {
		return err3
	}
	n0, err3b := common.WriteString(serverCommand, conn)
	if err3b != nil {
		return err3b
	}
	if n0 == 0 {
		return errors.New(fmt.Sprintf("Unable to send command: %s", serverCommand))
	}
	time.Sleep(3 * time.Second)
	n1, err4 := common.WriteString(destination, conn)
	if err4 != nil {
		return err4
	}
	if n1 == 0 {
		return errors.New(fmt.Sprintf("Unable to send destination folder: %s", destination))
	}
	time.Sleep(3 * time.Second)
	n2, err5 := common.WriteString(perm, conn)
	if err5 != nil {
		return err5
	}
	if n2 == 0 {
		return errors.New(fmt.Sprintf("Unable to send file permissions: %s", perm))
	}
	time.Sleep(3 * time.Second)
	n3, err6 := common.Write(data, conn)
	if err6 != nil {
		return err6
	}
	if n3 == 0 {
		return errors.New(fmt.Sprintf("Unable to send data -> len: %v", len(data)))
	}
	return nil
}
func (tranfer *tranfer) Helper() string {
	return "transfer-file [origin] [destination]\n  Parameters:\n    [origin]      origin file path\n    [destination] remote file path\n"
}

func New() common.Sender {
	return &tranfer{}
}
