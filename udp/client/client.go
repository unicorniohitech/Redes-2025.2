package client

import (
	"fmt"
	"net"
	"time"
	"udp/utils"

	"github.com/manifoldco/promptui"
	"go.uber.org/zap"
)

func StartClient(config *Config) error {
	logger := utils.GetLogger()

	serverAddr, err := net.ResolveUDPAddr("udp", config.AddressString())
	if err != nil {
		logger.Warn("Error resolving address", zap.Error(err))
		return err
	}

	logger.Info("UDP Address resolved!")

	for {
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
		requestPacket := utils.NewPacket(request.Bytes())

		conn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			logger.Warn("Error connecting to server", zap.Error(err))
			return err
		}

		logger.Info("Connected to server", zap.String("address", config.AddressString()))

		for i, p := range requestPacket {
			logger.Info("Sending packet", zap.Int("packet_index", i), zap.ByteString("payload", p.Payload))
			_, err = conn.Write(p.Bytes())
			if err != nil {
				logger.Warn("Error sending data to server", zap.Error(err))
				conn.Close()
				return err
			}
			time.Sleep(10 * time.Millisecond)
		}

		buffer := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			logger.Warn("Error reading response from server", zap.Error(err))
			conn.Close()
			return err
		}

		crc := utils.NewCRC()
		responsePacket, err := utils.ParsePacket(buffer[:n])
		if err != nil {
			logger.Warn("Error parsing response packet", zap.Error(err))
			conn.Close()
			return err
		}
		logger.Info(
			"Received response packet",
			zap.Uint16("control", responsePacket.Control),
			zap.ByteString("payload", responsePacket.Payload),
			zap.Uint16("crc", responsePacket.CRC),
		)
		if !crc.ValidatePacket(*responsePacket) {
			logger.Warn("Response packet CRC not valid")
			conn.Close()
			return fmt.Errorf("response packet CRC not valid")
		}
		conn.Close()

		statusCode, statusText, body := ParseHTTPResponse(string(responsePacket.Payload))

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
