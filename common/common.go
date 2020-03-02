package common

import ()

const (
	DEFAULT_IP_ADDRESS        string = "0.0.0.0"
	DEFAULT_CLIENT_IP_ADDRESS string = "127.0.0.1"
	DEFAULT_PORT              string = "49022"
)

type CertificateKeyPair struct {
	Cert string
	Key  string
}

type TCPServer interface {
	Start() error

	IsRunning() bool

	Stop()
}
