package main

import (
	"flag"
	"github.com/hellgate75/go-deploy/log"
	"github.com/hellgate75/go-tcp-server/client/worker"
	"github.com/hellgate75/go-tcp-server/common"
	"os"
	"strings"
	"time"
)

var Logger log.Logger = log.NewAppLogger("go-tcp-client", "INFO")

var certs string = ""
var keys string = ""
var host string = ""
var port string = ""
var verbosity string = ""
var fSet *flag.FlagSet

func init() {
	fSet = flag.NewFlagSet("go-tcp-client", flag.ContinueOnError)
	fSet.StringVar(&certs, "certs", "certs/server.pem", "Comma separated pem server certificate list")
	fSet.StringVar(&keys, "keys", "certs/server.key", "Comma separated server certs keys list")
	fSet.StringVar(&host, "ip", common.DEFAULT_CLIENT_IP_ADDRESS, "Server ip address")
	fSet.StringVar(&port, "port", common.DEFAULT_PORT, "Server port")
	fSet.StringVar(&verbosity, "verbosity", "INFO", "Logger verbosity level [TRACE,DEBUG,INFO,ERROR,FATAL] ")
}

func main() {
	if errParse := fSet.Parse(os.Args[1:]); errParse != nil {
		Logger.Errorf("Error in arguments parse: %s", errParse.Error())
		fSet.Usage()
		os.Exit(1)
	}
	var commands []string = make([]string, 0)
	var args []string = os.Args[1:]
	var hasToken bool = false
	for _, arg := range args {
		if "-h" == arg || "--help" == arg {
			fSet.Usage()
			os.Exit(0)
		}
		if "-" == arg[0:1] {
			hasToken = true
		} else if !hasToken {
			commands = append(commands, arg)
		} else {
			hasToken = false
		}

	}
	Logger.Infof("new verbosity: <%s>", strings.ToUpper(verbosity))
	Logger.Infof("logger verbosity: %v", Logger.GetVerbosity())
	if string(Logger.GetVerbosity()) != strings.ToUpper(verbosity) {
		Logger.Infof("Changing logger verbosity to: %s", strings.ToUpper(verbosity))
		Logger.SetVerbosity(log.VerbosityLevelFromString(strings.ToUpper(verbosity)))
	}
	var lenght int = len(certs)
	if lenght > len(keys) {
		lenght = len(keys)
	}
	var certPair common.CertificateKeyPair = common.CertificateKeyPair{
		Cert: certs,
		Key:  keys,
	}
	client := worker.NewClient(certPair, host, port)
	if len(commands) > 0 {
		var cmd string = commands[0]

		if strings.ToLower(cmd) == "help" ||
			strings.ToLower(cmd) == "--help" ||
			strings.ToLower(cmd) == "-h" {
			list := client.GetHelp()
			Logger.Info("List of commands:")
			for _, item := range list {
				Logger.Printf("- %s", item)
			}
			return

		}
		Logger.Infof("Summary:\nIp: %s\nPort: %s\ncerts: %v\nkeys: %v\n", host, port, certs, keys)
		client.Open(true)
		defer client.Close()
		var commandArgs []string = commands[1:]
		Logger.Infof("Command Args: (len: %v) %v", len(commandArgs), commandArgs)
		var params []interface{} = make([]interface{}, 0)
		for _, val := range commandArgs {
			params = append(params, val)
		}
		Logger.Debugf("Params: (len: %v) %v", len(params), params)
		err1 := client.ApplyCommand(cmd, params...)
		if err1 != nil {
			Logger.Errorf("Error sending command %s, Details: %s", cmd, err1.Error())
			exitClient(client)
			return
		}
		time.Sleep(3 * time.Second)
		answer, err2 := client.ReadAnswer()
		if err2 != nil {
			Logger.Errorf("Error reading respnse for command %s, Details: %s", cmd, err2.Error())
			exitClient(client)
			return
		}
		if "ok" == answer {
			Logger.Infof("Command Message '%s' sent and executed successfully!!", cmd)
		} else {
			Logger.Error("Command Message '%s' sent but failed!!")

		}
		Logger.Debugf("Response: %v", answer)
	}
	exitClient(client)
}

func exitClient(client common.TCPClient) {
	client.SendText("exit")
	time.Sleep(2 * time.Second)
	Logger.Info("Exit!!")
}
