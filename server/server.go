package server

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/log"
	"github.com/hellgate75/go-tcp-server/server/proxy"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var Logger log.Logger = nil

type tcpServer struct {
	Certs                     []common.CertificateKeyPair
	IpAddress                 string
	Port                      string
	RequiresChiphersAndCurves bool
	running                   bool
	conn                      []*net.Conn
	tlscon                    []*tls.Conn
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

var currentListener net.Listener

func (server *tcpServer) Start() error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			Logger.Errorf("Errors: %v", r)
			Logger.Fatal("TCP Server exit ...")
			os.Exit(0)
		}
	}()

	var certificates []tls.Certificate = make([]tls.Certificate, 0)

	for _, keyPair := range server.Certs {
		cert, err := tls.LoadX509KeyPair(keyPair.Cert, keyPair.Key)

		if err != nil {
			Logger.Fatalf("server: loadkeys: %s", err)
			panic("server: loadkeys:" + err.Error())
		}
		certificates = append(certificates, cert)
	}
	var config *tls.Config
	if !server.RequiresChiphersAndCurves {
		Logger.Warn("No Chiphers and TLS Curves required...")
		config = &tls.Config{
			Certificates: certificates,
		}

	} else {
		config = &tls.Config{
			Certificates:             certificates,
			MinVersion:               tls.VersionTLS10,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}
	}
	config.Rand = rand.Reader
	if server.IpAddress == "" {
		server.IpAddress = common.DEFAULT_IP_ADDRESS
	}
	if server.Port == "" {
		server.Port = common.DEFAULT_PORT
	}
	service := fmt.Sprintf("%s:%s", server.IpAddress, server.Port)
	listener, err := tls.Listen("tcp", service, config)
	if err != nil {
		Logger.Fatalf("server: listen: %v", err)
		if listener != nil {
			listener.Close()
		}
		panic("server: listen: " + err.Error())
	}
	currentListener = listener
	Logger.Infof("server: listen: %v", service)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprintf("%v", r))
				Logger.Fatalf("TCP Server exit ...")
			}
			Logger.Info("TCP Server exit ...")
		}()
		server.running = true
		for server.running {
			conn, errN := listener.Accept()
			if errN != nil {
				Logger.Errorf("server: accept: %s", errN)
				continue
			}
			server.conn = append(server.conn, &conn)
			defer conn.Close()
			Logger.Debugf("server: accepted from %s", conn.RemoteAddr())
			tlscon, ok := conn.(*tls.Conn)
			if ok {
				Logger.Debug("ok=true")
				state := tlscon.ConnectionState()
				for _, v := range state.PeerCertificates {
					Logger.Debug(x509.MarshalPKIXPublicKey(v.PublicKey))
				}
			}
			server.tlscon = append(server.tlscon, tlscon)
			go handleClient(tlscon, server)
		}
	}()
	return err
}

func handleClient(conn *tls.Conn, server *tcpServer) {
	defer func() {
		if r := recover(); r != nil {
			Logger.Errorf("Errors handling client request: %v", r)
			Logger.Error("Client connection error ...")
		}
		conn.Close()
		Logger.Info("Client connection exit ...")
	}()
	var buffSize int = 2048
	var open bool = true
	for open {
		str, errRead := common.ReadStringBuffer(buffSize, conn)
		if errRead != nil {
			Logger.Infof("server: conn: read error: %s", errRead)
			open = false
			return
		}
		var command string = str
		if command == "" {
			Logger.Debug("No command received ...")
			continue
		}
		Logger.Infof("Received command: <%s>", command)
		if "exit" == strings.ToLower(command) {
			open = false
			Logger.Info("Client exit ...")
			break
		} else if "shutdown" == strings.ToLower(command) {
			open = false
			Logger.Info("Shutdown server ...")
			common.WriteString("ok", conn)
			conn.Close()
			os.Exit(0)
		} else if "restart" == strings.ToLower(command) {
			open = false
			Logger.Info("Restarting server ...")
			executable, errExec := os.Executable()
			if errExec != nil {
				var message string = fmt.Sprintf("Error recovering executables -> Details: ", errExec.Error())
				Logger.Error(message)
				common.WriteString("ko:restart:"+message, conn)
				conn.Close()
				return
			}
			var execCmdText string
			Logger.Warnf("Discovered executables: %s, args: %v", executable, os.Args[1:])
			var cmdExecutor []string = make([]string, 0)
			if runtime.GOOS == "windows" {
				execCmdText = "cmd"
				cmdExecutor = append(cmdExecutor, "/C")
			} else {
				execCmdText = "sh"
				cmdExecutor = append(cmdExecutor, "-c")
			}

			cmdExecutor = append(cmdExecutor, executable)
			if len(os.Args) > 1 {
				cmdExecutor = append(cmdExecutor, os.Args[1:]...)
			}
			var cmd *exec.Cmd = exec.Command(execCmdText, cmdExecutor...)
			var path string
			path, errWd := os.Getwd()
			if errWd != nil {
				path = filepath.Dir(executable)
			}
			Logger.Warnf("Working Dir: %s", path)
			cmd.Dir = path
			Logger.Warnf("Command: %s", cmd.String())
			go func() {
				stdoutStderr, errCmd := cmd.CombinedOutput()
				Logger.Warnf("errCmd: %v", errCmd)
				Logger.Warnf("stdoutStderr: %v", stdoutStderr)
				if errCmd != nil {
					var message string = fmt.Sprintf("Error runninf executables -> Details: ", errCmd.Error())
					Logger.Error(message)
					return
				}
				Logger.Warnf("Executables: %s, ran successfully", executable)
			}()
			common.WriteString("ok", conn)
			conn.Close()
			server.Stop()
			currentListener.Close()
			time.Sleep(10 * time.Second)
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
		} else if command == "os-name" {
			time.Sleep(2 * time.Second)
			Logger.Infof("Sending OS type %s to client ...", runtime.GOOS)
			_, errWelcome := common.WriteString(string(runtime.GOOS), conn)
			if errWelcome != nil {
				Logger.Error("Error sending welcome message")
				continue
			}
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
			commander.SetLogger(Logger)
			errCom := commander.Execute(conn)
			if errCom != nil {
				var message string = "ko:command:"+command+"->"+errCom.Error()
				common.WriteString(message, conn)
			} else {
				common.WriteString("ok", conn)
			}
		}
	}
	Logger.Info("server: conn: closed")
}

func NewServer(certs []common.CertificateKeyPair, ipAddress string, port string, requiresChiphers bool) common.TCPServer {
	return &tcpServer{
		Certs:                     certs,
		IpAddress:                 ipAddress,
		Port:                      port,
		RequiresChiphersAndCurves: requiresChiphers,
		running:                   false,
		conn:                      make([]*net.Conn, 0),
		tlscon:                    make([]*tls.Conn, 0),
	}
}
