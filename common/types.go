package common

import (
	"crypto/tls"
)

type Commander interface {
	Execute(conn *tls.Conn) error
}
