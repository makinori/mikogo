package irc

import (
	"log/slog"
	"time"
)

func pingAllClients() {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	for _, client := range clients {
		client.ping()
	}
}

func init() {
	go func() {
		slog.Info("started global ping loop")
		for {
			time.Sleep(time.Second * 60)
			pingAllClients()
		}
	}()
}
