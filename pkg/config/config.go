package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

const (
	Version = "0.0.1"
)

type Config struct {
	ServerAddress []string  `json:"serverAddress"` // 代理服务器地址
	LogConfig     LogConfig `json:"logConfig"`
}

type LogConfig struct {
	Level       string `json:"level"`       // 日志级别
	Format      string `json:"format"`      // 日志格式
	MaxSize     int    `json:"maxSize"`     // 最大文件大小（MB）
	MaxAge      int    `json:"maxAge"`      // 最大文件保留天数
	Compress    bool   `json:"compress"`    // 是否压缩
	Filename    string `json:"filename"`    // 日志文件名
	ShowConsole bool   `json:"showConsole"` // 是否显示在控制台
}

var (
	config *Config
)

func InitConfig() {
	loadConfigFile()
}

func loadConfigFile() {

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.json"
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("读取配置文件失败")
	}
	config = new(Config)
	err = json.Unmarshal(b, config)
	if err != nil {
		log.Fatal("解析配置文件失败")
	}
}

func GetConfig() *Config {
	return config
}

func GetLocalIPs() (string, error) {
	var ips []string

	// 获取本机所有网络接口的地址
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// 遍历所有接口
	for _, iface := range interfaces {
		// 排除回环接口和无效接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取接口的地址列表
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println(err)
			continue
		}

		// 遍历接口的地址列表
		for _, addr := range addrs {
			// 检查地址的类型
			switch addr := addr.(type) {
			case *net.IPNet:
				// 排除IPv6地址
				if addr.IP.To4() != nil {
					ips = append(ips, addr.IP.String())
				}
			case *net.IPAddr:
				// 排除IPv6地址
				if addr.IP.To4() != nil {
					ips = append(ips, addr.IP.String())
				}
			}
		}
	}

	// 将IP地址使用逗号分隔并返回
	return strings.Join(ips, ","), nil
}
