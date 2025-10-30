package conf

import (
	"bit-labs.cn/owl/contract/foundation"
	"bytes"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/fsnotify/fsnotify"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Configure struct {
	confDir        string
	cfgFileNameMap map[string]map[string]any // fileName => map[string]any
	app            foundation.Application
	eventBus       EventBus.Bus
}

func NewConfigure(app foundation.Application, bus EventBus.Bus) *Configure {

	confDir := app.ConfigPath("conf")

	manager := Configure{
		confDir:        confDir,
		cfgFileNameMap: make(map[string]map[string]any),
		eventBus:       bus,
	}

	err := filepath.Walk(confDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		cfgMap := make(map[string]any)
		ext := strings.Replace(filepath.Ext(info.Name()), ".", "", -1)
		name := strings.Replace(info.Name(), "."+ext, "", -1)
		absPath, v := manager.load(name, ext, &cfgMap)
		cfgMap["abs-path"] = absPath
		cfgMap["vip"] = v
		manager.cfgFileNameMap[name] = cfgMap
		return nil
	})
	if err != nil {
		panic(err)
	}
	return &manager
}

func (i *Configure) GetConfig(key string, v any) error {
	var getter jsoniter.Any

	marshal, err := jsoniter.Marshal(i.cfgFileNameMap)
	if err != nil {
		return err
	}

	pathArr := strings.Split(key, ".")

	for _, path := range pathArr {
		if getter != nil {
			getter = jsoniter.Get([]byte(getter.ToString()), path)
		} else {
			getter = jsoniter.Get(marshal, path)
		}
	}

	if getter != nil {
		err = jsoniter.UnmarshalFromString(getter.ToString(), v)
	}
	return err
}

func (i *Configure) SaveConfig(fileName string, key string, value any) {
	cfg, ok := i.cfgFileNameMap[fileName]
	if ok {
		v := cfg["vip"].(*viper.Viper)
		v.Set(key, value)
		err := v.WriteConfig()
		if err != nil {
			fmt.Println("保存配置失败")
		}
	}
}

// load 读取文件中的配置
func (i *Configure) load(fileName, cfgType string, c any) (string, *viper.Viper) {

	confFilePath := filepath.Join(i.confDir, fileName+"."+cfgType)

	v := viper.New()
	v.SetConfigType(cfgType)
	v.SetConfigFile(confFilePath)
	//i.l.Info("配置文件: ", confFilePath)
	cfg, err := os.ReadFile(confFilePath)
	if err != nil {
		panic(fmt.Sprint("读取配置文件失败", err))
	}
	if err = v.ReadConfig(bytes.NewReader(cfg)); err != nil {
		panic(fmt.Sprint("读取配置文件失败", err))
	}

	// 转换为结构体
	if err := v.Unmarshal(&c); err != nil {
		panic("转为配置结构体失败")
	}

	// Watch for changes in the config file
	v.WatchConfig()
	// Register a callback function to handle the changes
	v.OnConfigChange(func(e fsnotify.Event) {
		i.eventBus.Publish(ConfigChangeEvent, e)
	})
	return confFilePath, v
}
