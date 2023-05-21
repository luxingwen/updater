package main

import (
	"os"
	"time"

	"updater"
)

func main() {

	var client *updater.Client
	var err error

	servers := make([]*updater.Server, 0)
	msghanlder := updater.NewMessageHandler(10)

	for _, item := range updater.GetConfig().ServerAddress {
		servers = append(servers, updater.NewServer(item))
	}
	for {
		client, err = updater.ConnectToServers(servers, msghanlder)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
	}

	msghanlder.HandleMessages(client, 10)

	client.Start()

	sig := make(chan os.Signal, 1)

	select {
	case <-sig:
		client.Stop()
	}

}
