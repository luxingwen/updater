package updater

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	ServerAddress []string `json:"serverAddress"`
}

var conf *Config

func init() {
	var err error

	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("读取配置文件失败")
	}
	conf = new(Config)
	err = json.Unmarshal(b, conf)
	if err != nil {
		log.Fatal("解析配置文件失败")
	}
}

func GetConfig() *Config {
	return conf
}
