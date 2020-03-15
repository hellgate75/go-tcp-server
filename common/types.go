package common

import (
	"crypto/tls"
	"github.com/hellgate75/go-tcp-common/log"
)

type Commander interface {
	Execute(conn *tls.Conn) error
	SetLogger(logger log.Logger)
}
