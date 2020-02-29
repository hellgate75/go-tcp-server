package shell

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-server/common"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type shell struct{}

func existsFile(file string) bool {
	_, err1 := os.Stat(file)
	if err1 != nil {
		return false
	}
	return true
}

func loadFile(path string) ([]byte, error) {
	file, err1 := os.Open(path)
	if err1 != nil {
		return nil, err1
	}
	return ioutil.ReadAll(file)
}

var serverCommand string = "shell"

func (shell *shell) SendMessage(conn *tls.Conn, params ...interface{}) error {
	var paramsLen int = len(params)
	var interactive string = "true"
	if paramsLen > 0 {
		if "true" != fmt.Sprintf("%v", params[0]) {
			interactive = "false"
		}

	}

	var shellCommandOrScript string = ""
	var isScriptFile bool = false
	if paramsLen > 1 {
		if "" != fmt.Sprintf("%v", params[1]) {
			shellCommandOrScript = fmt.Sprintf("%v", params[1])
			isScriptFile = len(shellCommandOrScript) > 5 && strings.Index(shellCommandOrScript, ".") >= len(shellCommandOrScript)-5
			interactive = "false"
		}
	}
	fmt.Printf("Shell Script: %s, Is Script: %v\n", shellCommandOrScript, isScriptFile)
	n0, err3b := common.WriteString(serverCommand, conn)
	if err3b != nil {
		return err3b
	}
	if n0 == 0 {
		return errors.New(fmt.Sprintf("Unable to send command: %s", serverCommand))
	}
	time.Sleep(3 * time.Second)
	n1, err4 := common.WriteString(interactive, conn)
	if err4 != nil {
		return err4
	}
	if n1 == 0 {
		return errors.New(fmt.Sprintf("Unable to send interactive: %s", interactive))
	}
	time.Sleep(3 * time.Second)
	if "" != shellCommandOrScript {
		var script string = ""
		if isScriptFile {
			if !existsFile(shellCommandOrScript) {
				common.WriteString("exit", conn)
				return errors.New(fmt.Sprintf("Script File %s doesn't exists!!", shellCommandOrScript))
			}
			n2, err5 := common.WriteString("script", conn)
			if err5 != nil {
				common.WriteString("exit", conn)
				return err5
			}
			if n2 == 0 {
				common.WriteString("exit", conn)
				return errors.New(fmt.Sprintf("Unable to send script file type: %v", isScriptFile))
			}
			fileName := shellCommandOrScript
			if strings.Contains(shellCommandOrScript, "/") {
				listX := strings.Split(shellCommandOrScript, "/")
				fileName = listX[len(listX)-1]
			} else if strings.Contains(shellCommandOrScript, "\\") {
				listX := strings.Split(shellCommandOrScript, "\\")
				fileName = listX[len(listX)-1]
			}
			n2, err5 = common.WriteString(fileName, conn)
			if err5 != nil {
				common.WriteString("exit", conn)
				return err5
			}
			if n2 == 0 {
				common.WriteString("exit", conn)
				return errors.New(fmt.Sprintf("Unable to send script file type: %v", isScriptFile))
			}
			content, errReadScript := loadFile(shellCommandOrScript)
			if errReadScript != nil {
				common.WriteString("exit", conn)
				return errors.New(fmt.Sprintf("Cannot read script File %s -> Details: %s", shellCommandOrScript, errReadScript.Error()))
			}
			script = string(content)
		} else {
			n2, err5 := common.WriteString("command", conn)
			if err5 != nil {
				common.WriteString("exit", conn)
				return err5
			}
			if n2 == 0 {
				common.WriteString("exit", conn)
				return errors.New(fmt.Sprintf("Unable to send COMMAND -> script file type: %v", isScriptFile))
			}
			script = shellCommandOrScript
		}
		time.Sleep(3 * time.Second)
		n3, err6 := common.Write([]byte(script), conn)
		if err6 != nil {
			common.WriteString("exit", conn)
			return err6
		}
		if n3 == 0 {
			common.WriteString("exit", conn)
			return errors.New(fmt.Sprintf("Unable to send data -> shell command: %v", script))
		}
		time.Sleep(3 * time.Second)
		content, errAnswer := common.Read(conn)
		if errAnswer != nil {
			return errors.New(fmt.Sprintf("Receive data -> shell command: %v", script))
		}
		fmt.Printf("Answer: %s\n", string(content))
	} else {
		n2, err5 := common.WriteString("shell", conn)
		if err5 != nil {
			common.WriteString("exit", conn)
			fmt.Println("Error: exit shell!!")
			return err5
		}
		if n2 == 0 {
			common.WriteString("exit", conn)
			fmt.Println("Error: exit shell!!")
			return errors.New("Unable to send shell command")
		}
		time.Sleep(3 * time.Second)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			var currentCommand string = scanner.Text()
			if "exit" == strings.ToLower(currentCommand) {
				fmt.Println("Request: exit shell!!")
				break
			}
			n3, err6 := common.WriteString(currentCommand, conn)
			if err6 != nil {
				common.WriteString("exit", conn)
				fmt.Println("Error: exit shell!!")
				return err6
			}
			if n3 == 0 {
				common.WriteString("exit", conn)
				fmt.Println("Error: exit shell!!")
				return errors.New(fmt.Sprintf("Unable to send command ->  %v", currentCommand))
			}
			time.Sleep(3 * time.Second)
			content, errAnswer := common.Read(conn)
			if errAnswer != nil {
				common.WriteString("exit", conn)
				fmt.Println("Error: exit shell!!")
				return errAnswer
			}
			fmt.Println("Response: ", string(content))
		}

		if err := scanner.Err(); err != nil {
			common.WriteString("exit", conn)
			fmt.Println("Error: exit shell!!")
			return err
		}

	}
	common.WriteString("exit", conn)
	return nil
}
func (shell *shell) Helper() string {
	return "shell [interactive] [script file|command]\n  Parameters:\n    [interactive]      (optional) interactive shell[true/false] (default: true)\n    [script file]      (optional) full path of local script file\n    [command]          (optional) shell command\n"
}

func New() common.Sender {
	return &shell{}
}
