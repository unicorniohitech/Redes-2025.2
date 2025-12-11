package client

import (
	"fmt"
	"net"
	"sync"
	"time"
	"udp/utils"

	"github.com/manifoldco/promptui"
	"go.uber.org/zap"
)

var packetStorage = utils.NewPacketStore()
var packetStorageMutex sync.Mutex

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

		buffer := make([]byte, 2048)
		responsePayload := []byte{}
		responseFinished := false
		for !responseFinished {
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				logger.Warn("Error reading from connection", zap.Error(err))
				continue
			}
			data := make([]byte, n)
			copy(data, buffer[:n])
			logger.Info("Received data", zap.ByteString("data", data))
			processResponse(data, &responsePayload, &responseFinished, remoteAddr, logger)
		}
		conn.Close()

		statusCode, statusText, body := ParseHTTPResponse(string(responsePayload))

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

func processResponse(data []byte, response *[]byte, responseFinished *bool, remoteAddr *net.UDPAddr, logger *zap.Logger) {
	defer logger.Info("Finished processing data", zap.String("remote_addr", remoteAddr.String()))

	packet, err := utils.ParsePacket(data)
	if err != nil {
		logger.Warn("Error parsing packet", zap.Error(err))
		return
	}
	logger.Info("Parsed packet", zap.Uint16("control", packet.Control), zap.Uint16("length", packet.Length), zap.ByteString("payload", packet.Payload), zap.Uint16("crc", packet.CRC))

	*response, *responseFinished = verifyPacket(packet, packetStorage, &packetStorageMutex, remoteAddr, logger)
}

func verifyPacket(packet utils.Packet, ps *utils.PacketStore, mux *sync.Mutex, remoteAddr *net.UDPAddr, logger *zap.Logger) ([]byte, bool) {
	defer logger.Info("Finished processing data", zap.String("remote_addr", remoteAddr.String()))

	crc := utils.NewCRC()
	if !crc.ValidatePacket(packet) {
		logger.Info("Packet CRC not valid", zap.String("remote_addr", remoteAddr.String()))
		return []byte{}, false
	}
	payload := packet.Payload

	if packet.Length > 0 {
		mux.Lock()
		ps.AddPacket(remoteAddr.String(), packet)
		if ps.IsComplete(remoteAddr.String()) {
			logger.Info("Packet complete", zap.String("remote_addr", remoteAddr.String()))
			packets := ps.Packets[remoteAddr.String()]
			payload = utils.GetCompletePayload(packets)
			logger.Info("Complete payload received", zap.ByteString("payload", payload))
			delete(ps.Packets, remoteAddr.String())
			mux.Unlock()
		} else {
			mux.Unlock()
			return []byte{}, false
		}
	}

	return payload, true
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
