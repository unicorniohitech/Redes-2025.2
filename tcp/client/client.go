package client

import (
	"fmt"
	"net"
	"time"

	"tcp/utils"

	"github.com/manifoldco/promptui"
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

		prompt := promptui.Select{
			Label: "Selecione um comando",
			Items: []string{"LIST", "LOOKUP", "INSERT", "UPDATE"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		var message string

		switch result {
		case "LIST":
			message = "LIST"
		case "LOOKUP":
			term := promptString("Digite o termo para busca:")
			message = fmt.Sprintf("LOOKUP %s", term)
		case "INSERT":
			term := promptString("Termo:")
			def := promptString("Definição:")
			message = fmt.Sprintf("INSERT %s %s", term, def)
		case "UPDATE":
			term := promptString("Termo:")
			def := promptString("Nova definição:")
			message = fmt.Sprintf("UPDATE %s %s", term, def)
		}

		request, err := ParseCommandToHTTPRequest(message)
		if err != nil {
			logger.Warn("Invalid command format", zap.Error(err))
			logger.Info("Usage: <METHOD> [term] [definition]")
			continue
		}

		_, err = conn.Write(request.Bytes())
		if err != nil {
			logger.Warn("Error sending data", zap.Error(err))
			return err
		}

		err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err != nil {
			logger.Warn("Error setting up read deadline", zap.Error(err))
			return err
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			logger.Warn("Error reading response", zap.Error(err))
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logger.Info("Read timeout: no response within 30 seconds")
				connOK = false
				continue
			} else {
				return err
			}
		}

		responseStr := string(buffer[:n])
		statusCode, statusText, body := ParseHTTPResponse(responseStr)

		if statusCode >= 200 && statusCode < 300 {
			fmt.Printf("%s SUCCESS (%d %s): %s\n", utils.GetEmoji(statusCode), statusCode, statusText, body)
		} else if statusCode >= 400 && statusCode < 500 {
			fmt.Printf("%s CLIENT ERROR (%d %s): %s\n", utils.GetEmoji(statusCode), statusCode, statusText, body)
		} else if statusCode >= 500 {
			fmt.Printf("%s SERVER ERROR (%d %s): %s\n", utils.GetEmoji(statusCode), statusCode, statusText, body)
		} else {
			fmt.Printf("%s RESPONSE (%d %s): %s\n", utils.GetEmoji(statusCode), statusCode, statusText, body)
		}
	}
}

func promptString(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}
	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}
	return result
}
