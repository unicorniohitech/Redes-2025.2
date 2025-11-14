package server

import (
	"net"
	"sync"

	"tcp/utils"

	"go.uber.org/zap"
)

func StartServer(config *Config) error {
	logger := utils.GetLogger()
	wg := &sync.WaitGroup{}

	listener, err := net.Listen("tcp", config.AddressString())
	if err != nil {
		logger.Warn("Error starting server", zap.Error(err))
		return err
	}
	defer listener.Close()
	defer wg.Wait()
	logger.Info("Server started", zap.String("address", config.AddressString()))

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn("Error accepting connection", zap.Error(err))
			continue
		}
		logger.Info("Client connected", zap.String("remote_addr", conn.RemoteAddr().String()))
		wg.Add(1)
		go handleConnection(conn, logger, wg)
	}
}

func handleConnection(conn net.Conn, logger *zap.Logger, wg *sync.WaitGroup) {
	defer func() {
		logger.Info("Client disconnected", zap.String("remote_addr", conn.RemoteAddr().String()))
		conn.Close()
		wg.Done()
	}()
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			logger.Warn("Error reading from connection", zap.Error(err))
			return
		}
		data := buffer[:n]
		logger.Info("Received data", zap.String("data", string(data)))
		wg.Add(1)
		go processData(data, conn, logger, wg)
	}
}

func processData(data []byte, conn net.Conn, logger *zap.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info("Processing data", zap.String("data", string(data)))

	/*
		==================================================
		Here you can implement any data processing logic.
		Use functions from server/utils.go as needed.
		==================================================
	*/
	processedData := []byte(ToUppercase(string(data)))

	/*
		==================================================
		End of data processing logic.
		==================================================
	*/

	_, err := conn.Write(processedData)
	if err != nil {
		logger.Warn("Error writing to connection", zap.Error(err))
	}
}
