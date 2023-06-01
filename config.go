package updater

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

const (
	Version = "0.0.1"
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
