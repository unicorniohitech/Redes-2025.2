package client

import (
	"bufio"
	"net"
	"os"

	"tcp/utils"

	"go.uber.org/zap"
)

func StartClient(config *Config) error {
	logger := utils.GetLogger()

	conn, err := net.Dial("tcp", config.AddressString())
	if err != nil {
		logger.Warn("Error connecting to server", zap.Error(err))
		return err
	}
	defer conn.Close()
	logger.Info("Connected to server", zap.String("address", config.AddressString()))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		logger.Info("Enter message to send (or Ctrl+C to quit):")
		if !scanner.Scan() {
			break
		}

		message := scanner.Text()
		if message == "" {
			continue
		}

		/*
			==================================================
			Here you can implement any data processing logic
			before sending. Use functions from client/utils.go
			as needed.
			==================================================
		*/
		processedMessage := []byte(ToLowercase(message))

		/*
			==================================================
			End of data processing logic.
			==================================================
		*/

		_, err := conn.Write(processedMessage)
		if err != nil {
			logger.Warn("Error sending data", zap.Error(err))
			return err
		}
		logger.Info("Sent data", zap.String("data", string(processedMessage)))

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			logger.Warn("Error reading response", zap.Error(err))
			return err
		}

		response := buffer[:n]
		logger.Info("Received response", zap.String("data", string(response)))
	}

	return nil
}
