package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"tcp/client"
	"tcp/server"
	"tcp/utils"

	"go.uber.org/zap"
)

func main() {
	logger := utils.GetLogger()
	defer logger.Sync()

	// Allow environment variables to override defaults (passed from docker-compose via -e)
	envHost := os.Getenv("HOST")
	envPort := os.Getenv("PORT")

	// compute defaults
	addrDefault := "localhost"
	if envHost != "" {
		addrDefault = envHost
	}
	portDefault := 9000
	if envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			portDefault = p
		}
	}

	// Define flags
	mode := flag.String("mode", "", "Mode to run: 'server' or 'client'")
	address := flag.String("address", addrDefault, "Address to bind/connect to")
	port := flag.Int("port", portDefault, "Port to bind/connect to")

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

		logger.Info("Starting TCP server", zap.String("address", config.AddressString()))
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
