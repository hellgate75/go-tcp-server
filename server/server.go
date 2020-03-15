package server

import (
	"crypto/tls"
	"fmt"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-server/server/proxy"
	restcomm "github.com/hellgate75/go-tcp-common/net/rest/common"
	restsrv "github.com/hellgate75/go-tcp-common/net/rest/tls/server"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var Logger log.Logger = log.NewLogger("go-tcp-server", "INFO")

func handleClient(conn *tls.Conn, server restcomm.RestServer) {
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
			//currentListener.Close()
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

func NewServer() restcomm.RestServer {
	return restsrv.NewHandleFunc(restsrv.TLSHandleFunc(handleClient), Logger)
}