package proxy

import (
	"fmt"
	"github.com/hellgate75/go-tcp-modules/server/proxy"
	"github.com/hellgate75/go-tcp-server/common"
	"github.com/hellgate75/go-tcp-common/log"
    "os"
    "path/filepath"
	"plugin"
)

var Logger log.Logger = nil

var UsePlugins bool = false
var PluginLibrariesFolder string = getDefaultPluginsFolder()
var PluginLibrariesExtension = "so"

func GetCommander(command string) (common.Commander, error) {
	if UsePlugins {
		fullPath := fmt.Sprintf("%s%s%s.%s", PluginLibrariesFolder, string(os.PathSeparator), command, PluginLibrariesExtension)
		Logger.Debugf("server.proxy.GetCommander() -> Loading library: %s", fullPath)
		plugin, err := plugin.Open(fullPath)
		if err == nil {
			sym, err2 := plugin.Lookup("GetCommander")
			if err2 != nil {
				commander, errPlugin := sym.(func(string)(common.Commander, error))(command)
				if errPlugin != nil {
					return nil, errPlugin
				}
				commander.SetLogger(Logger)
				return commander, nil
			}
		}
	}
	commander, errOrig := proxy.GetCommander(command)
	if errOrig != nil {
		return nil, errOrig
	}
	commander.SetLogger(Logger)
	return commander, nil
}

func getDefaultPluginsFolder() string {
    execPath, err := os.Executable()
    if err != nil {
    	pwd, errPwd := os.Getwd()
    	if errPwd != nil {
    		return filepath.Dir(".") + string(os.PathSeparator) + "modules"
		}
		return filepath.Dir(pwd) + string(os.PathSeparator) + "modules"
	}
    return filepath.Dir(execPath) + string(os.PathSeparator) + "modules"
}