package transfer

import (
	"crypto/tls"
	"github.com/gookit/color"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type tranfer struct{
	logger log.Logger
}

func (tranfer *tranfer) SetLogger(logger log.Logger) {
	tranfer.logger = logger
}

func (tranfer *tranfer) Execute(conn *tls.Conn) error {
	fileType, err0 := common.ReadString(conn)
	if err0 != nil {
		return err0
	}
	filePath, err1 := common.ReadString(conn)
	if err1 != nil {
		return err1
	}
	filePerm, err2 := common.ReadString(conn)
	if err2 != nil {
		return err2
	}
	perm, err4b := strconv.Atoi(filePerm)
	if err4b != nil {
		return err4b
	}
	if fileType == "folder" {
		if tranfer.logger != nil {
			tranfer.logger.Debugf("Creating folder: %s ...", filePath)
		} else {
			color.LightYellow.Printf("Creating folder: %s ... \n", filePath)
		}
		err3 := os.MkdirAll(filePath, os.FileMode(perm))
		if err3 != nil {
			if tranfer.logger != nil {
				tranfer.logger.Failuref("Errors creating folder: %s, Details: %s", filePath, err3.Error())
			} else {
				color.LightRed.Printf("Errors creating folder: %s, Details: %s \n", filePath, err3.Error())
			}
			return err3
		}
		if tranfer.logger != nil {
			tranfer.logger.Successf("Folder: %s created!!", filePath)
		} else {
			color.Green.Printf("Folder: %s created!!\n", filePath)
		}
	} else {
		data, err3 := common.Read(conn)
		if err3 != nil {
			return err3
		}
		var folder string = filepath.Dir(filePath)
		_, err4 := os.Stat(folder)
		if err4 != nil {
			if tranfer.logger != nil {
				tranfer.logger.Debugf("Creating folder: %s ...", folder)
			} else {
				color.LightYellow.Printf("Creating folder: %s ... \n", folder)
			}
			err4 = os.MkdirAll(folder, 0664)
			if err4 != nil {
				if tranfer.logger != nil {
					tranfer.logger.Failuref("Errors creating folder: %s, Details: %s", folder, err4.Error())
				} else {
					color.LightRed.Printf("Errors creating folder: %s, Details: %s \n", folder, err4.Error())
				}
			} else {
				if tranfer.logger != nil {
					tranfer.logger.Successf("Folder: %s created!!", folder)
				} else {
					color.Green.Printf("Folder: %s created!!\n", folder)
				}
			}

		}

		if tranfer.logger != nil {
			tranfer.logger.Debugf("Creating folder: %s ...", filePath)
		} else {
			color.LightYellow.Printf("Creating file: %s ... \n", filePath)
		}
		err5 := ioutil.WriteFile(filePath, data, os.FileMode(perm))
		if err5 != nil {
			if tranfer.logger != nil {
				tranfer.logger.Failuref("Errors: creating file: %s, Details: %s", filePath, err5.Error())
			} else {
				color.LightRed.Printf("Errors: creating file: %s, Details: %s\n", filePath, err5.Error())
			}
			return err5
		}
		if tranfer.logger != nil {
			tranfer.logger.Successf("File: %s created!!", filePath)
		} else {
			color.Green.Printf("File: %s created!!\n", filePath)
		}
	}
	return nil
}

func New() common.Commander {
	return &tranfer{}
}
