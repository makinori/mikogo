package env

import (
	"log/slog"
	"os"
)

var (
	_, DEV = os.LookupEnv("DEV")

	// it will always be connected here
	HOME_SERVER = getEnv("HOME_SERVER", "127.0.0.1:6697")

	NICK = getEnv("NICK", "mikogo")

	// only listen to command from this nick on home server
	OWNER = getEnv("OWNER", "maki")
)

func getEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	return fallback
}

func init() {
	if DEV {
		slog.Warn("running in develop mode!")
	}
	slog.Info("env", "HOME_SERVER", HOME_SERVER, "OWNER", OWNER)
}
