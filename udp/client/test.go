package client

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"udp/utils"

	"go.uber.org/zap"
)

// RunTestClient executa um cliente de teste que envia comandos LIST e LOOKUP a cada 1 segundo
func RunTestClient(address string, port int, interval time.Duration) error {
	logger := utils.GetLogger()

	serverAddrStr := fmt.Sprintf("%s:%d", address, port)
	logger.Info("Test Client Configuration",
		zap.String("server_address", serverAddrStr),
		zap.Duration("interval", interval),
	)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	nextCommand := "LIST"
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("Starting test client - Press Ctrl+C to stop")

	for {
		select {
		case <-sigChan:
			logger.Info("Shutdown signal received")
			return nil
		case <-ticker.C:
			command := nextCommand

			response, err := sendTestCommand(logger, serverAddrStr, command)
			if err != nil {
				logger.Warn("Error sending command", zap.String("command", command), zap.Error(err))
			}

			_, _, message := ParseHTTPResponse(string(response))
			if command == "LIST" {
				logger.Info("Dictionary contents", zap.String("dictionary", message))
				dict := DictionaryFromString(message)
				nextCommand = "LOOKUP " + dict.keys[rand.Intn(len(dict.keys))]
			} else {
				nextCommand = "LIST"
			}
		}
	}
}

func sendTestCommand(logger *zap.Logger, serverAddr string, command string) ([]byte, error) {
	logger.Info("Sending command", zap.String("command", command))

	// Resolve UDP address
	serverUDPAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("error resolving address: %w", err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, serverUDPAddr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}
	defer conn.Close()

	// Set read deadline (5 seconds)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Create request packet
	request, err := ParseCommandToHTTPRequest(command)
	if err != nil {
		return nil, fmt.Errorf("error parsing command: %w", err)
	}
	packet := utils.NewPacket(request.Bytes())

	// Send all packet fragments
	for i, p := range packet {
		logger.Debug("Sending packet fragment",
			zap.Int("fragment_index", i),
			zap.ByteString("payload", p.Payload),
		)

		_, err = conn.Write(p.Bytes())
		if err != nil {
			return nil, fmt.Errorf("error sending packet: %w", err)
		}

		// Small delay between fragments
		time.Sleep(10 * time.Millisecond)
	}

	// Receive response
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Parse response packet
	responsePacket, err := utils.ParsePacket(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	logger.Info("Response received",
		zap.Uint16("control", responsePacket.Control),
		zap.ByteString("payload", responsePacket.Payload),
		zap.Uint16("crc", responsePacket.CRC),
	)

	return responsePacket.Payload, nil
}

func DictionaryFromString(data string) *Dictionary {
	dict := &Dictionary{
		terms: make(map[string]string),
		keys:  []string{},
	}
	trimmed := data[1 : len(data)-1]
	if len(trimmed) == 0 {
		return dict
	}
	terms := strings.Split(trimmed, ", ")
	for _, term := range terms {
		dict.terms[term] = ""
		dict.keys = append(dict.keys, term)
	}
	return dict
}
