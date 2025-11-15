package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"tcp/utils"

	"go.uber.org/zap"
)

func StartClient(config *Config) error {
	logger := utils.GetLogger()
	connOK := false
	tryCount := 0

	conn, err := net.Dial("tcp", config.AddressString())
	if err != nil {
		logger.Warn("Error connecting to server", zap.Error(err))
		return err
	}
	connOK = true
	defer conn.Close()
	logger.Info("Connected to server", zap.String("address", config.AddressString()))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !connOK {
			logger.Warn("Connection to server lost. Trying to reconnect.")
			conn, err = net.Dial("tcp", config.AddressString())
			if err != nil {
				logger.Warn("Error reconnecting to server", zap.Error(err))
				tryCount++
				time.Sleep(5 * time.Second)
				continue
			}
			connOK = true
			tryCount = 0
		}
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
		request, err := ParseCommandToHTTPRequest(message)
		if err != nil {
			logger.Warn("Invalid command format", zap.Error(err))
			logger.Info("Usage: <METHOD> [term] [definition]")
			logger.Info("Examples:")
			logger.Info("  LIST")
			logger.Info("  LOOKUP golang")
			logger.Info("  INSERT golang A programming language")
			logger.Info("  UPDATE golang A statically typed language")
			continue
		}

		/*
			==================================================
			End of data processing logic.
			==================================================
		*/

		_, err = conn.Write(request.Bytes())
		if err != nil {
			logger.Warn("Error sending data", zap.Error(err))
			return err
		}
		// logger.Info("Sent request",
		// 	zap.String("method", request.Method),
		// 	zap.String("path", request.Path),
		// 	zap.String("body", request.Body))

		buffer := make([]byte, 1024)
		err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err != nil {
			logger.Warn("Error setting up read deadline", zap.Error(err))
			return err
		}
		n, err := conn.Read(buffer)
		if err != nil {
			logger.Warn("Error reading response", zap.Error(err))
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logger.Info("Read timeout: no response received within 30 seconds")
				connOK = false
				continue
			} else {
				return err
			}
		}

		responseStr := string(buffer[:n])
		// logger.Info("Received response", zap.String("response", responseStr))

		statusCode, statusText, message := ParseHTTPResponse(responseStr)

		// logger.Info("Received response",
		// 	zap.Int("status_code", statusCode),
		// 	zap.String("status", statusText),
		// 	zap.String("message", message))

		if statusCode >= 200 && statusCode < 300 {
			fmt.Printf("%s SUCCESS (%d %s): %s\n", utils.GetEmoji(statusCode), statusCode, statusText, message)
		} else if statusCode >= 400 && statusCode < 500 {
			fmt.Printf("%s CLIENT ERROR (%d %s): %s\n", utils.GetEmoji(statusCode), statusCode, statusText, message)
		} else if statusCode >= 500 {
			fmt.Printf("%s SERVER ERROR (%d %s): %s\n", utils.GetEmoji(statusCode), statusCode, statusText, message)
		} else {
			fmt.Printf("%s RESPONSE (%d %s): %s\n", utils.GetEmoji(statusCode), statusCode, statusText, message)
		}
	}

	return nil
}
