package main

import (
	"log/slog"

	"github.com/involvex/disgo-cli/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		slog.Error("failed to run command", "err", err)
	}
}
