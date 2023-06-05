package main

import (
	"os"
	"time"

	"updater"

	v1 "updater/controller/v1"
)

func main() {

	var client *updater.Client
	var err error

	servers := make([]*updater.Server, 0)
	msghanlder := updater.NewMessageHandler(10)

	v1.NewFileController(msghanlder)
	v1.NewAuthController(msghanlder)
	v1.NewScriptController(msghanlder)

	msghanlder.PrintRegisteredHandlers()

	for _, item := range updater.GetConfig().ServerAddress {
		servers = append(servers, updater.NewServer(item))
	}
	for {
		client, err = updater.ConnectToServers(servers, msghanlder)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	msghanlder.HandleMessages(client, 10)

	client.Start()

	sig := make(chan os.Signal, 1)

	select {
	case <-sig:
		client.Stop()
	}

}
