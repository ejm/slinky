package main

import (
	"log/slog"

	internal "github.com/ejm/slinky/internal"
	env "go-simpler.org/env"
)

func main() {
	var config internal.Config
	if err := env.Load(&config, nil); err != nil {
		slog.Error(err.Error())
		return
	}

	server, err := internal.NewServer(config)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	server.Run()
}
