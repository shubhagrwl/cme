package main

import (
	"cme/internal/app/api/server"
	"cme/internal/app/constants"
	"cme/internal/app/db"
	"cme/internal/app/service/logger"
	"cme/internal/config"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var err error
	// Returns a struct with values from env variables
	constants.Config, err = config.LoadConfig()
	if err != nil {
		panic(err.Error())
	}
	// Creates an empty context that can be passed around
	ctx := context.Background()

	// Initialize the logger
	logger.InitLogger()
	log := logger.Logger(ctx)

	// Initialize the database session
	session, err := db.Init(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a new DBService instance
	dbService := db.New(session)

	r := server.Init(ctx, dbService)
	if err := r.Run(fmt.Sprintf("%s:%s", constants.Config.HTTPServerConfig.HTTPSERVER_LISTEN, constants.Config.HTTPServerConfig.HTTPSERVER_PORT)); err != nil {
		log.Fatal("Server not able to startup with error: ", err)
	}

	// Set up a channel to listen for interrupt or termination signals
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-sigChannel

	// Close the session before exiting
	session.Close()
}
