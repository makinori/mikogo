default:
	@just --list

[group("dev")]
ergo:
	podman run --rm -it -p 6667:6667 -p 6697:6697 \
	ghcr.io/ergochat/ergo

alias s := start
[group("dev")]
start:
	DEV=1 go run .

