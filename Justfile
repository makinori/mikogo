export CGO_ENABLED := "0"

default:
	@just --list

# ephemeral on 6667 and 6697
[group("dev")]
ergo:
	podman run --rm -it -p 6667:6667 -p 6697:6697 ghcr.io/ergochat/ergo

# ephemeral on 6767 and 6797
[group("dev")]
ergo2:
	podman run --rm -it -p 6767:6667 -p 6797:6697 ghcr.io/ergochat/ergo

alias s := start
[group("dev")]
start:
	DEV=1 go run -ldflags="-s -w \
	-X 'github.com/makinori/mikogo/env.GIT_COMMIT=$(git rev-parse HEAD | head -c 8)'\
	" .

alias u := update
# git pull, build and restart quadlet
[group("server")]
update:
	git pull
	systemctl --user daemon-reload
	systemctl --user start mikogo-build
	systemctl --user restart mikogo