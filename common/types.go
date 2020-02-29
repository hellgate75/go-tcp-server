package common

import (
	"crypto/tls"
)

type Commander interface {
	Execute(conn *tls.Conn) error
}

type Sender interface {
	SendMessage(conn *tls.Conn, params ...interface{}) error
	Helper() string
}
