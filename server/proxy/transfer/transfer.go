package transfer

import (
	"crypto/tls"
	"github.com/hellgate75/go-tcp-server/common"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type tranfer struct{}

func (tranfer *tranfer) Execute(conn *tls.Conn) error {
	filePath, err1 := common.ReadString(conn)
	if err1 != nil {
		return err1
	}
	filePerm, err2 := common.ReadString(conn)
	if err2 != nil {
		return err2
	}
	data, err3 := common.Read(conn)
	if err3 != nil {
		return err3
	}
	var folder string = filepath.Dir(filePath)
	_, err4 := os.Stat(folder)
	if err4 != nil {
		os.MkdirAll(folder, 0664)
	}
	perm, err4b := strconv.Atoi(filePerm)
	if err4b != nil {
		return err4b
	}

	err5 := ioutil.WriteFile(filePath, data, os.FileMode(perm))
	if err5 != nil {
		return err5
	}
	return nil
}

func New() common.Commander {
	return &tranfer{}
}
