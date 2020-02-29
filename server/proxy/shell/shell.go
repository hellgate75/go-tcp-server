package shell

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-server/common"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type shell struct{}

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
	var cmd *exec.Cmd
	if isWindows() {
		fmt.Println("Command-Execute: Windows!!")
		cmd = exec.Command("cmd", "/C", command)
	} else {
		fmt.Println("Command-Execute: Linux!!")
		cmd = exec.Command("sh", "-c", command)
	}
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n", stdoutStderr), nil
}

func execLinuxCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n", stdoutStderr), nil
}

func (shell *shell) Execute(conn *tls.Conn) error {
	time.Sleep(2 * time.Second)
	interactive, err1 := common.ReadString(conn)
	if err1 != nil {
		return err1
	}

	time.Sleep(2 * time.Second)
	action, err2 := common.ReadString(conn)
	if err2 != nil {
		return err2
	}

	if "exit" == action {
		return errors.New("Shell interrupted from client")
	}

	time.Sleep(2 * time.Second)
	if "script" == action {
		if "true" == interactive {
			var message string = "Cannot run SCRIPT interactive!!"
			common.WriteString("ko:shell:script->"+message, conn)
			return errors.New(message)
		}
		fileName, err3 := common.ReadString(conn)
		if err3 != nil {
			common.WriteString("ko:shell:script->"+err3.Error(), conn)
			return err3
		}
		time.Sleep(2 * time.Second)
		data, err4 := common.Read(conn)
		if err4 != nil {
			common.WriteString("ko:shell:script->"+err4.Error(), conn)
			return err4
		}
		folder, errD := ioutil.TempDir(userHomeDir(), "go_tcp_server_")
		if errD != nil {
			common.WriteString("ko:shell:script->"+errD.Error(), conn)
			return errD
		}
		file := folder + getPathSeparator() + fileName

		errF := ioutil.WriteFile(file, data, 0777)
		if errF != nil {
			common.WriteString("ko:shell:script->"+errF.Error(), conn)
			return errF
		}
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
		if errExec != nil {
			common.WriteString("ko:shell:script::exec->"+errExec.Error(), conn)
			return errExec
		}
		common.Write([]byte(output), conn)
		common.WriteString("ok", conn)
	} else if "command" == action {
		if "true" == interactive {
			var message string = "Cannot run COMMAND interactive!!"
			common.WriteString("ko:shell:command->"+message, conn)
			return errors.New(message)
		}
		data, err3 := common.Read(conn)
		if err3 != nil {
			common.WriteString("ko:shell:command->"+err3.Error(), conn)
			return err3
		}
		fmt.Printf("Command Execute: %s\n", string(data))
		output, errExec := execCommand(string(data))
		if errExec != nil {
			common.WriteString("ko:shell:command::exec->"+errExec.Error(), conn)
			return errExec
		}
		common.Write([]byte(output), conn)
		common.WriteString("ok", conn)
	} else if "shell" == action {
		if "false" == interactive {
			var message string = "Cannot run SHELL non interactive!!"
			common.WriteString("ko:shell:shell->"+message, conn)
			return errors.New(message)
		}
		var command string = ""
		var err3 error
		for "exit" != command && err3 == nil {
			command, err3 = common.ReadString(conn)
			if err3 == nil {
				var output string
				output, err3 = execCommand(command)
				if err3 == nil {
					common.Write([]byte(output), conn)
					time.Sleep(2 * time.Second)
				}
			}
		}
		if err3 != nil {
			common.WriteString("ko:shell:shell->"+err3.Error(), conn)
			return err3
		}
		common.WriteString("ok", conn)
	} else {
		common.WriteString("ko:shell:all->Invalid action: "+action, conn)
		return errors.New("Invalid action: " + action)
	}
	return nil
}

func New() common.Commander {
	return &shell{}
}