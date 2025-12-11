package server

import (
	"net"
	"sync"
	"time"

	"udp/utils"

	"go.uber.org/zap"
)

var dict = NewDictionary()
var dictMutex sync.Mutex

var packetStorage = utils.NewPacketStore()
var packetStorageMutex sync.Mutex

func StartServer(config *Config) error {
	logger := utils.GetLogger()
	wg := &sync.WaitGroup{}

	addr, err := net.ResolveUDPAddr("udp", config.AddressString())
	if err != nil {
		logger.Warn("Error resolving address", zap.Error(err))
		return err
	}
	logger.Info("UDP Address resolved!", zap.String("address", config.AddressString()))
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Warn("Error listening on UDP", zap.Error(err))
		return err
	}
	defer conn.Close()
	logger.Info("Listening on: ", zap.String("address", config.AddressString()))
	wg.Add(1)
	go handleConnection(*conn, logger, wg)
	wg.Wait()

	return nil
}

func handleConnection(conn net.UDPConn, logger *zap.Logger, wg *sync.WaitGroup) {
	defer func() {
		logger.Info("Client disconnected", zap.String("remote_addr", conn.RemoteAddr().String()))
		conn.Close()
		wg.Done()
	}()
	buffer := make([]byte, 2048)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			logger.Warn("Error reading from connection", zap.Error(err))
			return
		}
		data := make([]byte, n)
		copy(data, buffer[:n])
		logger.Info("Received data", zap.ByteString("data", data))
		wg.Add(1)
		go processPacket(data, &conn, remoteAddr, logger, wg)
	}

}

func processPacket(data []byte, conn *net.UDPConn, remoteAddr *net.UDPAddr, logger *zap.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	defer logger.Info("Finished processing data", zap.String("remote_addr", remoteAddr.String()))

	packet, err := utils.ParsePacket(data)
	if err != nil {
		logger.Warn("Error parsing packet", zap.Error(err))
		return
	}
	logger.Info("Parsed packet", zap.Uint16("control", packet.Control), zap.Uint16("length", packet.Length), zap.ByteString("payload", packet.Payload), zap.Uint16("crc", packet.CRC))

	payload, complete := verifyPacket(packet, packetStorage, &packetStorageMutex, remoteAddr, logger)
	if !complete {
		return
	}

	responseData, err := processData(payload, logger)
	if err != nil {
		logger.Warn("Error processing data", zap.Error(err))
	}
	responsePacket := utils.NewPacket(responseData)
	for i := range responsePacket {
		_, err = conn.WriteToUDP(responsePacket[i].Bytes(), remoteAddr)
		if err != nil {
			logger.Warn("Error writing to UDP connection", zap.Error(err))
		}
		time.Sleep(10 * time.Millisecond) // Small delay to avoid packet loss
	}
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

func processData(data []byte, logger *zap.Logger) ([]byte, error) {
	logger.Info("Processing data", zap.ByteString("data", data))

	request, err := utils.ParseHTTPRequest(data)
	if err != nil {
		response := utils.HTTPResponse{
			StatusCode: 400,
			Message:    "Invalid request format: " + err.Error(),
		}
		logger.Warn("Invalid request", zap.Error(err))
		return response.Bytes(), err
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

	return response.Bytes(), nil
}
