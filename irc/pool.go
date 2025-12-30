package irc

import (
	"log/slog"
	"slices"
	"sync"

	"github.com/makinori/mikogo/db"
)

var (
	clients      = map[string]*Client{}
	clientsMutex = sync.Mutex{}
)

func IsConnectedSafe(address string) bool {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for _, client := range clients {
		if client.Address == address {
			return true
		}
	}
	return false
}

func Sync() error {
	servers, err := db.Servers.GetAll()
	if err != nil {
		return err
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	// ensure clients that need to be online first
	// also collect server names for later

	allServerNames := []string{}

	for name, server := range servers.AllFromBack() {
		allServerNames = append(allServerNames, name)

		if clients[name] == nil {
			clients[name] = newClient(server.Address)
		}

		if !clients[name].active {
			clients[name].init()
		}

		// TODO: also sync channels
	}

	// then remove clients that shouldnt be online

	for name := range clients {
		if !slices.Contains(allServerNames, name) {
			clients[name].delete()
			clients[name] = nil
		}
	}

	// then reconnect those that got a new address last
	// as we dont want to accidentally connect twice anywhere

	for name, server := range servers.AllFromBack() {
		client := clients[name]
		if clients[name] == nil {
			slog.Warn("server became unavailable mid-sync?")
			continue
		}

		if client.Address == server.Address {
			continue
		}

		slog.Info(
			"server address changed", "name", name,
			"from", client.Address, "to", server.Address,
		)

		client.Address = server.Address

		// only run reconnect if the client is connected
		// new address will be used regardless

		if client.state == ConnStateConnected {
			client.reconnect()
		}
	}

	return nil
}
