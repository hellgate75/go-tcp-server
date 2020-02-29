package server

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hellgate75/go-deploy/log"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/server/proxy"
	"net"
	"os"
	"strconv"
	"strings"
)

var Logger log.Logger = log.NewAppLogger("go-tcp-server", "INFO")

type tcpServer struct {
	Certs     []common.CertificateKeyPair
	IpAddress string
	Port      string
	running   bool
	conn      []*net.Conn
	tlscon    []*tls.Conn
}

func (server *tcpServer) IsRunning() bool {
	return server.running
}

func (server *tcpServer) Stop() {
	server.running = false
	for _, conn := range server.tlscon {
		if conn != nil {
			(*conn).CloseWrite()
		}
	}
	for _, conn := range server.conn {
		if conn != nil {
			(*conn).Close()
		}
	}
}

func (server *tcpServer) Start() error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	var certificates []tls.Certificate = make([]tls.Certificate, 0)

	for _, keyPair := range server.Certs {
		//cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
		cert, err := tls.LoadX509KeyPair(keyPair.Cert, keyPair.Key)

		if err != nil {
			Logger.Fatal(fmt.Sprintf("server: loadkeys: %s", err))
			panic("server: loadkeys:" + err.Error())
		}
		certificates = append(certificates, cert)
	}
	config := tls.Config{Certificates: certificates}
	config.Rand = rand.Reader
	if server.IpAddress == "" {
		server.IpAddress = common.DEFAULT_IP_ADDRESS
	}
	if server.Port == "" {
		server.Port = common.DEFAULT_PORT
	}
	service := fmt.Sprintf("%s:%s", server.IpAddress, server.Port)
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		Logger.Fatal(fmt.Sprintf("server: listen: %s", err))
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprintf("%v", r))
			}
		}()
		server.running = true
		for server.running {
			conn, errN := listener.Accept()
			if errN != nil {
				Logger.Error(fmt.Sprintf("server: accept: %s", errN))
				continue
			}
			server.conn = append(server.conn, &conn)
			defer conn.Close()
			Logger.Info(fmt.Sprintf("server: accepted from %s", conn.RemoteAddr()))
			tlscon, ok := conn.(*tls.Conn)
			if ok {
				Logger.Info("ok=true")
				state := tlscon.ConnectionState()
				for _, v := range state.PeerCertificates {
					Logger.Info(x509.MarshalPKIXPublicKey(v.PublicKey))
				}
			}
			server.tlscon = append(server.tlscon, tlscon)
			go handleClient(tlscon)
		}
	}()
	return err
}

func handleClient(conn *tls.Conn) {
	defer conn.Close()
	var buffSize int = 2048
	var open bool = true
	for open {
		str, errRead := common.ReadStringBuffer(buffSize, conn)
		if errRead != nil {
			Logger.Info(fmt.Sprintf("server: conn: read error: %s", errRead))
			open = false
			return
		}
		var command string = str
		if command == "" {
			continue
		}
		Logger.Info("server: conn: compute read")
		Logger.Infof("Received command: <%s>", command)
		if "exit" == strings.ToLower(command) {
			open = false
			Logger.Info("Client exit ...")
			break
		} else if "shutdown" == strings.ToLower(command) {
			open = false
			Logger.Info("Shutdown server ...")
			conn.Close()
			os.Exit(0)
		} else if len(command) > 12 && "buffer-size:" == strings.ToLower(command[:12]) {
			list := strings.Split(command, ":")
			size, errAtoi := strconv.Atoi(list[1])
			if errAtoi != nil {
				Logger.Errorf("Errors converting buffer size to: <%s> -> %s", list[1], errAtoi.Error())
				continue
			}
			Logger.Infof("Changing buffer size to: %v", size)
			buffSize = size
		} else {
			commander, errP := proxy.GetCommander(command)
			if errP != nil {
				var message string = fmt.Sprintf("Error to find command: <%s> -> %s", command, errP.Error())
				Logger.Error(message)
				common.WriteString("ko:"+message, conn)
				continue
			}
			if commander == nil {
				var message string = fmt.Sprintf("Error to reference command: <%s> !!", command)
				Logger.Error(message)
				common.WriteString("ko:"+message, conn)
				continue
			}
			errCom := commander.Execute(conn)
			if errCom != nil {
				common.WriteString("ko:command:"+command+"->"+errCom.Error(), conn)
			} else {
				common.WriteString("ok", conn)
			}
		}
	}
	Logger.Info("server: conn: closed")
}

func NewServer(certs []common.CertificateKeyPair, ipAddress string, port string) common.TCPServer {
	return &tcpServer{
		Certs:     certs,
		IpAddress: ipAddress,
		Port:      port,
		running:   false,
		conn:      make([]*net.Conn, 0),
		tlscon:    make([]*tls.Conn, 0),
	}
}
