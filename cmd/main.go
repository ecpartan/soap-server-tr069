package main

import (
	"context"
	"os"

	"github.com/ecpartan/soap-server-tr069/internal/config"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/server"
	"github.com/ecpartan/soap-server-tr069/tasks"
)

// RunServer run the server

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//logger.L(ctx).Info("Start logger")
	cfg := config.GetConfig()
	logger.InitLogger(os.Stdout)
	tasks.InitTasks()

	s, err := server.NewServer(ctx, cfg)

	if err != nil {
		logger.LogDebug("error creating server", err)
		return
	}
	s.Register()

	err = s.Run(ctx)

}
