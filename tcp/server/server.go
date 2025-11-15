package server

import (
	"net"
	"sync"

	"tcp/utils"

	"go.uber.org/zap"
)

var dict = NewDictionary()
var dictMutex sync.Mutex

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
		logger.Info("Received data", zap.ByteString("data", data))
		wg.Add(1)
		go processData(data, conn, logger, wg)
	}
}

func processData(data []byte, conn net.Conn, logger *zap.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info("Processing data", zap.ByteString("data", data))

	request, err := utils.ParseHTTPRequest(data)
	if err != nil {
		response := utils.HTTPResponse{
			StatusCode: 400,
			Message:    "Invalid request format: " + err.Error(),
		}
		logger.Warn("Invalid request", zap.Error(err))
		conn.Write(response.Bytes())
		return
	}

	logger.Info("Parsed request",
		zap.String("method", request.Method),
		zap.String("path", request.Path),
		zap.String("body", request.Body))

	/*
		==================================================
		Here you can implement any data processing logic.
		Use functions from server/utils.go as needed.
		==================================================
	*/
	response := ProcessDictCommand(request, dict, &dictMutex)

	/*
		==================================================
		End of data processing logic.
		==================================================
	*/

	_, err = conn.Write(response.Bytes())
	if err != nil {
		logger.Warn("Error writing to connection", zap.Error(err))
	}
}
