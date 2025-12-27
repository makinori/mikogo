package env

import (
	"log/slog"
	"os"
	"runtime"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

var (
	_, DEV = os.LookupEnv("DEV")

	NICK = getEnv("NICK", "mikogo")

	// only listen to command from this nick on home server
	OWNER = getEnv("OWNER", "maki")

	// it will always be connected here
	HOME_SERVER = getEnv("HOME_SERVER", "127.0.0.1:6697")

	// injected at build
	GIT_COMMIT string
)

func getEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	return fallback
}

func GetGoVersion() string {
	return strings.TrimPrefix(
		strings.SplitN(runtime.Version(), " ", 2)[0], "go",
	)
}

func init() {
	if DEV {
		slog.Warn("running in develop mode!")
	}
	slog.Info("version", "commit", GIT_COMMIT, "go", GetGoVersion())
	slog.Info("using",
		"nick", NICK,
		"owner", OWNER,
		"home", HOME_SERVER,
	)
}
