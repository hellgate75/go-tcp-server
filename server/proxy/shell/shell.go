package shell

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/log"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type shell struct{
	logger log.Logger
}

func getPathSeparator() string {
	if runtime.GOOS == "windows" {
		return "\\"
	}
	return string(os.PathSeparator)
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}
	return os.Getenv("HOME")
}

func isWindows() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return false
}

func execCommand(command string) (string, error) {
	color.Yellow.Printf("Execute command: %s\n", command)
	var cmd *exec.Cmd
	if isWindows() {
		//		fmt.Println("Command-Execute: Windows!!")
		cmd = exec.Command("cmd", "/C", command)
	} else {
		//		fmt.Println("Command-Execute: Linux!!")
		cmd = exec.Command("sh", "-c", command)
	}
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n", stdoutStderr), nil
}

func execLinuxCommand(command string) (string, error) {
	color.Yellow.Printf("Execute command: %s\n", command)
	cmd := exec.Command("sh", "-c", command)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n", stdoutStderr), nil
}

func (shell *shell) SetLogger(logger log.Logger) {
	shell.logger = logger
}

func (shell *shell) Execute(conn *tls.Conn) error {
	interactive, err1 := common.ReadString(conn)
	if err1 != nil {
		return err1
	}

	action, err2 := common.ReadString(conn)
	if err2 != nil {
		return err2
	}

	if "exit" == action {
		return errors.New("Shell interrupted from client")
	}

	if "script" == action {
		if "true" == interactive {
			var message string = "Cannot run SCRIPT interactive!!"
			return errors.New(message)
		}
		fileName, err3 := common.ReadString(conn)
		if err3 != nil {
			return err3
		}
		time.Sleep(2 * time.Second)
		data, err4 := common.Read(conn)
		if err4 != nil {
			return err4
		}
		folder, errD := ioutil.TempDir(userHomeDir(), "go_tcp_server_")
		if errD != nil {
			return errD
		}
		file := folder + getPathSeparator() + fileName

		errF := ioutil.WriteFile(file, data, 0777)
		if errF != nil {
			return errF
		}
		time.Sleep(2 * time.Second)
		common.WriteString("ok:continue shell", conn)
		var output string
		var errExec error
		if strings.Contains(strings.ToLower(file), ".exe") ||
			strings.Contains(strings.ToLower(file), ".bat") ||
			strings.Contains(strings.ToLower(file), ".ps") ||
			strings.Contains(strings.ToLower(file), ".cmd") {
			output, errExec = execCommand(file)
		} else {
			output, errExec = execLinuxCommand(file)
		}
		os.Remove(file)
		os.Remove(folder)
		if errExec != nil {
			common.Write([]byte(output), conn)
			return errExec
		}
		common.Write([]byte(output), conn)
	} else if "command" == action {
		if "true" == interactive {
			var message string = "Cannot run COMMAND interactive!!"
			return errors.New(message)
		}
		data, err3 := common.Read(conn)
		if err3 != nil {
			return err3
		}
		time.Sleep(2 * time.Second)
		common.WriteString("ok:continue shell", conn)
		output, errExec := execCommand(string(data))
		if errExec != nil {
			var message string = "shell:command (cmd:"+string(data)+") ::exec->"+errExec.Error()
			common.Write([]byte(message), conn)
			if shell.logger != nil {
				shell.logger.Errorf("Error excuting command: %s, Details: %s", string(data), errExec.Error())
			} else {
				color.LightRed.Printf("Error excuting command: %s, Details: %s", string(data), errExec.Error())
			}
			return errExec
		}
		common.Write([]byte(output), conn)
	} else if "shell" == action {
		if "false" == interactive {
			var message string = "Cannot run SHELL non interactive!!"
			return errors.New(message)
		}
		var command string = ""
		var err3 error
		for "exit" != strings.ToLower(command) && err3 == nil {
			command, err3 = common.ReadString(conn)
			if err3 == nil && "exit" != strings.ToLower(command) {
				if "" == command {
					continue
				}
				if shell.logger != nil {
					shell.logger.Debugf("Command: <%s>", command)
				} else {
					color.LightYellow.Sprintf("Command: <%s>", command)
				}
				var output string
				output, err3 = execCommand(command)
				//time.Sleep(2*time.Second)
				if err3 == nil {
					common.Write([]byte(output), conn)
				} else {
					common.Write([]byte(err3.Error()), conn)
				}
			} else if err3 != nil {
				if shell.logger != nil {
					shell.logger.Errorf("Command <%s> Error: %s\n", command, err3.Error())
				} else {
					color.LightRed.Printf("Command <%s> Error: %s\n", command, err3.Error())
				}
			}
		}
		if err3 != nil {
			return err3
		}
		common.WriteString("ok", conn)
	} else {
		return errors.New("Invalid action: " + action)
	}
	return nil
}

func New() common.Commander {
	return &shell{}
}
