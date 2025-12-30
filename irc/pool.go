package irc

import (
	"log/slog"
	"slices"
	"sync"

	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/ircf"
)

var (
	clients      = map[string]*Client{}
	clientsMutex = sync.RWMutex{}
)

func GetStateWithFormatting(name string) string {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	client, ok := clients[name]
	if !ok {
		return ircf.Bold().Color(98, 40).Format("not found")
	}

	switch client.state {
	case ConnStateConnecting:
		return ircf.Bold().Color(98, 41).Format("connecting")
	case ConnStateConnected:
		return ircf.Bold().Color(98, 43).Format("connected")
	case ConnStateDisconnected:
		return ircf.Bold().Color(98, 40).Format("disconnected")
	}

	return "no idea"
}

func Sync() error {
	servers, err := db.Servers.GetAll()
	if err != nil {
		return err
	}

	// write lock
	(func() {
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
	})()

	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

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
