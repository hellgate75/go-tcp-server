package worker

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-server/client/proxy"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/log"
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
	Logger.Debugf("client: connected to: %v", conn.RemoteAddr())

	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		Logger.Debug(x509.MarshalPKIXPublicKey(v.PublicKey))
		Logger.Debug(v.Subject)
	}
	Logger.Debug("client: handshake: ", state.HandshakeComplete)
	Logger.Debug("client: mutual: ", state.NegotiatedProtocolIsMutual)
	Logger.Info("client: Connected!!")
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
	Logger.Debugf("client: wrote %s (wrote: %d bytes)", message, n)
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
	Logger.Debugf("client: wrote %q (wrote: %d bytes)", message, n)
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

func (tc *tcpClient) Clone() common.TCPClient {
	return &tcpClient{
		Cert:      tc.Cert,
		IpAddress: tc.IpAddress,
		Port:      tc.Port,
	}
}

func (tcpClient *tcpClient) Close() error {
	if tcpClient.conn != nil {
		tcpClient.SendText("exit")
		tcpClient.conn.Close()
		tcpClient.conn = nil
	}
	return nil
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
