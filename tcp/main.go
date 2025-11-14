package main

import (
	"flag"
	"fmt"
	"os"

	"tcp/client"
	"tcp/server"
	"tcp/utils"

	"go.uber.org/zap"
)

func main() {
	logger := utils.GetLogger()
	defer logger.Sync()

	// Define flags
	mode := flag.String("mode", "", "Mode to run: 'server' or 'client'")
	address := flag.String("address", "localhost", "Address to bind/connect to")
	port := flag.Int("port", 8000, "Port to bind/connect to")

	flag.Parse()

	// Validate mode
	if *mode == "" {
		fmt.Println("Error: mode flag is required")
		fmt.Println("Usage: go run main.go -mode=<server|client> [-address=<address>] [-port=<port>]")
		os.Exit(1)
	}

	switch *mode {
	case "server":
		config := server.NewConfig()
		config.SetAddress(*address)
		config.SetPort(*port)

		logger.Info("Starting server", zap.String("address", config.AddressString()))
		if err := server.StartServer(config); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}

	case "client":
		config := client.NewConfig()
		config.SetAddress(*address)
		config.SetPort(*port)

		logger.Info("Starting client", zap.String("address", config.AddressString()))
		if err := client.StartClient(config); err != nil {
			logger.Fatal("Failed to start client", zap.Error(err))
		}

	default:
		fmt.Printf("Error: invalid mode '%s'\n", *mode)
		fmt.Println("Mode must be either 'server' or 'client'")
		os.Exit(1)
	}
}
