package main

import (
	"flag"
	"github.com/hellgate75/go-tcp-common/log"
	commonnet "github.com/hellgate75/go-tcp-common/net"
	restcomm "github.com/hellgate75/go-tcp-common/net/rest/common"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-server/server"
	"github.com/hellgate75/go-tcp-server/server/proxy"
	"os"
	"strconv"
	"strings"
	"time"
	//"fmt"
)

var Logger log.Logger = log.NewLogger("go-tcp-server", "INFO")

var caCert string = ""
var insecure bool = false
var certsStr string = ""
var keysStr string = ""
var host string = ""
var port string = ""
var verbosity string = ""
var requiresChiphers string = "true"
var readTimeout int64 = 0
var fSet *flag.FlagSet

func init() {
	fSet = flag.NewFlagSet("go-tcp-server", flag.ExitOnError)
	fSet.StringVar(&caCert, "ca-cert", "certs/ca.crt", "Ca Certificate file path")
	fSet.BoolVar(&insecure, "inssecure-keys", false, "USe insecure keys acceptance")
	fSet.StringVar(&certsStr, "certs", "certs/server.pem", "Comma separated pem server certificate files list")
	fSet.StringVar(&keysStr, "keys", "certs/server.key", "Comma separated server certs keys files list")
	fSet.StringVar(&host, "ip", common.DEFAULT_IP_ADDRESS, "Listening ip address")
	fSet.StringVar(&port, "port", common.DEFAULT_PORT, "Listening port")
	fSet.StringVar(&verbosity, "verbosity", "INFO", "Logger verbosity level [TRACE,DEBUG,INFO,ERROR,FATAL] ")
	fSet.StringVar(&requiresChiphers, "requires-chiphers", "true", "Requires Chiphers and Cuerves algorithms (true|false)")
	fSet.Int64Var(&readTimeout, "read-timeout", 5, "Message Read timeout in seconds, used to keep listening for answer from clients")
	fSet.StringVar(&proxy.PluginLibrariesFolder, "plugins-folder", proxy.PluginLibrariesFolder, "Folder where seek for plugin(s) library [Linux Only]")
	fSet.StringVar(&proxy.PluginLibrariesExtension, "plugins-extension", proxy.PluginLibrariesExtension, "File extension for plugin libraries [Linux Only]")
	fSet.BoolVar(&proxy.UsePlugins, "use-plugins", proxy.UsePlugins, "Enable/disable plugins [true|false] [Linux Only]")
	server.Logger = Logger
}

func main() {
	var args []string = os.Args
	for _, arg := range args {
		if "-h" == arg || "--help" == arg {
			fSet.Usage()
			os.Exit(0)
		}

	}
	if errParse := fSet.Parse(os.Args[1:]); errParse != nil {
		Logger.Errorf("Error in arguments parse: %s", errParse.Error())
		fSet.Usage()
		os.Exit(1)
	}
	commonnet.DEFAULT_TIMEOUT = time.Duration(readTimeout) * time.Second
	if string(Logger.GetVerbosity()) != strings.ToUpper(verbosity) {
		Logger.Debugf("Changing logger verbosity to: %s", strings.ToUpper(verbosity))
		Logger.SetVerbosity(log.VerbosityLevelFromString(strings.ToUpper(verbosity)))
		server.Logger.SetVerbosity(log.VerbosityLevelFromString(strings.ToUpper(verbosity)))
	}
	var certs = strings.Split(certsStr, ",")
	var keys = strings.Split(keysStr, ",")
	var lenght int = len(certs)
	if lenght > len(keys) {
		lenght = len(keys)
	}
	var certsPair []restcomm.CertificateKeyPair = make([]restcomm.CertificateKeyPair, 0)
	for i := 0; i < lenght; i++ {
		certsPair = append(certsPair, restcomm.CertificateKeyPair{
			Cert: certs[i],
			Key:  keys[i],
		})
	}
	Logger.Debugf("Summary:\nIp: %s\nPort: %s\ncerts: %v\nkeys: %v\ncaCert: %s\ninsecure: %v\n", host, port, certs, keys, caCert, insecure)
	server := server.NewServer()
	var portInt int
	portInt, _ = strconv.Atoi(port)
	if caCert != "" {
		insecure = true
	}
	defer func(){
		time.Sleep(5 * time.Second)
		server.Stop()
	}()
	errStart := server.StartTLS(host, int32(portInt), certsPair, caCert, insecure)
	if errStart != nil {
		Logger.Errorf("Server start-up error: %s\n", errStart.Error())
		panic(errStart.Error())

	}
	time.Sleep(2 * time.Second)
	Logger.Debugf("Running: %v", server.IsRunning())
	for server.IsRunning() {
//		fmt.Print(".")
		time.Sleep(30 * time.Second)
	}
	Logger.Debugf("Exit!!")
}
