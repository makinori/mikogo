package irc

import (
	"log/slog"

	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/ircf"
)

// > insert funny reimu image here

func ReportIncident(msg string) {
	homeClient := clients["home"]
	if homeClient == nil {
		slog.Error(
			"failed to find home client whilst reporting incident",
			"msg", msg,
		)
		return
	}

	if !homeClient.active {
		slog.Error(
			"home client not active whilst reporting incident",
			"msg", msg,
		)
		return
	}

	slog.Info("incident", "msg", msg)
	homeClient.Send(env.OWNER,
		ircf.Color(98, 40).Bold().Format("incident")+": "+msg,
	)
}
