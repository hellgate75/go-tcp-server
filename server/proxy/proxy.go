package proxy

import (
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-modules/server/proxy"
	"github.com/hellgate75/go-tcp-server/common"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

var Logger log.Logger = nil

var UsePlugins bool = false
var PluginLibrariesFolder string = getDefaultPluginsFolder()
var PluginLibrariesExtension = "so"

func GetCommander(command string) (common.Commander, error) {
	if UsePlugins {
		Logger.Debugf("client.proxy.GetSender() -> Loading library for command: %s", command)
		var commander common.Commander = nil
		forEachSenderInPlugins(command, func(commandersList []common.Commander) {
			if len(commandersList) > 0 {
				commander = commandersList[0]
			}
		})
		if commander != nil {
			return commander, nil
		}
	}
	commander, errOrig := proxy.GetCommander(command)
	if errOrig != nil {
		return nil, errOrig
	}
	commander.SetLogger(Logger)
	return commander, nil
}


func filterByExtension(fileName string) bool {
	n := len(PluginLibrariesExtension)
	fileNameLen := len(fileName)
	posix := fileNameLen - n
	return posix > 0 && strings.ToLower(fileName[posix:]) == strings.ToLower("." + PluginLibrariesExtension)
}

func listLibrariesInFolder(dirName string) []string {
	var out []string = make([]string, 0)
	_, err0 := os.Stat(dirName)
	if err0 == nil {
		lst, err1 := ioutil.ReadDir(dirName)
		if err1 == nil {
			for _,file := range lst {
				if file.IsDir() {
					fullDirPath := dirName + string(os.PathSeparator) + file.Name()
					newList := listLibrariesInFolder(fullDirPath)
					out = append(out, newList...)
				} else {
					if filterByExtension(file.Name()) {
						fullFilePath := dirName + string(os.PathSeparator) + file.Name()
						out = append(out, fullFilePath)

					}
				}
			}
		}
	}
	return out
}

func forEachSenderInPlugins(command string, callback func([]common.Commander)())  {
	var senders []common.Commander = make([]common.Commander, 0)
	dirName := PluginLibrariesFolder
	_, err0 := os.Stat(dirName)
	if err0 == nil {
		libraries := listLibrariesInFolder(dirName)
		for _,libraryFullPath := range libraries {
			Logger.Debugf("client.proxy.forEachSenderInPlugins() -> Loading help from library: %s", libraryFullPath)
			plugin, err := plugin.Open(libraryFullPath)
			if err == nil {
				sym, err2 := plugin.Lookup("GetCommander")
				if err2 != nil {
					commander, errPlugin := sym.(func(string)(common.Commander, error))(command)
					if errPlugin != nil {
						continue
					}
					commander.SetLogger(Logger)
					senders = append(senders, commander)
				}
			}
		}
	}
	callback(senders)
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