package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"udp/client"
	"udp/server"
	"udp/utils"

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
	portDefault := 8080
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
		fmt.Println("Usage: go run main.go -mode=<server|client|teste> [-address=<address>] [-port=<port>]")
		os.Exit(1)
	}

	switch *mode {
	case "server":
		config := server.NewConfig()
		config.SetAddress(*address)
		config.SetPort(*port)

		logger.Info("Starting UDP server", zap.String("address", config.AddressString()))
		if err := server.StartServer(config); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}

	case "client":
		config := client.NewConfig()
		config.SetAddress(*address)
		config.SetPort(*port)

		logger.Info("Starting UDP client", zap.String("address", config.AddressString()))
		client.StartClient(config)

	case "teste":
		logger.Info("Starting test client", zap.String("address", *address), zap.Int("port", *port))
		if err := client.RunTestClient(*address, *port, 1*time.Second); err != nil {
			logger.Fatal("Failed to run test client", zap.Error(err))
		}

	default:
		fmt.Printf("Error: invalid mode '%s'\n", *mode)
		fmt.Println("Mode must be either 'server', 'client', or 'teste'")
		os.Exit(1)
	}
}
