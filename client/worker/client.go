package worker

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hellgate75/go-deploy/log"
	"github.com/hellgate75/go-tcp-server/client/proxy"
	"github.com/hellgate75/go-tcp-server/common"
)

var Logger log.Logger = log.NewAppLogger("go-tcp-client", "INFO")

type tcpClient struct {
	Cert      common.CertificateKeyPair
	IpAddress string
	Port      string
	conn      *tls.Conn
}

func (tcpClient *tcpClient) Open(insecureSkipVerify bool) error {
	cert, err := tls.LoadX509KeyPair(tcpClient.Cert.Cert, tcpClient.Cert.Key)
	if err != nil {
		Logger.Fatalf("server: loadkeys: %s", err)
		return errors.New(fmt.Sprintf("server: loadkeys: %s", err))
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: insecureSkipVerify}
	service := fmt.Sprintf("%s:%s", tcpClient.IpAddress, tcpClient.Port)
	conn, err := tls.Dial("tcp", service, &config)
	if err != nil {
		Logger.Fatalf("client: dial: %s", err)
		return errors.New(fmt.Sprintf("client: dial: %s", err))
	}
	tcpClient.conn = conn
	Logger.Info("client: connected to: ", conn.RemoteAddr())

	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		Logger.Info(x509.MarshalPKIXPublicKey(v.PublicKey))
		Logger.Info(v.Subject)
	}
	Logger.Info("client: handshake: ", state.HandshakeComplete)
	Logger.Info("client: mutual: ", state.NegotiatedProtocolIsMutual)
	return nil
}

func (tcpClient *tcpClient) IsOpen() bool {
	return tcpClient.conn != nil
}

func (tcpClient *tcpClient) Send(message bytes.Buffer) error {
	n, err := common.Write(message.Bytes(), tcpClient.conn)
	if err != nil {
		Logger.Errorf("client: write: %s", err.Error())
		return errors.New(fmt.Sprintf("client: write: %s", err.Error()))
	}
	Logger.Infof("client: wrote %s (wrote: %d bytes)", message, n)
	if n == 0 {
		return errors.New(fmt.Sprintf("client: written bytes: %d", n))
	}
	return nil
}

func (tcpClient *tcpClient) SendText(message string) error {
	n, err := common.WriteString(message, tcpClient.conn)
	if err != nil {
		Logger.Errorf("client: write: %s", err.Error())
		return errors.New(fmt.Sprintf("client: write: %s", err.Error()))
	}
	Logger.Infof("client: wrote %q (wrote: %d bytes)", message, n)
	if n == 0 {
		return errors.New(fmt.Sprintf("client: written bytes: %d", n))
	}
	return nil
}

func (tcpClient *tcpClient) ApplyCommand(command string, params ...interface{}) error {
	sender, err := proxy.GetSender(command)
	if err != nil {
		Logger.Errorf("client: apply command: %s", err.Error())
		return errors.New(fmt.Sprintf("client: write: %s", err.Error()))
	}
	err = sender.SendMessage(tcpClient.conn, params...)
	if err != nil {
		Logger.Errorf("client: command (%s): %s", command, err.Error())
		return errors.New(fmt.Sprintf("client: command (%s): %s", command, err.Error()))
	}
	return nil
}

func (tcpClient *tcpClient) GetHelp() []string {
	return proxy.Help()
}

func (tcpClient *tcpClient) Close() {
	if tcpClient.conn != nil {
		tcpClient.conn.Close()
		tcpClient.conn = nil
	}

}

func NewClient(cert common.CertificateKeyPair, ipAddress string, port string) common.TCPClient {
	return &tcpClient{
		Cert:      cert,
		IpAddress: ipAddress,
		Port:      port,
	}
}

func (tcpClient *tcpClient) ReadAnswer() (string, error) {
	return common.ReadString(tcpClient.conn)
}

func (tcpClient *tcpClient) ReadDataPack() ([]byte, error) {
	return common.Read(tcpClient.conn)
}
