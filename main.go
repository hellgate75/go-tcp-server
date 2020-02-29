package main

import (
	"flag"
	"fmt"
	"github.com/hellgate75/go-deploy/log"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/server"
	"os"
	"strings"
	"time"
)

var Logger log.Logger = log.NewAppLogger("go-tcp-server", "INFO")

var certsStr = ""
var keysStr = ""
var host = ""
var port = ""
var verbosity string = ""
var fSet *flag.FlagSet

func init() {
	fSet = flag.NewFlagSet("go-tcp-server", flag.ExitOnError)
	fSet.StringVar(&certsStr, "certs", "certs/server.pem", "Comma separated pem server certificate list")
	fSet.StringVar(&keysStr, "keys", "certs/server.key", "Comma separated server certs keys list")
	fSet.StringVar(&host, "ip", common.DEFAULT_IP_ADDRESS, "Listening ip address")
	fSet.StringVar(&port, "port", common.DEFAULT_PORT, "Listening port")
	fSet.StringVar(&verbosity, "verbosity", "INFO", "Logger verbosity level [TRACE,DEBUG,INFO,ERROR,FATAL] ")
}

func main() {
	if errParse := fSet.Parse(os.Args[1:]); errParse != nil {
		Logger.Errorf("Error in arguments parse: %s", errParse.Error())
		fSet.Usage()
		os.Exit(1)
	}
	var args []string = os.Args
	for _, arg := range args {
		if "-h" == arg || "--help" == arg {
			fSet.Usage()
			os.Exit(0)
		}

	}
	if string(Logger.GetVerbosity()) != strings.ToUpper(verbosity) {
		Logger.Infof("Changing logger verbosity to: %s", strings.ToUpper(verbosity))
		Logger.SetVerbosity(log.VerbosityLevelFromString(strings.ToUpper(verbosity)))
	}
	var certs = strings.Split(certsStr, ",")
	var keys = strings.Split(keysStr, ",")
	var lenght int = len(certs)
	if lenght > len(keys) {
		lenght = len(keys)
	}
	var certsPair []common.CertificateKeyPair = make([]common.CertificateKeyPair, 0)
	for i := 0; i < lenght; i++ {
		certsPair = append(certsPair, common.CertificateKeyPair{
			Cert: certs[i],
			Key:  keys[i],
		})
	}
	Logger.Infof("Summary:\nIp: %s\nPort: %s\ncerts: %v\nkeys: %v\n", host, port, certs, keys)
	server := server.NewServer(certsPair, host, port)
	server.Start()
	time.Sleep(2 * time.Second)
	Logger.Infof("Running: %v", server.IsRunning())
	for server.IsRunning() {
		fmt.Print(".")
		time.Sleep(30 * time.Second)
	}
	Logger.Info("Exit!!")
}
